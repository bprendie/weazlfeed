package tui

import (
	"fmt"
	"strings"

	"github.com/bprendie/weazlfeed/internal/store"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m Model) View() string {
	if m.width == 0 {
		return "weazlfeed"
	}
	if m.lockMode != lockOpen {
		return m.lockView()
	}
	dims, bodyHeight := m.layout()
	contentWidth := max(20, m.width)
	header := m.header(contentWidth)
	body := lipgloss.JoinHorizontal(lipgloss.Top,
		m.panel("SOURCES", m.renderFeeds(dims.left, bodyHeight), dims.left, bodyHeight, m.focus == focusFeeds),
		m.panel("ITEMS", m.renderItems(dims.center, bodyHeight), dims.center, bodyHeight, m.focus == focusItems),
		m.panel("READER", m.renderStage(dims.right, bodyHeight), dims.right, bodyHeight, m.focus == focusArticle),
	)
	footer := m.footer()
	return m.styles.frame.Width(m.width).Render(lipgloss.JoinVertical(lipgloss.Left, header, body, footer))
}

func (m Model) panel(title, body string, width, height int, active bool) string {
	style := m.styles.panel.Width(panelContentWidth(m.styles.panel, width))
	if active {
		style = style.BorderForeground(crushPink)
	}
	lines := strings.Split(body, "\n")
	contentHeight := max(1, height-3)
	content := exactLines(append([]string{m.styles.status.Render(title)}, lines...), contentHeight+1)
	return style.Render(strings.Join(content, "\n"))
}

func (m Model) header(width int) string {
	if width < maxLineWidth(logo) || m.height < 18 {
		return gradientLogo("////// WeazlFeed //////")
	}
	return renderLogo(logo, width)
}

func (m Model) renderFeeds(width, height int) string {
	width = panelContentWidth(m.styles.panel, width)
	if len(m.feeds) == 0 {
		return m.styles.help.Render(truncate("No feeds yet. Add feeds in config, then run setup or refresh.", width))
	}
	rows := m.sourceRows()
	rows = windowSourceRows(rows, m.feedScroll, max(1, height-3))
	lines := make([]string, 0, len(rows))
	for i, row := range rows {
		rowIndex := m.feedScroll + i
		switch row.kind {
		case sourceSection:
			lines = append(lines, m.styles.status.Render(truncate(":: "+row.title+" ::", width)))
		case sourceFolder:
			marker := "[-]"
			if row.collapsed {
				marker = "[+]"
			}
			line := truncate("  "+marker+" "+row.title, width)
			if rowIndex == m.sourceCursor {
				line = m.styles.selected.Render(truncate("=> "+marker+" "+row.title, width))
			} else {
				line = m.styles.help.Render(line)
			}
			lines = append(lines, line)
		case sourceFeed:
			prefix := "  "
			if m.feeds[row.feedIndex].Type == "gopher" {
				prefix = "g>"
			}
			line := truncate(" - "+prefix+" "+row.title+" ["+intText(row.unread)+"]", width)
			if rowIndex == m.sourceCursor {
				line = m.styles.selected.Render(truncate("=> "+prefix+" "+row.title+" ["+intText(row.unread)+"]", width))
			} else {
				line = m.styles.item.Render(line)
			}
			lines = append(lines, line)
		}
	}
	return fitLines(lines, height-3)
}

func (m Model) renderItems(width, height int) string {
	if m.podcastMode() {
		return m.renderPodcastItems(width, height)
	}
	width = panelContentWidth(m.styles.panel, width)
	if len(m.items) == 0 {
		return m.styles.help.Render(truncate("No items loaded.", width))
	}
	var lines []string
	lines = append(lines, m.styles.help.Render("last signal / newest first"))
	for i, item := range m.visibleItems() {
		badges := badges(item)
		itemIndex := m.itemScroll + i
		var line string
		if itemIndex == m.itemCursor {
			line = truncate("=> "+badges+" "+item.Title, width)
			line = m.styles.selected.Render(line)
		} else {
			line = truncate(" - "+badges+" "+item.Title, width)
			line = m.styles.item.Render(line)
		}
		lines = append(lines, line)
	}
	return fitLines(lines, height-3)
}

