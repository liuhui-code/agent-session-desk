package recovery

import "github.com/liuhui-code/agent-session-desk/internal/session"

type Profile struct {
	Revision    int64       `json:"revision"`
	Declaration Declaration `json:"declaration"`
}

// Only the latest attempt per logical Session is retained. Historical Runs
// remain ordinary metadata; pending/unknown never authorize automatic retries.
type Attempt struct {
	ID             string  `json:"attempt_id"`
	SessionID      string  `json:"session_id"`
	RunID          string  `json:"run_id"`
	ConversationID string  `json:"conversation_id"`
	Revision       int64   `json:"profile_revision"`
	State          string  `json:"state"`
	Result         *Result `json:"result,omitempty"`
}

type Reservation struct {
	Attempt       Attempt
	PreviousRunID string
	Target        session.Session
	Launch        session.Launch
	Credential    session.AgentCredential
}
