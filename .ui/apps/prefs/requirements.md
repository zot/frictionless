# Prefs App - Requirements

Preferences app for managing frictionless user settings, starting with theme management. Provides visual theme selection with instant preview.

## Features

### Theme Management

1. **List installed themes** with visual swatches showing each theme's `--term-accent` color
2. **Switch themes** with instant preview - updates `<html>` class and localStorage
3. **Show current theme** indicator (checkmark or highlight)
4. **Persist selection** via `mcp:readSettings()`/`mcp:writeSettings()` to settings.json; localStorage is kept as a cache for instant theme application on page load

### Bundled Themes

Four themes ship by default (defined statically in Lua):

| Theme | Description |
|-------|-------------|
| clarity | Clean, editorial light theme with slate blue accent |
| lcars | Subtle Star Trek LCARS-inspired design |
| midnight | Modern dark theme with teal accent |
| ninja | Playful teal theme with cute cartoon ninjas |

Default theme is "lcars" (fallback when no setting exists).

### Update Check Preferences

1. **Toggle update checks** - sl-checkbox to enable/disable automatic update checks on startup
2. **Check Now button** - triggers an immediate update check and shows the result as a notification via `mcp:notify()`:
   - If no update: shows "You're up to date (vX.Y.Z)" success notification
   - If update available: shows "Update available: vX.Y.Z — use the star menu to update" primary notification
3. **Delegates to mcp** - uses `mcp:getUpdatePreference()`, `mcp:setUpdatePreference()`, and `mcp:checkForUpdates()` for persistence and check logic

### Tutorial

1. **Run Tutorial button** — triggers `mcp:startTutorial()` to re-run the first-run spotlight walkthrough
2. The button is always visible in a "Tutorial" section of the prefs panel

### Future Enhancements (not in initial release)

- Import custom themes from file/URL
- Delete user-added themes (protect bundled themes)
- Theme preview without committing
- Live CSS variable editor
