package recovery_test

import (
	"bytes"
	"encoding/json"
	"github.com/liuhui-code/agent-session-desk/internal/recovery"
	"os"
	"strings"
	"testing"
)

func TestDecodeDeclarationPreservesLiteralResume(t *testing.T) {
	data, err := os.ReadFile("../../docs/examples/agent-recovery.json")
	if err != nil {
		t.Fatal(err)
	}
	d, err := recovery.DecodeDeclaration(data)
	if err != nil {
		t.Fatal(err)
	}
	if d.ConversationID != "conversation-example-42" || d.Resume.Args[1] != d.ConversationID || d.Resume.Executable != `C:\Tools\MyAgent\agent.exe` {
		t.Fatalf("changed declaration: %#v", d)
	}
}

func TestDeclarationRejectsAmbiguousOrLossyJSON(t *testing.T) {
	data, err := os.ReadFile("../../docs/examples/agent-recovery.json")
	if err != nil {
		t.Fatal(err)
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, data); err != nil {
		t.Fatal(err)
	}
	valid := compact.String()
	for name, bad := range map[string]string{
		"duplicate":        strings.Replace(valid, `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1),
		"nested duplicate": strings.Replace(valid, `"env":{}`, `"env": {"A":"x","A":"y"}`, 1),
		"wrong case":       strings.Replace(valid, `"agent_id"`, `"AGENT_ID"`, 1),
		"null args":        strings.Replace(valid, `["resume","conversation-example-42"]`, `null`, 1),
		"null env":         strings.Replace(valid, `"env":{}`, `"env": null`, 1),
		"null version":     strings.Replace(valid, `"agent_version":"1.0.0"`, `"agent_version": null`, 1),
		"missing args":     strings.Replace(valid, `"args":["resume","conversation-example-42"],`, ``, 1),
		"boolean version":  strings.Replace(valid, `"schema_version":1`, `"schema_version":true`, 1),
		"LF":               strings.Replace(valid, `conversation-example-42`, `conversation-example-42\n`, 1),
		"surrogate":        strings.Replace(valid, `conversation-example-42`, `\ud800`, 1),
		"invalid UTF8":     strings.Replace(valid, `conversation-example-42`, string([]byte{0xff}), 1),
		"trailing":         valid + "{}",
		"BOM":              "\xef\xbb\xbf" + valid,
		"oversize":         valid + strings.Repeat(" ", recovery.MaxDeclarationBytes),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := recovery.DecodeDeclaration([]byte(bad)); err == nil {
				t.Fatal("ambiguous or lossy JSON accepted")
			}
		})
	}
}

func TestDeclarationVersionUsesExactBoundedNumericValue(t *testing.T) {
	b, e := os.ReadFile("../../docs/examples/agent-recovery.json")
	if e != nil {
		t.Fatal(e)
	}
	for _, number := range []string{"1.0", "1e0", "10e-1", "0.01e2"} {
		if _, e := recovery.DecodeDeclaration([]byte(strings.Replace(string(b), `"schema_version": 1`, `"schema_version": `+number, 1))); e != nil {
			t.Fatalf("version %s: %v", number, e)
		}
	}
	for _, number := range []string{"1.00000000000000001", "1e999999999999999999", "-1", "0"} {
		if _, e := recovery.DecodeDeclaration([]byte(strings.Replace(string(b), `"schema_version": 1`, `"schema_version": `+number, 1))); e == nil {
			t.Fatalf("invalid version %s", number)
		}
	}
}
