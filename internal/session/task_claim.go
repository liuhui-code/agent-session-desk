package session

import (
	"fmt"
	"time"
)

const DefaultClaimSeconds = 300
const MaxClaimSeconds = 3600

type TaskClaimCredential struct {
	ID        string `json:"claim_id"`
	MessageID string `json:"message_id"`
	Secret    string `json:"-"`
}

func (c TaskClaimCredential) Validate() error {
	if !ValidID(c.ID) || !ValidID(c.MessageID) {
		return fmt.Errorf("full claim and message IDs required")
	}
	return (AgentCredential{SessionID: c.ID, RunID: c.MessageID, Secret: c.Secret}).Validate()
}
func (c TaskClaimCredential) Digest() string { return (AgentCredential{Secret: c.Secret}).Digest() }

// A claim grants cooperative processing ownership, not proof of execution.
type TaskClaim struct {
	NewlyClaimed bool      `json:"newly_claimed"`
	ID           string    `json:"claim_id"`
	MessageID    string    `json:"message_id"`
	Owner        Session   `json:"owner"`
	State        string    `json:"state"`
	Reason       string    `json:"reason"`
	LeaseSeconds int       `json:"lease_seconds"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
