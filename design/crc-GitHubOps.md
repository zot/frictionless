# GitHubOps
**Requirements:** R239, R245, R246

`ui.githubInvestigate` and `ui.githubInstall`, defined as ui-engine host operations on every
session beside the checkpoint operations (`internal/mcp/github_ops.go`). Each returns at
once; its work runs off the executor; the result lands in `target[field]` and
`target[field .. "Error"]`, then the optional callback runs (ui-engine `specs/host-ops.md`).

## Knows
- deadlines: 60 seconds for investigate, 120 seconds for install
- appsDir: the server's `{base_dir}/apps`

## Does
- defineGitHubOps(vendedID): on the session's executor, define both operations through
  `ref.Lua().DefineHostOp`.
- githubInvestigate's op: one string argument, the URL; ParseURL (a bad URL is an op error,
  delivered without running work); work runs Investigate and returns AsLua (R239).
- githubInstall's op: two string arguments, URL and name; ParseURL; work runs Install and
  returns true (R245).
- Neither runs another program (R246).

## Collaborators
- GitHubInstall: the work
- ui-engine LuaRig.DefineHostOp: delivery, deadline, callback
- Server.setupMCPGlobal: calls defineGitHubOps beside defineCheckpointOps
- app-console: the caller

## Sequences
- seq-github-install.md
