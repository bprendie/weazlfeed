package tui

import "github.com/bprendie/weazlfeed/internal/store"

type sourceRowKind int

const (
	sourceSection sourceRowKind = iota
	sourceFolder
	sourceFeed
)

type sourceRow struct {
	kind      sourceRowKind
	section   string
	folder    string
	feedIndex int
	title     string
	collapsed bool
	unread    int
}

func (m Model) sourceRows() []sourceRow {
	rows := []sourceRow{}
	seenSection := map[string]bool{}
	seenFolder := map[string]bool{}
	for i, feed := range m.feeds {
		section := firstText(feed.Section, sectionFromFeed(feed))
		folder := firstText(feed.Folder, folderFromFeed(feed))
		if !seenSection[section] {
			rows = append(rows, sourceRow{kind: sourceSection, section: section, title: section})
			seenSection[section] = true
		}
		key := folderKey(section, folder)
		if !seenFolder[key] {
			collapsed := m.folderCollapsed(section, folder)
			rows = append(rows, sourceRow{kind: sourceFolder, section: section, folder: folder, title: folder, collapsed: collapsed})
			seenFolder[key] = true
		}
		if !m.folderCollapsed(section, folder) {
			rows = append(rows, sourceRow{kind: sourceFeed, section: section, folder: folder, feedIndex: i, title: feed.Title, unread: feed.Unread})
		}
	}
	return rows
}

func (m Model) visibleFeedIndices() []int {
	var indices []int
	for i, feed := range m.feeds {
		section := firstText(feed.Section, sectionFromFeed(feed))
		folder := firstText(feed.Folder, folderFromFeed(feed))
		if !m.folderCollapsed(section, folder) {
			indices = append(indices, i)
		}
	}
	return indices
}

func (m Model) sourceCursorRow() int {
	rows := m.sourceRows()
	for i, row := range rows {
		if row.kind == sourceFeed && row.feedIndex == m.feedCursor {
			return i
		}
	}
	return 0
}

func (m Model) folderCollapsed(section, folder string) bool {
	for _, item := range m.folders {
		if item.Section == section && item.Name == folder {
			return item.Collapsed
		}
	}
	return false
}

func (m *Model) setFolderCollapsed(section, folder string, collapsed bool) error {
	if err := m.store.SetFolderCollapsed(section, folder, collapsed); err != nil {
		return err
	}
	for i := range m.folders {
		if m.folders[i].Section == section && m.folders[i].Name == folder {
			m.folders[i].Collapsed = collapsed
			return nil
		}
	}
	m.folders = append(m.folders, store.Folder{Section: section, Name: folder, Collapsed: collapsed})
	return nil
}

func folderKey(section, folder string) string {
	return section + "\x00" + folder
}
