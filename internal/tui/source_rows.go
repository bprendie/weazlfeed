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

func (m Model) selectedSourceRow() (sourceRow, bool) {
	rows := m.sourceRows()
	if len(rows) == 0 {
		return sourceRow{}, false
	}
	index := clampInt(m.sourceCursor, 0, len(rows)-1)
	return rows[index], true
}

func (m Model) selectableSourceRows() []int {
	rows := m.sourceRows()
	indices := make([]int, 0, len(rows))
	for i, row := range rows {
		if row.kind != sourceSection {
			indices = append(indices, i)
		}
	}
	return indices
}

func (m Model) sourceCursorRow() int {
	return m.sourceCursor
}

func (m *Model) syncFeedCursorFromSource() {
	row, ok := m.selectedSourceRow()
	if ok && row.kind == sourceFeed {
		m.feedCursor = row.feedIndex
	}
}

func (m *Model) selectSourceRow(section, folder string) {
	for i, row := range m.sourceRows() {
		if row.kind == sourceFolder && row.section == section && row.folder == folder {
			m.sourceCursor = i
			m.syncFeedCursorFromSource()
			return
		}
	}
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
