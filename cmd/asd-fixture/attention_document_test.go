package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestAttentionFixturePrintsExistingMarkdownReferenceOnly(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.WriteFile(filepath.Join(root, "review.md"), []byte("# controlled fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	sink := &fixtureSink{stdout: &out}
	if err := printAttentionDocument(sink, "review.md"); err != nil || out.String() != "Review document: \"review.md\"\n" {
		t.Fatal(out.String(), err)
	}
	out.Reset()
	if err := printAttentionDocument(sink, ""); err != nil || out.Len() != 0 {
		t.Fatal("empty reference changed fixture")
	}
	if err := printAttentionDocument(sink, "absent.md"); err == nil || out.Len() != 0 {
		t.Fatal("missing file emitted a reference")
	}
}
