# Pluggable Themes Specification

**Language:** Go (server), CSS (themes), Lua (UI)
**Environment:** frictionless MCP server, browser-based UI
**Status:** Implemented

## Overview

A pluggable CSS-based theme system with runtime switching. Themes are CSS files with metadata in `@`-prefixed annotations.

## Architecture

- Theme CSS files in `.ui/html/themes/` (served at `/themes/`)
- Metadata extracted from CSS comment blocks using `@` annotations
- Runtime switching via HTML class on `<html>` element
- Theme preference persisted in localStorage

## Theme File Format

Themes are CSS files in `.ui/html/themes/` (served at `/themes/`) with metadata in comments:

### Metadata Comment Block
```css
/*
@theme lcars
@description Subtle Star Trek LCARS-inspired design

@class panel-header
  @description Header bar with bottom accent
  @usage Panel/section headers with title and action buttons
  @elements div, header
*/
```

### CSS Structure
- All rules prefixed with `.theme-{name}` class (e.g., `.theme-lcars .panel-header`)
- CSS variables scoped to theme class (e.g., `.theme-lcars { --term-accent: #E07A47; }`)
- Font imports included in theme file

### Base CSS
A `base.css` file provides:
- `:root` fallback variables for pseudo-elements
- Smooth transition rules between themes
- Global styles independent of theme

## Theme Block Injection

Frictionless injects a `<!-- #frictionless -->` block into `.ui/html/index.html`:

1. Read index.html
2. Remove any existing `<!-- #frictionless -->...<!-- /frictionless -->` block
3. Scan `.ui/html/themes/*.css` for theme files (excluding base.css)
4. Generate block containing:
   - Theme restore script (reads localStorage, sets `<html>` class)
   - `<link>` elements for base.css and each theme
5. Inject block at start of `<head>`
6. Write updated index.html

### When Injection Runs

- **Server startup:** inject once on start (existing behavior)
- **File watcher:** watch `.ui/html/index.html` for writes and re-inject when modified by external processes (e.g., `make cache` updating ui-engine assets)

The watcher must avoid infinite loops: when InjectThemeBlock itself writes the file, the watcher should not re-trigger. Use a guard flag or check whether the block is already present before injecting.

## Theme Switching

### Runtime (Browser)
```javascript
document.documentElement.className = 'theme-' + name;
localStorage.setItem('theme', name);
```

### Page Load
Inline script in `<head>` restores theme from localStorage before CSS loads.

## CLI Commands

- `theme list` - Scan `.css` files, parse metadata from comments
- `theme classes [THEME]` - Parse `@class` annotations from CSS comments
- `theme audit APP [THEME]` - Audit app's CSS class usage against theme

## Installation

### Bundled Themes
- `base.css` - Shared defaults and transitions
- `lcars.css` - LCARS-inspired dark theme (default)
- `clarity.css` - Light editorial theme
- `midnight.css` - Modern dark theme with teal
- `ninja.css` - Playful teal theme

### Install Process
- `frictionless install` copies theme CSS files to `.ui/html/themes/`

## MCP.DEFAULT.html Changes

- Remove embedded theme CSS (variables, classes)
- Keep only shell structural CSS wrapped in `@layer components`
- Themes load via index.html injection

## Prefs App

The Prefs app provides runtime theme management:
- List installed themes with accent color swatches
- Switch themes with instant preview
- Theme preference persists via localStorage
