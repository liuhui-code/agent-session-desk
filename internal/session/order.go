package session

import "sort"

// SessionOrder is an ordered snapshot of one group's currently discovered rows.
// It is presentation metadata, never a terminal address or Tab order.
type SessionOrder struct {
	Stamp    string
	TargetID string
	Siblings []Session
}

func OrderPeers(items []Session, targetID string) []Session {
	for _, target := range items {
		if target.ID == targetID {
			var peers []Session
			for _, v := range items {
				if GroupOf(v) == GroupOf(target) {
					peers = append(peers, v)
				}
			}
			return peers
		}
	}
	return nil
}

// SameOrderSnapshot ignores changing terminal observations, but checks every
// identity, membership, name and position that determines the reviewed rows.
func SameOrderSnapshot(a, b []Session) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		w := b[i]
		if v.ID != w.ID || v.RunID != w.RunID || v.BackendID != w.BackendID || v.Command != w.Command || GroupOf(v) != GroupOf(w) || v.GroupRevision != w.GroupRevision || v.Title != w.Title || v.TitleRevision != w.TitleRevision || v.Order != w.Order || v.OrderRevision != w.OrderRevision {
			return false
		}
	}
	return true
}

// SortSessions retains the legacy name/ID slots across groups (flat view),
// while ordering each group's own rows by its persisted manual positions.
// Unpositioned/new/moved-in rows follow positioned rows in name/ID order.
func SortSessions(items []Session) {
	lessName := func(a, b Session) bool {
		if a.Title == b.Title {
			return a.ID < b.ID
		}
		return a.Title < b.Title
	}
	sort.SliceStable(items, func(i, j int) bool { return lessName(items[i], items[j]) })
	groups := map[string][]Session{}
	for _, v := range items {
		id := GroupOf(v)
		groups[id] = append(groups[id], v)
	}
	for _, rows := range groups {
		sort.SliceStable(rows, func(i, j int) bool {
			a, b := rows[i], rows[j]
			if (a.OrderRevision > 0) != (b.OrderRevision > 0) {
				return a.OrderRevision > 0
			}
			if a.OrderRevision > 0 && a.Order != b.Order {
				return a.Order < b.Order
			}
			return lessName(a, b)
		})
	}
	next := map[string]int{}
	for i, v := range items {
		id := GroupOf(v)
		items[i] = groups[id][next[id]]
		next[id]++
	}
}
