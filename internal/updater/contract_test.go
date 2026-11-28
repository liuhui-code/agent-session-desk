package updater

import (
	"strings"
	"testing"
)

func TestContractRequiresSameFormatsAndRuntimeBeforeDownloadCanBeUsed(t *testing.T) {
	c := testContract()
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	c.Compatibility.Metadata++
	if c.Validate() == nil {
		t.Fatal("cross-schema update accepted")
	}
}

func TestContractAcceptsOnlyIdenticalStateFormatsAndCleanExactBuild(t *testing.T) {
	c := testContract()
	p := Probe{Protocol: 1, Platform: c.Platform, Revision: c.Revision, Compatibility: c.Compatibility}
	if err := c.Accept(p); err != nil {
		t.Fatal(err)
	}
	p.Compatibility.Metadata++
	if c.Accept(p) == nil {
		t.Fatal("cross-schema update accepted")
	}
	p.Compatibility = c.Compatibility
	p.Dirty = true
	if c.Accept(p) == nil {
		t.Fatal("dirty candidate accepted")
	}
	p.Dirty = false
	p.Revision = strings.Repeat("b", 40)
	if c.Accept(p) == nil {
		t.Fatal("wrong source accepted")
	}
}
func testContract() Contract {
	return Contract{Protocol: 1, Tag: "v0.1.0-test", Platform: "windows/amd64", Revision: strings.Repeat("a", 40), Product: "asd-windows-amd64.exe", SHA256: strings.Repeat("c", 64), Size: 3, Compatibility: Compatibility{Metadata: 11, UI: 7, Launch: 1, Keybindings: 1, RecentDirs: 1, Recovery: 1, Runtime: map[string]string{"psmux.exe": strings.Repeat("d", 64), "conpty.dll": strings.Repeat("e", 64), "OpenConsole.exe": strings.Repeat("f", 64)}}}
}

func TestAmbiguousUpdateJSONIsRejected(t *testing.T) {
	for _, text := range []string{`{"protocol":1,"protocol":1}`, `{"Protocol":1}`, `{"protocol":null}`} {
		var c Contract
		if decode([]byte(text), &c) == nil {
			t.Fatalf("ambiguous contract decoded: %s", text)
		}
	}
}
