package feed

import (
	"bytes"
	"encoding/xml"
	"errors"
	"strings"
	"time"
)

func Parse(body []byte) (Feed, error) {
	root, err := rootName(body)
	if err != nil {
		return Feed{}, err
	}
	switch strings.ToLower(root) {
	case "rss":
		return parseRSS(body)
	case "feed":
		return parseAtom(body)
	default:
		return Feed{}, errors.New("unsupported feed format")
	}
}

func rootName(body []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(body))
	for {
		token, err := dec.Token()
		if err != nil {
			return "", err
		}
		if start, ok := token.(xml.StartElement); ok {
			return start.Name.Local, nil
		}
	}
}

type rssDoc struct {
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title string    `xml:"title"`
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	GUID        string       `xml:"guid"`
	Title       string       `xml:"title"`
	Link        string       `xml:"link"`
	PubDate     string       `xml:"pubDate"`
	Description string       `xml:"description"`
	Content     string       `xml:"encoded"`
	Enclosure   rssEnclosure `xml:"enclosure"`
}

type rssEnclosure struct {
	URL  string `xml:"url,attr"`
	Type string `xml:"type,attr"`
}

func parseRSS(body []byte) (Feed, error) {
	var doc rssDoc
	if err := xml.Unmarshal(body, &doc); err != nil {
		return Feed{}, err
	}
	f := Feed{Title: cleanText(doc.Channel.Title), Type: "rss"}
	for _, src := range doc.Channel.Items {
		html := firstNonEmpty(src.Content, src.Description)
		item := Item{
			GUID:            firstNonEmpty(src.GUID, src.Link, src.Title),
			Title:           cleanText(src.Title),
			Link:            strings.TrimSpace(src.Link),
			PublishedAt:     parseDate(src.PubDate),
			ContentHTML:     html,
			ContentMarkdown: HTMLToMarkdown(html),
			EnclosureURL:    strings.TrimSpace(src.Enclosure.URL),
			EnclosureType:   strings.TrimSpace(src.Enclosure.Type),
		}
		f.Items = append(f.Items, item)
	}
	return f, nil
}

type atomDoc struct {
	Title   string      `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	ID        string     `xml:"id"`
	Title     string     `xml:"title"`
	Updated   string     `xml:"updated"`
	Published string     `xml:"published"`
	Summary   string     `xml:"summary"`
	Content   string     `xml:"content"`
	Links     []atomLink `xml:"link"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

func parseAtom(body []byte) (Feed, error) {
	var doc atomDoc
	if err := xml.Unmarshal(body, &doc); err != nil {
		return Feed{}, err
	}
	f := Feed{Title: cleanText(doc.Title), Type: "atom"}
	for _, src := range doc.Entries {
		html := firstNonEmpty(src.Content, src.Summary)
		link, encURL, encType := atomLinks(src.Links)
		item := Item{
			GUID:            firstNonEmpty(src.ID, link, src.Title),
			Title:           cleanText(src.Title),
			Link:            link,
			PublishedAt:     parseDate(firstNonEmpty(src.Published, src.Updated)),
			ContentHTML:     html,
			ContentMarkdown: HTMLToMarkdown(html),
			EnclosureURL:    encURL,
			EnclosureType:   encType,
		}
		f.Items = append(f.Items, item)
	}
	return f, nil
}

func atomLinks(links []atomLink) (string, string, string) {
	var page, encURL, encType string
	for _, link := range links {
		if link.Rel == "enclosure" || strings.HasPrefix(link.Type, "audio/") {
			encURL, encType = strings.TrimSpace(link.Href), strings.TrimSpace(link.Type)
			continue
		}
		if page == "" && (link.Rel == "" || link.Rel == "alternate") {
			page = strings.TrimSpace(link.Href)
		}
	}
	return page, encURL, encType
}

func parseDate(value string) time.Time {
	value = strings.TrimSpace(value)
	for _, layout := range []string{time.RFC1123Z, time.RFC1123, time.RFC3339, time.RFC822Z, time.RFC822} {
		if t, err := time.Parse(layout, value); err == nil {
			return t
		}
	}
	return time.Time{}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func cleanText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
