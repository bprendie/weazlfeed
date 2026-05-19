package tui

import "github.com/bprendie/weazlfeed/internal/store"

type sourceRowKind int

const (
	sourceSection sourceRowKind = iota
	sourceFolder
	sourceFeed
	sourceInterrogation
)

type sourceRow struct {
	kind      sourceRowKind
	section   string
	folder    string
	feedIndex int
	aiIndex   int
	title     string
	collapsed bool
	unread    int
}

func (m Model) sourceRows() []sourceRow {
	rows := []sourceRow{}
	sections := []string{"News", "Podcasts", "Gopher", "Interrogations"}
	sectionSeen := map[string]bool{"News": true, "Podcasts": true, "Gopher": true, "Interrogations": true}
	foldersBySection := map[string][]string{
		"News":     []string{"General"},
		"Podcasts": []string{"Search"},
		"Gopher":   []string{"Directory"},
	}
	folderSeen := map[string]bool{
		folderKey("News", "General"):     true,
		folderKey("Podcasts", "Search"):  true,
		folderKey("Gopher", "Directory"): true,
	}
	for _, folder := range m.folders {
		section := firstText(folder.Section, "News")
		if !sectionSeen[section] {
			sections = append(sections, section)
			sectionSeen[section] = true
		}
		key := folderKey(section, folder.Name)
		if !folderSeen[key] {
			foldersBySection[section] = append(foldersBySection[section], folder.Name)
			folderSeen[key] = true
		}
	}
	for _, feed := range m.feeds {
		section := firstText(feed.Section, sectionFromFeed(feed))
		folder := firstText(feed.Folder, folderFromFeed(feed))
		if !sectionSeen[section] {
			sections = append(sections, section)
			sectionSeen[section] = true
		}
		key := folderKey(section, folder)
		if !folderSeen[key] {
			foldersBySection[section] = append(foldersBySection[section], folder)
			folderSeen[key] = true
		}
	}
	for _, section := range sections {
		rows = append(rows, sourceRow{kind: sourceSection, section: section, title: section})
		if section == "Interrogations" {
			for i, out := range m.interrogations {
				rows = append(rows, sourceRow{kind: sourceInterrogation, section: section, aiIndex: i, title: interrogationTitle(out)})
			}
			continue
		}
		for _, folder := range foldersBySection[section] {
			collapsed := m.folderCollapsed(section, folder)
			rows = append(rows, sourceRow{kind: sourceFolder, section: section, folder: folder, title: folder, collapsed: collapsed})
			if collapsed {
				continue
			}
			for i, feed := range m.feeds {
				feedSection := firstText(feed.Section, sectionFromFeed(feed))
				feedFolder := firstText(feed.Folder, folderFromFeed(feed))
				if feedSection == section && feedFolder == folder {
					rows = append(rows, sourceRow{kind: sourceFeed, section: section, folder: folder, feedIndex: i, title: feed.Title, unread: feed.Unread})
				}
			}
		}
	}
	return rows
}

func interrogationTitle(out store.AIOutput) string {
	return firstText(out.ItemTitle, out.Prompt, "interrogation")
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

func (m *Model) revealPendingFeed() {
	if m.revealFeedID == 0 {
		return
	}
	_ = m.setFolderCollapsed(m.revealSection, m.revealFolder, false)
	for i, row := range m.sourceRows() {
		if row.kind == sourceFeed && m.feeds[row.feedIndex].ID == m.revealFeedID {
			m.sourceCursor = i
			m.feedCursor = row.feedIndex
			m.focus = focusFeeds
			m.ensureCursorVisible()
			break
		}
	}
	m.revealFeedID = 0
	m.revealSection = ""
	m.revealFolder = ""
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
	m.upsertLocalFolder(section, folder, collapsed)
	return nil
}

func folderKey(section, folder string) string {
	return section + "\x00" + folder
}
