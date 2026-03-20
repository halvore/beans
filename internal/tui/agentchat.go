package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hmans/beans/internal/agent"
	"github.com/hmans/beans/internal/ui"
)

// agentChatModel is the full-screen agent chat view
type agentChatModel struct {
	beanID   string
	title    string
	session  *agent.Session
	viewport viewport.Model
	input    textinput.Model
	width    int
	height   int

	// Signals back to App.Update
	done       bool   // wants to close
	pendingMsg string // message to send
	wantStop   bool   // wants to stop agent
}

func newAgentChatModel(beanID, title string, session *agent.Session, width, height int) agentChatModel {
	ti := textinput.New()
	ti.Placeholder = "Type a message..."
	ti.Focus()
	ti.CharLimit = 4096
	ti.Width = width - 4

	vp := viewport.New(width, height-5) // header(2) + input(1) + footer(1) + border
	vp.SetContent(renderMessages(session, width-4))

	// Auto-scroll to bottom
	vp.GotoBottom()

	return agentChatModel{
		beanID:   beanID,
		title:    title,
		session:  session,
		viewport: vp,
		input:    ti,
		width:    width,
		height:   height,
	}
}

func (m agentChatModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m agentChatModel) Update(msg tea.Msg) (agentChatModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 5
		m.input.Width = msg.Width - 4
		m.viewport.SetContent(renderMessages(m.session, m.width-4))

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.done = true
			return m, nil
		case "enter":
			text := strings.TrimSpace(m.input.Value())
			if text != "" {
				m.pendingMsg = text
				m.input.Reset()
			}
			return m, nil
		case "ctrl+x":
			m.wantStop = true
			return m, nil
		}
	}

	// Update text input
	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)

	// Update viewport
	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

func (m agentChatModel) updateSession(session *agent.Session) agentChatModel {
	m.session = session
	wasAtBottom := m.viewport.AtBottom()
	m.viewport.SetContent(renderMessages(session, m.width-4))
	if wasAtBottom {
		m.viewport.GotoBottom()
	}
	return m
}

func (m agentChatModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	statusBadge := "idle"
	statusColor := ui.ColorMuted
	if m.session != nil {
		switch m.session.Status {
		case agent.StatusRunning:
			statusBadge = "running"
			statusColor = lipgloss.Color("#0f0")
		case agent.StatusError:
			statusBadge = "error"
			statusColor = lipgloss.Color("#f00")
		}
		if m.session.PendingInteraction != nil {
			statusBadge = "needs input"
			statusColor = lipgloss.Color("#ff0")
		}
	}

	titleStr := m.title
	if len(titleStr) > m.width-30 {
		titleStr = titleStr[:m.width-33] + "..."
	}
	header := detailTitleStyle.Render(fmt.Sprintf(" Agent: %s ", titleStr)) +
		" " + lipgloss.NewStyle().Foreground(statusColor).Render("["+statusBadge+"]")
	header = lipgloss.NewStyle().Width(m.width).Render(header)

	// Input line
	inputView := m.input.View()

	// Footer
	footer := helpKeyStyle.Render("enter") + " " + helpStyle.Render("send") + "  " +
		helpKeyStyle.Render("ctrl+x") + " " + helpStyle.Render("stop") + "  " +
		helpKeyStyle.Render("ctrl+a/esc") + " " + helpStyle.Render("back")

	return header + "\n" + m.viewport.View() + "\n" + inputView + "\n" + footer
}

// renderMessages converts agent session messages into a styled string
func renderMessages(session *agent.Session, width int) string {
	if session == nil || len(session.Messages) == 0 {
		return lipgloss.NewStyle().
			Foreground(ui.ColorMuted).
			Render("No messages yet. Start a conversation or press esc to go back.")
	}

	var b strings.Builder
	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7af")).Bold(true)
	assistantStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#fff"))
	toolStyle := lipgloss.NewStyle().Foreground(ui.ColorMuted).Italic(true)
	infoStyle := lipgloss.NewStyle().Foreground(ui.ColorMuted).Italic(true)

	for _, msg := range session.Messages {
		switch msg.Role {
		case agent.RoleUser:
			b.WriteString(userStyle.Render("> " + truncateContent(msg.Content, width-2, 5)))
			b.WriteString("\n\n")

		case agent.RoleAssistant:
			// Try glamour rendering, fall back to plain text
			rendered := msg.Content
			if renderer := getGlamourRenderer(); renderer != nil {
				if out, err := renderer.Render(msg.Content); err == nil {
					rendered = strings.TrimSpace(out)
				}
			}
			b.WriteString(assistantStyle.Render(rendered))
			b.WriteString("\n\n")

		case agent.RoleTool:
			content := msg.Content
			if msg.Diff != "" {
				content += " (changed)"
			}
			b.WriteString(toolStyle.Render("  " + content))
			b.WriteString("\n")

		case agent.RoleInfo:
			b.WriteString(infoStyle.Render("  " + msg.Content))
			b.WriteString("\n")
		}
	}

	// Show tool invocations for current turn
	if session.Status == agent.StatusRunning && len(session.ToolInvocations) > 0 {
		last := session.ToolInvocations[len(session.ToolInvocations)-1]
		b.WriteString(toolStyle.Render(fmt.Sprintf("  [%s] %s...", last.Tool, last.Input)))
		b.WriteString("\n")
	}

	return b.String()
}

func truncateContent(s string, maxWidth, maxLines int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, "...")
	}
	return strings.Join(lines, "\n")
}
