package ui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"strings"
	"testing"
	"time"
)

func TestAttentionCenterCapacityAndRenderBoundaries(t *testing.T) {
	h, s := attentionUIFixture(t)
	items := make([]session.AttentionItem, session.MaxAttentionItems+1)
	for i := range items {
		items[i] = s.item
		items[i].ID = fmt.Sprintf("%032x", i+1)
		items[i].Title = "中文 🧪 " + strings.Repeat("title", 40)
		items[i].Detail = strings.Repeat("x", 8192)
	}
	h.send(sessionsMsg{sessions: sampleSessions(), attention: items})
	if len(h.m.attention.items) != session.MaxAttentionItems || cap(h.m.attention.items) > session.MaxAttentionItems {
		t.Fatal("unbounded center snapshot")
	}
	h.send(tea.KeyPressMsg{Code: 'a', Mod: tea.ModAlt})
	h.until(func() bool { return h.m.attention.center != nil })
	for _, size := range [][2]int{{160, 40}, {120, 32}, {63, 20}, {49, 12}, {27, 8}, {12, 4}, {2, 1}, {0, 0}} {
		h.send(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		view := h.m.View()
		lines := strings.Split(ansi.Strip(view.Content), "\n")
		if size[1] > 0 && len(lines) > size[1] {
			t.Fatal("height overflow", size)
		}
		for _, line := range lines {
			if ansi.StringWidth(line) > size[0] {
				t.Fatal("width overflow", size)
			}
		}
		if view.Cursor != nil {
			t.Fatal("modal exposed terminal cursor")
		}
	}
	// Copy from an overlapping old snapshot, then release every removed detail.
	h.m.acceptAttentionSnapshot(h.m.attention.items[:1])
	if h.m.attention.items[0].ID != items[0].ID {
		t.Fatal("overlap copy lost identity")
	}
	for _, item := range h.m.attention.items[1:cap(h.m.attention.items)] {
		if item.Detail != "" {
			t.Fatal("removed detail retained")
		}
	}
}
func BenchmarkAttentionCenterCapacityRender(b *testing.B) {
	m := New(context.Background(), nil)
	defer m.Close()
	m.width, m.height = 120, 32
	m.workspace = &workspaceState{}
	m.attention.center = &attentionCenter{}
	items := make([]session.AttentionItem, session.MaxAttentionItems)
	for i := range items {
		items[i] = session.AttentionItem{AttentionEvent: session.AttentionEvent{ID: fmt.Sprintf("%032x", i+1), Title: strings.Repeat("x", 256), Detail: strings.Repeat("d", 8192), Kind: "approval", State: "pending"}, Source: sampleSessions()[0], CreatedAt: time.Now()}
	}
	m.acceptAttentionSnapshot(items)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}
