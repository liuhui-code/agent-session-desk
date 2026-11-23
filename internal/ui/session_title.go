package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// Names remain presentation data. When their visible forms collide, reserve
// room for a distinguishing Session ID prefix; controls still bind full IDs.
func (m *Model) sessionTitle(v session.Session, peers []session.Session, width int) string {
	width = max(0, width)
	name := sessionName(v)
	short := cellTruncate(name, width, m.tokens.marks.ellipsis)
	collision := false
	for _, p := range peers {
		if p.ID != v.ID && cellTruncate(sessionName(p), width, m.tokens.marks.ellipsis) == short {
			collision = true
			break
		}
	}
	if !collision {
		return short
	}
	n := min(len(v.ID), m.tokens.marks.identityPrefixLength)
	for n < len(v.ID) {
		unique := true
		for _, p := range peers {
			if p.ID != v.ID && strings.HasPrefix(p.ID, v.ID[:n]) {
				unique = false
				break
			}
		}
		if unique {
			break
		}
		n++
	}
	suffix := m.tokens.marks.identityOpen + label(v.ID[:n]) + m.tokens.marks.identityClose
	suffixWidth := lipgloss.Width(suffix)
	if suffixWidth >= width {
		return ansi.Cut(suffix, max(0, suffixWidth-width), suffixWidth)
	}
	return cellTruncate(name, width-suffixWidth, m.tokens.marks.ellipsis) + suffix
}

func sessionName(v session.Session) string {
	if title := label(v.Title); title != "" {
		return title
	}
	return label(v.ID)
}

func (m *Model) titlePeers() []session.Session {
	items := append([]session.Session(nil), m.sessions...)
	for _, tab := range m.workspace.tabs {
		items = append(items, tab.session)
	}
	return items
}
