# PermissionHook

**Source Spec:** prompt-ui.md

## Responsibilities

### Knows
- mcpPortPath: Path to MCP port file (.ui-mcp/mcp-port)
- permissionsLogPath: Path to permissions log (.ui-mcp/permissions.log)

### Does
- ReadInput: Parse hook JSON from stdin (tool_name, tool_input)
- ReadMCPPort: Read port from file, exit 0 if not found (fallback to terminal)
- BuildPromptRequest: Construct API request JSON with message, options, timeout
- CallPromptAPI: POST to http://127.0.0.1:{port}/api/prompt
- LogDecision: Append decision to permissions.log for pattern analysis
- ReturnDecision: Output hook decision JSON to stdout

## Collaborators

- PromptHTTPServer: HTTP endpoint target
- ClaudeCode: Invokes hook, receives decision
- PermissionHistory: Log file read by MCP resource

## Sequences

- seq-prompt-flow.md: Hook script initiates prompt flow

## Notes

Fallback behavior: If MCP port file doesn't exist or HTTP call fails, script exits with 0 to let Claude fall back to terminal prompts.

Decision mapping:
- "allow" or "allow_session" -> `{"decision": "allow"}`
- "deny" -> `{"decision": "deny", "message": "User denied permission"}`
