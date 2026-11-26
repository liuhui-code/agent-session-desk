package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type taskClaimDocument struct {
	Version      int    `json:"schema_version"`
	ID           string `json:"claim_id"`
	MessageID    string `json:"message_id"`
	LeaseSeconds int    `json:"lease_seconds"`
	Token        string `json:"token"`
}

func loadTaskClaimFile(path string) (session.TaskClaimCredential, int, error) {
	var d taskClaimDocument
	if err := readPrivateDocument(path, &d); err != nil {
		return session.TaskClaimCredential{}, 0, err
	}
	if d.Version != 1 || d.LeaseSeconds < 1 || d.LeaseSeconds > session.MaxClaimSeconds {
		return session.TaskClaimCredential{}, 0, fmt.Errorf("invalid task claim file")
	}
	k := session.TaskClaimCredential{ID: d.ID, MessageID: d.MessageID, Secret: d.Token}
	return k, d.LeaseSeconds, k.Validate()
}

func ensureTaskClaimFile(path, id, msg string, seconds int) (session.TaskClaimCredential, error) {
	match := func() (session.TaskClaimCredential, error) {
		k, n, err := loadTaskClaimFile(path)
		if err == nil && (k.ID != id || k.MessageID != msg || n != seconds) {
			return session.TaskClaimCredential{}, fmt.Errorf("existing claim file belongs to different intent")
		}
		return k, err
	}
	if _, err := os.Lstat(path); err == nil {
		return match()
	} else if !os.IsNotExist(err) {
		return session.TaskClaimCredential{}, fmt.Errorf("claim destination unavailable")
	}
	s, err := session.NewAgentCredential(id, msg)
	if err != nil {
		return session.TaskClaimCredential{}, err
	}
	k := session.TaskClaimCredential{ID: id, MessageID: msg, Secret: s.Secret}
	data, err := json.Marshal(taskClaimDocument{1, id, msg, seconds, k.Secret})
	if err != nil {
		return k, err
	}
	created, err := publishPrivateDocument(path, data)
	if err != nil {
		return session.TaskClaimCredential{}, err
	}
	if !created {
		return match()
	}
	return k, nil
}
