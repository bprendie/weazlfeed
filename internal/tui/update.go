package tui

import (
	"strings"

	"github.com/bprendie/weazlfeed/internal/audio"
	"github.com/bprendie/weazlfeed/internal/store"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = size.Width, size.Height
		if m.lockMode != lockOpen {
			return m, nil
		}
	}
	if m.lockMode != lockOpen {
		return m.updateLock(msg)
	}
	if m.asking || m.folderInput || m.podcastInput {
		return m.updateInput(msg)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.updateKey(msg)
	case tea.MouseMsg:
		m.updateMouse(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if m.refreshing {
			return m, cmd
		}
	case feedsMsg:
		m.feeds, m.err = msg.feeds, errText(msg.err)
		m.folders = msg.folders
		if len(m.feeds) > 0 {
			return m, prefetchItemsCmd(m.store, m.feeds, m.hideSludge)
		}
	case itemsMsg:
		m.err = errText(msg.err)
		if msg.err == nil {
			m.itemCache[msg.feedID] = msg.items
			if m.focus == focusItems && m.selectedFeedID() == msg.feedID {
				m.items = msg.items
				m.podcasts = nil
				m.clamp()
				m.clearArticle()
			}
		}
	case allItemsMsg:
		m.err = errText(msg.err)
		if msg.err == nil {
			m.itemCache = msg.itemsByFeed
			if m.focus == focusItems {
				if items, ok := m.itemCache[m.selectedFeedID()]; ok {
					m.items = items
					m.podcasts = nil
					m.clamp()
					m.clearArticle()
				}
			}
		}
	case fetchMsg:
		m.refreshing = false
		m.itemCache = map[int64][]store.Item{}
		m.err = errText(msg.err)
		m.status = "refresh complete: checked " + intText(msg.checked) + " new " + intText(msg.added) + " failed " + intText(msg.failed)
		return m, loadFeedsCmd(m.store)
	case aiMsg:
		m.err = errText(msg.err)
		if msg.err == nil {
			m.setArticle(msg.text)
			m.status = "local extraction complete"
		}
	case articleMsg:
		m.err = errText(msg.err)
		if msg.err == nil {
			m.setArticle(msg.text)
			m.stageScroll = 0
			m.status = "gopher target loaded"
		}
	case podcastSearchMsg:
		m.err = errText(msg.err)
		if msg.err == nil {
			m.podcasts = msg.results
			m.itemCursor = 0
			m.itemScroll = 0
			m.focus = focusItems
			m.setArticle("Select a podcast result and press enter to subscribe.")
			m.status = "podcast search: " + intText(len(msg.results)) + " results"
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
	case "esc":
		m.retreat()
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
	case "left":
		if m.focus == focusFeeds {
			return m.toggleCurrentFolder(true)
		}
	case "right":
		if m.focus == focusFeeds {
			return m.toggleCurrentFolder(false)
		}
	case "R":
		m.refreshing = true
		m.itemCache = map[int64][]store.Item{}
		m.status = m.spinner.View() + " refreshing all sources"
		return m, tea.Batch(refreshAllCmd(m.store, m.ai), m.spinner.Tick)
	case "r":
		if len(m.feeds) > 0 {
			m.refreshing = true
			delete(m.itemCache, m.feeds[m.feedCursor].ID)
			m.status = m.spinner.View() + " refreshing " + m.feeds[m.feedCursor].Title
			return m, tea.Batch(refreshCmd(m.store, []store.Feed{m.feeds[m.feedCursor]}, m.ai), m.spinner.Tick)
		}
	case "h":
		m.hideSludge = !m.hideSludge
		m.itemCache = map[int64][]store.Item{}
		if len(m.feeds) > 0 {
			return m, prefetchItemsCmd(m.store, m.feeds, m.hideSludge)
		}
	case "enter":
		return m.activate()
	case " ":
		if m.focus == focusFeeds && len(m.feeds) > 0 {
			return m.pickOrDropFeed()
		}
		if m.paused {
			_ = m.player.Resume()
		} else {
			_ = m.player.TogglePause()
		}
		m.paused = !m.paused
		m.savePlayhead()
	case "s":
		m.stopAudio()
	case "n":
		if m.focus == focusFeeds {
			m.folderInput = true
			m.input.Placeholder = "new folder"
			m.input.Prompt = "folder> "
			m.input.EchoMode = textinput.EchoNormal
			m.input.Focus()
			return m, nil
		}
	case "p":
		m.podcastInput = true
		m.input.Placeholder = "podcast search"
		m.input.Prompt = "podcast> "
		m.input.EchoMode = textinput.EchoNormal
		m.input.Focus()
		return m, nil
	case "ctrl+a":
		if m.aiEnabled && len(m.items) > 0 {
			m.asking = true
			m.input.EchoMode = textinput.EchoNormal
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
			m.folderInput = false
			m.podcastInput = false
			m.input.Blur()
			m.input.SetValue("")
			m.input.Prompt = "interrogate> "
			return m, nil
		case "enter":
			question := strings.TrimSpace(m.input.Value())
			m.asking = false
			folderInput := m.folderInput
			podcastInput := m.podcastInput
			m.folderInput = false
			m.podcastInput = false
			m.input.Blur()
			m.input.SetValue("")
			m.input.Prompt = "interrogate> "
			if folderInput {
				return m.createFolder(question)
			}
			if podcastInput {
				if question == "" {
					return m, nil
				}
				m.status = "searching podcasts"
				return m, podcastSearchCmd(question)
			}
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
