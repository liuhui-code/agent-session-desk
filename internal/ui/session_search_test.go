package ui

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestSessionSearchPreservesPinnedPlaceholderAndKeyboardHintStyles(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '/', Text: "/"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "Local session search") })
	view := h.m.View().Content
	if !strings.Contains(view, "38;5;240") {
		t.Error("placeholder lost pinned Bubbles v1 index-240 color during v2 migration")
	}
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(ansi.Strip(line), "Enter Open") && regexp.MustCompile(`\x1b\[(?:[0-9]+;)*3(?:;[0-9]+)*m`).MatchString(line) {
			t.Error("keyboard legend inherited italic styling from empty-query tip")
		}
	}
}

func TestSessionSearchOpensOriginalRunWithoutSendingQueryToTerminal(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: 'd', Text: "draft-preserved"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '/', Text: "/"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "Local session search") })
	h.send(tea.PasteMsg{Content: "评审"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker second") })
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.starts) != 0 || strings.Join(service.opened, ",") != "first,second" {
		t.Fatalf("search restarted or opened wrong target: starts=%v opened=%v", service.starts, service.opened)
	}
	for _, stream := range service.streams {
		stream.mu.Lock()
		writes := strings.Join(stream.writes, ",")
		stream.mu.Unlock()
		if strings.Contains(writes, "评审") || strings.Contains(writes, "key:Enter") || strings.Contains(writes, "text:/") {
			t.Fatalf("search input leaked into terminal: %q", writes)
		}
	}
	first := service.streams["first"]
	first.mu.Lock()
	defer first.mu.Unlock()
	if strings.Count(strings.Join(first.writes, ","), "text:draft-preserved") != 1 {
		t.Fatal("opening search lost or duplicated the accepted draft")
	}
}

func TestSessionSearchWaitsForAcceptedTerminalWrite(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil })
	service.mu.Lock()
	stream := service.streams["first"]
	service.mu.Unlock()
	gate := make(chan struct{})
	stream.mu.Lock()
	stream.gate = gate
	stream.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "pending-draft"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '/', Text: "/"})
	if strings.Contains(h.m.View().Content, "Local session search") {
		t.Fatal("search took focus before accepted terminal write completed")
	}
	select {
	case <-stream.closed:
		t.Fatal("search cancelled the pending write")
	default:
	}
	close(gate)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "Local session search") })
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if strings.Count(strings.Join(stream.writes, ","), "text:pending-draft") != 1 {
		t.Fatal("draft was lost or replayed")
	}
}

func TestSessionSearchKeepsSelectionByRunWhenDiscoveryReorders(t *testing.T) {
	items := sampleSessions()
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return items, nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '/', Text: "/"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "Local session search") })
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(sessionsMsg{generation: h.m.listGeneration, sessions: []session.Session{items[1], items[0]}})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker second") })
}

func TestSessionSearchKeepsTerminalSlashAndModalNavigationSeparate(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: '/', Text: "/"})
	h.until(func() bool {
		service.mu.Lock()
		stream := service.streams["first"]
		service.mu.Unlock()
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return strings.Contains(strings.Join(stream.writes, ","), "text:/")
	})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '/', Text: "/"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "Local session search") })
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.MouseClickMsg{Button: tea.MouseLeft, X: 15, Y: 1})
	h.send(tea.PasteMsg{Content: "q"})
	if h.m.closed || !strings.Contains(h.m.View().Content, "Local session search") {
		t.Fatal("modal input changed application focus or quit")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	h.until(func() bool { return h.m.workspace.conn != nil })
	if h.m.workspace.focused || h.m.workspace.tabs[h.m.workspace.active].session.ID != "first" {
		t.Fatal("cancel switched active terminal or granted input focus")
	}
}

