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
- seq-mcp-notify.md

**Test Design:**
- test-MCP.md

---

## Level 2 <-> Level 3 (Design Models to Implementation)

*Implementation checkboxes updated to reflect actual code*

### crc-MCPServer.md
**Source Spec:** mcp.md
**Implementation:**
- [x] `internal/mcp/server.go` - MCP server (Lifecycle FSM, configuration, startup, notifications)
  - [x] SendNotification() wired to Lua mcp.notify() -> seq-mcp-notify.md
  - [x] getSessionCount callback for ui_status

### crc-MCPResource.md
**Source Spec:** mcp.md
**Implementation:**
- [ ] `internal/mcp/resources.go` - MCP resources (State root redirection via mcp.state, static resources from baseDir/resources)

### crc-MCPTool.md
**Source Spec:** mcp.md
**Implementation:**
- [x] `internal/mcp/tools.go` - MCP tools (configure, start, run, upload_viewdef, open_browser, status)
  - [x] handleStatus() for ui_status tool

---

## Integration with ui-engine

The MCP implementation integrates with ui-engine components:

**Session Management:**
- Uses ui-engine's Session and SessionManager for session lifecycle
- Creates backends via ui-engine's LuaBackend

**Lua Runtime:**
- Uses ui-engine's LuaRuntime for Lua code execution
- Exposes `mcp.state` and `mcp.notify` globals

**Variable Protocol:**
- Accesses state through ui-engine's VariableStore
- Uses ui-engine's ProtocolHandler for message processing

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
