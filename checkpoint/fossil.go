// Package checkpoint keeps per-app fossil checkpoints. It is imported by the
// frictionless binary and by ark, so both run the same hardened sequences.
// See specs/checkpoint.md.
//
// CRC: crc-Checkpoint.md | R202
package checkpoint

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	baselineMsg = "=== BASELINE ==="
	// ignoreGlob is set on every opened checkout; the aside directory (.#*)
	// and fossil's own files never enter a checkpoint.
	ignoreGlob = ".#*,.*~,*~,checkpoint.fossil,.fslckout"
)

// preservedBranches survive a baseline (R127).
var preservedBranches = []string{"updates", "local"}

// ErrNoFossil is returned by Find when no fossil binary is available. Its
// message tells the reader how to install one.
// CRC: crc-Checkpoint.md | R206, R207
var ErrNoFossil = errors.New(`fossil is not installed. To set up checkpoints, put it in ~/.claude/bin:

mkdir -p ~/.claude/bin
# Linux x86_64:
curl -sL "https://fossil-scm.org/home/uv/fossil-linux-x64-2.27.tar.gz" | tar -xzf - -C ~/.claude/bin fossil
# macOS ARM:
curl -sL "https://fossil-scm.org/home/uv/fossil-mac-arm-2.27.tar.gz" | tar -xzf - -C ~/.claude/bin fossil
# macOS x86_64:
curl -sL "https://fossil-scm.org/home/uv/fossil-mac-x64-2.27.tar.gz" | tar -xzf - -C ~/.claude/bin fossil
chmod +x ~/.claude/bin/fossil
# Windows: download fossil.exe from https://fossil-scm.org/home/uv/download.html into ~/.claude/bin`)

// Find locates fossil: ~/.claude/bin/fossil first (fossil.exe on Windows), then
// fossil on PATH.
// CRC: crc-Checkpoint.md | R206
func Find() (string, error) {
	name := "fossil"
	if runtime.GOOS == "windows" {
		name = "fossil.exe"
	}
	if home, err := os.UserHomeDir(); err == nil {
		bin := filepath.Join(home, ".claude", "bin", name)
		if _, err := os.Stat(bin); err == nil {
			return bin, nil
		}
	}
	if bin, err := exec.LookPath("fossil"); err == nil {
		return bin, nil
	}
	return "", ErrNoFossil
}

// App is one app's checkpoints: the repository Dir/checkpoint.fossil, worked on
// with the given fossil binary.
// CRC: crc-Checkpoint.md | R78, R208
type App struct {
	Dir    string // the app directory
	Fossil string // the fossil binary
}

func (a App) repo() string     { return filepath.Join(a.Dir, "checkpoint.fossil") }
func (a App) checkout() string { return filepath.Join(a.Dir, ".fslckout") }
func (a App) name() string     { return filepath.Base(a.Dir) }

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (a App) hasRepo() bool { return exists(a.repo()) }

// run executes one fossil command in Dir, unchecked, returning combined output.
func (a App) run(args ...string) (string, error) {
	cmd := exec.Command(a.Fossil, args...)
	cmd.Dir = a.Dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// step runs one fossil command as a checked step: on failure the error names the
// command and carries fossil's own output.
// CRC: crc-Checkpoint.md | R210
func (a App) step(args ...string) (string, error) {
	out, err := a.run(args...)
	if err != nil {
		return out, fmt.Errorf("fossil %s: %v\n%s", strings.Join(args, " "), err, strings.TrimSpace(out))
	}
	return out, nil
}

// resolvePath returns the symlink-resolved, cleaned form of p. Fossil reports
// resolved paths; app directories are routinely reached through links.
func resolvePath(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		p = r
	}
	return filepath.Clean(p)
}

// verify asks fossil which repository and checkout root it would use from Dir
// and errors unless both are the app's own. Fossil otherwise walks up into an
// enclosing checkout.
// CRC: crc-Checkpoint.md | R211
func (a App) verify() error {
	out, err := a.run("info")
	if err != nil {
		return fmt.Errorf("no open checkout in %s: %v\n%s", a.Dir, err, strings.TrimSpace(out))
	}
	var gotRepo, gotRoot string
	for _, line := range strings.Split(out, "\n") {
		if v, ok := strings.CutPrefix(line, "repository:"); ok {
			gotRepo = strings.TrimSpace(v)
		} else if v, ok := strings.CutPrefix(line, "local-root:"); ok {
			gotRoot = strings.TrimSpace(v)
		}
	}
	if gotRepo == "" || gotRoot == "" {
		return fmt.Errorf("no open checkout in %s (fossil info reports no repository or local-root)", a.Dir)
	}
	if resolvePath(gotRepo) != resolvePath(a.repo()) || resolvePath(gotRoot) != resolvePath(a.Dir) {
		return fmt.Errorf("checkout binding mismatch in %s:\n  expected repository %s at root %s\n  fossil reports repository %s at root %s",
			a.Dir, a.repo(), a.Dir, gotRepo, gotRoot)
	}
	return nil
}

// open binds a checkout in Dir: --nested binds here even inside an enclosing
// checkout, --keep leaves the working tree alone.
// CRC: crc-Checkpoint.md | R212
func (a App) open() error {
	_, err := a.step("open", a.repo(), "--force", "--nested", "--keep")
	return err
}

// ensureCheckout makes sure Dir has its own checkout of the app's repository,
// opening one when none is open. A checkout bound elsewhere is an error, never
// replaced.
// CRC: crc-Checkpoint.md | R211, R212
func (a App) ensureCheckout() error {
	if !exists(a.checkout()) {
		if err := a.open(); err != nil {
			return err
		}
	}
	if err := a.verify(); err != nil {
		return err
	}
	return a.settings()
}

// init creates the app's repository and opens it.
// CRC: crc-Checkpoint.md | Seq: seq-checkpoint.md#2.8 | R78, R212
func (a App) init() error {
	if _, err := a.step("init", a.repo(), "--project-name", a.name()); err != nil {
		return err
	}
	if err := a.open(); err != nil {
		return err
	}
	if err := a.verify(); err != nil {
		return err
	}
	if err := a.settings(); err != nil {
		return err
	}
	_, err := a.step("add", ".")
	return err
}

// settings configures an opened checkout. mtime-changes off makes commits hash
// content: the stat cache misses a same-size edit in the same second.
// CRC: crc-Checkpoint.md | R214
func (a App) settings() error {
	if _, err := a.step("settings", "mtime-changes", "off"); err != nil {
		return err
	}
	_, err := a.step("settings", "ignore-glob", ignoreGlob)
	return err
}
