package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hmans/beans/internal/agent"
	"github.com/hmans/beans/internal/graph"
	"github.com/hmans/beans/internal/graph/model"
	"github.com/hmans/beans/internal/worktree"
	"github.com/hmans/beans/pkg/beancore"
	"github.com/hmans/beans/pkg/config"
	"github.com/hmans/beans/pkg/safepath"
)

// viewState represents which view is currently active
type viewState int

const (
	viewList viewState = iota
	viewDetail
	viewTagPicker
	viewParentPicker
	viewStatusPicker
	viewTypePicker
	viewBlockingPicker
	viewPriorityPicker
	viewCreateModal
	viewHelpOverlay
	viewAgentChat
	viewInteraction
)

// Two-column layout constants
const (
	TwoColumnMinWidth = 120 // minimum terminal width for two-column layout
	RightPaneMaxWidth = 80  // max width of preview pane (text files follow 80 char convention)
)

// calculatePaneWidths returns (leftWidth, rightWidth) for two-column layout.
// Right pane is capped at RightPaneMaxWidth, left pane gets remaining space.
func calculatePaneWidths(totalWidth int) (int, int) {
	rightWidth := RightPaneMaxWidth
	if totalWidth-rightWidth < 40 { // ensure left pane has reasonable minimum
		rightWidth = totalWidth - 40
	}
	leftWidth := totalWidth - rightWidth - 1 // 1 for separator
	return leftWidth, rightWidth
}

// beansChangedMsg is sent when beans change on disk (via file watcher)
type beansChangedMsg struct{}

// cursorChangedMsg is sent when the list cursor moves to a different bean
type cursorChangedMsg struct {
	beanID string
}

// openTagPickerMsg requests opening the tag picker
type openTagPickerMsg struct{}

// tagSelectedMsg is sent when a tag is selected from the picker
type tagSelectedMsg struct {
	tag string
}

// clearFilterMsg is sent to clear any active filter
type clearFilterMsg struct{}

// copyBeanIDMsg requests copying bean ID(s) to the clipboard
type copyBeanIDMsg struct {
	ids []string
}

// openEditorMsg requests opening the editor for a bean
type openEditorMsg struct {
	beanID   string
	beanPath string
}

// editorFinishedMsg is sent when the editor closes
type editorFinishedMsg struct {
	err error
}

// agentUpdatedMsg is sent when any agent session changes (via global subscription)
type agentUpdatedMsg struct{}

// agentSessionMsg is sent when a specific agent session changes (per-bean subscription)
type agentSessionMsg struct {
	session *agent.Session
}

// startAgentMsg requests starting an agent for a bean
type startAgentMsg struct {
	beanID string
}

// openAgentChatMsg requests opening the agent chat for a bean
type openAgentChatMsg struct {
	beanID string
}

// openParentPickerMsg requests opening the parent picker for bean(s)
type openParentPickerMsg struct {
	beanIDs       []string // IDs of beans to update
	beanTitle     string   // Display title (single title or "N selected beans")
	beanTypes     []string // Types of the beans (to filter eligible parents)
	currentParent string   // Only meaningful for single bean
}

// App is the main TUI application model
type App struct {
	state          viewState
	list           listModel
	detail         detailModel
	preview        previewModel
	tagPicker      tagPickerModel
	parentPicker   parentPickerModel
	statusPicker   statusPickerModel
	typePicker     typePickerModel
	blockingPicker blockingPickerModel
	priorityPicker priorityPickerModel
	createModal    createModalModel
	helpOverlay    helpOverlayModel
	agentPanel     agentPanelModel
	agentChat      agentChatModel
	interaction    interactionModel
	history        []detailModel // stack of previous detail views for back navigation
	core           *beancore.Core
	resolver       *graph.Resolver
	config         *config.Config
	agentMgr       *agent.Manager
	worktreeMgr    *worktree.Manager
	width          int
	height         int
	program        *tea.Program // reference to program for sending messages from watcher

	// Key chord state - tracks partial key sequences like "g" waiting for "t"
	pendingKey string

	// Modal state - tracks view behind modal pickers
	previousState viewState

	// Agent state
	globalSubCh   chan struct{} // global agent subscription channel
	agentChatBean string       // beanID of the agent chat currently being viewed
	beanSubCh     chan struct{} // per-bean agent subscription channel

	// Editor state - tracks bean being edited to update updated_at on save
	editingBeanID      string
	editingBeanModTime time.Time
}

