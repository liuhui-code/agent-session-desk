package ui

import (
	"context"
	"sort"
	"strings"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type groupService interface {
	Workspace(context.Context) (session.Workspace, error)
	CreateGroup(context.Context, string, string, string) (session.Group, error)
	MoveToGroup(context.Context, session.Session, string) (session.Session, error)
}
type sidebarItem struct {
	groupID      string
	index, level int // index -1 denotes a header; Session indices remain stable.
}

func (m *Model) sidebarGroups() []session.Group {
	if len(m.groups) > 0 {
		return m.groups
	}
	return []session.Group{{ID: session.DefaultGroupID, Name: "My Sessions", Revision: 1}}
}
func (m *Model) groupByID(id string) session.Group {
	if id == "" {
		id = session.DefaultGroupID
	}
	for _, g := range m.sidebarGroups() {
		if g.ID == id {
			return g
		}
	}
	return session.Group{}
}
func (m *Model) selectedGroup() session.Group {
	if v, ok := m.selected(); ok {
		return m.groupByID(session.GroupOf(v))
	}
	g := m.groupByID(m.selectedGroupID)
	if g.ID == "" {
		return m.groupByID(session.DefaultGroupID)
	}
	return g
}
func (m *Model) groupCollapsed(id string) bool {
	if id == session.DefaultGroupID {
		return m.sidebarCollapsed
	}
	return m.collapsedGroups[id]
}
func (m *Model) sidebarItems() []sidebarItem {
	byParent := map[string][]session.Group{}
	for _, g := range m.sidebarGroups() {
		byParent[g.ParentID] = append(byParent[g.ParentID], g)
	}
	var result []sidebarItem
	seen := map[string]bool{}
	var visit func(string, int)
	visit = func(parent string, level int) {
		if level >= session.MaxGroupDepth {
			return
		}
		for _, g := range byParent[parent] {
			if seen[g.ID] {
				continue
			}
			seen[g.ID] = true
			result = append(result, sidebarItem{groupID: g.ID, index: -1, level: level})
			if m.groupCollapsed(g.ID) {
				continue
			}
			for i, v := range m.sessions {
				if session.GroupOf(v) == g.ID {
					result = append(result, sidebarItem{groupID: g.ID, index: i, level: level + 1})
				}
			}
			visit(g.ID, level+1)
		}
	}
	visit("", 0)
	return result
}
func (m *Model) sidebarSelectedItem(items []sidebarItem) int {
	for i, item := range items {
		if item.index >= 0 && item.index == m.cursor {
			return i
		}
		if item.index < 0 && m.cursor < 0 && item.groupID == m.selectedGroup().ID {
			return i
		}
	}
	return 0
}
func (m *Model) selectSidebarItem(item sidebarItem) {
	m.cursor, m.selectedGroupID = item.index, item.groupID
}
func (m *Model) wantedGroupCollapse(id string, collapsed bool) {
	if id == session.DefaultGroupID {
		m.preferences.wanted.DefaultGroupCollapsed = collapsed
		return
	}
	ids := map[string]bool{}
	for _, id := range strings.Split(m.preferences.wanted.CollapsedGroups, ",") {
		if id != "" {
			ids[id] = true
		}
	}
	if collapsed {
		ids[id] = true
	} else {
		delete(ids, id)
	}
	var sorted []string
	for id := range ids {
		sorted = append(sorted, id)
	}
	sort.Strings(sorted)
	m.preferences.wanted.CollapsedGroups = strings.Join(sorted, ",")
}
func sortedCollapsedIDs(values map[string]bool) []string {
	var ids []string
	for id, collapsed := range values {
		if collapsed {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}
func (m *Model) revealGroup(id string) {
	for depth := 0; id != "" && depth < session.MaxGroupDepth; depth++ {
		if id == session.DefaultGroupID {
			m.sidebarCollapsed = false
		} else {
			delete(m.collapsedGroups, id)
		}
		m.wantedGroupCollapse(id, false)
		id = m.groupByID(id).ParentID
	}
	m.viewOffset = 0
}

func (m *Model) reconcileSidebarSelection(previous session.Session, hadSelection bool) {
	if !m.groupedSidebar() {
		return
	}
	if v, ok := m.selected(); ok && hadSelection && sameRun(v, previous) && session.GroupOf(v) != session.GroupOf(previous) {
		m.revealGroup(session.GroupOf(v))
	}
	for _, item := range m.sidebarItems() {
		if item.index == m.cursor && (item.index >= 0 || item.groupID == m.selectedGroup().ID) {
			return
		}
	}
	// Initial discovery respects persisted folds. If its first Session is
	// hidden, select its outermost collapsed ancestor instead of a hidden row.
	g := m.selectedGroup()
	target := g.ID
	for depth := 0; g.ID != "" && depth < session.MaxGroupDepth; depth++ {
		if m.groupCollapsed(g.ID) {
			target = g.ID
		}
		if g.ParentID == "" {
			break
		}
		g = m.groupByID(g.ParentID)
	}
	m.cursor, m.selectedGroupID = -1, target
}
