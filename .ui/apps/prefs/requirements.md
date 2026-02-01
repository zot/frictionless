# Prefs App - Requirements

## Purpose

Preferences app for managing frictionless user settings, starting with theme management. Provides visual theme selection with instant preview.

## Features

### Theme Management

1. **List installed themes** with visual swatches showing each theme's `--term-accent` color
2. **Switch themes** with instant preview - updates `<html>` class and localStorage
3. **Show current theme** indicator (checkmark or highlight)
4. **Persist selection** via localStorage so it survives page reloads

### Future Enhancements (not in initial release)

- Import custom themes from file/URL
- Delete user-added themes (protect bundled themes)
- Theme preview without committing
- Live CSS variable editor
