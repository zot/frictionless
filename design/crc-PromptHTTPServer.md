# PromptHTTPServer

**Source Spec:** prompt-ui.md

## Responsibilities

### Knows
- port: HTTP server port (randomly assigned)
- portFilePath: Path to write port file (.ui-mcp/mcp-port)
- server: Reference to ui-engine Server

### Does
- Start: Bind to random port, write port file, begin serving
- Stop: Remove port file, shutdown HTTP server
- HandlePromptAPI: POST /api/prompt - parse request, call Server.Prompt(), return response

## Collaborators

- Server: Calls Server.Prompt() to trigger prompt flow
- MCPServer: Coordinates startup during ui_start
- OS: Port binding, file writing

## Sequences

- seq-prompt-flow.md: HTTP endpoint receives hook request
- seq-prompt-server-startup.md: Server startup includes prompt HTTP server
