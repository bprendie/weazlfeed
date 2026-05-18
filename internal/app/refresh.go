package app

import (
	"context"
	"time"

	"github.com/bprendie/weazlfeed/internal/feed"
	"github.com/bprendie/weazlfeed/internal/llm"
	"github.com/bprendie/weazlfeed/internal/store"
)

type RefreshResult struct {
	Added   int
	Failed  int
	Checked int
	Details []RefreshFeedResult
}

type RefreshFeedResult struct {
	Title       string
	URL         string
	Added       int
	Status      int
	NotModified bool
	Err         string
}

func Refresh(ctx context.Context, vault *store.Store, feeds []store.Feed, ai llm.Client) (RefreshResult, error) {
	client := feed.NewClient()
	result := RefreshResult{}
	rules, err := vault.Rules()
	if err != nil {
		return result, err
	}
	rulePrompts := make([]string, 0, len(rules))
	for _, rule := range rules {
		rulePrompts = append(rulePrompts, rule.Prompt)
	}
	useBouncer := len(rulePrompts) > 0 && ai.Available(ctx)
	for _, src := range feeds {
		result.Checked++
		added, status, notModified, err := refreshOne(ctx, vault, client, src, ai, rulePrompts, useBouncer)
		detail := RefreshFeedResult{Title: src.Title, URL: src.URL, Added: added, Status: status, NotModified: notModified}
		if err != nil {
			_ = vault.SetFeedStatus(src.ID, 0, "", "", err.Error())
			detail.Err = err.Error()
			result.Failed++
			result.Details = append(result.Details, detail)
			continue
		}
		result.Added += added
		result.Details = append(result.Details, detail)
	}
	return result, nil
}

func refreshOne(ctx context.Context, vault *store.Store, client feed.Client, src store.Feed, ai llm.Client, rules []string, useBouncer bool) (int, int, bool, error) {
	feedCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	etag, modified := src.ETag, src.LastModified
	if count, err := vault.ItemCount(src.ID); err == nil && count == 0 {
		etag, modified = "", ""
	}
	parsed, err := client.Fetch(feedCtx, src.URL, etag, modified)
	if err != nil {
		return 0, 0, false, err
	}
	_ = vault.SetFeedStatus(src.ID, parsed.Status, parsed.ETag, parsed.LastModified, "")
	if parsed.NotModified {
		return 0, parsed.Status, true, vault.MarkFetched(src.ID, time.Now())
	}
	title := firstText(parsed.Title, src.Title, src.URL)
	feedID, err := vault.UpsertFeed(title, src.URL, parsed.Type, src.Section, src.Folder, src.Category)
	if err != nil {
		return 0, parsed.Status, false, err
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
		}
		if useBouncer {
			flagged, err := ai.FlagSludge(ctx, item.ContentMarkdown, rules)
			if err == nil {
				item.SludgeFlag = flagged
				item.SludgeChecked = true
			}
		}
		ok, err := vault.UpsertItem(item)
		if err != nil {
			return added, parsed.Status, false, err
		}
		if ok {
			added++
		}
	}
	return added, parsed.Status, false, vault.MarkFetched(feedID, time.Now())
}

func firstText(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
