package session_test

import (
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestGroupNameSanitizesBeforeApplyingTheDisplayLimit(t *testing.T) {
	for _, input := range []string{"项目/甲..审核", strings.Repeat("!", 600) + "项目-甲-审核"} {
		name, err := session.NormalizeGroupName(input)
		if err != nil || name != "项目-甲-审核" {
			t.Fatalf("cleanup %q: %q %v", input, name, err)
		}
	}
	for _, input := range []string{"  ", "name\nnext", "bad\x00", string([]byte{0xff}), strings.Repeat("!", 4097), strings.Repeat("界", 51)} {
		if _, err := session.NormalizeGroupName(input); err == nil {
			t.Fatalf("invalid name accepted: %q", input)
		}
	}
}
