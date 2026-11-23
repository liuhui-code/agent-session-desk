package session

import (
	"slices"
	"testing"
)

func TestManualSessionOrderKeepsOtherGroupFlatSlotsAndNewRowsFollow(t *testing.T) {
	rows := []Session{{ID: "a", Title: "A", Order: 1, OrderRevision: 1}, {ID: "b", Title: "B", GroupID: "other"}, {ID: "c", Title: "C", Order: 0, OrderRevision: 1}, {ID: "d", Title: "0 new"}}
	SortSessions(rows)
	ids := []string{}
	for _, v := range rows {
		ids = append(ids, v.ID)
	}
	if !slices.Equal(ids, []string{"c", "a", "b", "d"}) {
		t.Fatalf("group slots/new rows: %v", ids)
	}
	original := append([]Session(nil), rows...)
	SortSessions(rows)
	if !SameOrderSnapshot(original, rows) {
		t.Fatal("order projection is not idempotent")
	}
	changed := append([]Session(nil), rows...)
	changed[0].TerminalState = TerminalExited
	changed[0].PaneExit = &PaneExit{PaneID: "%2"}
	if !SameOrderSnapshot(rows, changed) {
		t.Fatal("observation caused metadata conflict")
	}
	changed[0].RunID = "replacement"
	if SameOrderSnapshot(rows, changed) {
		t.Fatal("replacement Run accepted")
	}
}
