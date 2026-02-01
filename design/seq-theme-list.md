# Sequence: Theme List Command

**Requirements:** R40, R46

CLI command lists available themes with metadata.

```
┌─────┐       ┌──────────────┐       ┌────────────┐
│ CLI │       │ThemeManager  │       │ FileSystem │
└──┬──┘       └──────┬───────┘       └─────┬──────┘
   │                 │                     │
   │ theme list      │                     │
   ├────────────────>│                     │
   │                 │                     │
   │                 │ Scan themes/*.css   │
   │                 ├────────────────────>│
   │                 │<────────────────────┤
   │                 │ [base.css,          │
   │                 │  lcars.css,         │
   │                 │  minimal.css]       │
   │                 │                     │
   │                 │ Filter: !base.css   │
   │                 ├─┐                   │
   │                 │ │ [lcars.css,       │
   │                 │ │  minimal.css]     │
   │                 │<┘                   │
   │                 │                     │
   │                 │ For each theme:     │
   │                 │                     │
   │                 │ Read lcars.css      │
   │                 ├────────────────────>│
   │                 │<────────────────────┤
   │                 │                     │
   │                 │ ParseThemeMetadata  │
   │                 ├─┐ @theme lcars      │
   │                 │ │ @description ...  │
   │                 │<┘                   │
   │                 │                     │
   │                 │ Read minimal.css    │
   │                 ├────────────────────>│
   │                 │<────────────────────┤
   │                 │                     │
   │                 │ ParseThemeMetadata  │
   │                 ├─┐ @theme minimal    │
   │                 │ │ @description ...  │
   │                 │<┘                   │
   │                 │                     │
   │                 │ GetCurrentTheme()   │
   │                 ├─┐                   │
   │                 │ │ "lcars"           │
   │                 │<┘                   │
   │                 │                     │
   │<────────────────┤                     │
   │ {themes: [...], │                     │
   │  current: ...}  │                     │
   │                 │                     │
```

## Output Format

```json
{
  "themes": [
    {"name": "lcars", "description": "LCARS-inspired design"},
    {"name": "minimal", "description": "Clean minimal theme"}
  ],
  "current": "lcars"
}
```

## Notes

- base.css excluded from theme list (it's shared infrastructure)
- Metadata parsed from CSS comment block at top of file
- Missing @description defaults to empty string
