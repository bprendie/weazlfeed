package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) startDeleteFeed() (tea.Model, tea.Cmd) {
	if m.focus != focusFeeds {
		return m, nil
	}
	row, ok := m.selectedSourceRow()
	if !ok || (row.kind != sourceFeed && row.kind != sourceInterrogation) {
		m.status = "select a feed or interrogation to delete"
		return m, nil
	}
	m.deleteKind = "feed"
	m.confirmDelete = true
	if row.kind == sourceInterrogation {
		if row.aiIndex < 0 || row.aiIndex >= len(m.interrogations) {
			return m, nil
		}
		out := m.interrogations[row.aiIndex]
		m.deleteKind = "interrogation"
		m.deleteFeedID = out.ID
		m.deleteFeedTitle = interrogationTitle(out)
		m.status = "confirm delete"
		return m, nil
	}
	feed := m.feeds[row.feedIndex]
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
		kind := m.deleteKind
		m.clearDeleteFeed()
		m.status = "deleting " + title
		if kind == "interrogation" {
			return m, deleteInterrogationCmd(m.store, feedID, title)
		}
		return m, deleteFeedCmd(m.store, feedID, title)
	}
	return m, nil
}

func (m *Model) clearDeleteFeed() {
	m.confirmDelete = false
	m.deleteFeedID = 0
	m.deleteFeedTitle = ""
	m.deleteKind = ""
}
