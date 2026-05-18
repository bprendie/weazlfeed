package tui

import (
	"context"

	"github.com/bprendie/weazlfeed/internal/audio"
	"github.com/bprendie/weazlfeed/internal/config"
	"github.com/bprendie/weazlfeed/internal/llm"
	"github.com/bprendie/weazlfeed/internal/podcast"
	"github.com/bprendie/weazlfeed/internal/store"
	"github.com/charmbracelet/bubbles/spinner"
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
	cfg          config.Config
	cfgPath      string
	store        *store.Store
	styles       styles
	renderer     *glamour.TermRenderer
	ai           llm.Client
	player       audio.Player
	meter        *audio.Meter
	input        textinput.Model
	spinner      spinner.Model
	focus        focus
	width        int
	height       int
	feeds        []store.Feed
	folders      []store.Folder
	items        []store.Item
	podcasts     []podcast.Result
	feedCursor   int
	itemCursor   int
	feedScroll   int
	itemScroll   int
	stageScroll  int
	article      string
	status       string
	err          string
	hideSludge   bool
	aiEnabled    bool
	asking       bool
	folderInput  bool
	podcastInput bool
	paused       bool
	bars         []float64
	playingID    int64
	refreshing   bool
	pickedFeedID int64
}

func New(cfg config.Config, cfgPath string, vault *store.Store) Model {
	renderer, _ := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(80))
	input := textinput.New()
	input.Placeholder = "ask the active article"
	input.CharLimit = 500
	input.Prompt = "interrogate> "
	spin := spinner.New()
	spin.Spinner = spinner.Line
	return Model{
		cfg:        cfg,
		cfgPath:    cfgPath,
		store:      vault,
		styles:     newStyles(),
		renderer:   renderer,
		ai:         llm.New(cfg.Active()),
		input:      input,
		spinner:    spin,
		focus:      focusFeeds,
		hideSludge: cfg.UI.HideSludge,
		status:     "r refresh source | R refresh all | space pick/drop | n folder",
		aiEnabled:  llm.New(cfg.Active()).Available(context.Background()),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(seedFeedsCmd(m.store, m.cfg.Feeds), m.spinner.Tick)
}
