package session

import "time"

// HistoryRecord is a bounded metadata read model, independent of SQL and UI.
// Its nested terminal state is unknown; old evidence is not current liveness.
type HistoryRecord struct {
	Session         Session    `json:"session"`
	RecordedAt      time.Time  `json:"recorded_at"`
	StopConfirmedAt *time.Time `json:"stop_confirmed_at,omitempty"`
	PaneExit        *PaneExit  `json:"pane_exit,omitempty"`
}
