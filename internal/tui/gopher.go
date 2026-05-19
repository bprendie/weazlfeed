package tui

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bprendie/weazlfeed/internal/feed"
	"github.com/bprendie/weazlfeed/internal/store"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) openGopherTarget(item store.Item) (tea.Model, tea.Cmd) {
	switch gopherItemKind(item) {
	case "search":
		m.gopherSearchInput = true
		m.gopherSearchURL = item.Link
		m.input.SetValue("")
		m.input.Placeholder = "search " + item.Title
		m.input.Prompt = "gopher?> "
		m.input.Focus()
		m.status = "gopher search: " + item.Title
		return m, textinputBlink()
	case "binary", "image":
		m.confirmGopherDownload = true
		m.gopherDownloadItem = item
		m.status = "confirm gopher download"
		return m, nil
	}
	return m.dialGopher(item.Link, item.Title)
}

func (m Model) updateGopherDownload(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "esc", "n", "N":
		m.clearGopherDownload()
		m.status = "download cancelled"
		return m, nil
	case "enter", "y", "Y":
		item := m.gopherDownloadItem
		m.clearGopherDownload()
		m.status = "downloading " + item.Title
		return m, downloadGopherCmd(item)
	}
	return m, nil
}

func (m *Model) clearGopherDownload() {
	m.confirmGopherDownload = false
	m.gopherDownloadItem = store.Item{}
}

func (m Model) dialGopher(rawURL, title string) (tea.Model, tea.Cmd) {
	if cached, ok := m.gopherCache[rawURL]; ok {
		return m.applyCachedGopher(rawURL, title, cached), nil
	}
	m.focus = focusArticle
	m.startRendering("dialing gopher")
	m.clearArticle()
	m.gopherStack = append(m.gopherStack, append([]store.Item(nil), m.items...))
	m.gopherTrail = append(m.gopherTrail, firstText(title, gopherCrumb(rawURL)))
	m.gopherURLs = append(m.gopherURLs, rawURL)
	return m, tea.Batch(gopherPageCmd(rawURL, m.readerWidth()), m.spinner.Tick)
}

func (m Model) applyCachedGopher(rawURL, title string, cached gopherCacheEntry) tea.Model {
	m.gopherStack = append(m.gopherStack, append([]store.Item(nil), m.items...))
	m.gopherTrail = append(m.gopherTrail, firstText(title, gopherCrumb(rawURL)))
	m.gopherURLs = append(m.gopherURLs, rawURL)
	m.rendering = false
	m.renderAction = ""
	m.renderStartedAt = time.Time{}
	m.err = ""
	if len(cached.items) > 0 {
		m.items = append([]store.Item(nil), cached.items...)
		m.podcasts = nil
		m.itemCursor = 0
		m.itemScroll = 0
		m.stageScroll = 0
		m.focus = focusItems
		m.clearArticle()
		m.status = "gopher cache: " + intText(len(cached.items)) + " entries"
		return m
	}
	m.focus = focusArticle
	m.rawArticle = cached.text
	m.article = firstText(cached.rendered, renderMarkdownText(cached.text, m.readerWidth()))
	m.savedRawArticle = ""
	m.savedArticle = ""
	m.activeAIItem = store.Item{}
	m.articleMode = articleNormal
	m.stageScroll = 0
	m.status = "gopher cache: document"
	return m
}

func (m Model) runGopherSearch(query string) (tea.Model, tea.Cmd) {
	query = strings.TrimSpace(query)
	if query == "" {
		m.status = "gopher search cancelled"
		return m, nil
	}
	target := gopherSearchTargetURL(m.gopherSearchURL, query)
	return m.dialGopher(target, query)
}

func gopherSearchTargetURL(rawURL, query string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	u.RawQuery = url.QueryEscape(query)
	return u.String()
}

