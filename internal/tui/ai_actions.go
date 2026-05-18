package tui

import (
	"strings"

	"github.com/bprendie/weazlfeed/internal/store"
)

func (m *Model) showAIOutput(msg aiMsg) {
	if m.articleMode == articleNormal {
		m.savedRawArticle = m.rawArticle
		m.savedArticle = m.article
	}
	title := "AI TRIAGE"
	m.articleMode = articleTriage
	if msg.kind == "ask" {
		title = "INTERROGATION ROOM"
		m.articleMode = articleAsk
	}
	body := "# " + title + "\n\n"
	if msg.question != "" {
		body += "**Q:** " + msg.question + "\n\n"
	}
	body += strings.TrimSpace(msg.text)
	m.rawArticle = body
	m.article = m.renderMarkdown(body)
	m.stageScroll = 0
	m.focus = focusArticle
	if msg.cached {
		m.status = "cached local extraction"
	} else {
		m.status = "local extraction complete"
	}
}

func (m *Model) showInterrogation(out store.AIOutput) {
	body := "# INTERROGATION ROOM\n\n"
	body += "**Article:** " + firstText(out.ItemTitle, "saved article") + "\n\n"
	if out.Prompt != "" {
		body += "**Q:** " + out.Prompt + "\n\n"
	}
	body += strings.TrimSpace(out.Response)
	if out.ItemContent != "" {
		body += "\n\n---\n\n## Article Snapshot\n\n" + out.ItemContent
	}
	m.rawArticle = body
	m.article = m.renderMarkdown(body)
	m.savedRawArticle = ""
	m.savedArticle = ""
	m.articleMode = articleAsk
	m.status = "saved interrogation"
}

func (m *Model) restoreArticle() {
	if m.savedRawArticle != "" || m.savedArticle != "" {
		m.rawArticle = m.savedRawArticle
		m.article = m.savedArticle
		m.savedRawArticle = ""
		m.savedArticle = ""
	}
	m.articleMode = articleNormal
	m.stageScroll = 0
	m.status = "article restored"
}
