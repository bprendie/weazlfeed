package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bprendie/weazlfeed/internal/app"
	"github.com/bprendie/weazlfeed/internal/auth"
	"github.com/bprendie/weazlfeed/internal/config"
	"github.com/bprendie/weazlfeed/internal/llm"
	"github.com/bprendie/weazlfeed/internal/store"
)

func main() {
	all := flag.Bool("all", true, "refresh all feeds")
	flag.Parse()
	_ = all
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
	feeds, err := vault.Feeds()
	if err != nil {
		fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	result, err := app.Refresh(ctx, vault, feeds, llm.New(cfg.Active()))
	if err != nil {
		fatal(err)
	}
	fmt.Printf("Refresh complete: checked %d new %d failed %d\n", result.Checked, result.Added, result.Failed)
	for _, detail := range result.Details {
		state := fmt.Sprintf("status %d", detail.Status)
		if detail.NotModified {
			state = "not modified"
		}
		if detail.Err != "" {
			state = "failed: " + detail.Err
		}
		fmt.Printf("%-4d %s - %s\n", detail.Added, state, detail.Title)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "weazlfeed-refresh: %v\n", err)
	os.Exit(1)
}
