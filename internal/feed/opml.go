package feed

import (
	"encoding/xml"
	"io"
)

type OPML struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    OPMLHead `xml:"head"`
	Body    OPMLBody `xml:"body"`
}

type OPMLHead struct {
	Title string `xml:"title"`
}

type OPMLBody struct {
	Outlines []OPMLOutline `xml:"outline"`
}

type OPMLOutline struct {
	Text     string        `xml:"text,attr"`
	Title    string        `xml:"title,attr"`
	Type     string        `xml:"type,attr"`
	XMLURL   string        `xml:"xmlUrl,attr"`
	HTMLURL  string        `xml:"htmlUrl,attr"`
	Outlines []OPMLOutline `xml:"outline"`
}

type ImportFeed struct {
	Section string
	Folder  string
	Title   string
	URL     string
}

func ReadOPML(r io.Reader) ([]ImportFeed, error) {
	var doc OPML
	if err := xml.NewDecoder(r).Decode(&doc); err != nil {
		return nil, err
	}
	var feeds []ImportFeed
	for _, outline := range doc.Body.Outlines {
		walkOPML(outline, "", "", &feeds)
	}
	return feeds, nil
}

func walkOPML(outline OPMLOutline, section, folder string, feeds *[]ImportFeed) {
	label := firstNonEmpty(outline.Title, outline.Text)
	if outline.XMLURL != "" {
		*feeds = append(*feeds, ImportFeed{Section: firstNonEmpty(section, "News"), Folder: firstNonEmpty(folder, "General"), Title: label, URL: outline.XMLURL})
		return
	}
	if section == "" {
		section = label
	} else if folder == "" {
		folder = label
	}
	for _, child := range outline.Outlines {
		walkOPML(child, section, folder, feeds)
	}
}
