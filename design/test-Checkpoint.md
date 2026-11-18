# Test Design: Checkpoint
**Source:** crc-Checkpoint.md

Tests run the real fossil in temporary directories, because the defects behind this
package were interactions between our command sequences and fossil's actual behaviour; a
fake fossil would encode the same wrong assumptions. Failures are injected with a shim
around the real fossil, passed as `App.Fossil`, that fails on demand (`FAIL_ON=export`,
`FAIL_ON=import`) or makes `bundle import` exit 0 without importing
(`FAIL_ON=import-noop`). Skipped when fossil is not installed.

Nearly every property here passes by default when absent: an untouched working tree, a
checkout that did not walk up, a history that was not lost. The baseline alarms were first
pulled against the bash script on 2026-10-22; they move here with Go inject sites and are
pulled again.

## Test: a branch save leaves the working tree alone
**Purpose:** saving to "local" and returning to trunk keeps uncommitted edits on disk (R125, R126, R213)
**Input:** an app with a trunk checkpoint and an existing "local" branch (the first branch save takes the branch-missing path); edit a file; `Local("wip")`
**Expected:** the edited content is on disk; the "local" branch holds it; trunk does not
**Refs:** crc-Checkpoint.md, seq-checkpoint.md
**Code:** checkpoint/checkpoint_test.go
**Alarm:** 1
**Fire alarm:** return to trunk with `update trunk` instead of `checkout trunk --keep`; the edit is reverted on disk
**Inject:** checkpoint/checkpoint.go:App.branchSave
@alarm-hash: c317df3add8c11d9
**Pulled:** 2026-11-18 — rang: returning to trunk with `update trunk` in branchSave; "working tree changed by a branch save: app.lua = -- v1"

## Test: a closed leaf is reopened
**Purpose:** saving to a branch whose leaf is closed reopens it and commits (R125)
**Input:** an app with a "local" checkpoint; close the leaf; edit; `Local("again")`
**Expected:** no error; "local" holds the new content
**Refs:** crc-Checkpoint.md, seq-checkpoint.md
**Code:** checkpoint/checkpoint_test.go

## Test: a missing branch is created
**Purpose:** (R125, R126)
**Input:** an app with only trunk; `Update("first")`
**Expected:** no error; an "updates" branch exists holding the files
**Refs:** crc-Checkpoint.md
**Code:** checkpoint/checkpoint_test.go

## Test: an enclosing checkout is not walked into
**Purpose:** an app inside another fossil checkout gets its own, and the enclosing one is untouched (R211, R212)
**Input:** an enclosing checkout with one commit; an app directory inside it; `Save("x")`
**Expected:** the app has its own checkpoint; the enclosing checkout's timeline and status are unchanged
**Refs:** crc-Checkpoint.md, seq-checkpoint.md
**Code:** checkpoint/checkpoint_test.go
**Alarm:** 2
**Fire alarm:** open without `--nested`; the save fails or lands in the enclosing checkout
**Inject:** checkpoint/fossil.go:App.init
@alarm-hash: 1f39620661c3afe7
**Pulled:** 2026-11-18 — rang: init opened without --nested; "save inside an enclosing checkout: fossil open ... --force --keep: exit status 1"

## Test: a foreign binding stops the command
**Purpose:** a checkout bound to another repository is reported, never used or replaced (R211, R212)
**Input:** an app whose checkout is bound to an unrelated repository; `Save("x")`
**Expected:** an error naming the binding; the unrelated repository has no new check-in
**Refs:** crc-Checkpoint.md
**Code:** checkpoint/checkpoint_test.go
**Alarm:** 3
**Fire alarm:** make verify return nil; the save succeeds
**Inject:** checkpoint/fossil.go:App.verify
@alarm-hash: 3406f15f292345a0
**Pulled:** 2026-11-18 — rang: verify returned nil; "Save succeeded through a checkout bound to another repository"

## Test: a fossil failure surfaces
**Purpose:** no swallowed errors, no unconditional success (R210)
**Input:** an app whose `checkpoint.fossil` is not a repository; `Save("x")`
**Expected:** an error naming the failing fossil command (`fossil open`) and carrying its output
**Refs:** crc-Checkpoint.md
**Code:** checkpoint/checkpoint_test.go
**Alarm:** 4
**Fire alarm:** make step ignore fossil's exit status; the later verify still fails, but the error no longer names `fossil open`
**Inject:** checkpoint/fossil.go:App.step
@alarm-hash: 7e1d6705577a4ad4
**Pulled:** 2026-11-18 — rang: step ignored the exit status; "error does not name the failing fossil command (open): no open checkout in ..."

