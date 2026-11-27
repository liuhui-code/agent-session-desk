package manager

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/liuhui-code/agent-session-desk/internal/recovery"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (m *Manager) RegisterRecovery(ctx context.Context, c session.AgentCredential, d recovery.Declaration, revision int64) (recovery.Profile, error) {
	if err := d.Validate(); err != nil {
		return recovery.Profile{}, err
	}
	if err := m.authenticateAgent(ctx, c); err != nil {
		return recovery.Profile{}, err
	}
	return m.metadata.RegisterRecovery(ctx, c, d, revision)
}

func (m *Manager) ReportRecovery(ctx context.Context, c session.AgentCredential, r recovery.Result) (recovery.Attempt, error) {
	if err := r.Validate(); err != nil {
		return recovery.Attempt{}, err
	}
	if err := m.authenticateAgent(ctx, c); err != nil {
		return recovery.Attempt{}, err
	}
	if c.SessionID != r.SessionID || c.RunID != r.RunID {
		return recovery.Attempt{}, fmt.Errorf("recovery result credential mismatch")
	}
	a, err := m.metadata.RecoveryAttempt(ctx, r.SessionID)
	if err != nil {
		return a, err
	}
	if a.ID != r.AttemptID || a.RunID != r.RunID || a.ConversationID != r.ConversationID {
		return a, fmt.Errorf("recovery result attempt mismatch")
	}
	if a.State == "launch_unknown" {
		// Authentication above freshly verified this exact native backend Run.
		// Its existence reconciles a lost start receipt; no process is started.
		v, err := m.metadata.AgentIdentity(ctx, c)
		if err != nil {
			return a, err
		}
		if err = m.metadata.RecoveryStarted(ctx, v); err != nil {
			return a, err
		}
	}
	return m.metadata.RecordRecoveryResult(ctx, c, r)
}

type RecoveryRequest struct {
	SessionID     string
	PreviousRunID string
	Revision      int64
	ConfirmEnded  bool
	AttemptID     string
	Credential    session.AgentCredential
	ContextFile   string
}

type recoveryTerminal interface {
	StartRun(context.Context, session.Launch, string, string) (session.Session, error)
}

// Recover is an explicit single attempt, never a startup/reconnect side effect.
// ConfirmEnded is the user's statement; backend absence alone is not exit proof.
func (m *Manager) Recover(ctx context.Context, r RecoveryRequest) (recovery.Attempt, error) {
	var empty recovery.Attempt
	if !r.ConfirmEnded || !session.ValidID(r.SessionID) || !session.ValidID(r.PreviousRunID) || !session.ValidID(r.AttemptID) || r.Revision < 1 || r.Credential.SessionID != r.SessionID || r.Credential.RunID == r.PreviousRunID || !filepath.IsAbs(r.ContextFile) {
		return empty, fmt.Errorf("explicit ended confirmation, exact Session/previous Run/revision and private context required")
	}
	if err := r.Credential.Validate(); err != nil {
		return empty, err
	}
	starter, ok := m.Terminal.(recoveryTerminal)
	if !ok {
		return empty, fmt.Errorf("backend does not support reserved recovery Runs")
	}
	p, err := m.metadata.RecoveryProfile(ctx, r.SessionID)
	if err != nil {
		return empty, err
	}
	if p.Revision != r.Revision {
		return empty, fmt.Errorf("recovery profile revision conflict")
	}
	previous, err := m.metadata.RecoverySource(ctx, r.SessionID)
	if err != nil || previous.RunID != r.PreviousRunID {
		return empty, fmt.Errorf("latest Run changed; inspect history before recovery")
	}
	rows, err := m.Terminal.List(ctx)
	if err != nil {
		return empty, fmt.Errorf("backend state unknown; no recovery started: %w", err)
	}
	for _, v := range rows {
		if v.ID == r.SessionID {
			return empty, fmt.Errorf("original Session still has a backend terminal; reattach or explicitly stop it before recovery")
		}
	}
	launch := p.Declaration.Launch(previous.Title)
	for k := range p.Declaration.Resume.Env {
		if strings.HasPrefix(strings.ToUpper(k), "ASD_RECOVERY_") {
			return empty, fmt.Errorf("ASD_RECOVERY_ environment keys are reserved")
		}
	}
	launch.Env = append(launch.Env, "ASD_RECOVERY_CONTEXT_FILE="+r.ContextFile)
	launch, err = session.ResolveLaunch(launch)
	if err != nil {
		return empty, err
	}

	target := session.Session{ID: r.SessionID, RunID: r.Credential.RunID, BackendID: "asd_" + r.SessionID + "_" + r.Credential.RunID, Title: launch.Name, Command: launch.Executable, TerminalState: session.TerminalUnknown}
	a := recovery.Attempt{ID: r.AttemptID, SessionID: r.SessionID, RunID: target.RunID, ConversationID: p.Declaration.ConversationID, Revision: p.Revision}
	// No SQL transaction remains open during backend calls. Revision/latest-Run
	// CAS serializes concurrent product recoveries after the liveness observation.
	a, err = m.metadata.ReserveRecovery(ctx, recovery.Reservation{Attempt: a, PreviousRunID: r.PreviousRunID, Target: target, Launch: launch, Credential: r.Credential})
	if err != nil {
		return empty, fmt.Errorf("recovery reservation unconfirmed; inspect recovery-status, do not retry blindly: %w", err)
	}
	created, err := starter.StartRun(ctx, launch, target.ID, target.RunID)
	if err != nil {
		return a, fmt.Errorf("recovery attempt %s run %s launch outcome unknown; inspect status/backend before any new attempt: %w", a.ID, a.RunID, err)
	}
	if created.ID != target.ID || created.RunID != target.RunID || created.BackendID != target.BackendID || created.Title != target.Title || created.Command != target.Command {
		return a, fmt.Errorf("recovery backend identity mismatch; outcome unknown; do not retry")
	}
	if err = m.metadata.Save(ctx, created); err != nil {
		return a, fmt.Errorf("recovery terminal created; metadata unconfirmed; do not retry: %w", err)
	}
	if err = m.metadata.RecoveryStarted(ctx, created); err != nil {
		return a, fmt.Errorf("recovery terminal created; start receipt unconfirmed; do not retry: %w", err)
	}
	a.State = "started"
	current, err := m.metadata.RecoveryAttempt(ctx, target.ID)
	if err != nil {
		return a, fmt.Errorf("recovery attempt %s run %s started; receipt unavailable; do not retry: %w", a.ID, a.RunID, err)
	}
	if current.ID != a.ID || current.RunID != a.RunID {
		return a, fmt.Errorf("recovery attempt %s run %s receipt superseded; no retry", a.ID, a.RunID)
	}
	return current, nil
}
