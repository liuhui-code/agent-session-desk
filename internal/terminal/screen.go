// Package terminal models a bounded terminal screen and an interactive client.
// Terminal character attributes are program output, not application UI styling.
package terminal

// These are defensive transport limits, not accepted resource budgets.
const (
	MaxWidth      = 500
	MaxHeight     = 200
	MaxFrameBytes = 2 * 1024 * 1024
)

type Screen struct {
	Width, Height               int
	Cursor                      Cursor
	Rows                        []Row
	AlternateScreen, WantsMouse bool
}

type Cursor struct {
	X, Y    int
	Visible bool
	Shape   int
}

type Row struct{ Runs []Run }

// Run is a style run; Width is its total cell width, not one glyph width.
type Run struct {
	Text       string `json:"text"`
	Foreground string `json:"fg"`
	Background string `json:"bg"`
	Flags      uint8  `json:"flags"`
	Width      int    `json:"width"`
}
