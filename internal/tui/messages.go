package tui

import (
	"github.com/bprendie/weazlfeed/internal/audio"
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

type fetchMsg struct {
	checked int
	added   int
	failed  int
	err     error
}

type aiMsg struct {
	text string
	err  error
}

type meterMsg audio.Sample

type playheadTickMsg struct{}

type articleMsg struct {
	text string
	err  error
}

type podcastSearchMsg struct {
	results []podcast.Result
	err     error
}
