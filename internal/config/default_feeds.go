package config

func DefaultFeeds() []SeedFeed {
	return []SeedFeed{
		{Section: "News", Folder: "Tech", Title: "Ars Technica", URL: "https://arstechnica.com/feed/atom/"},
		{Section: "News", Folder: "Tech", Title: "The Verge", URL: "https://www.theverge.com/rss/index.xml"},
		{Section: "News", Folder: "Tech", Title: "Hacker News", URL: "https://news.ycombinator.com/rss"},
		{Section: "News", Folder: "World", Title: "BBC World", URL: "https://feeds.bbci.co.uk/news/world/rss.xml"},
		{Section: "News", Folder: "World", Title: "NPR News", URL: "https://feeds.npr.org/1001/rss.xml"},
		{Section: "News", Folder: "Sports", Title: "AP Sports", URL: "https://apnews.com/sports.rss"},
		{Section: "News", Folder: "Music", Title: "Pitchfork News", URL: "https://pitchfork.com/feed/feed-news/rss"},
		{Section: "News", Folder: "Music", Title: "Rolling Stone Music", URL: "https://www.rollingstone.com/music/feed/"},
		{Section: "Podcasts", Folder: "Technology", Title: "CoRecursive", URL: "https://corecursive.com/feed/"},
		{Section: "Gopher", Folder: "Directory", Title: "Floodgap Gopher", URL: "gopher://gopher.floodgap.com/"},
		{Section: "Gopher", Folder: "Directory", Title: "SDF Gopher", URL: "gopher://sdf.org/"},
	}
}
