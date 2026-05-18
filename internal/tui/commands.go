package tui

import (
	"context"
	"strings"
	"time"

	"github.com/bprendie/weazlfeed/internal/app"
	"github.com/bprendie/weazlfeed/internal/audio"
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
		return feedsMsg{feeds: feeds, folders: folders, err: firstErr(err, folderErr)}
	}
}

func loadFeedsCmd(vault *store.Store) tea.Cmd {
	return func() tea.Msg {
		feeds, err := vault.Feeds()
		folders, folderErr := vault.Folders()
		return feedsMsg{feeds: feeds, folders: folders, err: firstErr(err, folderErr)}
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

func aiCmd(ai llm.Client, mode string, item store.Item, question string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		var text string
		var err error
		if mode == "ask" {
			text, err = ai.Ask(ctx, item.ContentMarkdown, question)
		} else {
			text, err = ai.Triage(ctx, item.ContentMarkdown)
		}
		return aiMsg{text: text, err: err}
	}
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

func gopherPageCmd(url string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		parsed, err := feed.FetchGopher(ctx, url)
		if err != nil {
			return gopherMsg{url: url, err: err}
		}
		if len(parsed.Items) == 1 && parsed.Items[0].Link == url {
			return gopherMsg{url: url, text: parsed.Items[0].ContentMarkdown}
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
				EnclosureType:   gopherEnclosureType(item.Link),
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

func meterCmd(ch <-chan audio.Sample) tea.Cmd {
	return func() tea.Msg {
		sample, ok := <-ch
		if !ok {
			return meterMsg{}
		}
		return meterMsg(sample)
	}
}

func playheadTickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
		return playheadTickMsg{}
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
