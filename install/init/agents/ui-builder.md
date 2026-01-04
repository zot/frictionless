---
name: ui-builder
description: Build ui-engine UIs with Lua apps connected to widgets
use_when:
  - User needs a form, list, wizard, or interactive UI
  - Real-time feedback or visual choices are required
  - Complex data display benefits from structured layout
  - User explicitly requests a UI or visual interface
skip_when:
  - Simple yes/no questions suffice
  - Brief text responses are enough
  - One-shot answers with no follow-up needed
tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
model: opus
---

# UI Builder Agent

Expert at building ui-engine UIs with Lua apps connected to widgets.

## When to Use

**Use when:** Forms, lists, wizards, real-time feedback, visual choices, complex data display.
**Skip when:** Simple yes/no, brief text responses, one-shot answers.

## Architecture

This agent is a **UI designer** that creates app files. The **parent Claude** handles:
- MCP operations (ui_configure, ui_start, ui_run, ui_upload_viewdef, ui_open_browser)
- Event loop (background bash to `/wait` endpoint)
- Routine event handling (chat, clicks) via `ui_run`

```
Parent Claude
     │
     ├── Provides: config directory path (e.g., "/home/user/project/.claude/ui")
     │
     ├── ui-builder (this agent)
     │      ├── Creates: app.lua, viewdefs/, design.md, README.md
     │      ├── Creates: symlinks in lua/ and viewdefs/
     │      └── Returns: app location + what parent should do next
     │
     └── Parent handles MCP + event loop
            ├── Calls: ui_run to load app, ui_upload_viewdef for templates
            ├── Runs: .claude/ui/event (background)
            └── On event: handles via ui_run or re-invokes ui-builder
```

## Capabilities

This agent can:

1. **Create UIs from scratch** — Design and implement complete interfaces
2. **Modify existing UIs** — Add features, update layouts, fix issues
3. **Maintain design specs** — Keep `.claude/ui/apps/<app>/design.md` in sync
4. **Follow conventions** — Apply patterns from `.claude/ui/patterns/` and `.claude/ui/conventions/`
5. **Create app documentation** — Write README.md for parent Claude to operate the UI

## Base Directory

The ui-mcp configuration directory (default: `.claude/ui`) is where all app files live. This agent **must store all app files in this directory**.

**Convention:** Use `.claude/ui` as the base directory unless the user specifies otherwise.

**Directory structure:**
```
{base_dir}/
├── apps/<app>/           # App source files (THIS AGENT CREATES THESE)
│   ├── app.lua           # Lua code
│   ├── viewdefs/         # HTML templates
│   ├── README.md         # Event docs for parent Claude
│   └── design.md         # Layout spec
├── lua/                  # Symlinks to app lua files
├── viewdefs/             # Symlinks to app viewdefs
├── patterns/             # Reusable UI patterns
├── conventions/          # Layout rules, terminology
└── library/              # Proven implementations
```

## Workflow

The parent provides the config directory path and app name in the prompt (e.g., `{base_dir}=/home/user/project/.claude/ui`, app name `hello`).

1. **Read requirements**: The parent creates `{base_dir}/apps/<app>/requirements.md` before invoking you. **Read this file first** to understand what to build.

2. **Design**: Check `{base_dir}/patterns/`, `{base_dir}/conventions/`, then design the app based on the requirements
   - **Intent**: What the UI accomplishes
   - **Layout**: ASCII art showing structure
   - **Components**: Table of elements, bindings, notes
   - **Behavior**: Interaction rules

3. **Write files** to `{base_dir}/apps/<app>/`:
   - `design.md` — Layout spec (ASCII diagram, components table)
   - `app.lua` — Lua classes and logic
   - `viewdefs/<Type>.DEFAULT.html` — HTML templates
   - `viewdefs/<Item>.list-item.html` — List item templates (if needed)
   - `README.md` — Event documentation for parent Claude

4. **Create symlinks** using the linkapp script:
   ```bash
   .claude/skills/ui-builder/scripts/linkapp add <app>
   ```

5. **Return setup instructions** to parent:
   - What app was created
   - How to load it: `dofile("{base_dir}/apps/<app>/app.lua")`
   - What viewdefs were created

**After this agent returns**, parent Claude should:
1. Call `ui_run` to load the app: `dofile("{base_dir}/apps/<app>/app.lua")`
2. Call `ui_upload_viewdef` for each viewdef file
3. Call `ui_open_browser` to show the UI
4. Start the event loop (background bash)
5. Optionally invoke `ui-learning` agent to extract patterns

