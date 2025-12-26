# Sequence: MCP Run Code

**Source Spec:** interfaces.md (MCP Tools)

## Participants
- AI Agent: External AI assistant
- MCPServer: MCP interface for the UI server
- MCPTool (ui_run): Tool handler for Lua execution
- LuaRuntime: Embedded Lua VM manager
- LuaSession: Per-session Lua state
- LuaExecutor: Background thread for thread-safe Lua execution

## Scenario: AI executes arbitrary Lua code
```
     ┌────────┐                             ┌─────────┐          ┌────────────────┐                  ┌──────────┐                                      ┌──────────┐           ┌───────────┐
     │AI Agent│                             │MCPServer│          │MCPTool (ui_run)│                  │LuaRuntime│                                      │LuaSession│           │LuaExecutor│
     └────┬───┘                             └────┬────┘          └────────┬───────┘                  └─────┬────┘                                      └─────┬────┘           └─────┬─────┘
          │CallTool("ui_run", {code, sessionId}) │                        │                                │                                                 │                      │      
          │─────────────────────────────────────>│                        │                                │                                                 │                      │      
          │                                      │                        │                                │                                                 │                      │      
          │                                      │Handle("ui_run", args)  │                                │                                                 │                      │      
          │                                      │───────────────────────>│                                │                                                 │                      │      
          │                                      │                        │                                │                                                 │                      │      
          │                                      │                        │ExecuteInSession(sessionId, fn) │                                                 │                      │      
          │                                      │                        │───────────────────────────────>│                                                 │                      │      
          │                                      │                        │                                │                                                 │                      │      
          │                                      │                        │                                │                               Queue(fn)         │                      │      
          │                                      │                        │                                │───────────────────────────────────────────────────────────────────────>│      
          │                                      │                        │                                │                                                 │                      │      
          │                                      │                        │                               ┌┴┐                             Execute fn()       │                      │      
          │                                      │                        │                               │ │ <─────────────────────────────────────────────────────────────────────│      
          │                                      │                        │                               │ │                                                │                      │      
          │                                      │                        │                               │ │ ────┐                                          │                      │      
          │                                      │                        │                               │ │     │ Lock sessions                            │                      │      
          │                                      │                        │                               │ │ <───┘                                          │                      │      
          │                                      │                        │                               │ │                                                │                      │      
          │                                      │                        │                               │ │                  Get Session                   │                      │      
          │                                      │                        │                               │ │ ──────────────────────────────────────────────>│                      │      
          │                                      │                        │                               │ │                                                │                      │      
          │                                      │                        │                               │ │ ────┐                                          │                      │      
          │                                      │                        │                               │ │     │ Unlock                                   │                      │      
          │                                      │                        │                               │ │ <───┘                                          │                      │      
          │                                      │                        │                               │ │                                                │                      │      
          │                                      │                        │                               │ │ ────┐                                          │                      │      
          │                                      │                        │                               │ │     │ L.SetGlobal("session", sessionTable)     │                      │      
          │                                      │                        │                               │ │ <───┘                                          │                      │      
          │                                      │                        │                               │ │                                                │                      │      
          │                                      │                        │                               │ │ ────┐                                          │                      │      
          │                                      │                        │                               │ │     │ L.DoString(code)                         │                      │      
          │                                      │                        │                               │ │ <───┘                                          │                      │      
          │                                      │                        │                               └┬┘                                                │                      │      
          │                                      │                        │                                │                            Result / Error       │                      │      
          │                                      │                        │                                │ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ >│      
          │                                      │                        │                                │                                                 │                      │      
          │                                      │                        │                                │                                Result           │                      │      
          │                                      │                        │                                │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│      
          │                                      │                        │                                │                                                 │                      │      
          │                                      │                        │            Result              │                                                 │                      │      
          │                                      │                        │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                                                 │                      │      
          │                                      │                        │                                │                                                 │                      │      
          │                                      │      ToolResult        │                                │                                                 │                      │      
          │                                      │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                                │                                                 │                      │      
          │                                      │                        │                                │                                                 │                      │      
          │             ToolResult               │                        │                                │                                                 │                      │      
          │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                        │                                │                                                 │                      │      
     ┌────┴───┐                             ┌────┴────┐          ┌────────┴───────┐                  ┌─────┴────┐                                      ┌─────┴────┐           ┌─────┴─────┐
     │AI Agent│                             │MCPServer│          │MCPTool (ui_run)│                  │LuaRuntime│                                      │LuaSession│           │LuaExecutor│
     └────────┘                             └─────────┘          └────────────────┘                  └──────────┘                                      └──────────┘           └───────────┘
```
