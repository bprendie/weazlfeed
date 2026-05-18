package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bprendie/weazlfeed/internal/auth"
	"github.com/bprendie/weazlfeed/internal/config"
	"github.com/bprendie/weazlfeed/internal/store"
)

func main() {
	days := flag.Int("days", 30, "delete items older than this many days")
	keepUnread := flag.Bool("keep-unread", true, "keep unread items")
	keepPlayhead := flag.Bool("keep-playhead", true, "keep podcast items with saved playback")
	vacuum := flag.Bool("vacuum", true, "vacuum the database after pruning")
	flag.Parse()
	if *days <= 0 {
		fatal(fmt.Errorf("days must be greater than zero"))
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
	result, err := vault.Prune(store.PruneOptions{
		Before:       time.Now().AddDate(0, 0, -*days),
		KeepUnread:   *keepUnread,
		KeepPlayhead: *keepPlayhead,
		Vacuum:       *vacuum,
	})
	if err != nil {
		fatal(err)
	}
	fmt.Printf("Pruned %d items older than %d days\n", result.Deleted, *days)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "weazlfeed-prune: %v\n", err)
	os.Exit(1)
}
