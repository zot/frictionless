---
name: ui-basics
description: UI engine reference - bindings, state management, patterns. Foundation for ui-fast and ui-thorough.
---

# UI Basics

Frictionless is a platform for personal software that integrates Claude. Users download, mod, create, and share apps — and **this skill exists so you can help them build and modify their apps.**

Load this once for reference, then use `/ui-fast` or `/ui-thorough` for actual work.

## On Skill Load

**IMMEDIATELY invoke `/ui` using the Skill tool before doing anything else.** It covers the directory structure, helper script, debugging, and server lifecycle.

Then invoke `/frontend-design` using the Skill tool if it is available.

Then run this command:

```bash
.ui/mcp patterns
```

## How Frictionless Works

This is not a conventional web framework. It's an **object-oriented system where objects present themselves**.

- **Declarative frontend.** HTML uses `ui-*` bindings, not code. JS is only for browser-native capabilities (file I/O, DOM measurement, timing/animation) — see `.ui/patterns/` for established solutions.
- **Backend is source of truth.** State AND logic live in a persistent Lua session. Page refresh restores the view without resetting state.
- **The frontend creates most variables, not the backend.** The backend creates the root app variable; viewdef binding paths create the rest by reaching into backend objects.
- **Simple paths.** Properties, indexing, methods. No operators.
- **Vanilla Lua tables.** `prototype` tracks the schema and all instances (via weak refs); on hot-reload it detects schema changes and calls `mutate()` on every instance. Fields prefixed with `_` are private — not serialized to the frontend.
- **Viewdefs present objects.** Factoring the object model breaks large viewdefs into smaller ones. Viewdef namespaces select different presentations for the same object.
- **Domain vs Presenter.** Domain objects hold data and core behavior; presenter objects add UI state and actions (e.g., `delete()`, `isEditing`).
- **Three execution contexts:** Lua (all behavior whenever possible — fast, responsive), JS (browser APIs, DOM tricks — last resort), Claude (complex logic, external APIs — slow, event loop latency).
- **The event loop** (reactivity without subscriptions or signals):
  1. User makes a change in the browser
  2. Frontend sends variable update to server (binding paths are variables)
  3. Server applies the update
  4. Server checks ALL variables for changes (including method paths, which are recomputed)
  5. Server sends updates only for variables whose values changed
  6. Frontend applies updates to the UI

  This drives a key pattern: bind a method path to a backend computation, make a change on the frontend, the method fires during the variable check, and the frontend receives the result. Variables have `priority` (low/medium/high) because evaluation order can matter for dependent computations or frontend widget updates.

## Core Principles

1. **Use SOLID principles**
2. **Write idiomatic Lua code**

## Why design.md

`requirements.md` is the user's spec — what they want. `design.md` is Claude's interpretation — a compact intermediate form between requirements and code that serves multiple roles:

- **Verification:** Smaller than code+viewdefs, so the user can quickly read it and confirm their requirements were understood before code is written
- **Preview:** Warns the user what Claude is about to do with the code
- **Reference:** Quick lookup for event handling, data model, and methods
- **Anchor:** Without it, iterative modifications cause **drift** — features silently disappear as code evolves. The `/ui-fast` and `/ui-thorough` workflows enforce: read design → update design → update code → verify against design.

This is a 3-level architecture (as with `/mini-spec`): requirements → design → code. Each level is a progressively more detailed reification of the user's intent.

## File Operations

**ALWAYS use the Write tool** to create/update files. Do NOT use Bash heredocs.

---

# Reactivity & Session Lifecycle

## Hot-Loading

Both Lua and viewdefs hot-load from disk:
- `apps/<app>/app.lua` → re-executed, preserving state
- `apps/<app>/viewdefs/` → browser updates automatically

**Write order matters:** Code first, then viewdefs.

## Change Detection Details

- **Arrays:** Compared by-element. In-place mutations (`table.insert`, `table.remove`) are detected — no need to reassign.
- **`session.reloading`** is only true for hot-reload (file changes), NOT browser page reloads.
- **`ui-code` re-fires on page reload.** Clear `ui-code` properties when their action is complete to prevent stale re-fires.

---

# Object Model

**App directory structure:**
```
apps/my-app/
├── app.lua              # Main code (loads when user first displays the app)
├── init.lua             # Optional startup code (loads on server start)
├── viewdefs/            # HTML viewdefs for this app's types
│   ├── MyApp.DEFAULT.html
│   └── MyApp.Item.list-item.html
├── icon.html            # App icon (Bootstrap Icon)
├── favicon.svg          # Browser tab icon
└── README.md            # App description

.ui/storage/my-app/      # Optional local storage (isolated from app updates)
.ui/html/my-app          # Symlink to app dir (serves static files at /my-app/)
.ui/html/my-app-storage  # Symlink to storage dir (serves at /my-app-storage/)
```

Use `require("appname.module")` to split code into multiple files (see `mcp` app for examples).

```lua
-- App type name is PascalCase; instance global is camelCase of the type name
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
    myApp = MyApp:new()  -- camelCase instance global
end
```

