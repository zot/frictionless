# Prompt.DEFAULT Viewdef

**Source**: prompt-ui.md
**Route**: N/A (presenter switch via app._presenter)

## Data

See `crc-PromptViewdef.md`, `crc-Server.md`

```
pendingPrompt.id: string       -- Unique prompt identifier
pendingPrompt.message: string  -- Permission request message
pendingPrompt.options: array   -- [{label, value}, ...]
```

## Layout

```
+----------------------------------------------------------+
|                     (overlay dims page)                  |
|  +----------------------------------------------------+  |
|  |  sl-dialog [open]                                  |  |
|  |   +--------------------------------------------+   |  |
|  |   |                                            |   |  |
|  |   |  {pendingPrompt.message}                   |   |  |
|  |   |                                            |   |  |
|  |   +--------------------------------------------+   |  |
|  |                                                    |  |
|  |   ui-viewlist="pendingPrompt.options"              |  |
|  |   +----------------+  +----------------+  +------+ |  |
|  |   | {label}        |  | {label}        |  |{label}||  |
|  |   | [sl-button]    |  | [sl-button]    |  |      | |  |
|  |   +----------------+  +----------------+  +------+ |  |
|  |                                                    |  |
|  +----------------------------------------------------+  |
|                                                          |
+----------------------------------------------------------+
```

## Template

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

## Events

See `crc-PromptViewdef.md`

- `ui-action="respondToPrompt(_)"`: Calls `app:respondToPrompt(option)` in Lua
  - Lua calls `_G.promptResponse(id, value, label)`
  - Clears `app.pendingPrompt`
  - Switches `app._presenter` back to previous presenter

## CSS Classes

- `prompt-overlay`: Full-screen backdrop (optional styling)

## Behavior

1. When `app._presenter = "Prompt"`:
   - Viewdef renders with current `app.pendingPrompt` data
   - Shoelace dialog appears open

2. User clicks button:
   - `ui-action` triggers Lua method with option object
   - `_G.promptResponse` signals Go channel
   - Presenter switches back, dialog disappears

3. No close button or escape - user must choose an option

## Customization

Location: `.ui-mcp/viewdefs/Prompt.DEFAULT.html`

User-editable. Claude can modify via conversation:
- "Make buttons larger"
- "Add dark theme"
- "Show full command in message"
- "Add 'Remember for this directory' option"

Changes take effect on next prompt render.
