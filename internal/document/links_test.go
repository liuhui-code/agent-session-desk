package document

import (
	"github.com/charmbracelet/x/ansi"
	"testing"
)

func TestMarkdownLinkAtTerminalCell(t *testing.T) {
	for _, tc := range []struct{ text, hit, want string }{
		{"See docs/plan.md now", "plan", "docs/plan.md"},
		{"中文 `docs/设计.md` 已写好", "设计", "docs/设计.md"},
		{`Open "C:\work space\计划.md" today`, "计划", `C:\work space\计划.md`},
		{"[设计文档](docs/设计.md)", "设计文档", "docs/设计.md"},
		{"[notes](<C:/work space/notes.md>)", "notes", "C:/work space/notes.md"},
		{"see docs/test.md:12:3, done", "test", "docs/test.md:12:3"},
		{"file:///C:/work%20space/notes.md#intro", "notes", "file:///C:/work%20space/notes.md#intro"},
		{"not document.txt or script.md.exe", "script", ""},
		{"中文 👩‍💻 `a.md`", "a.md", "a.md"},
		{`See C:\work space\plan.md`, "plan", ""},
		{"[notes](docs/(old)plan.md)", "plan", ""},
	} {
		t.Run(tc.text, func(t *testing.T) {
			index := 0
			for index < len(tc.text) && tc.text[index:index+min(len(tc.hit), len(tc.text)-index)] != tc.hit {
				index++
			}
			col := ansi.StringWidth(tc.text[:index])
			got, ok := At(tc.text, col)
			if got != tc.want || ok != (tc.want != "") {
				t.Fatalf("cell %d: %q/%v want %q", col, got, ok, tc.want)
			}
		})
	}
	if _, ok := At("file.md next", 8); ok {
		t.Fatal("click outside path opened it")
	}
}
