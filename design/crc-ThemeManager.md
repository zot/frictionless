# ThemeManager

**Source Spec:** specs/pluggable-themes.md
**Requirements:** R40, R41, R42, R43, R44, R45, R46, R47, R48, R49, R50, R51, R52, R53, R136, R137, R138, R139, R140, R141, R142, R143

Manages theme CSS files and index.html injection.

## Knows

- themesDir: Path to `.ui/html/themes/` directory
- baseCSS: Name of base.css file (excluded from theme list)
- defaultTheme: Default theme name ("lcars")
- frictionlessMarkerStart: `<!-- #frictionless -->`
- frictionlessMarkerEnd: `<!-- /frictionless -->`

## Does

- **ListThemes(baseDir)**: Scans themes directory for .css files, returns theme names (excludes base.css)
- **GetCurrentTheme(baseDir)**: Reads default theme from config or returns "lcars"
- **GetThemeClasses(baseDir, theme)**: Parses CSS file for `@class` annotations
- **ParseThemeCSS(cssContent)**: Extracts all metadata from CSS comment block:
  - `@theme`, `@description` for theme-level metadata
  - `@class` blocks with `@description`, `@usage`, `@elements` attributes
- **InjectThemeBlock(baseDir)**: Updates index.html with frictionless block (skips if block already present)
- **GenerateThemeBlock(baseDir, themes, defaultTheme)**: Generates HTML with script + cache-busted link elements + favicon placeholder
- **ListThemesWithInfo(baseDir)**: Returns themes with descriptions, accent colors, current theme
- **GetThemeAccentColor(cssContent)**: Extracts `--term-accent` value from CSS
- **GetAllThemeClasses(baseDir)**: Scans all theme CSS files, returns deduplicated union of all `@class` entries
- **AuditAppTheme(baseDir, appName, theme)**: Compares app CSS classes against documented theme classes; empty theme uses all-themes list
- **WatchIndexHTML(baseDir, log)**: Watches index.html for writes; re-injects theme block if missing

## Collaborators

- **os/filepath**: File path operations
- **regexp**: CSS comment parsing
- **strings**: HTML manipulation
- **fsnotify**: File system watcher for index.html changes

## Sequences

- seq-theme-inject.md
- seq-theme-list.md
- seq-theme-audit.md
