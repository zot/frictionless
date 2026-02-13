# Prefs - Design

## Intent

User preferences panel for frictionless settings. Initial version focuses on theme management with visual swatches and instant switching.

## Layout

```
+--------------------------------------------------+
|  Preferences                          panel-header|
+--------------------------------------------------+
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
| _browserTheme | string | Set by JS on load via hidden input bridge |

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
| loadThemes() | Fetch theme list from server, populate _themes |
| currentTheme() | Returns _currentTheme |
| setCurrentTheme(name) | Update _currentTheme, apply theme |
| applyTheme(name) | Inject JS via mcp.code targeting `.prefs-inner` element |
| syncFromBrowser() | Copy _browserTheme to _currentTheme on load |
| mutate() | Add missing themes (e.g., ninja) during hot-reload |

### ThemeItem

| Method | Description |
|--------|-------------|
| isSelected() | Returns self.name == prefs._currentTheme |
| select() | Call prefs:setCurrentTheme(self.name) |
| swatchStyle() | Returns inline CSS for swatch background color |

## ViewDefs

| File | Type | Purpose |
|------|------|---------|
| Prefs.DEFAULT.html | Prefs | Main panel with header and theme list |
| Prefs.ThemeItem.list-item.html | ThemeItem | Theme card with radio, name, description, swatch |

## Events

None. Theme switching is handled entirely in Lua and client-side JavaScript.

## Theme Loading

On app load:
1. JS reads current theme from localStorage
2. JS sets hidden input `.browser-theme-bridge` value, triggering `_browserTheme` binding
3. `syncFromBrowser()` copies `_browserTheme` to `_currentTheme`
4. JS injects CSS `<link>` tags for all known themes
5. Themes are defined statically in Lua (clarity, lcars, midnight, ninja)

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
