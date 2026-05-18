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
	var lines []string
	seenFolders := map[string]bool{}
	for i, feed := range m.visibleFeeds() {
		section := firstText(feed.Section, sectionFromFeed(feed))
		folder := firstText(feed.Folder, folderFromFeed(feed))
		seenFolders[section+"/"+folder] = true
		if section != previousSection(m.feeds, m.feedScroll+i) {
			lines = append(lines, m.styles.status.Render(":: "+section+" ::"))
		}
		if folder != previousFolder(m.feeds, m.feedScroll+i) {
			lines = append(lines, m.styles.help.Render("  ["+folder+"]"))
		}
		prefix := "  "
		if feed.Type == "gopher" {
			prefix = "g>"
		}
		feedIndex := m.feedScroll + i
		var line string
		if feedIndex == m.feedCursor {
			line = truncate("=> "+prefix+" "+feed.Title+" ["+intText(feed.Unread)+"]", width)
			line = m.styles.selected.Render(line)
		} else {
			line = truncate(" - "+prefix+" "+feed.Title+" ["+intText(feed.Unread)+"]", width)
			line = m.styles.item.Render(line)
		}
		lines = append(lines, line)
	}
	for _, folder := range m.folders {
		key := folder.Section + "/" + folder.Name
		if seenFolders[key] {
			continue
		}
		lines = append(lines, m.styles.status.Render(":: "+folder.Section+" ::"))
		lines = append(lines, m.styles.help.Render("  ["+folder.Name+"]"))
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
	if m.asking || m.folderInput || m.podcastInput {
		return truncate(m.input.View(), width)
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
	picked := ""
	if m.pickedFeedID != 0 {
		picked = " | picked source"
	}
	parts := []string{
		m.styles.help.Render(truncate("[j/k] nav [pg] scroll [tab] node [enter] open [space] pick/drop [n] folder [p] podcast [r/R] refresh [q] quit", max(10, m.width))),
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

func previousSection(feeds []store.Feed, index int) string {
	if index <= 0 || index > len(feeds)-1 {
		return ""
	}
	return firstText(feeds[index-1].Section, sectionFromFeed(feeds[index-1]))
}

func previousFolder(feeds []store.Feed, index int) string {
	if index <= 0 || index > len(feeds)-1 {
		return ""
	}
	prev := feeds[index-1]
	current := feeds[index]
	if previousSection(feeds, index) != firstText(current.Section, sectionFromFeed(current)) {
		return ""
	}
	return firstText(prev.Folder, folderFromFeed(prev))
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
