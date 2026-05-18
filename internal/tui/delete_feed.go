package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) startDeleteFeed() (tea.Model, tea.Cmd) {
	if m.focus != focusFeeds {
		return m, nil
	}
	row, ok := m.selectedSourceRow()
	if !ok || row.kind != sourceFeed {
		m.status = "select a feed to delete"
		return m, nil
	}
	feed := m.feeds[row.feedIndex]
	m.confirmDelete = true
	m.deleteFeedID = feed.ID
	m.deleteFeedTitle = feed.Title
	m.status = "confirm delete"
	return m, nil
}

func (m Model) updateDeleteFeed(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "esc", "n", "N":
		m.clearDeleteFeed()
		m.status = "delete cancelled"
		return m, nil
	case "enter", "y", "Y":
		feedID := m.deleteFeedID
		title := m.deleteFeedTitle
		m.clearDeleteFeed()
		m.status = "deleting " + title
		return m, deleteFeedCmd(m.store, feedID, title)
	}
	return m, nil
}

func (m *Model) clearDeleteFeed() {
	m.confirmDelete = false
	m.deleteFeedID = 0
	m.deleteFeedTitle = ""
}
