package updater

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestEntrypointFollowsApplyAndRollback(t *testing.T) {
	e, c, product := setup(t)
	ctx := context.Background()
	check := func(want string) {
		t.Helper()
		got, err := Entrypoint(ctx, e.Root)
		if err != nil || got != filepath.Join(e.Root, want) {
			t.Fatal(got, err, want)
		}
	}
	check("old.exe")
	if err := e.Stage(ctx, c, bytes.NewReader(product)); err != nil {
		t.Fatal(err)
	}
	r, err := e.Apply(ctx)
	if err != nil {
		t.Fatal(err)
	}
	check(r.Current)
	if _, err = e.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	check("old.exe")
}

func TestEntrypointRefusesChangedProgramAndUnresolvedTransaction(t *testing.T) {
	for _, fault := range []string{"program", "intent"} {
		t.Run(fault, func(t *testing.T) {
			e, _, _ := setup(t)
			if fault == "program" {
				os.WriteFile(filepath.Join(e.Root, "old.exe"), []byte("changed"), 0600)
			} else {
				os.Mkdir(filepath.Join(e.Root, work), 0700)
				os.WriteFile(filepath.Join(e.Root, work, "intent.json"), []byte("{}"), 0600)
			}
			if _, err := Entrypoint(context.Background(), e.Root); err == nil {
				t.Fatal("unsafe launch target accepted")
			}
		})
	}
}
