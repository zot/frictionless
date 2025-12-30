# Sequence: Permission Prompt Flow (Viewdef-Based)

**Source Spec:** prompt-ui.md

## Participants

- ClaudeCode: AI assistant requesting permission
- PermissionHook: Shell script intercepting permission requests
- MCPServer: Background HTTP server with handlePrompt()
- PromptManager: Tracks pending prompts with response channels
- LuaRuntime: Executes Lua code, hosts mcp.promptResponse
- Browser: User's web browser rendering viewdef
- MCPViewdef: MCP.DEFAULT.html with prompt dialog

## Sequence

```
+----------+     +--------------+     +---------+     +-------------+     +----------+     +-------+     +-----------+
|ClaudeCode|     |PermissionHook|     |MCPServer|     |PromptManager|     |LuaRuntime|     |Browser|     |MCP Viewdef|
+----+-----+     +------+-------+     +----+----+     +------+------+     +----+-----+     +---+---+     +-----+-----+
     |                  |                  |                |                  |              |                |
     | PermissionRequest|                  |                |                  |              |                |
     | (stdin JSON)     |                  |                |                  |              |                |
     |----------------->|                  |                |                  |              |                |
     |                  |                  |                |                  |              |                |
     |                  |---+              |                |                  |              |                |
     |                  |   | ReadMCPPort()|                |                  |              |                |
     |                  |<--+ mcp-port     |                |                  |              |                |
     |                  |                  |                |                  |              |                |
     |                  | POST /api/prompt |                |                  |              |                |
     |                  | {message,options}|                |                  |              |                |
     |                  |----------------->|                |                  |              |                |
     |                  |                  |                |                  |              |                |
     |                  |                  | Prompt(session,|                  |              |                |
     |                  |                  | msg, opts, tmo)|                  |              |                |
     |                  |                  |--------------->|                  |              |                |
     |                  |                  |                |                  |              |                |
     |                  |                  |                | id, channel      |              |                |
     |                  |                  |                |---+              |              |                |
     |                  |                  |                |   | Generate UUID|              |                |
     |                  |                  |                |<--+              |              |                |
     |                  |                  |                |                  |              |                |
     |                  |                  |                | setPromptInLua() |              |                |
     |                  |                  |                | mcp.value={      |              |                |
     |                  |                  |                |   isPrompt=true, |              |                |
     |                  |                  |                |   message=...,   |              |                |
     |                  |                  |                |   options=...}   |              |                |
     |                  |                  |                |----------------->|              |                |
     |                  |                  |                |                  |              |                |
     |                  |                  |                |                  | Variable     |                |
     |                  |                  |                |                  | update event |                |
     |                  |                  |                |                  |------------->|                |
     |                  |                  |                |                  |              |                |
     |                  |                  |                |                  |              | Render         |
     |                  |                  |                |                  |              | MCP.DEFAULT    |
     |                  |                  |                |                  |              |--------------->|
     |                  |                  |                |                  |              |                |
     |                  |                  |                |                  |              |  +============+|
     |                  |                  |                |                  |              |  | User sees  ||
     |                  |                  |                |                  |              |  | dialog,    ||
     |                  |                  |                |                  |              |  | clicks btn ||
     |                  |                  |                |                  |              |  +============+|
     |                  |                  |                |                  |              |                |
     |                  |                  |                |                  |              | ui-action:     |
     |                  |                  |                |                  |              | respond()      |
     |                  |                  |                |                  |<-------------|                |
     |                  |                  |                |                  |              |                |
     |                  |                  |                |                  | Execute Lua: |                |
     |                  |                  |                |                  | option:respond()              |
     |                  |                  |                |                  |---+          |                |
     |                  |                  |                |                  |   |          |                |
     |                  |                  |                |                  |<--+          |                |
     |                  |                  |                |                  |              |                |
     |                  |                  |                |                  | mcp.promptResponse(id,val,lbl)|
     |                  |                  |                | Respond(id,resp) |              |                |
     |                  |                  |                |<-----------------|              |                |
     |                  |                  |                |                  |              |                |
     |                  |                  | channel recv   |                  |              |                |
     |                  |                  |<- - - - - - - -|                  |              |                |
     |                  |                  |                |                  |              |                |
     |                  | HTTP 200         |                |                  |              |                |
     |                  | {value, label}   |                |                  |              |                |
     |                  |<- - - - - - - - -|                |                  |              |                |
     |                  |                  |                |                  |              |                |
     |                  |---+              |                |                  |              |                |
     |                  |   | LogDecision()|                |                  |              |                |
     |                  |<--+ permissions.log               |                  |              |                |
     |                  |                  |                |                  |              |                |
     |                  |---+              |                |                  |              |                |
     |                  |   | MapToDecision()               |                  |              |                |
     |                  |<--+              |                |                  |              |                |
     |                  |                  |                |                  |              |                |
     |{decision: allow} |                  |                |                  |              |                |
     |<- - - - - - - - -|                  |                |                  |              |                |
+----+-----+     +------+-------+     +----+----+     +------+------+     +----+-----+     +---+---+     +-----+-----+
|ClaudeCode|     |PermissionHook|     |MCPServer|     |PromptManager|     |LuaRuntime|     |Browser|     |MCP Viewdef|
+----------+     +--------------+     +---------+     +-------------+     +----------+     +-------+     +-----------+
```

## Notes

### Key Design Points

- **No custom WebSocket messages**: Uses standard viewdef variable binding
- **mcp.value binding**: `mcp.value.isPrompt` triggers prompt dialog visibility
- **Lua callback bridge**: `mcp.promptResponse` connects Lua to Go channels
- **Blocking HTTP**: MCPServer.handlePrompt() blocks until channel receives response
- **Option respond()**: Each option has respond() method that calls mcp.promptResponse and clears mcp.value

### Error Conditions

- **MCP port file not found**: Hook exits 0 (falls back to terminal)
- **HTTP timeout**: PromptManager returns timeout error, handlePrompt returns HTTP error
- **Browser not connected**: Variable update has no recipient (prompt times out)
- **Invalid prompt ID**: Response ignored (logged)

### Timeout Handling

- PromptManager.Prompt() starts timer via context.WithTimeout
- On timeout: clearPromptInLua() called, error returned
- HTTP handler returns timeout error to hook
- Hook script exits 0 (falls back to terminal)