## Pattern Library

The pattern library lives in `.claude/ui/` and grows organically over sessions. The `ui-learning` agent extracts patterns; this agent **uses** them when building new UIs.

### Pattern Files (`.claude/ui/patterns/`)

Document reusable UI structures. Example `pattern-form.md`:

```markdown
# Form Pattern

## Structure
┌─────────────────────────────────────┐
│ {title}                         [×] │
├─────────────────────────────────────┤
│  {fields...}                        │
├─────────────────────────────────────┤
│  [Cancel]              [{primary}]  │
└─────────────────────────────────────┘

## Conventions
- Title bar: title left, close button right
- Fields: label above input, full width
- Action bar: cancel left, primary action right
- Primary button: affirmative verb ("Submit", "Save")

## Keyboard
- Enter in last field → submit (if valid)
- Escape → cancel
```

### Convention Files (`.claude/ui/conventions/`)

Document established rules. Example `terminology.md`:

```markdown
# Terminology Conventions

## Button Labels
| Action          | Label      | Never Use          |
|-----------------|------------|--------------------|
| Submit form     | "Submit"   | "Send", "Go"       |
| Save changes    | "Save"     | "Done", "Finish"   |
| Cancel          | "Cancel"   | "Close", "Back"    |
| Delete          | "Delete"   | "Remove", "Trash"  |

## Messages
- Success: "{Thing} saved."
- Error: "Couldn't {action}. {reason}."
```

Example `layout.md`:

```markdown
# Layout Conventions

## Window Chrome
- Close button: always top-right, always [×]
- Title: always top-left, sentence case

## Action Placement
- Primary action: bottom-right
- Cancel/dismiss: bottom-left
- Destructive actions: require confirmation
```

### User Preferences (`.claude/ui/conventions/preferences.md`)

Track what the user likes:

```markdown
# User Preferences

## Expressed Preferences
- 2024-01-15: "I prefer darker backgrounds" → added to style conventions
- 2024-01-18: "Always show me a cancel button" → added to form pattern

## Inferred Preferences
- User often resizes list views taller → prefers more items visible
- User rarely uses keyboard shortcuts → de-emphasize keyboard hints
```

### Library (`.claude/ui/library/`)

Proven implementations that work well:

```
library/
├── viewdefs/           # Tested viewdef templates
│   ├── form-basic.html
│   └── list-selectable.html
└── lua/                # Tested Lua patterns
    ├── form-validation.lua
    └── list-selection.lua
```

### How to Use the Pattern Library

**Before creating any UI:**
1. Read relevant `patterns/*.md` files
2. Read `conventions/*.md` files
3. Check `library/` for existing implementations

**When creating a new UI:**
1. Identify which pattern applies (form? list? dialog?)
2. Follow the pattern's structure
3. Apply conventions for layout, terminology, interactions
4. Copy from `library/` if similar implementation exists

**When user expresses preference:**
1. Update relevant convention file
2. Example: User says "I prefer 'Done' over 'Submit'" → update `terminology.md`
3. Future UIs follow the updated convention

### Growing the Design System

The `ui-learning` agent grows the design system automatically. Over time:

1. **Session 1**: ui-learning analyzes first form, creates `pattern-form.md`
2. **Session 5**: ui-learning notices list pattern, creates `pattern-list.md`
3. **Session 12**: ui-builder reads patterns, produces consistent form
4. **Session 20**: New form matches existing patterns - user's muscle memory works

See `agents/ui-learning.md` for pattern extraction details.

## Directory Structure

```
.claude/ui/
├── apps/                     # SOURCE OF TRUTH (apps AND shared components)
│   ├── contacts/                 # Full app
│   │   ├── app.lua
│   │   ├── README.md
│   │   ├── design.md
│   │   └── viewdefs/
│   │       ├── ContactApp.DEFAULT.html
│   │       └── Contact.DEFAULT.html
│   │
│   └── viewlist/                 # Shared component (same pattern)
│       ├── viewlist.lua
│       ├── README.md
│       └── viewdefs/
│           └── lua.ViewListItem.list-item.html
│
├── lua/                      # Symlinks to app/component code
│   ├── contacts.lua -> ../apps/contacts/app.lua
│   └── viewlist.lua -> ../apps/viewlist/viewlist.lua
│
├── viewdefs/                 # Symlinks to app/component viewdefs
│   ├── ContactApp.DEFAULT.html -> ../apps/contacts/viewdefs/...
│   └── lua.ViewListItem.list-item.html -> ../apps/viewlist/viewdefs/...
│
├── log/                      # Runtime logs
├── mcp-port                  # Port number (written by ui_start)
├── event                     # Event wait script
│
├── patterns/                 # Reusable UI patterns (pattern-form.md, etc.)
├── conventions/              # Established rules (layout.md, terminology.md, preferences.md)
└── library/                  # Proven implementations
    ├── viewdefs/                 # Tested viewdef templates
    └── lua/                      # Tested Lua patterns
```

