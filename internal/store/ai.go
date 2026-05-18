package store

import "database/sql"

func (s *Store) AIOutput(itemID int64, kind string) (AIOutput, error) {
	row := s.db.QueryRow(`
		SELECT id, item_id, kind, prompt, response, created_at
		FROM ai_outputs
		WHERE item_id = ? AND kind = ?
		ORDER BY id DESC
		LIMIT 1
	`, itemID, kind)
	return s.scanAIOutput(row)
}

func (s *Store) SaveAIOutput(itemID int64, kind, prompt, response string) error {
	var err error
	prompt, err = s.encryptText(prompt)
	if err != nil {
		return err
	}
	response, err = s.encryptText(response)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
		INSERT INTO ai_outputs(item_id, kind, prompt, response)
		VALUES(?, ?, ?, ?)
	`, itemID, kind, prompt, response)
	return err
}

func (s *Store) scanAIOutput(row scanner) (AIOutput, error) {
	var out AIOutput
	var created sql.NullString
	err := row.Scan(&out.ID, &out.ItemID, &out.Kind, &out.Prompt, &out.Response, &created)
	out.Prompt = s.decryptText(out.Prompt)
	out.Response = s.decryptText(out.Response)
	out.CreatedAt = parseTime(created)
	return out, err
}
