package tui

import (
	"github.com/bprendie/weazlfeed/internal/podcast"
	"github.com/bprendie/weazlfeed/internal/store"
)

type feedsMsg struct {
	feeds   []store.Feed
	folders []store.Folder
	err     error
}

type itemsMsg struct {
	feedID int64
	items  []store.Item
	err    error
}

type allItemsMsg struct {
	itemsByFeed map[int64][]store.Item
	err         error
}

type fetchMsg struct {
	checked int
	added   int
	failed  int
	err     error
}

type addFeedMsg struct {
	feedID  int64
	title   string
	section string
	folder  string
	added   int
	err     error
}

type deleteFeedMsg struct {
	feedID int64
	title  string
	err    error
}

type aiMsg struct {
	text string
	err  error
}

type playheadTickMsg struct{}

type audioTickMsg struct{}

type articleMsg struct {
	text string
	err  error
}

type gopherMsg struct {
	url   string
	items []store.Item
	text  string
	err   error
}

type readerMsg struct {
	item     store.Item
	raw      string
	rendered string
	err      error
}

type podcastSearchMsg struct {
	results []podcast.Result
	err     error
}