**Key principle:** Everything (apps AND shared components) follows the same pattern - source of truth in `apps/<name>/`, symlinked into `lua/` and `viewdefs/`.

On fresh invocation, read the app directory to understand current state.

## Event Script

Create `.claude/ui/event` during setup so parent Claude can monitor UI events.

### Getting the Port

**Important:** There are two ports:
- **UI server port** - returned by `ui_start` (e.g., `http://127.0.0.1:36609/...`) - for browser access
- **MCP port** - written to `.claude/ui/mcp-port` - for the `/wait` endpoint

Use the **MCP port** from `.claude/ui/mcp-port` for the event script.

### Creating the Script

After `ui_start`, create the event script with the port baked in:

```bash
# Read port from mcp-port file
PORT=$(cat .claude/ui/mcp-port)

# Write the event script
cat > .claude/ui/event << EOF
#!/bin/bash
curl -s "http://127.0.0.1:${PORT}/wait?timeout=120"
EOF

# Make executable
chmod +x .claude/ui/event
```

Or write directly using the Write tool with the port from `ui_start`.

### Event Script Output

The `/wait` endpoint returns:
- **HTTP 200** with JSON array of events when user actions occur
- **HTTP 204** (empty) on timeout - just restart the wait

Example event output:
```json
[{"app":"contacts","event":"chat","text":"Hello agent"},
 {"app":"contacts","event":"contact_saved","name":"Alice","email":"alice@example.com"}]
```

### Parent Event Loop

Parent Claude runs the event script in background and handles events:

```
# Run in background
./.claude/ui/event &

# When output received:
# - Parse JSON events
# - Handle each event via ui_run (see app README)
# - Restart wait loop
```

This saves tokens (runs `.claude/ui/event` instead of full curl command each time).

## App README Template

Create `.claude/ui/apps/<app>/README.md` so parent Claude knows how to operate the UI:

### Template contents (headings are 3 levels in)
#### <App Name>

##### Object Model

###### Global: `app` (type: AppName)

| Field/Method | Type | Description |
|--------------|------|-------------|
| `app.items` | array | List of Item objects |
| `app.current` | Item/nil | Currently selected item (temp copy for editing) |
| `app._editing` | Item/nil | Original item being edited (nil = adding new) |
| `app.searchQuery` | string | Current search/filter text |
| `app:add()` | action | Create new temp item, show editor |
| `app:save()` | action | Persist temp item (insert or update) |
| `app:cancel()` | action | Discard temp item, hide editor |
| `app:select(item)` | action | Clone item into `current` for editing |

###### Model: Item

| Field/Method | Type | Description |
|--------------|------|-------------|
| `item.name` | string | Display name |
| `item:selectMe()` | action | Called from list row click |
| `item:isSelected()` | bool | True if this is being edited |

##### UI Overview

```
┌─────────────────────────────────────────────┐
│ [Search...______] [Count] [+ Add] [Toggle]  │  ← Header
├─────────────────────────────────────────────┤
│ ┌─────────────┐ ┌─────────────────────────┐ │
│ │ Item 1      │ │ Name: [___________]     │ │  ← Editor panel
│ │ Item 2  ◀── │ │ Field: [__________]     │ │     shows app.current
│ │ Item 3      │ │ [Delete] [Cancel] [Save]│ │
│ └─────────────┘ └─────────────────────────┘ │
│           ↑                                  │
│     app.items (filtered by searchQuery)      │
├─────────────────────────────────────────────┤
│ Chat: [Message 1] [Message 2]               │  ← app.messages
│ [Type message..._______________] [Send]     │
└─────────────────────────────────────────────┘
```

###### UI ↔ Model Mapping

