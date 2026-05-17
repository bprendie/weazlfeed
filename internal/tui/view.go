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
	dims, bodyHeight := m.layout()
	contentWidth := max(20, m.width-4)
	header := renderLogo(logo, contentWidth)
	body := lipgloss.JoinHorizontal(lipgloss.Top,
		m.panel("INDEX", m.renderFeeds(dims.left, bodyHeight), dims.left, bodyHeight, m.focus == focusFeeds),
		m.panel("STREAM", m.renderItems(dims.center, bodyHeight), dims.center, bodyHeight, m.focus == focusItems),
		m.panel("STAGE", m.renderStage(dims.right, bodyHeight), dims.right, bodyHeight, m.focus == focusArticle),
	)
	footer := m.footer()
	return m.styles.frame.Width(m.width).Height(m.height).Render(lipgloss.JoinVertical(lipgloss.Left, header, body, footer))
}

func (m Model) panel(title, body string, width, height int, active bool) string {
	style := m.styles.panel.Width(width).Height(height)
	if active {
		style = style.BorderForeground(crushPink)
	}
	lines := strings.Split(body, "\n")
	contentHeight := max(1, height-3)
	return style.Render(m.styles.status.Render(title) + "\n" + fitLines(lines, contentHeight))
}

func (m Model) renderFeeds(width, height int) string {
	if len(m.feeds) == 0 {
		return m.styles.help.Render("No feeds yet. Add feeds in config, then run setup or refresh.")
	}
	var lines []string
	for i, feed := range m.visibleFeeds() {
		category := strings.ToUpper(firstText(feed.Category, "GENERAL"))
		currentCategory := previousCategory(m.feeds, m.feedScroll+i)
		if category != currentCategory {
			lines = append(lines, m.styles.help.Render(":: "+category+" ::"))
			currentCategory = category
		}
		prefix := "  "
		if feed.Type == "gopher" {
			prefix = "g>"
		}
		line := truncate(fmt.Sprintf("%s %s [%d]", prefix, feed.Title, feed.Unread), width-4)
		feedIndex := m.feedScroll + i
		if feedIndex == m.feedCursor {
			line = m.styles.selected.Render("=> " + strings.TrimSpace(line))
		} else {
			line = m.styles.item.Render("   " + line)
		}
		lines = append(lines, line)
	}
	return fitLines(lines, height-3)
}

func (m Model) renderItems(width, height int) string {
	if len(m.items) == 0 {
		return m.styles.help.Render("No items loaded.")
	}
	var lines []string
	lines = append(lines, m.styles.help.Render("last signal / newest first"))
	for i, item := range m.visibleItems() {
		badges := badges(item)
		line := truncate(badges+" "+item.Title, width-4)
		itemIndex := m.itemScroll + i
		if itemIndex == m.itemCursor {
			line = m.styles.selected.Render("=> " + line)
		} else {
			line = m.styles.item.Render(" - " + line)
		}
		lines = append(lines, line)
	}
	return fitLines(lines, height-3)
}

func (m Model) renderStage(width, height int) string {
	if m.asking {
		return m.input.View()
	}
	lines := strings.Split(m.article, "\n")
	lines = windowLines(lines, m.stageScroll, height-3)
	for i := range lines {
		lines[i] = truncate(lines[i], width-2)
	}
	return fitLines(lines, height-3)
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
		m.styles.help.Render("[j/k] nav  [pg] scroll  [tab] node  [enter] read/dial/play  [space] pause  [s] stop  [r] refresh  [q] quit"),
		m.styles.status.Render(ai + " | " + audioState + compactVisualizer(m.visualizer())),
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

type paneDims struct {
	left   int
	center int
	right  int
}

func (m Model) layout() (paneDims, int) {
	contentWidth := max(40, m.width-4)
	headerHeight := lipgloss.Height(renderLogo(logo, contentWidth))
	bodyHeight := clampInt(m.height-headerHeight-5, 6, max(6, m.height-3))
	innerWidth := max(40, m.width-8)
	left := clampInt(innerWidth/5, 18, 28)
	center := clampInt(innerWidth/3, 28, 44)
	right := max(24, innerWidth-left-center)
	return paneDims{left: left, center: center, right: right}, bodyHeight
}

func compactVisualizer(value string) string {
	if value == "" {
		return ""
	}
	return " | " + value
}

func (m Model) visibleFeeds() []store.Feed {
	_, bodyHeight := m.layout()
	return windowFeeds(m.feeds, m.feedScroll, max(1, bodyHeight-3))
}

func (m Model) visibleItems() []store.Item {
	_, bodyHeight := m.layout()
	return windowItems(m.items, m.itemScroll, max(1, bodyHeight-4))
}

func windowFeeds(feeds []store.Feed, start, count int) []store.Feed {
	start = clampInt(start, 0, len(feeds))
	end := clampInt(start+count, start, len(feeds))
	return feeds[start:end]
}

func windowItems(items []store.Item, start, count int) []store.Item {
	start = clampInt(start, 0, len(items))
	end := clampInt(start+count, start, len(items))
	return items[start:end]
}

func windowLines(lines []string, start, count int) []string {
	start = clampInt(start, 0, len(lines))
	end := clampInt(start+count, start, len(lines))
	return lines[start:end]
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

func clampInt(value, low, high int) int {
	if high < low {
		return low
	}
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func previousCategory(feeds []store.Feed, index int) string {
	if index <= 0 || index > len(feeds)-1 {
		return ""
	}
	return strings.ToUpper(firstText(feeds[index-1].Category, "GENERAL"))
}
