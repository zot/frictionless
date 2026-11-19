// Package applink activates and deactivates Frictionless apps. An app is active
// when {base_dir}/lua/{app} links to its directory {base_dir}/apps/{app}; that
// one link is the whole on/off bit (specs/applink.md).
//
// The link code is lifted from ark's dirlink.go, which solved Windows first:
// a relative symbolic link on Unix, a directory junction on Windows.
// CRC: crc-AppLink.md
package applink

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	appsDir     = "apps"
	luaDir      = "lua"
	viewdefsDir = "viewdefs" // only for old-scheme cleanup
)

// Activate links lua/{app} to apps/{app}. Activating an app already linked to
// its own directory changes nothing; a link there pointing elsewhere is
// replaced. A real directory at lua/{app} is never touched.
// CRC: crc-AppLink.md | Seq: seq-applink.md#1.1 | R184, R187, R188, R230
func Activate(baseDir, app string) error {
	if err := validName(app); err != nil {
		return err
	}
	appDir := filepath.Join(baseDir, appsDir, app)
	if fi, err := os.Stat(appDir); err != nil || !fi.IsDir() {
		return fmt.Errorf("app %q not found: %s is not a directory", app, appDir)
	}
	if err := os.MkdirAll(filepath.Join(baseDir, luaDir), 0755); err != nil {
		return err
	}
	link := filepath.Join(baseDir, luaDir, app)
	if fi, err := os.Lstat(link); err == nil {
		if !isLink(fi) {
			return fmt.Errorf("refusing to replace %s: it is not a link", link)
		}
		if !isOurLink(link, appDir) {
			if err := os.Remove(link); err != nil {
				return err
			}
		}
	}
	if !isOurLink(link, appDir) {
		if err := makeLink(link, appDir, app); err != nil {
			return err
		}
	}
	return removeOldLinks(baseDir, app)
}

// Deactivate removes the lua/{app} link. A real directory there is someone's
// data and is refused; an app that is not active is left as it is.
// CRC: crc-AppLink.md | Seq: seq-applink.md#2.1 | R185, R188, R231
func Deactivate(baseDir, app string) error {
	if err := validName(app); err != nil {
		return err
	}
	link := filepath.Join(baseDir, luaDir, app)
	fi, err := os.Lstat(link)
	switch {
	case os.IsNotExist(err):
	case err != nil:
		return err
	case !isLink(fi):
		return fmt.Errorf("refusing to remove %s: it is not a link", link)
	default:
		if err := os.Remove(link); err != nil {
			return err
		}
	}
	return removeOldLinks(baseDir, app)
}

