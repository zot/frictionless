---
name: ui-builder
description: **Use proactively** Use for any **UI** design, building, modifying, or testing. This is the required process for managing **all ui app code**. This supercedes other applicable skills when dealing with ui apps.
---

# UI Builder

Expert at building ui-engine UIs with Lua apps connected to widgets.

## Prerequisites

1. **Run `/ui` skill first** if you haven't already. It covers directory structure and how to run UIs after building.

## Helper Script

Use the `.ui/mcp` script for all MCP operations:

```bash
.ui/mcp status                      # Get server status
.ui/mcp run '<lua code>'            # Execute Lua code
.ui/mcp progress app % msg          # Report build progress
.ui/mcp linkapp add myapp           # Create symlinks
.ui/mcp audit myapp                 # Run code quality audit
.ui/mcp browser                     # Open browser to UI session
.ui/mcp display myapp               # Display app in browser
.ui/mcp event                       # Wait for next UI event
```

The script reads the MCP port automatically from `.ui/mcp-port` file.

## File Operations

**ALWAYS use the Write tool to create/update files.** Do NOT use Bash heredocs (`cat > file << 'EOF'`).

- Write tool works in background agents and allows user approval
- Bash heredocs are denied in background agents

## Core Principles
- use SOLID principles, comprehensive unit tests
- when adding code, verify whether it needs to be factored
- Code and specs as MINIMAL as possible
- write idiomatic Lua code

## Hot-Loading

**Both Lua code and viewdefs hot-load automatically from disk.** When you edit files:
- Lua files in `apps/<app>/app.lua` → re-executed, preserving app state
- Viewdef files in `apps/<app>/viewdefs/` → browser updates automatically
- `session.reloading` is `true` during reload, `false` otherwise — use to detect hot-reloads

**Write order matters:** Write code changes FIRST, then viewdefs. Viewdefs may reference new types/methods that must exist before the viewdef loads.

**Just write files to disk.** The server watches for changes and hot-loads automatically.

## Progress Reporting

Report build progress so the apps dashboard shows status:

```bash
.ui/mcp progress <app> <percent> <stage>
```

**Call progress at each phase:**

| Phase | Command |
|-------|---------|
| Starting | `.ui/mcp progress myapp 0 "starting..."` |
| Reading requirements | `.ui/mcp progress myapp 10 "reading requirements..."` |
| Designing | `.ui/mcp progress myapp 20 "designing..."` |
| Writing code | `.ui/mcp progress myapp 40 "writing code..."` |
| Writing viewdefs | `.ui/mcp progress myapp 60 "writing viewdefs..."` |
| Linking | `.ui/mcp progress myapp 80 "linking..."` |
| Auditing | `.ui/mcp progress myapp 90 "auditing..."` |
| Simplifying | `.ui/mcp progress myapp 95 "simplifying..."` |
| Complete | `.ui/mcp progress myapp 100 "complete"` |

**After all files are written**, trigger dashboard rescan:

```bash
.ui/mcp run "mcp:appUpdated('myapp')"
```

**When done**, send a final message (clears thinking status):

```bash
.ui/mcp run "if appConsole then appConsole:addAgentMessage('Done - brief description') end"
```

## MCP Global Methods

The `mcp` global provides methods for interacting with the MCP server:

| Method | Returns | Description |
|--------|---------|-------------|
| `mcp:status()` | table | Get server status including `base_dir` |
| `mcp:display(appName)` | string | Get URL for displaying an app (for iframes) |
| `mcp:appProgress(name, progress, stage)` | nil | Report build progress to dashboard |
| `mcp:appUpdated(name)` | nil | Trigger dashboard rescan after file changes |
| `mcp.pushState(event)` | nil | Send event to Claude agent |

**Important:** `mcp:display(appName)` expects a **string** app name, not an object. If you have an AppInfo object, pass `appInfo.name`.

