package feed

import "time"

type Feed struct {
	Title string
	Type  string
	Items []Item
}

type Item struct {
	GUID            string
	Title           string
	Link            string
	PublishedAt     time.Time
	ContentHTML     string
	ContentMarkdown string
	EnclosureURL    string
	EnclosureType   string
}
