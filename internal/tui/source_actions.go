package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) addURL(rawURL string) (tea.Model, tea.Cmd) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return m, nil
	}
	section, folder := m.selectedFolderTarget()
	m.status = m.spinner.View() + " adding source"
	m.refreshing = true
	return m, tea.Batch(addFeedCmd(m.store, rawURL, section, folder), m.spinner.Tick)
}

func (m Model) selectedFolderTarget() (string, string) {
	row, ok := m.selectedSourceRow()
	if !ok {
		return "News", "General"
	}
	section := firstText(row.section, "News")
	folder := row.folder
	if row.kind == sourceFeed {
		feed := m.feeds[row.feedIndex]
		section = firstText(feed.Section, sectionFromFeed(feed))
		folder = firstText(feed.Folder, folderFromFeed(feed))
	}
	if folder == "" {
		folder = "General"
	}
	return section, folder
}
