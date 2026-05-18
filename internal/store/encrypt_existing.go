package store

func (s *Store) EnsureEncrypted() error {
	if !s.unlocked {
		return nil
	}
	if err := s.encryptFeeds(); err != nil {
		return err
	}
	if err := s.encryptItems(); err != nil {
		return err
	}
	if err := s.encryptAIOutputs(); err != nil {
		return err
	}
	return s.encryptRules()
}

func (s *Store) encryptFeeds() error {
	rows, err := s.db.Query(`SELECT id, title, url, section, folder, category, coalesce(last_error, '') FROM feeds`)
	if err != nil {
		return err
	}
	defer rows.Close()
	type feedRow struct {
		id                                          int64
		title, url, section, folder, cat, lastError string
	}
	var items []feedRow
	for rows.Next() {
		var item feedRow
		if err := rows.Scan(&item.id, &item.title, &item.url, &item.section, &item.folder, &item.cat, &item.lastError); err != nil {
			return err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, item := range items {
		title, err := s.encryptText(s.decryptText(item.title))
		if err != nil {
			return err
		}
		url, err := s.encryptText(s.decryptText(item.url))
		if err != nil {
			return err
		}
		section, err := s.encryptText(s.decryptText(item.section))
		if err != nil {
			return err
		}
		folder, err := s.encryptText(s.decryptText(item.folder))
		if err != nil {
			return err
		}
		category, err := s.encryptText(s.decryptText(item.cat))
		if err != nil {
			return err
		}
		lastError, err := s.encryptText(s.decryptText(item.lastError))
		if err != nil {
			return err
		}
		if _, err := s.db.Exec(`
			UPDATE feeds SET title = ?, url = ?, section = ?, folder = ?, category = ?, last_error = ?, feed_key = ?
			WHERE id = ?
		`, title, url, section, folder, category, lastError, s.feedKey(item.url), item.id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) encryptItems() error {
	rows, err := s.db.Query(`
		SELECT id, title, link, content_html, content_markdown, enclosure_url FROM items
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	type itemRow struct {
		id                                     int64
		title, link, html, markdown, enclosure string
	}
	var items []itemRow
	for rows.Next() {
		var item itemRow
		if err := rows.Scan(&item.id, &item.title, &item.link, &item.html, &item.markdown, &item.enclosure); err != nil {
			return err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, item := range items {
		title, err := s.encryptText(s.decryptText(item.title))
		if err != nil {
			return err
		}
		link, err := s.encryptText(s.decryptText(item.link))
		if err != nil {
			return err
		}
		html, err := s.encryptText(s.decryptText(item.html))
		if err != nil {
			return err
		}
		markdown, err := s.encryptText(s.decryptText(item.markdown))
		if err != nil {
			return err
		}
		enclosure, err := s.encryptText(s.decryptText(item.enclosure))
		if err != nil {
			return err
		}
		if _, err := s.db.Exec(`
			UPDATE items
			SET title = ?, link = ?, content_html = ?, content_markdown = ?, enclosure_url = ?
			WHERE id = ?
		`, title, link, html, markdown, enclosure, item.id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) encryptRules() error {
	rows, err := s.db.Query(`SELECT id, rule_prompt FROM bouncer_rules`)
	if err != nil {
		return err
	}
	defer rows.Close()
	type ruleRow struct {
		id     int64
		prompt string
	}
	var rules []ruleRow
	for rows.Next() {
		var rule ruleRow
		if err := rows.Scan(&rule.id, &rule.prompt); err != nil {
			return err
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, rule := range rules {
		prompt, err := s.encryptText(s.decryptText(rule.prompt))
		if err != nil {
			return err
		}
		if _, err := s.db.Exec(`UPDATE bouncer_rules SET rule_prompt = ? WHERE id = ?`, prompt, rule.id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) encryptAIOutputs() error {
	rows, err := s.db.Query(`SELECT id, coalesce(item_title, ''), coalesce(item_content, ''), prompt, response FROM ai_outputs`)
	if err != nil {
		return err
	}
	defer rows.Close()
	type aiRow struct {
		id                               int64
		title, content, prompt, response string
	}
	var outputs []aiRow
	for rows.Next() {
		var out aiRow
		if err := rows.Scan(&out.id, &out.title, &out.content, &out.prompt, &out.response); err != nil {
			return err
		}
		outputs = append(outputs, out)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, out := range outputs {
		title, err := s.encryptText(s.decryptText(out.title))
		if err != nil {
			return err
		}
		content, err := s.encryptText(s.decryptText(out.content))
		if err != nil {
			return err
		}
		prompt, err := s.encryptText(s.decryptText(out.prompt))
		if err != nil {
			return err
		}
		response, err := s.encryptText(s.decryptText(out.response))
		if err != nil {
			return err
		}
		if _, err := s.db.Exec(`UPDATE ai_outputs SET item_title = ?, item_content = ?, prompt = ?, response = ? WHERE id = ?`, title, content, prompt, response, out.id); err != nil {
			return err
		}
	}
	return nil
}
