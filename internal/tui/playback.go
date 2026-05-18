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
}

func (m *Model) savePlayhead() {
	if m.playingID == 0 {
		return
	}
	_ = m.store.SetPlayhead(m.playingID, m.player.Position())
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
