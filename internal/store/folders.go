package store

func (s *Store) UpsertFolder(section, name string) error {
	if section == "" {
		section = "News"
	}
	if name == "" {
		return nil
	}
	_, err := s.db.Exec(`
		INSERT INTO folders(section, name) VALUES(?, ?)
		ON CONFLICT(section, name) DO NOTHING
	`, section, name)
	return err
}

func (s *Store) Folders() ([]Folder, error) {
	rows, err := s.db.Query(`
		SELECT id, section, name, collapsed FROM folders
		ORDER BY CASE section WHEN 'News' THEN 0 WHEN 'Podcasts' THEN 1 WHEN 'Gopher' THEN 2 ELSE 3 END,
			lower(name)
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var folders []Folder
	for rows.Next() {
		var f Folder
		var collapsed int
		if err := rows.Scan(&f.ID, &f.Section, &f.Name, &collapsed); err != nil {
			return nil, err
		}
		f.Collapsed = collapsed == 1
		folders = append(folders, f)
	}
	return folders, rows.Err()
}

func (s *Store) SetFolderCollapsed(section, name string, collapsed bool) error {
	if err := s.UpsertFolder(section, name); err != nil {
		return err
	}
	value := 0
	if collapsed {
		value = 1
	}
	_, err := s.db.Exec(`UPDATE folders SET collapsed = ? WHERE section = ? AND name = ?`, value, section, name)
	return err
}