**Important:** do not display the app after building it unless the user specifically requests it.

## Workflow

**At the start**, create tasks for Claude Code and the UI todo panel:

```
-- Claude Code tasks
TaskCreate: "Read requirements" (activeForm: "Reading requirements...")
TaskCreate: "Design changes" (activeForm: "Designing...")
TaskCreate: "Write code" (activeForm: "Writing code...")
TaskCreate: "Write viewdefs" (activeForm: "Writing viewdefs...")
TaskCreate: "Link and audit" (activeForm: "Auditing...")
TaskCreate: "Simplify code" (activeForm: "Simplifying...")

-- MCP todo panel (syncs progress bar + thinking messages)
.ui/mcp run "mcp:createTodos({'Read requirements', 'Design', 'Write code', 'Write viewdefs', 'Link and audit', 'Simplify'}, '<app>')"
```

Then work through each step, updating both TaskUpdate AND mcp:startTodoStep.

1. → TaskUpdate(status: in_progress), `.ui/mcp run "mcp:startTodoStep(1)"`, **Check for test issues**: If `{base_dir}/apps/<app>/TESTING.md` exists, read it and offer to resolve any Known Issues before proceeding

2. → **Read requirements** (step 1 already started above)
   - Check `{base_dir}/apps/<app>/requirements.md` first
   - If it does not exist, create it with human-readable prose (no ASCII art or tables)

3. → TaskUpdate("Read requirements": completed), TaskUpdate("Design changes": in_progress), `.ui/mcp run "mcp:startTodoStep(2)"`, **Design**
   - Check `{base_dir}/patterns/` for reusable patterns
   - Write `{base_dir}/apps/<app>/icon.html` with an emoji, `<sl-icon>`, or `<img>` element representing the app
   - Write the design in `{base_dir}/apps/<app>/design.md`:
      - **Intent**: What the UI accomplishes
      - **Layout**: ASCII wireframe showing structure
      - **Data Model**: Tables of types, fields, and descriptions
      - **Methods**: Actions each type performs
      - **ViewDefs**: Template files needed
      - **Events**: JSON examples of user interactions with **complete handling instructions**
        - Claude reads design.md to understand how to handle events — requirements.md may have detailed event handling that must be copied to design.md
        - Include: event name, JSON payload example, and exactly what Claude should do (spawn agent, call `.ui/mcp run`, respond via method, etc.)
        - If requirements.md has a "Claude Event Handling" or similar section, transfer all that information to design.md

4. → TaskUpdate("Design changes": completed), TaskUpdate("Write code": in_progress), `.ui/mcp run "mcp:startTodoStep(3)"`, **Write files** to `{base_dir}/apps/<app>/` (**code first, then viewdefs**):
   - `design.md` — design spec (first, for reference)
   - `app.lua` — Lua classes and logic (**write this before viewdefs**)
   - → TaskUpdate("Write code": completed), TaskUpdate("Write viewdefs": in_progress), `.ui/mcp run "mcp:startTodoStep(4)"`
   - `viewdefs/<Type>.DEFAULT.html` — HTML templates (after code exists)
   - `viewdefs/<Item>.list-item.html` — List item templates (if needed)

5. → TaskUpdate("Write viewdefs": completed), TaskUpdate("Link and audit": in_progress), `.ui/mcp run "mcp:startTodoStep(5)"`, **Create symlinks**

   ```bash
   .ui/mcp linkapp add <app>
   ```

