package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bprendie/weazlfeed/internal/auth"
	"github.com/bprendie/weazlfeed/internal/config"
	"github.com/bprendie/weazlfeed/internal/podcast"
	"github.com/bprendie/weazlfeed/internal/store"
)

func main() {
	add := flag.Int("add", 0, "add result number to Podcasts/Search")
	limit := flag.Int("limit", 20, "maximum search results")
	flag.Parse()
	query := strings.TrimSpace(strings.Join(flag.Args(), " "))
	if query == "" {
		fmt.Fprintln(os.Stderr, "usage: weazlfeed-podcast-search [-add N] query")
		os.Exit(2)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	results, err := podcast.NewClient().Search(ctx, query, *limit)
	if err != nil {
		fatal(err)
	}
	if len(results) == 0 {
		fmt.Println("No podcasts found.")
		return
	}
	if *add > 0 {
		addPodcast(results, *add)
		return
	}
	for i, result := range results {
		fmt.Printf("%2d. %s\n", i+1, result.Title)
		if result.Author != "" {
			fmt.Printf("    by %s\n", result.Author)
		}
		if result.EpisodeCount > 0 {
			fmt.Printf("    episodes: %s\n", strconv.Itoa(result.EpisodeCount))
		}
		fmt.Printf("    feed: %s\n", result.FeedURL)
	}
	fmt.Println("")
	fmt.Println("Add one with: weazlfeed-podcast-search -add N " + strconv.Quote(query))
}

func addPodcast(results []podcast.Result, index int) {
	if index < 1 || index > len(results) {
		fatal(fmt.Errorf("result %d is out of range", index))
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
	if err := auth.UnlockOrCreate(vault); err != nil {
		fatal(err)
	}
	result := results[index-1]
	if _, err := vault.UpsertFeed(result.Title, result.FeedURL, "rss", "Podcasts", "Search", "Search"); err != nil {
		fatal(err)
	}
	fmt.Println("Added podcast: " + result.Title)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "weazlfeed-podcast-search: %v\n", err)
	os.Exit(1)
}
