package tui

import (
	"net/url"
	"strings"

	"github.com/bprendie/weazlfeed/internal/store"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) openGopherTarget(item store.Item) (tea.Model, tea.Cmd) {
	if gopherItemKind(item) == "search" {
		m.gopherSearchInput = true
		m.gopherSearchURL = item.Link
		m.input.SetValue("")
		m.input.Placeholder = "search " + item.Title
		m.input.Prompt = "gopher?> "
		m.input.Focus()
		m.status = "gopher search: " + item.Title
		return m, textinputBlink()
	}
	return m.dialGopher(item.Link, item.Title)
}

func (m Model) dialGopher(rawURL, title string) (tea.Model, tea.Cmd) {
	if cached, ok := m.gopherCache[rawURL]; ok {
		return m.applyCachedGopher(rawURL, title, cached), nil
	}
	m.focus = focusArticle
	m.rendering = true
	m.clearArticle()
	m.status = "dialing gopher target"
	m.gopherStack = append(m.gopherStack, append([]store.Item(nil), m.items...))
	m.gopherTrail = append(m.gopherTrail, firstText(title, gopherCrumb(rawURL)))
	return m, tea.Batch(gopherPageCmd(rawURL), m.spinner.Tick)
}

func (m Model) applyCachedGopher(rawURL, title string, cached gopherCacheEntry) tea.Model {
	m.gopherStack = append(m.gopherStack, append([]store.Item(nil), m.items...))
	m.gopherTrail = append(m.gopherTrail, firstText(title, gopherCrumb(rawURL)))
	m.rendering = false
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
	m.setArticle(cached.text)
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
