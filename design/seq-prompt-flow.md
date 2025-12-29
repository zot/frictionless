# Sequence: Permission Prompt Flow (Viewdef-Based)

**Source Spec:** prompt-ui.md

## Participants

- ClaudeCode: AI assistant requesting permission
- PermissionHook: Shell script intercepting permission requests
- PromptHTTPServer: Background HTTP server for prompt API
- Server: ui-engine Server with Prompt() method
- PromptManager: Tracks pending prompts with response channels
- LuaRuntime: Executes Lua code, hosts _G.promptResponse
- Browser: User's web browser rendering viewdef
- PromptViewdef: HTML template with variable bindings

## Sequence

```
+----------+     +--------------+     +----------------+     +------+     +-------------+     +----------+     +-------+     +-------------+
|ClaudeCode|     |PermissionHook|     |PromptHTTPServer|     |Server|     |PromptManager|     |LuaRuntime|     |Browser|     |PromptViewdef|
+----+-----+     +------+-------+     +-------+--------+     +--+---+     +------+------+     +----+-----+     +---+---+     +------+------+
     |                  |                     |                 |                |                  |              |                |
     | PermissionRequest|                     |                 |                |                  |              |                |
     | (stdin JSON)     |                     |                 |                |                  |              |                |
     |----------------->|                     |                 |                |                  |              |                |
     |                  |                     |                 |                |                  |              |                |
     |                  |---+                 |                 |                |                  |              |                |
     |                  |   | ReadMCPPort()   |                 |                |                  |              |                |
     |                  |<--+ .ui-mcp/mcp-port|                 |                |                  |              |                |
     |                  |                     |                 |                |                  |              |                |
     |                  | POST /api/prompt    |                 |                |                  |              |                |
     |                  | {message,options}   |                 |                |                  |              |                |
     |                  |-------------------->|                 |                |                  |              |                |
     |                  |                     |                 |                |                  |              |                |
     |                  |                     | Prompt(session, |                |                  |              |                |
     |                  |                     | msg, options)   |                |                  |              |                |
     |                  |                     |---------------->|                |                  |              |                |
     |                  |                     |                 |                |                  |              |                |
     |                  |                     |                 | CreatePrompt() |                  |              |                |
     |                  |                     |                 |--------------->|                  |              |                |
     |                  |                     |                 |                |                  |              |                |
     |                  |                     |                 |  id, channel   |                  |              |                |
     |                  |                     |                 |<- - - - - - - -|                  |              |                |
     |                  |                     |                 |                |                  |              |                |
     |                  |                     |                 | Execute Lua:   |                  |              |                |
     |                  |                     |                 | app.pendingPrompt={...}          |              |                |
     |                  |                     |                 | app._presenter="Prompt"          |              |                |
     |                  |                     |                 |--------------------------------->|              |                |
     |                  |                     |                 |                |                  |              |                |
     |                  |                     |                 |                |                  | Variable     |                |
     |                  |                     |                 |                |                  | update event |                |
     |                  |                     |                 |                |                  |------------->|                |
     |                  |                     |                 |                |                  |              |                |
     |                  |                     |                 |                |                  |              | Render         |
     |                  |                     |                 |                |                  |              | Prompt.DEFAULT |
     |                  |                     |                 |                |                  |              |--------------->|
     |                  |                     |                 |                |                  |              |                |
     |                  |                     |                 |                |                  |              |  +============+|
     |                  |                     |                 |                |                  |              |  | User sees  ||
     |                  |                     |                 |                |                  |              |  | dialog,    ||
     |                  |                     |                 |                |                  |              |  | clicks btn ||
     |                  |                     |                 |                |                  |              |  +============+|
     |                  |                     |                 |                |                  |              |                |
     |                  |                     |                 |                |                  |              | ui-action:     |
     |                  |                     |                 |                |                  |              | respondToPrompt|
     |                  |                     |                 |                |                  |<-------------|                |
     |                  |                     |                 |                |                  |              |                |
     |                  |                     |                 |                |                  | Execute Lua: |                |
     |                  |                     |                 |                |                  | app:respondToPrompt(opt)      |
     |                  |                     |                 |                |                  |---+          |                |
     |                  |                     |                 |                |                  |   |          |                |
     |                  |                     |                 |                |                  |<--+          |                |
     |                  |                     |                 |                |                  |              |                |
     |                  |                     |                 |                |                  | _G.promptResponse(id,val,lbl) |
     |                  |                     |                 |                | Respond(id,resp) |              |                |
     |                  |                     |                 |                |<-----------------|              |                |
     |                  |                     |                 |                |                  |              |                |
     |                  |                     |                 | channel recv   |                  |              |                |
     |                  |                     |                 |<- - - - - - - -|                  |              |                |
     |                  |                     |                 |                |                  |              |                |
     |                  |                     |    response     |                |                  |              |                |
     |                  |                     |<- - - - - - - - |                |                  |              |                |
     |                  |                     |                 |                |                  |              |                |
     |                  | HTTP 200            |                 |                |                  |              |                |
     |                  | {value, label}      |                 |                |                  |              |                |
     |                  |<- - - - - - - - - - |                 |                |                  |              |                |
     |                  |                     |                 |                |                  |              |                |
     |                  |---+                 |                 |                |                  |              |                |
     |                  |   | LogDecision()   |                 |                |                  |              |                |
     |                  |<--+ permissions.log |                 |                |                  |              |                |
     |                  |                     |                 |                |                  |              |                |
     |                  |---+                 |                 |                |                  |              |                |
     |                  |   | MapToDecision() |                 |                |                  |              |                |
     |                  |<--+                 |                 |                |                  |              |                |
     |                  |                     |                 |                |                  |              |                |
     |{decision: allow} |                     |                 |                |                  |              |                |
     |<- - - - - - - - -|                     |                 |                |                  |              |                |
+----+-----+     +------+-------+     +-------+--------+     +--+---+     +------+------+     +----+-----+     +---+---+     +------+------+
|ClaudeCode|     |PermissionHook|     |PromptHTTPServer|     |Server|     |PromptManager|     |LuaRuntime|     |Browser|     |PromptViewdef|
+----------+     +--------------+     +----------------+     +------+     +-------------+     +----------+     +-------+     +-------------+
```

## Notes

### Key Design Points

- **No custom WebSocket messages**: Uses standard viewdef variable binding
- **Presenter switching**: `app._presenter = "Prompt"` triggers viewdef switch
- **Lua callback bridge**: `_G.promptResponse` connects Lua to Go channels
- **Blocking HTTP**: Server.Prompt() blocks until channel receives response

### Error Conditions

- **MCP port file not found**: Hook exits 0 (falls back to terminal)
- **HTTP timeout**: PromptManager closes channel with timeout error
- **Browser not connected**: Variable update has no recipient (prompt times out)
- **Invalid prompt ID**: Response ignored (logged)

### Timeout Handling

- Server.Prompt() starts timer on WaitForResponse
- On timeout: Cancel() closes channel, Server.Prompt() returns error
- HTTP handler returns timeout error to hook
- Hook script exits 0 (falls back to terminal)
