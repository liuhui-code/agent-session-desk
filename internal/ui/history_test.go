package ui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func historyRows() []session.HistoryRecord {
	return []session.HistoryRecord{{Session: session.Session{ID: "historical-session", RunID: "historical-run", Title: "History任务", Command: "agent.exe", TerminalState: session.TerminalUnknown}, RecordedAt: time.Now().UTC()}}
}

func TestHistoryOpensOnDemandAndEnterNeverAttaches(t *testing.T) {
	reads, attaches := 0, 0
	m := testModel(t, fakeService{history: func(context.Context, int) ([]session.HistoryRecord, error) { reads++; return historyRows(), nil }, attach: func(context.Context, session.Session) (*exec.Cmd, error) { attaches++; return nil, nil }})
	m.cursor = 1
	m.preview = "LIVE_PREVIEW"
	if reads != 0 {
		t.Fatal("history preloaded")
	}
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'h'})
	if cmd == nil {
		t.Fatal("history did not request metadata")
	}
	m.Update(cmd())
	view := ansi.Strip(m.View().Content)
	if reads != 1 || !strings.Contains(view, "History任务") || strings.Contains(view, "LIVE_PREVIEW") {
		t.Fatalf("history view: %s", view)
	}
	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		m.Update(cmd())
	}
	view = ansi.Strip(m.View().Content)
	if attaches != 0 || !strings.Contains(view, "historical-run") || !strings.Contains(view, "只读") {
		t.Fatalf("history Enter was not read-only: %s", view)
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.cursor != 1 || strings.Contains(ansi.Strip(m.View().Content), "History任务") {
		t.Fatal("history changed live selection or did not close")
	}
}

func TestHistoryCancellationDropsLateResultsAndBoundsRequests(t *testing.T) {
	first := make(chan context.Context, 1)
	release := make(chan struct{})
	result := make(chan tea.Msg, 1)
	reads := 0
	m := testModel(t, fakeService{history: func(ctx context.Context, _ int) ([]session.HistoryRecord, error) {
		if reads == 0 {
			reads++
			first <- ctx
			<-release
			rows := historyRows()
			rows[0].Session.Title = "STALE_HISTORY"
			return rows, nil
		}
		reads++
		return historyRows(), nil
	}})
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'h'})
	go func() { result <- cmd() }()
	ctx := <-first
	m.Update(tea.KeyPressMsg{Code: 'h'})
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("closing history did not cancel its read")
	}
	_, next := m.Update(tea.KeyPressMsg{Code: 'h'})
	if next != nil {
		t.Fatal("reopen started another concurrent read")
	}
	close(release)
	_, next = m.Update(<-result)
	if strings.Contains(ansi.Strip(m.View().Content), "STALE_HISTORY") {
		t.Fatal("late history result replaced current view")
	}
	if next == nil {
		t.Fatal("reopen did not resume after cancelled read finished")
	}
	m.Update(next())
	if reads != 2 || !strings.Contains(ansi.Strip(m.View().Content), "History任务") {
		t.Fatal("fresh history missing")
	}
}

func TestHistoryPagingKeepsChosenRecordVisibleAndPreservesLiveCursor(t *testing.T) {
	rows := []session.HistoryRecord{}
	for i := 0; i < 100; i++ {
		rows = append(rows, session.HistoryRecord{Session: session.Session{ID: fmt.Sprintf("session-%03d", i), RunID: fmt.Sprintf("run-%03d", i), Title: fmt.Sprintf("history-%03d", i)}})
	}
	m := testModel(t, fakeService{history: func(context.Context, int) ([]session.HistoryRecord, error) { return rows, nil }})
	m.height = 12
	m.cursor = 1
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'h'})
	m.Update(cmd())
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	if !strings.Contains(ansi.Strip(m.View().Content), "› history-099") {
		t.Fatal("last history selection is not visible")
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !strings.Contains(ansi.Strip(m.View().Content), "run-099") {
		t.Fatal("detail is not bound to chosen Run")
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m.Update(tea.KeyPressMsg{Code: tea.KeyHome})
	m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	if m.history.cursor == 0 || !strings.Contains(ansi.Strip(m.View().Content), fmt.Sprintf("› history-%03d", m.history.cursor)) {
		t.Fatal("page navigation did not keep selected item visible")
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyPgUp})
	if m.history.cursor != 0 || m.cursor != 1 {
		t.Fatal("history paging changed live selection")
	}
}

