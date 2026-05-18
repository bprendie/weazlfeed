package feed

import "time"

type Feed struct {
	Title        string
	Type         string
	Status       int
	ETag         string
	LastModified string
	NotModified  bool
	Items        []Item
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
	DurationSeconds int
}