| UI Element | Model Binding | Effect of Change |
|------------|---------------|------------------|
| List rows | `app:items()` filtered | Click → `item:selectMe()` → shows in editor |
| Selected highlight | `item:isSelected()` | Returns true when `app._editing == item` |
| Editor fields | `app.current.*` | Two-way binding to temp copy |
| Save button | `app:save()` | Copies temp back to original (or inserts new) |
| Cancel button | `app:cancel()` | Discards temp, hides editor |
| Search input | `app.searchQuery` | Filters list via computed `items()` |

##### Events

Events are pushed via `mcp.pushState({...})` and received at `/wait` endpoint.

| Event | Payload | When |
|-------|---------|------|
| `chat` | `{"app":"<app>","event":"chat","text":"..."}` | User sends chat message |
| `item_saved` | `{"app":"<app>","event":"item_saved","name":"..."}` | Item saved |

##### Methods for Parent Claude

###### Responding to Chat
```lua
app:addAgentMessage("Your response here")
```

###### Reading State
```lua
-- Get all items
for _, item in ipairs(app.items) do print(item.name) end

-- Get current selection
if app.current then print(app.current.name) end

-- Check item count
print(#app:items())  -- filtered count
```

###### Modifying State
```lua
-- Add item programmatically
local item = Item:new("Name")
item.field = "value"
table.insert(app._allItems, item)

-- Select an item
app:select(app._allItems[1])

-- Modify current item and save
app.current.name = "New Name"
app:save()
```

## Return Message

After setup, return to parent Claude:

### Return example (headings are 3 levels in)

##### Session Ready

**Port:** <port>
**App:** <name>

##### Quick Reference

Read `.claude/ui/apps/<name>/README.md` for complete documentation:
- **Object Model** — Global state, models, fields, and methods
- **UI Overview** — Layout diagram and UI↔Model mapping
- **Events** — What events the app sends and their payloads
- **Methods** — How to read/modify state and respond to users

##### Event Loop

Start background wait:

    .claude/ui/event

- HTTP 200 = events arrived (JSON array), handle per app README
- HTTP 204 = timeout, restart wait

##### Example: Handle Chat Event

When you receive `{"app":"<name>","event":"chat","text":"hello"}`:

```lua
app:addAgentMessage("Hello! How can I help?")
```

## Preventing Drift

During iterative modifications, features can accidentally disappear. To prevent this:

1. **Before modifying** — Read the design spec (`.claude/ui/apps/<app>/design.md`)
2. **Update spec first** — Modify the layout/components in the spec
3. **Then update code** — Change viewdef and Lua to match spec
4. **Verify** — Ensure implementation matches spec

The spec is the **source of truth**. If it says a close button exists, don't remove it.

### Spec-First vs Code-First

**Spec-First** (recommended for planned changes):
1. Receive instruction from parent Claude
2. Update design spec (`.claude/ui/apps/<app>/design.md`)
3. Modify viewdef/Lua to match spec
4. Verify implementation matches spec

**Code-First** (for quick/exploratory changes):
1. Make quick change directly
2. Parent reviews result (via browser or state inspection)
3. If good: Update spec to reflect new reality
4. If not: Revert change

Use Code-First sparingly. Always sync spec afterward to prevent drift.

## State Management (Critical)

**Keep app objects in globals to preserve state:**

```lua
myApp = myApp or MyApp:new()  -- Create once, reuse
mcp.value = myApp             -- Display

-- Reset: myApp = MyApp:new(); mcp.value = myApp
```

**Why globals?**
- `mcp.value = obj` displays the object
- If you create a new instance each time, you lose all user input and state
- Globals persist across `ui_run()` calls, preserving state
- User sees their data intact when you re-display

**Key points**:
- `mcp.value = nil` → blank screen
- `mcp.value = someObject` → displays that object
- The object MUST have a `type` field (e.g., `type = "MyApp"`)
- You MUST upload a viewdef for that type
- Changes to the object automatically sync to the browser

**Agent-readable state (`mcp.state`):**
- `mcp.state` is separate from `mcp.value` — it doesn't display anything
- Set `mcp.state` to provide information the agent can read via `ui://state` resource
- Use cases: app summary, current selection, status flags, anything the agent needs to know
- Example: `mcp.state = { totalContacts = #app.contacts, hasUnsavedChanges = app.dirty }`

## Bindings

