package checkpoint

import (
	"fmt"
	"strings"
)

// Save commits the current files to trunk, creating the repository on first
// use. saved is false when nothing changed.
// CRC: crc-Checkpoint.md | R69, R78
func (a App) Save(msg string) (saved bool, err error) {
	if !a.hasRepo() {
		if msg == "" {
			msg = "initial state"
		}
		if err := a.init(); err != nil {
			return false, err
		}
		_, err := a.step("commit", "-m", msg, "--no-warnings")
		return err == nil, err
	}
	if msg == "" {
		msg = "checkpoint"
	}
	if err := a.ensureCheckout(); err != nil {
		return false, err
	}
	if _, err := a.step("addremove"); err != nil {
		return false, err
	}
	changes, err := a.step("changes", "--quiet")
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(changes) == "" {
		return false, nil
	}
	_, err = a.step("commit", "-m", msg, "--no-warnings")
	return err == nil, err
}

// timeline returns the check-in lines, newest first, without fossil's footer.
func (a App) timeline() ([]string, error) {
	out, err := a.step("timeline", "-n", "0", "-t", "ci", "--oneline")
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, l := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if l == "" || strings.HasPrefix(l, "---") || strings.HasPrefix(l, "+++") {
			continue
		}
		lines = append(lines, l)
	}
	return lines, nil
}

// List returns one line per check-in, newest first.
// CRC: crc-Checkpoint.md | R70, R208
func (a App) List() ([]string, error) {
	if !a.hasRepo() {
		return nil, nil
	}
	if err := a.ensureCheckout(); err != nil {
		return nil, err
	}
	return a.timeline()
}

// Rollback restores the nth checkpoint, counting from the newest and skipping
// the baseline; n == 0 undoes the last fossil operation.
// CRC: crc-Checkpoint.md | R71
func (a App) Rollback(n int) error {
	if !a.hasRepo() {
		return fmt.Errorf("no checkpoints for %s", a.name())
	}
	if err := a.ensureCheckout(); err != nil {
		return err
	}
	if n == 0 {
		_, err := a.step("undo")
		return err
	}
	lines, err := a.timeline()
	if err != nil {
		return err
	}
	var commits []string
	for _, l := range lines {
		if !strings.Contains(l, baselineMsg) {
			commits = append(commits, l)
		}
	}
	if n < 1 || n > len(commits) {
		return fmt.Errorf("checkpoint %d not found", n)
	}
	// The forced checkout restores the tree by design: verify right before it.
	if err := a.verify(); err != nil {
		return err
	}
	_, err = a.step("checkout", strings.Fields(commits[n-1])[0], "--force")
	return err
}

// Diff returns the diff from the nth check-in (newest is 1) to the current files.
// CRC: crc-Checkpoint.md | R72
func (a App) Diff(n int) (string, error) {
	if !a.hasRepo() {
		return "", fmt.Errorf("no checkpoints for %s", a.name())
	}
	if err := a.ensureCheckout(); err != nil {
		return "", err
	}
	lines, err := a.timeline()
	if err != nil {
		return "", err
	}
	if n < 1 || n > len(lines) {
		return "", fmt.Errorf("checkpoint %d not found", n)
	}
	// diff exits non-zero when there are differences, so it is run unchecked.
	out, _ := a.run("diff", "--from", strings.Fields(lines[n-1])[0])
	return out, nil
}

// Count returns the number of trunk checkpoints, not counting the baseline.
// CRC: crc-Checkpoint.md | R75
func (a App) Count() (int, error) {
	if !a.hasRepo() {
		return 0, nil
	}
	if err := a.ensureCheckout(); err != nil {
		return 0, err
	}
	return a.trunkCount()
}

func (a App) trunkCount() (int, error) {
	out, err := a.step("timeline", "-n", "0", "-t", "ci", "--oneline", "-b", "trunk")
	if err != nil {
		return 0, err
	}
	count := 0
	for _, l := range strings.Split(out, "\n") {
		if l == "" || strings.Contains(l, baselineMsg) || strings.Contains(l, "initial empty check-in") ||
			strings.HasPrefix(l, "+++ ") || strings.HasPrefix(l, "--- ") {
			continue
		}
		count++
	}
	return count, nil
}

