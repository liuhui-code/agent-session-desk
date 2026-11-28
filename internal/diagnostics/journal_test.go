package diagnostics_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/diagnostics"
)

func TestReportKeepsBoundedUsefulFactsWithoutRawErrorsOrIdentity(t *testing.T) {
	var journal diagnostics.Journal
	for i := 0; i < 200; i++ {
		journal.Record(diagnostics.Connect, "private-session/private-project", 17*time.Millisecond, errors.New("password=FAKE_SECRET C:\\Users\\private\\project raw terminal content"))
	}
	report := journal.Snapshot(diagnostics.Runtime{}, "private-session/private-project", "available", "disconnected")
	data, err := report.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Events) != diagnostics.MaxEvents || !strings.Contains(string(data), "operation_failed") || !strings.Contains(string(data), "17") {
		t.Fatal("bounded timing/cause missing")
	}
	for _, secret := range []string{"FAKE_SECRET", "private-session", "private-project", "password", "raw terminal"} {
		if strings.Contains(string(data), secret) {
			t.Fatalf("report leaked %q", secret)
		}
	}
	journal.Record(diagnostics.Discovery, "", time.Second, context.DeadlineExceeded)
	data, _ = journal.Snapshot(diagnostics.Runtime{}, "", "unknown", "unknown").Encode()
	if !strings.Contains(string(data), "timeout") {
		t.Fatal("typed deadline cause lost")
	}
}

func TestEncodingRejectsUnallowlistedReportStrings(t *testing.T) {
	report := diagnostics.Report{Target: "SECRET_ID", Runtime: diagnostics.Runtime{Revision: "SECRET", Go: "SECRET", OS: "SECRET", Arch: "SECRET", Backend: "SECRET", BackendVersion: "SECRET"}, Terminal: "SECRET", Client: "SECRET", Events: []diagnostics.Event{{Operation: "SECRET", Target: "SECRET", Cause: "SECRET"}}}
	data, err := report.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "SECRET") {
		t.Fatal("unallowlisted report values leaked")
	}
}

func TestRawPathKeywordsDoNotBecomeFailureEvidence(t *testing.T) {
	err := errors.New(`open C:\identity\permission\closed\project: file unavailable`)
	if diagnostics.Cause(err) != "operation_failed" {
		t.Fatal("path text was treated as terminal failure evidence")
	}
}
