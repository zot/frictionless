---
name: ui-builder
description: use when **building or modifying ui-engine UIs** with Lua apps connected to widgets
---

# UI Builder

Expert at building ui-engine UIs with Lua apps connected to widgets.

## Prerequisite

**Run `/ui` skill first** if you haven't already. It covers directory structure and how to run UIs after building.

## Hot-Loading

**Both Lua code and viewdefs hot-load automatically from disk.** When you edit files:
- Lua files in `apps/<app>/app.lua` → re-executed, preserving app state
- Viewdef files in `apps/<app>/viewdefs/` → browser updates automatically

**Write order matters:** Write code changes FIRST, then viewdefs. Viewdefs may reference new types/methods that must exist before the viewdef loads.

**Never use `ui_upload_viewdef`** — just write files to disk. The server watches for changes and hot-loads automatically.

## Workflow

1. **Read requirements**: Check `{base_dir}/apps/<app>/requirements.md` first. If it does not exist, create it with human-readable prose (no ASCII art or tables)

2. **Design**:
   - Check `{base_dir}/patterns/` for reusable patterns
   - Write the design in `{base_dir}/apps/<app>/design.md`:
      - **Intent**: What the UI accomplishes
      - **Layout**: ASCII wireframe showing structure
      - **Data Model**: Tables of types, fields, and descriptions
      - **Methods**: Actions each type performs
      - **ViewDefs**: Template files needed
      - **Events**: JSON examples of user interactions

3. **Write files** to `{base_dir}/apps/<app>/` (**code first, then viewdefs**):
   - `design.md` — design spec (first, for reference)
   - `app.lua` — Lua classes and logic (**write this before viewdefs**)
   - `viewdefs/<Type>.DEFAULT.html` — HTML templates (after code exists)
   - `viewdefs/<Item>.list-item.html` — List item templates (if needed)

4. **Create symlinks** using the linkapp script:

   ```bash
   .claude/ui/linkapp add <app>
   ```

## Common Binding Mistakes

These are easy to get wrong:

| Wrong | Right |
|-------|-------|
| `ui-action="fn()"` on div | `ui-event-click="fn()"` on div |
| `ui-class="hidden:isCollapsed()"` | `ui-class-hidden="isCollapsed()"` |
| `ui-viewlist="items"` | `ui-view="items?wrapper=lua.ViewList"` |

`ui-action` only works on buttons. Use `ui-event-click` for other elements.

## Preventing Drift (Updates)

During iterative modifications, features can accidentally disappear:

1. **Before modifying** — Read the design spec (`design.md`)
2. **Update spec first** — Modify the layout/components in the spec
3. **Then update code** — Change viewdef and Lua to match spec
4. **Verify** — Ensure implementation matches spec

The spec is the **source of truth**. If it says a close button exists, don't remove it.

## State Management (Critical)

**Use `session:prototype()` with the idempotent pattern:**

```lua
-- 1. Declare prototype (always runs, preserves identity on reload)
MyApp = session:prototype("MyApp", {
    items = EMPTY,  -- EMPTY = nil but tracked
    name = ""
})

function MyApp:new(instance)
    instance = session:create(MyApp, instance)
    instance.items = instance.items or {}
    return instance
end

-- 2. Guard instance creation (idempotent)
if not session.reloading then
    myApp = MyApp:new()  -- lowercase camelCase matching app directory
end
```

**Why this pattern?**
- `session:prototype()` always runs → methods get updated on hot-reload
- `session.reloading` is true during hot-reload, false on initial load
- Instance creation only runs on first load → idempotent
- `session:create()` tracks instances for hot-reload migrations

**Key points**:
- The prototype MUST have a `type` field (set automatically by `session:prototype()`)
- Viewdefs must exist for that type
- Changes to objects automatically sync to the browser

**Agent-readable state (`mcp.pushState`):**
- Use `mcp.pushState({...})` to send events to the agent
- Events queue up and agent reads them via `/wait` endpoint
- Example: `mcp.pushState({ app = "myapp", event = "chat", text = userInput })`

## Behavior

| Location       | Use For                                           | Trade-offs                             |
|----------------|---------------------------------------------------|----------------------------------------|
| **Lua**        | All behavior whenever possible                    | Simpler, saves tokens, very responsive |
| **Claude**     | "Magical" stuff, complex logic, external APIs     | Slow turnaround (event loop latency)   |
| **JavaScript** | Extending presentation (browser APIs, DOM tricks) | Last resort, harder to maintain        |

**Prefer Lua.** Lua methods execute instantly when users click buttons or type.

## Bindings

