# Sequence: Prompt Server Startup

**Source Spec:** prompt-ui.md

## Participants

- AIAgent: External AI assistant (MCP Client)
- MCPServer: Protocol handler and lifecycle FSM
- MCPTool: Tool handler dispatcher
- PromptHTTPServer: Background HTTP server for prompt API
- PromptManager: Tracks pending prompts
- Server: ui-engine Server
- LuaRuntime: Lua execution environment
- OS: Operating system services

## Sequence

Extends seq-mcp-lifecycle.md Scenario 2 (Server Startup) to include prompt server initialization.

```
+-------+     +---------+     +-------+     +----------------+     +-------------+     +------+     +----------+     +--+
|AIAgent|     |MCPServer|     |MCPTool|     |PromptHTTPServer|     |PromptManager|     |Server|     |LuaRuntime|     |OS|
+---+---+     +----+----+     +---+---+     +-------+--------+     +------+------+     +--+---+     +----+-----+     ++-+
    |              |              |                 |                     |               |              |            |
    | ui_start     |              |                 |                     |               |              |            |
    |------------->|              |                 |                     |               |              |            |
    |              |              |                 |                     |               |              |            |
    |              | Handle       |                 |                     |               |              |            |
    |              |------------->|                 |                     |               |              |            |
    |              |              |                 |                     |               |              |            |
    |              |              |                                New()  |               |              |            |
    |              |              |-------------------------------------->|               |              |            |
    |              |              |                 |                     |               |              |            |
    |              |              |                 |         promptManager               |              |            |
    |              |              |<- - - - - - - - - - - - - - - - - - - |               |              |            |
    |              |              |                 |                     |               |              |            |
    |              |              | New(promptMgr)  |                     |               |              |            |
    |              |              |---------------->|                     |               |              |            |
    |              |              |                 |                     |               |              |            |
    |              |              |promptHTTPServer |                     |               |              |            |
    |              |              |<- - - - - - - - |                     |               |              |            |
    |              |              |                 |                     |               |              |            |
    |              |              | Start()         |                     |               |              |            |
    |              |              |---------------->|                     |               |              |            |
    |              |              |                 |                     |               |              |            |
    |              |              |                 |                  ListenOnRandomPort()              |            |
    |              |              |                 |----------------------------------------------------->|          |
    |              |              |                 |                     |               |              |            |
    |              |              |                 |                     |               |         port |            |
    |              |              |                 |<- - - - - - - - - - - - - - - - - - - - - - - - - - |            |
    |              |              |                 |                     |               |              |            |
    |              |              |                 |             WriteFile(.ui-mcp/mcp-port)            |            |
    |              |              |                 |----------------------------------------------------->|          |
    |              |              |                 |                     |               |              |            |
    |              |              | RegisterPromptCallback(promptMgr)     |               |              |            |
    |              |              |----------------------------------------------------------------------->|          |
    |              |              |                 |                     |               |              |            |
    |              |              |                 |                     |               |    Register  |            |
    |              |              |                 |                     |               |    _G.promptResponse      |
    |              |              |                 |                     |               |<-------------|            |
    |              |              |                 |                     |               |              |            |
    |              |              |---+             |                     |               |              |            |
    |              |              |   | StartUI()   |                     |               |              |            |
    |              |              |<--+             |                     |               |              |            |
    |              |              |                 |                     |               |              |            |
    | Success(url) |              |                 |                     |               |              |            |
    |<- - - - - - -|              |                 |                     |               |              |            |
+---+---+     +----+----+     +---+---+     +-------+--------+     +------+------+     +--+---+     +----+-----+     ++-+
|AIAgent|     |MCPServer|     |MCPTool|     |PromptHTTPServer|     |PromptManager|     |Server|     |LuaRuntime|     |OS|
+-------+     +---------+     +-------+     +----------------+     +-------------+     +------+     +----------+     +--+
```

## Notes

- PromptHTTPServer runs on separate port from UI server
- Port file enables hook script to discover prompt API endpoint
- `_G.promptResponse` callback registered before any prompts can occur
- Both servers must be running for full prompt flow to work

### Shutdown

On shutdown:
1. Stop PromptHTTPServer (removes mcp-port file)
2. Stop UI server
3. Cancel any pending prompts