6. → **Audit** (part of step 5, no new todo step):

   **Automated checks first** (via HTTP API):
   ```bash
   .ui/mcp audit $APP
   ```

   The tool checks Lua code AND viewdefs for:
   - Dead methods (defined but never called)
   - Missing `session.reloading` guard on instance creation
   - Global variable name doesn't match app directory
   - `<style>` blocks in list-item viewdefs
   - `item.` prefix in list-item viewdefs
   - `ui-action` on non-buttons
   - `ui-class="hidden:..."` (should use `ui-class-hidden`)
   - `ui-value` on checkboxes/switches
   - Operators in binding paths
   - HTML parse errors in viewdefs

   **Do not manually check viewdefs** — the tool handles all viewdef validation.

   **AI-based checks** (require reading comprehension):
   - Compare design.md against requirements.md — **every required feature must be represented**
   - Compare implementation against design.md — **every designed feature must be implemented**
   - Compare implementation against requirements.md — **every principle must be followed**
   - Feature gaps are violations that must be fixed before the task is complete
   - **Responsibility verification:** When requirements.md has explicit responsibility sections (e.g., "Lua Responsibilities", "Claude Responsibilities", "Data Flow"):
     - For each stated Lua responsibility, find the Lua code that implements it — if no code exists or it just sends an event to Claude, it's a violation
     - For each stated Claude responsibility, verify Lua does NOT implement it (Claude handles via events)
     - If requirements say "Lua-driven" or "all X happens in Lua", verify Lua actually does the work
     - Example violation: Requirements say "Lua scans directories" but code sends `refresh_request` event (Claude does scanning)
   - Check for missing `min-height: 0` on scrollable flex children
   - Check that Cancel buttons revert changes (see Edit/Cancel Pattern)

   **Fix violations** before the task is complete:

   1. **Dead methods NOT in design.md** → Delete them from `app.lua` now
   2. **Dead methods IN design.md** → Record in `TESTING.md` under `## Gaps` (design/code mismatch)
   3. **Other violations** (viewdef issues, missing guards) → Fix them in the code
   4. **Warnings** (external methods) → OK to ignore, these are called by Claude

   After fixing, **report any recorded gaps** to the user.

7. → TaskUpdate("Link and audit": completed), TaskUpdate("Simplify code": in_progress), `.ui/mcp run "mcp:startTodoStep(6)"`, **Simplify**

   Use the `code-simplifier` agent to refine the app code for clarity, consistency, and maintainability:

   ```
   Task tool with subagent_type="code-simplifier"
   prompt: "Simplify the code in {base_dir}/apps/<app>/app.lua"
   ```

   The agent will analyze and refine the Lua code while preserving functionality.

8. → TaskUpdate("Simplify code": completed), `.ui/mcp run "mcp:completeTodos()"`, **Complete**
   - Trigger dashboard rescan:
     ```bash
     .ui/mcp run "mcp:appUpdated('$APP')"
     ```

## Common Binding Mistakes

These are easy to get wrong:

| Wrong | Right |
|-------|-------|
| `ui-action="fn()"` on div | `ui-event-click="fn()"` on div |
| `ui-class="hidden:isCollapsed()"` | `ui-class-hidden="isCollapsed()"` |
| `ui-viewlist="items"` | `ui-view="items?wrapper=lua.ViewList"` |
| `<sl-checkbox ui-value="done">` | `<sl-checkbox ui-attr-checked="done">` |
| `<style>` in list-item viewdef | Put all styles in top-level viewdef |
| Save/Cancel both call `close()` | Save commits, Cancel restores snapshot |

`ui-action` only works on buttons. Use `ui-event-click` for other elements.
`ui-value` on checkboxes/switches renders the boolean as text. Use `ui-attr-checked` for display + event handler for changes:
```html
<sl-checkbox ui-attr-checked="done" ui-event-sl-change="toggle()">
```

## Edit/Cancel Pattern (Critical)

**Cancel must revert changes.** When editing with Save/Cancel buttons, Cancel restores the original values. Use snapshot/restore:

```lua
-- Nested prototype for tasks
Tasks.Task = session:prototype("Tasks.Task", {
    name = "",
    description = "",
    done = false,
    editing = false,
    _snapshot = EMPTY  -- stores original values
})
local Task = Tasks.Task

function Task:openEditor()
    -- Snapshot current values before editing
    self._snapshot = {
        name = self.name,
        description = self.description,
        done = self.done
    }
    self.editing = true
end

function Task:save()
    -- Just close - live bindings already updated the values
    self._snapshot = nil
    self.editing = false
end

function Task:cancel()
    -- Restore from snapshot
    if self._snapshot then
        self.name = self._snapshot.name
        self.description = self._snapshot.description
        self.done = self._snapshot.done
        self._snapshot = nil
    end
    self.editing = false
end
```

**Viewdef usage:**
```html
<sl-button ui-action="save()">Save</sl-button>
<sl-button ui-action="cancel()">Cancel</sl-button>
```

**Key points:**
- Snapshot on `openEditor()`, not on each keystroke
- `save()` discards snapshot (changes already applied via live binding)
- `cancel()` restores snapshot values
- Both clear snapshot and close editor

## Preventing Drift (Updates)

During iterative modifications, features can accidentally disappear:

1. **Before modifying** — Read the design spec (`design.md`)
2. **Update spec first** — Modify the layout/components in the spec
3. **Then update code** — Change viewdef and Lua to match spec
4. **Verify** — Ensure implementation matches spec

The spec is the **source of truth**. If it says a close button exists, don't remove it.

## State Management (Critical)

**Use `session:prototype()` with the namespace pattern:**

```lua
-- 1. Declare app prototype (serves as namespace)
-- init declares instance fields — only these are tracked for mutation
MaLuba = session:prototype("MaLuba", {
    items = EMPTY,  -- EMPTY: starts nil, but tracked for mutation
    name = ""
})

-- 2. Nested prototypes use dotted names
MaLuba.Item = session:prototype("MaLuba.Item", { name = "" })
local Item = MaLuba.Item  -- local shortcut

function MaLuba:new(instance)
    instance = session:create(MaLuba, instance)
    instance.items = instance.items or {}
    return instance
end

-- 3. Guard instance creation (idempotent)
if not session.reloading then
    maLuba = MaLuba:new()  -- global name = camelCase of app directory (ma-luba → maLuba)
end
```

**Why this pattern?**
- `session:prototype(name)` accepts arbitrary names (does not consult globals)
- The `name` becomes the prototype's `type` field, used for viewdef resolution (e.g., `"MaLuba.Item"` → `MaLuba.Item.list-item.html`)
- Each app creates only two globals: `Name` (prototype/namespace) and `name` (instance)
- Nested prototypes use dotted names: `MaLuba.Item` registered as `"MaLuba.Item"`
- `session.reloading` is true during hot-reload, false on initial load
- Instance creation only runs on first load → idempotent
- `session:create()` tracks instances for hot-reload migrations

**EMPTY pattern:**
Use `EMPTY` to declare optional fields that start nil but are tracked for mutation. When you remove a field from init, it's nil'd out on all instances.

**Key points**:
- The `type` field is set automatically by `session:prototype()` from the name argument
- Viewdefs must exist for that type (e.g., `MaLuba.DEFAULT.html`, `MaLuba.Item.list-item.html`)
- Changes to objects automatically sync to the browser

**Agent-readable state (`mcp.pushState`):**
- Use `mcp.pushState({...})` to send events to the agent
- Events queue up and agent reads them via `/wait` endpoint
- Example: `mcp.pushState({ app = "myapp", event = "chat", text = userInput })`

## Hot-Loading Mutations (Critical Timing)

When adding new fields to a prototype, existing instances need initialization. Use the `mutate()` method:

```lua
MaLuba = session:prototype("MaLuba", {
    items = EMPTY,
    name = "",
    newField = EMPTY  -- NEW: added in this change
})

function MaLuba:mutate()
    -- Initialize newField for existing instances
    if self.newField == nil then
        self.newField = {}
    end
end
```

**CRITICAL: All field additions and their `mutate()` methods must arrive in a SINGLE hot-load.**

