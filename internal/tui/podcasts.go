package tui

import (
	"fmt"

	"github.com/bprendie/weazlfeed/internal/podcast"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) podcastMode() bool {
	return len(m.podcasts) > 0
}

func (m Model) itemTargetCount() int {
	if m.podcastMode() {
		return len(m.podcasts)
	}
	return len(m.items)
}

func (m Model) visiblePodcasts() []podcast.Result {
	_, bodyHeight := m.layout()
	start := clampInt(m.itemScroll, 0, len(m.podcasts))
	end := clampInt(start+max(1, bodyHeight-4), start, len(m.podcasts))
	return m.podcasts[start:end]
}

func (m Model) renderPodcastItems(width, height int) string {
	width = panelContentWidth(m.styles.panel, width)
	lines := []string{m.styles.help.Render("podcast search / enter subscribes")}
	for i, result := range m.visiblePodcasts() {
		index := m.itemScroll + i
		title := result.Title
		if result.Author != "" {
			title += " / " + result.Author
		}
		line := truncate(" - [PODCAST] "+title, width)
		if index == m.itemCursor {
			line = m.styles.selected.Render(truncate("=> [PODCAST] "+title, width))
		} else {
			line = m.styles.item.Render(line)
		}
		lines = append(lines, line)
	}
	return fitLines(lines, height-3)
}

func (m Model) subscribePodcast() (tea.Model, tea.Cmd) {
	if !m.podcastMode() || m.itemCursor >= len(m.podcasts) {
		return m, nil
	}
	result := m.podcasts[m.itemCursor]
	if _, err := m.store.UpsertFeed(result.Title, result.FeedURL, "rss", "Podcasts", "Search", "Search"); err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.status = "subscribed podcast: " + result.Title
	m.podcasts = nil
	m.itemCursor = 0
	m.itemScroll = 0
	m.article = fmt.Sprintf("# %s\n\n%s", result.Title, result.FeedURL)
	return m, loadFeedsCmd(m.store)
}