func (m Model) renderStage(width, height int) string {
	width = panelContentWidth(m.styles.panel, width)
	if m.asking || m.folderInput || m.podcastInput || m.urlInput {
		return truncate(m.input.View(), width)
	}
	if m.rendering {
		return fitLines([]string{m.styles.status.Render(m.spinner.View() + " rendering reader")}, height-3)
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
	if m.refreshing {
		audioState = m.spinner.View() + " refreshing"
	}
	if m.rendering {
		audioState = m.spinner.View() + " rendering"
	}
	picked := ""
	if m.pickedFeedID != 0 {
		picked = " | picked source"
	}
	parts := []string{
		m.styles.help.Render(truncate("[j/k] nav [pg] scroll [enter] open/fold [esc/left] back [right] expand [space] pick/drop [a] add url [n] folder [p] podcast [r/R] refresh [q] quit", max(10, m.width))),
		m.styles.status.Render(ai + " | " + audioState + picked + compactVisualizer(m.visualizer())),
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
		switch firstText(item.EnclosureType, gopherEnclosureType(item.Link)) {
		case "gopher/directory":
			parts = append(parts, "[DIR]")
		case "gopher/search":
			parts = append(parts, "[SEARCH]")
		case "text/plain":
			parts = append(parts, "[TXT]")
		default:
			parts = append(parts, "[GOPHER]")
		}
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

func exactLines(lines []string, height int) []string {
	if height < 1 {
		return nil
	}
	if len(lines) > height {
		return lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return lines
}

type paneDims struct {
	left   int
	center int
	right  int
}

func (m Model) layout() (paneDims, int) {
	contentWidth := max(40, m.width)
	headerHeight := lipgloss.Height(m.header(contentWidth))
	footerHeight := 3
	bodyHeight := clampInt(m.height-headerHeight-footerHeight-2, 5, max(5, m.height-2))
	left, center, right := m.layoutWidths(max(30, m.width))
	return paneDims{left: left, center: center, right: right}, bodyHeight
}

func (m Model) layoutWidths(total int) (left, center, right int) {
	compact := clampInt(total/5, 12, 22)
	focused := max(12, total-(compact*2))
	switch m.focus {
	case focusFeeds:
		left, center, right = focused, compact, total-focused-compact
	case focusItems:
		center, left, right = focused, compact, total-focused-compact
	default:
		right, left, center = focused, compact, total-focused-compact
	}
	return max(8, left), max(8, center), max(8, right)
}

func panelContentWidth(style lipgloss.Style, outerW int) int {
	return max(1, outerW-style.GetHorizontalFrameSize())
}

func compactVisualizer(value string) string {
	if value == "" {
		return ""
	}
	return " | " + value
}

func (m Model) stageLineCount() int {
	if m.article == "" {
		return 1
	}
	return len(strings.Split(m.article, "\n"))
}

func (m Model) visibleItems() []store.Item {
	_, bodyHeight := m.layout()
	return windowItems(m.items, m.itemScroll, max(1, bodyHeight-4))
}

func windowSourceRows(rows []sourceRow, start, count int) []sourceRow {
	start = clampInt(start, 0, len(rows))
	end := clampInt(start+count, start, len(rows))
	return rows[start:end]
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
	if ansi.StringWidth(value) <= width {
		return value
	}
	return ansi.Truncate(value, width, "…")
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

func sectionFromFeed(feed store.Feed) string {
	if feed.Type == "gopher" {
		return "Gopher"
	}
	return "News"
}

func folderFromFeed(feed store.Feed) string {
	if feed.Folder != "" {
		return feed.Folder
	}
	return titleCase(firstText(feed.Category, "General"))
}

func titleCase(value string) string {
	value = strings.ToLower(value)
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