## Test: a same-second, same-size edit is saved
**Purpose:** commits hash content (R214)
**Input:** save; rewrite a file with same-size content and restore its old mtime; `Save("y")`
**Expected:** saved is true and the count rises
**Refs:** crc-Checkpoint.md
**Code:** checkpoint/checkpoint_test.go
**Alarm:** 5
**Fire alarm:** stop setting `mtime-changes off`; saved is false
**Inject:** checkpoint/fossil.go:App.settings
@alarm-hash: 1e2e22626e9fbdac
**Pulled:** 2026-11-18 — rang: mtime-changes off removed; "a same-size edit with an unchanged mtime was not saved"

## Test: baseline preserves branches
**Purpose:** a clean baseline restores "updates" and "local" and leaves no aside directory (R74, R127, R192)
**Input:** an app with an update checkpoint, a local checkpoint and a trunk checkpoint; `Baseline()`
**Expected:** no error; the branches are trunk, updates, local; Count is 0; no `.#baseline.*` in the app directory
**Refs:** crc-Checkpoint.md, seq-checkpoint.md
**Code:** checkpoint/baseline_test.go

## Test: a failed import leaves the repository byte-identical
**Purpose:** the data loss ark reported cannot recur (R192, R193)
**Input:** as above, with the shim and `FAIL_ON=import`
**Expected:** an error naming the failure; `checkpoint.fossil` and `.fslckout` hash as before; no `.#baseline.*` left
**Refs:** crc-Checkpoint.md, seq-checkpoint.md
**Code:** checkpoint/baseline_test.go
**Alarm:** 6
**Fire alarm:** make importBranches return nil without importing; the baseline reports success
**Inject:** checkpoint/baseline.go:App.importBranches
@alarm-hash: a4f3be12b4967f93
**Pulled:** 2026-11-18 — rang: importBranches returned nil; "baseline with FAIL_ON=import reported success"

## Test: a silent import is caught by the branch check
**Purpose:** an import that exits 0 but imports nothing still fails the baseline (R192, R193)
**Input:** `FAIL_ON=import-noop`
**Expected:** as the failed-import test
**Refs:** crc-Checkpoint.md
**Code:** checkpoint/baseline_test.go
**Alarm:** 7
**Fire alarm:** delete the branch-presence check after import; no error, and the repository lacks "local"
**Inject:** checkpoint/baseline.go:App.importBranches
@alarm-hash: a4f3be12b4967f93
**Pulled:** 2026-11-18 — rang: branch-presence check disabled; "baseline with FAIL_ON=import-noop reported success"

## Test: a failed export leaves the repository byte-identical
**Purpose:** a branch that cannot be exported stops the baseline before anything moves (R193)
**Input:** `FAIL_ON=export`
**Expected:** as the failed-import test, and the error names the branch that could not be exported
**Refs:** crc-Checkpoint.md
**Code:** checkpoint/baseline_test.go
**Alarm:** 8
**Fire alarm:** skip a branch whose export fails instead of stopping; the baseline reports success without the branch
**Inject:** checkpoint/baseline.go:App.exportBranches
@alarm-hash: c271502706331190
**Pulled:** 2026-11-18 — rang: failed export skipped with continue; "baseline with FAIL_ON=export reported success"

## Test: branch names match exactly
**Purpose:** `local-experiment` is not taken for `local` (R194)
**Input:** an app whose only extra branch is `local-experiment`; `Baseline()`
**Expected:** no error; no branch was preserved
**Refs:** crc-Checkpoint.md
**Code:** checkpoint/baseline_test.go
**Alarm:** 9
**Fire alarm:** match branch names by substring; the baseline fails exporting `local`
**Inject:** checkpoint/baseline.go:branchNames
@alarm-hash: 6508312ae6a782cd
**Pulled:** 2026-11-18 — rang: substring branch match in branchNames; "could not export branch local ... cannot resolve name: local"

