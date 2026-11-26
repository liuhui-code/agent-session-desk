//go:build windows

package instance

import (
	"encoding/json"
	"os"
	"testing"
	"unsafe"
)

// This read-only test adapter exposes only identity of its explicit synthetic
// scope. It is not a product command or a general process/window inspector.
func TestInstanceInspectHelper(t *testing.T) {
	output := os.Getenv("ASD_INSTANCE_INSPECT_OUTPUT")
	if output == "" {
		t.Skip("test observer helper")
	}
	key, err := ScopeKey("psmux", os.Getenv("ASD_INSTANCE_INSPECT_LOCATION"), os.Getenv("ASD_INSTANCE_INSPECT_METADATA"))
	if err != nil {
		t.Fatal(err)
	}
	name, err := objectName(key)
	if err != nil {
		t.Fatal(err)
	}
	hwnd := endpoint(name)
	var pid uint32
	if hwnd != 0 {
		windowPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	}
	value := map[string]any{"endpoint": hwnd, "owner_pid": pid, "target": boundTarget(hwnd)}
	file, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err = json.NewEncoder(file).Encode(value); err != nil {
		t.Fatal(err)
	}
}
