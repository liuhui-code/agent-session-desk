package ui

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

type deliveryTerminal struct{ *workspaceFixture }

func (s deliveryTerminal) Stop(context.Context, session.Session) error {
	panic("delivery must not stop a Run")
}

func deliveryHarness(t *testing.T) (*workspaceHarness, *workspaceFixture, *store.Store, session.Message) {
	t.Helper()
	rows := sampleSessions()
	for i := range rows {
		rows[i].ID = strings.Repeat(string(rune('1'+i*2)), 32)
		rows[i].RunID = strings.Repeat(string(rune('2'+i*2)), 32)
	}
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return rows, nil }}}
	db, e := store.Open(context.Background(), filepath.Join(t.TempDir(), "metadata.sqlite3"), "ui")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	m := manager.New(deliveryTerminal{service}, db)
	if _, e = m.List(context.Background()); e != nil {
		t.Fatal(e)
	}
	record, e := db.RecordOutput(context.Background(), rows[1], session.Preview{Text: "literal $() ` ;\n\u4e2d\u6587"})
	if e != nil {
		t.Fatal(e)
	}
	msg, e := m.QueueMessage(context.Background(), session.MessageRequest{RequestID: strings.Repeat("a", 32), TargetSessionID: rows[0].ID, TargetRunID: rows[0].RunID, SourceRunID: rows[1].RunID, SourceRecordID: record.ID})
	if e != nil {
		t.Fatal(e)
	}
	h := newWorkspaceHarness(t, m)
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	return h, service, db, msg
}
func openDelivery(t *testing.T, h *workspaceHarness) {
	h.until(func() bool { return h.m.workspace.pending == 0 })
	t.Helper()
	h.send(tea.KeyPressMsg{Code: 'm', Text: "m"})
	h.until(func() bool { return h.m.workspace.delivery != nil && !h.m.workspace.delivery.loading })
}

func TestMessageDeliveryDefaultsToCancelAndUsesOriginalFIFOOnce(t *testing.T) {
	h, s, db, msg := deliveryHarness(t)
	openDelivery(t, h)
	if box, fits := h.m.messageDeliveryBox(); !fits {
		t.Fatalf("source review does not fit normal workspace: %s", box)
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	if h.m.workspace.delivery != nil {
		t.Fatal("default Enter did not cancel")
	}
	before, e := db.Message(context.Background(), msg.ID)
	if e != nil || before.State != "queued" {
		t.Fatal("cancel claimed or wrote message")
	}
	openDelivery(t, h)
	h.send(tea.PasteMsg{Content: "NEVER INJECT THIS"})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool {
		return h.m.workspace.delivery != nil && h.m.workspace.delivery.submitted && !h.m.workspace.delivery.busy
	})
	got, e := db.Message(context.Background(), msg.ID)
	if e != nil || got.State != "input_written" {
		t.Fatalf("receipt: %+v %v", got, e)
	}
	s.mu.Lock()
	stream := s.streams[msg.Target.ID]
	s.mu.Unlock()
	stream.mu.Lock()
	writes := strings.Join(stream.writes, ";")
	stream.mu.Unlock()
	if strings.Count(writes, "paste:"+msg.Body) != 1 || strings.Contains(writes, "key:Enter") || strings.Contains(writes, "NEVER") {
		t.Fatalf("input boundaries: %q", writes)
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.send(tea.KeyPressMsg{Code: 'j', Mod: tea.ModAlt})
	if h.m.workspace.active != 0 {
		t.Fatal("completion dialog leaked navigation")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	if h.m.workspace.delivery != nil || h.m.workspace.focused {
		t.Fatal("closing delivery changed management focus")
	}
}
func TestMessageDeliveryNarrowReviewCannotConfirm(t *testing.T) {
	h, _, db, msg := deliveryHarness(t)
	openDelivery(t, h)
	h.send(tea.WindowSizeMsg{Width: 49, Height: 12})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	got, e := db.Message(context.Background(), msg.ID)
	if e != nil || got.State != "queued" {
		t.Fatal("unreviewable target was delivered")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
}

func TestMessageDeliveryPreviewPagesEntireBodyWithoutLabelTruncation(t *testing.T) {
	h, _, _, _ := deliveryHarness(t)
	openDelivery(t, h)
	h.m.workspace.delivery.message.Body = strings.Repeat("line\n", 1500) + "TAIL_MARKER"
	h.send(tea.KeyPressMsg{Code: tea.KeyEnd})
	if !strings.Contains(h.m.View().Content, "TAIL_MARKER") {
		t.Fatal("body tail is not reviewable")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyUp})
	if !strings.Contains(h.m.View().Content, "line") {
		t.Fatal("scroll from end did not move")
	}
}
