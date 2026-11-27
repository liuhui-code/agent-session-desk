package ui

import (
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
	"strings"
)

// Defensive retention bounds, not a resource acceptance budget. These are
// display snapshots only: no client, history, cursor authority or disk writes.
const workspaceSnapshotLimit = 4
const workspaceSnapshotBytes = 512 * 1024

type workspaceSnapshot struct {
	key    string
	bytes  int
	screen terminal.Screen
}

func screenRetentionBytes(s terminal.Screen) int {
	if s.Width < 1 || s.Height < 1 || s.CopyMode || s.CopyPrompt != nil || s.Selection != nil || s.CopyCursor != nil || len(s.CopyHighlights) > 0 {
		return workspaceSnapshotBytes + 1
	}
	// Charge generously for slice/object headers, allocator rounding and strings.
	n := 512 + 2*len(s.Status)
	for _, row := range s.Rows {
		n += 64
		for _, r := range row.Runs {
			n += 128 + 2*(len(r.Text)+len(r.Foreground)+len(r.Background))
		}
	}
	return n
}
func (w *workspaceState) rememberScreen() {
	if w.conn == nil || w.busy || w.err != nil || w.active < 0 || w.active >= len(w.tabs) {
		return
	}
	key := activityKey(w.tabs[w.active].session)
	size := screenRetentionBytes(w.screen)
	if size > workspaceSnapshotBytes {
		return
	}
	var next []workspaceSnapshot
	total := size
	for _, entry := range w.snapshots {
		if entry.key != key && len(next) < workspaceSnapshotLimit-1 && total+entry.bytes <= workspaceSnapshotBytes {
			next = append(next, entry)
			total += entry.bytes
		}
	}
	// Clone owned slices/strings: never retain a larger frame's backing storage.
	s := w.screen
	s.Status = strings.Clone(s.Status)
	s.Rows = append([]terminal.Row(nil), s.Rows...)
	for i := range s.Rows {
		s.Rows[i].Runs = append([]terminal.Run(nil), s.Rows[i].Runs...)
		for j := range s.Rows[i].Runs {
			r := &s.Rows[i].Runs[j]
			r.Text = strings.Clone(r.Text)
			r.Foreground = strings.Clone(r.Foreground)
			r.Background = strings.Clone(r.Background)
		}
	}
	w.snapshots = append([]workspaceSnapshot{{key: key, bytes: size, screen: s}}, next...)
}
func (w *workspaceState) lastScreen() terminal.Screen {
	if w.active < 0 || w.active >= len(w.tabs) || !w.tabs[w.active].live || w.tabs[w.active].session.TerminalState != session.TerminalAvailable {
		return terminal.Screen{}
	}
	key := activityKey(w.tabs[w.active].session)
	for _, entry := range w.snapshots {
		if entry.key == key {
			return entry.screen
		}
	}
	return terminal.Screen{}
}

func (w *workspaceState) closeViews() {
	w.disconnect()
	w.buffered = nil
	w.snapshots = nil
}
func (m *Model) workspaceConnecting() bool {
	return m.workspace != nil && (m.workspace.busy || m.workspace.transition == workspaceSwitch)
}
