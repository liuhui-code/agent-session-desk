package store

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Snapshot creates a consistent copy at a new path. It never overwrites an
// existing destination, including an empty file. The source is opened read-only.
func Snapshot(ctx context.Context, source, destination string) error {
	db, err := openReadOnly(source)
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err = validateMetadata(ctx, db); err != nil {
		return err
	}
	destination, err = filepath.Abs(destination)
	if err != nil {
		return err
	}
	if _, err = os.Lstat(destination); err == nil {
		return fmt.Errorf("snapshot destination must not exist")
	} else if !os.IsNotExist(err) {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(destination), ".asd-snapshot-*.sqlite3")
	if err != nil {
		return err
	}
	stage := f.Name()
	owned, err := f.Stat()
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	defer func() {
		if current, e := os.Lstat(stage); e == nil && owned != nil && os.SameFile(owned, current) {
			os.Remove(stage)
		}
	}()
	if err == nil {
		_, err = db.ExecContext(ctx, "VACUUM INTO ?", stage)
	}
	if err == nil {
		var output *sql.DB
		output, err = openReadOnly(stage)
		if err == nil {
			_, err = validateMetadata(ctx, output)
			output.Close()
		}
	}
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		return fmt.Errorf("snapshot failed before publication; source preserved: %w", err)
	}
	// A same-directory hard link publishes the complete database atomically and
	// refuses an existing destination, including one created after the precheck.
	// Unsupported filesystems fail explicitly; no overwrite-rename fallback.
	if err = os.Link(stage, destination); err != nil {
		return fmt.Errorf("publish snapshot (requires hard-link support and absent destination): %w", err)
	}
	if err = os.Remove(stage); err != nil {
		return fmt.Errorf("snapshot published at %s; temporary cleanup failed: %w", destination, err)
	}

	return nil
}

func openReadOnly(path string) (*sql.DB, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("metadata source must be a regular file")
	}
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	uri := url.URL{Scheme: "file", Path: p}
	uri.RawQuery = url.Values{"mode": {"ro"}, "_pragma": {"busy_timeout(1000)", "synchronous(FULL)", "cache_size(-1024)"}}.Encode()
	db, err := sql.Open("sqlite", uri.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return db, nil
}
