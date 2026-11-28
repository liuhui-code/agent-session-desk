package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/recovery"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/ui"
)

// Contexts can be held by live Agents. Never rotate, overwrite or reuse them.
const recoveryUIContextSlots = 32

func withRecoveryControls(settings ui.Options, service *manager.Manager, options backend.Options, stateDir string) ui.Options {
	settings.ReviewRecovery = func(ctx context.Context, source session.Session) (recovery.Review, error) {
		r, err := service.RecoveryReview(ctx, source.ID)
		if err != nil {
			return r, err
		}
		if r.Source.ID != source.ID || r.Source.RunID != source.RunID || r.Source.BackendID != source.BackendID {
			return r, fmt.Errorf("latest Run changed; refresh recovery review")
		}
		return r, nil
	}
	settings.RestoreRecovery = func(ctx context.Context, review recovery.Review) (recovery.Outcome, error) {
		var out recovery.Outcome
		if runtime.GOOS != "windows" {
			return out, fmt.Errorf("recovery execution requires native Windows")
		}
		current, err := settings.ReviewRecovery(ctx, review.Source)
		if err != nil {
			return out, err
		}
		if current.Profile == nil || review.Profile == nil || current.Profile.Revision != review.Profile.Revision {
			return out, fmt.Errorf("recovery profile changed; refresh review")
		}
		request, err := prepareUIRecoveryContext(options, stateDir, *current.Profile, current.Source.RunID)
		if err != nil {
			return out, err
		}
		out.Attempted = true
		out.Attempt, err = service.Recover(ctx, request)
		if err != nil {
			return out, err
		}
		rows, err := service.List(ctx)
		if err != nil {
			return out, fmt.Errorf("recovery launch recorded; current terminal observation unavailable")
		}
		for _, row := range rows {
			if row.ID == out.Attempt.SessionID && row.RunID == out.Attempt.RunID {
				out.Session = row
				return out, nil
			}
		}
		return out, fmt.Errorf("recovery launch recorded; target not observed; inspect status without retry")
	}
	return settings
}

func prepareUIRecoveryContext(options backend.Options, stateDir string, profile recovery.Profile, previous string) (manager.RecoveryRequest, error) {
	path, _, err := metadataPath(options, stateDir)
	if err != nil {
		return manager.RecoveryRequest{}, err
	}
	root := filepath.Dir(path)
	for slot := 1; slot <= recoveryUIContextSlots; slot++ {
		dir := filepath.Join(root, fmt.Sprintf("recovery-context-%02d", slot))
		_, err := os.Lstat(dir)
		if err == nil {
			continue
		}
		if !os.IsNotExist(err) {
			return manager.RecoveryRequest{}, fmt.Errorf("private recovery context slot unavailable")
		}
		// createRecoveryContext atomically creates this new directory. A race or
		// partial preparation is refused, never retried using a different identity.
		return createRecoveryContext(options, stateDir, dir, profile, previous)
	}
	return manager.RecoveryRequest{}, fmt.Errorf("private recovery context slots occupied; inspect old attempts before manual cleanup")
}
