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

// Test: fossil is found in its home, then on PATH (R206, R257)
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

// Test: the install hint names only its own platform (R257, R258, R259)
func TestInstallHintOwnPlatformOnly(t *testing.T) {
	archives := []string{"fossil-linux-x64", "fossil-mac-arm", "fossil-mac-x64", "fossil-w64"}
	cases := []struct {
		goos, goarch string
		archive      string   // "" means none
		want, reject []string // substrings the hint must and must not carry
	}{
		{"linux", "amd64", "fossil-linux-x64", []string{"~/.claude/bin", "chmod"}, []string{"Expand-Archive", "USERPROFILE"}},
		{"darwin", "arm64", "fossil-mac-arm", []string{"~/.claude/bin", "chmod"}, []string{"Expand-Archive", "USERPROFILE"}},
		{"darwin", "amd64", "fossil-mac-x64", []string{"~/.claude/bin", "chmod"}, []string{"Expand-Archive", "USERPROFILE"}},
		{"windows", "amd64", "fossil-w64", []string{"Expand-Archive", "%USERPROFILE%"}, []string{"chmod", "tar ", "mkdir -p"}},
		{"windows", "arm64", "fossil-w64", []string{"Expand-Archive", "%USERPROFILE%"}, []string{"chmod", "tar ", "mkdir -p"}},
		{"linux", "arm64", "", []string{"download.html", "~/.claude/bin", "linux/arm64"}, []string{"chmod", "Expand-Archive"}},
		{"freebsd", "amd64", "", []string{"download.html", "~/.claude/bin"}, []string{"chmod", "Expand-Archive"}},
	}
	for _, c := range cases {
		hint := installHint(c.goos, c.goarch)
		for _, a := range archives {
			if has := strings.Contains(hint, a); has != (a == c.archive) {
				t.Errorf("%s/%s: names %s = %v; hint:\n%s", c.goos, c.goarch, a, has, hint)
			}
		}
		for _, s := range c.want {
			if !strings.Contains(hint, s) {
				t.Errorf("%s/%s: hint lacks %q:\n%s", c.goos, c.goarch, s, hint)
			}
		}
		for _, s := range c.reject {
			if strings.Contains(hint, s) {
				t.Errorf("%s/%s: hint carries %q:\n%s", c.goos, c.goarch, s, hint)
			}
		}
	}
}