// New creates a new TUI application
func New(core *beancore.Core, cfg *config.Config, agentMgr *agent.Manager, wtMgr *worktree.Manager) *App {
	resolver := &graph.Resolver{Core: core}
	return &App{
		state:       viewList,
		core:        core,
		resolver:    resolver,
		config:      cfg,
		agentMgr:    agentMgr,
		worktreeMgr: wtMgr,
		list:        newListModel(resolver, cfg),
		preview:     newPreviewModel(nil, 0, 0),
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	cmds := []tea.Cmd{a.list.Init()}
	if a.agentMgr != nil {
		a.globalSubCh = a.agentMgr.SubscribeGlobal()
		cmds = append(cmds, waitForAgentUpdate(a.globalSubCh))
	}
	return tea.Batch(cmds...)
}

// waitForAgentUpdate blocks on the global subscription channel and returns an agentUpdatedMsg
func waitForAgentUpdate(ch chan struct{}) tea.Cmd {
	return func() tea.Msg {
		<-ch
		return agentUpdatedMsg{}
	}
}

// waitForSessionUpdate blocks on a per-bean subscription channel and returns an agentSessionMsg
func waitForSessionUpdate(mgr *agent.Manager, beanID string, ch chan struct{}) tea.Cmd {
	return func() tea.Msg {
		<-ch
		session := mgr.GetSession(beanID)
		return agentSessionMsg{session: session}
	}
}

// isTwoColumnMode returns true if the terminal width supports two-column layout
func (a *App) isTwoColumnMode() bool {
	return a.width >= TwoColumnMinWidth
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		// Update preview dimensions if in two-column mode
		if a.isTwoColumnMode() {
			_, rightWidth := calculatePaneWidths(a.width)
			a.preview.width = rightWidth
			a.preview.height = a.height - 2
		}

	case agentUpdatedMsg:
		// Refresh the agent panel with current session states
		if a.agentMgr != nil {
			a.agentPanel = a.agentPanel.refresh(a.agentMgr, a.resolver)

			// Check for pending interactions
			for _, ai := range a.agentPanel.agents {
				if ai.pendingInteraction != nil && a.state != viewInteraction {
					a.previousState = a.state
					a.interaction = newInteractionModel(ai.beanID, ai.title, ai.pendingInteraction, a.width, a.height)
					a.state = viewInteraction
					break
				}
			}

			return a, waitForAgentUpdate(a.globalSubCh)
		}
		return a, nil

	case agentSessionMsg:
		// Update the agent chat view with new session data
		if a.state == viewAgentChat {
			a.agentChat = a.agentChat.updateSession(msg.session)

			// Check for pending interactions
			if msg.session != nil && msg.session.PendingInteraction != nil && a.state != viewInteraction {
				a.previousState = a.state
				title := a.agentChatBean
				if b, err := a.resolver.Query().Bean(context.Background(), a.agentChatBean); err == nil && b != nil {
					title = b.Title
				}
				a.interaction = newInteractionModel(a.agentChatBean, title, msg.session.PendingInteraction, a.width, a.height)
				a.state = viewInteraction
			}
		}
		if a.beanSubCh != nil {
			return a, waitForSessionUpdate(a.agentMgr, a.agentChatBean, a.beanSubCh)
		}
		return a, nil

	case startAgentMsg:
		if a.agentMgr == nil {
			return a, nil
		}

		// Determine work directory
		workDir := a.core.Root()
		if a.worktreeMgr != nil {
			// Worktree mode: create a worktree for this bean
			wt, err := a.worktreeMgr.Create(msg.beanID)
			if err == nil {
				workDir = wt.Path
			}
			// In worktree mode, multiple agents can run — no need to stop others
		} else {
			// Direct mode: stop any existing agent for a different bean
			running := a.agentMgr.ListRunningSessions()
			for _, r := range running {
				if r.BeanID != msg.beanID {
					a.agentMgr.StopSession(r.BeanID)
				}
			}
		}

		// Build context message from bean
		contextMsg := ""
		if b, err := a.resolver.Query().Bean(context.Background(), msg.beanID); err == nil && b != nil {
			contextMsg = fmt.Sprintf("Please work on bean %s: %q\n\nType: %s | Status: %s\n", b.ID, b.Title, b.Type, b.Status)
			if b.Body != "" {
				contextMsg += fmt.Sprintf("\nDescription:\n%s", b.Body)
			}
		}
		if contextMsg == "" {
			contextMsg = fmt.Sprintf("Please work on bean %s", msg.beanID)
		}
		_ = a.agentMgr.SendMessage(msg.beanID, workDir, contextMsg, nil)
		// Open agent chat for this bean
		return a, func() tea.Msg { return openAgentChatMsg{beanID: msg.beanID} }

	case openAgentChatMsg:
		// Unsubscribe from previous bean if any
		if a.beanSubCh != nil && a.agentMgr != nil {
			a.agentMgr.Unsubscribe(a.agentChatBean, a.beanSubCh)
			a.beanSubCh = nil
		}
		a.agentChatBean = msg.beanID
		a.previousState = a.state
		a.state = viewAgentChat

		// Get current session
		session := a.agentMgr.GetSession(msg.beanID)
		title := msg.beanID
		if b, err := a.resolver.Query().Bean(context.Background(), msg.beanID); err == nil && b != nil {
			title = b.Title
		}
		a.agentChat = newAgentChatModel(msg.beanID, title, session, a.width, a.height)

		// Subscribe to updates
		a.beanSubCh = a.agentMgr.Subscribe(msg.beanID)
		return a, tea.Batch(
			a.agentChat.Init(),
			waitForSessionUpdate(a.agentMgr, msg.beanID, a.beanSubCh),
		)

	case tea.KeyMsg:
		// Clear status messages on any keypress
		a.list.statusMessage = ""
		a.detail.statusMessage = ""

		// Handle ctrl+a toggle for agent chat
		if msg.String() == "ctrl+a" && a.agentMgr != nil {
			if a.state == viewAgentChat {
				// Return to previous view
				prevState := a.previousState
				// Unsubscribe from per-bean updates
				if a.beanSubCh != nil {
					a.agentMgr.Unsubscribe(a.agentChatBean, a.beanSubCh)
					a.beanSubCh = nil
				}
				a.state = prevState
				return a, nil
			}
			// Try to open chat for current bean or most recent agent
			var targetBean string
			if a.state == viewDetail && a.detail.bean != nil {
				s := a.agentMgr.GetSession(a.detail.bean.ID)
				if s != nil {
					targetBean = a.detail.bean.ID
				}
			}
			if targetBean == "" {
				// Use most recently active agent (check all sessions, not just running)
				sessions := a.agentMgr.ListActiveSessions()
				if len(sessions) > 0 {
					targetBean = sessions[0].BeanID
				}
			}
			if targetBean != "" {
				return a, func() tea.Msg { return openAgentChatMsg{beanID: targetBean} }
			}
			return a, nil
		}

		// Handle key chord sequences
		if a.state == viewList && a.list.list.FilterState() != 1 {
			if a.pendingKey == "g" {
				a.pendingKey = ""
				switch msg.String() {
				case "t":
					// "g t" - go to tags
					return a, func() tea.Msg { return openTagPickerMsg{} }
				default:
					// Invalid second key, ignore the chord
				}
				// Don't forward this key since it was part of a chord attempt
				return a, nil
			}

			// Start of potential chord
			if msg.String() == "g" {
				a.pendingKey = "g"
				return a, nil
			}
		}

		// Clear pending key on any other key press
		a.pendingKey = ""

		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "?":
			// Open help overlay if not already showing it (and not in a picker/modal)
			if a.state == viewList || a.state == viewDetail {
				a.previousState = a.state
				a.helpOverlay = newHelpOverlayModel(a.width, a.height)
				a.state = viewHelpOverlay
				return a, a.helpOverlay.Init()
			}
		case "q":
			if a.state == viewDetail || a.state == viewTagPicker || a.state == viewParentPicker || a.state == viewStatusPicker || a.state == viewTypePicker || a.state == viewBlockingPicker || a.state == viewPriorityPicker || a.state == viewHelpOverlay {
				return a, tea.Quit
			}
			// For list, only quit if not filtering
			if a.state == viewList && a.list.list.FilterState() != 1 {
				return a, tea.Quit
			}
		}

	case cursorChangedMsg:
		// Update preview with the newly highlighted bean
		_, rightWidth := calculatePaneWidths(a.width)
		if msg.beanID != "" {
			bean, err := a.resolver.Query().Bean(context.Background(), msg.beanID)
			if err == nil && bean != nil {
				a.preview = newPreviewModel(bean, rightWidth, a.height-2)
			}
		} else {
			a.preview = newPreviewModel(nil, rightWidth, a.height-2)
		}
		return a, nil

	case beansLoadedMsg:
		// Forward to list view
		a.list, cmd = a.list.Update(msg)
		// Update preview with current cursor position
		_, rightWidth := calculatePaneWidths(a.width)
		if len(msg.items) == 0 {
			a.preview = newPreviewModel(nil, rightWidth, a.height-2)
		} else if item, ok := a.list.list.SelectedItem().(beanItem); ok {
			a.preview = newPreviewModel(item.bean, rightWidth, a.height-2)
		}
		return a, cmd

	case beansChangedMsg:
		// Beans changed on disk - refresh
		if a.state == viewDetail {
			// Try to reload the current bean via GraphQL
			updatedBean, err := a.resolver.Query().Bean(context.Background(), a.detail.bean.ID)
			if err != nil || updatedBean == nil {
				// Bean was deleted - return to list
				a.state = viewList
				a.history = nil
			} else {
				// Recreate detail view with fresh bean data
				a.detail = newDetailModel(updatedBean, a.resolver, a.config, a.width, a.height)
			}
		}
		// Trigger list refresh
		return a, a.list.loadBeans

	case openTagPickerMsg:
		// Collect all tags with their counts
		tags := a.collectTagsWithCounts()
		if len(tags) == 0 {
			// No tags in system, don't open picker
			return a, nil
		}
		a.tagPicker = newTagPickerModel(tags, a.width, a.height)
		a.state = viewTagPicker
		return a, a.tagPicker.Init()

	case tagSelectedMsg:
		a.state = viewList
		a.list.setTagFilter(msg.tag)
		return a, a.list.loadBeans

	case openParentPickerMsg:
		// Check if all bean types can have parents
		for _, beanType := range msg.beanTypes {
			if beancore.ValidParentTypes(beanType) == nil {
				// At least one bean type (e.g., milestone) cannot have parents - don't open the picker
				return a, nil
			}
		}
		a.previousState = a.state // Remember where we came from for the modal background
		a.parentPicker = newParentPickerModel(msg.beanIDs, msg.beanTitle, msg.beanTypes, msg.currentParent, a.resolver, a.config, a.width, a.height)
		a.state = viewParentPicker
		return a, a.parentPicker.Init()

	case closeParentPickerMsg:
		// Return to previous view and refresh in case beans changed while picker was open
		a.state = a.previousState
		return a, a.list.loadBeans

	case openStatusPickerMsg:
		a.previousState = a.state
		a.statusPicker = newStatusPickerModel(msg.beanIDs, msg.beanTitle, msg.currentStatus, a.config, a.width, a.height)
		a.state = viewStatusPicker
		return a, a.statusPicker.Init()

	case closeStatusPickerMsg:
		// Return to previous view and refresh in case beans changed while picker was open
		a.state = a.previousState
		return a, a.list.loadBeans

	case statusSelectedMsg:
		// Update all beans' status via GraphQL mutations
		for _, beanID := range msg.beanIDs {
			_, err := a.resolver.Mutation().UpdateBean(context.Background(), beanID, model.UpdateBeanInput{
				Status: &msg.status,
			})
			if err != nil {
				// Continue with other beans even if one fails
				continue
			}
		}
		// Return to the previous view and refresh
		a.state = a.previousState
		// Clear selection after batch edit
		clear(a.list.selectedBeans)
		if a.state == viewDetail && len(msg.beanIDs) == 1 {
			updatedBean, _ := a.resolver.Query().Bean(context.Background(), msg.beanIDs[0])
			if updatedBean != nil {
				a.detail = newDetailModel(updatedBean, a.resolver, a.config, a.width, a.height)
			}
		}
		return a, a.list.loadBeans

	case openTypePickerMsg:
		a.previousState = a.state
		a.typePicker = newTypePickerModel(msg.beanIDs, msg.beanTitle, msg.currentType, a.config, a.width, a.height)
		a.state = viewTypePicker
		return a, a.typePicker.Init()

	case closeTypePickerMsg:
		// Return to previous view and refresh in case beans changed while picker was open
		a.state = a.previousState
		return a, a.list.loadBeans

	case typeSelectedMsg:
		// Update all beans' type via GraphQL mutations
		for _, beanID := range msg.beanIDs {
			_, err := a.resolver.Mutation().UpdateBean(context.Background(), beanID, model.UpdateBeanInput{
				Type: &msg.beanType,
			})
			if err != nil {
				// Continue with other beans even if one fails
				continue
			}
		}
		// Return to the previous view and refresh
		a.state = a.previousState
		// Clear selection after batch edit
		clear(a.list.selectedBeans)
		if a.state == viewDetail && len(msg.beanIDs) == 1 {
			updatedBean, _ := a.resolver.Query().Bean(context.Background(), msg.beanIDs[0])
			if updatedBean != nil {
				a.detail = newDetailModel(updatedBean, a.resolver, a.config, a.width, a.height)
			}
		}
		return a, a.list.loadBeans

	case openPriorityPickerMsg:
		a.previousState = a.state
		a.priorityPicker = newPriorityPickerModel(msg.beanIDs, msg.beanTitle, msg.currentPriority, a.config, a.width, a.height)
		a.state = viewPriorityPicker
		return a, a.priorityPicker.Init()

	case closePriorityPickerMsg:
		// Return to previous view and refresh in case beans changed while picker was open
		a.state = a.previousState
		return a, a.list.loadBeans

	case prioritySelectedMsg:
		// Update all beans' priority via GraphQL mutations
		for _, beanID := range msg.beanIDs {
			_, err := a.resolver.Mutation().UpdateBean(context.Background(), beanID, model.UpdateBeanInput{
				Priority: &msg.priority,
			})
			if err != nil {
				// Continue with other beans even if one fails
				continue
			}
		}
		// Return to the previous view and refresh
		a.state = a.previousState
		// Clear selection after batch edit
		clear(a.list.selectedBeans)
		if a.state == viewDetail && len(msg.beanIDs) == 1 {
			updatedBean, _ := a.resolver.Query().Bean(context.Background(), msg.beanIDs[0])
			if updatedBean != nil {
				a.detail = newDetailModel(updatedBean, a.resolver, a.config, a.width, a.height)
			}
		}
		return a, a.list.loadBeans

	case openHelpMsg:
		a.previousState = a.state
		a.helpOverlay = newHelpOverlayModel(a.width, a.height)
		a.state = viewHelpOverlay
		return a, a.helpOverlay.Init()

	case closeHelpMsg:
		a.state = a.previousState
		return a, nil

	case openBlockingPickerMsg:
		a.previousState = a.state
		a.blockingPicker = newBlockingPickerModel(msg.beanID, msg.beanTitle, msg.currentBlocking, a.resolver, a.config, a.width, a.height)
		a.state = viewBlockingPicker
		return a, a.blockingPicker.Init()

	case closeBlockingPickerMsg:
		// Return to previous view and refresh in case beans changed while picker was open
		a.state = a.previousState
		return a, a.list.loadBeans

	case blockingConfirmedMsg:
		// Apply all blocking changes via GraphQL mutations
		for _, targetID := range msg.toAdd {
			_, err := a.resolver.Mutation().AddBlocking(context.Background(), msg.beanID, targetID, nil)
			if err != nil {
				// Continue with other changes even if one fails
				continue
			}
		}
		for _, targetID := range msg.toRemove {
			_, err := a.resolver.Mutation().RemoveBlocking(context.Background(), msg.beanID, targetID, nil)
			if err != nil {
				// Continue with other changes even if one fails
				continue
			}
		}
		// Return to previous view and refresh
		a.state = a.previousState
		if a.state == viewDetail {
			updatedBean, _ := a.resolver.Query().Bean(context.Background(), msg.beanID)
			if updatedBean != nil {
				a.detail = newDetailModel(updatedBean, a.resolver, a.config, a.width, a.height)
			}
		}
		return a, a.list.loadBeans

	case openCreateModalMsg:
		a.previousState = a.state
		a.createModal = newCreateModalModel(a.width, a.height)
		a.state = viewCreateModal
		return a, a.createModal.Init()

	case closeCreateModalMsg:
		a.state = a.previousState
		return a, nil

	case beanCreatedMsg:
		// Create the bean via GraphQL mutation with draft status
		draftStatus := "draft"
		createdBean, err := a.resolver.Mutation().CreateBean(context.Background(), model.CreateBeanInput{
			Title:  msg.title,
			Status: &draftStatus,
		})
		if err != nil {
			// TODO: Show error to user
			a.state = a.previousState
			return a, nil
		}
		// Return to list and open the new bean in editor
		a.state = viewList
		return a, tea.Batch(
			a.list.loadBeans,
			func() tea.Msg {
				return openEditorMsg{beanID: createdBean.ID, beanPath: createdBean.Path}
			},
		)

	case openEditorMsg:
		// Launch editor for the bean file
		editor := getEditor()
		fullPath, err := safepath.SafeJoin(a.core.Root(), msg.beanPath)
		if err != nil {
			a.list.statusMessage = fmt.Sprintf("unsafe bean path: %v", err)
			return a, nil
		}

		// Record the bean ID and file mod time before editing
		a.editingBeanID = msg.beanID
		if info, err := os.Stat(fullPath); err == nil {
			a.editingBeanModTime = info.ModTime()
		}

		c := exec.Command(editor, fullPath)
		return a, tea.ExecProcess(c, func(err error) tea.Msg {
			return editorFinishedMsg{err: err}
		})

	case editorFinishedMsg:
		// Editor closed - check if file was modified and update updated_at if so
		if a.editingBeanID != "" {
			if b, err := a.core.Get(a.editingBeanID); err == nil {
				fullPath, pathErr := safepath.SafeJoin(a.core.Root(), b.Path)
				if pathErr != nil {
					break
				}
				if info, err := os.Stat(fullPath); err == nil {
					if info.ModTime().After(a.editingBeanModTime) {
						// File was modified - reload from disk first to get user's changes,
						// then call Update to set updated_at
						_ = a.core.Load()
						if b, err = a.core.Get(a.editingBeanID); err == nil {
							_ = a.core.Update(b, nil)
						}
					}
				}
			}
			// Clear editing state
			a.editingBeanID = ""
			a.editingBeanModTime = time.Time{}
		}
		return a, nil

	case parentSelectedMsg:
		// Set the new parent via GraphQL mutation for all beans
		var parentID *string
		if msg.parentID != "" {
			parentID = &msg.parentID
		}
		for _, beanID := range msg.beanIDs {
			_, err := a.resolver.Mutation().SetParent(context.Background(), beanID, parentID, nil)
			if err != nil {
				// Continue with other beans even if one fails
				continue
			}
		}
		// Return to the previous view and refresh
		a.state = a.previousState
		// Clear selection after batch edit
		clear(a.list.selectedBeans)
		if a.state == viewDetail && len(msg.beanIDs) == 1 {
			// Refresh the bean to show updated parent
			updatedBean, _ := a.resolver.Query().Bean(context.Background(), msg.beanIDs[0])
			if updatedBean != nil {
				a.detail = newDetailModel(updatedBean, a.resolver, a.config, a.width, a.height)
			}
		}
		return a, a.list.loadBeans

	case clearFilterMsg:
		a.list.clearFilter()
		return a, a.list.loadBeans

	case copyBeanIDMsg:
		var statusMsg string
		text := strings.Join(msg.ids, ", ")
		if err := clipboard.WriteAll(text); err != nil {
			statusMsg = fmt.Sprintf("Failed to copy: %v", err)
		} else if len(msg.ids) == 1 {
			statusMsg = fmt.Sprintf("Copied %s to clipboard", msg.ids[0])
		} else {
			statusMsg = fmt.Sprintf("Copied %d bean IDs to clipboard", len(msg.ids))
		}

		// Set status on current view
		if a.state == viewList {
			a.list.statusMessage = statusMsg
		} else if a.state == viewDetail {
			a.detail.statusMessage = statusMsg
		}

		return a, nil

	case selectBeanMsg:
		// Push current detail view to history if we're already viewing a bean
		if a.state == viewDetail {
			a.history = append(a.history, a.detail)
		}
		a.state = viewDetail
		a.detail = newDetailModel(msg.bean, a.resolver, a.config, a.width, a.height)
		return a, a.detail.Init()

	case backToListMsg:
		// Pop from history if available, otherwise go to list
		if len(a.history) > 0 {
			a.detail = a.history[len(a.history)-1]
			a.history = a.history[:len(a.history)-1]
			// Stay in viewDetail state
		} else {
			a.state = viewList
			// Force list to pick up any size changes that happened while in detail view
			a.list, cmd = a.list.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
			return a, cmd
		}
		return a, nil
	}

	// Forward all messages to the current view
	switch a.state {
	case viewList:
		a.list, cmd = a.list.Update(msg)
	case viewDetail:
		a.detail, cmd = a.detail.Update(msg)
	case viewTagPicker:
		a.tagPicker, cmd = a.tagPicker.Update(msg)
	case viewParentPicker:
		a.parentPicker, cmd = a.parentPicker.Update(msg)
	case viewStatusPicker:
		a.statusPicker, cmd = a.statusPicker.Update(msg)
	case viewTypePicker:
		a.typePicker, cmd = a.typePicker.Update(msg)
	case viewPriorityPicker:
		a.priorityPicker, cmd = a.priorityPicker.Update(msg)
	case viewBlockingPicker:
		a.blockingPicker, cmd = a.blockingPicker.Update(msg)
	case viewCreateModal:
		a.createModal, cmd = a.createModal.Update(msg)
	case viewHelpOverlay:
		a.helpOverlay, cmd = a.helpOverlay.Update(msg)
	case viewAgentChat:
		a.agentChat, cmd = a.agentChat.Update(msg)
		// Check if chat view wants to close
		if a.agentChat.done {
			if a.beanSubCh != nil && a.agentMgr != nil {
				a.agentMgr.Unsubscribe(a.agentChatBean, a.beanSubCh)
				a.beanSubCh = nil
			}
			a.state = a.previousState
			// Re-send window size so the restored view renders correctly
			return a, func() tea.Msg {
				return tea.WindowSizeMsg{Width: a.width, Height: a.height}
			}
		}
		// Check if chat view wants to send a message
		if a.agentChat.pendingMsg != "" {
			msg := a.agentChat.pendingMsg
			a.agentChat.pendingMsg = ""
			_ = a.agentMgr.SendMessage(a.agentChatBean, a.core.Root(), msg, nil)
		}
		// Check if chat view wants to stop the agent
		if a.agentChat.wantStop {
			a.agentChat.wantStop = false
			a.agentMgr.StopSession(a.agentChatBean)
		}
	case viewInteraction:
		a.interaction, cmd = a.interaction.Update(msg)
		if a.interaction.done {
			if a.interaction.response != "" {
				_ = a.agentMgr.SendMessage(a.interaction.beanID, a.core.Root(), a.interaction.response, nil)
			}
			a.state = a.previousState
			return a, nil
		}
	}

	return a, cmd
}

