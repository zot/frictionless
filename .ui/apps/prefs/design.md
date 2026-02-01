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
| applyTheme(name) | Update <html> class and localStorage |

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
1. Read current theme from localStorage (via JavaScript bridge)
2. Fetch theme metadata from server via `/api/ui_theme?action=list`
3. Create ThemeItem objects with name, description, accent color
4. Parse accent color from theme CSS (or use API if available)
