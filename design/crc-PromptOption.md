# PromptOption (Lua Data)

**Source Spec:** specs/prompt-ui.md

**Implementation:**
- internal/mcp/prompt.go (option created in setPromptInLua)
- web/viewdefs/PromptOption.DEFAULT.html (viewdef)

## Responsibilities

### Knows
- type: "PromptOption" for viewdef resolution
- label: Display text for the option button
- value: Machine-readable value returned on selection

### Does
- respond(): Zero-argument method for viewdef buttons; calls mcp.promptResponse() and clears mcp.value

## Collaborators

- PromptManager: Creates option with respond() method in setPromptInLua()
- PromptOption viewdef: Renders option as button with ui-action="respond()"

## Sequences

- seq-prompt-flow.md: User clicks option button -> respond() -> mcp.promptResponse()

## Notes

Option created inline in Lua by setPromptInLua():
```lua
local option = {
  type = "PromptOption",
  label = opt.label,
  value = opt.value,
  respond = function(self)
    mcp.promptResponse(promptId, self.value, self.label)
    mcp.value = nil
  end
}
```