If they arrive in separate hot-loads, it fails silently:

| Hot-load | What Happens | Result |
|----------|--------------|--------|
| 1st: Add field | Calls `mutate()` | `mutate()` doesn't exist yet → field stays nil |
| 2nd: Add `mutate()` | Checks for init changes | Prototype init unchanged → `mutate()` not called |

**Why this happens:** Hot-reload only calls `mutate()` when the prototype's init table changes. Adding a method doesn't change the init table, so the second hot-load doesn't trigger mutation.

**Solution: Use atomic writes via temp file**

Hot-loading only watches files that are already loaded. Write to a temp copy, make all your changes, then `mv` to trigger exactly one hot-load:

```bash
# 1. Copy to temp file (not watched)
cp {base_dir}/apps/myapp/app.lua {base_dir}/apps/myapp/app.lua.tmp

# 2. Make ALL changes to the temp file
#    - Add new fields to prototype init
#    - Add/update mutate() method
#    - Add new methods
#    (multiple edits here don't trigger hot-loads)

# 3. Audit the temp file (see below)

# 4. Atomic replace triggers single hot-load
mv {base_dir}/apps/myapp/app.lua.tmp {base_dir}/apps/myapp/app.lua
```

**Audit the finished temp file before mv:**

1. **Identify new fields**: Compare temp file's prototype init against original
2. **Check for table/array fields**: Look for fields with `EMPTY` or `{}` defaults
3. **Verify mutate() coverage**: For each new table/array field, confirm `mutate()` initializes it
4. **Fix if needed**: Edit the temp file again (still no hot-load), then mv

If you're adding `outputLines = EMPTY`, your mutate() must have:
```lua
function App:mutate()
    if self.outputLines == nil then
        self.outputLines = {}
    end
end
```

**When mutate() is needed:**
- Adding array/table fields (need `{}` initialization)
- Adding fields that other code expects to be non-nil
- Removing fields (set to `nil` to clear from existing instances)
- Not needed for simple values with sensible nil defaults

**mutate() rules:**
- **Idempotent**: Must be safe to run multiple times (use `if self.field == nil then`)
- **Replaceable**: Contents can be completely rewritten each hot-load — no need to preserve old mutation code
- **Runs on all instances**: Called after hotload for every tracked instance whose prototype init changed

## Behavior

| Location       | Use For                                           | Trade-offs                             |
|----------------|---------------------------------------------------|----------------------------------------|
| **Lua**        | All behavior whenever possible                    | Simpler, saves tokens, very responsive |
| **Claude**     | "Magical" stuff, complex logic, external APIs     | Slow turnaround (event loop latency)   |
| **JavaScript** | Extending presentation (browser APIs, DOM tricks) | Last resort, harder to maintain        |

**Prefer Lua.** Lua methods execute instantly when users click buttons or type.

**JavaScript is available via:**
- `<script>` elements in viewdefs — static "library" code loaded once
- `ui-code` attribute — dynamic injection as-needed (see ui-code binding below)

**When JS is needed:**
- **App-local JS** (resize handlers, DOM tricks): Use `<script>` tags in the viewdef (after root element, before `</template>`)
- **Claude-triggered JS** (remote execution): The MCP shell provides `mcp.code` via `ui-code="code"` binding. Claude sets `mcp.code = "window.close()"` to execute JS remotely. Don't add `ui-code` bindings in apps — use the existing MCP shell capability.

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

**Namespace resolution (3-tier):**

When resolving which viewdef to use for a type:

1. Variable's `namespace` property → `Type.{namespace}`
2. Variable's `fallbackNamespace` property → `Type.{fallbackNamespace}`
3. Default → `Type.DEFAULT`

```html
<!-- Explicit namespace via ui-namespace -->
<div ui-view="contact" ui-namespace="COMPACT"/>

<!-- ViewList sets fallbackNamespace="list-item" automatically -->
<div ui-view="contacts?wrapper=lua.ViewList"/>
```