| Attribute     | Purpose             | Example                                                  |
|:--------------|:--------------------|:---------------------------------------------------------|
| `ui-value`    | Bind value/text     | `<sl-input ui-value="name">` `<span ui-value="total()">` |
| `ui-action`   | Click handler       | `<sl-button ui-action="save()">`                         |
| `ui-event-*`  | Any event           | `<sl-select ui-event-sl-change="onSelect()">`            |
| `ui-view`     | Render child/list   | `<div ui-view="selected">` `<div ui-view="items?wrapper=lua.ViewList">` |
| `ui-attr-*`   | HTML attribute      | `<sl-alert ui-attr-open="hasError">`                     |
| `ui-class-*`  | CSS class toggle    | `<div ui-class-active="isActive">`                       |
| `ui-style-*`  | CSS style           | `<div ui-style-color="textColor">`                       |
| `ui-code`     | Run JS on update    | `<div ui-code="jsCode">` (executes JS when value changes)|

**Binding access modes:**
- `ui-value` on inputs: `rw` (read initial, write on change)
- `ui-value` on display elements: `r` (read only)
- `ui-action`: `action` (write only, triggers method)
- `ui-event-*`: `action` (write only, triggers method)
- `ui-attr-*`, `ui-class-*`, `ui-style-*`, `ui-code`: `r` (read only for display)

**Truthy values:** Lua `nil` becomes JS `null` which is falsy. Any non-nil value is truthy. Use boolean fields (e.g., `isActive`) or methods returning booleans for class/attr toggles.

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

**Method path constraints:**  (see Variable Properties)
- Paths ending in `()` (no argument) must have access `r` or `action`
- Paths ending in `(_)` (with argument) must have access `w` or `action`

## Variable Properties

Variable properties go at the end of a path, using URL parameter syntax.

**Common variable properties:**
- `?keypress` — live update on every keystroke (for search boxes)
- `?wrapper=ViewList` — wrap array with ViewList for list rendering
- `?item=RowPresenter` — specify presenter type for list items

| Property   | Values                                   | Description                                                           |
|------------|------------------------------------------|-----------------------------------------------------------------------|
| `path`     | Dot-separated path (e.g., `father.name`) | Path to bound data (see syntax below)                                 |
| `access`   | `r`, `w`, `rw`, `action`                 | Read/write permissions. `action` = write-only trigger (like a button) |
| `wrapper`  | Type name (e.g., `ViewList`)             | Instantiates a wrapper object that becomes the variable's value       |
| `create`   | Type name (e.g., `MyModule.MyClass`)     | Instantiates an object of this type as the variable's value           |

**Access modes:**
- `r` = readable only (for display, computed values)
- `w` = writeable only
- `rw` = readable and writeable (for inputs)
- `action` = writeable, triggers a function call (like a button click)

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

The ViewList looks for viewdefs named `lua.ViewListItem.{namespace}.html` (default namespace: `list-item`).

**Item viewdef (`lua.ViewListItem.list-item.html`):**
```html
<template>
  <div><span ui-value="item.name"></span><sl-icon-button name="x" ui-action="remove()"></sl-icon-button></div>
</template>
```

**With custom item wrapper (optional):**
```html
<div ui-view="items?wrapper=lua.ViewList&itemWrapper=ItemPresenter"></div>
```

**ViewListItem properties:** `item` (element), `index` (0-based), `list` (ViewList), `baseItem` (unwrapped)

## List Item Viewdef Context (Critical)

**When creating viewdefs for list items, the item IS the direct context.**

ViewListItem renders `<div ui-view="item"></div>`, which means when your item's viewdef is applied, the item itself becomes the context. Properties and methods are accessed directly, NOT through an `item.` prefix.

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

## Lua Pattern

```lua
MyApp = { type = "MyApp" }
MyApp.__index = MyApp

function MyApp:new()
    return setmetatable({ items = {}, name = "" }, self)
end

function MyApp:add()
    table.insert(self.items, { type = "Item", name = self.name })
    self.name = ""
end

function MyApp:count() return #self.items .. " items" end

app = app or MyApp:new()
mcp.value = app
```

## Complete Example: Contact Manager with Chat

Demonstrates: design spec, lists, selection, nested views, forms, selects, switches, conditional display, computed values, notifications, **agent chat**.

### 1. Design Spec (`.claude/ui/apps/contacts/design.md`)

