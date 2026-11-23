package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestDialogLongFieldsKeepVisibleCaretAndValueAcrossResize(t *testing.T) {
	for _, mode := range []string{"launch", "launch-program", "launch-dir", "launch-args", "rename", "group-name", "group-dir", "group-rename", "search"} {
		for _, pattern := range []string{"abcdef/", "中文e\u0301👩‍💻1️⃣/"} {
			for _, middle := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/middle=%v", mode, pattern, middle), func(t *testing.T) {
					m := New(context.Background(), &workspaceFixture{})
					t.Cleanup(m.Close)
					m.width, m.height = 120, 32
					field := longFieldDialog(m, mode)
					value := strings.Repeat(pattern, 6) + "ZX" + strings.Repeat("tail/", 3) + "Q"
					field.SetValue(value)
					position := len([]rune(value))
					if middle {
						position = len([]rune(strings.Repeat(pattern, 6) + "Z"))
					}
					field.SetCursor(position)
					for _, size := range [][2]int{{120, 32}, {63, 25}, {49, 25}, {27, 12}, {120, 12}, {49, 12}, {120, 32}} {
						m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
						view := m.View()
						if field.Value() != value || field.Position() != position {
							t.Fatalf("%v resize changed value or logical caret: %q / %d", size, field.Value(), field.Position())
						}
						rows := strings.Split(view.Content, "\n")
						if len(rows) != size[1] {
							t.Fatalf("%v frame height %d", size, len(rows))
						}
						for _, row := range rows {
							if ansi.StringWidth(row) != size[0] || strings.ContainsRune(row, '\x00') {
								t.Fatalf("%v invalid row %q", size, row)
							}
						}
						c := view.Cursor
						if c == nil || c.X < 0 || c.X >= size[0] || c.Y < 0 || c.Y >= size[1] {
							t.Fatalf("%v focused caret missing/outside: %+v", size, c)
						}
						row := ansi.Strip(rows[c.Y])
						marker := "Q"
						if middle {
							marker = "Z"
						}
						at := strings.Index(row, marker)
						if at < 0 || ansi.StringWidth(row[:at+len(marker)]) != c.X {
							t.Fatalf("%v caret no longer after original %s: %+v, row %q", size, marker, c, row)
						}
					}
					m.Update(tea.KeyPressMsg{Code: '!', Text: "!"})
					runes := []rune(value)
					want := string(runes[:position]) + "!" + string(runes[position:])
					if field.Value() != want || field.Position() != position+1 {
						t.Fatal("edit after resize changed the original insertion target")
					}
				})
			}
		}
	}
}

func longFieldDialog(m *Model, mode string) *textinput.Model {
	switch mode {
	case "launch", "launch-program", "launch-dir", "launch-args":
		m.openLaunchDialog()
		index := map[string]int{"launch": 0, "launch-program": 1, "launch-dir": 2, "launch-args": 3}[mode]
		m.focusLaunchField(index)
		return &m.launch.fields[index]
	case "rename":
		s := sampleSessions()[0]
		m.renameTarget = &s
		m.openSessionRename()
		return &m.rename.input
	case "group-name", "group-dir", "group-rename":
		m.groupIntent = &groupDialog{rename: mode == "group-rename"}
		m.openGroupDialog()
		if mode == "group-dir" {
			m.groupDialog.inputs[0].Blur()
			m.groupDialog.focus = 1
			m.groupDialog.inputs[1].Focus()
		}
		return &m.groupDialog.inputs[m.groupDialog.focus]
	default:
		m.openSessionSearch()
		return &m.search.input
	}
}

func TestKeycapDisplayCaretKeepsLogicalRuneInsertionPosition(t *testing.T) {
	field := textinput.New()
	field.SetVirtualCursor(false)
	field.SetWidth(8)
	field.SetValue("1️⃣X")
	field.Focus()
	for position := 0; position <= len([]rune(field.Value())); position++ {
		field.SetCursor(position)
		view, cursor := inputFrame(field)
		want := ansi.StringWidth(field.Prompt) + ansi.StringWidth(string([]rune(field.Value())[:position]))
		if cursor == nil || cursor.X != want || field.Position() != position || !strings.Contains(ansi.Strip(view), "1️⃣X") {
			t.Fatalf("rune insertion %d projects to %+v, want cell %d, view %q", position, cursor, want, view)
		}
	}
}