func TestHistoryDetailsWrapScrollAndKeepObservationsSeparate(t *testing.T) {
	rows := historyRows()
	stopped := time.Date(2026, 11, 22, 3, 4, 5, 0, time.UTC)
	rows[0].RecordedAt = stopped.Add(-time.Hour)
	rows[0].StopConfirmedAt = &stopped
	rows[0].PaneExit = &session.PaneExit{PaneID: "%7", Source: "psmux.pane_dead", ObservedAt: stopped.Add(-time.Minute)}
	rows[0].Session.Title = "\x1b]52;c;ATTACK\a历史\n任务"
	rows[0].Session.Command = strings.Repeat("界", 40) + ".exe"
	m := testModel(t, fakeService{history: func(context.Context, int) ([]session.HistoryRecord, error) { return rows, nil }})
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'h'})
	m.Update(cmd())
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	full := ansi.Strip(m.View().Content)
	for _, want := range []string{"historical-session", "historical-run", "agent", "2026-11-22T03:04:05Z", "psmux.pane_dead", "%7", "Exit code: unknown"} {
		if !strings.Contains(full, want) {
			t.Fatalf("missing detail %q: %s", want, full)
		}
	}
	if strings.Contains(m.View().Content, "ATTACK") {
		t.Fatal("metadata replayed terminal control")
	}
	m.Update(tea.WindowSizeMsg{Width: 22, Height: 10})
	original := ansi.Strip(m.View().Content)
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	end := ansi.Strip(m.View().Content)
	if original == end || !strings.Contains(end, "Exit code: unknown") {
		t.Fatalf("cannot scroll wrapped details: %s", end)
	}
	if m.history.cursor != 0 {
		t.Fatal("detail scrolling selected a different Run")
	}
	for _, line := range strings.Split(m.View().Content, "\n") {
		if ansi.StringWidth(line) != 22 {
			t.Fatalf("detail width %d", ansi.StringWidth(line))
		}
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyHome})
	if ansi.Strip(m.View().Content) != original {
		t.Fatal("Home did not restore start of details")
	}
}

func TestHistoryRefreshPreservesRunIdentityAndHandlesErrors(t *testing.T) {
	rows := historyRows()
	second := rows[0]
	second.Session.RunID = "second-run"
	second.Session.Title = "SECOND"
	rows = append(rows, second)
	var readErr error
	reads := 0
	m := testModel(t, fakeService{history: func(context.Context, int) ([]session.HistoryRecord, error) { reads++; return rows, readErr }})
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'h'})
	m.Update(cmd())
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	rows = []session.HistoryRecord{second, rows[0]}
	_, cmd = m.Update(tea.KeyPressMsg{Code: 'r'})
	m.Update(cmd())
	if m.history.cursor != 0 || !m.history.detail || !strings.Contains(ansi.Strip(m.View().Content), "second-run") {
		t.Fatal("refresh changed the chosen Run")
	}
	readErr = fmt.Errorf("metadata unavailable\x1b]52;c;PRIVATE\a")
	_, cmd = m.Update(tea.KeyPressMsg{Code: 'r'})
	m.Update(cmd())
	view := m.View().Content
	if !strings.Contains(ansi.Strip(view), "读取History失败") || strings.Contains(view, "\x1b]52;") {
		t.Fatal("history failure missing or unsafe")
	}
	readErr = nil
	rows = rows[1:]
	_, cmd = m.Update(tea.KeyPressMsg{Code: 'r'})
	m.Update(cmd())
	if m.history.detail || m.history.cursor != 0 || reads != 4 {
		t.Fatal("removed Run kept stale detail or refresh not retried")
	}
	rows = nil
	_, cmd = m.Update(tea.KeyPressMsg{Code: 'r'})
	m.Update(cmd())
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.history.detail || !strings.Contains(ansi.Strip(m.View().Content), "No history records") {
		t.Fatal("empty history opened details")
	}
}

func TestHistoryRemainsAccessibleWhenLiveBackendFailsAndQuitCancelsRead(t *testing.T) {
	var ctx context.Context
	reads := 0
	m := testModel(t, fakeService{history: func(c context.Context, limit int) ([]session.HistoryRecord, error) {
		ctx = c
		reads++
		if limit != 100 {
			t.Fatalf("read limit=%d", limit)
		}
		return historyRows(), nil
	}})
	m.listErr = fmt.Errorf("backend missing")
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'h'})
	_, repeated := m.Update(tea.KeyPressMsg{Code: 'r'})
	if repeated != nil {
		t.Fatal("refresh overlapped pending read")
	}
	m.Update(cmd())
	if !strings.Contains(ansi.Strip(m.View().Content), "History任务") {
		t.Fatal("backend failure blocked history")
	}
	_, cmd = m.Update(tea.KeyPressMsg{Code: 'r'})
	// Execute in the test to observe the exact context supplied to the service.
	result := cmd()
	m.Update(tea.KeyPressMsg{Code: 'q'})
	if ctx.Err() == nil || reads != 2 {
		t.Fatal("quit did not cancel history context")
	}
	if _, next := m.Update(result); next != nil {
		t.Fatal("closed model accepted late history result")
	}
}

func TestHistoryViewportBoundsForListDetailAndTokenSpacing(t *testing.T) {
	m := testModel(t, fakeService{})
	m.history.open = true
	m.history.rows = historyRows()
	m.history.rows[0].Session.Title = "中🙂文\x1b]52;c;PRIVATE\a\n任务"
	for _, detail := range []bool{false, true} {
		m.history.detail = detail
		for _, w := range []int{1, 2, 10, 22, 49, 50, 79, 80, 100, 160} {
			for _, h := range []int{1, 4, 5, 8, 10, 24, 40} {
				m.width, m.height = w, h
				view := m.View().Content
				lines := strings.Split(view, "\n")
				if len(lines) != h || strings.Contains(view, "\x1b]52;") {
					t.Fatalf("unsafe/wrong height at %dx%d", w, h)
				}
				for _, line := range lines {
					if ansi.StringWidth(line) != w {
						t.Fatalf("wrong width at %dx%d: %q", w, h, line)
					}
				}
			}
		}
	}
	m.width, m.height = 160, 24
	m.tokens.layout.itemGap = 7
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "       r") {
		t.Fatal("history footer ignores item gap token")
	}
}
