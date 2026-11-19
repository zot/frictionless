# AppLinkOps
**Requirements:** R234

The Lua side of app activation: `ui.linkApp`, `ui.unlinkApp` and `ui.deleteApp`, defined on
every session beside the checkpoint operations (`internal/mcp/applink_ops.go`). Each is a
few file-system calls, so they are synchronous `ui.*` functions, not host operations.

## Knows
- appsBase: the server's base directory, which holds `apps/` and `lua/`

## Does
- defineAppLinkOps(vendedID): on the session's executor, set `ui.linkApp`, `ui.unlinkApp`
  and `ui.deleteApp` on the session's `ui` table. Each takes one string, calls
  applink.Activate, Deactivate or Delete, returns nothing on success, and raises a Lua
  error carrying the operation's error text otherwise (R234).

## Collaborators
- AppLink: the operations
- Server.setupMCPGlobal: calls defineAppLinkOps beside defineCheckpointOps
- app-console: the caller (its `requirements.md`, Download and Delete App)

## Sequences
- seq-applink.md