**Key points:**
- `session:prototype(name)` sets the `type` field for viewdef resolution
- `session.reloading` is true during hot-reload
- Each app defines a PascalCase type (`MyApp`) and a camelCase instance (`myApp`)

## Hot-Loading Mutations

When adding fields to a prototype, `mutate()` updates all live instances to match the new schema:

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

**Key rules:**
- **Mutation must be the last change to a file.** Hot-loading is very fast — making mutation the final edit ensures prototype and `mutate()` arrive together in one hot-load.
- **Overwrite `mutate()`, don't accumulate.** Each `mutate()` only handles the current delta — once an instance has been mutated, it already has the field.
- **Use atomic writes** when the user may be interacting with the app:
  ```bash
  cp app.lua app.lua.tmp   # Edit tmp
  mv app.lua.tmp app.lua   # Atomic replace
  ```

## Variable Wrappers

The `?wrapper=TypeName` property transforms a variable's value through a Lua type. ViewList is a built-in wrapper for arrays.

```lua
MyWrapper = session:prototype("MyWrapper", {
    variable = EMPTY,  -- the Variable object
    value = EMPTY,     -- convenience: variable's current value
})

function MyWrapper:new(variable)
    local existing = variable:getWrapper()
    if existing then
        existing.value = variable:getValue()
        return existing
    end
    local wrapper = session:create(MyWrapper)
    wrapper.variable = variable
    wrapper.value = variable:getValue()
    return wrapper
end
```

The wrapper receives the **variable** (not just the value). Check `variable:getWrapper()` to reuse existing wrappers and preserve state. Child paths navigate from the wrapper object.

---

# Bindings

| Attribute | Purpose | Example |
|-----------|---------|---------|
| `ui-value` | Bind value/text | `<sl-input ui-value="name">` |
| `ui-action` | Button click | `<sl-button ui-action="save()">` |
| `ui-event-click` | Any element click | `<div ui-event-click="toggle()">` |
| `ui-event-*` | Any event | `<sl-select ui-event-sl-change="onSelect()">` |
| `ui-event-keypress-*` | Specific key | `<sl-input ui-event-keypress-enter="submit()">` |
| `ui-event-keypress-ctrl-*` | Key + modifiers | `<sl-input ui-event-keypress-ctrl-s="save()">` (also `shift`, `alt`, `meta`) |
| `ui-view` | Render child/list | `<div ui-view="items?wrapper=lua.ViewList">` |
| `ui-attr-*` | HTML attribute | `<sl-alert ui-attr-open="hasError">` |
| `ui-class-*` | CSS class toggle | `<div ui-class-active="isActive">` |
| `ui-style-*` | CSS style | `<div ui-style-color="textColor">` |
| `ui-html` | Inject HTML content | `<div ui-html="description">` (use `?replace` to replace the element itself) |
| `ui-code` | Run JS from property | `<div ui-code="myJsCode">` — binds to a property containing JS, not inline code |
| `ui-namespace` | Set viewdef namespace | `<div ui-namespace="COMPACT">` |

## Common Mistakes

| Wrong | Right |
|-------|-------|
| `ui-action="fn()"` on div | `ui-event-click="fn()"` on div |
| `ui-class="hidden:expr"` | `ui-class-hidden="expr"` |
| `<sl-checkbox ui-value="done">` | `<sl-checkbox ui-attr-checked="done">` |
| `<sl-select ui-event-sl-change="...">` | `<sl-select ui-event-sl-input="...">` (sl-change doesn't fire) |
| `<style>` in list-item viewdef | Put styles in top-level viewdef |
| Operators in paths (`!value`) | Use methods (`isHidden()`) |
| Classes/styles on `ui-view="x?wrapper=lua.ViewList"` | Put them on a wrapper div (ViewList double-replaces, losing classes) |

## Variable Paths

- Property access: `name`, `nested.path`
- Array indexing: `0`, `1` (0-based in paths)
- Parent traversal: `..`
- Method calls: `getName()`, `setValue(_)`
- Path params: `path?wrapper=ViewList`

## Variable Properties

| Property | Values | Description |
|----------|--------|-------------|
| `access` | `r`, `w`, `rw`, `action` | Read/write permissions |
| `wrapper` | Type name | Wrap with this type |
| `keypress` | (flag) | Live update on keystroke |
| `scrollOnOutput` | (flag) | Auto-scroll on changes |
| `itemWrapper` | Type name | Wrap each list item |
| `create` | Type name | Create instance as value |
| `priority` | `low`, `medium`, `high` | Evaluation order during variable check |

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

---

# Styling

**Put ALL CSS in the main app viewdef only** (e.g. `MyApp.DEFAULT.html`). Never in list-item or other sub-viewdefs.

**Theme:** See `.ui/themes/theme.md` for CSS variables, colors, and reusable classes.

### Shoelace Component Styling

**Apps MUST defer Shoelace component styling to themes.** Do NOT add `::part()` overrides for Shoelace components (buttons, inputs, textareas, selects, dialogs, alerts, badges, spinners, progress bars, icon-buttons) in app viewdefs. These are styled by `base.css` (shared defaults) and theme CSS files (theme-specific overrides).

**Architecture:**
- `base.css` provides shared Shoelace defaults using `var(--term-*)` variables (no theme prefix)
- Each theme overrides only what's unique: font-family, border-radius, special effects (e.g. brume's glass/backdrop-filter)
- Theme-prefixed selectors (`.theme-brume sl-button::part(base)`) win over base.css by specificity

