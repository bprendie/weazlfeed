package tui

import (
	"context"
	"time"

	"github.com/bprendie/weazlfeed/internal/audio"
	"github.com/bprendie/weazlfeed/internal/config"
	"github.com/bprendie/weazlfeed/internal/llm"
	"github.com/bprendie/weazlfeed/internal/podcast"
	"github.com/bprendie/weazlfeed/internal/store"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
)

type focus int

const (
	focusFeeds focus = iota
	focusItems
	focusArticle
)

type lockMode int

const (
	lockOpen lockMode = iota
	lockUnlock
	lockCreate
	lockConfirm
)

type articleMode int

const (
	articleNormal articleMode = iota
	articleTriage
	articleAsk
)

type Model struct {
	cfg              config.Config
	cfgPath          string
	store            *store.Store
	styles           styles
	ai               llm.Client
	player           audio.Player
	meter            *audio.Meter
	input            textinput.Model
	spinner          spinner.Model
	focus            focus
	width            int
	height           int
	feeds            []store.Feed
	folders          []store.Folder
	items            []store.Item
	gopherStack      [][]store.Item
	itemCache        map[int64][]store.Item
	podcasts         []podcast.Result
	podcastCursor    int
	podcastScroll    int
	podcastSearching bool
	sourceCursor     int
	feedCursor       int
	itemCursor       int
	feedScroll       int
	itemScroll       int
	stageScroll      int
	rawArticle       string
	article          string
	savedArticle     string
	savedRawArticle  string
	articleMode      articleMode
	status           string
	err              string
	hideSludge       bool
	aiEnabled        bool
	asking           bool
	folderInput      bool
	podcastInput     bool
	urlInput         bool
	helpOpen         bool
	confirmDelete    bool
	deleteFeedID     int64
	deleteFeedTitle  string
	paused           bool
	energy           audio.Sample
	visualizer       Visualizer
	playingID        int64
	playingURL       string
	playingTitle     string
	playingTotal     int
	revealFeedID     int64
	revealSection    string
	revealFolder     string
	refreshing       bool
	rendering        bool
	aiWorking        bool
	aiAction         string
	aiStartedAt      time.Time
	pickedFeedID     int64
	lockMode         lockMode
	pendingPass      string
	unlocking        bool
}

func New(cfg config.Config, cfgPath string, vault *store.Store) Model {
	input := textinput.New()
	input.Placeholder = "ask the active article"
	input.CharLimit = 500
	input.Prompt = "interrogate> "
	spin := spinner.New()
	spin.Spinner = spinner.Line
	mode := lockUnlock
	if has, err := vault.HasLock(); err == nil && !has {
		mode = lockCreate
	}
	if mode != lockOpen {
		input.Placeholder = "vault password"
		input.Prompt = "vault> "
		input.EchoMode = textinput.EchoPassword
		input.Focus()
	}
	return Model{
		cfg:        cfg,
		cfgPath:    cfgPath,
		store:      vault,
		styles:     newStyles(),
		ai:         llm.New(cfg.Active()),
		itemCache:  map[int64][]store.Item{},
		input:      input,
		spinner:    spin,
		focus:      focusFeeds,
		hideSludge: cfg.UI.HideSludge,
		status:     "r refresh source | R refresh all | space pick/drop | n folder",
		aiEnabled:  llm.New(cfg.Active()).Available(context.Background()),
		lockMode:   mode,
		visualizer: NewVisualizer(harmonica.FPS(30)),
	}
}

func (m Model) Init() tea.Cmd {
	if m.lockMode != lockOpen {
		return textinput.Blink
	}
	return tea.Batch(seedFeedsCmd(m.store, m.cfg.Feeds), m.spinner.Tick)
}
