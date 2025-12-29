# MCPServer

**Source Spec:** specs/mcp.md

## Responsibilities

### Knows
- uiServer: Reference to UI server instance
- resources: List of available MCP resources
- tools: List of available MCP tools
- activeSession: Current session for AI interaction
- state: Lifecycle state (UNCONFIGURED, CONFIGURED, RUNNING)
- config: Server configuration (paths, I/O settings)
- getSessionCount: Callback to query active browser session count
- promptHTTPServer: Background HTTP server for prompt API
- promptManager: Manager for pending prompts

### Does
- initialize: Set up MCP server connection in UNCONFIGURED state
- configure: Transition to CONFIGURED state, setup directories and I/O (ui_configure)
- start: Transition to RUNNING state, launch HTTP server (ui_start)
- openBrowser: Launch system browser with conserve mode (ui_open_browser)
- listResources: Return available resources (session state)
- listTools: Return available tools (ui_configure, ui_start, ui_run, ui_upload_viewdef, ui_open_browser, ui_status)
- handleResourceRequest: Process resource queries (ui://state/{sessionId})
- handleToolCall: Execute tool operations by delegating to specific handlers
- sendNotification: Push events to AI client (wired to Lua mcp.notify())
- getStatus: Return current lifecycle state, URL, and session count
- shutdown: Clean up MCP connection
- startPromptServer: Start background HTTP server for prompt API
- handlePromptResponse: Receive promptResponse from browser, resolve pending prompt

## Collaborators

- MCPResource: Individual resource handlers
- MCPTool: Individual tool handlers
- SessionManager: Session operations
- LuaRuntime: Lua code execution and I/O redirection
- HTTPServer: Underlying HTTP service
- OS: Operating system interactions (filesystem, browser)
- PromptManager: Pending prompt tracking
- PromptHTTPServer: Background HTTP server for prompt API

## Sequences

- seq-mcp-lifecycle.md: Server configuration, startup, and browser launch
- seq-mcp-create-session.md: AI creating session
- seq-mcp-run.md: AI executing Lua code
- seq-mcp-get-state.md: AI inspecting state
- seq-mcp-notify.md: Lua code sending notifications to AI agent
- seq-prompt-flow.md: Permission prompt from hook to browser response