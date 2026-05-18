package store

import "database/sql"

func (s *Store) UpsertItem(item Item) (bool, error) {
	res, err := s.db.Exec(`
		INSERT INTO items(
			feed_id, guid, title, link, published_at, content_html, content_markdown,
			enclosure_url, enclosure_type, sludge_flag, sludge_checked, duration_seconds
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(feed_id, guid) DO NOTHING
	`, item.FeedID, item.GUID, item.Title, item.Link, nullTime(item.PublishedAt), item.ContentHTML,
		item.ContentMarkdown, item.EnclosureURL, item.EnclosureType, item.SludgeFlag, item.SludgeChecked, item.DurationSeconds)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil || n > 0 || item.DurationSeconds <= 0 {
		return n > 0, err
	}
	_, err = s.db.Exec(`
		UPDATE items
		SET duration_seconds = ?
		WHERE feed_id = ? AND guid = ? AND duration_seconds = 0
	`, item.DurationSeconds, item.FeedID, item.GUID)
	return n > 0, err
}

func (s *Store) Items(feedID int64, hideSludge bool) ([]Item, error) {
	query := `
		SELECT id, feed_id, guid, title, link, published_at,
			enclosure_url, enclosure_type, read_status, sludge_flag, sludge_checked, playhead_seconds, duration_seconds
		FROM items
		WHERE feed_id = ?`
	if hideSludge {
		query += ` AND sludge_flag = 0`
	}
	query += ` ORDER BY published_at DESC, id DESC`
	rows, err := s.db.Query(query, feedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Item
	for rows.Next() {
		item, err := scanItemSummary(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) Item(id int64) (Item, error) {
	row := s.db.QueryRow(`
		SELECT id, feed_id, guid, title, link, published_at, content_html, content_markdown,
			enclosure_url, enclosure_type, read_status, sludge_flag, sludge_checked, playhead_seconds, duration_seconds
		FROM items
		WHERE id = ?
	`, id)
	return scanItem(row)
}

func (s *Store) ItemCount(feedID int64) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM items WHERE feed_id = ?`, feedID).Scan(&count)
	return count, err
}

func (s *Store) MarkRead(id int64) error {
	_, err := s.db.Exec(`UPDATE items SET read_status = 1 WHERE id = ?`, id)
	return err
}

func (s *Store) SetSludge(id int64, flagged bool) error {
	_, err := s.db.Exec(`UPDATE items SET sludge_flag = ?, sludge_checked = 1 WHERE id = ?`, flagged, id)
	return err
}

func (s *Store) SetPlayhead(id int64, seconds int) error {
	if seconds < 0 {
		seconds = 0
	}
	_, err := s.db.Exec(`UPDATE items SET playhead_seconds = ? WHERE id = ?`, seconds, id)
	return err
}

func (s *Store) Rules() ([]BouncerRule, error) {
	rows, err := s.db.Query(`SELECT id, rule_prompt FROM bouncer_rules ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rules []BouncerRule
	for rows.Next() {
		var r BouncerRule
		if err := rows.Scan(&r.ID, &r.Prompt); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanItem(row scanner) (Item, error) {
	var item Item
	var published sql.NullString
	var read, sludge, checked int
	err := row.Scan(&item.ID, &item.FeedID, &item.GUID, &item.Title, &item.Link, &published,
		&item.ContentHTML, &item.ContentMarkdown, &item.EnclosureURL, &item.EnclosureType,
		&read, &sludge, &checked, &item.PlayheadSeconds, &item.DurationSeconds)
	item.PublishedAt = parseTime(published)
	item.ReadStatus = read == 1
	item.SludgeFlag = sludge == 1
	item.SludgeChecked = checked == 1
	return item, err
}

func scanItemSummary(row scanner) (Item, error) {
	var item Item
	var published sql.NullString
	var read, sludge, checked int
	err := row.Scan(&item.ID, &item.FeedID, &item.GUID, &item.Title, &item.Link, &published,
		&item.EnclosureURL, &item.EnclosureType, &read, &sludge, &checked, &item.PlayheadSeconds, &item.DurationSeconds)
	item.PublishedAt = parseTime(published)
	item.ReadStatus = read == 1
	item.SludgeFlag = sludge == 1
	item.SludgeChecked = checked == 1
	return item, err
}
