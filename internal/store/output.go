package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type outputFile struct {
	name   string
	Scope  string               `json:"scope"`
	Record session.OutputRecord `json:"record"`
}

func (s *Store) outputScope() string {
	digest := sha256.Sum256([]byte(s.scope))
	return hex.EncodeToString(digest[:])
}
func (s *Store) outputRoot() string { return filepath.Join(filepath.Dir(s.path), "output-records") }

// RecordOutput persists only a caller-requested screen snapshot. Capture must
// finish before this method: no backend call is made while holding the SQL lock.
func (s *Store) RecordOutput(ctx context.Context, v session.Session, p session.Preview) (session.OutputRecord, error) {
	var zero session.OutputRecord
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}
	defer tx.Rollback()
	var found bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM runs r JOIN sessions s ON s.scope=r.scope AND s.id=r.session_id WHERE r.scope=? AND r.id=? AND r.session_id=? AND r.backend_id=? AND r.launch_command=?)`, s.scope, v.RunID, v.ID, v.BackendID, v.Command).Scan(&found)
	if err != nil {
		return zero, err
	}
	if !found {
		return zero, fmt.Errorf("no exact metadata Run for output capture")
	}
	root := s.outputRoot()
	if _, err = prepareOutputRoot(root, true); err != nil {
		return zero, err
	}
	directory, err := scanOutput(ctx, root)
	files := directory.records
	if err != nil {
		return zero, err
	}
	sequence := uint64(1)
	for _, file := range files {
		if file.Record.Sequence >= sequence {
			sequence = file.Record.Sequence + 1
		}
	}
	for _, file := range files {
		if file.Scope == s.outputScope() && file.Record.RunID == v.RunID && (file.Record.SessionID != v.ID || file.Record.BackendID != v.BackendID) {
			return zero, fmt.Errorf("saved output identity conflict")
		}
	}
	for _, name := range directory.stages {
		if err = os.Remove(filepath.Join(root, name)); err != nil {
			return zero, err
		}
	}
	if sequence == 0 {
		return zero, fmt.Errorf("retained output sequence exhausted")
	}
	id, err := session.NewID()
	if err != nil {
		return zero, err
	}
	text := strings.ToValidUTF8(p.Text, "\uFFFD")
	truncated := p.Truncated || text != p.Text
	if len(text) > outputTextLimit {
		text = text[len(text)-outputTextLimit:]
		for len(text) > 0 && !utf8.RuneStart(text[0]) {
			text = text[1:]
		}
		truncated = true
	}
	record := session.OutputRecord{ID: id, SessionID: v.ID, RunID: v.RunID, BackendID: v.BackendID, Sequence: sequence, RecordedAt: time.Now().UTC(), Text: text, Truncated: truncated, Source: "screen_snapshot", Complete: false}
	data, err := json.Marshal(outputFile{Scope: s.outputScope(), Record: record})
	if err != nil {
		return zero, err
	}
	if len(data) > outputFileLimit {
		return zero, fmt.Errorf("output record exceeds file budget")
	}
	stage := filepath.Join(root, ".stage-"+id+".tmp")
	f, err := os.OpenFile(stage, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return zero, err
	}
	owned, err := f.Stat()
	defer func() {
		if current, e := os.Lstat(stage); e == nil && owned != nil && os.SameFile(owned, current) {
			os.Remove(stage)
		}
	}()
	if err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return zero, err
	}
	if err = ctx.Err(); err != nil {
		return zero, err
	}

	files = append(files, outputFile{name: id + ".json", Scope: s.outputScope(), Record: record})
	sort.Slice(files, func(i, j int) bool { return files[i].Record.Sequence > files[j].Record.Sequence })
	kept := 0
	perRun := map[string]int{}
	for _, file := range files {
		key := file.Scope + ":" + file.Record.RunID
		if perRun[key] >= outputPerRunLimit || kept >= outputTotalLimit {
			if err = os.Remove(filepath.Join(root, file.name)); err != nil {
				return zero, fmt.Errorf("output not published; retention may have pruned older captures: %w", err)
			}
			continue
		}
		kept++
		perRun[key]++
	}

	if err = os.Link(stage, filepath.Join(root, id+".json")); err != nil {
		return zero, fmt.Errorf("publish output (requires hard-link support): %w", err)
	}
	if err = os.Remove(stage); err != nil {
		return record, fmt.Errorf("output %s published; temporary cleanup failed: %w", id, err)
	}
	return record, nil
}

// Output reads preserved captures for an exact historical Run, without treating
// the record as an authorization to attach, restart, or send terminal input.
func (s *Store) Output(ctx context.Context, runID string) ([]session.OutputRecord, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var sessionID, backendID string
	if err = tx.QueryRowContext(ctx, "SELECT session_id,backend_id FROM runs WHERE scope=? AND id=?", s.scope, runID).Scan(&sessionID, &backendID); err != nil {
		return nil, fmt.Errorf("no exact metadata Run for saved output: %w", err)
	}
	ready, err := prepareOutputRoot(s.outputRoot(), false)
	if err != nil {
		return nil, err
	}
	if !ready {
		return []session.OutputRecord{}, nil
	}
	directory, err := scanOutput(ctx, s.outputRoot())
	files := directory.records
	if err != nil {
		return nil, err
	}
	result := []session.OutputRecord{}
	for _, file := range files {
		if file.Scope == s.outputScope() && file.Record.RunID == runID {
			if file.Record.SessionID != sessionID || file.Record.BackendID != backendID {
				return nil, fmt.Errorf("saved output identity conflict")
			}
			result = append(result, file.Record)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Sequence > result[j].Sequence })
	return result, nil
}
