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

Fixed at the bottom of the viewport, always visible. Displays `mcp.statusLine` text with `mcp.statusClass` CSS class applied. The `.thinking` class styles text as orange bold-italic.

### Processing Indicator

When the agent event loop is not connected to `/wait` (Claude is processing):
- A semi-transparent spinner overlays the 9-dot icon at 50% opacity
- Both the spinner and 9-dot button remain visible
- The 9-dot button stays clickable (spinner has pointer-events: none)
- Uses `mcp:pollingEvents()` server method - returns false when processing
- Spinner hidden via `ui-class-hidden="pollingEvents()"`

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
| MCP.DEFAULT.html | MCP | Shell with app view, menu button, icon grid dropdown, status bar |
| MCP.AppMenuItem.list-item.html | MCP.AppMenuItem | Icon card with icon HTML and name below |

## Events

None. App switching is handled entirely in Lua via `mcp:display()`.

## App Discovery (Lua)

On load, scan `{base_dir}/apps/` for directories containing `app.lua` (built apps only). Store names in `mcp._availableApps`.
