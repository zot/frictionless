// CRC: crc-Checkpoint.md | Test: test-Checkpoint.md
package checkpoint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustStatus(t *testing.T, a App) Status {
	t.Helper()
	st, err := a.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	return st
}

// Test: Status writes nothing (R218, R219, R220)
func TestStatusWritesNothing(t *testing.T) {
	a := newApp(t)
	if st := mustStatus(t, a); st.Count != 0 {
		t.Errorf("Count without a repository = %d, want 0", st.Count)
	}

	mustSave(t, a, "v1")
	writeFile(t, a.Dir, "app.lua", "-- v2\n")
	mustSave(t, a, "v2")
	fossilIn(t, a.Fossil, a.Dir, "close", "--force")
	if st := mustStatus(t, a); st.Count != 2 {
		t.Errorf("Count with the checkout closed = %d, want 2", st.Count)
	}
	if exists(a.checkout()) {
		t.Fatal("Status opened a checkout")
	}

	if _, err := a.Count(); err != nil { // reopens the checkout
		t.Fatal(err)
	}
	before := snapshot(t, a)
	mustStatus(t, a)
	after := snapshot(t, a)
	for name, sum := range before {
		if after[name] != sum {
			t.Errorf("Status changed %s", name)
		}
	}
}

// Test: Status reports downloaded and changed (R218)
func TestStatusReportsDownloadedAndChanged(t *testing.T) {
	a := newApp(t)
	if err := a.Install(); err != nil {
		t.Fatal(err)
	}
	if st := mustStatus(t, a); !st.Downloaded || st.Changed {
		t.Fatalf("after Install: %+v, want downloaded and unchanged", st)
	}
	writeFile(t, a.Dir, "app.lua", "-- edited\n")
	if st := mustStatus(t, a); !st.Changed {
		t.Fatalf("after an edit: %+v, want changed", st)
	}
}

// Test: Install creates original.fossil, and saves leave it out (R221, R222)
func TestInstallKeepsOriginalOutOfCheckpoints(t *testing.T) {
	a := newApp(t)
	if err := a.Install(); err != nil {
		t.Fatal(err)
	}
	if !exists(filepath.Join(a.Dir, "original.fossil")) {
		t.Fatal("Install made no original.fossil")
	}
	if changed, err := a.Changed(); err != nil || changed {
		t.Fatalf("Changed after Install = %v, %v; want false", changed, err)
	}
	writeFile(t, a.Dir, "app.lua", "-- edited\n")
	mustSave(t, a, "edit")
	if ls := fossilIn(t, a.Fossil, a.Dir, "ls"); strings.Contains(ls, "original.fossil") {
		t.Fatalf("a save committed original.fossil:\n%s", ls)
	}
}

// Test: a failed Install copy leaves nothing behind (R221)
func TestInstallFailedCopyLeavesNothing(t *testing.T) {
	a := newApp(t)
	blocker := filepath.Join(a.Dir, "original.fossil")
	if err := os.MkdirAll(filepath.Join(blocker, "x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := a.Install(); err == nil {
		t.Fatal("Install succeeded although original.fossil could not be replaced")
	}
	if m, _ := filepath.Glob(filepath.Join(a.Dir, ".#original.*")); len(m) > 0 {
		t.Fatalf("Install left its temporary behind: %v", m)
	}
}
