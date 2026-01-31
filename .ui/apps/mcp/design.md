# MCP - Design

## Intent

Outer shell for all frictionless apps. Displays the current app full-viewport and provides a floating app switcher menu in the top-right corner.

## Layout

```
+--------------------------------------------------+
|                                     [:::] or (o) <- 9-dot menu or spinner
|                                                  |
|              Current App (mcp.value)             |
|                  (full viewport)                 |
|                                                  |
|                                                  |
+--------------------------------------------------+
| Status: mcp.statusLine (with mcp.statusClass)    | <- status bar (always visible)
+--------------------------------------------------+
[hidden: ui-code element]
```

### Status Bar

Fixed at the bottom of the viewport, always visible. Compact horizontal padding (6px). Displays `mcp.statusLine` text with `mcp.statusClass` CSS class applied. The `.thinking` class styles text as orange bold-italic.

At the right end of the status bar are icons grouped tightly together in a `.mcp-status-toggles` container (no gap between icons):

| Icon | Action | Description |
|------|--------|-------------|
| ❓ question mark | openHelp() | Opens `/api/resource/` in new tab |
| 🔧 tools | openTools() | Opens app-console, selects current app |
| 🚀/💎 | toggleBuildMode() | fast / thorough |
| ⏳/🔄 | toggleBackground() | foreground / background |

Icon styling: minimal padding (2px vertical, 3px horizontal), no gap between icons. Click triggers action. Hover shows dynamic tooltip.

### Tools Icon

The tools icon shows the current app in app-console when clicked. If the current app has checkpoints (fast code), the icon glows orange via CSS `filter: drop-shadow()`. The tooltip shows "Go to App" normally or "Go to App - fast coded" when checkpoints exist.

### Processing Indicator

When the agent event loop is not connected to `/wait` (Claude is processing):
- The menu button enters a `.waiting` state via `ui-class-waiting="isWaiting()"`
- A pulsating orange glow effect animates around the button (CSS `button-pulse` animation)
- The grid icon dims to 30% opacity
- A wait time counter appears centered over the button
- The button remains fully clickable during wait state

### Wait Time Counter (Client-Local JavaScript)

Client-side JavaScript manages the counter display without server round-trips:
- A `<script>` block with `setInterval(200ms)` reads timestamp from hidden element
- Server provides `waitStartOffset()` - UNIX timestamp when wait started, or 0 if connected
- Counter calculates elapsed seconds client-side
- Counter shows seconds elapsed, empty when <= 5 seconds
- Bold orange text with glow, centered in button (`.mcp-wait-counter`)
- CSS controls visibility via parent `.waiting` class

### pushState Override

On load, idempotently override the global `pushState` function to:

1. **Inject build settings:**
   - `event.handler` = `"/ui-fast"` or `"/ui-thorough"`
   - `event.background` = `true` or `false`

2. **Warn on long wait times:**
   - If `mcp:waitTime() > 5` and not already notified, show warning notification
   - Resets `_notifiedForDisconnect` when Claude reconnects (waitTime returns to 0)

### Disconnect Check (checkDisconnectNotify)

Called on UI refresh via hidden span binding. Warns if:
- `waitTime() > 5` seconds AND
- `pendingEventCount() > 0` (events are waiting) AND
- Not already notified this disconnect period

This catches the case where user interacts with UI but Claude isn't listening.

### Menu Open State (Icon Grid)

```
+--------------------------------------------------+
|                                        [:::] <- 9-dot menu
|                            +-------------------+ |
|              Current App   | [icon] [icon] [icon]|
|                            |  app1   app2   app3 |
|                            | [icon] [icon] [icon]|
|                            |  app4   app5   app6 |
|                            +-------------------+ |
+--------------------------------------------------+
```

Icons arranged in rows of 3 (Z formation: left-to-right, then next row). Each cell shows the app's icon (from icon.html) with the app name below it.

## Data Model

### MCP (extends global mcp object)

The global `mcp` object is provided by the server. This app adds:

