# ViewdefsResource

**Source Spec:** prompt-ui.md

## Responsibilities

### Knows
- uri: "ui://viewdefs"
- viewdefDir: Path to viewdef directory (.ui-mcp/viewdefs/)

### Does
- GetContent: Return markdown describing viewdef locations and editing tips

## Collaborators

- MCPServer: Registers this resource
- AIAgent: Reads resource to discover customizable viewdefs

## Sequences

- (Static resource, no sequence needed)

## Notes

This MCP resource informs Claude about editable viewdefs, enabling conversational UI customization.

Content:
```markdown
# UI MCP Viewdefs

Viewdefs are HTML templates that control how UI elements are rendered.
You can edit these files to customize the appearance and behavior.

## Locations

- `.ui-mcp/viewdefs/Prompt.DEFAULT.html` - Permission prompt dialog

## Editing Tips

- Use Shoelace components (sl-button, sl-dialog, sl-input, etc.)
- Use `ui-value="path"` for data binding
- Use `ui-action="method()"` for button actions
- Changes take effect on next render
```
