package store

import "time"

type Feed struct {
	ID          int64
	Title       string
	URL         string
	Type        string
	Section     string
	Folder      string
	Category    string
	LastFetched time.Time
	Unread      int
}

type Item struct {
	ID              int64
	FeedID          int64
	GUID            string
	Title           string
	Link            string
	PublishedAt     time.Time
	ContentHTML     string
	ContentMarkdown string
	EnclosureURL    string
	EnclosureType   string
	ReadStatus      bool
	SludgeFlag      bool
	SludgeChecked   bool
	PlayheadSeconds int
}

type BouncerRule struct {
	ID     int64
	Prompt string
}
