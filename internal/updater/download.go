package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const repoURL = "https://github.com/liuhui-code/agent-session-desk/releases/download/"
const apiURL = "https://api.github.com/repos/liuhui-code/agent-session-desk/releases/tags/"

type Client struct{ HTTP *http.Client }
type asset struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	State  string `json:"state"`
	Digest string `json:"digest"`
	URL    string `json:"browser_download_url"`
}

func assetHost(u *url.URL) bool {
	if u.Scheme != "https" || u.User != nil || u.Port() != "" || u.Fragment != "" {
		return false
	}
	switch u.Host {
	case "github.com", "release-assets.githubusercontent.com", "objects.githubusercontent.com":
		return true
	}
	return false
}
func (c Client) client(binary bool) *http.Client {
	h := http.Client{Timeout: 2 * time.Minute}
	if c.HTTP != nil {
		h = *c.HTTP
	}
	h.Jar = nil
	if h.Transport == nil {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.DisableCompression = true
		h.Transport = t
	}
	if binary {
		h.CheckRedirect = func(r *http.Request, via []*http.Request) error {
			if len(via) > 3 || !assetHost(r.URL) {
				return fmt.Errorf("asset redirect refused")
			}
			return nil
		}
	} else {
		h.CheckRedirect = func(*http.Request, []*http.Request) error { return fmt.Errorf("metadata redirect refused") }
	}
	return &h
}
func (c Client) get(ctx context.Context, address string, binary bool) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", address, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "agent-session-desk")
	req.Header.Set("Accept", "application/vnd.github+json")
	if binary {
		req.Header.Set("Accept", "application/octet-stream")
	}
	res, err := c.client(binary).Do(req)
	if err != nil {
		return nil, fmt.Errorf("release download unavailable")
	}
	if res.StatusCode != 200 {
		res.Body.Close()
		return nil, fmt.Errorf("release download HTTP %d", res.StatusCode)
	}
	return res.Body, nil
}
func (c Client) assets(ctx context.Context, tag string) (asset, asset, error) {
	var zero asset
	if !tagPattern.MatchString(tag) {
		return zero, zero, fmt.Errorf("explicit published tag required")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	body, err := c.get(ctx, apiURL+tag, false)
	if err != nil {
		return zero, zero, err
	}
	defer body.Close()
	b, err := io.ReadAll(io.LimitReader(body, 256*1024+1))
	if err != nil || len(b) > 256*1024 {
		return zero, zero, fmt.Errorf("release metadata exceeds bound")
	}
	var v struct {
		Tag    string  `json:"tag_name"`
		Draft  bool    `json:"draft"`
		Assets []asset `json:"assets"`
	}
	if json.Unmarshal(b, &v) != nil || v.Draft || v.Tag != tag {
		return zero, zero, fmt.Errorf("published release identity mismatch")
	}
	var manifest, program asset
	for _, a := range v.Assets {
		switch a.Name {
		case "asd-update.json":
			if manifest.Name != "" {
				return zero, zero, fmt.Errorf("duplicate update asset")
			}
			manifest = a
		case "asd-windows-amd64.exe":
			if program.Name != "" {
				return zero, zero, fmt.Errorf("duplicate product asset")
			}
			program = a
		}
	}
	for _, a := range []asset{manifest, program} {
		limit := MaxProductBytes
		if a.Name == "asd-update.json" {
			limit = MaxManifestBytes
		}
		hash, ok := strings.CutPrefix(a.Digest, "sha256:")
		if a.State != "uploaded" || a.Size < 1 || a.Size > limit || !ok || !digest(hash, 32) || a.URL != repoURL+tag+"/"+a.Name {
			return zero, zero, fmt.Errorf("release lacks supported update assets or verified digests")
		}
	}
	return manifest, program, nil
}
func (e Engine) Download(ctx context.Context, tag string, c Client) error {
	return e.locked(ctx, func(f *files) error {
		if err := noIntent(f); err != nil {
			return err
		}
		if err := absent(f, work+"/pending.json"); err != nil {
			return err
		}
		manifest, program, err := c.assets(ctx, tag)
		if err != nil {
			return err
		}
		body, err := c.get(ctx, manifest.URL, true)
		if err != nil {
			return err
		}
		b, readErr := io.ReadAll(io.LimitReader(body, MaxManifestBytes+1))
		body.Close()
		if readErr != nil || int64(len(b)) != manifest.Size || "sha256:"+hashBytes(b) != manifest.Digest {
			return fmt.Errorf("update manifest download hash/size mismatch")
		}
		var contract Contract
		if decode(b, &contract) != nil || contract.Validate() != nil || contract.Tag != tag || contract.SHA256 != "" && "sha256:"+contract.SHA256 != program.Digest || contract.Size != program.Size {
			return fmt.Errorf("update contract differs from release assets")
		}
		product, err := c.get(ctx, program.URL, true)
		if err != nil {
			return err
		}
		defer product.Close()
		return e.stage(ctx, f, contract, product)
	})
}
