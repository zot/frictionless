# MCP

The outer shell for all frictionless apps. Displays the current app and provides navigation between apps.

## Architecture

**Uses the global `mcp` object directly.** This app does not create its own prototype — it renders the `mcp` object provided by the server. The `mcp` object has:
- `value` - the currently displayed app (set via `mcp:display(appName)`)
- `code` - JavaScript code to execute (for browser control)

## Layout

The shell is minimal chrome around the current app:
- Full viewport display of `mcp.value` (the current app)
- A 9-dot menu button in the top-right corner for app switching
- A status bar at the bottom showing `mcp.statusLine` in `mcp.statusColor`
- A hidden element with `ui-code="code"` for JavaScript execution

## App Switcher Menu

The 9-dot button (grid icon) in the top-right corner:
- Overlays the app content (always visible on top)
- Has a glow for easy visibility
- Opens a dropdown/popover menu with app icons
- Clicking an app calls `mcp:display(appName)` to switch
- Menu closes after selection

### Icon Grid Layout

Apps display as icons with names underneath:
- Each app's icon comes from its `icon.html` file (contains emoji, `<sl-icon>`, or `<img>`)
- App name displayed below each icon
- Icons arranged in rows of 3, Z formation (left-to-right, then next row)
- Clickable icon cards with hover effect

## Processing Indicator

Show a spinner overlaying the 9-dot button when the agent event loop is not connected to the `/wait` endpoint:
- Indicates that Claude is processing events
- Spinner overlays the 9-dot icon at 50% opacity (both visible)
- The 9-dot button remains clickable underneath
- When connected to `/wait`, the spinner is hidden

## Available Apps

The menu should list apps discovered from the apps directory. Since mcp is Lua-driven, it should scan for available apps on load using the same filesystem pattern as the `apps` app.

## JavaScript Execution

A hidden element binds to `mcp.code` via `ui-code`:
- When `mcp.code` changes, the JavaScript is executed
- Used for browser control (close window, open URLs, etc.)
- The element is invisible (display: none or similar)

## Events

This app does not send events to Claude. App switching is handled entirely in Lua via `mcp:display()`.

## Status Bar

A status bar at the bottom of the viewport:
- Always visible
- Displays `mcp.statusLine` text with `mcp.statusClass` CSS class
- The `.thinking` class styles text as orange bold-italic
- Maintains consistent height even when empty

## Styling

- No padding or margins around the app content
- The menu button should have subtle styling (semi-transparent, hover effect)
- Menu button should not interfere with app content interaction
- Status bar has subtle styling (light background, border-top)
