---
name: ui-basics
description: UI engine reference - bindings, state management, patterns. Foundation for ui-fast and ui-thorough.
---

# UI Basics

Reference material for ui-engine apps. Load this once, then use `/ui-fast` or `/ui-thorough` for actual work.

Make sure `/frontend-design` is loaded if available.

## On Skill Load

**Run this command immediately:**

```bash
.ui/mcp patterns
```

This shows available patterns in `.ui/patterns/` - reusable solutions for common ui-engine problems.

## Core Principles

1. **Use SOLID principles**
2. **Use Object-oriented principles**
3. **Write idiomatic Lua code** 

## Helper Script

```bash
.ui/mcp status              # Get server status
.ui/mcp run '<lua code>'    # Execute Lua code
.ui/mcp display myapp       # Display app in browser
.ui/mcp browser             # Open browser to UI session
.ui/mcp linkapp add myapp   # Create symlinks
.ui/mcp audit myapp         # Run code quality audit
.ui/mcp patterns            # List available patterns
```

## File Operations

**ALWAYS use the Write tool** to create/update files. Do NOT use Bash heredocs.

---

# State Management

## Prototype Pattern

```lua
-- Declare app prototype (serves as namespace)
MyApp = session:prototype("MyApp", {
    items = EMPTY,  -- EMPTY: starts nil, tracked for mutation
    name = ""
})

-- Nested prototypes use dotted names
MyApp.Item = session:prototype("MyApp.Item", { name = "" })
local Item = MyApp.Item  -- local shortcut

function MyApp:new(instance)
    instance = session:create(MyApp, instance)
    instance.items = instance.items or {}
    return instance
end

-- Guard instance creation (idempotent)
if not session.reloading then
    myApp = MyApp:new()  -- global name = camelCase of app dir
end
```

**Key points:**
- `session:prototype(name)` sets the `type` field for viewdef resolution
- `session.reloading` is true during hot-reload
- Each app creates two globals: `Name` (prototype) and `name` (instance)

## Hot-Loading

Both Lua and viewdefs hot-load from disk:
- `apps/<app>/app.lua` → re-executed, preserving state
- `apps/<app>/viewdefs/` → browser updates automatically

**Write order matters:** Code first, then viewdefs.

## Hot-Loading Mutations

When adding fields to a prototype, existing instances need initialization:

```lua
MyApp = session:prototype("MyApp", {
    items = EMPTY,
    newField = EMPTY  -- NEW field
})

function MyApp:mutate()
    if self.newField == nil then
        self.newField = {}
    end
end
```

**CRITICAL:** Field addition + `mutate()` must arrive in a SINGLE hot-load. Use atomic writes:

```bash
cp app.lua app.lua.tmp   # Edit tmp
mv app.lua.tmp app.lua   # Atomic replace
```

---

# Bindings

| Attribute | Purpose | Example |
|-----------|---------|---------|
| `ui-value` | Bind value/text | `<sl-input ui-value="name">` |
| `ui-action` | Button click | `<sl-button ui-action="save()">` |
| `ui-event-click` | Any element click | `<div ui-event-click="toggle()">` |
| `ui-event-*` | Any event | `<sl-select ui-event-sl-change="onSelect()">` |
| `ui-event-keypress-*` | Specific key | `<sl-input ui-event-keypress-enter="submit()">` |
| `ui-view` | Render child/list | `<div ui-view="items?wrapper=lua.ViewList">` |
| `ui-attr-*` | HTML attribute | `<sl-alert ui-attr-open="hasError">` |
| `ui-class-*` | CSS class toggle | `<div ui-class-active="isActive">` |
| `ui-style-*` | CSS style | `<div ui-style-color="textColor">` |
| `ui-code` | Run JS on update | `<div ui-code="jsCode">` |
| `ui-namespace` | Set viewdef namespace | `<div ui-namespace="COMPACT">` |

## Common Mistakes

| Wrong | Right |
|-------|-------|
| `ui-action="fn()"` on div | `ui-event-click="fn()"` on div |
| `ui-class="hidden:expr"` | `ui-class-hidden="expr"` |
| `<sl-checkbox ui-value="done">` | `<sl-checkbox ui-attr-checked="done">` |
| `<style>` in list-item viewdef | Put styles in top-level viewdef |
| Operators in paths (`!value`) | Use methods (`isHidden()`) |

## Variable Paths

