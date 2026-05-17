package feed

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

func FetchGopher(ctx context.Context, raw string) (Feed, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return Feed{}, err
	}
	port := u.Port()
	if port == "" {
		port = "70"
	}
	selector := strings.TrimPrefix(u.EscapedPath(), "/")
	if selector != "" {
		selector, _ = url.PathUnescape(selector)
		if len(selector) > 1 && isGopherType(selector[0]) {
			selector = selector[1:]
		}
	}
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(u.Hostname(), port))
	if err != nil {
		return Feed{}, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))
	if _, err := fmt.Fprintf(conn, "%s\r\n", selector); err != nil {
		return Feed{}, err
	}
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 4096), 2<<20)
	var lines []string
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if line == "." {
			break
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return Feed{}, err
	}
	return parseGopher(raw, u.Hostname(), port, lines), nil
}

func parseGopher(raw, host, port string, lines []string) Feed {
	title := "gopher://" + host
	items := make([]Item, 0, len(lines))
	var text []string
	for i, line := range lines {
		if line == "" {
			text = append(text, "")
			continue
		}
		fields := strings.Split(line[1:], "\t")
		if len(fields) >= 4 && isGopherType(line[0]) {
			item := gopherMenuItem(line[0], fields, i, host, port)
			if item.Title != "" {
				items = append(items, item)
			}
			continue
		}
		text = append(text, line)
	}
	if len(items) == 0 {
		body := strings.Join(text, "\n")
		items = append(items, Item{
			GUID:            raw,
			Title:           title,
			Link:            raw,
			ContentMarkdown: body,
			ContentHTML:     body,
			PublishedAt:     time.Now(),
		})
	}
	return Feed{Title: title, Type: "gopher", Items: items}
}

func gopherMenuItem(kind byte, fields []string, idx int, fallbackHost, fallbackPort string) Item {
	label := strings.TrimSpace(fields[0])
	selector := fields[1]
	host := firstNonEmpty(fields[2], fallbackHost)
	port := firstNonEmpty(fields[3], fallbackPort)
	link := "gopher://" + host
	if port != "" && port != "70" {
		link += ":" + port
	}
	link += "/" + string(kind) + selector
	body := fmt.Sprintf("Type: %s\nSelector: %s\nHost: %s\nPort: %s", gopherKind(kind), selector, host, port)
	return Item{
		GUID:            fmt.Sprintf("%s#%d:%s", link, idx, label),
		Title:           label,
		Link:            link,
		ContentMarkdown: body,
		ContentHTML:     body,
		PublishedAt:     time.Now(),
	}
}

func isGopherType(kind byte) bool {
	return strings.ContainsRune("013456789gIh", rune(kind))
}

func gopherKind(kind byte) string {
	switch kind {
	case '0':
		return "text"
	case '1':
		return "directory"
	case '7':
		return "search"
	case '8':
		return "telnet"
	case '9':
		return "binary"
	case 'g', 'I':
		return "image"
	case 'h':
		return "html"
	default:
		return "gopher"
	}
}
