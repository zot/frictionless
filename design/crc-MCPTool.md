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
- ui_configure: Configure and start server (stop existing, reinitialize, start HTTP servers, write port files). Use `.claude/ui` unless user specifies otherwise.
- ui_run: Execute Lua code in session context
- ui_upload_viewdef: Add dynamic view definition and push to frontend
- ui_open_browser: Open system browser to session URL (defaults to ?conserve=true)
- ui_status: Get server state, version, base_dir, URL, and session count
- ui_install: Install bundled files with version checking (skills, resources, viewdefs, scripts)

## Collaborators

- MCPServer: Registers and invokes tools, manages lifecycle, sends notifications
- SessionManager: Session creation
- VariableStore: Presenter creation/update
- ViewdefStore: Viewdef management
- LuaRuntime: Lua code loading
- Router: URL path registration
- SharedWorker: Frontend coordination for conserve mode (via browser)
- OS: Filesystem operations for installation and port file creation
- Bundle: Embedded files from `install/` directory (init/, resources/, viewdefs/, scripts)

## Sequences

- seq-mcp-lifecycle.md: Server lifecycle tools (configure, open_browser)
- seq-mcp-run.md: Code execution
- seq-mcp-create-presenter.md: Viewdef creation