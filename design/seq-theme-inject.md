# Sequence: Theme Block Injection

**Requirements:** R42, R43, R44, R45, R142, R143

Server startup injects theme support into index.html.

```
┌─────────┐       ┌──────────────┐       ┌────────────┐
│ Server  │       │ThemeManager  │       │ FileSystem │
└────┬────┘       └──────┬───────┘       └─────┬──────┘
     │                   │                     │
     │ Start()           │                     │
     ├──────────────────>│                     │
     │                   │                     │
     │                   │ Read index.html     │
     │                   ├────────────────────>│
     │                   │<────────────────────┤
     │                   │                     │
     │                   │ Scan themes/*.css   │
     │                   ├────────────────────>│
     │                   │<────────────────────┤
     │                   │ [lcars.css,         │
     │                   │  minimal.css]       │
     │                   │                     │
     │                   │ GetDefaultTheme()   │
     │                   ├─┐                   │
     │                   │ │ "lcars"           │
     │                   │<┘                   │
     │                   │                     │
     │                   │ Remove existing     │
     │                   │ #frictionless block │
     │                   ├─┐                   │
     │                   │ │                   │
     │                   │<┘                   │
     │                   │                     │
     │                   │ Generate block:     │
     │                   │ <script>restore     │
     │                   │ <link>base.css      │
     │                   │ <link>lcars.css     │
     │                   │ <link>minimal.css   │
     │                   ├─┐                   │
     │                   │ │                   │
     │                   │<┘                   │
     │                   │                     │
     │                   │ Inject after <head> │
     │                   ├─┐                   │
     │                   │ │                   │
     │                   │<┘                   │
     │                   │                     │
     │                   │ Write index.html    │
     │                   ├────────────────────>│
     │                   │<────────────────────┤
     │<──────────────────┤                     │
     │                   │                     │
```

## Generated Block

```html
<!-- #frictionless -->
<script>
  document.documentElement.className = 'theme-' + (localStorage.getItem('theme') || 'lcars');
</script>
<link rel="stylesheet" href="/themes/base.css?v=1737500000">
<link rel="stylesheet" href="/themes/lcars.css?v=1737500000">
<link rel="stylesheet" href="/themes/minimal.css?v=1737500000">
<link rel="icon" id="app-favicon" href="data:,">
<!-- /frictionless -->
```

Cache-busting: each `<link>` gets `?v={modtime}` from the CSS file's modification
timestamp (Unix seconds). Browser reloads CSS when the file changes on disk.

## Scenario 2: File Watcher Re-injection

After startup, the server watches `index.html` for writes. When an external process
(e.g., `make cache`) overwrites the file, the watcher re-injects the theme block.

```
┌──────────┐    ┌──────────────┐    ┌────────────┐
│ fsnotify │    │ThemeManager  │    │ FileSystem │
└────┬─────┘    └──────┬───────┘    └─────┬──────┘
     │                 │                  │
     │ Write event     │                  │
     ├────────────────>│                  │
     │                 │                  │
     │                 │ Read index.html  │
     │                 ├─────────────────>│
     │                 │<─────────────────┤
     │                 │                  │
     │                 │ Has #frictionless│
     │                 │ block?           │
     │                 ├─┐                │
     │                 │ │ No             │
     │                 │<┘                │
     │                 │                  │
     │                 │ InjectThemeBlock │
     │                 ├─────────────────>│
     │                 │<─────────────────┤
     │                 │                  │
```

If block is already present, the watcher does nothing.

## Notes

- Block injected at start of `<head>` (after opening tag)
- Script runs before CSS loads to prevent flash
- base.css always first, then themes alphabetically
- Existing block removed before injection (idempotent)
- File watcher ensures block survives external overwrites (e.g., ui-engine updates)
- Cache busting via `?v={modtime}` on each CSS link (cssModTime helper)
- Favicon placeholder (`<link id="app-favicon">`) set dynamically by each app's viewdef script

## CSS Cache Busting Strategy

Two layers of CSS need cache busting:

1. **Theme CSS** (index.html): Server-side mod-time stamps via `cssModTime()` in Go.
   Re-injected on file change by the watcher.

2. **App CSS** (viewdefs): Client-side `Date.now()` nonce via `<script>` that
   dynamically creates `<link>` elements. Runs on each viewdef load, so every page
   load/hot-reload gets fresh CSS.

The MCP shell uses `.mcp-shell .hidden` (specificity 0,2,0) to ensure the `.hidden` class
always overrides app viewdef display rules without `!important`. The `.hidden.showing`
pattern (also 0,2,0) wins by source order for panel toggling.
