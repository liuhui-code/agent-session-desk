//go:build !windows

package backend

import "errors"

func verifyBundleProcess(uint32, uint64, string, string) error {
	return errors.New("Windows terminal bundle cannot be used on this platform")
}
