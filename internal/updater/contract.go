// Package updater implements explicit, bounded portable program updates. It
// never starts a workbench, stops a Run or restores a metadata database.
package updater

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
)

const MaxProductBytes int64 = 64 << 20
const MaxManifestBytes int64 = 32 << 10
const maxAlphaBytes int64 = 1 << 20
const work = ".asd-update"

var tagPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
var leafPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,95}\.exe$`)

type Compatibility struct {
	Metadata    int               `json:"metadata"`
	UI          int               `json:"ui"`
	Launch      int               `json:"launch"`
	Keybindings int               `json:"keybindings"`
	RecentDirs  int               `json:"recent_dirs"`
	Recovery    int               `json:"recovery"`
	Runtime     map[string]string `json:"runtime"`
}
type Probe struct {
	Protocol      int           `json:"protocol"`
	Platform      string        `json:"platform"`
	Revision      string        `json:"revision"`
	Dirty         bool          `json:"dirty"`
	Compatibility Compatibility `json:"compatibility"`
}
type Contract struct {
	Protocol      int           `json:"protocol"`
	Tag           string        `json:"tag"`
	Platform      string        `json:"platform"`
	Revision      string        `json:"revision"`
	Product       string        `json:"product"`
	SHA256        string        `json:"sha256"`
	Size          int64         `json:"size"`
	Compatibility Compatibility `json:"compatibility"`
}

func digest(s string, n int) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == n && s == fmt.Sprintf("%x", b)
}
func (c Contract) Validate() error {
	if c.Protocol != 1 || !tagPattern.MatchString(c.Tag) || c.Platform != "windows/amd64" || !digest(c.Revision, 20) || c.Product != "asd-windows-amd64.exe" || !digest(c.SHA256, 32) || c.Size < 1 || c.Size > MaxProductBytes {
		return fmt.Errorf("invalid update contract")
	}
	p := c.Compatibility
	if p.Metadata != 11 || p.UI != 7 || p.Launch != 1 || p.Keybindings != 1 || p.RecentDirs != 1 || p.Recovery != 1 || len(p.Runtime) != 3 {
		return fmt.Errorf("unsupported update state formats")
	}
	for _, name := range []string{"psmux.exe", "conpty.dll", "OpenConsole.exe"} {
		if !digest(p.Runtime[name], 32) {
			return fmt.Errorf("incomplete runtime fingerprints")
		}
	}
	return nil
}
func (c Contract) Accept(p Probe) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if p.Protocol != 1 || p.Platform != c.Platform || p.Dirty || p.Revision != c.Revision || !reflect.DeepEqual(p.Compatibility, c.Compatibility) {
		return fmt.Errorf("candidate probe does not match verified contract")
	}
	return nil
}
func decode(b []byte, dst any) error {
	v, err := inspect(b)
	if err != nil || !exactFields(v, reflect.TypeOf(dst).Elem()) {
		return fmt.Errorf("ambiguous update JSON")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(dst); err != nil {
		return fmt.Errorf("invalid update document")
	}
	var tail any
	if d.Decode(&tail) != io.EOF {
		return fmt.Errorf("trailing update document")
	}
	return nil
}
