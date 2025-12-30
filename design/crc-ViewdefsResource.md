# ViewdefsResource

**Source Spec:** prompt-ui.md

**Implementation:** internal/mcp/resources.go (handleGetPromptViewdefsResource)

## Responsibilities

### Knows
- uri: "ui://prompt/viewdefs"
- baseDir: Server base directory for path construction

### Does
- GetContent: Return markdown describing viewdef locations and editing tips

## Collaborators

- MCPServer: Registers this resource via registerResources()
- AIAgent: Reads resource to discover customizable viewdefs

## Sequences

- (Static resource, no sequence needed)

## Notes

This MCP resource informs Claude about editable viewdefs, enabling conversational UI customization.

The resource dynamically constructs viewdef paths using baseDir:
- `{baseDir}/viewdefs/Prompt.DEFAULT.html` - Permission prompt dialog
- `{baseDir}/viewdefs/Feedback.DEFAULT.html` - Default app UI