func gopherCrumb(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Hostname() == "" {
		return rawURL
	}
	selector := strings.TrimPrefix(u.EscapedPath(), "/")
	if selector == "" {
		return u.Hostname()
	}
	selector, _ = url.PathUnescape(selector)
	if len(selector) > 1 && isGopherURLType(selector[0]) {
		selector = selector[1:]
	}
	selector = strings.Trim(selector, "/")
	if selector == "" {
		return u.Hostname()
	}
	parts := strings.Split(selector, "/")
	return firstText(parts[len(parts)-1], u.Hostname())
}

func gopherItemKind(item store.Item) string {
	switch item.EnclosureType {
	case "gopher/directory":
		return "directory"
	case "gopher/search":
		return "search"
	case "gopher/info":
		return "info"
	case "text/plain":
		return "text"
	case "image/gopher":
		return "image"
	case "text/html":
		return "html"
	case "gopher/telnet":
		return "telnet"
	case "application/octet-stream":
		return "binary"
	}
	switch gopherEnclosureType(item.Link) {
	case "gopher/directory":
		return "directory"
	case "gopher/search":
		return "search"
	case "text/plain":
		return "text"
	case "image/gopher":
		return "image"
	case "text/html":
		return "html"
	case "gopher/telnet":
		return "telnet"
	case "application/octet-stream":
		return "binary"
	default:
		return "gopher"
	}
}

func isGopherURLType(kind byte) bool {
	return strings.ContainsRune("013456789gIhi", rune(kind))
}

func textinputBlink() tea.Cmd {
	return textinput.Blink
}

func (m Model) bookmarkGopherLocation() (tea.Model, tea.Cmd) {
	rawURL := m.currentGopherURL()
	if rawURL == "" {
		m.status = "no gopher location to bookmark"
		return m, nil
	}
	title := m.currentGopherTitle()
	feedID, err := m.store.UpsertFeed(title, rawURL, "gopher", "Gopher", "Bookmarks", "Bookmarks")
	if err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.revealFeedID = feedID
	m.revealSection = "Gopher"
	m.revealFolder = "Bookmarks"
	m.status = "bookmarked gopher: " + title
	return m, loadFeedsCmd(m.store)
}

func (m Model) currentGopherURL() string {
	if len(m.gopherURLs) > 0 {
		return m.gopherURLs[len(m.gopherURLs)-1]
	}
	if len(m.feeds) > 0 && m.feedCursor >= 0 && m.feedCursor < len(m.feeds) && m.feeds[m.feedCursor].Type == "gopher" {
		return m.feeds[m.feedCursor].URL
	}
	return ""
}

func (m Model) currentGopherTitle() string {
	if len(m.gopherTrail) > 0 {
		return firstText(m.gopherTrail[len(m.gopherTrail)-1], gopherCrumb(m.currentGopherURL()))
	}
	if len(m.feeds) > 0 && m.feedCursor >= 0 && m.feedCursor < len(m.feeds) && m.feeds[m.feedCursor].Type == "gopher" {
		return firstText(m.feeds[m.feedCursor].Title, gopherCrumb(m.feeds[m.feedCursor].URL))
	}
	return firstText(gopherCrumb(m.currentGopherURL()), "gopher bookmark")
}

func downloadGopherCmd(item store.Item) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		data, err := feed.FetchGopherBytes(ctx, item.Link, 512*1024*1024)
		if err != nil {
			return gopherDownloadMsg{err: err}
		}
		dir, err := gopherDownloadDir()
		if err != nil {
			return gopherDownloadMsg{err: err}
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return gopherDownloadMsg{err: err}
		}
		path := filepath.Join(dir, gopherDownloadName(item))
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return gopherDownloadMsg{err: err}
		}
		return gopherDownloadMsg{path: path}
	}
}

func gopherDownloadDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "weazlfeed-gopher", nil
	}
	return filepath.Join(home, "Downloads", "weazlfeed-gopher"), nil
}

func gopherDownloadName(item store.Item) string {
	name := strings.TrimSpace(item.Title)
	if name == "" {
		name = gopherCrumb(item.Link)
	}
	name = regexp.MustCompile(`[^A-Za-z0-9._-]+`).ReplaceAllString(name, "_")
	name = strings.Trim(name, "._-")
	if name == "" {
		name = "gopher_payload"
	}
	return name
}
