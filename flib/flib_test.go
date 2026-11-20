// CRC: crc-FlibRuntime.md | Test: test-AppActivation.md
package flib

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zot/ui-engine/cli"
)

// Test: BuiltinApps reaches the ui-engine config (R190)
func TestBuiltinAppsReachUIEngine(t *testing.T) {
	want := []string{"ark", "mcp"}
	rt, err := New(Config{Dir: t.TempDir(), BuiltinApps: want})
	if err != nil {
		t.Fatal(err)
	}
	if got := rt.Cfg.Server.BuiltinApps; !reflect.DeepEqual(got, want) {
		t.Errorf("ui-engine BuiltinApps = %v, want %v", got, want)
	}
}

// Test: ViewdefCollisions passes the ui-engine's live list through (R191)
func TestViewdefCollisionsPassThrough(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "lua"), 0o755)
	for _, name := range []string{"aa", "zz"} {
		vd := filepath.Join(dir, "apps", name, "viewdefs")
		os.MkdirAll(vd, 0o755)
		os.WriteFile(filepath.Join(vd, "T.DEFAULT.html"), []byte("<template>"+name+"</template>"), 0o644)
		if err := os.Symlink(filepath.Join("..", "apps", name), filepath.Join(dir, "lua", name)); err != nil {
			t.Skipf("cannot create symlinks: %v", err)
		}
	}
	rt, err := New(Config{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	got := rt.ViewdefCollisions()
	if len(got) != 1 || got[0].Key != "T.DEFAULT" || got[0].Winner != "aa" || got[0].Loser != "zz" {
		t.Errorf("ViewdefCollisions() = %+v, want T.DEFAULT, aa over zz", got)
	}
}

// Test: WithLuaUpdate reports a missing session the way WithLua does (R200)
func TestWithLuaUpdateNoSession(t *testing.T) {
	rt, err := New(Config{Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	called := false
	err = rt.WithLuaUpdate(func(*cli.LuaRuntime) error {
		called = true
		return nil
	})
	if err == nil || err.Error() != errNoActiveSession {
		t.Errorf("WithLuaUpdate() error = %v, want %q", err, errNoActiveSession)
	}
	if called {
		t.Error("WithLuaUpdate ran the closure with no session")
	}
}

// Test: WithLuaUpdate pushes the closure's changes to a connected browser (R200)
func TestWithLuaUpdatePushesToBrowser(t *testing.T) {
	const (
		probeVarID = 2
		pushed     = "pushed-by-update"
		pushWait   = 2 * time.Second
	)
	dir, err := os.MkdirTemp("", "flib")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.MkdirAll(filepath.Join(dir, "lua"), 0o755)
	os.MkdirAll(filepath.Join(dir, "log"), 0o755)
	// A README.md skips auto-install, which needs the bundled binary.
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test\n**Version: 99.0.0**\n"), 0o644)

	rt, err := New(Config{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if err := rt.Configure(); err != nil {
		t.Fatal(err)
	}
	base, err := rt.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Shutdown(context.Background())

	// Connect as the browser and watch mcp.probe (variable 1 is the mcp table).
	wsURL := "ws" + strings.TrimPrefix(base, "http") + "/ws/" + rt.mcpServer.GetCurrentSessionID()
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", wsURL, err)
	}
	defer conn.Close()
	msgs := readAll(conn)
	create := fmt.Sprintf(`{"type":"create","data":{"id":%d,"parentId":1,"properties":{"path":"probe"}}}`, probeVarID)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(create)); err != nil {
		t.Fatal(err)
	}
	// Let the create's own batch go out before the change under test.
	drain(msgs, 300*time.Millisecond)

	err = rt.WithLuaUpdate(func(L *cli.LuaRuntime) error {
		return L.State.DoString(fmt.Sprintf("mcp.probe = %q", pushed))
	})
	if err != nil {
		t.Fatal(err)
	}
	// No further browser message is sent, so only an active push delivers it.
	if got, ok := waitFor(msgs, pushed, pushWait); !ok {
		t.Errorf("browser received %q after WithLuaUpdate, want an update carrying %q", got, pushed)
	}
}

// readAll forwards every websocket message to the returned channel until
// the connection closes. One reader for the connection's lifetime, since a
// read deadline that expires leaves a gorilla connection unusable.
func readAll(conn *websocket.Conn) <-chan string {
	ch := make(chan string, 64)
	go func() {
		defer close(ch)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			ch <- string(msg)
		}
	}()
	return ch
}

// drain collects messages from ch until none arrives within wait.
func drain(ch <-chan string, wait time.Duration) string {
	var sb strings.Builder
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return sb.String()
			}
			sb.WriteString(msg)
		case <-time.After(wait):
			return sb.String()
		}
	}
}

