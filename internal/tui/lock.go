package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) lockView() string {
	width := max(40, m.width)
	header := m.header(width)
	title := "Unlock vault"
	if m.lockMode == lockCreate {
		title = "Create vault"
	}
	if m.lockMode == lockConfirm {
		title = "Confirm vault"
	}
	body := title + "\n\n" + m.input.View()
	if m.err != "" {
		body += "\n\n" + m.styles.error.Render(m.err)
	}
	panelW := clampInt(width-8, 32, 76)
	box := m.styles.panel.Width(panelW).Render(body)
	return m.styles.frame.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, header, box))
}

func (m Model) updateLock(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if ok {
		switch key.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m.submitLock()
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) submitLock() (tea.Model, tea.Cmd) {
	password := m.input.Value()
	if strings.TrimSpace(password) == "" {
		m.err = "password is required"
		return m, nil
	}
	switch m.lockMode {
	case lockUnlock:
		if err := m.store.Unlock(password); err != nil {
			m.err = err.Error()
			m.input.SetValue("")
			return m, nil
		}
		return m.finishUnlock()
	case lockCreate:
		m.pendingPass = password
		m.input.SetValue("")
		m.input.Placeholder = "confirm vault password"
		m.lockMode = lockConfirm
		m.err = ""
		return m, nil
	case lockConfirm:
		if password != m.pendingPass {
			m.err = "passwords do not match"
			m.input.SetValue("")
			return m, nil
		}
		if err := m.store.CreateLock(password); err != nil {
			m.err = err.Error()
			return m, nil
		}
		return m.finishUnlock()
	default:
		return m, nil
	}
}

func (m Model) finishUnlock() (tea.Model, tea.Cmd) {
	m.lockMode = lockOpen
	m.pendingPass = ""
	m.input.SetValue("")
	m.input.EchoMode = textinput.EchoNormal
	m.input.Placeholder = "ask the active article"
	m.input.Prompt = "interrogate> "
	m.input.Blur()
	m.err = ""
	m.status = "vault unlocked"
	return m, seedFeedsCmd(m.store, m.cfg.Feeds)
}
