package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hmans/beans/internal/agent"
	"github.com/hmans/beans/internal/graph"
	"github.com/hmans/beans/internal/ui"
)

// agentInfo holds display info about an active agent
type agentInfo struct {
	beanID             string
	title              string
	status             agent.SessionStatus
	pendingInteraction *agent.PendingInteraction
}

// agentPanelModel shows a compact agent status panel
type agentPanelModel struct {
	agents []agentInfo
}

// refresh rebuilds the agent panel from current manager state
func (m agentPanelModel) refresh(mgr *agent.Manager, resolver *graph.Resolver) agentPanelModel {
	sessions := mgr.ListActiveSessions()
	agents := make([]agentInfo, 0, len(sessions))

	for _, s := range sessions {
		title := s.BeanID
		if b, err := resolver.Query().Bean(context.Background(), s.BeanID); err == nil && b != nil {
			title = b.Title
		}
		session := mgr.GetSession(s.BeanID)
		var pending *agent.PendingInteraction
		if session != nil {
			pending = session.PendingInteraction
		}
		agents = append(agents, agentInfo{
			beanID:             s.BeanID,
			title:              title,
			status:             s.Status,
			pendingInteraction: pending,
		})
	}

	// Also include idle sessions with pending interactions
	// (they show as idle but need attention)
	return agentPanelModel{agents: agents}
}

// View renders the compact panel content
func (m agentPanelModel) View() string {
	if len(m.agents) == 0 {
		return ""
	}

	var b strings.Builder

	// Header — show running count vs total
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorPrimary)
	running := 0
	for _, a := range m.agents {
		if a.status == agent.StatusRunning {
			running++
		}
	}
	headerText := fmt.Sprintf("Agents: %d", len(m.agents))
	if running > 0 {
		headerText = fmt.Sprintf("Agents: %d (%d running)", len(m.agents), running)
	}
	b.WriteString(headerStyle.Render(headerText))
	b.WriteString("\n")

	// Per-agent lines (cap at 6 to avoid taking too much space)
	maxLines := 6
	if len(m.agents) < maxLines {
		maxLines = len(m.agents)
	}
	for i := 0; i < maxLines; i++ {
		a := m.agents[i]
		icon := statusIcon(a.status, a.pendingInteraction)
		title := a.title
		if len(title) > 25 {
			title = title[:22] + "..."
		}
		b.WriteString(fmt.Sprintf("%s %s\n", icon, title))
	}
	if len(m.agents) > maxLines {
		b.WriteString(fmt.Sprintf("  +%d more\n", len(m.agents)-maxLines))
	}

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorPrimary).
		Padding(0, 1)

	return border.Render(b.String())
}

// Overlay composites the panel onto the bottom-right of the rendered output
func (m agentPanelModel) Overlay(base string, width, height int) string {
	if len(m.agents) == 0 {
		return base
	}

	panel := m.View()
	panelLines := strings.Split(panel, "\n")
	panelHeight := len(panelLines)
	panelWidth := lipgloss.Width(panel)

	bgLines := strings.Split(base, "\n")
	for len(bgLines) < height {
		bgLines = append(bgLines, "")
	}
	if len(bgLines) > height {
		bgLines = bgLines[:height]
	}

	// Position: bottom-right with 1 char margin
	startY := height - panelHeight - 1
	startX := width - panelWidth - 1
	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}

	for i, panelLine := range panelLines {
		bgY := startY + i
		if bgY >= 0 && bgY < len(bgLines) {
			bgLines[bgY] = overlayLine(bgLines[bgY], panelLine, startX, width)
		}
	}

	return strings.Join(bgLines, "\n")
}

func statusIcon(status agent.SessionStatus, pending *agent.PendingInteraction) string {
	if pending != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0")).Render("!")
	}
	switch status {
	case agent.StatusRunning:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#0f0")).Render("~")
	case agent.StatusError:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f00")).Render("x")
	default:
		return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(".")
	}
}
