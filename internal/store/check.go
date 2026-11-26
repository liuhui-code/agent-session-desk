package store

import (
	"context"
	"database/sql"
	"fmt"
)

// CheckResult describes database structure only, never live terminal state.
type CheckResult struct {
	SchemaVersion int    `json:"schema_version"`
	Integrity     string `json:"integrity"`
	ForeignKeys   string `json:"foreign_keys"`
}

// Check opens an existing database read-only. It never initializes or migrates.
func Check(ctx context.Context, path string) (CheckResult, error) {
	db, err := openReadOnly(path)
	if err != nil {
		return CheckResult{}, err
	}
	defer db.Close()
	version, err := validateMetadata(ctx, db)
	if err != nil {
		return CheckResult{}, err
	}
	return CheckResult{version, "ok", "ok"}, nil
}

// validateMetadata checks a read transaction without initializing or migrating.
// It returns no stored user content in diagnostics.
func validateMetadata(ctx context.Context, db *sql.DB) (int, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var app, version int
	if err = tx.QueryRowContext(ctx, "PRAGMA application_id").Scan(&app); err != nil {
		return 0, err
	}
	if err = tx.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return 0, err
	}
	if app != 0x41534431 || (version < 1 || version > 7) {
		return 0, fmt.Errorf("unsupported metadata application/schema; source preserved")
	}
	var integrity string
	if err = tx.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrity); err != nil {
		return 0, err
	}
	if integrity != "ok" {
		return 0, fmt.Errorf("metadata integrity check failed; source preserved")
	}
	var journal string
	if err = tx.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journal); err != nil {
		return 0, err
	}
	if journal != "delete" {
		return 0, fmt.Errorf("metadata requires DELETE journal mode")
	}
	queries := []string{`SELECT s.scope,s.id,s.title,s.command,r.id,r.session_id,r.backend_id,r.recorded_ns,r.stopped_ns FROM sessions s JOIN runs r ON s.scope=r.scope AND s.id=r.session_id LIMIT 0`}
	if version >= 2 {
		queries = append(queries, `SELECT scope,run_id,pane_id,source,observed_ns FROM pane_exits LIMIT 0`)
	}
	if version >= 3 {
		queries = append(queries, `SELECT scope,session_id,title,revision FROM session_names LIMIT 0`)
	}
	if version >= 4 {
		queries = append(queries, groupColumns, membershipColumns)
	}
	if version >= 5 {
		queries = append(queries, positionColumns)
	}
	if version >= 6 {
		queries = append(queries, messageColumns+" LIMIT 0")
	}
	if version >= 7 {
		queries = append(queries, replyColumns+" LIMIT 0", "SELECT scope,run_id,token_hash FROM agent_credentials LIMIT 0")
	}
	for _, query := range queries {
		rows, e := tx.QueryContext(ctx, query)
		if e != nil {
			return 0, e
		}
		if e = rows.Close(); e != nil {
			return 0, e
		}
	}
	rows, err := tx.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return 0, err
	}
	violation := rows.Next()
	err = rows.Err()
	rows.Close()
	if err != nil {
		return 0, err
	}
	if violation {
		return 0, fmt.Errorf("metadata foreign-key check failed; source preserved")
	}
	if version >= 4 {
		if err = validateGroupScopes(ctx, tx); err != nil {
			return 0, err
		}
	}
	if version >= 5 {
		if err = validatePositions(ctx, tx); err != nil {
			return 0, err
		}
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return version, nil
}
