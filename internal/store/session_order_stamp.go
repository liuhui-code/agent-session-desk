package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"hash"
)

// Stream a group-scoped review stamp, including historical identities. This
// detects metadata insertions between discovery and the write transaction,
// without retaining a second list of history or claiming backend atomicity.
func (s *Store) sessionOrderStamps(ctx context.Context, q rowsReader, groups []session.Group) (map[string]string, error) {
	hashes := map[string]hash.Hash{}
	for _, g := range groups {
		hashes[g.ID] = sha256.New()
	}
	rows, err := q.QueryContext(ctx, `SELECT s.id,coalesce(m.group_id,'default'),coalesce(m.revision,0),coalesce(n.revision,0),coalesce(p.ordinal,0),coalesce(p.revision,0),r.id,r.backend_id,coalesce(r.stopped_ns,0) FROM sessions s JOIN runs r ON r.scope=s.scope AND r.session_id=s.id LEFT JOIN session_groups m ON m.scope=s.scope AND m.session_id=s.id LEFT JOIN session_names n ON n.scope=s.scope AND n.session_id=s.id LEFT JOIN session_positions p ON p.scope=s.scope AND p.session_id=s.id WHERE s.scope=? ORDER BY 2,s.id,r.id`, s.scope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, group, run, backend string
		var member, name, ordinal, position, stopped int64
		if err = rows.Scan(&id, &group, &member, &name, &ordinal, &position, &run, &backend, &stopped); err != nil {
			return nil, err
		}
		if h := hashes[group]; h != nil {
			b, _ := json.Marshal([]any{id, group, member, name, ordinal, position, run, backend, stopped})
			h.Write(b)
			h.Write([]byte{'\n'})
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	result := map[string]string{}
	for id, h := range hashes {
		result[id] = hex.EncodeToString(h.Sum(nil))
	}
	return result, nil
}
