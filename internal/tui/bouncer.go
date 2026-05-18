package tui

import (
	"context"
	"strings"
	"time"

	"github.com/bprendie/weazlfeed/internal/llm"
	"github.com/bprendie/weazlfeed/internal/store"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateBouncer(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.bouncerInput {
		return m.updateBouncerInput(msg)
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc", "ctrl+b":
			m.bouncerOpen = false
			m.status = "bouncer closed"
		case "n":
			m.bouncerInput = true
			m.input.SetValue("")
			m.input.Placeholder = "flag SEO sludge, crypto shilling, low-signal AI filler..."
			m.input.Prompt = "rule> "
			m.input.Focus()
		case "d", "ctrl+d":
			if len(m.bouncerRules) > 0 {
				rule := m.bouncerRules[m.bouncerCursor]
				return m, deleteBouncerRuleCmd(m.store, rule.ID)
			}
		case "j", "down":
			m.moveBouncer(1)
		case "k", "up":
			m.moveBouncer(-1)
		case "s":
			return m.scanSelectedWithBouncer()
		}
	}
	return m, nil
}

func (m Model) handleBouncerRulesMsg(msg bouncerRulesMsg) (tea.Model, tea.Cmd) {
	m.err = errText(msg.err)
	if msg.err == nil {
		m.bouncerRules = msg.rules
	}
	return m, nil
}

func (m Model) handleBouncerActionMsg(msg bouncerActionMsg) (tea.Model, tea.Cmd) {
	m.err = errText(msg.err)
	if msg.err == nil {
		m.bouncerRules = msg.rules
		m.bouncerInput = false
		m.input.Blur()
		m.input.SetValue("")
		m.input.Prompt = "interrogate> "
		m.status = "bouncer rules updated"
	}
	return m, nil
}

func (m Model) handleBouncerScanMsg(msg bouncerScanMsg) (tea.Model, tea.Cmd) {
	m.bouncerScanning = false
	m.err = errText(msg.err)
	if msg.err == nil {
		m.applyBouncerScan(msg)
		m.status = "bouncer passed: " + msg.scan.Title
		if msg.scan.Flagged {
			m.status = "bouncer flagged sludge: " + msg.scan.Title
		}
	}
	return m, nil
}

func (m *Model) applyBouncerScan(msg bouncerScanMsg) {
	if msg.scan.ItemID == 0 {
		return
	}
	for i := range m.items {
		if m.items[i].ID == msg.scan.ItemID {
			m.items[i].SludgeFlag = msg.scan.Flagged
			m.items[i].SludgeChecked = true
			m.itemCache[m.items[i].FeedID] = m.items
			return
		}
	}
}

func (m Model) updateBouncerInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc", "ctrl+c":
			m.bouncerInput = false
			m.input.Blur()
			m.input.SetValue("")
			m.input.Prompt = "interrogate> "
			return m, nil
		case "enter":
			prompt := strings.TrimSpace(m.input.Value())
			if prompt == "" {
				return m, nil
			}
			return m, addBouncerRuleCmd(m.store, prompt)
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) moveBouncer(delta int) {
	if len(m.bouncerRules) == 0 {
		m.bouncerCursor = 0
		return
	}
	m.bouncerCursor = clampInt(m.bouncerCursor+delta, 0, len(m.bouncerRules)-1)
}

func (m Model) scanSelectedWithBouncer() (tea.Model, tea.Cmd) {
	if len(m.bouncerRules) == 0 {
		m.status = "add a bouncer rule first"
		return m, nil
	}
	if !m.aiEnabled {
		m.status = "ai offline"
		return m, nil
	}
	item, ok := m.currentAIItem()
	if !ok {
		m.status = "open an item before scanning"
		return m, nil
	}
	m.bouncerScanning = true
	m.status = "bouncer scanning: " + item.Title
	return m, tea.Batch(scanBouncerItemCmd(m.store, m.ai, item, m.bouncerRules), m.spinner.Tick)
}

func loadBouncerRulesCmd(vault *store.Store) tea.Cmd {
	return func() tea.Msg {
		rules, err := vault.Rules()
		return bouncerRulesMsg{rules: rules, err: err}
	}
}

func addBouncerRuleCmd(vault *store.Store, prompt string) tea.Cmd {
	return func() tea.Msg {
		err := vault.AddRule(prompt)
		rules, rulesErr := vault.Rules()
		return bouncerActionMsg{rules: rules, err: firstErr(err, rulesErr)}
	}
}

func deleteBouncerRuleCmd(vault *store.Store, id int64) tea.Cmd {
	return func() tea.Msg {
		err := vault.DeleteRule(id)
		rules, rulesErr := vault.Rules()
		return bouncerActionMsg{rules: rules, err: firstErr(err, rulesErr)}
	}
}

func scanBouncerItemCmd(vault *store.Store, ai llm.Client, item store.Item, rules []store.BouncerRule) tea.Cmd {
	return func() tea.Msg {
		if item.ContentMarkdown == "" && item.ID != 0 {
			full, err := vault.Item(item.ID)
			if err != nil {
				return bouncerScanMsg{err: err}
			}
			item = full
		}
		prompts := make([]string, 0, len(rules))
		for _, rule := range rules {
			prompts = append(prompts, rule.Prompt)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		flagged, err := ai.FlagSludge(ctx, trimAIInput(item.ContentMarkdown), prompts)
		if err != nil {
			return bouncerScanMsg{err: err}
		}
		if item.ID != 0 {
			err = vault.SetSludge(item.ID, flagged)
		}
		return bouncerScanMsg{scan: store.BouncerScan{ItemID: item.ID, Title: item.Title, Flagged: flagged}, err: err}
	}
}
