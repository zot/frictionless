# Architecture

**Entry point to the design - shows how design elements are organized into logical systems**

**Sources**: mcp.md

---

## MCP Integration System

**Purpose**: AI assistant integration via Model Context Protocol

The MCP server provides tools and resources for AI assistants to interact with UI applications. It manages the lifecycle of UI sessions, provides state access, and enables programmatic UI control.

### Components

**MCPServer**: Core server implementing MCP protocol
- Lifecycle state machine (Unconfigured → Configured → Running)
- Configuration management
- Tool and resource registration

**MCPResource**: Static and dynamic resource providers
- State root redirection via `mcp.state`
- Static resources from `baseDir/resources`

**MCPTool**: AI assistant tools
- `configure` - Set UI configuration
- `start` - Start the UI server
- `run` - Execute Lua code
- `upload_viewdef` - Upload view definitions
- `open_browser` - Launch browser to UI

### Design Elements

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

**Test Design:**
- test-MCP.md

---

## Integration with ui-engine

This MCP server integrates with the ui-engine project:
- Uses `internal/lua/runtime.go` for Lua execution
- Creates sessions via ui-engine's session management
- Accesses variable state through ui-engine's variable store
- Delivers viewdefs through ui-engine's viewdef system

---

*This file serves as the architectural "main program" for MCP - start here to understand the design structure*
