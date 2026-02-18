# CheckpointManager

**Requirements:** R69, R70, R71, R72, R73, R74, R75, R76, R77, R78, R79, R125, R126, R127

## Knows

- `app`: Application name
- `app_dir`: Path to `{base_dir}/apps/{app}`
- `repo`: Path to `checkpoint.fossil`
- `FOSSIL_BIN`: Path to `~/.claude/bin/fossil`
- `bundle`: Temporary bundle file path for preserved branch export

## Does

- **ensure_fossil**: Download fossil if missing, detect platform
- **checkpoint_save**: Init repo if new, addremove, commit if changes
- **checkpoint_list**: Show timeline (excluding baseline and artifacts)
- **checkpoint_rollback**: Checkout Nth commit (or undo if no N)
- **checkpoint_diff**: Diff from Nth commit to current
- **checkpoint_clear**: Alias for baseline (reset repo)
- **checkpoint_baseline**: Export preserved branches as bundle, close/remove repo, create fresh with current state, import bundle
- **checkpoint_count**: Count commits excluding baseline
- **checkpoint_update**: Verify no fast checkpoints, switch to "updates" branch (create if needed), commit current state, switch back to trunk
- **checkpoint_local**: Switch to "local" branch (create if needed), commit current state, switch back to trunk
- **notify_ui**: Reset `appConsole._checkpointsTime` via mcp run

## Collaborators

- **MCPScript**: Parent script that invokes checkpoint commands
- **fossil**: Fossil SCM binary for version control
- **curl**: For downloading fossil binary
