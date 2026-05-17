package feed

import (
	"strings"

	"golang.org/x/net/html"
)

func HTMLToMarkdown(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	root, err := html.Parse(strings.NewReader("<div>" + raw + "</div>"))
	if err != nil {
		return cleanText(raw)
	}
	var b strings.Builder
	renderNode(&b, root, 0)
	return strings.TrimSpace(compactLines(b.String()))
}

func renderNode(b *strings.Builder, n *html.Node, depth int) {
	if n.Type == html.TextNode {
		text := cleanText(n.Data)
		if text != "" {
			writeSpace(b)
			b.WriteString(text)
		}
		return
	}
	if n.Type != html.ElementNode && n.Type != html.DocumentNode {
		return
	}
	switch n.Data {
	case "script", "style", "img", "picture", "source", "iframe":
		return
	case "h1", "h2", "h3":
		b.WriteString("\n\n")
		b.WriteString(strings.Repeat("#", headingDepth(n.Data)))
		b.WriteString(" ")
	case "p", "div", "section", "article", "br":
		b.WriteString("\n\n")
	case "li":
		b.WriteString("\n- ")
	case "blockquote":
		b.WriteString("\n\n> ")
	}
	if n.Data == "a" {
		renderLink(b, n, depth)
		return
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		renderNode(b, child, depth+1)
	}
}

func renderLink(b *strings.Builder, n *html.Node, depth int) {
	textStart := b.Len()
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		renderNode(b, child, depth+1)
	}
	href := attr(n, "href")
	if href == "" || b.Len() == textStart {
		return
	}
	b.WriteString(" (")
	b.WriteString(href)
	b.WriteString(")")
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return strings.TrimSpace(a.Val)
		}
	}
	return ""
}

func writeSpace(b *strings.Builder) {
	s := b.String()
	if s == "" || strings.HasSuffix(s, " ") || strings.HasSuffix(s, "\n") {
		return
	}
	b.WriteByte(' ')
}

func headingDepth(tag string) int {
	switch tag {
	case "h1":
		return 1
	case "h2":
		return 2
	default:
		return 3
	}
}

func compactLines(value string) string {
	lines := strings.Split(value, "\n")
	out := make([]string, 0, len(lines))
	blank := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if !blank {
				out = append(out, "")
			}
			blank = true
			continue
		}
		out = append(out, line)
		blank = false
	}
	return strings.Join(out, "\n")
}
