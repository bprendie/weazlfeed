package tui

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/x/ansi"
)

const glamourMaxBytes = 24 * 1024

func (m Model) renderMarkdown(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	dims, _ := m.layout()
	width := max(20, panelContentWidth(m.styles.panel, dims.right)-2)
	if len(text) > glamourMaxBytes {
		return fastWrapMarkdown(text, width)
	}
	renderer, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(width))
	if err != nil {
		return fastWrapMarkdown(text, width)
	}
	rendered, err := renderer.Render(text)
	if err != nil {
		return fastWrapMarkdown(text, width)
	}
	return strings.TrimRight(rendered, "\n")
}

func fastWrapMarkdown(text string, width int) string {
	width = max(20, width)
	paragraphs := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(paragraphs)*2)
	for _, line := range paragraphs {
		line = strings.TrimSpace(line)
		if line == "" {
			if len(out) > 0 && out[len(out)-1] != "" {
				out = append(out, "")
			}
			continue
		}
		out = append(out, wrapLine(line, width)...)
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n")
}

func wrapLine(line string, width int) []string {
	if ansi.StringWidth(line) <= width {
		return []string{line}
	}
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	current := ""
	for _, word := range words {
		if current == "" {
			current = word
			continue
		}
		next := current + " " + word
		if ansi.StringWidth(next) <= width {
			current = next
			continue
		}
		lines = append(lines, current)
		current = word
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