func TestSessionSearchFollowsRankedResultsAndFitsBoundedWindows(t *testing.T) {
	items := make([]session.Session, 16)
	for i := range items {
		items[i] = session.Session{ID: fmt.Sprintf("target-%d", i), RunID: fmt.Sprintf("run-%d", i), Title: fmt.Sprintf("Review-%02d 中文", i), Command: `C:\tools\pwsh.exe`, TerminalState: session.TerminalAvailable}
	}
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return items, nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '/', Text: "/"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "Local session search") })
	for i := 0; i < 15; i++ {
		h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	if !strings.Contains(ansi.Strip(h.m.View().Content), "› Review-15") {
		t.Fatal("result window did not follow selected item")
	}
	for _, size := range [][2]int{{120, 32}, {63, 25}, {49, 25}, {27, 10}, {10, 4}, {1, 1}, {0, 0}} {
		h.send(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		v := h.m.View()
		if size[0] == 0 || size[1] == 0 {
			if v.Content != "" || v.Cursor != nil {
				t.Fatal("zero viewport retained content/cursor")
			}
			continue
		}
		rows := strings.Split(v.Content, "\n")
		if len(rows) != size[1] {
			t.Fatalf("height %d != %d", len(rows), size[1])
		}
		for _, row := range rows {
			if ansi.StringWidth(row) != size[0] {
				t.Fatalf("row width %d != %d", ansi.StringWidth(row), size[0])
			}
		}
		if v.Cursor != nil && (v.Cursor.X < 0 || v.Cursor.X >= size[0] || v.Cursor.Y < 0 || v.Cursor.Y >= size[1]) {
			t.Fatal("cursor outside viewport")
		}
		if size[0] >= 49 && !strings.Contains(ansi.Strip(v.Content), "› Review-15") {
			t.Fatal("resizing lost selected result")
		}
	}
}

func TestSessionSearchRanksExactBeforeFuzzyAndKeepsUnicodeRules(t *testing.T) {
	for _, tc := range []struct {
		name, query string
		want        string
	}{
		{"exact before typo", "  REVIEW  ", "exact"},
		{"program path", "pwsh.exe", "fuzzy"},
		{"short Unicode exact", "评审", "unicode"},
		{"no canonical normalization", "e\u0301", ""},
		{"empty query original order", "   ", "fuzzy"},
		{"no result", "zzzzzzzzz", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			items := []session.Session{{ID: "fuzzy", RunID: "a", Title: "Reviwe", Command: `C:\tools\pwsh.exe`, TerminalState: session.TerminalAvailable}, {ID: "exact", RunID: "b", Title: "Review", TerminalState: session.TerminalAvailable}, {ID: "unicode", RunID: "c", Title: "评审 é", TerminalState: session.TerminalAvailable}}
			service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return items, nil }}}
			h := newWorkspaceHarness(t, service)
			h.until(func() bool { return h.m.workspace.conn != nil })
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			h.send(tea.KeyPressMsg{Code: '/', Text: "/"})
			h.until(func() bool { return strings.Contains(h.m.View().Content, "Local session search") })
			h.send(tea.PasteMsg{Content: tc.query})
			h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
			if tc.want == "" {
				if !strings.Contains(h.m.View().Content, "Local session search") || !strings.Contains(h.m.View().Content, "No matching sessions") {
					t.Fatal("empty result activated hidden target")
				}
				return
			}
			h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker "+tc.want) })
			if h.m.workspace.tabs[h.m.workspace.active].session.ID != tc.want {
				t.Fatal("ranked result activated wrong identity")
			}
		})
	}
}

func TestSessionSearchDoesNotRetargetWhenSelectedRunChangesDuringRefresh(t *testing.T) {
	items := sampleSessions()
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return items, nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '/', Text: "/"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "Local session search") })
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	replacement := append([]session.Session(nil), items...)
	replacement[1].RunID = "replacement-run"
	h.send(sessionsMsg{generation: h.m.listGeneration, sessions: replacement})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !strings.Contains(h.m.View().Content, "Local session search") {
		t.Fatal("refresh silently retargeted Enter after the selected Run was replaced")
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	if strings.Join(service.opened, ",") != "first" {
		t.Fatalf("unexpected terminal open: %v", service.opened)
	}
}
