package tui

import (
	"context"

	"github.com/bprendie/weazlfeed/internal/audio"
	"github.com/bprendie/weazlfeed/internal/config"
	"github.com/bprendie/weazlfeed/internal/llm"
	"github.com/bprendie/weazlfeed/internal/store"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

type focus int

const (
	focusFeeds focus = iota
	focusItems
	focusArticle
)

type Model struct {
	cfg        config.Config
	cfgPath    string
	store      *store.Store
	styles     styles
	renderer   *glamour.TermRenderer
	ai         llm.Client
	player     audio.Player
	meter      *audio.Meter
	input      textinput.Model
	focus      focus
	width      int
	height     int
	feeds      []store.Feed
	items      []store.Item
	feedCursor int
	itemCursor int
	article    string
	status     string
	err        string
	hideSludge bool
	aiEnabled  bool
	asking     bool
	paused     bool
	bars       []float64
	playingID  int64
}

func New(cfg config.Config, cfgPath string, vault *store.Store) Model {
	renderer, _ := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(80))
	input := textinput.New()
	input.Placeholder = "ask the active article"
	input.CharLimit = 500
	input.Prompt = "interrogate> "
	return Model{
		cfg:        cfg,
		cfgPath:    cfgPath,
		store:      vault,
		styles:     newStyles(),
		renderer:   renderer,
		ai:         llm.New(cfg.Active()),
		input:      input,
		focus:      focusFeeds,
		hideSludge: cfg.UI.HideSludge,
		status:     "r refresh | tab focus | enter read/play | ctrl+a ask | ctrl+t triage",
		aiEnabled:  llm.New(cfg.Active()).Available(context.Background()),
	}
}

func (m Model) Init() tea.Cmd {
	return seedFeedsCmd(m.store, m.cfg.Feeds)
}
