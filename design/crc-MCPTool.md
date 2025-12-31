# MCPTool

**Source Spec:** specs/mcp.md

## Responsibilities

### Knows
- name: Tool identifier
- description: Human-readable description
- inputSchema: JSON schema for tool parameters
- handler: Function to execute tool

### Does
- define: Define tool schema (name, description, input schema)
- handle: Execute tool logic (interface implementation)

### Standard Tools
- ui_configure: Prepare server environment (files, logs, I/O, agent installation)
- ui_start: Launch HTTP server
- ui_run: Execute Lua code in session context
- ui_upload_viewdef: Add dynamic view definition and push to frontend
- ui_open_browser: Open system browser to session URL (defaults to ?conserve=true)
- ui_status: Get server state, URL, and session count

## Collaborators

- MCPServer: Registers and invokes tools, manages lifecycle, sends notifications
- SessionManager: Session creation
- VariableStore: Presenter creation/update
- ViewdefStore: Viewdef management
- LuaRuntime: Lua code loading
- Router: URL path registration
- SharedWorker: Frontend coordination for conserve mode (via browser)
- OS: Filesystem operations for agent file installation

## Sequences

- seq-mcp-lifecycle.md: Server lifecycle tools (configure, start, open_browser)
- seq-mcp-run.md: Code execution
- seq-mcp-create-presenter.md: Viewdef creation