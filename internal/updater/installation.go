package updater

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Record struct {
	Revision      string        `json:"revision"`
	Entry         string        `json:"entrypoint"`
	SHA256        string        `json:"sha256"`
	Compatibility Compatibility `json:"compatibility"`
}
type State struct {
	Protocol int               `json:"protocol"`
	Previous Record            `json:"previous"`
	Owned    map[string]string `json:"owned"`
	Backup   string            `json:"backup,omitempty"`
	Tag      string            `json:"tag"`
}
type installation struct {
	raw       []byte
	doc       map[string]json.RawMessage
	entry     string
	namespace string
	hashes    map[string]string
	state     State
}

func (f *files) installation() (installation, error) {
	var a installation
	b, err := f.read("alpha.json", maxAlphaBytes)
	if err != nil {
		return a, fmt.Errorf("managed alpha.json installation required")
	}
	a.raw = b
	b = []byte(strings.TrimPrefix(string(b), "\xef\xbb\xbf"))
	if _, err := inspect(b); err != nil {
		return a, err
	}
	if json.Unmarshal(b, &a.doc) != nil || json.Unmarshal(a.doc["entrypoint"], &a.entry) != nil || !leafPattern.MatchString(a.entry) || json.Unmarshal(a.doc["files"], &a.hashes) != nil || !digest(a.hashes[a.entry], 32) {
		return a, fmt.Errorf("invalid portable entrypoint manifest")
	}
	if json.Unmarshal(a.doc["namespace"], &a.namespace) != nil || !tagPattern.MatchString(a.namespace) {
		return a, fmt.Errorf("explicit portable namespace required")
	}
	for i := 1; i <= 3; i++ {
		name := slotName(i)
		if strings.EqualFold(a.entry, name) && a.entry != name {
			return a, fmt.Errorf("case alias of managed program slot")
		}
		for key := range a.hashes {
			if strings.EqualFold(key, name) && key != name {
				return a, fmt.Errorf("case alias in program fingerprints")
			}
		}
	}
	if data, ok := a.doc["asd_update"]; ok {
		if decode(data, &a.state) != nil || a.state.Protocol != 1 || !leafPattern.MatchString(a.state.Previous.Entry) || !digest(a.state.Previous.SHA256, 32) || !digest(a.state.Previous.Revision, 20) || len(a.state.Owned) > 3 || (slotIndex(strings.ToLower(a.state.Previous.Entry)) > 0 && a.state.Previous.Entry != strings.ToLower(a.state.Previous.Entry)) {
			return a, fmt.Errorf("invalid previous update receipt")
		}
		if a.state.Backup != "" && a.state.Backup != "backup-1.sqlite3" && a.state.Backup != "backup-2.sqlite3" && a.state.Backup != "backup-3.sqlite3" {
			return a, fmt.Errorf("invalid backup receipt")
		}
		for name, h := range a.state.Owned {
			if slotIndex(name) == 0 || !digest(h, 32) {
				return a, fmt.Errorf("invalid managed update slot")
			}
		}
	}
	h, err := f.hash(a.entry)
	if err != nil || h != a.hashes[a.entry] {
		return a, fmt.Errorf("current program hash changed")
	}
	return a, nil
}
func slotIndex(name string) int {
	for i := 1; i <= 3; i++ {
		if name == slotName(i) {
			return i
		}
	}
	return 0
}
func slotName(i int) string { return fmt.Sprintf("asd-update-%d.exe", i) }
func (a installation) next(target Record, tag, backup string) ([]byte, error) {
	owned := map[string]string{}
	for n, h := range a.state.Owned {
		owned[n] = h
	}
	if slotIndex(target.Entry) > 0 {
		owned[target.Entry] = target.SHA256
	}
	for i := 1; i <= 3; i++ {
		delete(a.hashes, slotName(i))
	}
	previous := Record{Entry: a.entry, SHA256: a.hashes[a.entry], Compatibility: target.Compatibility}
	json.Unmarshal(a.doc["ui_source_commit"], &previous.Revision)
	// The old entry may itself be a managed slot removed above.
	if previous.SHA256 == "" {
		previous.SHA256 = a.state.Owned[a.entry]
	}
	a.hashes[target.Entry] = target.SHA256
	a.hashes[previous.Entry] = previous.SHA256
	a.doc["entrypoint"], _ = json.Marshal(target.Entry)
	a.doc["ui_source_commit"], _ = json.Marshal(target.Revision)
	a.doc["files"], _ = json.Marshal(a.hashes)
	a.doc["asd_update"], _ = json.Marshal(State{1, previous, owned, backup, tag})
	data, err := json.MarshalIndent(a.doc, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte{0xef, 0xbb, 0xbf}, data...), nil
}
