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

## Notes

- Block injected at start of `<head>` (after opening tag)
- Script runs before CSS loads to prevent flash
- base.css always first, then themes alphabetically
- Existing block removed before injection (idempotent)