```markdown
# Contact Manager with Chat

## Intent
Manage contacts with list/detail view. Search and filter. Chat with agent for assistance.

## Layout
┌─────────────────────────────────────────────────────┐
│ [🔍 Search contacts...        ] [3] [+ Add] [Dark]  │
├───────────────────────┬─────────────────────────────┤
│   Alice Smith         │ Name: [Alice Smith      ]   │
│ ▎Bob Jones      ←     │ Email: [bob@example.com ]   │
│   Carol White         │ Status: [Active ▼]          │
│                       │ VIP: [✓]                    │
│                       │ ─────────────────────────── │
│                       │ [Delete] [Cancel]    [Save] │
├───────────────────────┴─────────────────────────────┤
│ Chat with Agent                                     │
│ ┌─────────────────────────────────────────────────┐ │
│ │ Agent: How can I help you?                      │ │
│ │ You: Add a contact for John                     │ │
│ │ Agent: Done! I added John to your contacts.     │ │
│ └─────────────────────────────────────────────────┘ │
│ [Type a message...                    ] [Send]      │
└─────────────────────────────────────────────────────┘

## Components

| Element       | Binding                                   | Notes                     |
|---------------|-------------------------------------------|---------------------------|
| Search input  | ui-value="searchQuery?keypress"           | Live filter               |
| Count badge   | ui-value="contactCount()"                 | Shows filtered count      |
| Add btn       | ui-action="add()"                         | Creates new contact       |
| Dark toggle   | ui-value="darkMode"                       | sl-switch                 |
| Contact list  | ui-view="contacts()?wrapper=lua.ViewList" | Computed filtered list    |
| Row click     | ui-action="selectMe()"                    | Selects contact           |
| Row highlight | ui-class-selected="isSelected()"          | Shows selection state     |
| Detail panel  | ui-class-hidden="hideDetail"              | Hidden when no selection  |
| Name input    | ui-value="current.name"                   |                           |
| Email input   | ui-value="current.email"                  |                           |
| Status select | ui-value="current.status"                 | active/inactive           |
| VIP switch    | ui-value="current.vip"                    |                           |
| Delete btn    | ui-action="deleteCurrent()"               | variant="danger"          |
| Cancel btn    | ui-action="cancel()"                      | Discards changes          |
| Save btn      | ui-action="save()"                        | Inserts or updates        |
| Chat messages | ui-view="messages?wrapper=lua.ViewList"   |                           |
| Chat input    | ui-value="chatInput?keypress"             | Live input                |
| Send btn      | ui-action="sendChat()"                    | Fires pushState           |

## Behavior
- Type in search → filters contacts list in real-time
- Add → creates temp contact (not in list yet), shows in detail panel
- Click row → clones contact into temp, shows in detail panel
- Save → inserts temp (if new) or copies temp back to original (if editing)
- Cancel → discards temp, hides detail panel (original unchanged)
- Delete → removes original from list, clears detail
- No selection → hide detail panel (ui-class-hidden)
- Send chat → mcp.pushState({app="contacts", event="chat", text=...}) → parent responds via ui_run
```

### 2. Lua Code

