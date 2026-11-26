//go:build windows

package document

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestMarkdownWindowsOpenAssociationContract(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("Windows x64 target")
	}
	if unsafe.Sizeof(shellExecuteInfo{}) != 112 {
		t.Fatal("SHELLEXECUTEINFO ABI")
	}
	path := filepath.Join(t.TempDir(), "中文 & $(literal) notes.md")
	if err := os.WriteFile(path, []byte("# fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	called := false
	err := openWith(context.Background(), path, func(info *shellExecuteInfo) error {
		called = true
		if windows.UTF16PtrToString(info.File) != path || windows.UTF16PtrToString(info.Verb) != "open" || info.Parameters != nil || info.Class != nil || info.Mask != 0x500 || info.Process != 0 || info.Show != 1 {
			t.Fatal("document path became a command or association changed")
		}
		return nil
	})
	if err != nil || !called {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := openWith(ctx, path, func(*shellExecuteInfo) error { t.Fatal("cancelled request invoked Shell"); return nil }); err == nil {
		t.Fatal("cancellation lost")
	}
	t.Log("native COM/ABI and separated path contract; ShellExecuteEx replaced by observer, default viewer not launched")
}
