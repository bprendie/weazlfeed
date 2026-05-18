package tui

import (
	"strconv"
	"strings"

	"github.com/bprendie/weazlfeed/internal/store"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) pickOrDropFeed() (tea.Model, tea.Cmd) {
	row, ok := m.selectedSourceRow()
	if !ok || (m.pickedFeedID == 0 && row.kind != sourceFeed) {
		m.status = "space picks up sources only; press enter to fold folders"
		return m, nil
	}
	section, folder := m.selectedFolderTarget()
	if m.pickedFeedID == 0 {
		feed := m.feeds[row.feedIndex]
		m.pickedFeedID = feed.ID
		m.status = "picked source: " + feed.Title
		return m, nil
	}
	picked := m.findFeed(m.pickedFeedID)
	if picked == nil {
		m.pickedFeedID = 0
		m.status = "picked source vanished"
		return m, nil
	}
	if err := m.store.MoveFeed(picked.ID, section, folder); err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.status = "moved " + picked.Title + " to " + section + "/" + folder
	m.pickedFeedID = 0
	return m, loadFeedsCmd(m.store)
}

func (m Model) createFolder(name string) (tea.Model, tea.Cmd) {
	name = strings.TrimSpace(name)
	if name == "" || len(m.feeds) == 0 {
		return m, nil
	}
	section := "News"
	if row, ok := m.selectedSourceRow(); ok {
		section = row.section
		if row.kind == sourceFeed {
			feed := m.feeds[row.feedIndex]
			section = firstText(feed.Section, sectionFromFeed(feed))
		}
	}
	if err := m.store.UpsertFolder(section, name); err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.status = "folder ready: " + section + "/" + name + " (space drops a picked source)"
	if m.pickedFeedID != 0 {
		picked := m.findFeed(m.pickedFeedID)
		if picked != nil {
			if err := m.store.MoveFeed(picked.ID, section, name); err != nil {
				m.err = err.Error()
				return m, nil
			}
			m.pickedFeedID = 0
			m.status = "moved " + picked.Title + " to " + section + "/" + name
			return m, loadFeedsCmd(m.store)
		}
	}
	return m, nil
}

func (m Model) findFeed(id int64) *store.Feed {
	for i := range m.feeds {
		if m.feeds[i].ID == id {
			return &m.feeds[i]
		}
	}
	return nil
}

func (m *Model) page(delta int) {
	_, bodyHeight := m.layout()
	step := max(1, bodyHeight-4)
	switch m.focus {
	case focusFeeds:
		m.move(delta * step)
		return
	case focusItems:
		m.itemCursor += delta * step
	case focusArticle:
		m.stageScroll += delta * step
	}
	m.clamp()
	m.ensureCursorVisible()
	if m.focus == focusArticle {
		m.clampScrolls()
	}
}

func (m *Model) home() {
	switch m.focus {
	case focusFeeds:
		m.sourceCursor, m.feedScroll = 0, 0
		for _, index := range m.selectableSourceRows() {
			m.sourceCursor = index
			break
		}
		m.syncFeedCursorFromSource()
	case focusItems:
		m.itemCursor, m.itemScroll = 0, 0
	case focusArticle:
		m.stageScroll = 0
	}
}

func (m *Model) end() {
	switch m.focus {
	case focusFeeds:
		indices := m.selectableSourceRows()
		if len(indices) > 0 {
			m.sourceCursor = indices[len(indices)-1]
			m.syncFeedCursorFromSource()
		}
	case focusItems:
		m.itemCursor = m.itemTargetCount() - 1
	case focusArticle:
		m.stageScroll = m.stageLineCount()
	}
	m.clamp()
	m.ensureCursorVisible()
	if m.focus == focusArticle {
		m.clampScrolls()
	}
}

func (m *Model) updateMouse(msg tea.MouseMsg) {
	switch msg.Type {
	case tea.MouseWheelDown:
		m.scrollFocused(3)
	case tea.MouseWheelUp:
		m.scrollFocused(-3)
	}
}

func (m *Model) scrollFocused(delta int) {
	switch m.focus {
	case focusFeeds:
		m.move(delta)
		return
	case focusItems:
		m.itemCursor += delta
	case focusArticle:
		m.stageScroll += delta
	}
	m.clamp()
	m.ensureCursorVisible()
	if m.focus == focusArticle {
		m.clampScrolls()
	}
}

