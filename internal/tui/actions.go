package tui

import (
	"strconv"
	"strings"

	"github.com/bprendie/weazlfeed/internal/audio"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) move(delta int) {
	switch m.focus {
	case focusFeeds:
		m.feedCursor += delta
		m.clamp()
		if len(m.feeds) > 0 {
			m.itemCursor = 0
			m.items = nil
			m.article = ""
		}
	case focusItems:
		m.itemCursor += delta
		m.clamp()
		m.renderArticle()
	}
}

func (m Model) activate() (tea.Model, tea.Cmd) {
	if m.focus == focusFeeds && len(m.feeds) > 0 {
		return m, loadItemsCmd(m.store, m.feeds[m.feedCursor].ID, m.hideSludge)
	}
	if len(m.items) == 0 {
		return m, nil
	}
	item := m.items[m.itemCursor]
	_ = m.store.MarkRead(item.ID)
	if strings.HasPrefix(strings.ToLower(item.Link), "gopher://") {
		m.status = "dialing gopher target"
		return m, gopherArticleCmd(item.Link)
	}
	if item.EnclosureURL != "" && strings.HasPrefix(item.EnclosureType, "audio/") {
		m.stopAudio()
		if err := m.player.Play(item.EnclosureURL, item.PlayheadSeconds); err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.playingID = item.ID
		m.paused = false
		m.status = "playing " + item.Title
		tick := playheadTickCmd()
		if meter, err := audio.StartMeter(item.EnclosureURL); err == nil {
			m.meter = meter
			return m, tea.Batch(meterCmd(meter.Samples()), tick)
		}
		return m, tick
	}
	m.renderArticle()
	return m, loadItemsCmd(m.store, item.FeedID, m.hideSludge)
}

func (m *Model) stopAudio() {
	m.savePlayhead()
	m.player.Stop()
	if m.meter != nil {
		m.meter.Stop()
		m.meter = nil
	}
	m.playingID = 0
	m.paused = false
	m.bars = nil
}

func (m *Model) savePlayhead() {
	if m.playingID == 0 {
		return
	}
	_ = m.store.SetPlayhead(m.playingID, m.player.Position())
}

func (m *Model) clamp() {
	if m.feedCursor < 0 {
		m.feedCursor = 0
	}
	if m.feedCursor >= len(m.feeds) && len(m.feeds) > 0 {
		m.feedCursor = len(m.feeds) - 1
	}
	if m.itemCursor < 0 {
		m.itemCursor = 0
	}
	if m.itemCursor >= len(m.items) && len(m.items) > 0 {
		m.itemCursor = len(m.items) - 1
	}
}

func (m *Model) renderArticle() {
	if len(m.items) == 0 {
		m.article = "No items for this feed. Press r to refresh."
		return
	}
	item := m.items[m.itemCursor]
	text := item.ContentMarkdown
	if text == "" {
		text = item.Link
	}
	if rendered, err := m.renderer.Render(text); err == nil {
		m.article = rendered
		return
	}
	m.article = text
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
