# Test Design: CheckpointManager
**Source:** crc-CheckpointManager.md

Each test copies `install/mcp` into a temp base dir with one app, points
`HOME` at a temp dir whose `.claude/bin/fossil` is a shim around the real
fossil, and runs `mcp checkpoint baseline APP`. The shim fails on demand
(`FAIL_ON=export`, `FAIL_ON=import`) or makes `bundle import` exit 0
without importing (`FAIL_ON=import-noop`). Skipped when fossil is not
installed at `~/.claude/bin/fossil`.

## Test: baseline preserves branches
**Purpose:** a clean baseline restores "updates" and "local" and leaves no aside dir (R127, R192)
**Input:** app with an update checkpoint, a local checkpoint and a fast checkpoint
**Expected:** exit 0; stdout says preserved branches restored; branch list holds trunk, updates, local; checkpoint count is 0; no `.#baseline.*` in the app dir
**Refs:** crc-CheckpointManager.md
**Code:** internal/mcp/checkpoint_test.go

## Test: failed import leaves repo byte-identical
**Purpose:** the ark-reported data loss cannot recur (R192, R193)
**Input:** as above, with `FAIL_ON=import`
**Expected:** non-zero exit; stderr names the failure; stdout has no "Baseline set"; `checkpoint.fossil` and `.fslckout` hash identically to before; no `.#baseline.*` left
**Refs:** crc-CheckpointManager.md
**Code:** internal/mcp/checkpoint_test.go
**Fire alarm:** make baseline_import ignore the import's exit status (drop `|| return 1` from the import line) together with the branch check; red is a zero exit and changed hashes
**Inject:** install/mcp:baseline_import
**Pulled:** 2026-10-22 — rang: import subtest exit 0, stdout claimed success, both files changed

## Test: silent import is caught by the branch check
**Purpose:** an import that exits 0 but imports nothing still fails the baseline (R192, R193)
**Input:** `FAIL_ON=import-noop`
**Expected:** same as the failed-import test
**Refs:** crc-CheckpointManager.md
**Code:** internal/mcp/checkpoint_test.go
**Fire alarm:** delete the branch-presence check in baseline_import; red is a zero exit and a repo missing "local"
**Inject:** install/mcp:baseline_import
**Pulled:** 2026-10-22 — rang: import-noop subtest exit 0, both files changed

## Test: failed export leaves repo byte-identical
**Purpose:** a branch that cannot be exported aborts before anything moves (R193)
**Input:** `FAIL_ON=export`
**Expected:** as the failed-import test, and stderr names the branch that could not be exported
**Refs:** crc-CheckpointManager.md
**Code:** internal/mcp/checkpoint_test.go
**Fire alarm:** make a failed export fall through instead of aborting (the old `&& has_bundle=true` shape); the import step then catches the missing bundle and still rolls back, so red is only the stderr report lacking "could not export branch updates"
**Inject:** install/mcp:checkpoint_baseline
**Pulled:** 2026-10-22 — rang: export subtest stderr lacked "could not export branch updates" (silent before the stderr assertion was added)

## Test: branch names match exactly
**Purpose:** `local-experiment` is not taken for `local` (R194)
**Input:** app whose only extra branch is `local-experiment`
**Expected:** exit 0; stdout "Baseline set for APP" without "preserved branches"
**Refs:** crc-CheckpointManager.md
**Code:** internal/mcp/checkpoint_test.go
**Fire alarm:** restore the substring match (drop the `sed` strip and use `grep -q "$1"`); red is a non-zero exit, because exporting branch `local` fails
**Inject:** install/mcp:baseline_has_branch
**Pulled:** 2026-10-22 — rang: exact-names test, baseline exit status 1
