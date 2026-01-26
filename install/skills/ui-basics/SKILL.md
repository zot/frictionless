---
name: ui-basics
description: UI engine reference - bindings, state management, patterns. Foundation for ui-fast and ui-thorough.
---

# UI Basics

Reference material for ui-engine apps. Load this once, then use `/ui-fast` or `/ui-thorough` for actual work.

## Helper Script

```bash
.ui/mcp status              # Get server status
.ui/mcp run '<lua code>'    # Execute Lua code
.ui/mcp display myapp       # Display app in browser
.ui/mcp browser             # Open browser to UI session
.ui/mcp linkapp add myapp   # Create symlinks
.ui/mcp audit myapp         # Run code quality audit
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
<!-- Select --> <sl-select ui-value="status"><sl-option value="a">A</sl-option></sl-select>
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

# Patterns

## Edit/Cancel Pattern

```lua
function Item:openEditor()
    self._snapshot = { name = self.name }
    self.editing = true
end

function Item:save()
    self._snapshot = nil
    self.editing = false
end

function Item:cancel()
    if self._snapshot then
        self.name = self._snapshot.name
        self._snapshot = nil
    end
    self.editing = false
end
```

## Viewport Fitting

```css
html, body {
  margin: 0;
  padding: 0;
  overflow: hidden;
}
.my-app {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}
.scrollable-area {
  flex: 1;
  min-height: 0;  /* CRITICAL */
  overflow-y: auto;
}
```

---

# Styling

**Put ALL CSS in top-level viewdefs only.** Never in list-item viewdefs.

```html
<template>
  <style>
    .my-app { padding: 1rem; }
    .hidden { display: none !important; }
  </style>
  <div class="my-app">...</div>
</template>
```

---

# MCP Methods

| Method | Description |
|--------|-------------|
| `mcp:status()` | Get server status including `base_dir` |
| `mcp:display(appName)` | Get URL for displaying an app |
| `mcp:appProgress(name, progress, stage)` | Report build progress |
| `mcp:appUpdated(name)` | Trigger dashboard rescan |
| `mcp.pushState(event)` | Send event to Claude agent |
