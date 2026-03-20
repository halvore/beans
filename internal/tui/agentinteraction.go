package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hmans/beans/internal/agent"
	"github.com/hmans/beans/internal/ui"
)

// interactionModel handles blocking agent interactions (plan approval, questions)
type interactionModel struct {
	beanID      string
	title       string
	interaction *agent.PendingInteraction
	cursor      int
	viewport    viewport.Model
	width       int
	height      int

	// Signals back to App.Update
	done     bool
	response string
}

func newInteractionModel(beanID, title string, interaction *agent.PendingInteraction, width, height int) interactionModel {
	vp := viewport.New(width*60/100, height/2)

	if interaction.Type == agent.InteractionExitPlan && interaction.PlanContent != "" {
		vp.SetContent(interaction.PlanContent)
	}

	return interactionModel{
		beanID:      beanID,
		title:       title,
		interaction: interaction,
		viewport:    vp,
		width:       width,
		height:      height,
	}
}

func (m interactionModel) Init() tea.Cmd {
	return nil
}

func (m interactionModel) Update(msg tea.Msg) (interactionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Dismiss without responding (agent stays blocked)
			m.done = true
			return m, nil

		case "enter":
			m.done = true
			m.response = m.getResponse()
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			maxIdx := m.maxOptions() - 1
			if m.cursor < maxIdx {
				m.cursor++
			}
		}
	}

	if m.interaction.Type == agent.InteractionExitPlan {
		var vpCmd tea.Cmd
		m.viewport, vpCmd = m.viewport.Update(msg)
		return m, vpCmd
	}

	return m, nil
}

func (m interactionModel) getResponse() string {
	switch m.interaction.Type {
	case agent.InteractionExitPlan:
		if m.cursor == 0 {
			return "yes"
		}
		return "no"
	case agent.InteractionEnterPlan:
		if m.cursor == 0 {
			return "yes"
		}
		return "no"
	case agent.InteractionAskUser:
		if len(m.interaction.Questions) > 0 && len(m.interaction.Questions[0].Options) > 0 {
			opts := m.interaction.Questions[0].Options
			if m.cursor < len(opts) {
				return opts[m.cursor].Label
			}
		}
		return "yes"
	}
	return "yes"
}

func (m interactionModel) maxOptions() int {
	switch m.interaction.Type {
	case agent.InteractionExitPlan, agent.InteractionEnterPlan:
		return 2 // approve / reject
	case agent.InteractionAskUser:
		if len(m.interaction.Questions) > 0 && len(m.interaction.Questions[0].Options) > 0 {
			return len(m.interaction.Questions[0].Options)
		}
		return 2
	}
	return 2
}

func (m interactionModel) View() string {
	modalWidth := max(50, min(70, m.width*60/100))

	// Title
	titleText := "Agent needs input"
	if m.interaction.Type == agent.InteractionExitPlan {
		titleText = "Approve plan?"
	} else if m.interaction.Type == agent.InteractionEnterPlan {
		titleText = "Switch to plan mode?"
	} else if m.interaction.Type == agent.InteractionAskUser {
		titleText = "Agent question"
	}

	header := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorPrimary).Render(titleText)
	subtitle := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(m.title)

	var content strings.Builder
	content.WriteString(header + "\n" + subtitle + "\n\n")

	switch m.interaction.Type {
	case agent.InteractionExitPlan:
		// Show plan content in viewport
		if m.interaction.PlanContent != "" {
			planPreview := m.interaction.PlanContent
			lines := strings.Split(planPreview, "\n")
			if len(lines) > 15 {
				lines = lines[:15]
				lines = append(lines, "...")
			}
			planStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#aaa"))
			content.WriteString(planStyle.Render(strings.Join(lines, "\n")))
			content.WriteString("\n\n")
		}
		m.renderOptions(&content, []string{"Approve", "Reject"})

	case agent.InteractionEnterPlan:
		content.WriteString("The agent wants to switch to plan mode (read-only).\n\n")
		m.renderOptions(&content, []string{"Approve", "Reject"})

	case agent.InteractionAskUser:
		if len(m.interaction.Questions) > 0 {
			q := m.interaction.Questions[0]
			if q.Header != "" {
				content.WriteString(lipgloss.NewStyle().Bold(true).Render(q.Header) + "\n")
			}
			if q.Question != "" {
				content.WriteString(q.Question + "\n\n")
			}
			if len(q.Options) > 0 {
				labels := make([]string, len(q.Options))
				for i, opt := range q.Options {
					labels[i] = opt.Label
				}
				m.renderOptions(&content, labels)
			} else {
				m.renderOptions(&content, []string{"Yes", "No"})
			}
		}
	}

	// Footer
	footer := helpKeyStyle.Render("enter") + " " + helpStyle.Render("confirm") + "  " +
		helpKeyStyle.Render("j/k") + " " + helpStyle.Render("navigate") + "  " +
		helpKeyStyle.Render("esc") + " " + helpStyle.Render("dismiss")

	content.WriteString("\n" + footer)

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorPrimary).
		Padding(1, 2).
		Width(modalWidth)

	return border.Render(content.String())
}

func (m interactionModel) renderOptions(b *strings.Builder, options []string) {
	for i, opt := range options {
		cursor := "  "
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#fff"))
		if i == m.cursor {
			cursor = "> "
			style = style.Foreground(ui.ColorPrimary).Bold(true)
		}
		fmt.Fprintf(b, "%s%s\n", cursor, style.Render(opt))
	}
}

// ModalView renders the interaction modal over a background
func (m interactionModel) ModalView(bgView string, fullWidth, fullHeight int) string {
	modal := m.View()
	return overlayModal(bgView, modal, fullWidth, fullHeight)
}
