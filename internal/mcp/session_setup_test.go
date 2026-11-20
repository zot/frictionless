// CRC: crc-MCPServer.md | Test: test-SessionSetupHook.md | R253, R254
package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	lua "github.com/yuin/gopher-lua"
	"github.com/zot/ui-engine/cli"
)

// configuredServer is createTestServerWithSessionSetup stopped before the session is
// created, so hooks can be registered first. It returns the server and its base dir.
func configuredServer(t *testing.T) (*Server, string) {
	t.Helper()
	baseDir := filepath.Join(t.TempDir(), ".claude", "ui")
	os.MkdirAll(filepath.Join(baseDir, "log"), 0o755)
	os.MkdirAll(filepath.Join(baseDir, "lua"), 0o755)
	// A README.md skips auto-install, which needs the bundled binary.
	os.WriteFile(filepath.Join(baseDir, "README.md"), []byte("# Test\n**Version: 99.0.0**\n"), 0o644)
	cfg := cli.DefaultConfig()
	cfg.Server.Dir = baseDir
	uiServer := cli.NewServer(cfg)
	s := NewServer(cfg, uiServer, uiServer.GetViewdefManager(),
		func(int) (string, error) { return "http://127.0.0.1:12345", nil },
		func() int { return 0 })
	if err := s.Configure(baseDir); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	t.Cleanup(func() { s.Stop() })
	return s, baseDir
}

// Test: a hook defines a host operation the session can call (R253, R254)
func TestSessionSetupHookDefinesHostOp(t *testing.T) {
	s, _ := configuredServer(t)
	var mu sync.Mutex
	var record []string
	note := func(what string) {
		mu.Lock()
		defer mu.Unlock()
		record = append(record, what)
	}
	s.AddSessionSetupHook(func(cli.LuaRig) error { note("one"); return nil })
	s.AddSessionSetupHook(func(rig cli.LuaRig) error {
		note("two")
		rig.DefineHostOp("hookOp", 0, func([]lua.LValue) (func(context.Context) (any, error), error) {
			return func(context.Context) (any, error) { return "from-hook", nil }, nil
		})
		return nil
	})
	if _, err := s.StartAndCreateSession(); err != nil {
		t.Fatalf("StartAndCreateSession: %v", err)
	}

	mu.Lock()
	got := strings.Join(record, ",")
	mu.Unlock()
	if got != "one,two" {
		t.Fatalf("hooks ran as %q, want \"one,two\" (registration order, R253)", got)
	}
	luaValue(t, s, `hk = {} ui.hookOp(hk, "out", function() hk.done = true end) return "started"`)
	waitLua(t, s, "hk.done", "the hook-defined host operation's callback")
	if out := luaValue(t, s, "return tostring(hk.out)"); !strings.Contains(out, "from-hook") {
		t.Fatalf("ui.hookOp delivered %s, want from-hook (R254)", out)
	}
}

// Test: hooks run again when reconfiguring replaces the session (R253)
func TestSessionSetupHookRunsAfterReconfigure(t *testing.T) {
	s, baseDir := configuredServer(t)
	runs := 0
	s.AddSessionSetupHook(func(rig cli.LuaRig) error {
		runs++
		rig.Session().State.SetGlobal("hookRuns", lua.LNumber(runs))
		return nil
	})
	if _, err := s.StartAndCreateSession(); err != nil {
		t.Fatalf("StartAndCreateSession: %v", err)
	}
	// Replace the session the way handleConfigure does.
	if err := s.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if err := s.Configure(baseDir); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	if _, err := s.StartAndCreateSession(); err != nil {
		t.Fatalf("StartAndCreateSession again: %v", err)
	}
	if runs != 2 {
		t.Fatalf("the hook ran %d times over two sessions, want 2 (R253)", runs)
	}
	if got := luaValue(t, s, "return tostring(hookRuns)"); !strings.Contains(got, "2") {
		t.Fatalf("the new session saw hookRuns = %s, want 2", got)
	}
}
