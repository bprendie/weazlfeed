package store

const schema = `
CREATE TABLE IF NOT EXISTS feeds (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	url TEXT NOT NULL UNIQUE,
	type TEXT NOT NULL DEFAULT 'rss',
	last_fetched TEXT
);

CREATE TABLE IF NOT EXISTS items (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	feed_id INTEGER NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
	guid TEXT NOT NULL,
	title TEXT NOT NULL,
	link TEXT,
	published_at TEXT,
	content_html TEXT,
	content_markdown TEXT,
	enclosure_url TEXT,
	enclosure_type TEXT,
	read_status INTEGER NOT NULL DEFAULT 0,
	sludge_flag INTEGER NOT NULL DEFAULT 0,
	sludge_checked INTEGER NOT NULL DEFAULT 0,
	playhead_seconds INTEGER NOT NULL DEFAULT 0,
	UNIQUE(feed_id, guid)
);

CREATE TABLE IF NOT EXISTS bouncer_rules (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	rule_prompt TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_items_feed_date ON items(feed_id, published_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_items_unread ON items(read_status);

ALTER TABLE items ADD COLUMN playhead_seconds INTEGER NOT NULL DEFAULT 0;
`
