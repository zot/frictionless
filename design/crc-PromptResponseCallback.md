# PromptResponseCallback

**Source Spec:** prompt-ui.md

## Responsibilities

### Knows
- promptManager: Reference to PromptManager (closure capture)

### Does
- HandleResponse: Called from Lua as _G.promptResponse(id, value, label), signals waiting channel

## Collaborators

- LuaRuntime: Registered as global function
- PromptManager: Calls Respond() to signal waiting goroutine
- PromptViewdef: Lua app:respondToPrompt() calls this callback

## Sequences

- seq-prompt-flow.md: Callback bridges Lua response to Go channel

## Notes

This is a Go function registered in Lua as `_G.promptResponse`. When the user clicks a button in the Prompt viewdef, the Lua action handler calls this function to signal the waiting HTTP request.

```go
// Registration
runtime.SetGlobal("promptResponse", func(id, value, label string) {
    promptManager.Respond(id, value, label)
})
```

```lua
-- Called from app:respondToPrompt(option)
_G.promptResponse(self.pendingPrompt.id, option.value, option.label)
```
