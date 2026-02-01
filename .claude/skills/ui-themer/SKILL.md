---
name: ui-themer
description: Analyze and standardize theme class usage in Frictionless UI apps
---

# UI Themer

Workflow for auditing and standardizing semantic CSS class usage across UI apps.

## Commands

| Command | Description |
|---------|-------------|
| `mcp theme list` | List available themes and current theme |
| `mcp theme classes [THEME]` | Show semantic classes with descriptions |
| `mcp theme audit APP [THEME]` | Audit an app's CSS class usage |

## Workflow: Standardize an App

1. **Audit the app**
   ```bash
   .ui/mcp theme audit my-app
   ```

   This returns:
   - `undocumented_classes`: Classes used but not in theme (may need documenting)
   - `unused_theme_classes`: Theme classes not used by this app (OK - not all apps use all classes)
   - `summary`: Counts of documented vs undocumented

2. **Review undocumented classes**

   For each undocumented class, decide:
   - **Document it**: Add to the theme CSS metadata block and implement the styles
   - **Replace it**: Use an existing theme class if one fits
   - **Keep it**: App-specific classes are fine (e.g., `chat-message` in a chat app)

3. **Add new theme classes** (if needed)

   Edit `{base_dir}/html/themes/lcars.css` — add to the metadata block:
   ```css
   /*
   @theme lcars
   @description LCARS-inspired design

   @class new-class-name
     @description What it does visually
     @usage When/where to use it
     @elements div, header
   */
   ```

   Then add the CSS rules:
   ```css
   .theme-lcars .new-class-name {
     /* styling */
   }
   ```

## Theme CSS Format

Theme files are CSS with `@` metadata annotations in a comment block:

```css
/*
@theme lcars
@description Subtle Star Trek LCARS-inspired design

@class panel-header
  @description Header bar with bottom accent
  @usage Panel/section headers with title and action buttons
  @elements div, header

@class section-header
  @description Collapsible section header with hover effects
  @usage Collapsible sections within panels
  @elements div
*/

.theme-lcars {
  --term-bg: #0a0a0f;
  --term-accent: #E07A47;
  /* ... */
}

.theme-lcars .panel-header {
  border-bottom: 3px solid var(--term-accent);
}
```

Metadata fields:
- `@theme`: Theme identifier (matches filename without .css)
- `@description`: One-line theme description
- `@class`: Start a semantic class definition
  - `@description`: What the class does visually
  - `@usage`: When/where to use it
  - `@elements`: HTML elements it's typically applied to

## Class Naming Conventions

- **Structural**: `panel-header`, `section-header`, `input-area`
- **State modifiers**: `selected`, `expanded`, `active`
- **Compound**: Use with base class, e.g., `.item.selected`

Avoid:
- Overly specific names: `chat-message-header` (use `panel-header`)
- Presentation-focused names: `orange-border` (describe purpose, not appearance)

## Current Theme

The active theme is stored in localStorage and applied via an HTML class:

```html
<html class="theme-lcars">
```

Users switch themes in the **Prefs** app. The selection persists across sessions.

All apps inherit styles from the active theme automatically.

## Example Audit Output

```json
{
  "app": "app-console",
  "theme": "lcars",
  "undocumented_classes": [
    {"class": "chat-message", "file": "AppConsole.Chat.html", "line": 15},
    {"class": "app-item", "file": "AppConsole.AppInfo.list-item.html", "line": 3}
  ],
  "unused_theme_classes": ["input-area"],
  "summary": {
    "total": 12,
    "documented": 10,
    "undocumented": 2
  }
}
```

This app uses 12 distinct CSS classes, 10 match documented theme patterns, and 2 are app-specific.