- Property access: `name`, `nested.path`
- Array indexing: `0`, `1` (0-based in paths)
- Parent traversal: `..`
- Method calls: `getName()`, `setValue(_)`
- Path params: `path?wrapper=ViewList`

**No operators in paths.** For negation, create a method.

## Variable Properties

| Property | Values | Description |
|----------|--------|-------------|
| `access` | `r`, `w`, `rw`, `action` | Read/write permissions |
| `wrapper` | Type name | Wrap with this type |
| `keypress` | (flag) | Live update on keystroke |
| `scrollOnOutput` | (flag) | Auto-scroll on changes |
| `itemWrapper` | Type name | Wrap each list item |
| `create` | Type name | Create instance as value |

---

# Widgets

```html
<!-- Text --> <span ui-value="name"></span>
<!-- Input --> <sl-input ui-value="email" label="Email"></sl-input>
<!-- Live --> <sl-input ui-value="search?keypress">
<!-- Button --> <sl-button ui-action="save()">Save</sl-button>
<!-- Select or Dropdown --> <sl-select ui-value="status"><sl-option value="a">A</sl-option></sl-select>
<!-- Check --> <sl-checkbox ui-attr-checked="agreed">Agree</sl-checkbox>
<!-- Switch --> <sl-switch ui-attr-checked="dark">Dark</sl-switch>
<!-- Rating --> <sl-rating ui-value="stars"></sl-rating>
<!-- Hide --> <div ui-class-hidden="isHidden()">Content</div>
<!-- Alert --> <sl-alert ui-attr-open="err" variant="danger"><span ui-value="msg"></span></sl-alert>
<!-- Badge --> <sl-badge variant="success"><span ui-value="count"></span></sl-badge>
<!-- Child --> <div ui-view="selectedItem"></div>
```

---

# Lists

```html
<div ui-view="items?wrapper=lua.ViewList"></div>
```

List item viewdef (`MyApp.Item.list-item.html`):
```html
<template>
  <div ui-event-mousedown="select()">
    <span ui-value="name"></span>
  </div>
</template>
```

**In list-item viewdefs, the item IS the context.** Use `name`, not `item.name`.

## Select Dropdowns with Dynamic Options

```html
<sl-select ui-value="selectedId" label="Pick one">
  <span ui-view="items()?wrapper=lua.ViewList" ui-namespace="my-option"></span>
</sl-select>
```

Viewdef (`lua.ViewListItem.my-option.html`):
```html
<template>
  <sl-option ui-attr-value="index">
    <span ui-value="item.name"></span>
  </sl-option>
</template>
```

---

# Common Patterns

**Run `.ui/mcp patterns` to see available patterns** in `.ui/patterns/`.

---

# Styling

**Put ALL CSS in the main app viewdef only** (e.g. `MyApp.DEFAULT.html`). Never in list-item viewdefs or other non-main viewdefs.

```html
<template>
  <style>
    .my-app { padding: 1rem; }
    .hidden { display: none !important; }
  </style>
  <div class="my-app">...</div>
</template>
```

**Theme:** See `.ui/themes/theme.md` for CSS variables, colors, and reusable classes. Apps inherit base component styles from the MCP shell.

### Semantic Theme Classes

| Class | Description | Usage |
|-------|-------------|-------|
| `.panel-header` | Header bar with bottom accent | Panel/section headers with title and action buttons |
| `.panel-header-left` | Header bar with left accent | Detail panels where accent is on the left side |
| `.section-header` | Collapsible section header | Expandable/collapsible sections with hover feedback |
| `.item` | Base class for list items | Standard list item styling |
| `.selected` | Selected state modifier | Apply to items with selection state (use with `.item`) |
| `.input-area` | Input area with top accent | Chat/command input areas |

**Compose theme + app classes** (Tailwind-style):
```html
<div class="panel-header app-list-header">
```
- Theme class (`.panel-header`) - provides themed styling (accent bars, sweeps)
- App class (`.app-list-header`) - adds app-specific layout/overrides

This keeps theming swappable while preserving app-specific needs.

**Auditing theme usage:** Run `.ui/mcp theme audit myapp` to check which classes are documented vs app-specific.

---

# MCP Methods

| Method | Description |
|--------|-------------|
| `mcp:status()` | Get server status including `base_dir` |
| `mcp:display(appName)` | Get URL for displaying an app |
| `mcp:appProgress(name, progress, stage)` | Report build progress |
| `mcp:appUpdated(name)` | Trigger dashboard rescan |
| `mcp.pushState(event)` | Send event to Claude agent |
