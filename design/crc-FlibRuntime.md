# FlibRuntime

**Source Spec:** specs/mcp.md (Section 2.6)
**Requirements:** R161, R162, R163, R164, R165, R166, R167

## Responsibilities

### Knows
- cfg: ui-engine configuration (Dir, Host, port, Lua settings)
- uiServer: ui-engine Server instance
- mcpServer: MCP Server instance

### Does
- New: Allocate ui-engine server, MCP server, and cleanup worker from Config (R162)
- Configure: Prepare base directory, run auto-install if needed (R163)
- Start: Start UI HTTP server, create Lua session with mcp global; return base URL (R164)
- RegisterAPI: Mount /api/*, /wait, /state, /variables on external mux (R165)
- StartAPI: Start standalone HTTP API server; return port (R166)
- Shutdown: Graceful shutdown of UI and API servers, remove port files (R167)

## Collaborators

- MCPServer: Delegates Configure, StartAndCreateSession, RegisterAPIRoutes, StartHTTPServer, ShutdownHTTPServer
- UIServer (cli.Server): Provides StartAsync, GetSessions, GetViewdefManager, Shutdown, StartCleanupWorker
- Config: Holds Dir and Host; converted to cli.Config internally

## Sequences

- seq-mcp-lifecycle.md: Configure and Start mirror the standalone server lifecycle
