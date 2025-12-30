# PromptManager

**Source Spec:** prompt-ui.md

**Implementation:** internal/mcp/prompt.go

## Responsibilities

### Knows
- pending: Map of prompt ID to pendingPrompt (response channel)
- server: Reference to ui-engine Server for ExecuteInSession
- runtime: Reference to LuaRuntime for code execution
- mu: Mutex for thread-safe access

### Does
- Prompt: Generate unique ID, create response channel, set prompt in Lua, block until response or timeout
- Respond: Receive response (called from Lua callback), send to channel
- setPromptInLua: Execute Lua to set mcp.value with prompt data
- clearPromptInLua: Execute Lua to clear mcp.value on timeout

## Collaborators

- MCPServer: handlePrompt() calls Prompt()
- LuaRuntime: mcp.promptResponse callback calls Respond()
- Server: ExecuteInSession for Lua execution

## Sequences

- seq-prompt-flow.md: Full prompt lifecycle from hook to response