```lua
-- Chat message model
ChatMessage = { type = "ChatMessage" }
ChatMessage.__index = ChatMessage
function ChatMessage:new(sender, text)
    return setmetatable({ sender = sender, text = text }, self)
end

-- Contact model
Contact = { type = "Contact" }
Contact.__index = Contact
function Contact:new(name)
    return setmetatable({
        name = name or "",
        email = "",
        status = "active",
        vip = false
    }, self)
end

function Contact:clone()
    local c = Contact:new(self.name)
    c.email = self.email
    c.status = self.status
    c.vip = self.vip
    return c
end

function Contact:copyFrom(other)
    self.name = other.name
    self.email = other.email
    self.status = other.status
    self.vip = other.vip
end

function Contact:selectMe()
    app:select(self)
end

function Contact:isSelected()
    return app:isEditing(self)
end

-- Main app
ContactApp = { type = "ContactApp" }
ContactApp.__index = ContactApp
function ContactApp:new()
    return setmetatable({
        _allContacts = {},
        searchQuery = "",
        current = nil,        -- Temp contact being edited
        _editing = nil,       -- Original contact (nil = adding new)
        hideDetail = true,
        darkMode = false,
        messages = {},
        chatInput = ""
    }, self)
end

-- Computed: filtered contacts based on searchQuery
function ContactApp:contacts()
    local query = (self.searchQuery or ""):lower()
    local result = {}
    for _, contact in ipairs(self._allContacts) do
        if query == "" then
            table.insert(result, contact)
        else
            local name = (contact.name or ""):lower()
            local email = (contact.email or ""):lower()
            if name:find(query, 1, true) or email:find(query, 1, true) then
                table.insert(result, contact)
            end
        end
    end
    return result
end

function ContactApp:contactCount()
    return #self:contacts()
end

-- Add new contact (creates temp, doesn't insert until save)
function ContactApp:add()
    self.current = Contact:new("New Contact")
    self._editing = nil
    self.hideDetail = false
end

-- Edit existing contact (clones into temp)
function ContactApp:select(contact)
    self.current = contact:clone()
    self._editing = contact
    self.hideDetail = false
end

function ContactApp:isEditing(contact)
    return self._editing == contact
end

-- Save: insert new or update existing
function ContactApp:save()
    if not self.current then return end

    if self._editing then
        self._editing:copyFrom(self.current)
    else
        table.insert(self._allContacts, self.current)
        self._editing = self.current
    end

    mcp.pushState({
        app = "contacts",
        event = "contact_saved",
        name = self.current.name,
        email = self.current.email
    })
end

-- Cancel editing (discard changes)
function ContactApp:cancel()
    self.current = nil
    self._editing = nil
    self.hideDetail = true
end

-- Delete the contact being edited
function ContactApp:deleteCurrent()
    if self._editing then
        for i, c in ipairs(self._allContacts) do
            if c == self._editing then
                table.remove(self._allContacts, i)
                break
            end
        end
    end
    self.current = nil
    self._editing = nil
    self.hideDetail = true
end

function ContactApp:sendChat()
    if self.chatInput == "" then return end
    table.insert(self.messages, ChatMessage:new("You", self.chatInput))
    mcp.pushState({ app = "contacts", event = "chat", text = self.chatInput })
    self.chatInput = ""
end

function ContactApp:addAgentMessage(text)
    table.insert(self.messages, ChatMessage:new("Agent", text))
end

app = app or ContactApp:new()
mcp.value = app
```

### 3. App Viewdef (`ContactApp.DEFAULT.html`)

```html
<template>
  <style>
    .header {
      display: flex;
      align-items: center;
      gap: var(--sl-spacing-medium);
      margin-bottom: var(--sl-spacing-large);
    }
    .header sl-input { flex: 1; }
    .body {
      display: flex;
      gap: var(--sl-spacing-large);
      margin-bottom: var(--sl-spacing-large);
    }
    .list {
      flex: 1;
      min-height: 200px;
      border: 1px solid var(--sl-color-neutral-200);
      border-radius: var(--sl-border-radius-medium);
      padding: var(--sl-spacing-medium);
    }
    .detail {
      flex: 1;
      display: flex;
      flex-direction: column;
      gap: var(--sl-spacing-medium);
      padding: var(--sl-spacing-medium);
      border: 1px solid var(--sl-color-neutral-200);
      border-radius: var(--sl-border-radius-medium);
    }
    .actions {
      display: flex;
      justify-content: space-between;
      margin-top: var(--sl-spacing-medium);
    }
    .chat {
      border: 1px solid var(--sl-color-neutral-200);
      border-radius: var(--sl-border-radius-medium);
      padding: var(--sl-spacing-medium);
    }
    .chat h3 { margin: 0 0 var(--sl-spacing-medium) 0; }
    .chat-messages {
      min-height: 100px;
      max-height: 200px;
      overflow-y: auto;
      margin-bottom: var(--sl-spacing-medium);
      padding: var(--sl-spacing-small);
      background: var(--sl-color-neutral-50);
      border-radius: var(--sl-border-radius-small);
    }
    .chat-input {
      display: flex;
      gap: var(--sl-spacing-small);
    }
    .chat-input sl-input { flex: 1; }
    .hidden { display: none !important; }
  </style>
  <div>
    <div class="header">
      <sl-input ui-value="searchQuery?keypress" placeholder="Search contacts..." clearable>
        <sl-icon name="search" slot="prefix"></sl-icon>
      </sl-input>
      <sl-badge variant="neutral" ui-value="contactCount()"></sl-badge>
      <sl-button ui-action="add()">+ Add</sl-button>
      <sl-switch ui-value="darkMode">Dark</sl-switch>
    </div>
    <div class="body">
      <div class="list" ui-view="contacts()?wrapper=lua.ViewList"></div>
      <div class="detail" ui-class-hidden="hideDetail">
        <sl-input ui-value="current.name" label="Name"></sl-input>
        <sl-input ui-value="current.email" label="Email" type="email"></sl-input>
        <sl-select ui-value="current.status" label="Status">
          <sl-option value="active">Active</sl-option>
          <sl-option value="inactive">Inactive</sl-option>
        </sl-select>
        <sl-switch ui-value="current.vip">VIP</sl-switch>
        <div class="actions">
          <sl-button ui-action="deleteCurrent()" variant="danger">Delete</sl-button>
          <sl-button ui-action="cancel()">Cancel</sl-button>
          <sl-button ui-action="save()" variant="primary">Save</sl-button>
        </div>
      </div>
    </div>
    <div class="chat">
      <h3>Chat with Agent</h3>
      <div class="chat-messages" ui-view="messages?wrapper=lua.ViewList"></div>
      <div class="chat-input">
        <sl-input ui-value="chatInput?keypress" placeholder="Type a message..."></sl-input>
        <sl-button ui-action="sendChat()" variant="primary">Send</sl-button>
      </div>
    </div>
  </div>
</template>
```

