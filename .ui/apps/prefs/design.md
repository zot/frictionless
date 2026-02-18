# Prefs - Design

## Intent

User preferences panel for frictionless settings. Initial version focuses on theme management with visual swatches and instant switching.

## Layout

```
+--------------------------------------------------+
|  Preferences                          panel-header|
+--------------------------------------------------+
|                                                   |
|  Updates                           section-header |
|  [x] Check for updates on startup  [Check Now]   |
|                                                   |
|  Themes                            section-header |
|  +---------------------------------------------+ |
|  | [*] LCARS                          [=====]  | | <- selected theme
|  |     Subtle Star Trek LCARS-inspired design  | |
|  +---------------------------------------------+ |
|  | [ ] Minimal                        [=====]  | | <- theme option
|  |     Clean, low-contrast theme               | |
|  +---------------------------------------------+ |
|                                                   |
+--------------------------------------------------+
```

The swatch `[=====]` shows the theme's `--term-accent` color as a small rounded rectangle.

## Data Model

### Prefs

| Field | Type | Description |
|-------|------|-------------|
| _themes | ThemeItem[] | List of available themes |
| _currentTheme | string | Currently selected theme name |

### ThemeItem

| Field | Type | Description |
|-------|------|-------------|
| name | string | Theme identifier (e.g., "lcars") |
| description | string | Theme description from CSS metadata |
| accentColor | string | Value of --term-accent for swatch |
| _prefs | ref | Reference to parent Prefs for callbacks |

## Methods

### Prefs

| Method | Description |
|--------|-------------|
| themes() | Returns _themes for binding |
| currentTheme() | Returns _currentTheme |
| setCurrentTheme(name) | Update _currentTheme, write to settings.json, apply theme |
| applyTheme(name) | Inject JS via mcp.code targeting `.prefs-inner` element |
| loadThemeFromSettings() | Read theme from .ui/storage/settings.json, apply it |
| checkUpdates() | Returns current update-check preference via `mcp:getUpdatePreference()` |
| toggleCheckUpdates() | Toggles update-check preference via `mcp:setUpdatePreference()` |
| checkNow() | Runs `mcp:checkForUpdates()`, then shows notification: success "Up to date" or info with version and "Update Now" button via `mcp:startUpdate()` |

### ThemeItem

| Method | Description |
|--------|-------------|
| isSelected() | Returns self.name == prefs._currentTheme |
| select() | Call prefs:setCurrentTheme(self.name) |
| swatchStyle() | Returns inline CSS for swatch background color |

## ViewDefs

| File | Type | Purpose |
|------|------|---------|
| Prefs.DEFAULT.html | Prefs | Main panel with header, theme list, and updates section |
| Prefs.ThemeItem.list-item.html | ThemeItem | Theme card with radio, name, description, swatch |

## Events

None. Theme switching is handled entirely in Lua and client-side JavaScript.

## Theme Loading

Theme persistence uses `.ui/storage/settings.json` as the source of truth. localStorage is kept as a cache for instant theme application on page load (index.html reads it before Lua loads). Settings I/O is handled by the global `mcp:readSettings()` and `mcp:writeSettings()` methods.

On app load:
1. Lua reads settings.json via `loadThemeFromSettings()` (calls `mcp:readSettings()`)
2. Sets `_currentTheme` from the file (falls back to "lcars" if missing)
3. Applies theme via JS (sets document class + mirrors to localStorage)
4. Themes are defined statically in Lua (clarity, lcars, midnight, ninja)

On theme change:
1. `setCurrentTheme(name)` writes to settings.json via `mcp:writeSettings()`
2. JS applies the theme class and mirrors to localStorage

## Styling Notes

### Inner Wrapper Pattern

The viewdef uses a `.prefs-inner` wrapper div inside the root `.prefs-panel`:

```html
<div class="prefs-panel">
  <div class="prefs-inner">
    <!-- header, content, bridge elements -->
  </div>
</div>
```

This is required because ui-engine merges the viewdef root element into the container, making its children direct children of `.mcp-app-container`. The MCP shell applies `> div { height: 100% !important }` to direct child divs. Without the inner wrapper, every child div (header, content) would get `height: 100%`, breaking the flex layout.

The `.prefs-inner` div receives the flex column layout (`display: flex; flex-direction: column; height: 100%`), while `.prefs-panel` only sets font, background, and color.

JS functions (`applyTheme`, `injectThemeCSS`) are stored on the `.prefs-inner` element, so `applyTheme()` in Lua targets `document.querySelector('.prefs-inner')`.
