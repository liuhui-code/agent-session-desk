package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// AgentCredential is a delegated local capability, not process authentication.
// Secret is persisted only in an explicit credential file, never public JSON.
type AgentCredential struct {
	SessionID string `json:"session_id"`
	RunID     string `json:"run_id"`
	Secret    string `json:"-"`
}

func NewAgentCredential(sid, rid string) (AgentCredential, error) {
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return AgentCredential{}, fmt.Errorf("generate agent credential failed")
	}
	c := AgentCredential{sid, rid, hex.EncodeToString(secret[:])}
	return c, c.Validate()
}

func (c AgentCredential) Validate() error {
	if !ValidID(c.SessionID) || !ValidID(c.RunID) || len(c.Secret) != 64 || c.Secret != strings.ToLower(c.Secret) {
		return fmt.Errorf("invalid agent credential")
	}
	if _, err := hex.DecodeString(c.Secret); err != nil {
		return fmt.Errorf("invalid agent credential")
	}
	return nil
}

func (c AgentCredential) Digest() string {
	sum := sha256.Sum256([]byte(c.Secret))
	return hex.EncodeToString(sum[:])
}

type ReplyRequest struct {
	ID      string `json:"reply_id"`
	ReplyTo string `json:"reply_to"`
	Body    string `json:"body"`
}

func (r ReplyRequest) Validate() error {
	if !ValidID(r.ID) || !ValidID(r.ReplyTo) {
		return fmt.Errorf("full reply ID and reply_to required")
	}
	return ValidateMessageBody(r.Body)
}

// Reply is one explicitly submitted result. It does not certify completion.
type Reply struct {
	ID         string    `json:"reply_id"`
	ReplyTo    string    `json:"reply_to"`
	Source     Session   `json:"source"`
	Body       string    `json:"body"`
	RecordedAt time.Time `json:"recorded_at"`
}
