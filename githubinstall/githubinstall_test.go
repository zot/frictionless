// CRC: crc-GitHubInstall.md | Test: test-GitHubInstall.md
package githubinstall

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

// fakeGitHub serves listings (keyed by directory path), raw files (keyed by
// path from the repo root) and one branch archive, and points the package's
// endpoints at itself for the test's duration. It returns the count of
// archive requests.
func fakeGitHub(t *testing.T, listings map[string][]Entry, files map[string]string, archive []byte) *atomic.Int32 {
	t.Helper()
	var archiveHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		switch {
		case strings.HasPrefix(p, "/api/repos/u/r/contents/"):
			entries, ok := listings[strings.TrimPrefix(p, "/api/repos/u/r/contents/")]
			if !ok || r.URL.Query().Get("ref") != "main" {
				http.NotFound(w, r)
				return
			}
			json.NewEncoder(w).Encode(entries)
		case strings.HasPrefix(p, "/raw/u/r/main/"):
			text, ok := files[strings.TrimPrefix(p, "/raw/u/r/main/")]
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Write([]byte(text))
		case p == "/archive/u/r/archive/refs/heads/main.zip":
			archiveHits.Add(1)
			w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	api, raw, arc := APIBase, RawBase, ArchiveBase
	APIBase, RawBase, ArchiveBase = srv.URL+"/api", srv.URL+"/raw", srv.URL+"/archive"
	t.Cleanup(func() { APIBase, RawBase, ArchiveBase = api, raw, arc })
	return &archiveHits
}

// zipOf builds an archive from name -> content, in the given order.
func zipOf(t *testing.T, names []string, content map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, n := range names {
		w, err := zw.Create(n)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(content[n]))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

var demoSrc = Source{User: "u", Repo: "r", Branch: "main", Path: "apps/demo", Name: "demo"}

// appsDirIn makes base/apps and returns it.
func appsDirIn(t *testing.T, base string) string {
	t.Helper()
	apps := filepath.Join(base, "apps")
	if err := os.MkdirAll(apps, 0o755); err != nil {
		t.Fatal(err)
	}
	return apps
}

// onlyEntries fails unless dir holds exactly the named entries.
func onlyEntries(t *testing.T, dir string, want ...string) {
	t.Helper()
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range ents {
		got = append(got, e.Name())
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("%s holds %v, want %v", dir, got, want)
	}
}

// Test: a tree URL parses, and nothing else does (R236)
func TestParseURL(t *testing.T) {
	src, err := ParseURL("https://github.com/u/r/tree/main/apps/demo")
	if err != nil {
		t.Fatal(err)
	}
	if src != demoSrc {
		t.Errorf("ParseURL = %+v, want %+v", src, demoSrc)
	}
	for _, bad := range []string{
		"https://github.com/u/r/blob/main/apps/demo",
		"https://github.com/u/r/tree/main",
		"https://example.com/u/r/tree/main/apps/demo",
	} {
		if _, err := ParseURL(bad); err == nil {
			t.Errorf("ParseURL(%q) succeeded, want an error", bad)
		}
	}
}

// Test: investigate lists and fetches what the review shows (R237, R238)
func TestInvestigateFetchesReviewable(t *testing.T) {
	appFiles := []Entry{
		{"requirements.md", "file"}, {"design.md", "file"}, {"app.lua", "file"},
		{"extra.lua", "file"}, {"README.md", "file"}, {"viewdefs", "dir"},
	}
	viewdefs := []Entry{{"App.DEFAULT.html", "file"}, {"App.Item.list-item.html", "file"}}
	files := map[string]string{}
	for _, n := range []string{"requirements.md", "design.md", "app.lua", "extra.lua", "README.md", "viewdefs/App.DEFAULT.html", "viewdefs/App.Item.list-item.html"} {
		files["apps/demo/"+n] = "text of " + n
	}
	fakeGitHub(t, map[string][]Entry{"apps/demo": appFiles, "apps/demo/viewdefs": viewdefs}, files, nil)

	inv, err := Investigate(context.Background(), demoSrc)
	if err != nil {
		t.Fatal(err)
	}
	if len(inv.Files) != len(appFiles) || len(inv.Viewdefs) != 2 {
		t.Errorf("listed %d files and %d viewdefs, want %d and 2", len(inv.Files), len(inv.Viewdefs), len(appFiles))
	}
	for _, rel := range []string{"requirements.md", "design.md", "app.lua", "extra.lua", "viewdefs/App.DEFAULT.html", "viewdefs/App.Item.list-item.html"} {
		if inv.Contents[rel] != "text of "+rel {
			t.Errorf("Contents[%q] = %q, want its text", rel, inv.Contents[rel])
		}
	}
	if _, ok := inv.Contents["README.md"]; ok {
		t.Error("Contents holds README.md, which the review does not show")
	}
}

// Test: a missing file is absent, a missing listing is an error (R239)
func TestInvestigateMissing(t *testing.T) {
	fakeGitHub(t,
		map[string][]Entry{"apps/demo": {{"app.lua", "file"}, {"design.md", "file"}}},
		map[string]string{"apps/demo/app.lua": "lua"}, nil)

	inv, err := Investigate(context.Background(), demoSrc)
	if err != nil {
		t.Fatalf("an unfetchable file failed the investigation: %v", err)
	}
	if _, ok := inv.Contents["design.md"]; ok {
		t.Error("Contents holds design.md, which answered 404")
	}
	if inv.Contents["app.lua"] != "lua" {
		t.Errorf("Contents[app.lua] = %q, want lua", inv.Contents["app.lua"])
	}

	missing := demoSrc
	missing.Path, missing.Name = "apps/nothing", "nothing"
	if _, err := Investigate(context.Background(), missing); err == nil {
		t.Error("an unfetchable listing succeeded, want an error")
	}
}

// Test: install extracts only the app, with its prefix removed (R241, R244)
func TestInstallExtractsOnlyTheApp(t *testing.T) {
	names := []string{"r-main/apps/demo/app.lua", "r-main/apps/demo/viewdefs/D.html", "r-main/README.md", "r-main/apps/other/app.lua"}
	content := map[string]string{
		"r-main/apps/demo/app.lua": "demo lua", "r-main/apps/demo/viewdefs/D.html": "<template/>",
		"r-main/README.md": "readme", "r-main/apps/other/app.lua": "other lua",
	}
	fakeGitHub(t, nil, nil, zipOf(t, names, content))
	apps := appsDirIn(t, t.TempDir())

	if err := Install(context.Background(), demoSrc, apps, "demo"); err != nil {
		t.Fatal(err)
	}
	for rel, want := range map[string]string{"app.lua": "demo lua", "viewdefs/D.html": "<template/>"} {
		got, err := os.ReadFile(filepath.Join(apps, "demo", rel))
		if err != nil || string(got) != want {
			t.Errorf("apps/demo/%s = %q, %v; want %q", rel, got, err, want)
		}
	}
	onlyEntries(t, apps, "demo")
	onlyEntries(t, filepath.Join(apps, "demo"), "app.lua", "viewdefs")
	if fi, err := os.Stat(filepath.Join(apps, "demo")); err != nil || fi.Mode().Perm() != 0o755 {
		t.Errorf("apps/demo mode = %v, %v; want 0755 like any app directory", fi.Mode().Perm(), err)
	}
}

// Test: an existing app is never overwritten (R240)
func TestInstallRefusesExisting(t *testing.T) {
	hits := fakeGitHub(t, nil, nil, zipOf(t, []string{"r-main/apps/demo/app.lua"}, map[string]string{"r-main/apps/demo/app.lua": "new"}))
	apps := appsDirIn(t, t.TempDir())
	keep := filepath.Join(apps, "demo", "keep.txt")
	os.MkdirAll(filepath.Dir(keep), 0o755)
	os.WriteFile(keep, []byte("mine"), 0o644)

	if err := Install(context.Background(), demoSrc, apps, "demo"); err == nil {
		t.Fatal("installing over an existing app succeeded, want an error")
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("the archive was requested %d times, want 0: the existence check must come first", n)
	}
	if got, _ := os.ReadFile(keep); string(got) != "mine" {
		t.Errorf("keep.txt = %q, want it unchanged", got)
	}
	onlyEntries(t, apps, "demo")
}

// Test: an entry escaping the app fails the install (R242, R244)
func TestInstallRefusesZipSlip(t *testing.T) {
	names := []string{"r-main/apps/demo/app.lua", "r-main/apps/demo/../../escaped.txt"}
	fakeGitHub(t, nil, nil, zipOf(t, names, map[string]string{names[0]: "lua", names[1]: "gotcha"}))
	base := t.TempDir()
	apps := appsDirIn(t, base)

	if err := Install(context.Background(), demoSrc, apps, "demo"); err == nil {
		t.Fatal("an escaping entry installed, want an error")
	}
	filepath.WalkDir(base, func(p string, d os.DirEntry, err error) error {
		if d != nil && d.Name() == "escaped.txt" {
			t.Errorf("escaped.txt was written at %s", p)
		}
		return nil
	})
	onlyEntries(t, apps)
}

// Test: an oversize archive is refused (R243)
func TestInstallRefusesOversize(t *testing.T) {
	old := MaxBytes
	MaxBytes = 1024
	t.Cleanup(func() { MaxBytes = old })
	name := "r-main/apps/demo/app.lua"
	fakeGitHub(t, nil, nil, zipOf(t, []string{name}, map[string]string{name: strings.Repeat("x", 4096)}))
	apps := appsDirIn(t, t.TempDir())

	if err := Install(context.Background(), demoSrc, apps, "demo"); err == nil {
		t.Fatal("a 4 KB app installed under a 1 KB cap, want an error")
	}
	onlyEntries(t, apps)
}

// The R248 name guard: a name that would land outside apps/ changes nothing.
func TestInstallRefusesBadName(t *testing.T) {
	hits := fakeGitHub(t, nil, nil, nil)
	apps := appsDirIn(t, t.TempDir())
	for _, name := range []string{"", ".", "..", "a/b", `a\b`, "x..y"} {
		if err := Install(context.Background(), demoSrc, apps, name); err == nil {
			t.Errorf("Install as %q succeeded, want an error (R248)", name)
		}
	}
	if hits.Load() != 0 {
		t.Error("a bad name reached the network")
	}
	onlyEntries(t, apps)
}