## Variable Paths

**Path syntax:**
- Property access: `name`, `nested.path`
- Array indexing: `0`, `1` (0-based in paths, 1-based in Lua)
- Parent traversal: `..`
- Method calls: `getName()`, `setValue(_)`
- Path params: `contacts?wrapper=ViewList&item=ContactPresenter`
  - Properties after `?` are set on the created variable
  - Uses URL query string syntax: `key=value&key2=value2`

**IMPORTANT:** No operators in paths. Operators like negation, equality, logical-and, plus, etc. are NOT valid. For negation, create a method (e.g., `isCollapsed()` returning `not self.expanded` instead of using the negation operator).

**Truthy values:** Lua `nil` becomes JS `null` which is falsy. Any non-nil value is truthy. Use boolean fields (e.g., `isActive`) or methods returning booleans for class/attr toggles.

**Method path constraints:**  (see Variable Properties)
- Paths ending in `()` (no argument) must have access `r` or `action`
- Paths ending in `(_)` (with argument) must have access `w` or `action`

**Nullish path handling:**

Path traversal uses nullish coalescing (like JavaScript's `?.`). If any segment resolves to `nil`:
- **Read direction:** The binding displays empty/default value instead of erroring
- **Write direction:** Fails gracefully

This allows bindings like `ui-value="selectedContact.firstName"` to work when `selectedContact` is nil (e.g., nothing selected).

### Read/Write Method Paths

Methods can act as read/write properties by ending the path in `()` with `access=rw`:

```html
<input ui-value="value()?access=rw">
```

On read, the method is called with no arguments. On write, the value is passed as an argument. In Lua, use varargs:

```lua
function MyPresenter:value(...)
    if select('#', ...) > 0 then
        self._value = select(1, ...)  -- write
    end
    return self._value  -- read
end
```

## Variable Properties

`<sl-input ui-value="name?prop1=val1&prop2=val2"></sl-input>`

**Only use properties listed here.** Do not invent new properties like `negate=true` — they don't exist. For boolean inversions, create a Lua method (e.g., `notEditing()` returning `not self.editing`).

| Property  | Values                                   | Description                                                           |
|-----------|------------------------------------------|-----------------------------------------------------------------------|
| `access`  | `r`, `w`, `rw`, `action`                 | Read/write permissions for variables                                  |
| `wrapper` | Type name (e.g., `lua.ViewList`)         | Wrap with this type                                                   |
| `keypress`| (flag)                                   | Live update on every keystroke                                        |
| `scrollOnOutput` | (flag)                            | Auto-scroll to bottom when content changes                            |
| `item` | wrapper type                                | Specify wrapper type for ViewList items                               |
| `create` | Type name (e.g., `Contact`)              | Create instance of this type as variable value                        |

**Default ui-value access by element type:**
- Native inputs (`input`, `textarea`, `select`): `rw`
- Interactive Shoelace (`sl-input`, `sl-textarea`, `sl-select`, `sl-checkbox`, `sl-radio`, `sl-radio-group`, `sl-radio-button`, `sl-switch`, `sl-range`, `sl-color-picker`, `sl-rating`): `rw`
- Read-only Shoelace (`sl-progress-bar`, `sl-progress-ring`, `sl-qr-code`, `sl-option`, `sl-copy-button`): `r`
- Non-interactive elements (`div`, `span`, etc.): `r`

Custom wrappers may define additional properties, but only use them when the design explicitly specifies a wrapper that documents those properties.

## Widgets

```html
<!-- Text --> <span ui-value="name"></span> <div ui-value="compute()"></div>
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

**Shoelace Tips:**
- `sl-badge` has no `value` attribute — it displays its child element. Use `<sl-badge><span ui-value="count"></span></sl-badge>`

## Lists

**Standard pattern (using ui-view with wrapper):**

```html
<!-- In app viewdef -->
<div ui-view="items?wrapper=lua.ViewList"></div>
```

**IMPORTANT:** Always use `ui-view` with `wrapper=lua.ViewList` for lists, which wrap their items in with ui-view attributes.

**Selectable lists:** For lists where clicking an item selects it, bind `ui-event-mousedown` (not `ui-event-click`) unless the user states otherwise. This provides immediate visual feedback before the click completes.

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
<div class="tree-item" ui-event-mousedown="invoke()">
  <span ui-value="name"></span>
</div>
```

**WRONG** (would look for `item.name` on the TreeItem, which doesn't exist):

```html
<!-- TreeItem.list-item.html - BROKEN -->
<div class="tree-item" ui-event-mousedown="item.invoke()">
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
-- 1. Declare app prototype (serves as namespace)
MaLuba = session:prototype("MaLuba", {
    items = EMPTY,
    name = ""
})

-- 2. Nested prototypes use dotted names
MaLuba.Item = session:prototype("MaLuba.Item", { name = "" })
local Item = MaLuba.Item  -- local shortcut for cleaner code

function MaLuba:new(instance)
    instance = session:create(MaLuba, instance)
    instance.items = instance.items or {}
    return instance
end

function MaLuba:add()
    local item = session:create(Item, { name = self.name })
    table.insert(self.items, item)
    self.name = ""
end

-- 3. Guard instance creation (idempotent)
if not session.reloading then
    -- global name = camelCase of app directory (ma-luba → maLuba)
    maLuba = MaLuba:new()
end
```

## Styling

**Put ALL CSS in top-level object viewdefs only.** Never in index.html or list-item viewdefs.

```html
<template>
  <style>
    .ma-luba { padding: 1rem; }
    .hidden { display: none !important; }
  </style>
  <div class="ma-luba">...</div>
</template>
```

**Rules:**
- ALL styles go in the top-level viewdef (e.g., `MaLuba.DEFAULT.html`)
- **NEVER put `<style>` blocks in list-item viewdefs** — they get duplicated for each item
- Styles cascade down to nested viewdefs automatically
- Use Shoelace CSS variables (e.g., `var(--sl-spacing-medium)`) for consistency
- The `.hidden` utility class is commonly needed for `ui-class-hidden` bindings

## Viewport Fitting (Critical)

**Apps must fit within the viewport without causing page scroll.** Content that overflows should scroll within its container, not the page.

**Required CSS pattern for full-height apps:**

```html
<template>
  <style>
    html, body {
      margin: 0;
      padding: 0;
      overflow: hidden;  /* Prevent page scroll */
    }
    .my-app {
      display: flex;
      flex-direction: column;
      height: 100vh;
      overflow: hidden;  /* Contain children */
    }
    .scrollable-area {
      flex: 1;
      min-height: 0;     /* CRITICAL: allows flex child to shrink */
      overflow-y: auto;  /* Scroll within container */
    }
  </style>
</template>
```

**Key rules:**
1. Set `html, body { margin: 0; padding: 0; overflow: hidden; }` to prevent page scroll
2. Root container needs `height: 100vh` and `overflow: hidden`
3. Flex children that should scroll need BOTH `min-height: 0` AND `overflow-y: auto`
4. The `min-height: 0` is essential - without it, flex items won't shrink below content size

## Complete Example

See the `examples/` directory for a complete Contact Manager with Chat:
- `examples/requirements.md` — Requirements spec
- `examples/design.md` — Design spec
- `examples/app.lua` — Lua code (shows namespace pattern: `Contacts`, `Contacts.Contact`, `Contacts.ChatMessage`)
- `examples/viewdefs/Contacts.DEFAULT.html` — App viewdef
- `examples/viewdefs/Contacts.Contact.list-item.html` — Contact item viewdef
- `examples/viewdefs/Contacts.ChatMessage.list-item.html` — Chat message viewdef
