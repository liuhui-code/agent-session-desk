package terminal

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"
)

// Render emits only text and generated SGR character attributes, never raw VT
// received from the child. These attributes are terminal content, not UI chrome.
func (s Screen) Render(width, height int) string {
	width = min(max(0, width), MaxWidth)
	height = min(max(0, height), MaxHeight)
	if width == 0 || height == 0 {
		return ""
	}
	rows := make([]string, height)
	for i := range rows {
		var line strings.Builder
		if i < len(s.Rows) {
			for _, run := range s.Rows[i].Runs {
				text := strings.Map(func(r rune) rune {
					if unicode.IsControl(r) {
						return -1
					}
					return r
				}, run.Text)
				codes := runSGR(run)
				if codes != "" {
					line.WriteString("\x1b[" + codes + "m")
				}
				line.WriteString(text)
				if codes != "" {
					line.WriteString("\x1b[0m")
				}
			}
		}
		row := ansi.Truncate(line.String(), width, "")
		row += strings.Repeat(" ", max(0, width-ansi.StringWidth(row)))
		rows[i] = row
	}
	return strings.Join(rows, "\n")
}

func runSGR(run Run) string {
	var codes []string
	// psmux's flags encode dim,bold,italic,underline,inverse,blink,hidden,strike.
	for i, code := range []string{"2", "1", "3", "4", "7", "5", "8", "9"} {
		if run.Flags&(1<<i) != 0 {
			codes = append(codes, code)
		}
	}
	if color := colorSGR(run.Foreground, 38); color != "" {
		codes = append(codes, color)
	}
	if color := colorSGR(run.Background, 48); color != "" {
		codes = append(codes, color)
	}
	return strings.Join(codes, ";")
}

func colorSGR(value string, kind int) string {
	if value == "default" || value == "" {
		return ""
	}
	if suffix, ok := strings.CutPrefix(value, "idx:"); ok {
		n, err := strconv.Atoi(suffix)
		if err == nil && n >= 0 && n <= 255 {
			return fmt.Sprintf("%d;5;%d", kind, n)
		}
	}
	if suffix, ok := strings.CutPrefix(value, "rgb:"); ok {
		parts := strings.Split(suffix, ",")
		if len(parts) != 3 {
			return ""
		}
		var rgb [3]int
		for i, part := range parts {
			n, err := strconv.Atoi(part)
			if err != nil || n < 0 || n > 255 {
				return ""
			}
			rgb[i] = n
		}
		return fmt.Sprintf("%d;2;%d;%d;%d", kind, rgb[0], rgb[1], rgb[2])
	}
	return ""
}
