package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

type Engine struct {
	Root    string
	Offline func(context.Context, func() error) error
	Probe   func(context.Context, string) (Probe, error)
	Check   func(context.Context, string) (int, error)
	Backup  func(context.Context, string, string) error
}
type plan struct {
	Protocol int      `json:"protocol"`
	Contract Contract `json:"contract"`
	Base     string   `json:"base"`
	Target   string   `json:"target"`
	Replace  string   `json:"replace,omitempty"`
}
type intent struct {
	Protocol int    `json:"protocol"`
	Before   string `json:"before"`
	After    string `json:"after"`
	Target   string `json:"target,omitempty"`
	SHA256   string `json:"sha256,omitempty"`
}
type Receipt struct {
	Action   string `json:"action"`
	Current  string `json:"current"`
	Previous string `json:"previous,omitempty"`
	Backup   string `json:"metadata_backup,omitempty"`
	Note     string `json:"note"`
}

func absent(f *files, name string) error {
	if _, err := f.root.Lstat(name); os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("occupied update file %s; inspect status and repair/discard", name)
}
func noIntent(f *files) error { return absent(f, work+"/intent.json") }
func (e Engine) Stage(ctx context.Context, c Contract, product io.Reader) error {
	return e.locked(ctx, func(f *files) error { return e.stage(ctx, f, c, product) })
}
func (e Engine) stage(ctx context.Context, f *files, c Contract, product io.Reader) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if err := noIntent(f); err != nil {
		return err
	}
	for _, name := range []string{"pending.json", "candidate.exe", "product.tmp", "pending.json.tmp"} {
		if err := absent(f, work+"/"+name); err != nil {
			return err
		}
	}
	a, err := f.installation()
	if err != nil {
		return err
	}
	if err = f.runtime(c.Compatibility); err != nil {
		return err
	}
	p := plan{Protocol: 1, Contract: c, Base: hashBytes(a.raw)}
	for i := 1; i <= 3; i++ {
		name := slotName(i)
		if name == a.entry || name == a.state.Previous.Entry {
			continue
		}
		if _, referenced := a.hashes[name]; referenced {
			continue
		}
		h, err := f.hash(name)
		if os.IsNotExist(err) {
			p.Target = name
			break
		}
		if err != nil {
			return err
		}
		if h != a.state.Owned[name] {
			return fmt.Errorf("unowned inactive program slot; no overwrite")
		}
		p.Target = name
		p.Replace = h
		break
	}
	if p.Target == "" {
		return fmt.Errorf("no inactive program slot")
	}
	if err = f.fresh(work+"/product.tmp", product, c.Size); err != nil {
		return err
	}
	h, err := f.hash(work + "/product.tmp")
	st, _ := f.root.Stat(work + "/product.tmp")
	if err != nil || st == nil || st.Size() != c.Size || h != c.SHA256 || ctx.Err() != nil {
		f.root.Remove(work + "/product.tmp")
		return fmt.Errorf("download incomplete, cancelled or SHA256 mismatch")
	}
	if err = f.root.Rename(work+"/product.tmp", work+"/candidate.exe"); err != nil {
		return err
	}
	b, _ := json.Marshal(p)
	return f.publish(work+"/pending.json", b)
}
func (f *files) runtime(c Compatibility) error {
	st, err := f.root.Lstat("terminal-runtime")
	if err != nil || !safeInfo(st, true) {
		return fmt.Errorf("runtime directory unavailable or linked")
	}
	for n, want := range c.Runtime {
		h, err := f.hash("terminal-runtime/" + n)
		if err != nil || h != want {
			return fmt.Errorf("runtime fingerprint mismatch; no backend replacement")
		}
	}
	return nil
}
func (f *files) pending() (plan, error) {
	var p plan
	b, err := f.read(work+"/pending.json", MaxManifestBytes*2)
	if err != nil {
		return p, err
	}
	if decode(b, &p) != nil || p.Protocol != 1 || p.Contract.Validate() != nil || !digest(p.Base, 32) || slotIndex(p.Target) == 0 || (p.Replace != "" && !digest(p.Replace, 32)) {
		return p, fmt.Errorf("invalid pending update")
	}
	return p, nil
}
func (e Engine) offline(ctx context.Context, f func() error) error {
	if e.Offline == nil {
		return fmt.Errorf("offline update gate unavailable")
	}
	return e.Offline(ctx, f)
}
func (e Engine) Apply(ctx context.Context) (Receipt, error) {
	var r Receipt
	err := e.locked(ctx, func(f *files) error {
		return e.offline(ctx, func() error {
			if err := noIntent(f); err != nil {
				return err
			}
			if err := absent(f, "alpha.update.tmp"); err != nil {
				return err
			}
			p, err := f.pending()
			if err != nil {
				return err
			}
			a, err := f.installation()
			if err != nil {
				return err
			}
			if hashBytes(a.raw) != p.Base || p.Target == a.entry || p.Target == a.state.Previous.Entry {
				return fmt.Errorf("installation changed; discard and download again")
			}
			if e.Probe == nil || e.Backup == nil || e.Check == nil {
				return fmt.Errorf("update verification unavailable")
			}
			if err = f.runtime(p.Contract.Compatibility); err != nil {
				return err
			}
			h, err := f.hash(work + "/candidate.exe")
			if err != nil || h != p.Contract.SHA256 {
				return fmt.Errorf("staged program hash changed")
			}
			old, err := e.Probe(ctx, filepath.Join(f.path, a.entry))
			if err != nil || old.Protocol != 1 || old.Dirty || !digest(old.Revision, 20) || old.Platform != p.Contract.Platform || !reflect.DeepEqual(old.Compatibility, p.Contract.Compatibility) {
				return fmt.Errorf("current program cannot guarantee compatible rollback")
			}
			next, err := e.Probe(ctx, filepath.Join(f.path, work, "candidate.exe"))
			if err != nil {
				return fmt.Errorf("candidate probe unavailable")
			}
			if err = p.Contract.Accept(next); err != nil {
				return err
			}
			h, err = f.hash(p.Target)
			if p.Replace == "" {
				if !os.IsNotExist(err) {
					return fmt.Errorf("inactive slot appeared")
				}
			} else if err != nil || h != p.Replace {
				return fmt.Errorf("inactive slot changed")
			}
			if err = f.dataDirs(); err != nil {
				return err
			}
			metadata := filepath.Join(f.path, "data", "state", "metadata.sqlite3")
			_, statErr := f.root.Lstat("data/state/metadata.sqlite3")
			hasMetadata := statErr == nil
			if statErr != nil && !os.IsNotExist(statErr) {
				return statErr
			}

			backupLeaf := fmt.Sprintf("backup-%d.sqlite3", slotIndex(p.Target))
			if backupLeaf == a.state.Backup {
				return fmt.Errorf("backup still referenced; no replacement")
			}
			if st, err := f.root.Lstat(work + "/" + backupLeaf); err == nil {
				if !safeInfo(st, false) || a.state.Owned[p.Target] == "" {
					return fmt.Errorf("unowned backup slot; no overwrite")
				}
				version, err := e.Check(ctx, filepath.Join(f.path, work, backupLeaf))
				if err != nil || version != p.Contract.Compatibility.Metadata {
					return fmt.Errorf("inactive backup integrity unverified")
				}
			} else if !os.IsNotExist(err) {
				return err
			}
			backup := ""
			if hasMetadata {
				schema, err := e.Check(ctx, metadata)
				if err != nil || schema != p.Contract.Compatibility.Metadata {
					return fmt.Errorf("metadata format/integrity check failed")
				}
				backup = backupLeaf
			}
			a.doc["ui_source_commit"], _ = json.Marshal(old.Revision)
			data, err := a.next(Record{Entry: p.Target, SHA256: p.Contract.SHA256, Compatibility: p.Contract.Compatibility, Revision: p.Contract.Revision}, p.Contract.Tag, backup)
			if err != nil {
				return err
			}
			for _, n := range []string{"backup.tmp", "backup.capture.tmp", "program.tmp"} {
				if err = absent(f, work+"/"+n); err != nil {
					return err
				}
			}
			i := intent{1, hashBytes(a.raw), hashBytes(data), p.Target, p.Contract.SHA256}
			b, _ := json.Marshal(i)
			if err = f.publish(work+"/intent.json", b); err != nil {
				return err
			}
			if hasMetadata {
				if err = e.Backup(ctx, metadata, filepath.Join(f.path, work, "backup.tmp")); err != nil {
					return fmt.Errorf("metadata backup failed; original entrypoint retained")
				}
				schema, err := e.Check(ctx, filepath.Join(f.path, work, "backup.tmp"))
				if err != nil || schema != p.Contract.Compatibility.Metadata {
					return fmt.Errorf("metadata backup verification failed")
				}
				if err = f.root.Rename(work+"/backup.tmp", work+"/"+backup); err != nil {
					return err
				}
			}
			src, err := f.root.Open(work + "/candidate.exe")
			if err != nil {
				return err
			}
			err = f.fresh(work+"/program.tmp", src, MaxProductBytes)
			src.Close()
			if err != nil {
				return err
			}
			if h, err = f.hash(work + "/program.tmp"); err != nil || h != p.Contract.SHA256 {
				return fmt.Errorf("program copy hash mismatch")
			}
			if err = f.root.Rename(work+"/program.tmp", p.Target); err != nil {
				return err
			}
			if h, err = f.hash(a.entry); err != nil || h != a.hashes[a.entry] {
				return fmt.Errorf("current program changed before publication")
			}
			if err = f.runtime(p.Contract.Compatibility); err != nil {
				return err
			}
			if err = f.replaceAlpha(ctx, a.raw, data); err != nil {
				return err
			}
			r = Receipt{"applied", p.Target, a.entry, backup, "Program switched; start the original shortcut once. No Run restarted; conversation state not inferred."}
			return f.finish()
		})
	})
	return r, err
}
func (f *files) dataDirs() error {
	for _, n := range []string{"data", "data/state"} {
		st, err := f.root.Lstat(n)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil || !safeInfo(st, true) {
			return fmt.Errorf("state directory must not be linked")
		}
	}
	if st, err := f.root.Lstat("data/state/metadata.sqlite3"); err == nil && !safeInfo(st, false) {
		return fmt.Errorf("metadata must not be linked")
	}
	return nil
}
func (f *files) remove(name string) error {
	st, err := f.root.Lstat(name)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil || !safeInfo(st, false) {
		return fmt.Errorf("unexpected update file; preserved")
	}
	return f.root.Remove(name)
}
func (f *files) finish() error {
	for _, n := range []string{"candidate.exe", "pending.json", "intent.json"} {
		if err := f.remove(work + "/" + n); err != nil {
			return fmt.Errorf("entrypoint may be committed; inspect status and repair: %w", err)
		}
	}
	return nil
}
