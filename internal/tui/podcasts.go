package tui

import (
	"github.com/bprendie/weazlfeed/internal/store"
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
	m.refreshing = true
	return m, refreshCmd(m.store, []store.Feed{{
		ID:       feedID,
		Title:    result.Title,
		URL:      result.FeedURL,
		Type:     "rss",
		Section:  "Podcasts",
		Folder:   folder,
		Category: folder,
	}}, m.ai)
}
