# Prompt (Lua Data)

**Source Spec:** specs/prompt-ui.md

**Implementation:** internal/mcp/prompt.go (setPromptInLua creates the Lua table)

## Responsibilities

### Knows
- id: Unique prompt identifier for response correlation
- message: Description of what permission is being requested
- options: Array of option objects with label, value, and respond()
- isPrompt: Boolean flag to trigger viewdef visibility

### Does
- (Created by Go): PromptManager.setPromptInLua() creates mcp.value table

## Collaborators

- PromptManager: Creates prompt table in Lua via setPromptInLua()
- mcp.value: Bound to viewdef for rendering
- PromptViewdef: Renders prompt UI with message and option buttons

## Sequences

- seq-prompt-flow.md: Go sets mcp.value = {...}, user clicks option, option:respond() called

## Notes

Prompt data is created inline in Lua by setPromptInLua():
```lua
mcp.value = {
  isPrompt = true,
  id = promptId,
  message = "Claude wants to use: Bash",
  options = {
    {type = "PromptOption", label = "Allow once", value = "allow", respond = function(self) ... end},
    ...
  }
}
```

No separate Lua class file - structure is created dynamically by Go.
