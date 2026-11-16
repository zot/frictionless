# FlibRuntime

**Source Spec:** specs/mcp.md (Section 2.6)
**Requirements:** R161-167, R177, R179-182, R190, R191, R198-200

## Responsibilities

### Knows
- cfg: ui-engine configuration (Dir, Host, port, Lua settings)
- uiServer: ui-engine Server instance
- mcpServer: MCP Server instance

### Does
- New: Copy Config.BuiltinApps into the ui-engine config's Server.BuiltinApps (R190). Allocate ui-engine server, MCP server, and cleanup worker from Config (R162). When Config.Dir is non-empty, derive the ui-engine backend control socket as {Dir}/ui.sock instead of the shared /tmp/ui.sock default, isolating each embedded runtime's control socket (R182)
- Configure: Prepare base directory, run auto-install if needed (R163)
- Start: Start UI HTTP server, create Lua session with mcp global; return base URL (R164)
- RegisterAPI: Mount `/api/*`, `/wait`, `/state`, `/variables` on external mux (R165)
- StartAPI: Start standalone HTTP API server; return port (R166)
- WithLua: PASSIVE. Execute closure in Lua executor via the Lua session's own ExecuteInSession; never runs afterBatch, so nothing is pushed to the browser. For setup such as registering Go functions, which must not push. Resolves vendedID internally (R179)
- WithLuaUpdate: ACTIVE twin of WithLua. Same vendedID resolution, same errors, and hands fn the same LuaRuntime, but executes through uiServer.ExecuteInSession, which queues on the session's executor and runs afterBatch when a browser is connected, so the closure's changes reach the browser at once. For delivering Go-side events into Lua (R200)
- UIHandleFunc: Register custom HTTP handler on the UI server's mux; delegates to uiServer.HttpEndpoint.HandleFunc (R181)
- ViewdefCollisions: Return uiServer.ViewdefCollisions() unchanged (R191)
- ThemeBlock: Return generated theme HTML block string for in-memory injection (R198)
- InjectAllThemeBlocks: Patch all HTML files on disk with theme block (R199, via mcp.InjectAllThemeBlocks, R196)
- Shutdown: Graceful shutdown of UI and API servers, remove port files (R167)

## Collaborators

- MCPServer: Delegates Configure, StartAndCreateSession, RegisterAPIRoutes, StartHTTPServer, ShutdownHTTPServer
- UIServer (cli.Server): Provides ExecuteInSession (active path for WithLuaUpdate), GetLuaSession, StartAsync, GetSessions, GetViewdefManager, ViewdefCollisions, Shutdown, StartCleanupWorker
- Config: Holds Dir, Host, Project, Port, and BuiltinApps; converted to cli.Config internally. Project sets MCPServer.ProjectDir for install targeting.

## Sequences

- seq-mcp-lifecycle.md: Configure and Start mirror the standalone server lifecycle
