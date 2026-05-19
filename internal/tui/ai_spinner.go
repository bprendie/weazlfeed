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

var renderThinkingPhrases = []string{
	"burning_html_off_the_wire",
	"wrapping_raw_signal",
	"deansi_frying_the_payload",
	"polishing_terminal_glyphs",
	"unrolling_the_scrollback",
	"taming_longform_static",
	"cutting_tracking_pixels",
	"compressing_the_feed_blob",
	"painting_neon_markdown",
	"warming_the_reader_tube",
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

func (m Model) renderWorkingText() string {
	if len(renderThinkingPhrases) == 0 || m.renderStartedAt.IsZero() {
		return firstText(m.renderAction, "rendering_reader")
	}
	phase := min(len(renderThinkingPhrases)-1, int(time.Since(m.renderStartedAt)/(3*time.Second)))
	start := int((m.renderStartedAt.UnixNano() / int64(time.Millisecond)) % int64(len(renderThinkingPhrases)))
	phrase := renderThinkingPhrases[(start+phase)%len(renderThinkingPhrases)]
	return firstText(m.renderAction, "rendering reader") + " | " + phrase
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
