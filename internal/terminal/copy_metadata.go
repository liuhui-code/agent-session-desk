package terminal

import "errors"

type nativeCopyDecoration struct {
	StartRow   *int     `json:"sel_start_row"`
	StartCol   *int     `json:"sel_start_col"`
	EndRow     *int     `json:"sel_end_row"`
	EndCol     *int     `json:"sel_end_col"`
	Mode       string   `json:"sel_mode"`
	CursorRow  *int     `json:"copy_cursor_row"`
	CursorCol  *int     `json:"copy_cursor_col"`
	Highlights [][4]int `json:"copy_hl"`
}

func (d nativeCopyDecoration) decode(copyMode bool, width, height int) (*Selection, *Cursor, []CopyHighlight, error) {
	if !copyMode {
		return nil, nil, nil, nil
	}
	invalid := errors.New("native copy metadata exceeds screen bounds")
	var selection *Selection
	if d.StartRow != nil || d.StartCol != nil || d.EndRow != nil || d.EndCol != nil {
		if d.StartRow == nil || d.StartCol == nil || d.EndRow == nil || d.EndCol == nil {
			return nil, nil, nil, invalid
		}
		sr, sc, er, ec := *d.StartRow, *d.StartCol, *d.EndRow, *d.EndCol
		if sr < 0 || er < sr || er >= height || sc < 0 || sc >= width || ec < 0 || ec >= width {
			return nil, nil, nil, invalid
		}
		mode := d.Mode
		if mode == "" {
			mode = "char"
		}
		if mode != "char" && mode != "line" && mode != "rect" {
			return nil, nil, nil, invalid
		}
		selection = &Selection{sr, sc, er, ec, mode}
	}
	var cursor *Cursor
	if d.CursorRow != nil || d.CursorCol != nil {
		// During resize the native search prompt can retain its old body
		// caret. Hide that optional caret until it is inside the new viewport;
		// its valid prompt, screen and target identity remain usable.
		if d.CursorRow != nil && d.CursorCol != nil && *d.CursorRow >= 0 && *d.CursorRow < height && *d.CursorCol >= 0 && *d.CursorCol < width {
			cursor = &Cursor{X: *d.CursorCol, Y: *d.CursorRow, Visible: true}
		}
	}
	if len(d.Highlights) > width*height*4 {
		return nil, nil, nil, invalid
	}
	highlights := make([]CopyHighlight, 0, len(d.Highlights))
	for _, h := range d.Highlights {
		if h[0] < 0 || h[0] >= height || h[1] < 0 || h[2] < h[1] || h[2] >= width || h[3] < 0 || h[3] > 3 {
			return nil, nil, nil, invalid
		}
		highlights = append(highlights, CopyHighlight{h[0], h[1], h[2], h[3]})
	}
	return selection, cursor, highlights, nil
}
