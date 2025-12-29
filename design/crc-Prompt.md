# Prompt (Lua)

**Source Spec:** specs/prompt-ui.md

**Implementation:** web/lua/prompt.lua

## Responsibilities

### Knows
- id: Unique prompt identifier for response correlation
- message: Description of what permission is being requested
- options: Array of PromptOption instances

### Does
- new(tbl): Create Prompt instance, wrapping raw options in PromptOption instances
- respondWith(value, label): Call mcp.promptResponse(), clear pendingPrompt, restore previous presenter

## Collaborators

- PromptOption: Child objects representing selectable options
- mcp: Go-provided global for sending response back to prompt API
- session: Access to app state for clearing pendingPrompt
- PromptViewdef: Renders prompt UI with message and option buttons

## Sequences

- seq-prompt-flow.md: Go sets app.pendingPrompt = Prompt:new(...), user clicks option, respondWith() called

## Notes

Uses prototype pattern following ui-engine conventions:
- `Prompt.__index = Prompt`
- `Prompt:new(tbl)` creates instances via `setmetatable(tbl, self)`
- Instance methods use `self` receiver
