# CheckpointOps
**Requirements:** R223, R224, R225

The checkpoint host operations each session's Lua gets: `ui.checkpointStatus` and
`ui.checkpointInstall`. Defined by the MCP server where it builds the `mcp` global, on
ui-engine's host operations (`LuaView.DefineHostOp`), so the fossil work runs off the
session's executor and the result lands in a field the caller names.

## Knows
- appsDir: `{base_dir}/apps`
- statusDeadline (60s) and installDeadline (120s)

## Does
- define: on the executor, inside setupMCPGlobal, define both operations on the session's
  LuaView (R223).
- checkpointStatus op: read `names` (a Lua list of strings) on the executor; work: Find
  fossil, then for each name `App.Status()`; deliver a map from name to
  `{count, downloaded, changed}`, or `{error}` for an app whose status failed, so one
  unreadable app does not hide the others (R224).
- checkpointInstall op: read `name` on the executor; work: Find fossil, `App.Install()`;
  deliver true (R225).

## Collaborators
- MCPServer: setupMCPGlobal calls define
- LuaView (ui-engine): DefineHostOp
- Checkpoint: Find, App.Status, App.Install

## Sequences
- seq-checkpoint.md
