package store

import "database/sql"

func (s *Store) AIOutput(itemID int64, kind string) (AIOutput, error) {
	row := s.db.QueryRow(`
		SELECT id, item_id, kind, coalesce(item_title, ''), coalesce(item_content, ''), prompt, response, created_at
		FROM ai_outputs
		WHERE item_id = ? AND kind = ?
		ORDER BY id DESC
		LIMIT 1
	`, itemID, kind)
	return s.scanAIOutput(row)
}

func (s *Store) AIOutputs(kind string) ([]AIOutput, error) {
	rows, err := s.db.Query(`
		SELECT id, item_id, kind, coalesce(item_title, ''), coalesce(item_content, ''), prompt, response, created_at
		FROM ai_outputs
		WHERE kind = ?
		ORDER BY id DESC
	`, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var outputs []AIOutput
	for rows.Next() {
		out, err := s.scanAIOutput(rows)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, out)
	}
	return outputs, rows.Err()
}

func (s *Store) SaveAIOutput(item Item, kind, prompt, response string) error {
	var err error
	title, err := s.encryptText(item.Title)
	if err != nil {
		return err
	}
	content, err := s.encryptText(item.ContentMarkdown)
	if err != nil {
		return err
	}
	prompt, err = s.encryptText(prompt)
	if err != nil {
		return err
	}
	response, err = s.encryptText(response)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
		INSERT INTO ai_outputs(item_id, kind, item_title, item_content, prompt, response)
		VALUES(?, ?, ?, ?, ?, ?)
	`, item.ID, kind, title, content, prompt, response)
	return err
}

func (s *Store) DeleteAIOutput(id int64) error {
	_, err := s.db.Exec(`DELETE FROM ai_outputs WHERE id = ?`, id)
	return err
}

func (s *Store) scanAIOutput(row scanner) (AIOutput, error) {
	var out AIOutput
	var created sql.NullString
	err := row.Scan(&out.ID, &out.ItemID, &out.Kind, &out.ItemTitle, &out.ItemContent, &out.Prompt, &out.Response, &created)
	out.ItemTitle = s.decryptText(out.ItemTitle)
	out.ItemContent = s.decryptText(out.ItemContent)
	out.Prompt = s.decryptText(out.Prompt)
	out.Response = s.decryptText(out.Response)
	out.CreatedAt = parseTime(created)
	return out, err
}
