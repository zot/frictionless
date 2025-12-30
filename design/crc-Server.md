# Server (Prompt Integration)

**Source Spec:** prompt-ui.md

**Implementation:** internal/mcp/server.go (MCPServer contains PromptManager)

## Responsibilities

### Knows
- promptManager: Reference to PromptManager for prompt lifecycle
- uiServer: Reference to ui-engine Server for ExecuteInSession
- runtime: Reference to Lua runtime for callback registration

### Does
- handlePrompt: HTTP handler that triggers prompt flow
- setupMCPGlobal: Register mcp.promptResponse callback in Lua

## Collaborators

- PromptManager: Prompt() creates prompts and waits for responses
- LuaRuntime: mcp.promptResponse callback bridges Lua to Go
- PromptViewdef: Renders prompt UI based on mcp.value

## Sequences

- seq-prompt-flow.md: handlePrompt() orchestrates the prompt flow

## Notes

Prompt functionality is integrated into MCPServer rather than extending ui-engine Server:
1. handlePrompt() receives POST /api/prompt request
2. Calls promptManager.Prompt() which sets mcp.value in Lua
3. mcp.value triggers viewdef update showing prompt dialog
4. User clicks option, Lua calls mcp.promptResponse()
5. Go callback signals channel, handlePrompt() returns response
