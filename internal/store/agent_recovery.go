package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/liuhui-code/agent-session-desk/internal/recovery"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

const MaxRecoveryProfiles = 128
const recoverySchema = `
 ALTER TABLE runs ADD COLUMN launch_title TEXT NOT NULL DEFAULT '';
 ALTER TABLE runs ADD COLUMN launch_command TEXT NOT NULL DEFAULT '';
 UPDATE runs SET launch_title=(SELECT title FROM sessions s WHERE s.scope=runs.scope AND s.id=runs.session_id),
 launch_command=(SELECT command FROM sessions s WHERE s.scope=runs.scope AND s.id=runs.session_id);
 CREATE TABLE run_launches(scope TEXT NOT NULL,session_id TEXT NOT NULL,run_id TEXT NOT NULL,document TEXT NOT NULL CHECK(length(CAST(document AS BLOB))<=65536),PRIMARY KEY(scope,session_id),FOREIGN KEY(scope,run_id) REFERENCES runs(scope,id));
 CREATE TABLE recovery_profiles(scope TEXT NOT NULL,session_id TEXT NOT NULL,source_run_id TEXT NOT NULL,revision INTEGER NOT NULL CHECK(revision>0),document TEXT NOT NULL CHECK(length(CAST(document AS BLOB))<=16384),PRIMARY KEY(scope,session_id),FOREIGN KEY(scope,session_id) REFERENCES sessions(scope,id),FOREIGN KEY(scope,source_run_id) REFERENCES runs(scope,id));
 CREATE TABLE recovery_attempts(scope TEXT NOT NULL,session_id TEXT NOT NULL,id TEXT NOT NULL,run_id TEXT NOT NULL,conversation_id TEXT NOT NULL,revision INTEGER NOT NULL,state TEXT NOT NULL,result TEXT,PRIMARY KEY(scope,session_id),UNIQUE(scope,id),FOREIGN KEY(scope,run_id) REFERENCES runs(scope,id));`

var recoveryQueries = []string{
	"SELECT scope,id,launch_title,launch_command FROM runs LIMIT 0",
	"SELECT scope,session_id,run_id,document FROM run_launches LIMIT 0",
	"SELECT scope,session_id,source_run_id,revision,document FROM recovery_profiles LIMIT 0",
	"SELECT scope,session_id,id,run_id,conversation_id,revision,state,result FROM recovery_attempts LIMIT 0",
}

func (s *Store) RecoveryProfile(ctx context.Context, id string) (recovery.Profile, error) {
	if !session.ValidID(id) {
		return recovery.Profile{}, fmt.Errorf("full Session ID required")
	}
	return s.recoveryProfile(ctx, s.db, id)
}
func (s *Store) recoveryProfile(ctx context.Context, q rowReader, id string) (recovery.Profile, error) {
	var p recovery.Profile
	var b string
	if err := q.QueryRowContext(ctx, `SELECT revision,document FROM recovery_profiles WHERE scope=? AND session_id=?`, s.scope, id).Scan(&p.Revision, &b); err != nil {
		return p, fmt.Errorf("recovery profile unavailable")
	}
	d, err := recovery.DecodeDeclaration([]byte(b))
	p.Declaration = d
	return p, err
}

func (s *Store) RegisterRecovery(ctx context.Context, c session.AgentCredential, d recovery.Declaration, expected int64) (recovery.Profile, error) {
	var empty recovery.Profile
	if err := d.Validate(); err != nil {
		return empty, err
	}
	if d.SessionID != c.SessionID || d.RunID != c.RunID || expected < 0 {
		return empty, fmt.Errorf("recovery declaration binding/revision mismatch")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return empty, err
	}
	defer tx.Rollback()
	if _, err = s.agentIdentity(ctx, tx, c); err != nil {
		return empty, err
	}
	latest, err := s.latestRecoveryRun(ctx, tx, d.SessionID)
	if err != nil || latest.RunID != c.RunID {
		return empty, fmt.Errorf("only latest Run may declare recovery")
	}
	b, _ := json.Marshal(d)
	var old string
	var revision int64
	err = tx.QueryRowContext(ctx, `SELECT revision,document FROM recovery_profiles WHERE scope=? AND session_id=?`, s.scope, d.SessionID).Scan(&revision, &old)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return empty, err
	}
	if old == string(b) {
		return recovery.Profile{Revision: revision, Declaration: d}, nil
	}
	if revision != expected {
		return empty, fmt.Errorf("recovery profile revision conflict; read current profile")
	}
	if revision == 0 {
		var count int
		if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM recovery_profiles`).Scan(&count); err != nil {
			return empty, err
		}
		if count >= MaxRecoveryProfiles {
			return empty, fmt.Errorf("recovery profile capacity reached")
		}
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO recovery_profiles VALUES(?,?,?,?,?) ON CONFLICT(scope,session_id) DO UPDATE SET source_run_id=excluded.source_run_id,revision=excluded.revision,document=excluded.document`, s.scope, d.SessionID, d.RunID, revision+1, string(b)); err != nil {
		return empty, err
	}
	return recovery.Profile{Revision: revision + 1, Declaration: d}, tx.Commit()
}

func (s *Store) latestRecoveryRun(ctx context.Context, q rowReader, id string) (session.Session, error) {
	var v session.Session
	err := q.QueryRowContext(ctx, `SELECT r.session_id,r.id,r.backend_id,coalesce(n.title,s.title),r.launch_command FROM runs r JOIN sessions s ON s.scope=r.scope AND s.id=r.session_id LEFT JOIN session_names n ON n.scope=s.scope AND n.session_id=s.id WHERE r.scope=? AND r.session_id=? ORDER BY r.recorded_ns DESC,r.id DESC LIMIT 1`, s.scope, id).Scan(&v.ID, &v.RunID, &v.BackendID, &v.Title, &v.Command)
	return v, err
}

// RecoverySource is a metadata-only latest-Run observation, not liveness proof.
func (s *Store) RecoverySource(ctx context.Context, id string) (session.Session, error) {
	if !session.ValidID(id) {
		return session.Session{}, fmt.Errorf("full Session ID required")
	}
	return s.latestRecoveryRun(ctx, s.db, id)
}
