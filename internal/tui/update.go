package tui

import (
	"time"

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
	if m.helpOpen {
		return m.updateHelp(msg)
	}
	if m.confirmDelete {
		return m.updateDeleteFeed(msg)
	}
	if m.bouncerOpen {
		return m.updateBouncer(msg)
	}
	if m.podcastInput {
		return m.updatePodcastDirectory(msg)
	}
	if m.asking || m.folderInput || m.urlInput {
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
		if m.refreshing || m.rendering || m.aiWorking {
			return m, cmd
		}
	case aiTickMsg:
		if m.aiWorking {
			return m, aiTickCmd()
		}
	case feedsMsg:
		m.feeds, m.err = msg.feeds, errText(msg.err)
		m.folders = msg.folders
		m.interrogations = msg.interrogations
		m.bouncerRules = msg.rules
		m.revealPendingFeed()
		m.clamp()
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
	case addFeedMsg:
		m.refreshing = false
		m.itemCache = map[int64][]store.Item{}
		m.err = errText(msg.err)
		if msg.err == nil && msg.title != "" {
			m.status = "added " + msg.title + " -> " + msg.section + "/" + msg.folder + " (" + intText(msg.added) + " items)"
			return m, loadFeedsCmd(m.store)
		}
	case deleteFeedMsg:
		m.itemCache = map[int64][]store.Item{}
		m.err = errText(msg.err)
		if msg.err == nil {
			if msg.kind == "interrogation" {
				m.status = "deleted interrogation: " + msg.title
			} else {
				m.items = nil
				m.clearArticle()
				m.status = "deleted feed: " + msg.title
			}
			return m, loadFeedsCmd(m.store)
		}
	case aiMsg:
		m.aiWorking = false
		m.aiAction = ""
		m.aiStartedAt = time.Time{}
		m.aiReqIn = msg.inTokens
		m.aiReqOut = msg.outTokens
		m.err = errText(msg.err)
		if msg.err == nil {
			m.showAIOutput(msg)
			if msg.kind == "ask" {
				return m, loadFeedsCmd(m.store)
			}
		}
	case articleMsg:
		m.rendering = false
		m.err = errText(msg.err)
		if msg.err == nil {
			m.setArticle(msg.text)
			m.stageScroll = 0
			m.status = "gopher target loaded"
		}
	case gopherMsg:
		m.rendering = false
		m.err = errText(msg.err)
		if msg.err == nil && len(msg.items) > 0 {
			m.items = msg.items
			m.podcasts = nil
			m.itemCursor = 0
			m.itemScroll = 0
			m.stageScroll = 0
			m.focus = focusItems
			m.clearArticle()
			m.status = "gopher menu: " + intText(len(msg.items)) + " entries"
		}
		if msg.err == nil && len(msg.items) == 0 {
			m.focus = focusArticle
			m.setArticle(msg.text)
			m.stageScroll = 0
			m.status = "gopher document loaded"
		}
	case readerMsg:
		m.rendering = false
		m.err = errText(msg.err)
		if msg.err == nil {
			m.rawArticle = msg.raw
			m.article = msg.rendered
			m.stageScroll = 0
			m.status = "reader ready"
			for i := range m.items {
				if m.items[i].ID == msg.item.ID {
					m.items[i] = msg.item
					m.items[i].ReadStatus = true
					if msg.item.FeedID != 0 {
						m.itemCache[msg.item.FeedID] = m.items
					}
					break
				}
			}
		}
	case interrogationMsg:
		m.rendering = false
		m.err = errText(msg.err)
		if msg.err == nil {
			m.rawArticle = msg.raw
			m.article = msg.rendered
			m.savedRawArticle = ""
			m.savedArticle = ""
			m.articleMode = articleAsk
			m.stageScroll = 0
			m.status = "saved interrogation"
		}
	case podcastSearchMsg:
		if !m.podcastInput {
			return m, nil
		}
		m.podcastSearching = false
		m.err = errText(msg.err)
		if msg.err == nil {
			m.podcasts = msg.results
			m.podcastCursor = 0
			m.podcastScroll = 0
			m.input.Blur()
			m.status = "podcast search: " + intText(len(msg.results)) + " results"
		}
	case bouncerRulesMsg:
		return m.handleBouncerRulesMsg(msg)
	case bouncerActionMsg:
		return m.handleBouncerActionMsg(msg)
	case bouncerScanMsg:
		return m.handleBouncerScanMsg(msg)
	case playheadTickMsg:
		if m.playingID != 0 {
			m.savePlayhead()
			return m, playheadTickCmd()
		}
	case audioTickMsg:
		m.drainMeter()
		m.visualizer.Step(m.player.Active() && !m.paused, m.energy)
		if m.playingID != 0 {
			return m, audioTickCmd()
		}
	}
	return m, nil
}