// waitFor collects messages from ch until one carries want or wait passes.
func waitFor(ch <-chan string, want string, wait time.Duration) (string, bool) {
	var sb strings.Builder
	deadline := time.After(wait)
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return sb.String(), false
			}
			sb.WriteString(msg)
			if strings.Contains(msg, want) {
				return sb.String(), true
			}
		case <-deadline:
			return sb.String(), false
		}
	}
}

// Test: WithLuaUpdate works with dangerous execution disallowed (R250)
func TestWithLuaUpdateMakesNoDangerousCall(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "lua"), 0o755)
	os.MkdirAll(filepath.Join(dir, "log"), 0o755)
	// A README.md skips auto-install, which needs the bundled binary.
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test\n**Version: 99.0.0**\n"), 0o644)
	rt, err := New(Config{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if err := rt.Configure(); err != nil {
		t.Fatal(err)
	}
	if _, err := rt.Start(); err != nil {
		t.Fatal(err)
	}
	defer rt.Shutdown(context.Background())
	rt.Cfg.Session.AllowDangerousExecution = false // ui-engine's coming default: a dangerous call panics

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("WithLuaUpdate made a dangerous call: %v (R250)", r)
			}
		}()
		if err := rt.WithLuaUpdate(func(L *cli.LuaRuntime) error {
			return L.State.DoString("reviewedProbe = true")
		}); err != nil {
			t.Fatalf("WithLuaUpdate: %v", err)
		}
	}()
	var got string
	rt.WithLua(func(L *cli.LuaRuntime) error {
		got = L.State.GetGlobal("reviewedProbe").String()
		return nil
	})
	if got != "true" {
		t.Errorf("reviewedProbe = %s after WithLuaUpdate, want true", got)
	}
}

// startedRuntime is a configured and started flib runtime in a temp dir.
func startedRuntime(t *testing.T) *Runtime {
	t.Helper()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "lua"), 0o755)
	os.MkdirAll(filepath.Join(dir, "log"), 0o755)
	// A README.md skips auto-install, which needs the bundled binary.
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test\n**Version: 99.0.0**\n"), 0o644)
	rt, err := New(Config{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if err := rt.Configure(); err != nil {
		t.Fatal(err)
	}
	if _, err := rt.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { rt.Shutdown(context.Background()) })
	return rt
}

// Test: a hook registered after Start runs for the existing session (R255)
func TestOnSessionSetupAfterStartRunsNow(t *testing.T) {
	rt := startedRuntime(t)
	if err := rt.OnSessionSetup(func(rig cli.LuaRig) error {
		return rig.Session().State.DoString("lateHook = true")
	}); err != nil {
		t.Fatalf("OnSessionSetup: %v", err)
	}
	var got string
	rt.WithLua(func(L *cli.LuaRuntime) error {
		got = L.State.GetGlobal("lateHook").String()
		return nil
	})
	if got != "true" {
		t.Errorf("lateHook = %s after OnSessionSetup, want true: a hook registered after Start must run for the existing session (R255)", got)
	}
}

// Test: a hook's error reaches the caller (R256)
func TestOnSessionSetupReturnsHookError(t *testing.T) {
	rt := startedRuntime(t)
	err := rt.OnSessionSetup(func(cli.LuaRig) error { return errors.New("hook says no") })
	if err == nil || !strings.Contains(err.Error(), "hook says no") {
		t.Fatalf("OnSessionSetup = %v, want the hook's error (R256)", err)
	}
}
