# CheckpointManager

**Requirements:** R69-79, R125-127, R192-194

## Knows

- `app`: Application name
- `app_dir`: Path to `{base_dir}/apps/{app}`
- `repo`: Path to `checkpoint.fossil`
- `FOSSIL_BIN`: Path to `~/.claude/bin/fossil`
- `aside`: `{app_dir}/.#baseline.XXXXXX` holding the old repo, its checkout file, and one bundle per preserved branch while a baseline is in progress (dot-prefixed, so `fossil add` never captures it)

## Does

- **ensure_fossil**: Download fossil if missing, detect platform
- **checkpoint_save**: Init repo if new, addremove, commit if changes
- **checkpoint_list**: Show timeline (excluding baseline and artifacts)
- **checkpoint_rollback**: Checkout Nth commit (or undo if no N)
- **checkpoint_diff**: Diff from Nth commit to current
- **checkpoint_clear**: Alias for baseline (reset repo)
- **checkpoint_baseline**: Export each preserved branch (exact name match) to its own bundle in `aside`, move the old repo and `.fslckout` into `aside`, create a fresh repo with current state, import the bundles; on any failure, delete the new repo and move the old one back untouched, report on stderr, exit non-zero; on success, discard `aside`
- **baseline_import**: Import each bundle and confirm its branch is present by exact name; fail if either step fails
- **baseline_has_branch**: True when fossil has a branch with exactly the given name
- **checkpoint_count**: Count commits excluding baseline
- **checkpoint_update**: Verify no fast checkpoints, switch to "updates" branch (create if needed), commit current state, switch back to trunk
- **checkpoint_local**: Switch to "local" branch (create if needed), commit current state, switch back to trunk
- **notify_ui**: Reset `appConsole._checkpointsTime` via mcp run

## Collaborators

- **MCPScript**: Parent script that invokes checkpoint commands
- **fossil**: Fossil SCM binary for version control
- **curl**: For downloading fossil binary
