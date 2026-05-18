package tui

import (
	"fmt"
	"time"
)

var aiThinkingPhrases = []string{
	"hacking_the_gibson",
	"jacking_into_the_matrix",
	"breaching_corporate_ice",
	"overclocking_neural_link",
	"tracing_the_uplink",
	"decrypting_sector_7",
	"sniffing_data_packets",
	"bypassing_firewall_01",
	"rerouting_the_mainframe",
	"mapping_the_grid",
	"ghosting_the_network",
	"prying_open_the_vault",
	"optimizing_cyberdeck",
	"wheezing_the_juice",
	"chilling_the_tokens",
	"taxing_the_gig",
}

func (m Model) aiWorkingText() string {
	if len(aiThinkingPhrases) == 0 || m.aiStartedAt.IsZero() {
		return firstText(m.aiAction, "local_model_working")
	}
	phase := min(3, int(time.Since(m.aiStartedAt)/(20*time.Second)))
	start := int((m.aiStartedAt.UnixNano() / int64(time.Millisecond)) % int64(len(aiThinkingPhrases)))
	phrase := aiThinkingPhrases[(start+phase)%len(aiThinkingPhrases)]
	return firstText(m.aiAction, "ai") + " | " + phrase
}

func (m Model) aiMetricsView(width int) string {
	budget := m.cfg.Active().ContextWindow
	if budget <= 0 {
		budget = 32768
	}
	m.contextBar.Width = min(28, max(10, width/4))
	pct := min(1.0, float64(m.aiReqIn)/float64(budget))
	elapsed := 0
	if !m.aiStartedAt.IsZero() {
		elapsed = int(time.Since(m.aiStartedAt).Seconds())
	}
	out := "..."
	if !m.aiWorking {
		out = intText(m.aiReqOut)
	}
	text := fmt.Sprintf("ctx %s %d/%d  in %d  out %s  %ds", m.contextBar.ViewAs(pct), m.aiReqIn, budget, m.aiReqIn, out, elapsed)
	return m.styles.help.Render(truncate(text, max(20, width)))
}
