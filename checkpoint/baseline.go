package checkpoint

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Baseline makes the current files the new baseline, restoring the preserved
// branches into the new repository, and returns the branches it restored.
// The old repository is set aside, not deleted, and discarded only once the new
// one is complete; on any failure it is put back byte for byte.
// CRC: crc-Checkpoint.md | Seq: seq-checkpoint.md#2.1 | R73, R74, R127, R192, R193
func (a App) Baseline() (restored []string, err error) {
	if !a.hasRepo() {
		// Seq: seq-checkpoint.md#2.1
		return nil, a.newBaseline()
	}
	// Seq: seq-checkpoint.md#2.2 — an open checkout is only verified, not
	// re-configured: a settings write would change the repository before it is
	// set aside.
	if exists(a.checkout()) {
		err = a.verify()
	} else {
		err = a.ensureCheckout()
	}
	if err != nil {
		return nil, err
	}
	// Seq: seq-checkpoint.md#2.4
	aside, err := os.MkdirTemp(a.Dir, ".#baseline.")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(aside)
	branches, err := a.exportBranches(aside)
	if err != nil {
		return nil, err
	}
	// Seq: seq-checkpoint.md#2.6
	if err := a.verify(); err != nil {
		return nil, err
	}
	// Seq: seq-checkpoint.md#2.7
	oldRepo, oldCheckout := filepath.Join(aside, "old.fossil"), filepath.Join(aside, "old.fslckout")
	if err := os.Rename(a.repo(), oldRepo); err != nil {
		return nil, err
	}
	if err := os.Rename(a.checkout(), oldCheckout); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, a.restore(oldRepo, oldCheckout, err)
	}
	if err := a.newBaseline(); err != nil {
		return nil, a.restore(oldRepo, oldCheckout, err)
	}
	if err := a.importBranches(aside, branches); err != nil {
		return nil, a.restore(oldRepo, oldCheckout, err)
	}
	return branches, nil
}

// newBaseline creates the repository and commits the current files as the baseline.
// Seq: seq-checkpoint.md#2.8
func (a App) newBaseline() error {
	if err := a.init(); err != nil {
		return err
	}
	_, err := a.step("commit", "-m", baselineMsg, "--no-warnings")
	return err
}

// restore removes the new repository and puts the old one and its checkout back.
// CRC: crc-Checkpoint.md | Seq: seq-checkpoint.md#2.11 | R193
func (a App) restore(oldRepo, oldCheckout string, cause error) error {
	var errs []error
	for _, p := range []string{a.repo(), a.checkout()} {
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, err)
		}
	}
	if err := os.Rename(oldRepo, a.repo()); err != nil {
		errs = append(errs, err)
	}
	if exists(oldCheckout) {
		if err := os.Rename(oldCheckout, a.checkout()); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("baseline failed for %s: %v\nrestoring the old repository also failed (it is in %s): %v",
			a.name(), cause, filepath.Dir(oldRepo), errors.Join(errs...))
	}
	return fmt.Errorf("baseline failed for %s; checkpoints unchanged: %w", a.name(), cause)
}

// exportBranches writes each preserved branch that exists to its own bundle in
// aside and returns their names. A branch that cannot be exported stops the
// baseline before anything moves.
// CRC: crc-Checkpoint.md | Seq: seq-checkpoint.md#2.3 | R127, R193, R194
func (a App) exportBranches(aside string) ([]string, error) {
	names, err := a.branchNames()
	if err != nil {
		return nil, err
	}
	var branches []string
	for _, b := range preservedBranches {
		if !names[b] {
			continue
		}
		if _, err := a.step("bundle", "export", filepath.Join(aside, b+".bundle"), "--branch", b, "--standalone"); err != nil {
			return nil, fmt.Errorf("baseline aborted for %s: could not export branch %s; checkpoints unchanged: %w", a.name(), b, err)
		}
		branches = append(branches, b)
	}
	return branches, nil
}

// importBranches imports each bundle and confirms its branch arrived, by exact
// name: an import that exits 0 having imported nothing still fails.
// CRC: crc-Checkpoint.md | Seq: seq-checkpoint.md#2.9 | R192, R193, R194
func (a App) importBranches(aside string, branches []string) error {
	for _, b := range branches {
		if _, err := a.step("bundle", "import", filepath.Join(aside, b+".bundle"), "--force", "--publish"); err != nil {
			return err
		}
		names, err := a.branchNames()
		if err != nil {
			return err
		}
		if !names[b] {
			return fmt.Errorf("branch %s is missing after its import", b)
		}
	}
	return nil
}

// branchNames returns the repository's branch names. Each line of `branch list`
// is one name, with "* " marking the current branch; names are compared whole,
// so local-experiment is never taken for local.
// CRC: crc-Checkpoint.md | R194
func (a App) branchNames() (map[string]bool, error) {
	out, err := a.step("branch", "list")
	if err != nil {
		return nil, err
	}
	names := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		if name := strings.TrimSpace(strings.TrimLeft(line, "* ")); name != "" {
			names[name] = true
		}
	}
	return names, nil
}
