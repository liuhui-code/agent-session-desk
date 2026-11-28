package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

type Status struct {
	Current        string `json:"current"`
	Previous       string `json:"previous,omitempty"`
	Pending        bool   `json:"pending"`
	RepairRequired bool   `json:"repair_required"`
}

func (e Engine) Status(ctx context.Context) (Status, error) {
	var out Status
	err := e.locked(ctx, func(f *files) error {
		a, err := f.installation()
		if err != nil {
			return err
		}
		out.Current = a.entry
		out.Previous = a.state.Previous.Entry
		for n, dst := range map[string]*bool{"pending.json": &out.Pending, "intent.json": &out.RepairRequired} {
			_, err = f.root.Lstat(work + "/" + n)
			if err == nil {
				*dst = true
			} else if !os.IsNotExist(err) {
				return err
			}
		}
		return nil
	})
	return out, err
}
func (e Engine) Rollback(ctx context.Context) (Receipt, error) {
	var r Receipt
	err := e.locked(ctx, func(f *files) error {
		return e.offline(ctx, func() error {
			if err := noIntent(f); err != nil {
				return err
			}
			if err := absent(f, work+"/pending.json"); err != nil {
				return err
			}
			if err := absent(f, "alpha.update.tmp"); err != nil {
				return err
			}
			a, err := f.installation()
			if err != nil {
				return err
			}
			target := a.state.Previous
			if target.Entry == "" || target.Entry == a.entry || e.Probe == nil || e.Check == nil {
				return fmt.Errorf("verified previous program unavailable")
			}
			h, err := f.hash(target.Entry)
			if err != nil || h != target.SHA256 {
				return fmt.Errorf("previous program hash changed; no rollback")
			}
			p, err := e.Probe(ctx, filepath.Join(f.path, target.Entry))
			if err != nil || p.Protocol != 1 || p.Dirty || p.Platform != "windows/amd64" || p.Revision != target.Revision || !reflect.DeepEqual(p.Compatibility, target.Compatibility) {
				return fmt.Errorf("previous program compatibility unverified")
			}
			if err = f.runtime(target.Compatibility); err != nil {
				return err
			}
			if err = e.checkData(ctx, f, target.Compatibility.Metadata); err != nil {
				return err
			}
			data, err := a.next(target, "rollback", a.state.Backup)
			if err != nil {
				return err
			}
			i := intent{Before: hashBytes(a.raw), After: hashBytes(data), Protocol: 1}
			b, _ := json.Marshal(i)
			if err = f.publish(work+"/intent.json", b); err != nil {
				return err
			}
			if err = f.replaceAlpha(ctx, a.raw, data); err != nil {
				return err
			}
			r = Receipt{"rolled_back", target.Entry, a.entry, a.state.Backup, "Program only; current metadata and background Runs preserved. No workbench started."}
			return f.finish()
		})
	})
	return r, err
}
func (e Engine) checkData(ctx context.Context, f *files, want int) error {
	if err := f.dataDirs(); err != nil {
		return err
	}
	_, err := f.root.Lstat("data/state/metadata.sqlite3")
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	got, err := e.Check(ctx, filepath.Join(f.path, "data", "state", "metadata.sqlite3"))
	if err != nil || got != want {
		return fmt.Errorf("current metadata format/integrity mismatch; no rollback")
	}
	return nil
}
func (e Engine) Repair(ctx context.Context) (Receipt, error) {
	var r Receipt
	err := e.locked(ctx, func(f *files) error {
		return e.offline(ctx, func() error {
			var i intent
			b, err := f.read(work+"/intent.json", MaxManifestBytes)
			if err != nil {
				return err
			}
			if decode(b, &i) != nil || i.Protocol != 1 || !digest(i.Before, 32) || !digest(i.After, 32) || (i.Target != "" && (slotIndex(i.Target) == 0 || !digest(i.SHA256, 32))) {
				return fmt.Errorf("invalid interrupted update; preserve files for inspection")
			}
			a, err := f.installation()
			if err != nil {
				return err
			}
			stamp := hashBytes(a.raw)
			switch stamp {
			case i.Before:
				if i.Target != "" {
					h, err := f.hash(i.Target)
					if err == nil && h == i.SHA256 {
						if i.Target == a.entry || i.Target == a.state.Previous.Entry {
							return fmt.Errorf("target is referenced; no removal")
						}
						if err = f.remove(i.Target); err != nil {
							return err
						}
					} else if err != nil && !os.IsNotExist(err) {
						return err
					}
				}
				if i.Target != "" && a.state.Owned[i.Target] == "" {
					name := fmt.Sprintf("backup-%d.sqlite3", slotIndex(i.Target))
					if name == a.state.Backup {
						return fmt.Errorf("backup remains referenced")
					}
					if err = f.remove(work + "/" + name); err != nil {
						return err
					}
				}
				r = Receipt{Action: "aborted_preparation", Current: a.entry, Previous: a.state.Previous.Entry, Note: "Original pointer retained; no program started, no database restored. Explicit apply may be reviewed again."}
			case i.After:
				r = Receipt{Action: "finalized_switch", Current: a.entry, Previous: a.state.Previous.Entry, Note: "Completed pointer retained; no startup replay. Use previous program rollback if needed."}
			default:
				return fmt.Errorf("entrypoint differs from both transaction states; no automatic repair")
			}
			for _, name := range []string{work + "/program.tmp", work + "/backup.tmp", work + "/backup.capture.tmp", "alpha.update.tmp"} {
				if err = f.remove(name); err != nil {
					return err
				}
			}
			// A completed switch can lose pending files during cleanup. Repair never
			// reruns the backup, executes a candidate, or changes terminal/process state.
			if stamp == i.After {
				return f.finish()
			}
			return f.remove(work + "/intent.json")
		})
	})
	return r, err
}
func (e Engine) Discard(ctx context.Context) error {
	return e.locked(ctx, func(f *files) error {
		if err := noIntent(f); err != nil {
			return err
		}
		for _, n := range []string{"candidate.exe", "pending.json", "product.tmp", "pending.json.tmp", "intent.json.tmp", "manifest.tmp"} {
			if err := f.remove(work + "/" + n); err != nil {
				return err
			}
		}
		return nil
	})
}
