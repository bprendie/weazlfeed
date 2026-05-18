package store

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateLockEncryptsExistingRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.sqlite3")
	vault, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer vault.Close()

	if _, err := vault.db.Exec(`
		INSERT INTO feeds(title, url, type, section, folder, category)
		VALUES('Plain Feed', 'https://example.com/feed.xml', 'rss', 'News', 'General', 'GENERAL')
	`); err != nil {
		t.Fatal(err)
	}
	if _, err := vault.db.Exec(`
		INSERT INTO items(feed_id, guid, title, link, content_html, content_markdown, enclosure_url, enclosure_type)
		VALUES(1, 'guid-1', 'Plain Item', 'https://example.com/item', '<p>Secret</p>', 'Secret', 'https://example.com/audio.mp3', 'audio/mpeg')
	`); err != nil {
		t.Fatal(err)
	}

	if err := vault.CreateLock("test-password"); err != nil {
		t.Fatal(err)
	}
	itemForAI := Item{ID: 1, Title: "Plain Item", ContentMarkdown: "Secret"}
	if err := vault.SaveAIOutput(itemForAI, "triage", "prompt", "response"); err != nil {
		t.Fatal(err)
	}

	var rawTitle, rawURL, rawItem, rawAI string
	if err := vault.db.QueryRow(`SELECT title, url FROM feeds WHERE id = 1`).Scan(&rawTitle, &rawURL); err != nil {
		t.Fatal(err)
	}
	if err := vault.db.QueryRow(`SELECT content_markdown FROM items WHERE id = 1`).Scan(&rawItem); err != nil {
		t.Fatal(err)
	}
	if err := vault.db.QueryRow(`SELECT response FROM ai_outputs WHERE item_id = 1`).Scan(&rawAI); err != nil {
		t.Fatal(err)
	}
	for label, value := range map[string]string{"feed title": rawTitle, "feed url": rawURL, "item": rawItem, "ai": rawAI} {
		if !strings.HasPrefix(value, encryptedPrefix) {
			t.Fatalf("%s was not encrypted: %q", label, value)
		}
	}

	feeds, err := vault.Feeds()
	if err != nil {
		t.Fatal(err)
	}
	if got := feeds[0].URL; got != "https://example.com/feed.xml" {
		t.Fatalf("decrypted feed URL = %q", got)
	}
	item, err := vault.Item(1)
	if err != nil {
		t.Fatal(err)
	}
	if item.ContentMarkdown != "Secret" || item.EnclosureURL != "https://example.com/audio.mp3" {
		t.Fatalf("decrypted item mismatch: %#v", item)
	}
	out, err := vault.AIOutput(1, "triage")
	if err != nil {
		t.Fatal(err)
	}
	if out.Response != "response" {
		t.Fatalf("decrypted AI output = %q", out.Response)
	}
}
