// CRC: crc-Checkpoint.md | Test: test-Checkpoint.md
package checkpoint

import (
	"slices"
	"strings"
	"testing"
)

// withHistory gives the app an update checkpoint, a local checkpoint and a
// trunk checkpoint.
func withHistory(t *testing.T, a App) {
	t.Helper()
	if _, err := a.Baseline(); err != nil {
		t.Fatal(err)
	}
	writeFile(t, a.Dir, "app.lua", "-- update\n")
	if err := a.Update("up"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, a.Dir, "app.lua", "-- local\n")
	if err := a.Local("loc"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, a.Dir, "app.lua", "-- fast\n")
	mustSave(t, a, "fast")
}

func branchList(t *testing.T, a App) []string {
	t.Helper()
	names, err := a.branchNames()
	if err != nil {
		t.Fatal(err)
	}
	var list []string
	for n := range names {
		list = append(list, n)
	}
	slices.Sort(list)
	return list
}

// Test: baseline preserves branches (R74, R127, R192)
func TestBaselinePreservesBranches(t *testing.T) {
	a := newApp(t)
	withHistory(t, a)
	restored, err := a.Baseline()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(restored, ",") != "updates,local" {
		t.Errorf("restored %v, want [updates local]", restored)
	}
	if got := strings.Join(branchList(t, a), ","); got != "local,trunk,updates" {
		t.Errorf("branches after baseline: %s", got)
	}
	if n, _ := a.Count(); n != 0 {
		t.Errorf("Count after baseline = %d, want 0", n)
	}
	assertNoAside(t, a)
}

// baselineFails runs a baseline through the failing shim and checks that the
// repository and checkout are exactly as before.
func baselineFails(t *testing.T, failOn, wantInErr string) {
	t.Helper()
	a := newApp(t)
	withHistory(t, a)
	before := snapshot(t, a)
	_, err := withShim(t, a, failOn).Baseline()
	if err == nil {
		t.Fatalf("baseline with FAIL_ON=%s reported success", failOn)
	}
	if !strings.Contains(err.Error(), wantInErr) {
		t.Errorf("error lacks %q: %v", wantInErr, err)
	}
	after := snapshot(t, a)
	for name, sum := range before {
		if after[name] != sum {
			t.Errorf("%s changed by a failed baseline", name)
		}
	}
	assertNoAside(t, a)
}

// Test: a failed import leaves the repository byte-identical (R192, R193)
func TestBaselineFailedImportRestores(t *testing.T) {
	baselineFails(t, "import", "checkpoints unchanged")
}

// Test: a silent import is caught by the branch check (R192, R193)
func TestBaselineSilentImportCaught(t *testing.T) {
	baselineFails(t, "import-noop", "missing after its import")
}

// Test: a failed export leaves the repository byte-identical (R193)
func TestBaselineFailedExportStops(t *testing.T) {
	baselineFails(t, "export", "could not export branch updates")
}

// Test: branch names match exactly (R194)
func TestBaselineExactBranchNames(t *testing.T) {
	a := newApp(t)
	mustSave(t, a, "v1")
	fossilIn(t, a.Fossil, a.Dir, "commit", "--branch", "local-experiment", "-m", "x", "--allow-empty", "--no-warnings")
	fossilIn(t, a.Fossil, a.Dir, "checkout", "trunk", "--keep")
	restored, err := a.Baseline()
	if err != nil {
		t.Fatalf("baseline with only local-experiment: %v", err)
	}
	if len(restored) != 0 {
		t.Errorf("restored %v; local-experiment is not local", restored)
	}
}
