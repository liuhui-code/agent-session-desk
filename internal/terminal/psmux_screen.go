package terminal

import (
	"encoding/json"
	"errors"
	"fmt"
	"unicode"
	"unicode/utf8"
)

func (c *PSMux) readScreen() (Screen, error) {
	for c.scanner.Scan() {
		line := c.scanner.Bytes()
		if string(line) == "NC" || len(line) == 0 {
			continue
		}
		var frame struct {
			Session string `json:"session_name"`
			Status  string `json:"status_message"`
			Layout  struct {
				Type       string `json:"type"`
				ID         int    `json:"id"`
				Rows       int    `json:"rows"`
				Cols       int    `json:"cols"`
				CursorRow  int    `json:"cursor_row"`
				CursorCol  int    `json:"cursor_col"`
				Hidden     bool   `json:"hide_cursor"`
				Shape      int    `json:"cursor_shape"`
				Active     bool   `json:"active"`
				Alternate  bool   `json:"alternate_screen"`
				Mouse      bool   `json:"wants_mouse"`
				CopyMode   bool   `json:"copy_mode"`
				Offset     uint64 `json:"scroll_offset"`
				ViewOffset uint64 `json:"view_offset"`
				Runs       []Row  `json:"rows_v2"`
			} `json:"layout"`
		}
		if err := json.Unmarshal(line, &frame); err != nil {
			return Screen{}, errors.New("invalid persistent terminal frame")
		}
		p := frame.Layout
		if frame.Session != c.options.Session || p.Type != "leaf" || p.ID != c.options.Pane || !p.Active {
			return Screen{}, errors.New("terminal Session/pane changed; reconnect explicitly")
		}
		if !validSize(p.Cols, p.Rows) || p.CursorCol < 0 || p.CursorCol > p.Cols || p.CursorRow < 0 || p.CursorRow >= p.Rows || len(p.Runs) > p.Rows {
			return Screen{}, fmt.Errorf("terminal frame exceeds screen bounds (size=%dx%d cursor=%d,%d rows=%d)", p.Cols, p.Rows, p.CursorCol, p.CursorRow, len(p.Runs))
		}
		for _, row := range p.Runs {
			if len(row.Runs) > p.Cols {
				return Screen{}, errors.New("terminal row exceeds screen bounds")
			}
			cells := 0
			for _, run := range row.Runs {
				cells += run.Width
				if !utf8.ValidString(run.Text) || (run.Width < 0 || run.Width > p.Cols || cells > p.Cols || (run.Width == 0 && run.Text != "")) {
					return Screen{}, errors.New("invalid terminal cell")
				}
				for _, r := range run.Text {
					if unicode.IsControl(r) {
						return Screen{}, errors.New("terminal cell contains control characters")
					}
				}
			}
		}
		// The pinned native server can expose col==cols at the right margin
		// while wrap is pending. Display that one-past-column state on the last
		// cell; larger columns and all out-of-range rows remain rejected.
		offset := max(p.Offset, p.ViewOffset)
		return Screen{Width: p.Cols, Height: p.Rows, Cursor: Cursor{min(p.CursorCol, p.Cols-1), p.CursorRow, !p.Hidden && !p.CopyMode && offset == 0, p.Shape}, Rows: p.Runs, AlternateScreen: p.Alternate, WantsMouse: p.Mouse, CopyMode: p.CopyMode, ScrollbackOffset: offset, Status: frame.Status}, nil
	}
	if c.scanner.Err() != nil {
		return Screen{}, errors.New("terminal stream read failed or exceeded frame limit")
	}
	return Screen{}, errors.New("terminal disconnected; Run state is unknown")
}

func validSize(w, h int) bool { return w > 0 && w <= MaxWidth && h > 0 && h <= MaxHeight }
