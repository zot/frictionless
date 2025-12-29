# PromptOption (Lua)

**Source Spec:** specs/prompt-ui.md

**Implementation:** web/lua/prompt.lua

## Responsibilities

### Knows
- label: Display text for the option button
- value: Machine-readable value returned on selection
- _prompt: Reference to parent Prompt instance

### Does
- respond(): Zero-argument method for viewdef buttons; calls parent Prompt:respondWith()

## Collaborators

- Prompt: Parent that contains this option; receives response via respondWith()
- PromptViewdef: Renders options as buttons with ui-click="respond()"

## Sequences

- seq-prompt-flow.md: User clicks option button → respond() → Prompt:respondWith()
