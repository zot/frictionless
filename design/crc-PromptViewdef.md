# PromptViewdef

**Source Spec:** prompt-ui.md

## Responsibilities

### Knows
- Template HTML with Shoelace dialog
- Data bindings: ui-value="pendingPrompt.message", ui-viewlist="pendingPrompt.options"
- Action binding: ui-action="respondToPrompt(_)"

### Does
- RenderPromptDialog: Display modal with message and option buttons
- BindOptions: Render button for each option using ui-viewlist
- HandleButtonClick: Trigger app:respondToPrompt(option) via ui-action

## Collaborators

- VariableStore: Reads app.pendingPrompt for rendering
- LuaRuntime: Action calls app:respondToPrompt() method

## Sequences

- seq-prompt-flow.md: Viewdef renders prompt and handles user click

## Notes

Location: `.ui-mcp/viewdefs/Prompt.DEFAULT.html`

This is a standard viewdef that uses ui-engine's variable binding system. No custom WebSocket messages needed.

```html
<template>
  <div class="prompt-overlay">
    <sl-dialog open label="Permission Request">
      <p ui-value="pendingPrompt.message"></p>
      <div ui-viewlist="pendingPrompt.options">
        <sl-button ui-action="respondToPrompt(_)">
          <span ui-value="label"></span>
        </sl-button>
      </div>
    </sl-dialog>
  </div>
</template>
```

The viewdef is user-editable, allowing Claude to customize the prompt UI via conversation.
