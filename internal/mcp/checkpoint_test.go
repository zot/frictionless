// CRC: crc-CheckpointManager.md | Test: test-CheckpointManager.md
package mcp

import (
	"bytes"
	"crypto/sha256"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// checkpointEnv is a base dir holding a copy of install/mcp and one app,
// with HOME pointing at a fossil shim that fails on demand via FAIL_ON.
type checkpointEnv struct {
	t       *testing.T
	baseDir string
	appDir  string
	home    string
}

const fossilShim = `#!/bin/bash
case "$FAIL_ON:$1 $2" in
    "export:bundle export"|"import:bundle import") echo "shim: injected $FAIL_ON failure" >&2; exit 1 ;;
    "import-noop:bundle import") exit 0 ;;
esac
exec "%s" "$@"
`

func newCheckpointEnv(t *testing.T) *checkpointEnv {
	t.Helper()
	realHome, _ := os.UserHomeDir()
	realFossil := filepath.Join(realHome, ".claude", "bin", "fossil")
	if _, err := os.Stat(realFossil); err != nil {
		t.Skipf("fossil not installed at %s", realFossil)
	}
	script, err := os.ReadFile(filepath.Join("..", "..", "install", "mcp"))
	if err != nil {
		t.Fatal(err)
	}
	e := &checkpointEnv{t: t, baseDir: t.TempDir(), home: t.TempDir()}
	e.appDir = filepath.Join(e.baseDir, "apps", "demo")
	binDir := filepath.Join(e.home, ".claude", "bin")
	for _, d := range []string{e.appDir, binDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	e.write(filepath.Join(e.baseDir, "mcp"), string(script), 0o755)
	e.write(filepath.Join(binDir, "fossil"), strings.Replace(fossilShim, "%s", realFossil, 1), 0o755)
	e.write(filepath.Join(e.appDir, "app.lua"), "-- v1\n", 0o644)
	return e
}

func (e *checkpointEnv) write(path, content string, mode os.FileMode) {
	e.t.Helper()
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		e.t.Fatal(err)
	}
}

// mcp runs `mcp checkpoint ARGS...` and returns stdout, stderr and the exit error.
func (e *checkpointEnv) mcp(failOn string, args ...string) (string, string, error) {
	e.t.Helper()
	cmd := exec.Command(filepath.Join(e.baseDir, "mcp"), append([]string{"checkpoint"}, args...)...)
	// FRICTIONLESS_MCP makes the UI notify a no-op: no server runs here
	cmd.Env = append(os.Environ(), "HOME="+e.home, "FAIL_ON="+failOn, "FRICTIONLESS_MCP=1")
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	return out.String(), errOut.String(), err
}

func (e *checkpointEnv) mustMcp(args ...string) {
	e.t.Helper()
	if out, errOut, err := e.mcp("", args...); err != nil {
		e.t.Fatalf("mcp checkpoint %v: %v\n%s%s", args, err, out, errOut)
	}
}

// fossil runs the shimmed fossil in the app dir with no injection.
func (e *checkpointEnv) fossil(args ...string) string {
	e.t.Helper()
	cmd := exec.Command(filepath.Join(e.home, ".claude", "bin", "fossil"), args...)
	cmd.Dir = e.appDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		e.t.Fatalf("fossil %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func (e *checkpointEnv) branches() []string {
	var names []string
	for _, line := range strings.Split(e.fossil("branch", "list"), "\n") {
		if name := strings.TrimLeft(line, "* "); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// withHistory gives the app an update, a local and a fast checkpoint.
func (e *checkpointEnv) withHistory() {
	e.mustMcp("update", "demo", "u1")
	e.mustMcp("local", "demo", "l1")
	e.write(filepath.Join(e.appDir, "app.lua"), "-- v2\n", 0o644)
	e.mustMcp("save", "demo", "s1")
}

// snapshot hashes the repo and checkout files.
func (e *checkpointEnv) snapshot() map[string][32]byte {
	e.t.Helper()
	sums := map[string][32]byte{}
	for _, name := range []string{"checkpoint.fossil", ".fslckout"} {
		data, err := os.ReadFile(filepath.Join(e.appDir, name))
		if err != nil {
			e.t.Fatal(err)
		}
		sums[name] = sha256.Sum256(data)
	}
	return sums
}

func (e *checkpointEnv) assertNoAside() {
	e.t.Helper()
	if left, _ := filepath.Glob(filepath.Join(e.appDir, ".#baseline.*")); len(left) > 0 {
		e.t.Errorf("aside dir left behind: %v", left)
	}
}

func hasAll(have []string, want ...string) bool {
	set := map[string]bool{}
	for _, b := range have {
		set[b] = true
	}
	for _, w := range want {
		if !set[w] {
			return false
		}
	}
	return true
}

func TestCheckpointBaselinePreservesBranches(t *testing.T) {
	e := newCheckpointEnv(t)
	e.withHistory()
	out, errOut, err := e.mcp("", "baseline", "demo")
	if err != nil {
		t.Fatalf("baseline: %v\n%s", err, errOut)
	}
	if !strings.Contains(out, "preserved branches restored") {
		t.Errorf("stdout = %q, want preserved branches restored", out)
	}
	if b := e.branches(); !hasAll(b, "trunk", "updates", "local") {
		t.Errorf("branches = %v, want trunk, updates, local", b)
	}
	if count, _, _ := e.mcp("", "count", "demo"); strings.TrimSpace(count) != "0" {
		t.Errorf("checkpoint count = %q, want 0", count)
	}
	e.assertNoAside()
}

// The data loss ark reported: a failed import must leave the old repo intact.
func TestCheckpointBaselineFailureLeavesRepoIdentical(t *testing.T) {
	// Each failure must be reported with the step that failed
	reports := map[string]string{
		"import":      "Baseline failed for demo; checkpoints unchanged",
		"import-noop": "Baseline failed for demo; checkpoints unchanged",
		"export":      "could not export branch updates; checkpoints unchanged",
	}
	for failOn, report := range reports {
		t.Run(failOn, func(t *testing.T) {
			e := newCheckpointEnv(t)
			e.withHistory()
			before := e.snapshot()
			out, errOut, err := e.mcp(failOn, "baseline", "demo")
			if err == nil {
				t.Errorf("baseline exited 0 despite injected %s failure", failOn)
			}
			if strings.Contains(out, "Baseline set") {
				t.Errorf("stdout claims success: %q", out)
			}
			if !strings.Contains(errOut, report) {
				t.Errorf("stderr = %q, want %q", errOut, report)
			}
			after := e.snapshot()
			for name, sum := range before {
				if after[name] != sum {
					t.Errorf("%s changed after failed baseline", name)
				}
			}
			e.assertNoAside()
		})
	}
}

func TestCheckpointBaselineExactBranchNames(t *testing.T) {
	e := newCheckpointEnv(t)
	e.mustMcp("save", "demo", "s1")
	e.fossil("commit", "--branch", "local-experiment", "-m", "exp", "--allow-empty", "--no-warnings")
	e.fossil("update", "trunk")
	out, errOut, err := e.mcp("", "baseline", "demo")
	if err != nil {
		t.Fatalf("baseline: %v\n%s", err, errOut)
	}
	if strings.TrimSpace(out) != "Baseline set for demo" {
		t.Errorf("stdout = %q, want plain baseline (local-experiment is not local)", out)
	}
}
