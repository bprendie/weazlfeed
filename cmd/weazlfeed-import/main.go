package main

import (
	"fmt"
	"os"

	"github.com/bprendie/weazlfeed/internal/config"
	"github.com/bprendie/weazlfeed/internal/feed"
	"github.com/bprendie/weazlfeed/internal/store"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: weazlfeed-import feeds.opml")
		os.Exit(2)
	}
	cfg, _, err := config.Load()
	if err != nil {
		fatal(err)
	}
	vault, err := store.Open(cfg.Database.Path)
	if err != nil {
		fatal(err)
	}
	defer vault.Close()
	f, err := os.Open(os.Args[1])
	if err != nil {
		fatal(err)
	}
	defer f.Close()
	feeds, err := feed.ReadOPML(f)
	if err != nil {
		fatal(err)
	}
	for _, src := range feeds {
		feedType := "rss"
		if feed.IsGopher(src.URL) {
			feedType = "gopher"
		}
		if _, err := vault.UpsertFeed(src.Title, src.URL, feedType, src.Section, src.Folder, src.Folder); err != nil {
			fatal(err)
		}
	}
	fmt.Printf("Imported %d feeds\n", len(feeds))
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "weazlfeed-import: %v\n", err)
	os.Exit(1)
}
