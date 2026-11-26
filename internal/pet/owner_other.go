//go:build !windows

package pet

import "fmt"

type Owner struct{}

func CaptureOwner() (*Owner, error) { return nil, fmt.Errorf("desktop reminder requires Windows") }
func (*Owner) Activate() bool       { return false }
func (*Owner) Close()               {}
