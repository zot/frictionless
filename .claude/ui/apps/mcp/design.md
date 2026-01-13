# MCP - Design

## Intent

Outer shell for all ui-mcp apps. Displays the current app full-viewport and provides a floating app switcher menu in the top-right corner.

## Layout

```
+--------------------------------------------------+
|                                        [:::] <- 9-dot menu
|                                                  |
|              Current App (mcp.value)             |
|                  (full viewport)                 |
|                                                  |
|                                                  |
+--------------------------------------------------+
[hidden: ui-code element]
```

### Menu Open State

```
+--------------------------------------------------+
|                                        [:::] <- 9-dot menu
|                                     +----------+ |
|              Current App            | contacts | |
|                                     | tasks    | |
|                                     | apps     | |
|                                     +----------+ |
+--------------------------------------------------+
```

## Data Model

### MCP (extends global mcp object)

The global `mcp` object is provided by the server. This app adds:

| Field | Type | Description |
|-------|------|-------------|
| value | object | Currently displayed app (set by mcp:display) |
| code | string | JavaScript to execute via ui-code |
| _availableApps | string[] | List of discovered app names |
| menuOpen | boolean | Whether app menu is visible |

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

### AppMenuItem (wrapper for app name string)

| Method | Description |
|--------|-------------|
| name() | Returns the app name |
| select() | Calls mcp:selectApp(self.name) |

## ViewDefs

| File | Type | Purpose |
|------|------|---------|
| MCP.DEFAULT.html | MCP | Shell with app view, menu button, menu dropdown |
| AppMenuItem.list-item.html | AppMenuItem | Menu item row |

## Events

None. App switching is handled entirely in Lua via `mcp:display()`.

## App Discovery (Lua)

On load, scan `{base_dir}/apps/` for directories containing `app.lua` (built apps only). Store names in `mcp._availableApps`.
