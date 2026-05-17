package config

func DefaultFeeds() []SeedFeed {
	return []SeedFeed{
		{Category: "TECH", Title: "Ars Technica", URL: "https://arstechnica.com/feed/atom/"},
		{Category: "TECH", Title: "The Verge", URL: "https://www.theverge.com/rss/index.xml"},
		{Category: "TECH", Title: "Hacker News", URL: "https://news.ycombinator.com/rss"},
		{Category: "WORLD", Title: "BBC World", URL: "https://feeds.bbci.co.uk/news/world/rss.xml"},
		{Category: "WORLD", Title: "NPR News", URL: "https://feeds.npr.org/1001/rss.xml"},
		{Category: "SPORTS", Title: "AP Sports", URL: "https://apnews.com/sports.rss"},
		{Category: "MUSIC", Title: "Pitchfork News", URL: "https://pitchfork.com/feed/feed-news/rss"},
		{Category: "MUSIC", Title: "Rolling Stone Music", URL: "https://www.rollingstone.com/music/feed/"},
		{Category: "GOPHER", Title: "Floodgap Gopher", URL: "gopher://gopher.floodgap.com/"},
		{Category: "GOPHER", Title: "SDF Gopher", URL: "gopher://sdf.org/"},
	}
}
