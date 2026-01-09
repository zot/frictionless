# Design: ui-mcp

**Source Spec:** specs/mcp.md

## Intent

MCP (Model Context Protocol) server for AI assistants to control browser-based UIs. Provides tools for session management, Lua execution, and browser interaction.

## Artifacts

### CRC Cards
- [x] crc-MCPServer.md → `internal/mcp/server.go`
- [x] crc-MCPResource.md → `internal/mcp/resources.go`
- [x] crc-MCPTool.md → `internal/mcp/tools.go`

### Sequences
- [x] seq-mcp-lifecycle.md → `internal/mcp/server.go`, `internal/mcp/tools.go`
- [x] seq-mcp-create-session.md → `internal/mcp/server.go`
- [x] seq-mcp-create-presenter.md → `internal/mcp/server.go`
- [x] seq-mcp-receive-event.md → `internal/mcp/tools.go`
- [x] seq-mcp-run.md → `internal/mcp/tools.go`
- [x] seq-mcp-get-state.md → `internal/mcp/resources.go`
- [ ] seq-mcp-state-wait.md → `internal/mcp/server.go` (handleWait, notifyStateChange)

### Test Designs
- [ ] test-MCP.md → `tools_test.go`

## Systems

### MCP Integration System
AI assistant integration via Model Context Protocol
- crc-MCPServer.md, crc-MCPResource.md, crc-MCPTool.md
- seq-mcp-lifecycle.md, seq-mcp-create-session.md, seq-mcp-create-presenter.md
- seq-mcp-receive-event.md, seq-mcp-run.md, seq-mcp-get-state.md, seq-mcp-state-wait.md

### Transport System
Support multiple MCP transport modes:
- **Stdio** (`mcp` command): JSON-RPC 2.0 over stdin/stdout
- **SSE** (`serve` command): Server-Sent Events over HTTP

### HTTP Endpoints (MCP port)
Debug and inspect runtime state:
- `GET /wait`: Long-poll for mcp.pushState() events
- `GET /variables`: Interactive variable tree view
- `GET /state`: Current session state JSON

## Gaps

### Unimplemented
- [ ] handleWait() - GET /wait long-poll endpoint
- [ ] notifyStateChange() - signal waiting HTTP clients
- [ ] atomicSwapQueue() - atomically swap mcp.state
- [ ] SafeExecuteInSession() - panic recovery wrapper
- [ ] triggerBrowserUpdate() - push state changes to browsers
- [ ] handleInstall() - bundled file installation

### Test Coverage
- [ ] Test code needs to be aligned with test-MCP.md design
