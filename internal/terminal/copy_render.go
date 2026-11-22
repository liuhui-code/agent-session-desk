package terminal

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"
)

func (s Screen) copyRowRoles(row, width int) []int {
	roles := make([]int, width)
	var events [4][]int
	for i := range events {
		events[i] = make([]int, width+1)
	}
	for _, h := range s.CopyHighlights {
		start, end := max(0, h.Start), min(width-1, h.End)
		if h.Row != row || h.Kind < 0 || h.Kind >= len(events) || start > end {
			continue
		}
		events[h.Kind][start]++
		events[h.Kind][end+1]--
	}
	var counts [4]int
	for col := range roles {
		for kind := range counts {
			counts[kind] += events[kind][col]
		}
		// Native priority: current match, match, marked cell, marked line.
		for _, kind := range []int{1, 0, 3, 2} {
			if counts[kind] > 0 {
				roles[col] = kind + 2
				break
			}
		}
	}
	selection := s.Selection
	if selection == nil || row < selection.StartRow || row > selection.EndRow {
		return roles
	}
	start, end := 0, width-1
	switch selection.Mode {
	case "rect":
		start, end = min(selection.StartCol, selection.EndCol), max(selection.StartCol, selection.EndCol)
	case "line":
	default:
		if selection.StartRow == selection.EndRow {
			start, end = min(selection.StartCol, selection.EndCol), max(selection.StartCol, selection.EndCol)
		} else {
			if row == selection.StartRow {
				start = selection.StartCol
			}
			if row == selection.EndRow {
				end = selection.EndCol
			}
		}
	}
	for col := max(0, start); col <= min(width-1, end); col++ {
		roles[col] = 1 // Selection paints over search and marks.
	}
	return roles
}

func paintCopyRun(run Run, start int, roles []int, styles CopyStyles) string {
	text := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, run.Text)
	var out, segment strings.Builder
	role, col := -1, start
	flush := func() {
		if segment.Len() == 0 {
			return
		}
		style := run
		var overlay Run
		if role == 1 {
			overlay = styles.Selection
		} else if role >= 2 {
			overlay = styles.Highlights[role-2]
		}
		if overlay.Foreground != "" {
			style.Foreground = overlay.Foreground
		}
		if overlay.Background != "" {
			style.Background = overlay.Background
		}
		style.Flags |= overlay.Flags
		codes := runSGR(style)
		if codes != "" {
			out.WriteString("\x1b[" + codes + "m")
		}
		out.WriteString(segment.String())
		if codes != "" {
			out.WriteString("\x1b[0m")
		}
		segment.Reset()
	}
	var state byte
	for len(text) > 0 && col < len(roles) {
		cluster, cells, consumed, next := ansi.DecodeSequence(text, state, nil)
		if consumed <= 0 {
			break
		}
		current := roles[max(0, col)]
		if current != role {
			flush()
			role = current
		}
		segment.WriteString(cluster)
		col += cells
		text, state = text[consumed:], next
	}
	flush()
	return out.String()
}
