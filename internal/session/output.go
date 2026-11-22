package session

import "time"

// OutputRecord is a captured terminal screen. It is neither a conversation nor
// a complete stream; Sequence orders retained captures, not lifetime output.
type OutputRecord struct {
	ID         string    `json:"record_id"`
	SessionID  string    `json:"session_id"`
	RunID      string    `json:"run_id"`
	BackendID  string    `json:"backend_id"`
	Sequence   uint64    `json:"sequence"`
	RecordedAt time.Time `json:"recorded_at"`
	Text       string    `json:"text"`
	Truncated  bool      `json:"truncated"`
	Source     string    `json:"source"`
	Complete   bool      `json:"complete"`
}
