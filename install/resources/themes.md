# Themes

Frictionless uses CSS-based themes for consistent visual styling. Switch themes instantly — your choice persists across sessions.

## Overview

Themes control:
- **Colors** — backgrounds, text, accents, status indicators
- **Typography** — font families for code and UI text
- **Semantic classes** — consistent styling patterns across apps

Theme switching happens via an HTML class on the root element (e.g., `theme-lcars`). No page reload required.

## Available Themes

| Theme | Description | Accent Color |
|-------|-------------|--------------|
| **lcars** (default) | Subtle Star Trek LCARS-inspired design | Orange `#E07A47` |
| **clarity** | Clean, editorial light theme | Slate blue `#3b6ea5` |
| **midnight** | Modern dark theme | Teal `#2dd4bf` |
| **ninja** | Playful teal with cartoon ninja silhouettes | Dark `#2d2d2d` |

## Switching Themes

1. Open the **Prefs** app from the App Console
2. Select your preferred theme
3. Changes apply immediately and persist via localStorage

## For Developers

### CSS Variables

All themes define these variables on their root class (e.g., `.theme-lcars`):

| Variable | Purpose |
|----------|---------|
| `--term-bg` | Primary background |
| `--term-bg-elevated` | Elevated surface (cards, modals) |
| `--term-bg-hover` | Hover state background |
| `--term-bg-panel` | Panel background |
| `--term-border` | Border color |
| `--term-text` | Primary text |
| `--term-text-dim` | Secondary text |
| `--term-text-muted` | Muted/disabled text |
| `--term-accent` | Primary accent color |
| `--term-accent-glow` | Accent with transparency (for glows) |
| `--term-accent-bright` | Brighter accent variant |
| `--term-accent-dim` | Dimmer accent variant |
| `--term-success` | Success state |
| `--term-warning` | Warning state |
| `--term-danger` | Danger/error state |
| `--term-error` | Error state (alias for danger) |
| `--term-info` | Informational state |
| `--term-mono` | Monospace font family |
| `--term-sans` | Sans-serif font family |

### Semantic Classes

Use these classes for consistent cross-theme styling:

| Class | Description | Usage |
|-------|-------------|-------|
| `panel-header` | Header bar with bottom accent | Panel/section headers with title and action buttons |
| `panel-header-left` | Header with left accent bar | Detail panels, secondary headers |
| `section-header` | Collapsible section header with hover effects | Collapsible sections within panels |
| `selected-item` | Selected item with gradient/highlight | List items in selected state |
| `input-area` | Input area with top accent | Chat input, form areas |

### Creating Custom Themes

1. Create a CSS file in `.ui/html/themes/<name>.css`

2. Add metadata annotations at the top of the file:

```css
/*
@theme my-theme
@description A brief description of your theme

@class panel-header
  @description Header bar with accent
  @usage Panel headers
  @elements div, header

@class section-header
  @description Collapsible section header
  @usage Sections within panels
  @elements div
*/
```

3. Define your theme's CSS variables:

```css
.theme-my-theme {
  --term-bg: #1a1a2e;
  --term-accent: #e94560;
  /* ... other variables ... */
}
```

4. Style the semantic classes:

```css
.theme-my-theme .panel-header {
  border-bottom: 3px solid var(--term-accent);
}
```

5. Restart the server — your theme appears automatically

### Metadata Format

Theme files use `@` annotations in CSS comments:

- `@theme <name>` — Theme identifier (required)
- `@description <text>` — One-line description
- `@class <name>` — Start a class definition block
  - `@description <text>` — What the class does visually
  - `@usage <text>` — When/where to use it
  - `@elements <list>` — HTML elements it's typically applied to

## CLI Tools

Audit and inspect themes from the command line:

```bash
# List available themes
frictionless theme list

# Show semantic classes for a theme
frictionless theme classes [THEME]

# Audit an app's CSS class usage
frictionless theme audit APP [THEME]
```

The audit command reports:
- **undocumented_classes** — Classes used but not defined in theme
- **unused_theme_classes** — Theme classes not used by this app
- **summary** — Counts of documented vs undocumented usage
