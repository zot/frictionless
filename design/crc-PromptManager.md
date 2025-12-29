# PromptManager

**Source Spec:** prompt-ui.md

## Responsibilities

### Knows
- pendingPrompts: Map of prompt ID to response channel
- mu: Mutex for thread-safe access

### Does
- CreatePrompt: Generate unique ID, create response channel, store in map
- WaitForResponse: Block on channel until response or timeout
- Respond: Receive response (called from Lua callback), send to channel, remove from map
- Cancel: Close channel with timeout error, remove from map

## Collaborators

- Server: Server.Prompt() creates prompts and waits for responses
- LuaRuntime: _G.promptResponse callback calls Respond()

## Sequences

- seq-prompt-flow.md: Full prompt lifecycle from hook to response
