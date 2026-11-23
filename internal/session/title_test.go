package session_test

import (
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestRenameNamePolicyPreservesLiteralUnicodeAndRejectsTerminalControls(t *testing.T) {
	for _, value := range []string{"", " \u3000", "a\nb", "a\rb", "a\tb", "a\x1bb", "a\x00b", "a\x7fb", "a\u0085b", "a\u2028b", "a\u2029b", "\xff", strings.Repeat("a", 129), strings.Repeat("a", 4097)} {
		if _, err := session.NormalizeRenameTitle(value); err == nil {
			t.Errorf("accepted invalid name %q", value)
		}
	}
	for _, value := range []string{"中文 ; 'literal'", "a  b", "é", "e\u0301", "👩‍💻", strings.Repeat("🙂", 128)} {
		got, err := session.NormalizeRenameTitle(" \u3000" + value + " ")
		if err != nil || got != value {
			t.Errorf("literal changed: %q => %q %v", value, got, err)
		}
	}
}
