package store

import (
	"database/sql"
	"sort"
	"time"
)

func (s *Store) UpsertFeed(title, url, feedType, section, folder, category string) (int64, error) {
	if section == "" {
		section = sectionFromType(feedType)
	}
	if folder == "" {
		folder = folderFromCategory(category)
	}
	category = firstText(category, folder, "General")
	if err := s.UpsertFolder(section, folder); err != nil {
		return 0, err
	}
	feedKey := s.feedKey(url)
	title, err := s.encryptText(title)
	if err != nil {
		return 0, err
	}
	encryptedURL, err := s.encryptText(url)
	if err != nil {
		return 0, err
	}
	section, err = s.encryptText(section)
	if err != nil {
		return 0, err
	}
	folder, err = s.encryptText(folder)
	if err != nil {
		return 0, err
	}
	category, err = s.encryptText(category)
	if err != nil {
		return 0, err
	}
	_, err = s.db.Exec(`
		INSERT INTO feeds(title, url, type, section, folder, category, feed_key) VALUES(?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(feed_key) DO UPDATE SET
			title=excluded.title,
			url=excluded.url,
			type=excluded.type,
			section=excluded.section,
			folder=excluded.folder,
			category=excluded.category
	`, title, encryptedURL, feedType, section, folder, category, feedKey)
	if err != nil {
		return 0, err
	}
	var id int64
	err = s.db.QueryRow(`SELECT id FROM feeds WHERE feed_key = ?`, feedKey).Scan(&id)
	return id, err
}

func (s *Store) Feeds() ([]Feed, error) {
	rows, err := s.db.Query(`
		SELECT f.id, f.title, f.url, f.type, f.section, f.folder, f.category, f.last_fetched,
			coalesce(f.etag, ''), coalesce(f.last_modified, ''), coalesce(f.last_error, ''),
			coalesce(f.last_status, 0),
			COUNT(CASE WHEN i.read_status = 0 THEN 1 END) AS unread
		FROM feeds f
		LEFT JOIN items i ON i.feed_id = f.id
		GROUP BY f.id
		ORDER BY
			CASE f.section WHEN 'News' THEN 0 WHEN 'Podcasts' THEN 1 WHEN 'Gopher' THEN 2 ELSE 3 END,
			lower(f.folder), lower(f.title)
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var feeds []Feed
	for rows.Next() {
		var f Feed
		var fetched sql.NullString
		if err := rows.Scan(&f.ID, &f.Title, &f.URL, &f.Type, &f.Section, &f.Folder, &f.Category,
			&fetched, &f.ETag, &f.LastModified, &f.LastError, &f.LastStatus, &f.Unread); err != nil {
			return nil, err
		}
		f.Title = s.decryptText(f.Title)
		f.URL = s.decryptText(f.URL)
		f.Section = s.decryptText(f.Section)
		f.Folder = s.decryptText(f.Folder)
		f.Category = s.decryptText(f.Category)
		f.LastError = s.decryptText(f.LastError)
		f.LastFetched = parseTime(fetched)
		feeds = append(feeds, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(feeds, func(i, j int) bool {
		return feedSortKey(feeds[i]) < feedSortKey(feeds[j])
	})
	return feeds, nil
}

func sectionFromType(feedType string) string {
	if feedType == "gopher" {
		return "Gopher"
	}
	return "News"
}

func folderFromCategory(category string) string {
	switch category {
	case "", "GENERAL":
		return "General"
	case "TECH":
		return "Tech"
	case "WORLD":
		return "World"
	case "SPORTS":
		return "Sports"
	case "MUSIC":
		return "Music"
	case "GOPHER":
		return "Directory"
	default:
		return category
	}
}

func firstText(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (s *Store) MarkFetched(feedID int64, when time.Time) error {
	_, err := s.db.Exec(`UPDATE feeds SET last_fetched = ? WHERE id = ?`, when.UTC().Format(time.RFC3339), feedID)
	return err
}

func (s *Store) SetFeedStatus(feedID int64, status int, etag, modified, lastError string) error {
	lastError, err := s.encryptText(lastError)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
		UPDATE feeds
		SET last_status = ?, etag = coalesce(nullif(?, ''), etag),
			last_modified = coalesce(nullif(?, ''), last_modified), last_error = ?
		WHERE id = ?
	`, status, etag, modified, lastError, feedID)
	return err
}

func (s *Store) MoveFeed(feedID int64, section, folder string) error {
	if err := s.UpsertFolder(section, folder); err != nil {
		return err
	}
	encSection, err := s.encryptText(section)
	if err != nil {
		return err
	}
	encFolder, err := s.encryptText(folder)
	if err != nil {
		return err
	}
	encCategory, err := s.encryptText(folder)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE feeds SET section = ?, folder = ?, category = ? WHERE id = ?`, encSection, encFolder, encCategory, feedID)
	return err
}

func (s *Store) DeleteFeed(feedID int64) error {
	_, err := s.db.Exec(`DELETE FROM feeds WHERE id = ?`, feedID)
	return err
}

func feedSortKey(f Feed) string {
	section := "3:" + f.Section
	switch f.Section {
	case "News":
		section = "0:News"
	case "Podcasts":
		section = "1:Podcasts"
	case "Gopher":
		section = "2:Gopher"
	}
	return section + "\x00" + f.Folder + "\x00" + f.Title
}
