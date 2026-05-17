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
			if _, err := vault.UpsertFeed(title, seed.URL, feedType, seed.Category); err != nil {
				return feedsMsg{err: err}
			}
		}
		feeds, err := vault.Feeds()
		return feedsMsg{feeds: feeds, err: err}
	}
}

func loadFeedsCmd(vault *store.Store) tea.Cmd {
	return func() tea.Msg {
		feeds, err := vault.Feeds()
		return feedsMsg{feeds: feeds, err: err}
	}
}

func loadItemsCmd(vault *store.Store, feedID int64, hide bool) tea.Cmd {
	return func() tea.Msg {
		items, err := vault.Items(feedID, hide)
		return itemsMsg{items: items, err: err}
	}
}

func refreshCmd(vault *store.Store, feeds []store.Feed, ai llm.Client) tea.Cmd {
	return func() tea.Msg {
		client := feed.NewClient()
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		added := 0
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
			parsed, err := client.Fetch(ctx, src.URL)
			if err != nil {
				return fetchMsg{added: added, err: err}
			}
			title := firstText(parsed.Title, src.Title, src.URL)
			feedID, err := vault.UpsertFeed(title, src.URL, parsed.Type, src.Category)
			if err != nil {
				return fetchMsg{added: added, err: err}
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
					return fetchMsg{added: added, err: err}
				}
				if ok {
					added++
				}
			}
			_ = vault.MarkFetched(feedID, time.Now())
		}
		return fetchMsg{added: added}
	}
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
