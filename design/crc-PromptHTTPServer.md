# PromptHTTPServer

**Source Spec:** prompt-ui.md

**Implementation:** internal/mcp/server.go (integrated into MCPServer)

## Responsibilities

### Knows
- mcpPort: HTTP server port (randomly assigned)
- httpServer: The HTTP server instance
- baseDir: Path for port file (.ui-mcp/mcp-port)

### Does
- StartPromptServer: Bind to random port, return port number
- WritePortFile: Write port to mcp-port file
- RemovePortFile: Clean up port file on shutdown
- ShutdownPromptServer: Gracefully stop HTTP server
- handlePrompt: POST /api/prompt - parse request, call PromptManager.Prompt(), return response
- handleDebugVariables: GET /debug/variables - render variable tree
- handleDebugState: GET /debug/state - render state JSON

## Collaborators

- PromptManager: handlePrompt() calls Prompt()
- MCPServer: Coordinates startup during runMCP
- OS: Port binding, file writing

## Sequences

- seq-prompt-flow.md: HTTP endpoint receives hook request
- seq-prompt-server-startup.md: Server startup includes prompt HTTP server
