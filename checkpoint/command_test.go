// CRC: crc-Checkpoint.md | Test: test-Checkpoint.md
package checkpoint

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// Test: Command prints, and calls onChange only after a changing verb (R79, R204, R208, R209)
func TestCommandPrintsAndNotifiesOnChange(t *testing.T) {
	a := newApp(t)
	appsDir := filepath.Dir(a.Dir)
	changes := 0
	run := func(args ...string) string {
		t.Helper()
		var out bytes.Buffer
		if err := Command(args, appsDir, &out, func() { changes++ }); err != nil {
			t.Fatalf("Command %v: %v", args, err)
		}
		return out.String()
	}

	if out := run("save", "demo", "first"); out != "Created checkpoint: first\n" {
		t.Errorf("save printed %q", out)
	}
	if changes != 1 {
		t.Errorf("onChange ran %d times after save, want 1", changes)
	}
	if out := run("count", "demo"); out != "1\n" {
		t.Errorf("count printed %q, want 1", out)
	}
	if out := run("list", "demo"); strings.Count(out, "\n") != 2 || !strings.Contains(out, "first") {
		t.Errorf("list printed %q, want the checkpoint and fossil's initial check-in", out)
	}
	if changes != 1 {
		t.Errorf("onChange ran %d times in all, want 1: count and list change nothing", changes)
	}
	if err := Command([]string{"bogus", "demo"}, appsDir, &bytes.Buffer{}, nil); err == nil {
		t.Error("an unknown verb was not an error")
	}
	if err := Command([]string{"count", "missing"}, appsDir, &bytes.Buffer{}, nil); err == nil || !strings.Contains(err.Error(), "app not found") {
		t.Errorf("a missing app: err = %v", err)
	}
}
