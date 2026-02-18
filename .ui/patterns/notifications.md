---
description: Show dismissable toast notifications using the MCP shell's built-in notification system. Supports Shoelace alert variants (success, warning, danger, primary, neutral).
---

# Notifications

Show toast-style alerts that the user can dismiss. Notifications are managed by the MCP shell and rendered in the `.mcp-notifications` container.

## When to Use

- Confirming an action succeeded ("Saved", "Up to date")
- Warning the user about a problem ("Fetch failed — try the bookmarklet")
- Showing transient status messages that don't need persistent UI

## Usage

From any app's Lua code:

```lua
mcp:notify("message text", "variant")
```

### Variants

| Variant | Use For | Color |
|---------|---------|-------|
| `"success"` | Confirmations, positive outcomes | Green |
| `"primary"` | Informational, action available | Theme accent |
| `"warning"` | Caution, degraded state | Yellow/orange |
| `"danger"` | Errors, failures (default) | Red |
| `"neutral"` | Neutral info | Gray |

The default variant is `"danger"` if omitted.

### Examples

```lua
-- Success confirmation
mcp:notify("You're up to date (v0.18.0)", "success")

-- Action available
mcp:notify("Update available: v0.19.0 — use the star menu to update", "primary")

-- Warning with suggestion
mcp:notify("Couldn't fetch that page — try the bookmarklet instead", "warning")

-- Error (default variant)
mcp:notify("Failed to save data")
```

## How It Works

The MCP shell maintains a `_notifications` array of `MCP.Notification` objects. Each notification renders as a Shoelace `<sl-alert>` with a close button. When the user clicks the close button, `dismiss()` removes the notification from the array.

### Rendering

Notifications render in the MCP shell viewdef via:

```html
<div class="mcp-notifications" ui-view="notifications()?wrapper=lua.ViewList"></div>
```

Each notification uses `MCP.Notification.list-item.html`:

```html
<template>
  <div class="mcp-notification">
    <sl-alert ui-attr-variant="variant" open closable ui-event-sl-after-hide="dismiss()">
      <span ui-value="message"></span>
    </sl-alert>
  </div>
</template>
```

## Key Points

- **Available globally** — `mcp:notify()` works from any app since `mcp` is the global MCP shell object
- **Auto-dismissable** — users close notifications via the X button; no timer needed
- **No setup required** — the notification list and viewdef are part of the MCP shell, not your app
- **Plain text only** — the message is rendered via `ui-value` (text content, not HTML)
