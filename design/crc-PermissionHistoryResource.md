# PermissionHistoryResource

**Source Spec:** prompt-ui.md

## Responsibilities

### Knows
- uri: "ui://permissions/history"
- logPath: Path to permissions log (.ui-mcp/permissions.log)

### Does
- GetContent: Read permissions.log, return JSON array of recent decisions

## Collaborators

- MCPServer: Registers this resource
- AIAgent: Reads resource to analyze permission patterns
- PermissionHook: Appends decisions to log file

## Sequences

- (Resource read, no sequence needed)

## Notes

This MCP resource enables Claude to proactively improve the permission UI based on user patterns.

Log entry format (JSONL):
```json
{"tool": "Bash", "input": {"command": "grep ..."}, "choice": "allow_session", "timestamp": "2025-01-01T12:00:00Z"}
```

Resource content:
```json
{
  "decisions": [
    {"tool": "Bash", "command": "grep ...", "choice": "allow_session", "timestamp": "..."},
    {"tool": "Read", "path": "/home/...", "choice": "allow", "timestamp": "..."}
  ]
}
```

Resource description suggests analysis patterns:
- Frequent "Always allow X" -> suggest as default
- Consistent allow for patterns -> offer auto-allow
- Custom options usage tracking
