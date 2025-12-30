# PromptViewdef

**Source Spec:** prompt-ui.md

**Implementation:**
- web/viewdefs/MCP.DEFAULT.html - Main MCP viewdef with prompt dialog
- web/viewdefs/PromptOption.DEFAULT.html - Option button viewdef

## Responsibilities

### Knows
- Template HTML with conditional prompt dialog
- Data bindings: ui-value="value.message", ui-viewlist="value.options"
- Visibility: ui-class-visible="value.isPrompt"

### Does
- RenderPromptDialog: Display dialog when value.isPrompt is set
- BindOptions: Render PromptOption viewdef for each option via ui-viewlist
- HandleButtonClick: Trigger option:respond() via ui-action

## Collaborators

- VariableStore: Reads mcp.value for rendering
- PromptOption viewdef: Renders individual option buttons

## Sequences

- seq-prompt-flow.md: Viewdef renders prompt and handles user click

## Notes

The prompt UI uses the MCP.DEFAULT viewdef which shows a prompt dialog when `mcp.value.isPrompt` is set. Each option is rendered using PromptOption.DEFAULT viewdef.

MCP.DEFAULT.html structure:
```html
<div class="prompt-dialog" ui-class-visible="value.isPrompt">
  <h3>Permission Request</h3>
  <p class="prompt-message" ui-value="value.message"></p>
  <div class="prompt-options" ui-viewlist="value.options"></div>
</div>
```

PromptOption.DEFAULT.html:
```html
<button class="prompt-option" ui-value="label" ui-action="respond()"></button>
```

The viewdefs are user-editable, allowing Claude to customize the prompt UI via conversation.