| Attribute             | Purpose                     | Example                                                                      |
|:----------------------|:----------------------------|:-----------------------------------------------------------------------------|
| `ui-value`            | Bind value/text             | `<sl-input ui-value="name">` `<span ui-value="total()">`                     |
| `ui-action`           | Click handler (buttons)     | `<sl-button ui-action="save()">`                                             |
| `ui-event-click`      | Click handler (any element) | `<div ui-event-click="toggle()">`                                            |
| `ui-event-*`          | Any event                   | `<sl-select ui-event-sl-change="onSelect()">`                                |
| `ui-event-keypress-*` | Specific key press          | `<sl-input ui-event-keypress-enter="submit()">`                              |
| `ui-view`             | Render child/list           | `<div ui-view="selected">` `<div ui-view="items?wrapper=lua.ViewList">`      |
| `ui-attr-*`           | HTML attribute              | `<sl-alert ui-attr-open="hasError">`                                         |
| `ui-class-*`          | CSS class toggle            | `<div ui-class-active="isActive">`                                           |
| `ui-style-*`          | CSS style                   | `<div ui-style-color="textColor">`                                           |
| `ui-code`             | Run JS on update            | `<div ui-code="jsCode">` (executes JS when value changes)                    |
| `ui-namespace`        | Set viewdef namespace       | `<div ui-namespace="COMPACT"><div ui-view="item"></div></div>`               |

**Keypress bindings:**
`ui-event-keypress-*` fires only when the specified key is pressed (enter, escape, tab, left, right, up, down, space, {letter}):
- `ui-event-keypress-enter` - Enter/Return key
- `ui-event-keypress-escape` - Escape key
- `ui-event-keypress-ctrl-enter` - Ctrl+Enter (modifiers: `ctrl`, `shift`, `alt`, `meta`, which can be combined)

**ui-code binding:**

Execute JavaScript when a variable's value changes. The code has access to:
- `element` - The bound DOM element
- `value` - The new value from the variable
- `variable` - The variable object (for accessing widget/properties)
- `store` - The VariableStore

```html
<!-- Close browser when closeWindow becomes truthy -->
<div ui-code="closeWindow" style="display:none;"></div>
```

```lua
-- In Lua: set the JS code, then trigger it
app.closeWindow = "if (value) window.close()"
-- Later, to close:
app.closeWindow = "window.close()"  -- or set a trigger value
```

Use cases: auto-close window, trigger downloads, custom DOM manipulation, browser APIs.

## Variable Paths

**Path syntax:**
- Property access: `name`, `nested.path`
- Array indexing: `0`, `1` (0-based in paths, 1-based in Lua)
- Parent traversal: `..`
- Method calls: `getName()`, `setValue(_)`
- Path params: `contacts?wrapper=ViewList&item=ContactPresenter`
  - Properties after `?` are set on the created variable
  - Uses URL query string syntax: `key=value&key2=value2`

**IMPORTANT:** No operators in paths! `!`, `==`, `&&`, `+`, etc. are NOT valid. For negation, create a method (e.g., `isCollapsed()` returning `not self.expanded` instead of `!expanded`).

**Truthy values:** Lua `nil` becomes JS `null` which is falsy. Any non-nil value is truthy. Use boolean fields (e.g., `isActive`) or methods returning booleans for class/attr toggles.

**Method path constraints:**  (see Variable Properties)
- Paths ending in `()` (no argument) must have access `r` or `action`
- Paths ending in `(_)` (with argument) must have access `w` or `action`

**Nullish path handling:**

