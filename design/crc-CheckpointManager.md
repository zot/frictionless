# CheckpointManager

**Requirements:** R69, R70, R71, R72, R73, R74, R75, R76, R77, R78, R79

## Knows

- `app`: Application name
- `app_dir`: Path to `{base_dir}/apps/{app}`
- `repo`: Path to `checkpoint.fossil`
- `FOSSIL_BIN`: Path to `~/.claude/bin/fossil`

## Does

- **ensure_fossil**: Download fossil if missing, detect platform
- **checkpoint_save**: Init repo if new, addremove, commit if changes
- **checkpoint_list**: Show timeline (excluding baseline and artifacts)
- **checkpoint_rollback**: Checkout Nth commit (or undo if no N)
- **checkpoint_diff**: Diff from Nth commit to current
- **checkpoint_clear**: Alias for baseline (reset repo)
- **checkpoint_baseline**: Close/remove repo, create fresh with current state
- **checkpoint_count**: Count commits excluding baseline
- **notify_ui**: Reset `appConsole._checkpointsTime` via mcp run

## Collaborators

- **MCPScript**: Parent script that invokes checkpoint commands
- **fossil**: Fossil SCM binary for version control
- **curl**: For downloading fossil binary
