# ThemeManager

**Source Spec:** specs/pluggable-themes.md
**Requirements:** R39, R40, R41, R42, R43, R44, R45, R46, R47, R48, R49, R50, R51, R52

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
- **InjectThemeBlock(baseDir)**: Updates index.html with frictionless block
- **GenerateThemeBlock(themes, defaultTheme)**: Generates HTML with script + link elements

## Collaborators

- **os/filepath**: File path operations
- **regexp**: CSS comment parsing
- **strings**: HTML manipulation

## Sequences

- seq-theme-inject.md
- seq-theme-list.md
