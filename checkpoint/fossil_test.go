// CRC: crc-Checkpoint.md | Test: test-Checkpoint.md
package checkpoint

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Test: fossil is found in its home, then on PATH (R206, R207)
func TestFindHomeThenPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell-script stand-in for fossil")
	}
	name := "fossil"
	pathDir, home := t.TempDir(), t.TempDir()
	onPath := filepath.Join(pathDir, name)
	if err := os.WriteFile(onPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("PATH", pathDir)

	if got, err := Find(); err != nil || got != onPath {
		t.Fatalf("with fossil only on PATH: Find() = %q, %v; want %q", got, err, onPath)
	}

	inHome := filepath.Join(home, ".claude", "bin", name)
	if err := os.MkdirAll(filepath.Dir(inHome), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inHome, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got, err := Find(); err != nil || got != inHome {
		t.Fatalf("with fossil in both: Find() = %q, %v; want the home one %q", got, err, inHome)
	}

	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())
	_, err := Find()
	if !errors.Is(err, ErrNoFossil) {
		t.Fatalf("with no fossil: err = %v, want ErrNoFossil", err)
	}
	if !strings.Contains(err.Error(), "~/.claude/bin") || !strings.Contains(err.Error(), "fossil-scm.org") {
		t.Errorf("ErrNoFossil does not say how to install: %v", err)
	}
}