// collectTagsWithCounts returns all tags with their usage counts
func (a *App) collectTagsWithCounts() []tagWithCount {
	beans, _ := a.resolver.Query().Beans(context.Background(), nil)
	tagCounts := make(map[string]int)
	for _, b := range beans {
		for _, tag := range b.Tags {
			tagCounts[tag]++
		}
	}

	tags := make([]tagWithCount, 0, len(tagCounts))
	for tag, count := range tagCounts {
		tags = append(tags, tagWithCount{tag: tag, count: count})
	}

	return tags
}

// renderTwoColumnView renders the list and preview side by side with app-global footer
func (a *App) renderTwoColumnView() string {
	leftWidth, rightWidth := calculatePaneWidths(a.width)
	contentHeight := a.height - 1 // Reserve 1 line for footer

	// Render left pane (list) with constrained width, no footer
	leftPane := a.list.ViewConstrained(leftWidth, contentHeight)

	// Render right pane (preview) with same height
	a.preview.width = rightWidth
	a.preview.height = contentHeight
	rightPane := a.preview.View()

	// Compose columns
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// App-global footer spans full width
	footer := a.list.Footer()

	return columns + "\n" + footer
}

// View renders the current view
func (a *App) View() string {
	var base string
	switch a.state {
	case viewList:
		if a.isTwoColumnMode() {
			base = a.renderTwoColumnView()
		} else {
			base = a.list.View()
		}
	case viewDetail:
		base = a.detail.View()
	case viewTagPicker:
		base = a.tagPicker.View()
	case viewParentPicker:
		base = a.parentPicker.ModalView(a.getBackgroundView(), a.width, a.height)
	case viewStatusPicker:
		base = a.statusPicker.ModalView(a.getBackgroundView(), a.width, a.height)
	case viewTypePicker:
		base = a.typePicker.ModalView(a.getBackgroundView(), a.width, a.height)
	case viewPriorityPicker:
		base = a.priorityPicker.ModalView(a.getBackgroundView(), a.width, a.height)
	case viewBlockingPicker:
		base = a.blockingPicker.ModalView(a.getBackgroundView(), a.width, a.height)
	case viewCreateModal:
		base = a.createModal.ModalView(a.getBackgroundView(), a.width, a.height)
	case viewHelpOverlay:
		base = a.helpOverlay.ModalView(a.getBackgroundView(), a.width, a.height)
	case viewAgentChat:
		base = a.agentChat.View()
	case viewInteraction:
		base = a.interaction.ModalView(a.getBackgroundView(), a.width, a.height)
	}
	if base == "" {
		return ""
	}

	// Overlay agent panel on views that aren't the agent chat itself
	if a.state != viewAgentChat && len(a.agentPanel.agents) > 0 {
		base = a.agentPanel.Overlay(base, a.width, a.height)
	}

	return base
}