Path traversal uses nullish coalescing (like JavaScript's `?.`). If any segment resolves to `nil`:
- **Read direction:** The binding displays empty/default value instead of erroring
- **Write direction:** Fails gracefully

This allows bindings like `ui-value="selectedContact.firstName"` to work when `selectedContact` is nil (e.g., nothing selected).

## Variable Properties

`<sl-input ui-value="name?prop1=val1,prop2=val2"></sl-input>`

| Property  | Values                                   | Description                                                           |
|-----------|------------------------------------------|-----------------------------------------------------------------------|
| `access`  | `r`, `w`, `rw`, `action`                 | Read/write permissions for variables                                  |
| `wrapper` | Type name (e.g., `lua.ViewList`)         | Wrap with this type                                                   |
| `keypress`| (flag)                                   | Live update on every keystroke                                        |
| `scrollOnOutput` | (flag)                            | Auto-scroll to bottom when content changes                            |
| `item` | wrapper type                                | specify wrapper type for ViewList items                               |

## Widgets

```html
<!-- Text --> <span ui-value="name"></span> <div ui-value="compute()"></div>
<!-- Input --> <sl-input ui-value="email" label="Email"></sl-input>
<!-- Live --> <sl-input ui-value="search?keypress">
<!-- Button --> <sl-button ui-action="save()">Save</sl-button>
<!-- Select --> <sl-select ui-value="status"><sl-option value="a">A</sl-option></sl-select>
<!-- Check --> <sl-checkbox ui-value="agreed">Agree</sl-checkbox>
<!-- Switch --> <sl-switch ui-value="dark">Dark</sl-switch>
<!-- Rating --> <sl-rating ui-value="stars"></sl-rating>
<!-- Hide --> <div ui-class-hidden="isHidden()">Content</div>
<!-- Alert --> <sl-alert ui-attr-open="err" variant="danger"><span ui-value="msg"></span></sl-alert>
<!-- Child --> <div ui-view="selectedItem"></div>
```

## Lists

**Standard pattern (using ui-view with wrapper):**

```html
<!-- In app viewdef -->
<div ui-view="items?wrapper=lua.ViewList"></div>
```

**IMPORTANT:** Always use `ui-view` with `wrapper=lua.ViewList` for lists, which wrap their items in with ui-view attributes.

**Item viewdef (`lua.ViewListItem.list-item.html`):**

```html
<template>
  <div><span ui-value="item.name"></span></div>
</template>
```

## List Item Viewdef Context (Critical)

**When creating viewdefs for list items, the item IS the direct context.**

**CORRECT** (item is the context):

```html
<!-- TreeItem.list-item.html -->
<div class="tree-item" ui-action="invoke()">
  <span ui-value="name"></span>
</div>
```

**WRONG** (would look for `item.name` on the TreeItem, which doesn't exist):

```html
<!-- TreeItem.list-item.html - BROKEN -->
<div class="tree-item" ui-action="item.invoke()">
  <span ui-value="item.name"></span>
</div>
```

**Why this happens:**
1. ViewList wraps each array element in a ViewListItem
2. ViewListItem's viewdef (`lua.ViewListItem.list-item.html`) renders `<div ui-view="item"></div>`
3. This `ui-view="item"` makes the wrapped item the context for its own viewdef
4. So inside `TreeItem.list-item.html`, `name` refers to `TreeItem.name` directly

**Where you DO use `item.` prefix:**
- Inside `lua.ViewListItem.list-item.html` itself (where `item` is a property of ViewListItem)
- NOT in the item's own viewdef (e.g., `TreeItem.list-item.html`)

## ui-viewlist is a special case for lists

`ui-viewlist` is lower level and only used when element children must have a specific type, like `sl-select` with its `sl-option` items. Otherwise use a regular list.

`ui-viewlist` expects an exemplar child element which it will clone and use with `ui-view` for each list item.

Example:

```html
<sl-select ui-viewlist="options">
  <sl-option></sl-option>
</sl-select>
```

The `<sl-option>` tells the frontend to use `<sl-option ui-view="options[N]"></sl-option>` for each item `N`.

## Lua Pattern

```lua
-- 1. Declare prototypes (always runs, updates methods)
MyApp = session:prototype("MyApp", {
    items = EMPTY,
    name = ""
})

function MyApp:new(instance)
    instance = session:create(MyApp, instance)
    instance.items = instance.items or {}
    return instance
end

function MyApp:add()
    local item = session:create(Item, { name = self.name })
    table.insert(self.items, item)
    self.name = ""
end

Item = session:prototype("Item", { name = "" })

-- 2. Guard instance creation and any other immediately run code (idempotent)
if not session.reloading then
    -- assign the app variable here
    myApp = MyApp:new()
end
```

## Styling

**Put all CSS in top-level object viewdefs, NOT in index.html.**

```html
<template>
  <style>
    .my-app { padding: 1rem; }
    .hidden { display: none !important; }
  </style>
  <div class="my-app">...</div>
</template>
```

**Tips:**
- Put all styles in a `<style>` block in top-level object viewdefs
- These styles apply to the entire rendered tree including nested viewdefs
- Use Shoelace CSS variables (e.g., `var(--sl-spacing-medium)`) for consistency
- The `.hidden` utility class is commonly needed for `ui-class-hidden` bindings

## Complete Example

See the `examples/` directory for a complete Contact Manager with Chat:
- `examples/design.md` — Design spec
- `examples/code.lua` — Lua code
- `examples/ContactApp.DEFAULT.html` — App viewdef
- `examples/Contact.list-item.html` — Contact item viewdef
- `examples/ChatMessage.list-item.html` — Chat message viewdef
