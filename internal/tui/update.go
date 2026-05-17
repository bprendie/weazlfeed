package tui

import (
	"strings"

	"github.com/bprendie/weazlfeed/internal/audio"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.asking {
		return m.updateInput(msg)
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		return m.updateKey(msg)
	case tea.MouseMsg:
		m.updateMouse(msg)
	case feedsMsg:
		m.feeds, m.err = msg.feeds, errText(msg.err)
		if len(m.feeds) > 0 {
			return m, loadItemsCmd(m.store, m.feeds[m.feedCursor].ID, m.hideSludge)
		}
	case itemsMsg:
		m.items, m.err = msg.items, errText(msg.err)
		m.clamp()
		m.renderArticle()
	case fetchMsg:
		m.err = errText(msg.err)
		m.status = "refresh complete: new items " + intText(msg.added)
		return m, loadFeedsCmd(m.store)
	case aiMsg:
		m.err = errText(msg.err)
		if msg.err == nil {
			m.article = msg.text
			m.status = "local extraction complete"
		}
	case articleMsg:
		m.err = errText(msg.err)
		if msg.err == nil {
			m.article = msg.text
			m.stageScroll = 0
			m.status = "gopher target loaded"
		}
	case meterMsg:
		sample := audio.Sample(msg)
		m.bars = sample.Bands
		if m.meter != nil {
			return m, meterCmd(m.meter.Samples())
		}
	case playheadTickMsg:
		if m.playingID != 0 {
			m.savePlayhead()
			return m, playheadTickCmd()
		}
	}
	return m, nil
}

func (m Model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.stopAudio()
		return m, tea.Quit
	case "tab":
		m.focus = (m.focus + 1) % 3
	case "j", "down":
		m.move(1)
	case "k", "up":
		m.move(-1)
	case "pgdown":
		m.page(1)
	case "pgup":
		m.page(-1)
	case "home":
		m.home()
	case "end":
		m.end()
	case "r":
		m.status = "refreshing feeds"
		return m, refreshCmd(m.store, m.feeds, m.ai)
	case "h":
		m.hideSludge = !m.hideSludge
		if len(m.feeds) > 0 {
			return m, loadItemsCmd(m.store, m.feeds[m.feedCursor].ID, m.hideSludge)
		}
	case "enter":
		return m.activate()
	case " ":
		if m.paused {
			_ = m.player.Resume()
		} else {
			_ = m.player.TogglePause()
		}
		m.paused = !m.paused
		m.savePlayhead()
	case "s":
		m.stopAudio()
	case "ctrl+a":
		if m.aiEnabled && len(m.items) > 0 {
			m.asking = true
			m.input.Focus()
			return m, nil
		}
	case "ctrl+t":
		if m.aiEnabled && len(m.items) > 0 {
			m.status = "extracting critical points"
			return m, aiCmd(m.ai, "triage", m.items[m.itemCursor], "")
		}
	}
	return m, nil
}

func (m Model) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc", "ctrl+c":
			m.asking = false
			m.input.Blur()
			m.input.SetValue("")
			return m, nil
		case "enter":
			question := strings.TrimSpace(m.input.Value())
			m.asking = false
			m.input.Blur()
			m.input.SetValue("")
			if question != "" && len(m.items) > 0 {
				m.status = "interrogating active article"
				return m, aiCmd(m.ai, "ask", m.items[m.itemCursor], question)
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}
