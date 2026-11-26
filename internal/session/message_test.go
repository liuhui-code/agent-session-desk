package session

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMessageContractBodyBoundariesAndLiteralText(t *testing.T) {
	for _, n := range []int{0, 1, MaxMessageBytes - 1, MaxMessageBytes, MaxMessageBytes + 1} {
		err := ValidateMessageBody(strings.Repeat("x", n))
		if (err == nil) != (n > 0 && n <= MaxMessageBytes) {
			t.Errorf("body byte boundary %d: %v", n, err)
		}
	}
	for _, body := range []string{"中🙂 e\u0301 ' \" \\ $();\r\n\t", strings.Repeat("中", MaxMessageBytes/3)} {
		if err := ValidateMessageBody(body); err != nil {
			t.Fatal(err)
		}
	}
	for _, body := range []string{"x\x00", "x\x1b", "x\b", "x\x7f", string([]byte{0xff}), strings.Repeat("中", MaxMessageBytes/3+1)} {
		if ValidateMessageBody(body) == nil {
			t.Fatal("unsupported control, invalid UTF-8 or oversized body accepted")
		}
	}
}

func TestMessageContractIdempotencyAndJSONReceipt(t *testing.T) {
	r := MessageRequest{RequestID: strings.Repeat("a", 32), TargetSessionID: strings.Repeat("b", 32), TargetRunID: strings.Repeat("c", 32), Body: "literal"}
	fp, err := r.Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*MessageRequest){
		func(v *MessageRequest) { v.Body += " " },
		func(v *MessageRequest) { v.TargetSessionID = strings.Repeat("d", 32) },
		func(v *MessageRequest) { v.TargetRunID = strings.Repeat("d", 32) },
	} {
		other := r
		change(&other)
		got, err := other.Fingerprint()
		if err != nil || got == fp {
			t.Fatal("distinct intent did not change idempotency fingerprint")
		}
	}
	data, err := json.Marshal(Message{ID: r.RequestID, Target: Session{ID: r.TargetSessionID, RunID: r.TargetRunID}, Body: r.Body, State: "input_written", Reason: "agent_acceptance_unknown", Fingerprint: fp})
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]json.RawMessage
	if err = json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"message_id", "target", "body", "state", "reason", "created_at", "updated_at"} {
		if _, ok := doc[key]; !ok {
			t.Fatalf("public JSON receipt field missing: %s", key)
		}
	}
	for _, key := range []string{"Fingerprint", "fingerprint", "attempt", "accepted", "completed"} {
		if _, ok := doc[key]; ok {
			t.Fatalf("private or unproved state in JSON receipt: %s", key)
		}
	}
}
