package store

import (
	"database/sql"
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
	_, err := s.db.Exec(`
		INSERT INTO feeds(title, url, type, section, folder, category) VALUES(?, ?, ?, ?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET title=excluded.title, type=excluded.type
	`, title, url, feedType, section, folder, category)
	if err != nil {
		return 0, err
	}
	var id int64
	err = s.db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, url).Scan(&id)
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
		f.LastFetched = parseTime(fetched)
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
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
	_, err := s.db.Exec(`
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
	_, err := s.db.Exec(`UPDATE feeds SET section = ?, folder = ?, category = ? WHERE id = ?`, section, folder, folder, feedID)
	return err
}

func (s *Store) DeleteFeed(feedID int64) error {
	_, err := s.db.Exec(`DELETE FROM feeds WHERE id = ?`, feedID)
	return err
}
