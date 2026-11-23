package store_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestGroupMoveCASRejectsABAAndOtherScopeWithoutChangingRun(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "groups.sqlite3")
	a, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	other, err := store.Open(ctx, path, "scope-b")
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if err = a.Save(ctx, sample()); err != nil {
		t.Fatal(err)
	}
	g, err := a.CreateGroup(ctx, "", "A", "")
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := other.CreateGroup(ctx, "", "foreign", "")
	if err != nil {
		t.Fatal(err)
	}
	moved, err := a.MoveToGroup(ctx, sample(), g.ID)
	if err != nil {
		t.Fatal(err)
	}
	back, err := b.MoveToGroup(ctx, moved, session.DefaultGroupID)
	if err != nil || back.GroupRevision != 2 {
		t.Fatalf("move back: %+v %v", back, err)
	}
	if _, err = a.MoveToGroup(ctx, sample(), g.ID); !errors.Is(err, store.ErrGroupConflict) {
		t.Fatalf("ABA accepted: %v", err)
	}
	if _, err = a.MoveToGroup(ctx, back, foreign.ID); err == nil {
		t.Fatal("cross-scope group accepted")
	}
	for _, field := range []string{"id", "run", "backend", "command"} {
		bad := back
		switch field {
		case "id":
			bad.ID = "00000000000000000000000000000000"
		case "run":
			bad.RunID = "00000000000000000000000000000000"
		case "backend":
			bad.BackendID = "another"
		case "command":
			bad.Command = "another"
		}
		if _, err = a.MoveToGroup(ctx, bad, g.ID); err == nil {
			t.Fatalf("wrong %s accepted", field)
		}
	}
	if err = a.Save(ctx, back); err == nil {
		t.Fatal("projected metadata used as raw discovery")
	}
	if err = a.Save(ctx, sample()); err != nil {
		t.Fatal(err)
	}
	view, err := b.Workspace(ctx, []session.Session{sample()})
	if err != nil || view.Sessions[0].GroupID != "" || view.Sessions[0].GroupRevision != 2 {
		t.Fatalf("discovery overwrote membership: %+v %v", view, err)
	}
}

func TestGroupCreateDuplicateIsIdempotentAndKeepsEmptyGroups(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "groups.sqlite3"), "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	g, err := db.CreateGroup(ctx, "", "Project..Review", `C:\work`)
	if err != nil {
		t.Fatal(err)
	}
	again, err := db.CreateGroup(ctx, "", "Project/Review", `C:\work`)
	if err != nil || again != g {
		t.Fatalf("duplicate changed group: %+v %v", again, err)
	}
	if _, err = db.CreateGroup(ctx, "", g.Name, `C:\other`); !errors.Is(err, store.ErrGroupConflict) {
		t.Fatalf("duplicate changed directory: %v", err)
	}
	if _, err = db.CreateGroup(ctx, "missing", "child", ""); err == nil {
		t.Fatal("missing parent accepted")
	}
	groups, err := db.Groups(ctx)
	if err != nil || len(groups) != 2 || groups[1] != g {
		t.Fatalf("empty group lost: %+v %v", groups, err)
	}
}

func TestGroupMigrationConflictsRollbackTheWholeUpgrade(t *testing.T) {
	for _, version := range []string{"v1", "v2", "v3"} {
		for _, table := range []string{"groups", "session_groups"} {
			t.Run(version+"/"+table, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "metadata.sqlite3")
				switch version {
				case "v1":
					createV1(t, path)
				case "v2":
					createV2(t, path)
				case "v3":
					createV3(t, path)
				}
				raw, err := sql.Open("sqlite", path)
				if err != nil {
					t.Fatal(err)
				}
				_, err = raw.Exec("CREATE TABLE " + table + "(unrelated TEXT)")
				raw.Close()
				if err != nil {
					t.Fatal(err)
				}
				before, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				db, err := store.Open(context.Background(), path, "scope-a")
				if err == nil {
					db.Close()
					t.Fatal("conflicting table accepted")
				}
				after, err := os.ReadFile(path)
				if err != nil || sha256.Sum256(before) != sha256.Sum256(after) {
					t.Fatal("failed upgrade changed source")
				}
			})
		}
	}
}

func TestRecoveryRefusesCyclicGroupMetadataWithoutChangingSource(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "groups.sqlite3")
	db, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	g, err := db.CreateGroup(ctx, "", "cycle", "")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = raw.Exec(`UPDATE groups SET parent_id=id WHERE id=?`, g.ID)
	raw.Close()
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.Check(ctx, path); err == nil {
		t.Fatal("cyclic tree declared healthy")
	}
	if err = store.Snapshot(ctx, path, filepath.Join(dir, "backup.sqlite3")); err == nil {
		t.Fatal("cyclic tree published in backup")
	}
	after, err := os.ReadFile(path)
	if err != nil || sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("invalid tree source changed")
	}
}

func TestWorkspaceRejectsAnInvalidStoredDisplayName(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.Save(ctx, sample()); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Rename(ctx, sample(), "valid"); err != nil {
		t.Fatal(err)
	}
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = raw.Exec(`UPDATE session_names SET title=?`, "synthetic\x1b[2J")
	raw.Close()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Workspace(ctx, []session.Session{sample()}); err == nil {
		t.Fatal("invalid stored display name projected")
	}
}

func TestConcurrentGroupMovesHaveOneWinnerAndPreserveTheOtherRevision(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	a, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if err = a.Save(ctx, sample()); err != nil {
		t.Fatal(err)
	}
	g1, err := a.CreateGroup(ctx, "", "A", "")
	if err != nil {
		t.Fatal(err)
	}
	g2, err := b.CreateGroup(ctx, "", "B", "")
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	for i, db := range []*store.Store{a, b} {
		target := []string{g1.ID, g2.ID}[i]
		go func() { <-start; _, e := db.MoveToGroup(ctx, sample(), target); results <- e }()
	}
	close(start)
	wins, conflicts := 0, 0
	for i := 0; i < 2; i++ {
		e := <-results
		if e == nil {
			wins++
		} else if errors.Is(e, store.ErrGroupConflict) {
			conflicts++
		} else {
			t.Fatal(e)
		}
	}
	view, err := a.Workspace(ctx, []session.Session{sample()})
	if err != nil || wins != 1 || conflicts != 1 || view.Sessions[0].GroupRevision != 1 {
		t.Fatalf("lost update: %+v wins=%d conflicts=%d %v", view, wins, conflicts, err)
	}
}

func TestGroupCreationEnforcesDocumentedCountAndDepthBounds(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	parent := ""
	for depth := 0; depth < session.MaxGroupDepth; depth++ {
		g, e := db.CreateGroup(ctx, parent, "nested", "")
		if e != nil {
			t.Fatalf("depth %d: %v", depth, e)
		}
		parent = g.ID
	}
	if _, err = db.CreateGroup(ctx, parent, "too deep", ""); err == nil {
		t.Fatal("depth bound not enforced")
	}
	for i := session.MaxGroupDepth + 1; i < session.MaxGroups; i++ {
		if _, err = db.CreateGroup(ctx, "", fmt.Sprintf("root %d", i), ""); err != nil {
			t.Fatalf("count %d: %v", i, err)
		}
	}
	if _, err = db.CreateGroup(ctx, "", "too many", ""); err == nil {
		t.Fatal("count bound not enforced")
	}
	groups, err := db.Groups(ctx)
	if err != nil || len(groups) != session.MaxGroups {
		t.Fatalf("rejection changed valid groups: %d %v", len(groups), err)
	}
}