**What apps CAN style:** Layout properties (padding, margin, gap, flex, grid), structural properties (font-size on specific elements), and app-specific classes. What they must NOT style: colors, backgrounds, borders, and box-shadows on Shoelace `::part()` selectors.

### Semantic Theme Classes

**Live discovery:** Run `.ui/mcp theme classes` to get the authoritative list of all semantic classes across all installed themes — including user-added themes. Always check this before writing viewdefs for a new app or major feature.

These additional classes are used in viewdefs but not declared as `@class` in theme CSS:

| Class | Description |
|-------|-------------|
| `.item` | Base class for list items |
| `.selected` | Selected state modifier (use with `.item`) |

Compose theme + app classes: `<div class="panel-header app-list-header">` — theme class for styling, app class for layout overrides.

**Theme audit:** Run `.ui/mcp theme audit APP` after writing viewdefs to catch undocumented classes and missed opportunities to use semantic classes.

### Creating Themes

Every theme CSS file **must** have a comment block at the top with these annotations (required by `theme list`, `theme classes`, and `theme audit`):

```css
/*
@theme my-theme
@description Short theme description

@class panel-header
  @description Header bar with bottom accent
  @usage Panel/section headers with title and action buttons
  @elements div, header

@class another-class
  @description What this class is for
  @usage When to use it
  @elements div
*/
```

**Required annotations:**
- `@theme` — theme name (must match the CSS filename without extension)
- `@description` — theme-level description
- `@class` blocks — one per semantic class, each with `@description`, `@usage`, and `@elements`

Before creating a theme, run `.ui/mcp theme classes` to see the existing semantic classes. A new theme should declare and style all of them and may add new ones. `@description` should describe what the class looks like in *this* theme (e.g. "Header bar with soft bottom glow"). `@usage` should be generic and structural — it describes *when* to use the class, not how it looks (e.g. "Panel/section headers with title and action buttons").

## Favicons

Each app has `favicon.svg` in its app directory — a Bootstrap Icon SVG with `fill="#E07A47"`. Add a `<script>` as the **last child** of `<template>` in the DEFAULT viewdef:

```html
<script>document.getElementById('app-favicon').href='data:image/svg+xml;base64,...'</script>
```

Generate the base64: `base64 -w0 apps/myapp/favicon.svg`

The MCP shell app (`mcp`) must NOT set a favicon — it wraps other apps.

---

# JavaScript API

`window.uiApp.updateValue(elementId, value?)` — send a value from JS to Lua (e.g., file pickers, clipboard). See `.ui/patterns/js-to-lua-bridge.md` for the full pattern.

---

# Architecture Balance

Two spectrums to balance:
- Objects: God Object ←→ Ravioli Objects
- Viewdefs: Monolithic ←→ Ravioli Viewdefs

**God object signs** (time to extract):
- 15+ methods mixing concerns on root object
- Multiple "current selections" (selected, selectedResume, editingItem)
- Many `selected.X` paths in viewdef (Law of Demeter smell)
- Proliferating show/hide/is*View methods

**Ravioli signs** (over-factored):
- Jumping between 5+ files to trace a simple flow
- Objects/viewdefs with only 2-3 members
- Factoring for purity rather than benefit

**Extract when:**
- Sub-object has 10+ bindings in viewdef
- View has distinct state that should reset on navigation
- Clear separation of concerns improves maintainability

**Keep together when:**
- Views share most state
- UI is tightly coupled to parent layout
- Separation adds files without clarity

See `.scratch/APP-DESIGN.md` for detailed patterns and examples.

---

# MCP Methods

| Method | Description |
|--------|-------------|
| `mcp:status()` | Get server status including `base_dir` |
| `mcp:display(appName)` | Get URL for displaying an app |
| `mcp:appUpdated(name)` | Trigger dashboard rescan |
| `mcp.pushState(event)` | Send event to Claude agent |

## Progress (visible to user in UI)

| Method | Description |
|--------|-------------|
| `mcp:createTodos(steps, appName)` | Create progress steps (e.g., `{'Write code', 'Write viewdefs'}`) |
| `mcp:startTodoStep(n)` | Mark step n as in-progress |
| `mcp:completeTodos()` | Mark all steps complete |
| `mcp:addAgentMessage(msg)` | Show a message from Claude in the UI |

Use alongside Claude Code's `TaskCreate`/`TaskUpdate` — MCP progress is for user visibility, TaskCreate is for work tracking.
