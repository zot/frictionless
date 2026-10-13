// CRC: crc-FlibRuntime.md | Test: test-AppActivation.md
package flib

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
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
