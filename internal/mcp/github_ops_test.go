// CRC: crc-GitHubOps.md | Test: test-GitHubInstall.md
package mcp

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/frictionless/githubinstall"
)

// Test: the host operations deliver into their fields (R239, R245)
func TestGitHubOpsDeliver(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("r-main/apps/demo/app.lua")
	w.Write([]byte("-- demo"))
	zw.Close()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/repos/u/r/contents/apps/demo":
			w.Write([]byte(`[{"name":"app.lua","type":"file"}]`))
		case "/raw/u/r/main/apps/demo/app.lua":
			w.Write([]byte("-- demo"))
		case "/archive/u/r/archive/refs/heads/main.zip":
			w.Write(buf.Bytes())
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	api, raw, arc := githubinstall.APIBase, githubinstall.RawBase, githubinstall.ArchiveBase
	githubinstall.APIBase, githubinstall.RawBase, githubinstall.ArchiveBase = srv.URL+"/api", srv.URL+"/raw", srv.URL+"/archive"
	t.Cleanup(func() { githubinstall.APIBase, githubinstall.RawBase, githubinstall.ArchiveBase = api, raw, arc })

	var base string
	s, cleanup := createTestServerWithSessionSetup(t, func(baseDir string) {
		base = baseDir
		os.MkdirAll(filepath.Join(baseDir, "apps"), 0o755)
	})
	t.Cleanup(cleanup)

	luaValue(t, s, `gh = {} ui.githubInvestigate(gh, "inv", "https://github.com/u/r/tree/main/apps/demo", function() gh.invDone = true end) return "started"`)
	waitLua(t, s, "gh.invDone", "the investigate callback")
	if got := luaValue(t, s, `return tostring(gh.inv and gh.inv.contents["app.lua"])`); !strings.Contains(got, "-- demo") {
		t.Fatalf(`inv.contents["app.lua"] = %s (error %s), want the served text`, got, luaValue(t, s, "return tostring(gh.invError)"))
	}

	luaValue(t, s, `ui.githubInstall(gh, "ins", "https://github.com/u/r/tree/main/apps/demo", "demo", function() gh.insDone = true end) return "started"`)
	waitLua(t, s, "gh.insDone", "the install callback")
	if got := luaValue(t, s, "return tostring(gh.ins)"); !strings.Contains(got, "true") {
		t.Fatalf("ins = %s (error %s), want true", got, luaValue(t, s, "return tostring(gh.insError)"))
	}
	if _, err := os.Stat(filepath.Join(base, "apps", "demo", "app.lua")); err != nil {
		t.Errorf("apps/demo/app.lua was not installed: %v", err)
	}
}