// getBackgroundView returns the view to show behind modal pickers
func (a *App) getBackgroundView() string {
	switch a.previousState {
	case viewList:
		return a.list.View()
	case viewDetail:
		return a.detail.View()
	default:
		return a.list.View()
	}
}

// getEditor returns the user's preferred editor using the fallback chain:
// $VISUAL -> $EDITOR -> vi -> nano
func getEditor() string {
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	// Fallback chain: vi is more universal, nano as last resort
	if _, err := exec.LookPath("vi"); err == nil {
		return "vi"
	}
	return "nano"
}

// Run starts the TUI application with file watching
func Run(core *beancore.Core, cfg *config.Config, agentMgr *agent.Manager, wtMgr *worktree.Manager) error {
	app := New(core, cfg, agentMgr, wtMgr)
	p := tea.NewProgram(app, tea.WithAltScreen())

	// Store reference to program for sending messages from watcher
	app.program = p

	// Clean up agent manager on exit
	if agentMgr != nil {
		defer agentMgr.Shutdown()
		defer agentMgr.UnsubscribeGlobal(app.globalSubCh)

		// Refresh bean list when an agent completes a turn
		agentMgr.SetOnTurnComplete(func(beanID string) {
			if app.program != nil {
				app.program.Send(beansChangedMsg{})
			}
		})
	}

	// Start file watching
	if err := core.StartWatching(); err != nil {
		return err
	}
	defer core.Unwatch()

	// Subscribe to bean events
	eventCh, unsubscribe := core.Subscribe()
	defer unsubscribe()

	// Forward events to TUI in a goroutine
	go func() {
		for range eventCh {
			// Send message to TUI when beans change
			if app.program != nil {
				app.program.Send(beansChangedMsg{})
			}
		}
	}()

	_, err := p.Run()
	return err
}
