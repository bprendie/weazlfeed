package store

import (
	"database/sql"
	"time"
)

func (s *Store) UpsertFeed(title, url, feedType, category string) (int64, error) {
	if category == "" {
		category = "GENERAL"
	}
	res, err := s.db.Exec(`
		INSERT INTO feeds(title, url, type, category) VALUES(?, ?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET title=excluded.title, type=excluded.type, category=excluded.category
	`, title, url, feedType, category)
	if err != nil {
		return 0, err
	}
	if id, err := res.LastInsertId(); err == nil && id > 0 {
		return id, nil
	}
	var id int64
	err = s.db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, url).Scan(&id)
	return id, err
}

func (s *Store) Feeds() ([]Feed, error) {
	rows, err := s.db.Query(`
		SELECT f.id, f.title, f.url, f.type, f.category, f.last_fetched,
			COUNT(CASE WHEN i.read_status = 0 THEN 1 END) AS unread
		FROM feeds f
		LEFT JOIN items i ON i.feed_id = f.id
		GROUP BY f.id
		ORDER BY lower(f.category), lower(f.title)
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var feeds []Feed
	for rows.Next() {
		var f Feed
		var fetched sql.NullString
		if err := rows.Scan(&f.ID, &f.Title, &f.URL, &f.Type, &f.Category, &fetched, &f.Unread); err != nil {
			return nil, err
		}
		f.LastFetched = parseTime(fetched)
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
}

func (s *Store) MarkFetched(feedID int64, when time.Time) error {
	_, err := s.db.Exec(`UPDATE feeds SET last_fetched = ? WHERE id = ?`, when.UTC().Format(time.RFC3339), feedID)
	return err
}
