package updater

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestCorruptDownloadNeverPublishesPendingUpdate(t *testing.T) {
	e, c, _ := setup(t)
	contract, _ := json.Marshal(c)
	asset := func(name string, b []byte) map[string]any {
		return map[string]any{"name": name, "size": len(b), "state": "uploaded", "digest": "sha256:" + sum(b), "browser_download_url": "https://github.com/liuhui-code/agent-session-desk/releases/download/" + c.Tag + "/" + name}
	}
	metadata, _ := json.Marshal(map[string]any{"tag_name": c.Tag, "assets": []any{asset("asd-update.json", contract), asset(c.Product, []byte("new"))}})
	client := Client{HTTP: &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Authorization") != "" {
			t.Fatal("credential attached")
		}
		body := string(metadata)
		if strings.HasSuffix(r.URL.Path, "asd-update.json") {
			body = string(contract)
		} else if strings.HasSuffix(r.URL.Path, ".exe") {
			body = "bad"
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: r}, nil
	})}}
	if err := e.Download(context.Background(), c.Tag, client); err == nil {
		t.Fatal("corrupt asset accepted")
	}
	s, err := e.Status(context.Background())
	if err != nil || s.Pending || s.Current != "old.exe" {
		t.Fatal(s, err)
	}
}

func TestVerifiedDownloadStagesWithoutExecutingAProgram(t *testing.T) {
	e, c, _ := setup(t)
	contract, _ := json.Marshal(c)
	asset := func(name string, b []byte) map[string]any {
		return map[string]any{"name": name, "size": len(b), "state": "uploaded", "digest": "sha256:" + sum(b), "browser_download_url": "https://github.com/liuhui-code/agent-session-desk/releases/download/" + c.Tag + "/" + name}
	}
	metadata, _ := json.Marshal(map[string]any{"tag_name": c.Tag, "assets": []any{asset("asd-update.json", contract), asset(c.Product, []byte("new"))}})
	client := Client{HTTP: &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		body := string(metadata)
		if strings.HasSuffix(r.URL.Path, "asd-update.json") {
			body = string(contract)
		} else if strings.HasSuffix(r.URL.Path, ".exe") {
			body = "new"
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: r}, nil
	})}}
	e.Probe = func(context.Context, string) (Probe, error) {
		t.Fatal("download executed program")
		return Probe{}, nil
	}
	if err := e.Download(context.Background(), c.Tag, client); err != nil {
		t.Fatal(err)
	}
	s, err := e.Status(context.Background())
	if err != nil || !s.Pending || s.Current != "old.exe" {
		t.Fatal(s, err)
	}
}
func TestAssetRedirectRefusesHTTPAndForeignHosts(t *testing.T) {
	for _, target := range []string{"http://release-assets.githubusercontent.com/product", "https://example.com/product", "https://github.com.evil.example/product"} {
		t.Run(target, func(t *testing.T) {
			calls := 0
			client := Client{HTTP: &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				return &http.Response{StatusCode: 302, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{"Location": {target}}, Request: r}, nil
			})}}
			if _, err := client.get(context.Background(), "https://github.com/liuhui-code/agent-session-desk/releases/download/test/exe", true); err == nil || calls != 1 {
				t.Fatal("unsafe redirect followed", calls, err)
			}
		})
	}
}
