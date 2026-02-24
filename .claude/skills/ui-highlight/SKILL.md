---
name: ui-highlight
description: draw the user's attention to a UI element by highlighting it with an overlay, ring, and connector line
---

Be sure you have loaded /ui-basics.

# Highlighting UI Elements

The MCP shell has a built-in `window.uiApp.highlight(elementId)` function that draws the user's attention to any element — overlay, box-shadow ring, and a dashed connector line from the user's mouse position. Mouse tracking is always active; no setup needed.

All you need is the element's DOM id. Sources for element IDs:
- **Variables:** each variable has an `elementId` property (see `/ui-variables`)
- **Viewdefs:** elements with `ui-*` bindings get auto-assigned IDs like `ui-42`
- **Playwright:** snapshot refs map to element IDs

## Usage

### Via mcp.code (JS execution)

```lua
mcp.code = [=[window.uiApp.highlight("ui-42")// ]=] .. os.time()
```

**Important:** `mcp.code` uses change detection — re-assigning the same value is a no-op. Append a nonce (e.g. `.. os.time()`) to force re-execution.

**Bash escaping:** When JS is passed through `.ui/mcp run` via bash, `!` gets escaped to `\!`, producing invalid JS. **Avoid `!` in JS code** — use `== null` instead of `!el`, `== 0` instead of `!width`, etc.

### Via rich messages (inline "here" links)

Use `mcp:addRichMessage()` with `mcp:highlightLink()` to embed clickable highlight links directly in chat messages:

```lua
mcp:addRichMessage(
    "I updated the status bar. Click "
    .. mcp:highlightLink("ui-42", "here")
    .. " to see it."
)
```

This renders a chat message with a clickable "here" link that highlights the element when clicked. No `mcp.code` nonce needed — the highlight fires on each click.

`mcp:highlightLink(elementId, label)` returns an HTML anchor tag. You can include multiple links in one message:

```lua
mcp:addRichMessage(
    "The " .. mcp:highlightLink("ui-10", "search box")
    .. " filters the " .. mcp:highlightLink("ui-20", "results list") .. "."
)
```

## What it does

- **Box-shadow ring** on the element (orange, respects border-radius)
- **Translucent overlay** matching the element's bounds
- **SVG dashed connector line** from the user's last mouse position to the element center
- Everything fades after 3 seconds (1s solid + 2s fade)

## Rich messages

All chat messages now render markdown. `mcp:addAgentMessage("**bold** text")` renders with formatting. For raw HTML control (highlight links, custom widgets), use `mcp:addRichMessage(html)`.

| Method | Description |
|--------|-------------|
| `mcp:addAgentMessage(text)` | Add message with markdown rendering |
| `mcp:addRichMessage(html)` | Add message with raw HTML (for highlight links, etc.) |
| `mcp:highlightLink(id, label)` | Returns anchor HTML that highlights element on click |
| `mcp:renderMarkdown(text)` | Convert markdown to HTML fragment (available for custom use) |

## Workflow

1. Get the target element's DOM id (from variables, viewdefs, or Playwright)
2. Either:
   - **Rich message:** `mcp:addRichMessage("Click " .. mcp:highlightLink(id, "here") .. " to see it.")`
   - **JS execution:** `mcp.code = [=[window.uiApp.highlight("id")// ]=] .. os.time()` + `mcp:addAgentMessage("...")`

The connector line draws from the user's last known mouse position to the target, giving a natural "pointing at" effect. Since the user interacts with the chat panel via mouse, the position is typically near the chat area.
