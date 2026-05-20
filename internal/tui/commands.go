package tui

import (
	"context"
	"strings"
	"time"

	"github.com/bprendie/weazlfeed/internal/app"
	"github.com/bprendie/weazlfeed/internal/config"
	"github.com/bprendie/weazlfeed/internal/feed"
	"github.com/bprendie/weazlfeed/internal/llm"
	"github.com/bprendie/weazlfeed/internal/podcast"
	"github.com/bprendie/weazlfeed/internal/store"
	tea "github.com/charmbracelet/bubbletea"
)

func seedFeedsCmd(vault *store.Store, seeds []config.SeedFeed) tea.Cmd {
	return func() tea.Msg {
		for _, seed := range seeds {
			if seed.URL == "" {
				continue
			}
			title := seed.Title
			if title == "" {
				title = seed.URL
			}
			feedType := "rss"
			if feed.IsGopher(seed.URL) {
				feedType = "gopher"
			}
			if _, err := vault.UpsertFeed(title, seed.URL, feedType, seed.Section, seed.Folder, seed.Category); err != nil {
				return feedsMsg{err: err}
			}
		}
		feeds, err := vault.Feeds()
		folders, folderErr := vault.Folders()
		interrogations, aiErr := vault.AIOutputs("ask")
		rules, rulesErr := vault.Rules()
		return feedsMsg{feeds: feeds, folders: folders, interrogations: interrogations, rules: rules, err: firstErr(err, folderErr, aiErr, rulesErr)}
	}
}

func loadFeedsCmd(vault *store.Store) tea.Cmd {
	return func() tea.Msg {
		feeds, err := vault.Feeds()
		folders, folderErr := vault.Folders()
		interrogations, aiErr := vault.AIOutputs("ask")
		rules, rulesErr := vault.Rules()
		return feedsMsg{feeds: feeds, folders: folders, interrogations: interrogations, rules: rules, err: firstErr(err, folderErr, aiErr, rulesErr)}
	}
}

func loadItemsCmd(vault *store.Store, feedID int64, hide bool) tea.Cmd {
	return func() tea.Msg {
		items, err := vault.Items(feedID, hide)
		return itemsMsg{feedID: feedID, items: items, err: err}
	}
}

func prefetchItemsCmd(vault *store.Store, feeds []store.Feed, hide bool) tea.Cmd {
	return func() tea.Msg {
		cache := make(map[int64][]store.Item, len(feeds))
		for _, feed := range feeds {
			items, err := vault.Items(feed.ID, hide)
			if err != nil {
				return allItemsMsg{err: err}
			}
			cache[feed.ID] = items
		}
		return allItemsMsg{itemsByFeed: cache}
	}
}

func refreshAllCmd(vault *store.Store, ai llm.Client) tea.Cmd {
	return func() tea.Msg {
		feeds, err := vault.Feeds()
		if err != nil {
			return fetchMsg{err: err}
		}
		return refresh(vault, feeds, ai)
	}
}

func refreshCmd(vault *store.Store, feeds []store.Feed, ai llm.Client) tea.Cmd {
	return func() tea.Msg {
		return refresh(vault, feeds, ai)
	}
}

func refresh(vault *store.Store, feeds []store.Feed, ai llm.Client) tea.Msg {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	result, err := app.Refresh(ctx, vault, feeds, ai)
	return fetchMsg{checked: result.Checked, added: result.Added, failed: result.Failed, err: err}
}