// Update saves the current files on the "updates" branch. It refuses while
// trunk holds checkpoints that have not been consolidated.
// CRC: crc-Checkpoint.md | R125
func (a App) Update(msg string) error {
	if msg == "" {
		msg = "update"
	}
	if !a.hasRepo() {
		return a.firstBranchSave("updates", msg)
	}
	if err := a.ensureCheckout(); err != nil {
		return err
	}
	count, err := a.trunkCount()
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%d checkpoint(s) exist. Consolidate before updating", count)
	}
	return a.branchSave("updates", msg)
}

// Local saves the current files on the "local" branch.
// CRC: crc-Checkpoint.md | R126
func (a App) Local(msg string) error {
	if msg == "" {
		msg = "local"
	}
	if !a.hasRepo() {
		return a.firstBranchSave("local", msg)
	}
	if err := a.ensureCheckout(); err != nil {
		return err
	}
	return a.branchSave("local", msg)
}

// firstBranchSave is a branch save with no repository yet: the files become the
// trunk baseline, and the branch starts with an empty commit holding the same
// content. The checkout ends on trunk.
func (a App) firstBranchSave(branch, msg string) error {
	if err := a.init(); err != nil {
		return err
	}
	if _, err := a.step("commit", "-m", baselineMsg, "--no-warnings"); err != nil {
		return err
	}
	if _, err := a.step("commit", "--branch", branch, "-m", msg, "--allow-empty", "--no-warnings"); err != nil {
		return err
	}
	_, err := a.step("checkout", "trunk", "--keep")
	return err
}

// branchSave commits the current files onto a side branch without touching the
// working tree: `checkout --keep` retargets the checkout and leaves every file
// alone. `fossil update` would extract trunk over the tree.
// CRC: crc-Checkpoint.md | Seq: seq-checkpoint.md#3.1 | R125, R126, R213
func (a App) branchSave(branch, msg string) error {
	out, err := a.run("checkout", branch, "--keep")
	if err != nil {
		if !strings.Contains(out, "cannot resolve name") {
			return fmt.Errorf("fossil checkout %s --keep: %v\n%s", branch, err, strings.TrimSpace(out))
		}
		// Seq: seq-checkpoint.md#3.2.1 — the branch does not exist yet
		if _, err := a.step("addremove"); err != nil {
			return err
		}
		if _, err := a.step("commit", "--branch", branch, "-m", msg, "--allow-empty", "--no-warnings"); err != nil {
			return err
		}
		return a.returnToTrunk(nil)
	}
	if _, err := a.step("addremove"); err != nil {
		return a.returnToTrunk(err)
	}
	out, err = a.run("commit", "-m", msg, "--allow-empty", "--no-warnings")
	if err != nil && strings.Contains(out, "closed leaf") {
		// Seq: seq-checkpoint.md#3.3.1 — --raw matters: the plain form cancels a
		// sym-closed tag and leaves the leaf closed.
		if _, err := a.step("tag", "cancel", "--raw", "closed", branch); err != nil {
			return a.returnToTrunk(err)
		}
		out, err = a.run("commit", "-m", msg, "--allow-empty", "--no-warnings")
	}
	if err != nil {
		return a.returnToTrunk(fmt.Errorf("fossil commit on %s: %v\n%s", branch, err, strings.TrimSpace(out)))
	}
	return a.returnToTrunk(nil)
}

// returnToTrunk retargets the checkout to trunk without touching files. A prior
// error stays primary; a failure to return is reported beside it.
func (a App) returnToTrunk(prior error) error {
	if _, err := a.step("checkout", "trunk", "--keep"); err != nil {
		if prior != nil {
			return fmt.Errorf("%v\nadditionally, returning to trunk failed: %v", prior, err)
		}
		return err
	}
	return prior
}
