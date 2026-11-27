package release

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"tag_name":"v0.1.0","published_at":"2026-01-01T00:00:00Z"}]`))
	}))
	defer s.Close()
	v, e := check(context.Background(), s.Client(), s.URL)
	if e != nil || !strings.HasSuffix(v.URL, "v0.1.0") {
		t.Fatal(v, e)
	}
}
func TestReject(t *testing.T) {
	for _, b := range []string{`[{"tag_name":"../x","published_at":"2026-01-01T00:00:00Z"}]`, `[]`, strings.Repeat("x", 262145)} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(b)) }))
		_, e := check(context.Background(), s.Client(), s.URL)
		s.Close()
		if e == nil {
			t.Fatal("accepted")
		}
	}
}
