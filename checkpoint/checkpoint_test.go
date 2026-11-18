// CRC: crc-Checkpoint.md | Test: test-Checkpoint.md
package checkpoint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test: a branch save leaves the working tree alone (R125, R126, R213)
func TestBranchSaveLeavesWorkingTree(t *testing.T) {
	a := newApp(t)
	mustSave(t, a, "v1")
	// An existing branch: the first save takes the branch-missing path.
	if err := a.Local("seed"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, a.Dir, "app.lua", "-- wip\n")
	if err := a.Local("wip"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, a.Dir, "app.lua"); got != "-- wip\n" {
		t.Fatalf("working tree changed by a branch save: app.lua = %q, want the uncommitted edit", got)
	}
	if got := catRev(t, a, "app.lua", "local"); got != "-- wip\n" {
		t.Errorf("local branch holds %q, want the edit", got)
	}
	if got := catRev(t, a, "app.lua", "trunk"); got != "-- v1\n" {
		t.Errorf("trunk holds %q, want v1", got)
	}
}

// Test: a closed leaf is reopened (R125)
func TestClosedLeafReopened(t *testing.T) {
	a := newApp(t)
	mustSave(t, a, "v1")
	for _, args := range [][]string{
		{"commit", "--branch", "local", "-m", "skeleton", "--allow-empty", "--no-warnings"},
		{"amend", "current", "--close"},
		{"checkout", "trunk", "--keep"},
	} {
		fossilIn(t, a.Fossil, a.Dir, args...)
	}
	writeFile(t, a.Dir, "app.lua", "-- v2\n")
	if err := a.Local("again"); err != nil {
		t.Fatalf("save to a closed leaf: %v", err)
	}
	if got := catRev(t, a, "app.lua", "local"); got != "-- v2\n" {
		t.Errorf("local holds %q, want v2", got)
	}
}

// Test: a missing branch is created (R125, R126)
func TestMissingBranchCreated(t *testing.T) {
	a := newApp(t)
	if _, err := a.Baseline(); err != nil {
		t.Fatal(err)
	}
	if err := a.Update("first"); err != nil {
		t.Fatal(err)
	}
	names, err := a.branchNames()
	if err != nil {
		t.Fatal(err)
	}
	if !names["updates"] {
		t.Fatalf("no updates branch after Update; branches %v", names)
	}
	if got := catRev(t, a, "app.lua", "updates"); got != "-- v1\n" {
		t.Errorf("updates holds %q", got)
	}
}

// newOuterCheckout makes an enclosing fossil checkout and returns its root.
func newOuterCheckout(t *testing.T, fossil string) string {
	t.Helper()
	outer := t.TempDir()
	repo := filepath.Join(outer, "outer.fossil")
	fossilIn(t, fossil, outer, "init", repo, "--project-name", "outer")
	fossilIn(t, fossil, outer, "open", repo, "--force")
	writeFile(t, outer, "readme", "outer\n")
	fossilIn(t, fossil, outer, "add", "readme")
	fossilIn(t, fossil, outer, "commit", "-m", "outer", "--no-warnings")
	return outer
}

// Test: an enclosing checkout is not walked into (R211, R212)
func TestEnclosingCheckoutNotWalkedInto(t *testing.T) {
	fossil := testFossil(t)
	outer := newOuterCheckout(t, fossil)
	before := fossilIn(t, fossil, outer, "timeline", "-t", "ci", "--oneline")
	a := App{Dir: filepath.Join(outer, "app"), Fossil: fossil}
	if err := os.Mkdir(a.Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, a.Dir, "app.lua", "-- v1\n")

	if _, err := a.Save("x"); err != nil {
		t.Fatalf("save inside an enclosing checkout: %v", err)
	}
	if err := a.verify(); err != nil {
		t.Errorf("app checkout is not its own: %v", err)
	}
	if after := fossilIn(t, fossil, outer, "timeline", "-t", "ci", "--oneline"); after != before {
		t.Errorf("enclosing checkout's timeline changed:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// Test: a foreign binding stops the command (R211, R212)
func TestForeignBindingStops(t *testing.T) {
	a := newApp(t)
	mustSave(t, a, "v1")
	other := filepath.Join(t.TempDir(), "other.fossil")
	fossilIn(t, a.Fossil, a.Dir, "init", other)
	fossilIn(t, a.Fossil, a.Dir, "close", "--force")
	fossilIn(t, a.Fossil, a.Dir, "open", other, "--force", "--nested", "--keep")
	otherBefore := fossilIn(t, a.Fossil, a.Dir, "timeline", "-n", "0", "-t", "ci", "--oneline", "-R", other)

	writeFile(t, a.Dir, "app.lua", "-- v2\n")
	_, err := a.Save("x")
	if err == nil {
		t.Fatal("Save succeeded through a checkout bound to another repository")
	}
	if !strings.Contains(err.Error(), "binding mismatch") {
		t.Errorf("error does not name the binding: %v", err)
	}
	if after := fossilIn(t, a.Fossil, a.Dir, "timeline", "-n", "0", "-t", "ci", "--oneline", "-R", other); after != otherBefore {
		t.Errorf("the other repository gained a check-in:\n%s", after)
	}
}

// Test: a fossil failure surfaces (R210)
func TestFossilFailureSurfaces(t *testing.T) {
	a := newApp(t)
	writeFile(t, a.Dir, "checkpoint.fossil", "not a repository\n")
	_, err := a.Save("x")
	if err == nil {
		t.Fatal("Save reported success against a corrupt repository")
	}
	if !strings.Contains(err.Error(), "fossil open") {
		t.Errorf("error does not name the failing fossil command (open): %v", err)
	}
}

// Test: a same-second, same-size edit is saved (R214)
func TestSameSecondEditSaved(t *testing.T) {
	a := newApp(t)
	mustSave(t, a, "v1")
	path := filepath.Join(a.Dir, "app.lua")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, a.Dir, "app.lua", "-- v2\n") // same size as "-- v1\n"
	if err := os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}
	saved, err := a.Save("y")
	if err != nil {
		t.Fatal(err)
	}
	if !saved {
		t.Fatal("a same-size edit with an unchanged mtime was not saved")
	}
	if n, _ := a.Count(); n != 2 {
		t.Errorf("Count = %d, want 2", n)
	}
}