// List reports the active apps: the entries in lua/ that are links to
// directories, sorted by name.
// CRC: crc-AppLink.md | R186
func List(baseDir string) ([]string, error) {
	dir := filepath.Join(baseDir, luaDir)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		fi, err := os.Lstat(path)
		if err != nil || !isLink(fi) {
			continue
		}
		if target, err := os.Stat(path); err == nil && target.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// Delete deactivates an app and removes its directory. Only a plainly named,
// real directory under apps/ is removed.
// CRC: crc-AppLink.md | Seq: seq-applink.md#2.2 | R232
func Delete(baseDir, app string) error {
	if err := validName(app); err != nil {
		return err
	}
	if err := Deactivate(baseDir, app); err != nil {
		return err
	}
	appDir := filepath.Join(baseDir, appsDir, app)
	fi, err := os.Lstat(appDir)
	if err != nil {
		return fmt.Errorf("app %q not found: %w", app, err)
	}
	if isLink(fi) || !fi.IsDir() {
		return fmt.Errorf("refusing to delete %s: it is not a real directory", appDir)
	}
	return os.RemoveAll(appDir)
}

// validName accepts a plain app name: not empty, no path separator, no "..".
// CRC: crc-AppLink.md | Seq: seq-applink.md#2.2.1 | R232
func validName(app string) error {
	if app == "" || app == "." || strings.ContainsAny(app, `/\`) || strings.Contains(app, "..") {
		return fmt.Errorf("invalid app name %q", app)
	}
	return nil
}

// Repair remakes each activation link that no longer resolves, when its app
// directory exists, and reports the rest. It only walks links that exist, so it
// never creates one: a deactivated app stays deactivated. A junction stores an
// absolute target, so on Windows a moved base directory needs this.
// CRC: crc-AppLink.md | Seq: seq-applink.md#3.1 | R228
func Repair(baseDir string, logf func(format string, args ...any)) {
	dir := filepath.Join(baseDir, luaDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		link := filepath.Join(dir, e.Name())
		fi, err := os.Lstat(link)
		if err != nil || !isLink(fi) {
			continue
		}
		if _, err := os.Stat(link); err == nil {
			continue // resolves
		}
		app := e.Name()
		appDir := filepath.Join(baseDir, appsDir, app)
		if st, err := os.Stat(appDir); err != nil || !st.IsDir() {
			logf("activation link %s: app %s is gone; left as it is", link, app)
			continue
		}
		if err := os.Remove(link); err == nil {
			err = makeLink(link, appDir, app)
			if err != nil {
				logf("activation link %s: could not repair: %v", link, err)
				continue
			}
		}
		logf("activation link %s repaired", link)
	}
}

// isLink reports whether fi is a symbolic link or, on Windows, a junction,
// which Lstat reports as ModeIrregular.
// CRC: crc-AppLink.md | R229
func isLink(fi os.FileInfo) bool {
	return fi.Mode()&(os.ModeSymlink|os.ModeIrregular) != 0
}

// isOurLink reports whether link resolves to appDir. Both sides are resolved,
// so a relative Unix target and an absolute junction compare alike.
// CRC: crc-AppLink.md
func isOurLink(link, appDir string) bool {
	got, err := filepath.EvalSymlinks(link)
	if err != nil {
		return false
	}
	want, err := filepath.EvalSymlinks(appDir)
	return err == nil && got == want
}

// removeOldLinks removes links from the old per-file scheme: lua/{app}.lua, and
// each viewdefs/ link into the app's viewdefs/.
// CRC: crc-AppLink.md | R188
func removeOldLinks(baseDir, app string) error {
	old := filepath.Join(baseDir, luaDir, app+".lua")
	if fi, err := os.Lstat(old); err == nil && isLink(fi) {
		if err := os.Remove(old); err != nil {
			return err
		}
	}
	dir := filepath.Join(baseDir, viewdefsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil // no viewdefs/ is the normal case now
	}
	into := appsDir + "/" + app + "/" + viewdefsDir + "/"
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		fi, err := os.Lstat(path)
		if err != nil || !isLink(fi) {
			continue
		}
		target, err := os.Readlink(path)
		if err == nil && strings.Contains(filepath.ToSlash(target), into) {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
	}
	return nil
}

// Command runs `linkapp add|remove|list [APP]` against baseDir.
// CRC: crc-AppLink.md | R233
func Command(args []string, baseDir string, out io.Writer) error {
	usage := errors.New("usage: linkapp add|remove|list [APP]")
	if len(args) == 0 {
		return usage
	}
	switch args[0] {
	case "add", "remove":
		if len(args) != 2 {
			return usage
		}
		if args[0] == "add" {
			if err := Activate(baseDir, args[1]); err != nil {
				return err
			}
			fmt.Fprintf(out, "Done: %s active\n", args[1])
			return nil
		}
		if err := Deactivate(baseDir, args[1]); err != nil {
			return err
		}
		fmt.Fprintf(out, "Done: %s inactive\n", args[1])
		return nil
	case "list":
		names, err := List(baseDir)
		if err != nil {
			return err
		}
		fmt.Fprintln(out, "Active apps:")
		for _, n := range names {
			fmt.Fprintf(out, "  %s\n", n)
		}
		return nil
	}
	return usage
}
