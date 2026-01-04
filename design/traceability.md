# Traceability Map

## Level 1 <-> Level 2 (Specs to Design Models)

### mcp.md

**CRC Cards:**
- crc-MCPServer.md
- crc-MCPResource.md
- crc-MCPTool.md

**Sequence Diagrams:**
- seq-mcp-lifecycle.md
- seq-mcp-create-session.md
- seq-mcp-create-presenter.md
- seq-mcp-receive-event.md
- seq-mcp-run.md
- seq-mcp-get-state.md
- seq-mcp-state-wait.md

**Test Design:**
- test-MCP.md

---

## Level 2 <-> Level 3 (Design Models to Implementation)

*Implementation checkboxes updated to reflect actual code*

### crc-MCPServer.md
**Source Spec:** mcp.md
**Implementation:**
- [x] `internal/mcp/server.go` - MCP server (Lifecycle FSM, configuration, startup)
  - [x] Configure() -> seq-mcp-lifecycle.md
  - [x] Start() -> seq-mcp-lifecycle.md
  - [x] Stop() - session restart support (159bc52)
  - [x] getSessionCount callback for ui_status
  - [x] ServeSSE() - SSE transport mode (5a11c38)
  - [x] handleVariables() - GET /variables (uses currentVendedID)
  - [x] handleState() - GET /state (uses currentVendedID)
  - [ ] handleWait() - GET /wait long-poll endpoint -> seq-mcp-state-wait.md
  - [ ] notifyStateChange() - signal waiting HTTP clients when mcp.pushState() called
  - [ ] atomicSwapQueue() - atomically swap mcp.state with empty table in Lua context

### crc-MCPResource.md
**Source Spec:** mcp.md
**Implementation:**
- [x] `internal/mcp/resources.go` - MCP resources
  - [x] handleGetStateResource() - ui://state
  - [x] handleGetVariablesResource() - ui://variables
  - [x] handleGetStaticResource() - ui://{path}

### crc-MCPTool.md
**Source Spec:** mcp.md
**Implementation:**
- [x] `internal/mcp/tools.go` - MCP tools
  - [x] handleConfigure() -> seq-mcp-lifecycle.md
    - [x] Directory creation
    - [x] Lua I/O redirection
    - [x] Resource extraction
    - [x] Installation check (returns install_needed hint)
  - [x] handleStart() -> seq-mcp-lifecycle.md
    - [x] Port selection (UI and MCP servers)
    - [x] Port file creation (ui-port, mcp-port)
  - [x] handleRun() -> seq-mcp-run.md (browser update trigger - ccbbf4f)
  - [x] handleUploadViewdef()
  - [x] handleOpenBrowser()
  - [x] handleStatus()
  - [ ] handleInstall() -> seq-mcp-lifecycle.md (Scenario 1a)
    - [ ] Bundled file installation (init/, resources/, viewdefs/, scripts)

---

## Integration with ui-engine

The MCP implementation integrates with ui-engine components:

**Session Management:**
- Uses ui-engine's Session and SessionManager for session lifecycle
- Creates backends via ui-engine's LuaBackend

**Lua Runtime:**
- Uses ui-engine's LuaRuntime for Lua code execution
- Exposes `mcp.state` global as event queue (array of event objects)
- `mcp.pushState({...})` queues event and triggers notifyStateChange() to wake waiting HTTP clients
- Wait endpoint atomically swaps queue with empty table to prevent event loss

**Variable Protocol:**
- Accesses state through ui-engine's VariableStore
- Uses ui-engine's ProtocolHandler for message processing

**Viewdef System:**
- Uses ui-engine's viewdef system for rendering
- Presenter switching via app._presenter triggers viewdef change

---

## External Dependencies

**ui-engine project:**
- Session management (`internal/session/`)
- Lua runtime (`internal/lua/`)
- Variable protocol (`internal/variable/`, `internal/protocol/`)
- Viewdef system (`internal/viewdef/`)

**MCP SDK:**
- Model Context Protocol implementation
- Tool and resource registration