func addFeedCmd(vault *store.Store, rawURL, section, folder string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		rawURL = strings.TrimSpace(rawURL)
		if rawURL == "" {
			return addFeedMsg{}
		}
		feedType := "rss"
		if feed.IsGopher(rawURL) {
			feedType = "gopher"
			section = "Gopher"
			if folder == "" || folder == "Search" || folder == "General" {
				folder = "Directory"
			}
		}
		parsed, err := feed.NewClient().Fetch(ctx, rawURL, "", "")
		if err != nil {
			return addFeedMsg{err: err}
		}
		if parsed.Type != "" {
			feedType = parsed.Type
		}
		if feedHasAudio(parsed) {
			section = "Podcasts"
			if folder == "" || folder == "Directory" || folder == "General" {
				folder = "Search"
			}
		}
		if section == "" {
			section = "News"
		}
		if folder == "" {
			folder = "General"
		}
		title := firstText(parsed.Title, rawURL)
		feedID, err := vault.UpsertFeed(title, rawURL, feedType, section, folder, folder)
		if err != nil {
			return addFeedMsg{err: err}
		}
		added := 0
		for _, parsedItem := range parsed.Items {
			item := store.Item{
				FeedID:          feedID,
				GUID:            firstText(parsedItem.GUID, parsedItem.Link, parsedItem.Title),
				Title:           firstText(parsedItem.Title, "untitled"),
				Link:            parsedItem.Link,
				PublishedAt:     parsedItem.PublishedAt,
				ContentHTML:     parsedItem.ContentHTML,
				ContentMarkdown: parsedItem.ContentMarkdown,
				EnclosureURL:    parsedItem.EnclosureURL,
				EnclosureType:   parsedItem.EnclosureType,
				DurationSeconds: parsedItem.DurationSeconds,
				SludgeChecked:   true,
			}
			ok, err := vault.UpsertItem(item)
			if err != nil {
				return addFeedMsg{err: err}
			}
			if ok {
				added++
			}
		}
		_ = vault.MarkFetched(feedID, time.Now())
		_ = vault.SetFeedStatus(feedID, parsed.Status, parsed.ETag, parsed.LastModified, "")
		return addFeedMsg{feedID: feedID, title: title, section: section, folder: folder, added: added}
	}
}

func deleteFeedCmd(vault *store.Store, feedID int64, title string) tea.Cmd {
	return func() tea.Msg {
		return deleteFeedMsg{id: feedID, kind: "feed", title: title, err: vault.DeleteFeed(feedID)}
	}
}

func deleteInterrogationCmd(vault *store.Store, id int64, title string) tea.Cmd {
	return func() tea.Msg {
		return deleteFeedMsg{id: id, kind: "interrogation", title: title, err: vault.DeleteAIOutput(id)}
	}
}

func feedHasAudio(parsed feed.Feed) bool {
	for _, item := range parsed.Items {
		if strings.HasPrefix(item.EnclosureType, "audio/") || looksAudioURL(item.EnclosureURL) {
			return true
		}
	}
	return false
}

func looksAudioURL(raw string) bool {
	raw = strings.ToLower(strings.TrimSpace(raw))
	return strings.HasSuffix(raw, ".mp3") || strings.HasSuffix(raw, ".m4a") ||
		strings.HasSuffix(raw, ".aac") || strings.HasSuffix(raw, ".ogg") ||
		strings.HasSuffix(raw, ".opus") || strings.HasSuffix(raw, ".wav")
}

const aiMaxChars = 12000

func aiCmd(vault *store.Store, ai llm.Client, mode string, item store.Item, question string) tea.Cmd {
	return func() tea.Msg {
		if item.ContentMarkdown == "" && item.ID != 0 {
			full, err := vault.Item(item.ID)
			if err != nil {
				return aiMsg{itemID: item.ID, kind: mode, question: question, err: err}
			}
			item = full
		}
		if item.ID == 0 && item.GUID == "" {
			item.GUID = "interrogation:" + item.Title
		}
		inTokens := estimateTokens(item.ContentMarkdown) + estimateTokens(question)
		if mode == "triage" && item.ID != 0 {
			if out, err := vault.AIOutput(item.ID, "triage"); err == nil && out.Response != "" {
				return aiMsg{itemID: item.ID, kind: mode, text: out.Response, cached: true, inTokens: inTokens, outTokens: estimateTokens(out.Response)}
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		body := trimAIInput(item.ContentMarkdown)
		var text string
		var err error
		if mode == "ask" {
			text, err = ai.Ask(ctx, body, question)
		} else {
			text, err = ai.Triage(ctx, body)
		}
		if err == nil && item.ID != 0 && text != "" {
			_ = vault.SaveAIOutput(item, mode, question, text)
		}
		return aiMsg{itemID: item.ID, kind: mode, question: question, text: text, inTokens: inTokens, outTokens: estimateTokens(text), err: err}
	}
}

func trimAIInput(markdown string) string {
	markdown = strings.TrimSpace(markdown)
	if len(markdown) <= aiMaxChars {
		return markdown
	}
	return markdown[:aiMaxChars] + "\n\n[TRUNCATED FOR LOCAL MODEL SPEED]"
}

func estimateTokens(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	return max(1, len([]rune(text))/4)
}

func podcastSearchCmd(query string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		results, err := podcast.NewClient().Search(ctx, query, 20)
		return podcastSearchMsg{results: results, err: err}
	}
}

func gopherArticleCmd(url string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		parsed, err := feed.FetchGopher(ctx, url)
		if err != nil {
			return articleMsg{err: err}
		}
		if len(parsed.Items) == 1 && parsed.Items[0].Link == url {
			return articleMsg{text: parsed.Items[0].ContentMarkdown}
		}
		text := "# " + parsed.Title + "\n\n"
		for _, item := range parsed.Items {
			text += "- " + item.Title + "\n  " + item.Link + "\n"
		}
		return articleMsg{text: text}
	}
}

