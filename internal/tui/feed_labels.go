package tui

import (
	"strings"

	"github.com/bprendie/weazlfeed/internal/store"
)

func folderFromFeed(feed store.Feed) string {
	if feed.Folder != "" {
		return feed.Folder
	}
	return titleCase(firstText(feed.Category, "General"))
}

func titleCase(value string) string {
	value = strings.ToLower(value)
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
