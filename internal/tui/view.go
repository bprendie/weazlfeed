package tui

import (
	"fmt"
	"strings"

	"github.com/bprendie/weazlfeed/internal/store"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width == 0 {
		return "weazlfeed"
	}
	contentWidth := max(40, m.width-4)
	header := "\n" + renderLogo(logo, contentWidth)
	bodyHeight := max(8, m.height-lipgloss.Height(header)-6)
	leftW := max(24, m.width/5)
	centerW := max(34, m.width/3)
	rightW := max(30, m.width-leftW-centerW-10)
	body := lipgloss.JoinHorizontal(lipgloss.Top,
		m.panel("INDEX", m.renderFeeds(leftW, bodyHeight), leftW, bodyHeight, m.focus == focusFeeds),
		m.panel("STREAM", m.renderItems(centerW, bodyHeight), centerW, bodyHeight, m.focus == focusItems),
		m.panel("STAGE", m.renderStage(rightW, bodyHeight), rightW, bodyHeight, m.focus == focusArticle),
	)
	footer := m.footer()
	return m.styles.frame.Width(m.width).Height(m.height).Render(lipgloss.JoinVertical(lipgloss.Left, header, body, footer))
}

func (m Model) panel(title, body string, width, height int, active bool) string {
	style := m.styles.panel.Width(width).Height(height)
	if active {
		style = style.BorderForeground(crushPink)
	}
	return style.Render(m.styles.status.Render(title) + "\n" + body)
}

func (m Model) renderFeeds(width, height int) string {
	if len(m.feeds) == 0 {
		return m.styles.help.Render("No feeds yet. Add feeds in config, then run setup or refresh.")
	}
	var lines []string
	currentCategory := ""
	for i, feed := range m.feeds {
		category := strings.ToUpper(firstText(feed.Category, "GENERAL"))
		if category != currentCategory {
			if len(lines) > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, m.styles.help.Render(":: "+category+" ::"))
			currentCategory = category
		}
		prefix := "  "
		if feed.Type == "gopher" {
			prefix = "g>"
		}
		line := truncate(fmt.Sprintf("%s %s [%d]", prefix, feed.Title, feed.Unread), width-4)
		if i == m.feedCursor {
			line = m.styles.selected.Render("=> " + strings.TrimSpace(line))
		} else {
			line = m.styles.item.Render("   " + line)
		}
		lines = append(lines, line)
	}
	return fitLines(lines, height-2)
}

func (m Model) renderItems(width, height int) string {
	if len(m.items) == 0 {
		return m.styles.help.Render("No items loaded.")
	}
	var lines []string
	lines = append(lines, m.styles.help.Render("last signal / newest first"))
	for i, item := range m.items {
		badges := badges(item)
		line := truncate(badges+" "+item.Title, width-4)
		if i == m.itemCursor {
			line = m.styles.selected.Render("=> " + line)
		} else {
			line = m.styles.item.Render(" - " + line)
		}
		lines = append(lines, line)
	}
	return fitLines(lines, height-2)
}

func (m Model) renderStage(width, height int) string {
	if m.asking {
		return m.input.View()
	}
	lines := strings.Split(m.article, "\n")
	for i := range lines {
		lines[i] = truncate(lines[i], width-2)
	}
	return fitLines(lines, height-2)
}

func (m Model) footer() string {
	ai := "ai off"
	if m.aiEnabled {
		ai = "ai on"
	}
	audioState := "audio idle"
	if m.player.Active() {
		audioState = "audio live"
	}
	parts := []string{
		m.styles.help.Render("[j/k] nav  [tab] node  [enter] read/dial/play  [space] pause  [s] stop  [r] refresh  [h] sludge  [q] quit"),
		m.styles.status.Render(ai + " | " + audioState),
		m.visualizer(),
	}
	if m.err != "" {
		parts = append(parts, m.styles.error.Render(m.err))
	} else {
		parts = append(parts, m.styles.help.Render(m.status))
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) visualizer() string {
	if len(m.bars) == 0 {
		return ""
	}
	blocks := []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	var b strings.Builder
	for _, value := range m.bars {
		idx := int(value * float64(len(blocks)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		b.WriteString(blocks[idx])
	}
	return m.styles.status.Render(b.String())
}

func badges(item store.Item) string {
	var parts []string
	if !item.ReadStatus {
		parts = append(parts, "[UNREAD]")
	}
	if item.EnclosureURL != "" && strings.HasPrefix(item.EnclosureType, "audio/") {
		if item.PlayheadSeconds > 0 {
			parts = append(parts, fmt.Sprintf("[AUDIO %s]", formatClock(item.PlayheadSeconds)))
		} else {
			parts = append(parts, "[AUDIO]")
		}
	}
	if item.SludgeFlag {
		parts = append(parts, "[SLUDGE]")
	}
	if strings.HasPrefix(strings.ToLower(item.Link), "gopher://") {
		parts = append(parts, "[GOPHER]")
	}
	return strings.Join(parts, " ")
}

func fitLines(lines []string, height int) string {
	if height < 1 {
		return ""
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}

func truncate(value string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width <= 1 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "…"
}

func formatClock(seconds int) string {
	min := seconds / 60
	sec := seconds % 60
	if min >= 60 {
		return fmt.Sprintf("%d:%02d:%02d", min/60, min%60, sec)
	}
	return fmt.Sprintf("%d:%02d", min, sec)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
