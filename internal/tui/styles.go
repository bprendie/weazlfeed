package tui

import "github.com/charmbracelet/lipgloss"

const (
	crushPink   = lipgloss.Color("#F25D94")
	crushPurple = lipgloss.Color("#7D56F4")
	crushMint   = lipgloss.Color("#04B575")
	crushGold   = lipgloss.Color("#F7D774")
	ink         = lipgloss.Color("#FAFAFA")
	muted       = lipgloss.Color("#8E8E93")
	panel       = lipgloss.Color("#181820")
	border      = lipgloss.Color("#3D315B")
)

type styles struct {
	frame    lipgloss.Style
	header   lipgloss.Style
	panel    lipgloss.Style
	status   lipgloss.Style
	help     lipgloss.Style
	selected lipgloss.Style
	item     lipgloss.Style
	error    lipgloss.Style
	badge    lipgloss.Style
}

func newStyles() styles {
	return styles{
		frame:  lipgloss.NewStyle().Foreground(ink).Background(lipgloss.Color("#0D0D12")).Padding(0, 2),
		header: lipgloss.NewStyle().Foreground(crushPink).Bold(true),
		panel: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(border).
			Background(panel).
			Padding(0, 1).
			Margin(0),
		status:   lipgloss.NewStyle().Foreground(crushMint).Bold(true),
		help:     lipgloss.NewStyle().Foreground(muted),
		selected: lipgloss.NewStyle().Foreground(crushPink).Bold(true),
		item:     lipgloss.NewStyle().Foreground(ink),
		error:    lipgloss.NewStyle().Foreground(crushGold).Bold(true),
		badge:    lipgloss.NewStyle().Foreground(crushMint).Bold(true),
	}
}
