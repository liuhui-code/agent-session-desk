package session

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const MaxMessageBytes = 64 * 1024

type MessageRequest struct {
	RequestID       string `json:"request_id"`
	TargetSessionID string `json:"target_session_id"`
	TargetRunID     string `json:"target_run_id"`
	Body            string `json:"body,omitempty"`
	SourceRunID     string `json:"source_run_id,omitempty"`
	SourceRecordID  string `json:"source_record_id,omitempty"`
	StartLine       int    `json:"start_line,omitempty"`
	EndLine         int    `json:"end_line,omitempty"`
}

type OutputReference struct {
	SessionID  string    `json:"session_id"`
	RunID      string    `json:"run_id"`
	RecordID   string    `json:"record_id"`
	RecordedAt time.Time `json:"recorded_at"`
	StartLine  int       `json:"start_line"`
	EndLine    int       `json:"end_line"`
	Kind       string    `json:"kind"`
	Truncated  bool      `json:"truncated"`
	Complete   bool      `json:"complete"`
}

type Message struct {
	ID          string           `json:"message_id"`
	Target      Session          `json:"target"`
	Source      *OutputReference `json:"source,omitempty"`
	Body        string           `json:"body"`
	State       string           `json:"state"`
	Reason      string           `json:"reason,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Fingerprint string           `json:"-"`
}

// DeliveryAttempt is a single durable claim, never an agent acknowledgement.
type DeliveryAttempt struct {
	Message Message
	Token   string
}

func ValidID(id string) bool {
	if len(id) != 32 || id != strings.ToLower(id) {
		return false
	}
	_, err := hex.DecodeString(id)
	return err == nil
}

func ValidateMessageBody(body string) error {
	if len(body) == 0 || len(body) > MaxMessageBytes || !utf8.ValidString(body) {
		return fmt.Errorf("message must be nonempty UTF-8 within 64 KiB")
	}
	for _, r := range body {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return fmt.Errorf("message contains unsupported control characters")
		}
	}
	return nil
}

func (r MessageRequest) Fingerprint() (string, error) {
	if !ValidID(r.RequestID) || !ValidID(r.TargetSessionID) || !ValidID(r.TargetRunID) {
		return "", fmt.Errorf("full request, target Session and Run IDs required")
	}
	if r.SourceRecordID != "" {
		if r.Body != "" || !ValidID(r.SourceRunID) || !ValidID(r.SourceRecordID) || r.StartLine < 0 || r.EndLine < 0 || (r.EndLine > 0 && r.StartLine > r.EndLine) {
			return "", fmt.Errorf("invalid snapshot source or line range")
		}
	} else if r.SourceRunID != "" || r.StartLine != 0 || r.EndLine != 0 {
		return "", fmt.Errorf("snapshot range requires a source record")
	} else if err := ValidateMessageBody(r.Body); err != nil {
		return "", err
	}
	data, _ := json.Marshal(r)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
