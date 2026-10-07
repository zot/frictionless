// CRC: crc-FlibRuntime.md | Test: test-AppActivation.md
package flib

import (
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
