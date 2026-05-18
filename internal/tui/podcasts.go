package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) itemTargetCount() int {
	return len(m.items)
}

func (m Model) subscribePodcast() (tea.Model, tea.Cmd) {
	if len(m.podcasts) == 0 || m.podcastCursor >= len(m.podcasts) {
		return m, nil
	}
	result := m.podcasts[m.podcastCursor]
	folder := "Search"
	if row, ok := m.selectedSourceRow(); ok && row.section == "Podcasts" && row.folder != "" {
		folder = row.folder
	}
	feedID, err := m.store.UpsertFeed(result.Title, result.FeedURL, "rss", "Podcasts", folder, folder)
	if err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.status = "subscribed podcast: " + result.Title + " -> " + folder
	m.revealFeedID = feedID
	m.revealSection = "Podcasts"
	m.revealFolder = folder
	m.podcastInput = false
	m.podcasts = nil
	m.podcastCursor = 0
	m.podcastScroll = 0
	m.podcastSearching = false
	m.input.Blur()
	m.input.SetValue("")
	m.input.Prompt = "interrogate> "
	return m, loadFeedsCmd(m.store)
}
