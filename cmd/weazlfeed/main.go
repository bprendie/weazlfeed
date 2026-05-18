package main

import (
	"fmt"
	"os"

	"github.com/bprendie/weazlfeed/internal/config"
	"github.com/bprendie/weazlfeed/internal/store"
	"github.com/bprendie/weazlfeed/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg, cfgPath, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	vault, err := store.Open(cfg.Database.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "vault: %v\n", err)
		os.Exit(1)
	}
	defer vault.Close()

	model := tui.New(cfg, cfgPath, vault)
	if _, err := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tui: %v\n", err)
		os.Exit(1)
	}
}
