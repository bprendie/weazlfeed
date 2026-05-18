package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc", "ctrl+c":
			m.asking = false
			m.folderInput = false
			m.urlInput = false
			m.input.Blur()
			m.input.SetValue("")
			m.input.Prompt = "interrogate> "
			return m, nil
		case "enter":
			question := strings.TrimSpace(m.input.Value())
			folderInput := m.folderInput
			urlInput := m.urlInput
			m.asking = false
			m.folderInput = false
			m.urlInput = false
			m.input.Blur()
			m.input.SetValue("")
			m.input.Prompt = "interrogate> "
			if folderInput {
				return m.createFolder(question)
			}
			if urlInput {
				return m.addURL(question)
			}
			if question != "" {
				item, ok := m.currentAIItem()
				if !ok {
					return m, nil
				}
				m.aiWorking = true
				m.aiAction = "interrogating article"
				m.aiStartedAt = time.Now()
				m.aiReqIn = estimateTokens(item.ContentMarkdown) + estimateTokens(question)
				m.aiReqOut = 0
				m.status = "interrogating active article"
				return m, tea.Batch(aiCmd(m.store, m.ai, "ask", item, question), m.spinner.Tick, aiTickCmd())
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}