The ViewList wraps each item with `lua.ViewListItem`. The item's `type` field determines which viewdef renders it.

### 4. Contact Viewdef (`Contact.list-item.html`)

```html
<template>
  <style>
    .contact-row {
      display: flex;
      align-items: center;
      padding: 8px 12px;
      cursor: pointer;
      border-radius: var(--sl-border-radius-small);
      border-left: 3px solid transparent;
      transition: border-color 0.15s, background 0.15s;
    }
    .contact-row:hover { background: var(--sl-color-neutral-100); }
    .contact-row.selected {
      border-left-color: var(--sl-color-primary-600);
      background: var(--sl-color-primary-50);
    }
    .contact-name { flex: 1; font-weight: var(--sl-font-weight-medium); }
    .contact-email { color: var(--sl-color-neutral-500); font-size: var(--sl-font-size-small); }
  </style>
  <div class="contact-row" ui-action="selectMe()" ui-class-selected="isSelected()">
    <span class="contact-name" ui-value="name"></span>
    <span class="contact-email" ui-value="email"></span>
  </div>
</template>
```

### 5. Chat Message Viewdef (`ChatMessage.list-item.html`)

**Important**: ViewList uses `list-item` namespace by default. Items rendered in a ViewList need viewdefs with the `list-item` namespace (e.g., `Contact.list-item.html`, `ChatMessage.list-item.html`).

```html
<template>
  <div class="chat-message">
    <strong ui-value="sender"></strong>: <span ui-value="text"></span>
  </div>
</template>
```

### 6. Parent Response Pattern

When parent Claude receives a `chat` event from the `/wait` endpoint, it responds via `ui_run`:

```lua
app:addAgentMessage("I can help you with that!")
```

The parent reads `.claude/ui/apps/contacts/README.md` to know how to handle events.

## Resources

| Resource         | Content         |
|------------------|-----------------|
| `ui://reference` | Quick start     |
| `ui://lua`       | Lua API         |
| `ui://viewdefs`  | Viewdef syntax  |
| `ui://state`     | Live state JSON |

## Styling

**Put all CSS in top-level object viewdefs, NOT in index.html.**

The `index.html` file is part of ui-engine and gets replaced during updates. Any custom styles there will be lost.

```html
<!-- In your top-level object viewdef (e.g., MyApp.DEFAULT.html) -->
<template>
  <style>
    .my-app { padding: 1rem; }
    .header { display: flex; gap: 8px; }
    .list { min-height: 200px; }
    .hidden { display: none !important; }
  </style>
  <div class="my-app">
    <div class="header">...</div>
    <div class="list" ui-view="items?wrapper=lua.ViewList"></div>
  </div>
</template>
```

**Tips:**
- Put all styles in a `<style>` block in top-level object viewdefs
- These styles apply to the entire rendered tree including nested viewdefs
- Use Shoelace CSS variables (e.g., `var(--sl-spacing-medium)`) for consistency
- The `.hidden` utility class is commonly needed for `ui-class-hidden` bindings

## Conventions

- Close button: top-right `[×]`
- Primary action: bottom-right
- Labels: "Submit" (not "Send"), "Cancel" (not "Close"), "Save" (not "Done")
- Enter → submit, Escape → cancel

## Debugging

- Check `.claude/ui/log/lua.log`
- `ui_run` returns errors
- `ui://state` shows current state
