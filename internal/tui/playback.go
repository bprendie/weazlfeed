package tui

import (
	"github.com/bprendie/weazlfeed/internal/audio"
	"github.com/bprendie/weazlfeed/internal/store"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) playAudio(item store.Item) (tea.Model, tea.Cmd) {
	m.stopAudio()
	if err := m.player.Play(item.EnclosureURL, item.PlayheadSeconds); err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.playingID = item.ID
	m.playingURL = item.EnclosureURL
	m.playingTitle = item.Title
	m.playingTotal = item.DurationSeconds
	m.paused = false
	m.status = "playing " + item.Title
	if meter, err := audio.StartMeter(item.EnclosureURL, item.PlayheadSeconds); err == nil {
		m.meter = meter
	}
	return m, tea.Batch(playheadTickCmd(), audioTickCmd())
}

func (m *Model) stopAudio() {
	m.savePlayhead()
	m.player.Stop()
	if m.meter != nil {
		m.meter.Stop()
		m.meter = nil
	}
	m.playingID = 0
	m.playingURL = ""
	m.playingTitle = ""
	m.playingTotal = 0
	m.paused = false
	m.energy = audio.Sample{}
	m.status = "audio stopped"
}

func (m *Model) savePlayhead() {
	if m.playingID == 0 {
		return
	}
	position := m.player.Position()
	_ = m.store.SetPlayhead(m.playingID, position)
	m.updatePlayhead(position)
}

func (m *Model) updatePlayhead(position int) {
	for i := range m.items {
		if m.items[i].ID == m.playingID {
			m.items[i].PlayheadSeconds = position
			if m.items[i].FeedID != 0 {
				m.itemCache[m.items[i].FeedID] = m.items
			}
			return
		}
	}
}

func (m *Model) seekAudio(delta int) {
	if m.playingID == 0 || m.playingURL == "" {
		return
	}
	if m.meter != nil {
		m.meter.Stop()
		m.meter = nil
	}
	if err := m.player.Seek(delta); err != nil {
		m.err = err.Error()
		return
	}
	m.paused = false
	m.savePlayhead()
	position := m.player.Position()
	if meter, err := audio.StartMeter(m.playingURL, position); err == nil {
		m.meter = meter
	}
	m.status = "seek " + audioPosition(position, m.playingTotal)
}
