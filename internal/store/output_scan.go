package store

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	outputTextLimit   = 64 * 1024
	outputFileLimit   = 512 * 1024
	outputPerRunLimit = 2
	outputTotalLimit  = 32
)

type outputDirectory struct {
	records []outputFile
	stages  []string
}

func scanOutput(ctx context.Context, root string) (outputDirectory, error) {
	var result outputDirectory
	dir, err := os.Open(root)
	if err != nil {
		return result, err
	}
	entries, err := dir.ReadDir(outputTotalLimit + 3)
	dir.Close()
	if err != nil && err != io.EOF {
		return result, err
	}
	if len(entries) > outputTotalLimit+2 {
		return result, fmt.Errorf("output directory exceeds entry budget; files preserved")
	}
	sequences := map[uint64]bool{}
	perRun := map[string]int{}
	for _, entry := range entries {
		if err = ctx.Err(); err != nil {
			return result, err
		}
		name := entry.Name()
		if name == ".format" {
			continue
		}
		stage := strings.HasPrefix(name, ".stage-") && strings.HasSuffix(name, ".tmp") && outputHex(strings.TrimSuffix(strings.TrimPrefix(name, ".stage-"), ".tmp"), 32)
		record := strings.HasSuffix(name, ".json") && outputHex(strings.TrimSuffix(name, ".json"), 32)
		if !stage && !record {
			return result, fmt.Errorf("unrecognized output file; directory preserved")
		}
		path := filepath.Join(root, name)
		info, e := os.Lstat(path)
		if e != nil {
			return result, e
		}
		if !info.Mode().IsRegular() || info.Size() > outputFileLimit {
			return result, fmt.Errorf("output file is linked, oversized or non-regular; preserved")
		}
		if stage {
			result.stages = append(result.stages, name)
			if len(result.stages) > 1 {
				return result, fmt.Errorf("output staging budget exceeded; preserved")
			}
			continue
		}
		f, e := os.Open(path)
		if e != nil {
			return result, e
		}
		data, e := io.ReadAll(io.LimitReader(f, outputFileLimit+1))
		closeErr := f.Close()
		if e == nil {
			e = closeErr
		}
		if e != nil {
			return result, e
		}
		if len(data) > outputFileLimit {
			return result, fmt.Errorf("output file budget exceeded; preserved")
		}
		var file outputFile
		if e = json.Unmarshal(data, &file); e != nil {
			return result, fmt.Errorf("invalid output record; file preserved")
		}
		r := file.Record
		if !outputHex(file.Scope, 64) || !outputHex(r.ID, 32) || r.ID+".json" != name || !outputHex(r.SessionID, 32) || !outputHex(r.RunID, 32) || r.BackendID == "" || len(r.BackendID) > 4096 || strings.ContainsRune(r.BackendID, 0) || r.Source != "screen_snapshot" || r.Complete || r.RecordedAt.IsZero() || r.Sequence == 0 || len(r.Text) > outputTextLimit || !utf8.ValidString(r.Text) {
			return result, fmt.Errorf("unsupported output identity, source or bounds; file preserved")
		}
		if sequences[r.Sequence] {
			return result, fmt.Errorf("duplicate retained output sequence; files preserved")
		}
		sequences[r.Sequence] = true
		key := file.Scope + ":" + r.RunID
		perRun[key]++
		if perRun[key] > outputPerRunLimit || len(result.records) >= outputTotalLimit {
			return result, fmt.Errorf("published output retention budget exceeded; files preserved")
		}
		file.name = name
		result.records = append(result.records, file)
	}
	return result, nil
}

func outputHex(value string, n int) bool {
	if len(value) != n {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
