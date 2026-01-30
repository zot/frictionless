---
name: lcars
description: Subtle Star Trek LCARS-inspired design
classes:
  - name: panel-header
    description: Header bar with bottom accent
    usage: Panel/section headers with title and action buttons
    elements: [div, header]
  - name: panel-header-left
    description: Header bar with left accent
    usage: Detail panels where accent is on the left side
    elements: [div, header]
  - name: section-header
    description: Collapsible section header
    usage: Expandable/collapsible sections with hover feedback
    elements: [div]
  - name: item
    description: Base class for list items
    usage: Standard list item styling
    elements: [div, li]
  - name: selected
    description: Selected state modifier (use with .item)
    usage: Apply to items with selection state
    elements: [div, li]
  - name: input-area
    description: Input area with top accent
    usage: Chat/command input areas
    elements: [div]
---

# LCARS Theme

Subtle Star Trek LCARS-inspired design. References the aesthetic without cosplay.

## Design Elements

- **Pill-shaped buttons/badges** - `border-radius: 999px`
- **Accent bars** - Orange left/top/bottom borders on headers
- **Sweep corners** - Rounded L-shapes where accent bars meet
- **Gradient selections** - Fade from accent color to transparent

## CSS Variables

Defined in `MCP.DEFAULT.html`, inherited by all apps:

```css
/* Backgrounds */
--term-bg: #0a0a0f;           /* Dark base */
--term-bg-elevated: #12121a;   /* Panels, cards */
--term-bg-hover: #1a1a24;      /* Hover states */
--term-bg-panel: #0d0d14;      /* Side panels */

/* Borders */
--term-border: #2a2a3a;
--term-border-bright: #3a3a4a;

/* Text */
--term-text: #e0e0e8;          /* Primary */
--term-text-dim: #8888a0;      /* Secondary */
--term-text-muted: #5a5a70;    /* Tertiary */

/* Accent (orange) */
--term-accent: #E07A47;
--term-accent-glow: rgba(224, 122, 71, 0.4);
--term-accent-dim: rgba(224, 122, 71, 0.15);
--term-accent-bright: #ff9966;

/* Status colors */
--term-success: #4ade80;
--term-success-dim: rgba(74, 222, 128, 0.15);
--term-warning: #fbbf24;
--term-warning-dim: rgba(251, 191, 36, 0.15);
--term-danger: #f87171;
--term-danger-dim: rgba(248, 113, 113, 0.15);
--term-info: #60a5fa;

/* Fonts */
--term-mono: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
--term-sans: 'Space Grotesk', system-ui, sans-serif;
```

## MCP Shell Styles

Applied to all apps via `MCP.DEFAULT.html`:

```css
/* Pill-shaped buttons */
sl-button::part(base) {
  border-radius: 999px;
}

/* Pill-shaped badges */
sl-badge::part(base) {
  border-radius: 999px;
  padding: 0 0.75em;
}

/* Status bar - L-shaped sweep corner */
.mcp-status-bar {
  border-top: 3px solid var(--term-accent);
  border-left: 4px solid var(--term-accent);
  border-radius: 8px 0 0 0;
}

/* Status toggles - pill shaped */
.mcp-build-mode-toggle {
  border-radius: 999px;
}

/* LCARS scrollbars - pill-shaped thumbs */
/* Note: Playwright doesn't render border-radius on scrollbar thumbs */
::-webkit-scrollbar {
  width: 16px;
  height: 16px;
}
::-webkit-scrollbar-track {
  background: var(--term-bg);
}
::-webkit-scrollbar-thumb {
  background: var(--term-accent);
  border: 4px solid var(--term-bg);
  border-radius: 8px;
  background-clip: padding-box;
}
::-webkit-scrollbar-thumb:hover {
  background: var(--term-accent-bright);
}

/* Menu button - circular */
.mcp-menu-button {
  border-radius: 999px;
}

/* Dropdown menu - left accent bar */
.mcp-menu-dropdown {
  border-left: 4px solid var(--term-accent);
  border-radius: 0 8px 8px 8px;
}
```

## Reusable Classes

These semantic class names can be styled differently by other themes.

### `.panel-header` - Header bar with bottom accent

Use for panel/section headers with title and action buttons.

```css
.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-bottom: 3px solid var(--term-accent);
  border-radius: 0 0 6px 0;  /* sweep corner */
}

.panel-header h2,
.panel-header h3,
.panel-header h4 {
  margin: 0;
  font-family: var(--term-mono);
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}
```

### `.panel-header-left` - Header bar with left accent

Use for detail panels where accent is on the left side.

```css
.panel-header-left {
  border-left: 4px solid var(--term-accent);
  border-radius: 0 0 0 8px;  /* sweep corner on left */
  padding-left: 12px;
}
```

### `.section-header` - Collapsible section header

Use for expandable/collapsible sections with hover feedback.

```css
.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 8px 0 8px 10px;
  border-left: 3px solid var(--term-border);
  border-radius: 0 0 0 6px;
  transition: border-color 0.15s, color 0.15s;
}

.section-header:hover {
  border-color: var(--term-accent);
  color: var(--term-accent);
}
```

### `.item.selected` - Selected list item

Use for list items with selection state.

```css
.item.selected {
  background: linear-gradient(90deg,
    var(--term-accent-dim) 0%,
    transparent 100%
  );
  border-left: 6px solid var(--term-accent);
  border-radius: 0 6px 6px 0;
}
```

### `.input-area` - Input area with top accent

Use for chat/command input areas.

```css
.input-area {
  border-top: 3px solid var(--term-accent);
  border-radius: 8px 0 0 0;  /* sweep corner */
  padding-top: 12px;
}
```

## Clearances

The floating menu button is positioned at `top: 12px; right: 12px` and is 48px wide. Headers extending to the right edge need ~72px right padding to clear it.

## App-Specific Notes

### app-console

Uses these LCARS patterns:
- `.app-list-header` - panel-header with bottom accent (sweep right)
- `.todo-header` - panel-header with bottom accent (sweep right)
- `.detail-header` - panel-header-left (sweep left)
- `.section-header` - for Requirements, Gaps, Issues sections
- `.app-item.selected` - gradient selection with left accent
- `.chat-input-row` - input-area with top accent
- `.panel-header` (Chat/Lua tabs) - left accent bar