func (m *Model) drainMeter() {
	if m.meter == nil {
		m.energy = audio.Sample{}
		return
	}
	for {
		select {
		case sample, ok := <-m.meter.Samples():
			if !ok {
				m.meter = nil
				return
			}
			m.energy = sample
		default:
			return
		}
	}
}

func (m Model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.stopAudio()
		return m, tea.Quit
	case "tab":
		m.focus = (m.focus + 1) % 3
	case "esc":
		if m.playingID != 0 {
			m.stopAudio()
			return m, nil
		}
		if m.articleMode != articleNormal {
			m.restoreArticle()
			return m, nil
		}
		m.retreat()
	case "ctrl+k", "?", "f1":
		m.helpOpen = true
		m.status = "help"
	case "ctrl+b":
		m.bouncerOpen = true
		m.bouncerInput = false
		m.bouncerCursor = clampInt(m.bouncerCursor, 0, max(0, len(m.bouncerRules)-1))
		m.status = "bouncer rules"
		return m, loadBouncerRulesCmd(m.store)
	case "ctrl+d":
		return m.startDeleteFeed()
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
		m.retreat()
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
		if row, ok := m.selectedSourceRow(); ok && row.kind == sourceFeed {
			feed := m.feeds[row.feedIndex]
			m.refreshing = true
			delete(m.itemCache, feed.ID)
			m.status = m.spinner.View() + " refreshing " + feed.Title
			return m, tea.Batch(refreshCmd(m.store, []store.Feed{feed}, m.ai), m.spinner.Tick)
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
			if m.playingURL != "" {
				if meter, err := audio.StartMeter(m.playingURL, m.player.Position()); err == nil {
					m.meter = meter
				}
			}
		} else {
			_ = m.player.TogglePause()
			if m.meter != nil {
				m.meter.Stop()
				m.meter = nil
			}
		}
		m.paused = !m.paused
		m.savePlayhead()
	case ",", "<":
		m.seekAudio(-10)
	case ".", ">":
		m.seekAudio(30)
	case "f":
		if m.focus == focusItems {
			return m.finishPodcastItem()
		}
	case "n":
		if m.focus == focusFeeds {
			m.folderInput = true
			m.input.Placeholder = "new folder"
			m.input.Prompt = "folder> "
			m.input.EchoMode = textinput.EchoNormal
			m.input.Focus()
			return m, nil
		}
	case "a":
		if m.focus == focusFeeds {
			m.urlInput = true
			m.input.Placeholder = "feed or gopher url"
			m.input.Prompt = "url> "
			m.input.EchoMode = textinput.EchoNormal
			m.input.Focus()
			return m, nil
		}
	case "p":
		m.podcastInput = true
		m.podcasts = nil
		m.podcastCursor = 0
		m.podcastScroll = 0
		m.podcastSearching = false
		m.input.SetValue("")
		m.input.Placeholder = "podcast search"
		m.input.Prompt = "podcast> "
		m.input.EchoMode = textinput.EchoNormal
		m.input.Focus()
		return m, nil
	case "ctrl+a":
		if m.aiEnabled && (len(m.items) > 0 || m.activeAIItem.ContentMarkdown != "") {
			m.asking = true
			m.input.EchoMode = textinput.EchoNormal
			m.input.Focus()
			return m, nil
		}
	case "ctrl+t":
		if m.aiEnabled && len(m.items) > 0 {
			m.aiWorking = true
			m.aiAction = "extracting critical points"
			m.aiStartedAt = time.Now()
			m.aiReqIn = estimateTokens(m.items[m.itemCursor].ContentMarkdown)
			m.aiReqOut = 0
			m.status = "extracting critical points"
			return m, tea.Batch(aiCmd(m.store, m.ai, "triage", m.items[m.itemCursor], ""), m.spinner.Tick, aiTickCmd())
		}
	}
	return m, nil
}