| Field | Type | Description |
|-------|------|-------------|
| value | object | Currently displayed app (set by mcp:display) |
| code | string | JavaScript to execute via ui-code |
| _availableApps | string[] | List of discovered app names |
| menuOpen | boolean | Whether app menu is visible |
| statusLine | string | Status text to display (server-provided) |
| statusClass | string | CSS class for status bar styling (e.g., "thinking") |
| _notifications | Notification[] | Active notification toasts |
| buildMode | string | "fast" or "thorough" - global build mode setting |
| runInBackground | boolean | Whether to run builds in background |
| _notifiedForDisconnect | boolean | Whether disconnect warning has been shown (prevents duplicate notifications) |

## Methods

### MCP (added to global mcp)

| Method | Description |
|--------|-------------|
| availableApps() | Returns _availableApps for binding |
| toggleMenu() | Toggle menuOpen state |
| closeMenu() | Set menuOpen to false |
| menuHidden() | Returns not menuOpen (for ui-class-hidden) |
| selectApp(name) | Call mcp:display(name), close menu |
| scanAvailableApps() | Scan apps/ directory for available apps |
| pollingEvents() | Server-provided: true if agent is connected to /wait endpoint |
| waitTime() | Server-provided: seconds since last agent connection to /wait |
| isWaiting() | Returns true if waitTime() > 0 (for ui-class-waiting binding) |
| pendingEventCount() | Server-provided: number of events waiting to be processed |
| waitStartOffset() | Returns UNIX timestamp when wait started, or 0 if connected (for client-side counter) |
| checkDisconnectNotify() | Check if Claude appears disconnected and show warning notification if needed |
| notify(message, variant) | Show a notification toast (variant: danger, warning, success, primary, neutral) |
| notifications() | Returns _notifications for binding |
| dismissNotification(n) | Remove notification from list |
| openHelp() | Open /api/resource/ in new browser tab using mcp:status().mcp_port |
| openTools() | Display app-console and select the current app |
| currentAppName() | Returns kebab-case name of current app from mcp.value.type |
| currentAppHasCheckpoints() | Returns true if current app has checkpoints (via appConsole:findApp) |
| currentAppNoCheckpoints() | Returns not currentAppHasCheckpoints() |
| toolsTooltip() | Returns "Go to App - fast coded" if checkpoints, else "Go to App" |
| toggleBuildMode() | Toggle between "fast" and "thorough" modes |
| isFastMode() | Returns true if buildMode is "fast" |
| isThoroughMode() | Returns true if buildMode is "thorough" |
| buildModeTooltip() | Returns tooltip text for current mode |
| toggleBackground() | Toggle between foreground and background execution |
| isBackground() | Returns true if runInBackground is true |
| isForeground() | Returns true if runInBackground is false |
| backgroundTooltip() | Returns tooltip text for current execution mode |

### MCP.Notification (notification toast)

| Field | Type | Description |
|-------|------|-------------|
| message | string | Notification text |
| variant | string | Shoelace alert variant (danger, warning, success, primary, neutral) |
| _mcp | ref | Reference to mcp for dismiss callback |

| Method | Description |
|--------|-------------|
| dismiss() | Calls mcp:dismissNotification(self) |

### MCP.AppMenuItem (wrapper for app info)

| Field | Type | Description |
|-------|------|-------------|
| _name | string | App directory name |
| _iconHtml | string | HTML content from icon.html |
| _mcp | ref | Reference to mcp for callbacks |

| Method | Description |
|--------|-------------|
| name() | Returns the app name |
| iconHtml() | Returns the icon HTML content |
| select() | Calls mcp:selectApp(self._name) |

## ViewDefs

| File | Type | Purpose |
|------|------|---------|
| MCP.DEFAULT.html | MCP | Shell with app view, menu button, icon grid dropdown, notifications, status bar |
| MCP.AppMenuItem.list-item.html | MCP.AppMenuItem | Icon card with icon HTML and name below |
| MCP.Notification.list-item.html | MCP.Notification | Toast notification with message and close button |

## Events

None. App switching is handled entirely in Lua via `mcp:display()`.

## App Discovery (Lua)

On load, scan `{base_dir}/apps/` for directories containing `app.lua` (built apps only). Store names in `mcp._availableApps`.
