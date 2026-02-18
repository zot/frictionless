# Prefs App - Requirements

Preferences app for managing frictionless user settings, starting with theme management. Provides visual theme selection with instant preview.

## Features

### Theme Management

1. **List installed themes** with visual swatches showing each theme's `--term-accent` color
2. **Switch themes** with instant preview - updates `<html>` class and localStorage
3. **Show current theme** indicator (checkmark or highlight)
4. **Persist selection** via `mcp:readSettings()`/`mcp:writeSettings()` to settings.json; localStorage is kept as a cache for instant theme application on page load

### Update Check Preferences

1. **Toggle update checks** - sl-checkbox to enable/disable automatic update checks on startup
2. **Check Now button** - triggers an immediate update check and shows the result as a notification:
   - If no update: shows "Up to date" success notification (auto-dismisses)
   - If update available: shows version info with an "Update Now" button that triggers the update flow, plus a dismiss option
3. **Delegates to mcp** - uses `mcp:getUpdatePreference()`, `mcp:setUpdatePreference()`, and `mcp:checkForUpdates()` for persistence and check logic

### Tutorial

1. **Run Tutorial button** — triggers `mcp:startTutorial()` to re-run the first-run spotlight walkthrough
2. The button is always visible in a "Tutorial" section of the prefs panel

### Future Enhancements (not in initial release)

- Import custom themes from file/URL
- Delete user-added themes (protect bundled themes)
- Theme preview without committing
- Live CSS variable editor
