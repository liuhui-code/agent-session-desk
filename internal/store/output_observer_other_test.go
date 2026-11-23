//go:build !windows

package store_test

import "os"

func openOutputObserver(path string) (*os.File, error) { return os.Open(path) }
