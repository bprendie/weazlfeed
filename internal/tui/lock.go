package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
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
	if m.unlocking {
		body += "\n\n" + gradientStatus(m.spinner.View()+" decrypting vault")
	}
	if m.err != "" {
		body += "\n\n" + m.styles.error.Render(m.err)
	}
	panelW := clampInt(width-8, 32, 76)
	box := m.styles.panel.Width(panelW).Render(body)
	return m.styles.frame.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, header, box))
}

func (m Model) updateLock(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if !m.unlocking {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case lockMsg:
		m.unlocking = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.input.SetValue("")
			m.input.Focus()
			return m, nil
		}
		return m.finishUnlock()
	}
	if m.unlocking {
		return m, nil
	}
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
		m.unlocking = true
		m.err = ""
		m.input.Blur()
		m.input.SetValue("")
		return m, tea.Batch(unlockVaultCmd(m, password), m.spinner.Tick)
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
		m.unlocking = true
		m.err = ""
		m.input.Blur()
		m.input.SetValue("")
		return m, tea.Batch(createVaultCmd(m, password), m.spinner.Tick)
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

func unlockVaultCmd(m Model, password string) tea.Cmd {
	return func() tea.Msg {
		return lockMsg{err: m.store.Unlock(password)}
	}
}

func createVaultCmd(m Model, password string) tea.Cmd {
	return func() tea.Msg {
		return lockMsg{err: m.store.CreateLock(password)}
	}
}
