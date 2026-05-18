package tui

import (
	"context"
	"time"

	"github.com/bprendie/weazlfeed/internal/audio"
	"github.com/bprendie/weazlfeed/internal/config"
	"github.com/bprendie/weazlfeed/internal/feed"
	"github.com/bprendie/weazlfeed/internal/llm"
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
		return itemsMsg{items: items, err: err}
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
	client := feed.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	added := 0
	failed := 0
	rules, err := vault.Rules()
	if err != nil {
		return fetchMsg{err: err}
	}
	rulePrompts := make([]string, 0, len(rules))
	for _, rule := range rules {
		rulePrompts = append(rulePrompts, rule.Prompt)
	}
	useBouncer := len(rulePrompts) > 0 && ai.Available(ctx)
	for _, src := range feeds {
		parsed, err := client.Fetch(ctx, src.URL, src.ETag, src.LastModified)
		if err != nil {
			_ = vault.SetFeedStatus(src.ID, 0, "", "", err.Error())
			failed++
			continue
		}
		_ = vault.SetFeedStatus(src.ID, parsed.Status, parsed.ETag, parsed.LastModified, "")
		if parsed.NotModified {
			_ = vault.MarkFetched(src.ID, time.Now())
			continue
		}
		title := firstText(parsed.Title, src.Title, src.URL)
		feedID, err := vault.UpsertFeed(title, src.URL, parsed.Type, src.Section, src.Folder, src.Category)
		if err != nil {
			return fetchMsg{added: added, failed: failed, err: err}
		}
		for _, parsedItem := range parsed.Items {
			item := store.Item{
				FeedID:          feedID,
				GUID:            parsedItem.GUID,
				Title:           firstText(parsedItem.Title, "untitled"),
				Link:            parsedItem.Link,
				PublishedAt:     parsedItem.PublishedAt,
				ContentHTML:     parsedItem.ContentHTML,
				ContentMarkdown: parsedItem.ContentMarkdown,
				EnclosureURL:    parsedItem.EnclosureURL,
				EnclosureType:   parsedItem.EnclosureType,
			}
			if useBouncer {
				flagged, err := ai.FlagSludge(ctx, item.ContentMarkdown, rulePrompts)
				if err == nil {
					item.SludgeFlag = flagged
					item.SludgeChecked = true
				}
			}
			ok, err := vault.UpsertItem(item)
			if err != nil {
				return fetchMsg{added: added, failed: failed, err: err}
			}
			if ok {
				added++
			}
		}
		_ = vault.MarkFetched(feedID, time.Now())
	}
	return fetchMsg{added: added, failed: failed}
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
