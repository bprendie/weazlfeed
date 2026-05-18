package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderURLModal(bodyHeight int) string {
	outerWidth := clampInt(m.width-8, 32, 88)
	contentWidth := max(20, outerWidth-4)
	input := m.input
	input.Width = contentWidth
	targetSection, targetFolder := m.selectedFolderTarget()
	target := targetSection + "/" + targetFolder
	lines := []string{
		m.styles.status.Render("ADD SOURCE URL"),
		m.styles.help.Render(truncate("target: "+target+" | gopher and audio feeds auto-route", contentWidth)),
		"",
		input.View(),
		"",
		m.styles.help.Render("enter add | esc cancel"),
	}
	modal := m.styles.panel.
		Width(contentWidth).
		BorderForeground(crushPink).
		Padding(1, 2).
		Render(strings.Join(lines, "\n"))
	return lipgloss.Place(max(1, m.width), max(1, bodyHeight), lipgloss.Center, lipgloss.Center, modal)
}

func (m Model) renderHelpModal(bodyHeight int) string {
	outerWidth := clampInt(m.width-8, 36, 92)
	contentWidth := max(24, outerWidth-4)
	contentHeight := max(6, bodyHeight-4)
	lines := strings.Split(m.helpText(), "\n")
	for i := range lines {
		lines[i] = truncate(lines[i], contentWidth)
	}
	content := strings.Join(exactLines(lines, contentHeight), "\n")
	modal := m.styles.panel.
		Width(contentWidth).
		Height(contentHeight).
		BorderForeground(crushPink).
		Padding(1, 2).
		Render(content)
	return lipgloss.Place(max(1, m.width), max(1, bodyHeight), lipgloss.Center, lipgloss.Center, modal)
}

func (m Model) renderAudioModal(bodyHeight int) string {
	outerWidth := clampInt(m.width-8, 42, 96)
	contentWidth := max(28, outerWidth-4)
	title := truncate(firstText(m.playingTitle, "audio"), contentWidth)
	state := "PLAYING"
	if m.paused {
		state = "PAUSED"
	}
	position := audioPosition(m.player.Position(), m.playingTotal)
	lines := []string{
		m.styles.status.Render(state + " " + position),
		m.styles.item.Render(title),
		"",
		m.visualizer.View(),
		"",
		m.styles.help.Render("space pause/resume | < -10s | > +30s | s stop/close | esc close"),
	}
	modal := m.styles.panel.
		Width(contentWidth).
		BorderForeground(crushPink).
		Padding(1, 2).
		Render(strings.Join(lines, "\n"))
	return lipgloss.Place(max(1, m.width), max(1, bodyHeight), lipgloss.Center, lipgloss.Center, modal)
}
