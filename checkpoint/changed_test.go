// CRC: crc-Checkpoint.md | Test: test-Checkpoint.md
package checkpoint

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

// installed makes an app as app-console installs a downloaded one: a baseline,
// then checkpoint.fossil copied to original.fossil. It also leaves an editor
// backup file, which never counts.
func installed(t *testing.T) App {
	t.Helper()
	a := newApp(t)
	writeFile(t, a.Dir, "util.lua", "-- util\n")
	if _, err := a.Baseline(); err != nil {
		t.Fatal(err)
	}
	src, err := os.Open(a.repo())
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	dst, err := os.Create(filepath.Join(a.Dir, "original.fossil"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		t.Fatal(err)
	}
	dst.Close()
	writeFile(t, a.Dir, "app.lua~", "-- backup\n")
	return a
}

func expectChanged(t *testing.T, a App, want bool, what string) {
	t.Helper()
	got, err := a.Changed()
	if err != nil {
		t.Fatalf("%s: Changed: %v", what, err)
	}
	if got != want {
		t.Errorf("%s: Changed = %v, want %v", what, got, want)
	}
}

// Test: changed since install (R215, R216, R217)
func TestChangedSinceInstall(t *testing.T) {
	expectChanged(t, newApp(t), false, "an app without original.fossil")

	a := installed(t)
	writeFile(t, a.Dir, "app.lua", "-- edited\n")
	expectChanged(t, a, true, "a file edited")

	a = installed(t)
	writeFile(t, a.Dir, "new.lua", "-- new\n")
	expectChanged(t, a, true, "a file added")

	a = installed(t)
	if err := os.Remove(filepath.Join(a.Dir, "util.lua")); err != nil {
		t.Fatal(err)
	}
	expectChanged(t, a, true, "a file removed")
}

// Test: fossil's own files do not count as changes (R216)
func TestChangedIgnoresFossilFiles(t *testing.T) {
	expectChanged(t, installed(t), false, "an untouched installed app")
}
