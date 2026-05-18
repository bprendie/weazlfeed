package main

import (
	"fmt"
	"os"

	"github.com/bprendie/weazlfeed/internal/auth"
	"github.com/bprendie/weazlfeed/internal/config"
	"github.com/bprendie/weazlfeed/internal/store"
)

func main() {
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
	fmt.Println("Vault unlocked and encrypted-at-rest migration complete.")
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "weazlfeed-vault: %v\n", err)
	os.Exit(1)
}
