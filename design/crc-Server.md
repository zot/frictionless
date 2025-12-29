# Server (ui-engine extension)

**Source Spec:** prompt-ui.md

## Responsibilities

### Knows
- promptManager: Reference to PromptManager
- luaRuntime: Reference to Lua runtime for state manipulation

### Does
- Prompt: Set app.pendingPrompt in Lua, switch presenter to "Prompt", block on channel until response
- RegisterPromptCallback: Register _G.promptResponse callback in Lua runtime

## Collaborators

- PromptManager: Creates prompts and waits for responses
- LuaRuntime: Sets app state and registers global callback
- PromptViewdef: Renders prompt UI based on app.pendingPrompt

## Sequences

- seq-prompt-flow.md: Server.Prompt() orchestrates the prompt flow

## Notes

This extends the existing ui-engine Server with prompt capabilities. The Prompt() method:
1. Generates prompt ID via PromptManager.CreatePrompt()
2. Executes Lua to set `app.pendingPrompt = {...}` and `app._presenter = "Prompt"`
3. Blocks on PromptManager.WaitForResponse(id, timeout)
4. Returns response when Lua callback signals channel