func (m Model) activate() (tea.Model, tea.Cmd) {
	if m.focus == focusFeeds && len(m.feeds) > 0 {
		row, ok := m.selectedSourceRow()
		if !ok {
			return m, nil
		}
		if row.kind == sourceFolder {
			return m.toggleCurrentFolder(!row.collapsed)
		}
		if row.kind == sourceInterrogation {
			if row.aiIndex < 0 || row.aiIndex >= len(m.interrogations) {
				return m, nil
			}
			m.focus = focusArticle
			m.stageScroll = 0
			m.showInterrogation(m.interrogations[row.aiIndex])
			return m, nil
		}
		if row.kind != sourceFeed {
			return m, nil
		}
		feed := m.feeds[row.feedIndex]
		m.feedCursor = row.feedIndex
		m.podcasts = nil
		m.gopherStack = nil
		m.focus = focusItems
		m.status = "opened source: " + feed.Title
		if m.useCachedItems(feed.ID) {
			return m, nil
		}
		m.itemCursor = 0
		m.itemScroll = 0
		m.stageScroll = 0
		return m, loadItemsCmd(m.store, feed.ID, m.hideSludge)
	}
	if len(m.items) == 0 {
		return m, nil
	}
	item := m.items[m.itemCursor]
	if strings.HasPrefix(strings.ToLower(item.Link), "gopher://") {
		m.focus = focusArticle
		m.rendering = true
		m.clearArticle()
		m.status = "dialing gopher target"
		m.gopherStack = append(m.gopherStack, append([]store.Item(nil), m.items...))
		return m, tea.Batch(gopherPageCmd(item.Link), m.spinner.Tick)
	}
	if item.EnclosureURL != "" && strings.HasPrefix(item.EnclosureType, "audio/") {
		return m.playAudio(item)
	}
	if item.ID != 0 {
		_ = m.store.MarkRead(item.ID)
	}
	m.items[m.itemCursor].ReadStatus = true
	if item.FeedID != 0 {
		m.itemCache[item.FeedID] = m.items
	}
	m.focus = focusArticle
	m.stageScroll = 0
	m.rendering = true
	m.clearArticle()
	m.status = m.spinner.View() + " rendering reader"
	return m, tea.Batch(renderReaderCmd(m.store, item, m.readerWidth()), m.spinner.Tick)
}

func (m *Model) clamp() {
	rows := m.sourceRows()
	if m.sourceCursor < 0 {
		m.sourceCursor = 0
	}
	if m.sourceCursor >= len(rows) && len(rows) > 0 {
		m.sourceCursor = len(rows) - 1
	}
	if len(rows) > 0 && rows[m.sourceCursor].kind == sourceSection {
		indices := m.selectableSourceRows()
		if len(indices) > 0 {
			m.sourceCursor = indices[0]
		}
	}
	if m.feedCursor < 0 {
		m.feedCursor = 0
	}
	if m.feedCursor >= len(m.feeds) && len(m.feeds) > 0 {
		m.feedCursor = len(m.feeds) - 1
	}
	visible := m.visibleFeedIndices()
	if len(visible) > 0 {
		found := false
		for _, index := range visible {
			if index == m.feedCursor {
				found = true
				break
			}
		}
		if !found {
			m.feedCursor = visible[0]
		}
	}
	if m.itemCursor < 0 {
		m.itemCursor = 0
	}
	itemCount := m.itemTargetCount()
	if m.itemCursor >= itemCount && itemCount > 0 {
		m.itemCursor = itemCount - 1
	}
}

func (m *Model) ensureCursorVisible() {
	_, bodyHeight := m.layout()
	visible := max(1, bodyHeight-3)
	switch m.focus {
	case focusFeeds:
		row := m.sourceCursorRow()
		if row < m.feedScroll {
			m.feedScroll = row
		}
		if row >= m.feedScroll+visible {
			m.feedScroll = row - visible + 1
		}
	case focusItems:
		if m.itemCursor < m.itemScroll {
			m.itemScroll = m.itemCursor
		}
		if m.itemCursor >= m.itemScroll+visible {
			m.itemScroll = m.itemCursor - visible + 1
		}
	}
	m.clampListScrolls()
}

func (m *Model) clampListScrolls() {
	_, bodyHeight := m.layout()
	visible := max(1, bodyHeight-3)
	m.feedScroll = clampInt(m.feedScroll, 0, max(0, len(m.sourceRows())-visible))
	m.itemScroll = clampInt(m.itemScroll, 0, max(0, m.itemTargetCount()-visible))
}

func (m *Model) clampScrolls() {
	m.clampListScrolls()
	_, bodyHeight := m.layout()
	visible := max(1, bodyHeight-3)
	m.stageScroll = clampInt(m.stageScroll, 0, max(0, m.stageLineCount()-visible))
}

func (m *Model) renderArticle() {
	if len(m.items) == 0 {
		m.setArticle("No items for this feed. Press r to refresh.")
		return
	}
	item := m.items[m.itemCursor]
	text := item.ContentMarkdown
	if text == "" {
		text = item.Link
	}
	m.setArticle(text)
}

func (m *Model) showItemHint() {
	if len(m.items) == 0 {
		m.clearArticle()
		return
	}
	m.clearArticle()
}

func (m *Model) setArticle(text string) {
	m.rawArticle = text
	m.article = m.renderMarkdown(text)
	m.savedRawArticle = ""
	m.savedArticle = ""
	m.articleMode = articleNormal
}

func (m *Model) clearArticle() {
	m.rawArticle = ""
	m.article = ""
	m.savedRawArticle = ""
	m.savedArticle = ""
	m.articleMode = articleNormal
}

func (m *Model) rerenderArticle() {
	if m.rawArticle != "" {
		m.article = m.renderMarkdown(m.rawArticle)
	}
	if m.savedRawArticle != "" {
		m.savedArticle = m.renderMarkdown(m.savedRawArticle)
	}
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func intText(n int) string {
	return strconv.Itoa(n)
}
