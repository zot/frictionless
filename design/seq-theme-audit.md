# Sequence: Theme Audit Command

**Requirements:** R48, R49, R138, R139

Audit app viewdef CSS class usage against theme-documented classes.

```
┌─────┐       ┌──────────────┐       ┌────────────┐
│ CLI │       │ThemeManager  │       │ FileSystem │
└──┬──┘       └──────┬───────┘       └─────┬──────┘
   │                 │                     │
   │ theme audit APP │                     │
   │ [THEME]         │                     │
   ├────────────────>│                     │
   │                 │                     │
   │                 │ THEME provided?     │
   │                 ├─┐                   │
   │                 │ │ yes: GetTheme     │
   │                 │ │   Classes(theme)  │
   │                 │ │ no: GetAllTheme   │
   │                 │ │   Classes()       │
   │                 │<┘                   │
   │                 │                     │
   │                 │ Scan theme CSS      │
   │                 ├────────────────────>│
   │                 │<────────────────────┤
   │                 │                     │
   │                 │ Parse @class blocks │
   │                 ├─┐                   │
   │                 │ │ documented =      │
   │                 │ │  [panel-header,   │
   │                 │ │   sidebar-panel,  │
   │                 │ │   content-card,   │
   │                 │ │   ...]            │
   │                 │<┘                   │
   │                 │                     │
   │                 │ Scan app viewdefs   │
   │                 ├────────────────────>│
   │                 │<────────────────────┤
   │                 │ HTML files          │
   │                 │                     │
   │                 │ Extract CSS classes │
   │                 ├─┐                   │
   │                 │ │ used =            │
   │                 │ │  [panel-header,   │
   │                 │ │   unknown-class,  │
   │                 │ │   ...]            │
   │                 │<┘                   │
   │                 │                     │
   │                 │ Compare sets        │
   │                 ├─┐                   │
   │                 │ │ undocumented =    │
   │                 │ │   used - docs     │
   │                 │ │ unused =          │
   │                 │ │   docs - used     │
   │                 │<┘                   │
   │                 │                     │
   │<────────────────┤                     │
   │ {documented,    │                     │
   │  undocumented,  │                     │
   │  unused}        │                     │
   │                 │                     │
```

## Notes

- No theme argument: scans all theme CSS files, deduplicates classes
- Single theme argument: scans only that theme's CSS file
- Viewdefs scanned from `apps/{app}/viewdefs/*.html`
- CSS files scanned from `apps/{app}/css/*.css`
