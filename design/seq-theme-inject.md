# Sequence: Theme Block Injection

**Requirements:** R42, R43, R44, R45

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
<link rel="stylesheet" href="/themes/base.css">
<link rel="stylesheet" href="/themes/lcars.css">
<link rel="stylesheet" href="/themes/minimal.css">
<!-- /frictionless -->
```

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
