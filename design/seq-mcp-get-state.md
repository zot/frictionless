# Sequence: MCP Get State

**Source Spec:** mcp.md (MCP Tools)

## Participants
- AI Agent: External AI assistant
- MCPServer: MCP interface for the UI server
- MCPTool (ui_get_state): Tool handler for state inspection
- LuaRuntime: Embedded Lua VM manager
- LuaSession: Per-session Lua state
- VariableStore: Backend for variable storage
- ChangeTracker: Tracker for session variable state

## Scenario: AI inspects current application state
```
     ┌────────┐                             ┌─────────┐            ┌──────────────────────┐           ┌──────────┐           ┌──────────┐           ┌─────────────┐          ┌─────────────┐
     │AI Agent│                             │MCPServer│            │MCPTool (ui_get_state)│           │LuaRuntime│           │LuaSession│           │VariableStore│          │ChangeTracker│
     └────┬───┘                             └────┬────┘            └───────────┬──────────┘           └─────┬────┘           └─────┬────┘           └──────┬──────┘          └──────┬──────┘
          │CallTool("ui_get_state", {sessionId}) │                             │                            │                      │                       │                        │       
          │─────────────────────────────────────>│                             │                            │                      │                       │                        │       
          │                                      │                             │                            │                      │                       │                        │       
          │                                      │Handle("ui_get_state", args) │                            │                      │                       │                        │       
          │────────────────────────────>│                            │                      │                       │                        │       
          │                                      │                             │                            │                      │                       │                        │       
          │                                      │                             │ GetLuaSession(sessionId)   │                      │                       │                        │       
          │                                      │                             │───────────────────────────>│                      │                       │                        │       
          │                                      │                             │                            │                      │                       │                        │       
          │                                      │                             │          session           │                      │                       │                        │       
          │                                      │                             │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                      │                       │                        │       
          │                                      │                             │                            │                      │                       │                        │       
          │                                      │                             │                   GetTracker()                    │                       │                        │       
          │                                      │                             │──────────────────────────────────────────────────>│                       │                        │       
          │                                      │                             │                            │                      │                       │                        │       
          │                                      │                             │                     tracker│                      │                       │                        │       
          │                                      │                             │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │                       │                        │       
          │                                      │                             │                            │                      │                       │                        │       
          │                                      │                             │                            │             GetVariable(1)                   │                        │       
          │                                      │                             │───────────────────────────────────────────────────────────────────────────────────────────────────>│       
          │                                      │                             │                            │                      │                       │                        │       
          │                                      │                             │                            │                   v1 │                       │                        │       
          │                                      │                             │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│       
          │                                      │                             │                            │                      │                       │                        │       
          │                                      │                             │                            │       ToValueJSONBytes(v1.Value)             │                        │       
          │                                      │                             │───────────────────────────────────────────────────────────────────────────────────────────────────>│       
          │                                      │                             │                            │                      │                       │                        │       
          │                                      │                             │                            │                jsonBytes                     │                        │       
          │                                      │                             │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│       
          │                                      │                             │                            │                      │                       │                        │       
          │                                      │   ToolResult(jsonBytes)     │                            │                      │                       │                        │       
          │                                      │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │                            │                      │                       │                        │       
          │                                      │                             │                            │                      │                       │                        │       
          │             ToolResult               │                             │                            │                      │                       │                        │       
          │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                             │                            │                      │                       │                        │       
     ┌────┴───┐                             ┌────┴────┐            ┌───────────┴──────────┐           ┌─────┴────┐           ┌─────┴────┐           ┌──────┴──────┐          ┌──────┴──────┐
     │AI Agent│                             │MCPServer│            │MCPTool (ui_get_state)│           │LuaRuntime│           │LuaSession│           │VariableStore│          │ChangeTracker│
     └────────┘                             └─────────┘            └──────────────────────┘           └──────────┘           └──────────┘           └─────────────┘          └─────────────┘
```
