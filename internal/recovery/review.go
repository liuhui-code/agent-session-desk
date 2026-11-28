package recovery

import "github.com/liuhui-code/agent-session-desk/internal/session"

// Review is one metadata snapshot, never evidence that a native process ended.
// Launch and declaration values are private review data, not diagnostics.
type Review struct {
	Source  session.Session
	Profile *Profile
	Attempt *Attempt
	Launch  *session.Launch
}

type Outcome struct {
	Attempted bool // Recover was entered; an error must not authorize a blind retry.
	Attempt   Attempt
	Session   session.Session
}
