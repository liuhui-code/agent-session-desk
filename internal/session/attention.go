package session

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
	"unicode"
	"unicode/utf8"
)

const MaxAttentionItems = 256
const MaxAttentionEventBytes = 16384

type AttentionEvent struct {
	ID      string `json:"attention_id"`
	Version int64  `json:"version"`
	Kind    string `json:"kind"`
	State   string `json:"state"`
	Title   string `json:"title"`
	Detail  string `json:"detail,omitempty"`
}

func (e AttentionEvent) Fingerprint() (string, error) {
	if !ValidID(e.ID) || e.Version < 1 || (e.Kind != "approval" && e.Kind != "question") || (e.State != "pending" && e.State != "resolved" && e.State != "withdrawn") {
		return "", fmt.Errorf("invalid attention identity, version, kind or state")
	}
	if e.Title == "" || len(e.Title) > 256 || !utf8.ValidString(e.Title) {
		return "", fmt.Errorf("attention title must be 1-256 UTF-8 bytes")
	}
	for _, r := range e.Title {
		if unicode.IsControl(r) {
			return "", fmt.Errorf("attention title must be one line without controls")
		}
	}
	if len(e.Detail) > 8192 {
		return "", fmt.Errorf("attention detail exceeds 8 KiB")
	}
	if e.Detail != "" {
		if err := ValidateMessageBody(e.Detail); err != nil {
			return "", err
		}
	}
	data, err := json.Marshal(e)
	if err != nil || len(data) > MaxAttentionEventBytes {
		return "", fmt.Errorf("attention event exceeds 16 KiB")
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

type AttentionItem struct {
	AttentionEvent
	EventID           string     `json:"event_id"`
	Source            Session    `json:"source"`
	ViewedAt          *time.Time `json:"viewed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	TerminalAvailable bool       `json:"terminal_available"`
}
