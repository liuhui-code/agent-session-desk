package manager_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestRecoveryReviewReadsSavedLaunchWithoutStartingOrDiscoveringTerminal(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "review")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	original := observed()
	if err = db.Save(ctx, original); err != nil {
		t.Fatal(err)
	}
	launch := session.Launch{Name: original.Title, Executable: original.Command, Args: []string{"literal $SECRET", ""}, Dir: t.TempDir(), Env: []string{"PRIVATE_TEST_VALUE=secret"}}
	if err = db.SaveLaunch(ctx, original, launch); err != nil {
		t.Fatal(err)
	}
	backend := &terminal{err: context.DeadlineExceeded}
	svc := manager.New(backend, db)
	review, err := svc.RecoveryReview(ctx, original.ID)
	if err != nil || review.Source.RunID != original.RunID || review.Launch == nil || review.Launch.Args[0] != launch.Args[0] || review.Profile != nil || review.Attempt != nil {
		t.Fatalf("review missing or invented data: %+v %v", review, err)
	}
	if backend.starts != 0 || backend.stops != 0 {
		t.Fatal("review changed native terminal")
	}
}
