// CRC: crc-MCPServer.md | Test: test-AppActivation.md
package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

// addApp writes apps/<name>/app.lua and activates it with only the
// lua/<name> directory link (no lua/<name>.lua entry link).
func addApp(t *testing.T, baseDir, name, code string) {
	t.Helper()
	appDir := filepath.Join(baseDir, "apps", name)
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "app.lua"), []byte(code), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join("..", "apps", name), filepath.Join(baseDir, "lua", name)); err != nil {
		t.Skipf("cannot create symlinks: %v", err)
	}
}

func runLua(t *testing.T, s *Server, code string) string {
	t.Helper()
	result, err := callHandleRun(s, code)
	if err != nil {
		t.Fatalf("handleRun(%q): %v", code, err)
	}
	if result.IsError {
		t.Fatalf("handleRun(%q) tool error: %v", code, result.Content)
	}
	return getTextContent(result)
}

// Test: the mcp app's entry extends the mcp global (R183)
func TestMCPAppEntryExtendsGlobal(t *testing.T) {
	s, cleanup := createTestServerWithSessionSetup(t, func(baseDir string) {
		addApp(t, baseDir, "mcp", "mcp.extended = 'yes'\n")
	})
	defer cleanup()
	if got := runLua(t, s, "return mcp.extended"); got != `"yes"` {
		t.Errorf("mcp.extended = %q, want the mcp app's entry to have run", got)
	}
}

// Test: mcp:app loads an app's entry through its activation link (R189)
func TestAppEntryLoadsThroughActivationLink(t *testing.T) {
	s, cleanup := createTestServerWithSessionSetup(t, func(baseDir string) {
		addApp(t, baseDir, "demo", "demo = {loaded = 'yes'}\n")
	})
	defer cleanup()
	if got := runLua(t, s, "return mcp:app('demo').loaded"); got != `"yes"` {
		t.Errorf("mcp:app('demo').loaded = %q, want yes", got)
	}
}
