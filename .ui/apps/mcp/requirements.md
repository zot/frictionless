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

### Wait Time Counter (Client-Local)

Display a counter inside the spinner showing seconds elapsed since the spinner appeared:
- Entirely client-side JavaScript — no server round-trips needed
- When spinner becomes visible (`pollingEvents()` returns false), store timestamp and start interval
- Interval updates counter display every second
- Counter shows seconds elapsed as bold orange text with black glow for contrast
- Hidden when elapsed time is 5 seconds or less
- When spinner hides (`pollingEvents()` becomes true), clear interval and hide counter

### Busy Notification

When pushing an event via `mcp.pushState`:
- Idempotently override the global `pushState` function to add this behavior
- Before pushing, check `mcp:waitTime()`
- If waitTime exceeds 5 seconds, show a notification: "Claude might be busy. Use /ui events to reconnect."
- Use the "warning" variant for this notification
- Only show once per disconnect period (reset when Claude reconnects)

Additionally, on UI refresh, check if events are pending and Claude appears disconnected:
- If `waitTime() > 5` and there are pending events, show the same warning
- This catches interactions that don't go through pushState

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
- Compact padding (6px horizontal)

### Help Icon

A help icon (question mark) at the right edge of the status bar, before the build toggles:
- Clicking opens `localhost:{MCP-PORT}/api/resource/` in a new browser tab
- Shows resources documentation index
- Tooltip: "Documentation"

### Build Settings Toggles

Two toggle icons on the right edge of the status bar, grouped tightly together:

| Toggle | Icons | Values |
|--------|-------|--------|
| Build mode | 🚀 rocket / 💎 diamond | fast / thorough |
| Execution | ⏳ hourglass / 🔄 arrows | foreground / background |

- Icons have minimal padding (3px horizontal) for a compact appearance
- Click toggles between states
- Hover shows tooltip describing current value with "(click to change)"

## Notifications

Agents can display notifications to alert users of important events (errors, warnings, info):

- `mcp:notify(message, variant)` - Show a notification toast
- `variant` can be: "danger" (red), "warning" (yellow), "success" (green), "primary" (blue), "neutral" (gray)
- Default variant is "danger" (most notifications are errors)
- Notifications appear as Shoelace alerts, auto-dismiss after 5 seconds
- Multiple notifications stack vertically
- Each notification has a close button for manual dismissal
- Notifications appear in top-right corner, below the menu button

## Styling

- No padding or margins around the app content
- The menu button should have subtle styling (semi-transparent, hover effect)
- Menu button should not interfere with app content interaction
- Status bar has subtle styling (light background, border-top)
