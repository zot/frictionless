// Tests for app activation (test-AppLink.md). Each runs over a temporary base
// directory; the Windows junction cannot run here and is listed as untested.
// CRC: crc-AppLink.md | Test: test-AppLink.md | R184, R185, R186, R187, R188, R226, R228, R229, R230, R231, R232, R233
package applink

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// newBase makes a base directory holding apps/<name> for each name.
func newBase(t *testing.T, names ...string) string {
	t.Helper()
	base := filepath.Join(t.TempDir(), "base")
	for _, n := range names {
		if err := os.MkdirAll(filepath.Join(base, appsDir, n, viewdefsDir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(base, appsDir), 0o755); err != nil {
		t.Fatal(err)
	}
	return base
}

func symlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

// Test: activate makes one relative link and cleans up the old scheme (R184, R187, R188, R226)
func TestActivateMakesOneLinkAndCleansOldScheme(t *testing.T) {
	base := newBase(t, "demo")
	symlink(t, "../apps/demo/app.lua", filepath.Join(base, luaDir, "demo.lua"))
	symlink(t, "../apps/demo/viewdefs/Demo.DEFAULT.html", filepath.Join(base, viewdefsDir, "Demo.DEFAULT.html"))
	if err := Activate(base, "demo"); err != nil {
		t.Fatal(err)
	}
	if target, err := os.Readlink(filepath.Join(base, luaDir, "demo")); err != nil || target != "../apps/demo" {
		t.Fatalf("lua/demo target = %q, %v; want ../apps/demo", target, err)
	}
	if exists(filepath.Join(base, luaDir, "demo.lua")) {
		t.Error("old-scheme link lua/demo.lua survived Activate (R188)")
	}
	if exists(filepath.Join(base, viewdefsDir, "Demo.DEFAULT.html")) {
		t.Error("old-scheme viewdefs link survived Activate (R188)")
	}
}

// Test: the Unix link target is relative (R226)
func TestLinkSurvivesMovedBase(t *testing.T) {
	base := newBase(t, "demo")
	if err := Activate(base, "demo"); err != nil {
		t.Fatal(err)
	}
	moved := base + "-moved"
	if err := os.Rename(base, moved); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Stat(filepath.Join(moved, luaDir, "demo")); err != nil || !fi.IsDir() {
		t.Fatalf("after moving the base, lua/demo no longer resolves: %v (R226)", err)
	}
}

// Test: activate is idempotent and replaces a foreign link (R230)
func TestActivateIdempotentAndReplacesForeignLink(t *testing.T) {
	base := newBase(t, "demo", "other")
	for i := 0; i < 2; i++ {
		if err := Activate(base, "demo"); err != nil {
			t.Fatalf("activate #%d: %v", i+1, err)
		}
	}
	link := filepath.Join(base, luaDir, "demo")
	os.Remove(link)
	symlink(t, "../apps/other", link)
	if err := Activate(base, "demo"); err != nil {
		t.Fatal(err)
	}
	if target, _ := os.Readlink(link); target != "../apps/demo" {
		t.Fatalf("foreign link not replaced: lua/demo -> %q (R230)", target)
	}
}

// Test: activating a missing app is an error (R230)
func TestActivateMissingApp(t *testing.T) {
	base := newBase(t)
	err := Activate(base, "ghost")
	if err == nil || !strings.Contains(err.Error(), filepath.Join(appsDir, "ghost")) {
		t.Fatalf("Activate ghost: %v, want an error naming apps/ghost", err)
	}
	if exists(filepath.Join(base, luaDir, "ghost")) {
		t.Fatal("Activate made a link for a missing app")
	}
}

// Test: deactivate refuses a real directory (R185, R231)
func TestDeactivateRefusesRealDirectory(t *testing.T) {
	base := newBase(t, "demo")
	real := filepath.Join(base, luaDir, "demo")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Deactivate(base, "demo"); err == nil {
		t.Fatal("Deactivate removed a real directory without error (R231)")
	}
	if !exists(real) {
		t.Fatal("Deactivate removed the real directory at lua/demo (R231)")
	}
	os.Remove(real)
	if err := Activate(base, "demo"); err != nil {
		t.Fatal(err)
	}
	if err := Deactivate(base, "demo"); err != nil || exists(real) {
		t.Fatalf("Deactivate of an active app: %v, link left: %v", err, exists(real))
	}
	if err := Deactivate(base, "demo"); err != nil {
		t.Fatalf("Deactivate of an inactive app: %v, want nothing", err)
	}
}

// Test: list reports only directory links (R186)
func TestListReportsOnlyDirectoryLinks(t *testing.T) {
	base := newBase(t, "demo")
	if err := Activate(base, "demo"); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(base, luaDir, "main.lua"), []byte("--\n"), 0o644)
	os.MkdirAll(filepath.Join(base, luaDir, "plain"), 0o755)
	names, err := List(base)
	if err != nil || len(names) != 1 || names[0] != "demo" {
		t.Fatalf("List = %v, %v; want [demo] (R186)", names, err)
	}
}

// Test: delete guards its name and its target (R232)
func TestDeleteGuards(t *testing.T) {
	base := newBase(t, "demo")
	victim := filepath.Join(base, "victim")
	os.MkdirAll(victim, 0o755)
	for _, bad := range []string{"", "a/b", "..", "../victim"} {
		if err := Delete(base, bad); err == nil {
			t.Errorf("Delete(%q) succeeded, want an error (R232)", bad)
		}
	}
	if !exists(victim) {
		t.Fatal("a bad name reached outside apps/ and removed victim (R232)")
	}
	target := filepath.Join(t.TempDir(), "kept")
	os.MkdirAll(target, 0o755)
	symlink(t, target, filepath.Join(base, appsDir, "linked"))
	if err := Delete(base, "linked"); err == nil {
		t.Error("Delete of a link under apps/ succeeded, want an error")
	}
	if !exists(target) {
		t.Fatal("Delete followed a link and removed its target")
	}
	if err := Activate(base, "demo"); err != nil {
		t.Fatal(err)
	}
	if err := Delete(base, "demo"); err != nil {
		t.Fatal(err)
	}
	if exists(filepath.Join(base, appsDir, "demo")) || exists(filepath.Join(base, luaDir, "demo")) {
		t.Fatal("Delete left the app directory or its link")
	}
}

// Test: repair remakes only stale links (R228)
func TestRepairRemakesOnlyStaleLinks(t *testing.T) {
	base := newBase(t, "demo", "idle")
	symlink(t, "/nowhere/apps/demo", filepath.Join(base, luaDir, "demo"))
	symlink(t, "/nowhere/apps/gone", filepath.Join(base, luaDir, "gone"))
	var logged []string
	Repair(base, func(format string, args ...any) { logged = append(logged, format) })
	if fi, err := os.Stat(filepath.Join(base, luaDir, "demo")); err != nil || !fi.IsDir() {
		t.Errorf("stale lua/demo was not remade: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, luaDir, "gone")); err == nil {
		t.Error("lua/gone resolves, but its app does not exist")
	}
	if len(logged) != 2 {
		t.Errorf("Repair logged %d lines, want 2 (one repair, one gone)", len(logged))
	}
	if exists(filepath.Join(base, luaDir, "idle")) {
		t.Fatal("Repair created a link for an app that was not active (R228)")
	}
}

// fakeInfo is a FileInfo with a chosen mode.
type fakeInfo struct{ mode fs.FileMode }

func (f fakeInfo) Name() string       { return "x" }
func (f fakeInfo) Size() int64        { return 0 }
func (f fakeInfo) Mode() fs.FileMode  { return f.mode }
func (f fakeInfo) ModTime() time.Time { return time.Time{} }
func (f fakeInfo) IsDir() bool        { return f.mode.IsDir() }
func (f fakeInfo) Sys() any           { return nil }

// Test: a junction counts as a link (R229)
func TestIsLinkCountsJunctions(t *testing.T) {
	for _, c := range []struct {
		mode fs.FileMode
		want bool
	}{{fs.ModeSymlink, true}, {fs.ModeIrregular, true}, {fs.ModeDir, false}, {0, false}} {
		if got := isLink(fakeInfo{c.mode}); got != c.want {
			t.Errorf("isLink(%v) = %v, want %v (R229)", c.mode, got, c.want)
		}
	}
}

// Test: the command runs each verb (R233)
func TestCommand(t *testing.T) {
	base := newBase(t, "demo")
	var out bytes.Buffer
	if err := Command([]string{"add", "demo"}, base, &out); err != nil {
		t.Fatal(err)
	}
	if !exists(filepath.Join(base, luaDir, "demo")) {
		t.Fatal("linkapp add made no link")
	}
	out.Reset()
	if err := Command([]string{"list"}, base, &out); err != nil || !strings.Contains(out.String(), "demo") {
		t.Fatalf("linkapp list: %q, %v", out.String(), err)
	}
	if err := Command([]string{"remove", "demo"}, base, &out); err != nil || exists(filepath.Join(base, luaDir, "demo")) {
		t.Fatalf("linkapp remove: %v", err)
	}
	if err := Command(nil, base, &out); err == nil {
		t.Fatal("linkapp with no arguments succeeded, want a usage error")
	}
}
