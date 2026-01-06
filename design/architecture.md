# Architecture

**Entry point to the design - shows how design elements are organized into logical systems**

**Sources**: mcp.md

---

## MCP Integration System

**Purpose**: AI assistant integration via Model Context Protocol

**Design Elements:**
- crc-MCPServer.md
- crc-MCPResource.md
- crc-MCPTool.md
- seq-mcp-lifecycle.md
- seq-mcp-create-session.md
- seq-mcp-create-presenter.md
- seq-mcp-receive-event.md
- seq-mcp-run.md
- seq-mcp-get-state.md
- seq-mcp-state-wait.md
- test-MCP.md

---

## MCP Resources

**Purpose**: Expose UI metadata to AI agents

**Design Elements:**
- crc-MCPResource.md

---

## Transport System

**Purpose**: Support multiple MCP transport modes for different use cases

**Transport Modes:**
- **Stdio** (`mcp` command): JSON-RPC 2.0 over stdin/stdout for AI agent integration
- **SSE** (`serve` command): Server-Sent Events over HTTP for standalone debugging

**Design Elements:**
- crc-MCPServer.md (ServeStdio, ServeSSE methods)
- seq-mcp-lifecycle.md

---

## HTTP Endpoints

**Purpose**: Debug and inspect runtime state (uses server's distinguished session)

**Two-Port Architecture:**
- **UI Server**: Serves HTML/JS and WebSocket connections
- **MCP Server**: Serves debug endpoints below
- Port files written to `{base_dir}/ui-port` and `{base_dir}/mcp-port`

**Endpoints (MCP port):**
- `GET /wait`: Long-poll for mcp.pushState() events
- `GET /variables`: Interactive variable tree view
- `GET /state`: Current session state JSON

**Design Elements:**
- crc-MCPServer.md (handleWait, handleVariables, handleState)
- seq-mcp-state-wait.md

---

## Integration with ui-engine

This MCP server integrates with the ui-engine project:
- Uses `internal/lua/runtime.go` for Lua execution
- Creates sessions via ui-engine's session management
- Accesses variable state through ui-engine's variable store
- Delivers viewdefs through ui-engine's viewdef system

---

## Cross-Cutting Concerns

### Browser Update Mechanism

**Purpose**: Ensure browser UIs refresh when server-side state changes

**Mechanism** (see mcp.md Section 4.1):
- MCP server delegates to ui-server's `Server.ExecuteInSession` method
- `ExecuteInSession` queues function through session executor (serializing with WebSocket messages)
- Executes function, then calls `afterBatch` to push state changes to browsers
- Any operation needing browser update can call `ExecuteInSession` with an empty function

**Panic Recovery**:
- MCP server MUST wrap `ExecuteInSession` with `SafeExecuteInSession`
- Catches Lua errors/panics and returns them as errors
- Prevents crashes from propagating to MCP process

**Design Elements:**
- crc-MCPServer.md (SafeExecuteInSession, triggerBrowserUpdate)
- seq-mcp-state-wait.md (browser update after queue drain)

---

*This file serves as the architectural "main program" - start here to understand the design structure*
