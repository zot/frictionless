# PromptResponseCallback

**Source Spec:** prompt-ui.md

**Implementation:** internal/mcp/tools.go (registered in setupMCPGlobal)

## Responsibilities

### Knows
- promptManager: Reference to PromptManager (closure capture via MCPServer)

### Does
- HandleResponse: Called from Lua as mcp.promptResponse(id, value, label), signals waiting channel

## Collaborators

- LuaRuntime: Registered on mcp table
- PromptManager: Calls Respond() to signal waiting goroutine
- PromptOption: Lua option:respond() calls this callback

## Sequences

- seq-prompt-flow.md: Callback bridges Lua response to Go channel

## Notes

This is a Go function registered in Lua on the `mcp` table. When the user clicks a button in the MCP viewdef, the Lua respond() method calls this function to signal the waiting HTTP request.

```go
// Registration in setupMCPGlobal()
L.SetField(mcpTable, "promptResponse", L.NewFunction(func(L *lua.LState) int {
    id := L.CheckString(1)
    value := L.CheckString(2)
    label := L.CheckString(3)
    s.promptManager.Respond(id, value, label)
    return 0
}))
```

```lua
-- Called from option:respond() in setPromptInLua
respond = function(self)
    mcp.promptResponse(promptId, self.value, self.label)
    mcp.value = nil
end
```
