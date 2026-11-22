package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

const exitSchema = `CREATE TABLE pane_exits(scope TEXT NOT NULL,run_id TEXT NOT NULL,pane_id TEXT NOT NULL,source TEXT NOT NULL,observed_ns INTEGER NOT NULL,PRIMARY KEY(scope,run_id),FOREIGN KEY(scope,run_id) REFERENCES runs(scope,id));`

func (s *Store) saveExit(ctx context.Context, tx *sql.Tx, v session.Session) error {
	if v.PaneExit == nil {
		return nil
	}
	e := v.PaneExit
	_, err := tx.ExecContext(ctx, `INSERT INTO pane_exits VALUES(?,?,?,?,?) ON CONFLICT DO NOTHING`, s.scope, v.RunID, e.PaneID, e.Source, e.ObservedAt.UnixNano())
	if err != nil {
		return err
	}
	_, err = s.matchesExit(ctx, tx, v)
	return err
}

func (s *Store) matchesExit(ctx context.Context, q rowReader, v session.Session) (bool, error) {
	if v.PaneExit == nil {
		return true, nil
	}
	var pane, source string
	err := q.QueryRowContext(ctx, `SELECT pane_id,source FROM pane_exits WHERE scope=? AND run_id=?`, s.scope, v.RunID).Scan(&pane, &source)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if pane != v.PaneExit.PaneID || source != v.PaneExit.Source {
		return true, fmt.Errorf("pane exit identity conflict; existing observation preserved")
	}
	return true, nil
}

var exitPanePattern = regexp.MustCompile(`^%[0-9]+$`)

func validateExit(v session.Session) error {
	e := v.PaneExit
	if e == nil {
		return nil
	}
	if v.TerminalState != session.TerminalExited || !exitPanePattern.MatchString(e.PaneID) || len(e.PaneID) > 64 || (e.Source != "psmux.pane_dead" && e.Source != "tmux.pane_dead") || e.ObservedAt.IsZero() || e.ExitCode != nil {
		return fmt.Errorf("unverified pane exit observation; no agent exit code is established")
	}
	return nil
}
