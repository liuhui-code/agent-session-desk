package release

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

const Version = "0.1.0-alpha.2"

func Build() string {
	v := Version
	if b, ok := debug.ReadBuildInfo(); ok {
		for _, s := range b.Settings {
			if s.Key == "vcs.revision" && len(s.Value) >= 7 {
				v += " " + s.Value[:7]
			}
			if s.Key == "vcs.modified" && s.Value == "true" {
				v += " dirty"
			}
		}
	}
	return v
}

type Info struct {
	Tag     string `json:"tag"`
	URL     string `json:"url"`
	Preview bool   `json:"preview"`
}

func Check(ctx context.Context) (Info, error) {
	return check(ctx, &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return fmt.Errorf("redirect refused") }}, "https://api.github.com/repos/liuhui-code/agent-session-desk/releases?per_page=5")
}
func check(ctx context.Context, c *http.Client, endpoint string) (Info, error) {
	r, e := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if e != nil {
		return Info{}, e
	}
	r.Header.Set("Accept", "application/vnd.github+json")
	r.Header.Set("User-Agent", "agent-session-desk")
	s, e := c.Do(r)
	if e != nil {
		return Info{}, fmt.Errorf("release check unavailable")
	}
	defer s.Body.Close()
	if s.StatusCode != 200 {
		return Info{}, fmt.Errorf("release check HTTP %d", s.StatusCode)
	}
	b, e := io.ReadAll(io.LimitReader(s.Body, 262145))
	if e != nil || len(b) > 262144 {
		return Info{}, fmt.Errorf("release metadata exceeds bound")
	}
	var rows []struct {
		Tag       string    `json:"tag_name"`
		Draft     bool      `json:"draft"`
		Preview   bool      `json:"prerelease"`
		Published time.Time `json:"published_at"`
	}
	if json.Unmarshal(b, &rows) != nil {
		return Info{}, fmt.Errorf("invalid release metadata")
	}
	var best time.Time
	var out Info
	for _, v := range rows {
		if v.Draft || !v.Published.After(best) || !validTag(v.Tag) {
			continue
		}
		best = v.Published
		out = Info{v.Tag, "https://github.com/liuhui-code/agent-session-desk/releases/tag/" + v.Tag, v.Preview}
	}
	if out.Tag == "" {
		return out, fmt.Errorf("no published release")
	}
	return out, nil
}
func validTag(s string) bool {
	if s == "" || len(s) > 64 {
		return false
	}
	return strings.IndexFunc(s, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '-' || r == '_')
	}) < 0
}
