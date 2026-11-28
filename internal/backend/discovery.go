package backend

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// Retain only bounded identity hints, never terminal output or cached liveness.
// Each missing name must be independently revalidated on the current poll.
const maxDiscoveryHints = 128

type discoveryHints struct {
	mu      sync.Mutex
	records map[string]liveRecord
}

func (b *Backend) List(ctx context.Context) ([]session.Session, error) {
	b.discovery.mu.Lock()
	defer b.discovery.mu.Unlock()
	out, err := b.run(ctx, maxControlBytes, "list-sessions", "-F", "#{session_name}")
	if err != nil {
		if noServer(err) {
			b.discovery.records = nil
			return []session.Session{}, nil
		}
		return nil, err
	}
	if out.truncated {
		return nil, errors.New("backend session list exceeded M0 output limit")
	}
	records := make(map[string]liveRecord)
	for _, name := range strings.Fields(out.text) {
		if !namePattern.MatchString(name) {
			continue
		}
		r, err := b.readRecord(ctx, name)
		if errors.Is(err, errUnpublishedRecord) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read live session %s: %w", name, err)
		}
		records[name] = r
	}
	if b.kind == "psmux" {
		// A successful global enumeration can temporarily omit a live original.
		// Query exact original targets; never manufacture Available from old data.
		names := make([]string, 0, len(b.discovery.records))
		for name := range b.discovery.records {
			if _, ok := records[name]; !ok {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		for _, name := range names {
			old := b.discovery.records[name]
			current, err := b.readRecord(ctx, name)
			if err != nil || current != old {
				// Do not retain an unreachable hint forever. This poll is uncertain;
				// subsequent discovery can remove the row without claiming an exit.
				delete(b.discovery.records, name)
				return nil, fmt.Errorf("omitted original session %s could not be verified", name)
			}
			records[name] = current
		}
	}
	names := make([]string, 0, len(records))
	for name := range records {
		names = append(names, name)
	}
	sort.Strings(names)
	var sessions []session.Session
	next := make(map[string]liveRecord)
	for _, name := range names {
		r := records[name]
		state, exit, program := b.observePaneProgram(ctx, name, r.PaneID, true)
		sessions = append(sessions, session.Session{ID: r.SessionID, RunID: r.RunID, BackendID: name, Title: r.Title, Command: r.Command, CurrentProgram: program, TerminalState: state, PaneExit: exit})
		if b.kind == "psmux" && len(next) < maxDiscoveryHints {
			next[name] = r
		}
	}
	b.discovery.records = next
	sort.Slice(sessions, func(i, j int) bool {
		if sessions[i].Title == sessions[j].Title {
			return sessions[i].ID < sessions[j].ID
		}
		return sessions[i].Title < sessions[j].Title
	})
	return sessions, nil
}
