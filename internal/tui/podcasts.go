package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) itemTargetCount() int {
	return len(m.items)
}

func (m Model) currentFeedIsPodcast() bool {
	if len(m.feeds) == 0 || m.feedCursor < 0 || m.feedCursor >= len(m.feeds) {
		return false
	}
	return m.feeds[m.feedCursor].Section == "Podcasts"
}

func (m Model) finishPodcastItem() (tea.Model, tea.Cmd) {
	if !m.currentFeedIsPodcast() || len(m.items) == 0 || m.itemCursor >= len(m.items) {
		return m, nil
	}
	item := m.items[m.itemCursor]
	if item.ID == 0 {
		return m, nil
	}
	_ = m.store.MarkRead(item.ID)
	if item.DurationSeconds > 0 {
		_ = m.store.SetPlayhead(item.ID, item.DurationSeconds)
		item.PlayheadSeconds = item.DurationSeconds
	}
	item.ReadStatus = true
	m.items[m.itemCursor] = item
	if item.FeedID != 0 {
		m.itemCache[item.FeedID] = m.items
	}
	m.status = "finished: " + item.Title
	return m, nil
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
	if result.FeedURL == "" {
		m.status = "podcast result has no feed url"
		return m, nil
	}
	m.podcastInput = false
	m.podcasts = nil
	m.podcastCursor = 0
	m.podcastScroll = 0
	m.podcastSearching = false
	m.input.Blur()
	m.input.SetValue("")
	m.input.Prompt = "interrogate> "
	m.refreshing = true
	m.status = m.spinner.View() + " subscribing podcast: " + result.Title
	return m, tea.Batch(addFeedCmd(m.store, result.FeedURL, "Podcasts", folder), m.spinner.Tick)
}
