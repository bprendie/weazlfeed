package store

import "time"

type PruneOptions struct {
	Before       time.Time
	KeepUnread   bool
	KeepPlayhead bool
	Vacuum       bool
}

type PruneResult struct {
	Deleted int64
}

func (s *Store) Prune(opts PruneOptions) (PruneResult, error) {
	query := `DELETE FROM items WHERE published_at IS NOT NULL AND published_at != '' AND published_at < ?`
	args := []any{opts.Before.UTC().Format(time.RFC3339)}
	if opts.KeepUnread {
		query += ` AND read_status = 1`
	}
	if opts.KeepPlayhead {
		query += ` AND playhead_seconds = 0`
	}
	res, err := s.db.Exec(query, args...)
	if err != nil {
		return PruneResult{}, err
	}
	deleted, err := res.RowsAffected()
	if err != nil {
		return PruneResult{}, err
	}
	if opts.Vacuum {
		if _, err := s.db.Exec(`VACUUM`); err != nil {
			return PruneResult{}, err
		}
	}
	return PruneResult{Deleted: deleted}, nil
}
