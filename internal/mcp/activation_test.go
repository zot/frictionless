// CRC: crc-MCPServer.md | Test: test-AppActivation.md
package mcp

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

// Test: linkapp makes exactly one link and cleans up the old scheme (R184-R188)
func TestLinkappOneLink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("linkapp is a bash script")
	}
	base := t.TempDir()
	script, err := os.ReadFile(filepath.Join("..", "..", "install", "linkapp"))
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(base, "linkapp"), script, 0o755)
	os.MkdirAll(filepath.Join(base, "apps", "demo", "viewdefs"), 0o755)
	os.WriteFile(filepath.Join(base, "apps", "demo", "app.lua"), []byte("--"), 0o644)
	// Old-scheme links left by an earlier linkapp.
	os.MkdirAll(filepath.Join(base, "lua"), 0o755)
	os.MkdirAll(filepath.Join(base, "viewdefs"), 0o755)
	os.Symlink("../apps/demo/app.lua", filepath.Join(base, "lua", "demo.lua"))
	os.Symlink("../apps/demo/viewdefs/D.DEFAULT.html", filepath.Join(base, "viewdefs", "D.DEFAULT.html"))

	run := func(args ...string) string {
		out, err := exec.Command(filepath.Join(base, "linkapp"), args...).CombinedOutput()
		if err != nil {
			t.Fatalf("linkapp %v: %v\n%s", args, err, out)
		}
		return string(out)
	}

	run("add", "demo")
	if target, err := os.Readlink(filepath.Join(base, "lua", "demo")); err != nil || target != "../apps/demo" {
		t.Errorf("lua/demo -> %q (%v), want ../apps/demo", target, err)
	}
	for _, old := range []string{"lua/demo.lua", "viewdefs/D.DEFAULT.html"} {
		if _, err := os.Lstat(filepath.Join(base, old)); err == nil {
			t.Errorf("old-scheme link %s survived add", old)
		}
	}
	if out := run("list"); !strings.Contains(out, "  demo\n") {
		t.Errorf("list output %q does not report demo", out)
	}

	run("remove", "demo")
	if _, err := os.Lstat(filepath.Join(base, "lua", "demo")); err == nil {
		t.Error("lua/demo survived remove")
	}
}
