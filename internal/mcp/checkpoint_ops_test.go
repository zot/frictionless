// CRC: crc-CheckpointOps.md | Test: test-Checkpoint.md
package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zot/frictionless/checkpoint"
)

// checkpointOpsServer starts an MCP server session whose apps directory holds a
// "good" app with one checkpoint and a "bad" app whose repository is not one.
func checkpointOpsServer(t *testing.T) (*Server, string) {
	t.Helper()
	fossil, err := checkpoint.Find()
	if err != nil {
		t.Skip("fossil is not installed")
	}
	var appsDir string
	s, cleanup := createTestServerWithSessionSetup(t, func(baseDir string) {
		appsDir = filepath.Join(baseDir, "apps")
		for name, files := range map[string]map[string]string{
			"good": {"app.lua": "-- good\n"},
			"bad":  {"app.lua": "-- bad\n", "checkpoint.fossil": "not a repository\n"},
		} {
			dir := filepath.Join(appsDir, name)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			for f, content := range files {
				if err := os.WriteFile(filepath.Join(dir, f), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
		}
		if _, err := (checkpoint.App{Dir: filepath.Join(appsDir, "good"), Fossil: fossil}).Save("one"); err != nil {
			t.Fatal(err)
		}
	})
	t.Cleanup(cleanup)
	return s, appsDir
}

// luaValue runs code returning a string and returns the tool's text.
func luaValue(t *testing.T, s *Server, code string) string {
	t.Helper()
	result, err := callHandleRun(s, code)
	if err != nil {
		t.Fatalf("run %q: %v", code, err)
	}
	if result.IsError {
		t.Fatalf("run %q: %v", code, result.Content)
	}
	return getTextContent(result)
}

// waitLua polls until the Lua expression is true.
func waitLua(t *testing.T, s *Server, expr, what string) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for !strings.Contains(luaValue(t, s, "return tostring("+expr+")"), "true") {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// Test: checkpointStatus reports every app, isolating failures (R223, R224)
func TestCheckpointStatusIsolatesFailures(t *testing.T) {
	s, _ := checkpointOpsServer(t)
	luaValue(t, s, `cpt = {} ui.checkpointStatus(cpt, "s", {"good", "bad"}, function() cpt.done = true end) return "started"`)
	waitLua(t, s, "cpt.done == true", "the checkpointStatus delivery")

	if got := luaValue(t, s, "return tostring(cpt.sError)"); !strings.Contains(got, "nil") {
		t.Fatalf("the whole operation failed: sError = %s", got)
	}
	if got := luaValue(t, s, "return tostring(cpt.s.good.count)"); !strings.Contains(got, "1") {
		t.Errorf("good app's count = %s, want 1", got)
	}
	if got := luaValue(t, s, "return tostring(cpt.s.bad.error ~= nil)"); !strings.Contains(got, "true") {
		t.Errorf("bad app has no error entry: %s", got)
	}
}

// Test: checkpointInstall installs and delivers true (R223, R225)
func TestCheckpointInstallDeliversTrue(t *testing.T) {
	s, appsDir := checkpointOpsServer(t)
	luaValue(t, s, `cpi = {} ui.checkpointInstall(cpi, "done", "good", function() cpi.cb = true end) return "started"`)
	waitLua(t, s, "cpi.cb == true", "the checkpointInstall delivery")

	if got := luaValue(t, s, "return tostring(cpi.done)"); !strings.Contains(got, "true") {
		t.Fatalf("done = %s (error %s), want true", got, luaValue(t, s, "return tostring(cpi.doneError)"))
	}
	if _, err := os.Stat(filepath.Join(appsDir, "good", "original.fossil")); err != nil {
		t.Errorf("no original.fossil after checkpointInstall: %v", err)
	}
}
