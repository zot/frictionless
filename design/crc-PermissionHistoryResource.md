# PermissionHistoryResource

**Source Spec:** prompt-ui.md

**Implementation:** internal/mcp/resources.go (handleGetPermissionsHistoryResource)

## Responsibilities

### Knows
- uri: "ui://permissions/history"
- baseDir: Server base directory for log path (.ui-mcp/permissions.log)

### Does
- GetContent: Read permissions.log, parse JSONL, return last 50 decisions with analysis hints

## Collaborators

- MCPServer: Registers this resource via registerResources()
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

Resource content includes analysis_hints for Claude:
```json
{
  "decisions": [...],
  "analysis_hints": "Analyze patterns to proactively improve the permission UI..."
}
```

Analysis patterns suggested:
- Frequent "Always allow X" -> suggest as default
- Consistent allow for patterns -> offer auto-allow
- Custom options usage tracking