## Test: fossil is found in its home, then on PATH
**Purpose:** (R206, R207)
**Input:** HOME with no fossil and PATH holding one; then HOME with one; then neither
**Expected:** the PATH fossil; then the HOME one; then ErrNoFossil whose message names `~/.claude/bin` and a download
**Refs:** crc-Checkpoint.md
**Code:** checkpoint/fossil_test.go
**Alarm:** 10
**Fire alarm:** skip the PATH lookup; the first case returns ErrNoFossil
**Inject:** checkpoint/fossil.go:Find
@alarm-hash: 38720bff568753e8
**Pulled:** 2026-11-18 — rang: PATH lookup removed from Find; `with fossil only on PATH: Find() = "", fossil is not installed`

## Test: changed since install
**Purpose:** Changed reports edits, additions and removals against `original.fossil`, and nothing else (R215, R216, R217)
**Input:** an app without `original.fossil`; then baselined with `original.fossil` copied as app-console does, plus an editor backup file; then each of: a file edited, a file added, a file removed (each from a fresh copy)
**Expected:** false; false; true for each change
**Refs:** crc-Checkpoint.md, seq-checkpoint.md
**Code:** checkpoint/changed_test.go
**Alarm:** 11
**Fire alarm:** compare only the files in the tarball; the added file reads unchanged
**Inject:** checkpoint/changed.go:App.Changed
@alarm-hash: 684c00738e86901f
**Pulled:** 2026-11-18 — rang: added files not counted in Changed; "a file added: Changed = false, want true"

## Test: fossil's own files do not count as changes
**Purpose:** (R216)
**Input:** an app baselined with `original.fossil`, untouched
**Expected:** Changed is false although `checkpoint.fossil`, `original.fossil` and `.fslckout` are not in the tarball
**Refs:** crc-Checkpoint.md
**Code:** checkpoint/changed_test.go
**Alarm:** 12
**Fire alarm:** drop the ignore filtering from the walk; Changed is true
**Inject:** checkpoint/changed.go:App.Changed
@alarm-hash: 684c00738e86901f
**Pulled:** 2026-11-18 — rang: ignore filtering dropped from the walk; "an untouched installed app: Changed = true, want false"

## Test: Command prints, and calls onChange only after a changing verb
**Purpose:** (R79, R204, R208, R209)
**Input:** `Command` with `save`, then `count`, then `list`, counting onChange calls; then an unknown verb
**Expected:** save prints its message and onChange ran once; count prints "1" and list prints the checkpoint (and fossil's initial check-in), with no further onChange; an unknown verb and a missing app are errors
**Refs:** crc-Checkpoint.md, seq-checkpoint.md
**Code:** checkpoint/command_test.go
**Alarm:** 13
**Fire alarm:** call onChange after every verb; it runs three times
**Inject:** checkpoint/command.go:Command
@alarm-hash: 992089c6c4072472
**Pulled:** 2026-11-18 — rang: onChange called after every verb; "onChange ran 3 times in all, want 1"

## Test: the binary's notify never changes the exit status
**Purpose:** with no server running, a successful command still exits 0 (R203, R205)
**Input:** `runCheckpoint` with a base directory holding one app and no `mcp-port`; `save`
**Expected:** exit 0
**Refs:** crc-Checkpoint.md, seq-checkpoint.md
**Code:** cmd/frictionless/checkpoint_test.go
**Alarm:** 14
**Fire alarm:** return exit 1 when the notify fails; the exit is 1
**Inject:** cmd/frictionless/checkpoint.go:runCheckpoint
@alarm-hash: 62664469152bee06
**Pulled:** 2026-11-18 — rang: runCheckpoint returned 1 after the notify; "save exited 1 with no server running, want 0"

## Test: the script hands checkpoint off to the binary
**Purpose:** `mcp checkpoint` adds nothing to the binary's arguments, output or exit status (R201)
**Input:** `install/mcp` in a temp base dir; HOME's `.claude/bin` holding a fossil and a stub `frictionless` that prints its arguments and exits 3
**Expected:** the stub's arguments are `checkpoint count demo --dir BASE`; its output is the script's; the exit is 3
**Refs:** crc-MCPScript.md
**Code:** internal/mcp/checkpoint_test.go
**Alarm:** 15
**Fire alarm:** run the binary without `exec` and exit 0 after it; the exit is 0
**Inject:** install/mcp:checkpoint_cmd
**Pulled:** 2026-11-18 — rang: binary run without exec, then exit 0; "script exit = <nil>, want the binary's exit status 3"