func gopherPageCmd(url string, width int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		parsed, err := feed.FetchGopher(ctx, url)
		if err != nil {
			return gopherMsg{url: url, err: err}
		}
		if len(parsed.Items) == 1 && parsed.Items[0].Link == url {
			text := parsed.Items[0].ContentMarkdown
			return gopherMsg{url: url, text: text, rendered: renderMarkdownText(text, width)}
		}
		items := make([]store.Item, 0, len(parsed.Items))
		for _, item := range parsed.Items {
			body := item.ContentMarkdown
			if body == "" {
				body = item.Link
			}
			items = append(items, store.Item{
				GUID:            firstText(item.GUID, item.Link, item.Title),
				Title:           firstText(item.Title, "untitled"),
				Link:            item.Link,
				PublishedAt:     item.PublishedAt,
				ContentHTML:     item.ContentHTML,
				ContentMarkdown: body,
				EnclosureType:   firstText(item.EnclosureType, gopherEnclosureType(item.Link)),
				ReadStatus:      true,
				SludgeChecked:   true,
			})
		}
		return gopherMsg{url: url, items: items}
	}
}

func gopherEnclosureType(link string) string {
	kind := gopherLinkKind(link)
	switch kind {
	case "1":
		return "gopher/directory"
	case "0":
		return "text/plain"
	case "7":
		return "gopher/search"
	case "g", "I":
		return "image/gopher"
	case "h":
		return "text/html"
	case "8":
		return "gopher/telnet"
	case "9":
		return "application/octet-stream"
	default:
		return "application/gopher"
	}
}

func gopherLinkKind(link string) string {
	parts := strings.SplitN(link, "/", 4)
	if len(parts) < 4 || parts[3] == "" {
		return "1"
	}
	return parts[3][:1]
}

func renderReaderCmd(vault *store.Store, item store.Item, width int) tea.Cmd {
	return func() tea.Msg {
		if item.ContentMarkdown == "" && item.ContentHTML == "" {
			full, err := vault.Item(item.ID)
			if err != nil {
				return readerMsg{err: err}
			}
			item = full
		}
		text := item.ContentMarkdown
		if text == "" {
			text = item.Link
		}
		return readerMsg{item: item, raw: text, rendered: renderMarkdownText(text, width)}
	}
}

func renderPrettyReaderCmd(item store.Item, raw string, width int) tea.Cmd {
	return func() tea.Msg {
		text := strings.TrimSpace(raw)
		if text == "" {
			if item.ContentMarkdown != "" {
				text = item.ContentMarkdown
			} else {
				text = item.Link
			}
		}
		return readerMsg{item: item, raw: text, rendered: renderPrettyMarkdownText(text, width), forced: true}
	}
}

func renderInterrogationCmd(out store.AIOutput, width int) tea.Cmd {
	return func() tea.Msg {
		text := interrogationBody(out)
		return interrogationMsg{raw: text, rendered: renderMarkdownText(text, width)}
	}
}

func playheadTickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
		return playheadTickMsg{}
	})
}

func audioTickCmd() tea.Cmd {
	return tea.Tick(time.Second/30, func(time.Time) tea.Msg {
		return audioTickMsg{}
	})
}

func aiTickCmd() tea.Cmd {
	return tea.Tick(time.Second/10, func(time.Time) tea.Msg {
		return aiTickMsg{}
	})
}

func firstText(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
