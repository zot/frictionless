# MCPServer

**Source Spec:** specs/mcp.md

## Responsibilities

### Knows
- uiServer: Reference to UI server instance
- mcpServer: Reference to MCP HTTP server instance
- resources: List of available MCP resources
- tools: List of available MCP tools
- activeSession: Current session for AI interaction
- state: Lifecycle state (CONFIGURED, RUNNING)
- config: Server configuration (paths, I/O settings)
- uiPort: UI server port (serves HTML/JS/WebSocket)
- mcpPort: MCP server port (serves /state, /wait, /variables)
- getSessionCount: Callback to query active browser session count
- currentVendedID: Current session's vended ID for cleanup on reconfigure
- stateWaiters: Waiting HTTP requests for current session (channels)
- mcpStateQueue: Event queue for current session (mcp.state)

### Does
- initialize: Set up MCP server, auto-install if README.md missing, start in CONFIGURED state
- configure: Reconfigure to different base_dir, auto-install if needed (ui_configure)
- start: Transition to RUNNING state, launch HTTP server (ui_start)
- stop: Destroy current session, reset to CONFIGURED (enables session restart)
- openBrowser: Launch system browser with conserve mode (ui_open_browser)
- listResources: Return available resources (ui://state, ui://variables)
- listTools: Return available tools (ui_configure, ui_start, ui_run, ui_upload_viewdef, ui_open_browser, ui_status, ui_install)
- handleResourceRequest: Process resource queries (ui://state uses currentVendedID)
- handleToolCall: Execute tool operations by delegating to specific handlers
- handleWait: HTTP long-poll endpoint for state changes (GET /wait, uses currentVendedID); after draining queue, calls SafeExecuteInSession with empty function to trigger browser update
- notifyStateChange: Signal waiting HTTP clients when mcp.pushState() called
- atomicSwapQueue: Atomically swap mcp.state with empty table, return accumulated events
- SafeExecuteInSession: Wraps ui-server's ExecuteInSession with panic recovery; converts Lua errors/panics to errors
- triggerBrowserUpdate: Call SafeExecuteInSession with empty function to push state changes to browsers
- getStatus: Return current lifecycle state, URL, and session count
- shutdown: Clean up MCP connection
- serveSSE: Start MCP server on HTTP with SSE transport (serve command)
- handleVariables: Render interactive variable tree (GET /variables)
- handleState: Return session state JSON (GET /state)
- setupMCPGlobal: Register mcp global table in Lua (mcp.type, mcp.value, mcp.pushState, mcp:pollingEvents, mcp:display, mcp:status)
- loadMCPLua: Load `{base_dir}/lua/mcp.lua` if it exists, extending the mcp global
- loadAppInitFiles: Scan `{base_dir}/apps/*/` and load `init.lua` from each app directory if it exists

## Collaborators

- MCPResource: Individual resource handlers
- MCPTool: Individual tool handlers
- SessionManager: Session operations
- LuaRuntime: Lua code execution and I/O redirection
- HTTPServer: Underlying HTTP service
- UIServer: Provides ExecuteInSession for session-context execution with afterBatch browser updates
- OS: Operating system interactions (filesystem, browser)

## Sequences

- seq-mcp-lifecycle.md: Server configuration, startup, and browser launch
- seq-mcp-create-session.md: AI creating session
- seq-mcp-run.md: AI executing Lua code
- seq-mcp-get-state.md: AI inspecting state
- seq-mcp-state-wait.md: Agent waiting for state changes via HTTP long-poll (GET /wait)
