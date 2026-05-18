package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "esc", "ctrl+k", "?", "f1":
		m.helpOpen = false
		m.status = "ready"
	}
	return m, nil
}

func (m Model) helpText() string {
	return `WeazlFeed help

Navigation
  j/k or arrows       Move within the active pane.
  tab                 Expand/cycle the active pane.
  enter               Open a source, item, folder, podcast result, or Gopher target.
  esc or left         Move back a pane, or back through Gopher menu history.
  pgup/pgdown         Page through the active pane.
  home/end            Jump to top or bottom.
  ctrl+k, ?, f1       Open or close this command screen.
  q or ctrl+c         Quit.

Sources and folders
  a                   Add a feed, podcast RSS URL, or Gopher URL to the selected folder.
  n                   Create a folder in the selected section.
  space               Pick up a source; move to a folder/source; space again to drop.
  enter on folder     Fold or unfold the folder.
  right               Expand the selected folder.
  r                   Refresh selected source.
  R                   Refresh all sources.
  h                   Hide or show items flagged as sludge.

Podcasts and audio
  p                   Open the podcast directory.
  /                   Focus podcast search inside the directory.
  enter or a          Subscribe the selected podcast result.
  enter on audio      Play or resume with mpv.
  space               Pause or resume audio when not picking/dropping a source.
  < or ,              Seek backward 10 seconds.
  > or .              Seek forward 30 seconds.
  esc                 Stop/close audio and save playhead.

Reader and local AI
  enter on article    Render the article.
  ctrl+a              Ask the local model about the active article.
  ctrl+t              Extract critical points from the active article.

Gopher
  gopher:// URLs are routed into the Gopher section automatically.
  Directory entries open as nested menus in the items pane.
  Text entries open in the reader.
  esc or left walks back through the Gopher menu stack.`
}
