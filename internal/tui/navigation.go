package tui

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) move(delta int) {
	switch m.focus {
	case focusFeeds:
		indices := m.selectableSourceRows()
		if len(indices) == 0 {
			return
		}
		pos := 0
		for i, index := range indices {
			if index == m.sourceCursor {
				pos = i
				break
			}
		}
		pos = clampInt(pos+delta, 0, len(indices)-1)
		m.sourceCursor = indices[pos]
		m.syncFeedCursorFromSource()
		m.clamp()
		m.ensureCursorVisible()
		m.itemCursor = 0
		m.itemScroll = 0
		m.stageScroll = 0
		m.items = nil
		m.podcasts = nil
		m.clearArticle()
	case focusItems:
		m.itemCursor += delta
		m.clamp()
		m.ensureCursorVisible()
	case focusArticle:
		m.stageScroll += delta
		m.clampScrolls()
	}
}

func (m *Model) retreat() {
	if m.focus > focusFeeds {
		m.focus--
	}
	m.status = "back"
	m.clamp()
	m.ensureCursorVisible()
}

func (m Model) toggleCurrentFolder(collapsed bool) (tea.Model, tea.Cmd) {
	row, ok := m.selectedSourceRow()
	if !ok {
		return m, nil
	}
	section, folder := row.section, row.folder
	if row.kind == sourceFeed {
		feed := m.feeds[row.feedIndex]
		section = firstText(feed.Section, sectionFromFeed(feed))
		folder = firstText(feed.Folder, folderFromFeed(feed))
	}
	if row.kind == sourceSection || folder == "" {
		return m, nil
	}
	if err := m.setFolderCollapsed(section, folder, collapsed); err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.status = "folder " + folder
	if collapsed {
		m.status += " collapsed"
	} else {
		m.status += " expanded"
	}
	if collapsed {
		m.selectSourceRow(section, folder)
	}
	m.ensureCursorVisible()
	return m, nil
}

func (m Model) selectedFeedID() int64 {
	if len(m.feeds) == 0 || m.feedCursor < 0 || m.feedCursor >= len(m.feeds) {
		return 0
	}
	return m.feeds[m.feedCursor].ID
}

func (m *Model) useCachedItems(feedID int64) bool {
	items, ok := m.itemCache[feedID]
	if !ok {
		return false
	}
	m.items = items
	m.podcasts = nil
	m.itemCursor = 0
	m.itemScroll = 0
	m.stageScroll = 0
	m.clamp()
	m.clearArticle()
	return true
}
