// CRC: crc-AppLinkOps.md | Test: test-AppLink.md | R234
package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test: the Lua bindings succeed quietly and raise on error (R234)
func TestAppLinkBindings(t *testing.T) {
	var base string
	s, cleanup := createTestServerWithSessionSetup(t, func(baseDir string) {
		base = baseDir
		if err := os.MkdirAll(filepath.Join(baseDir, "apps", "demo"), 0o755); err != nil {
			t.Fatal(err)
		}
	})
	t.Cleanup(cleanup)

	if got := luaValue(t, s, `return select('#', ui.linkApp("demo"))`); got != "0" {
		t.Errorf("ui.linkApp returned %s values, want 0", got)
	}
	if _, err := os.Stat(filepath.Join(base, "lua", "demo")); err != nil {
		t.Fatalf("ui.linkApp made no link: %v", err)
	}
	got := luaValue(t, s, `local ok, err = pcall(ui.deleteApp, "../x"); return tostring(ok) .. "|" .. tostring(err)`)
	if !strings.HasPrefix(strings.Trim(got, `"`), "false|") || !strings.Contains(got, "../x") {
		t.Fatalf("pcall(ui.deleteApp, \"../x\") = %s, want false with an error naming the app (R234)", got)
	}
}
