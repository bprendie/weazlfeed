package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updatePodcastDirectory(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc", "ctrl+c":
			m.closePodcastDirectory()
			return m, nil
		case "/":
			m.input.Focus()
			return m, nil
		case "enter":
			if m.input.Focused() {
				query := strings.TrimSpace(m.input.Value())
				if query == "" {
					return m, nil
				}
				m.status = "searching podcasts"
				return m, podcastSearchCmd(query)
			}
			return m.subscribePodcast()
		case "a":
			if !m.input.Focused() {
				return m.subscribePodcast()
			}
		case "j", "down":
			if !m.input.Focused() {
				m.movePodcast(1)
				return m, nil
			}
		case "k", "up":
			if !m.input.Focused() {
				m.movePodcast(-1)
				return m, nil
			}
		case "pgdown":
			if !m.input.Focused() {
				m.movePodcast(8)
				return m, nil
			}
		case "pgup":
			if !m.input.Focused() {
				m.movePodcast(-8)
				return m, nil
			}
		}
	}
	if m.input.Focused() {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) movePodcast(delta int) {
	if len(m.podcasts) == 0 {
		return
	}
	m.podcastCursor = clampInt(m.podcastCursor+delta, 0, len(m.podcasts)-1)
	visible := max(1, m.podcastVisibleRows())
	if m.podcastCursor < m.podcastScroll {
		m.podcastScroll = m.podcastCursor
	}
	if m.podcastCursor >= m.podcastScroll+visible {
		m.podcastScroll = m.podcastCursor - visible + 1
	}
}

func (m Model) podcastVisibleRows() int {
	_, bodyHeight := m.layout()
	contentHeight := max(10, bodyHeight-4)
	return max(1, contentHeight-4)
}

func (m *Model) closePodcastDirectory() {
	m.podcastInput = false
	m.podcasts = nil
	m.podcastCursor = 0
	m.podcastScroll = 0
	m.input.Blur()
	m.input.SetValue("")
	m.input.Prompt = "interrogate> "
	m.status = "podcast directory closed"
}
