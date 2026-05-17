package tui

import (
	"github.com/bprendie/weazlfeed/internal/audio"
	"github.com/bprendie/weazlfeed/internal/store"
)

type feedsMsg struct {
	feeds []store.Feed
	err   error
}

type itemsMsg struct {
	items []store.Item
	err   error
}

type fetchMsg struct {
	added int
	err   error
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
