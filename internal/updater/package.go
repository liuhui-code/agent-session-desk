package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PackageContract is an explicitly requested local release build operation.
// It does not publish a release. Product bytes are rechecked after the probe.
func PackageContract(ctx context.Context, program, tag, output string, probe func(context.Context, string) (Probe, error)) error {
	if !tagPattern.MatchString(tag) {
		return fmt.Errorf("explicit release tag required")
	}
	path, err := filepath.Abs(program)
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer root.Close()
	f := &files{root, filepath.Dir(path)}
	hash, err := f.hash(filepath.Base(path))
	if err != nil {
		return err
	}
	p, err := probe(ctx, path)
	if err != nil {
		return err
	}
	st, err := root.Stat(filepath.Base(path))
	if err != nil {
		return err
	}
	c := Contract{Protocol: 1, Tag: tag, Platform: p.Platform, Revision: p.Revision, Product: "asd-windows-amd64.exe", SHA256: hash, Size: st.Size(), Compatibility: p.Compatibility}
	if err = c.Accept(p); err != nil {
		return err
	}
	if h, e := f.hash(filepath.Base(path)); e != nil || h != hash {
		return fmt.Errorf("release program changed during probe")
	}
	data, _ := json.MarshalIndent(c, "", "  ")
	out, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	dest, err := os.OpenRoot(filepath.Dir(out))
	if err != nil {
		return err
	}
	defer dest.Close()
	return (&files{dest, filepath.Dir(out)}).publish(filepath.Base(out), data)
}
