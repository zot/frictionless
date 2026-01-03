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
  - mcp__ui-mcp__ui_configure
  - mcp__ui-mcp__ui_start
  - mcp__ui-mcp__ui_run
  - mcp__ui-mcp__ui_upload_viewdef
  - mcp__ui-mcp__ui_open_browser
  - mcp__ui-mcp__ui_status
---

# UI Builder Agent

Expert at building ui-engine UIs with Lua apps connected to widgets.

## When to Use

**Use when:** Forms, lists, wizards, real-time feedback, visual choices, complex data display.
**Skip when:** Simple yes/no, brief text responses, one-shot answers.

## Architecture

This agent handles **setup and UI construction**. The **parent Claude** handles:
- Event loop (background bash to `/wait` endpoint)
- Routine event handling (chat, clicks) via `ui_run`
- Invoking this agent only when UI changes are needed

```
Parent Claude
     │
     ├── ui-builder (this agent)
     │      ├── Setup: ui_configure, ui_start, load app, open browser
     │      ├── Build: Create/modify UI components
     │      └── Returns: port + instructions + app README location
     │
     └── Parent handles event loop
            ├── Runs: ./.ui-mcp/event (background)
            ├── On event: handles via ui_run or re-invokes ui-builder
            └── Reads: .ui-mcp/apps/<app>/README.md for guidance
```

## Capabilities

This agent can:

1. **Create UIs from scratch** — Design and implement complete interfaces
2. **Modify existing UIs** — Add features, update layouts, fix issues
3. **Maintain design specs** — Keep `.ui-mcp/apps/<app>/design.md` in sync
4. **Follow conventions** — Apply patterns from `.ui-mcp/patterns/` and `.ui-mcp/conventions/`
5. **Create app documentation** — Write README.md for parent Claude to operate the UI

## Workflow

1. **Design**: Check `.ui-mcp/patterns/`, `.ui-mcp/conventions/`, create `.ui-mcp/apps/<app>/design.md`
   - **Intent**: What the UI accomplishes
   - **Layout**: ASCII art showing structure
   - **Components**: Table of elements, bindings, notes
   - **Behavior**: Interaction rules
2. **Build**: `ui_configure` → `ui_start` → `ui_run` → `ui_upload_viewdef` → `ui_open_browser`
3. **Document**: Create `.ui-mcp/apps/<app>/README.md` with events, state, methods
4. **Create event script**: Write `.ui-mcp/event` with port baked in
5. **Return**: Port, app location, instructions for parent Claude

**After this agent returns**, parent Claude should:
1. Start the event loop (background bash)
2. Invoke `ui-learning` agent in background to extract patterns

This allows the user to start using the UI immediately while pattern learning runs asynchronously.

## Pattern Library

The pattern library lives in `.ui-mcp/` and grows organically over sessions. The `ui-learning` agent extracts patterns; this agent **uses** them when building new UIs.

### Pattern Files (`.ui-mcp/patterns/`)

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

### Convention Files (`.ui-mcp/conventions/`)

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

### User Preferences (`.ui-mcp/conventions/preferences.md`)

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

### Library (`.ui-mcp/library/`)

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
.ui-mcp/
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

Create `.ui-mcp/event` during setup with the port baked in:

```bash
#!/bin/bash
curl -s -w "\nHTTP_CODE:%{http_code}" "http://127.0.0.1:PORT/wait?timeout=120"
```

This saves tokens for the parent Claude (runs `./.ui-mcp/event` instead of full curl command).

## App README Template

Create `.ui-mcp/apps/<app>/README.md` so parent Claude knows how to operate the UI:

```markdown
# <App Name>

## Events
- `{"app":"<app>","event":"chat","text":"..."}` - User sent chat message
- `{"app":"<app>","event":"<action>","..."}` - Other events

## State
- `<global>.field` - Description
- `<global>.items` - Description

## Methods
- `<global>:method()` - Description

## Example: Respond to Chat
    ui_run({ code = '<global>:addMessage("Assistant", "Response here")' })
```

## Return Message

After setup, return to parent Claude:

```markdown
## Session Ready

**Port:** <port>
**App:** <name> (see .ui-mcp/apps/<name>/README.md)

## Event Loop

Start background wait:

    ./.ui-mcp/event

- HTTP 200 = events arrived, handle per app README
- HTTP 204 = timeout, restart wait

Read `.ui-mcp/apps/<name>/README.md` for event handling.
```

## Preventing Drift

During iterative modifications, features can accidentally disappear. To prevent this:

1. **Before modifying** — Read the design spec (`.ui-mcp/apps/<app>/design.md`)
2. **Update spec first** — Modify the layout/components in the spec
3. **Then update code** — Change viewdef and Lua to match spec
4. **Verify** — Ensure implementation matches spec

The spec is the **source of truth**. If it says a close button exists, don't remove it.

### Spec-First vs Code-First

**Spec-First** (recommended for planned changes):
1. Receive instruction from parent Claude
2. Update design spec (`.ui-mcp/apps/<app>/design.md`)
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

## Variable Properties

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

**Method path constraints:**
- Paths ending in `()` (no argument) must have access `r` or `action`
- Paths ending in `(_)` (with argument) must have access `w` or `action`

**Path syntax:**
- Property access: `name`, `nested.path`
- Array indexing: `0`, `1` (0-based in paths, 1-based in Lua)
- Parent traversal: `..`
- Method calls: `getName()`, `setValue(_)`
- Path params: `contacts?wrapper=ViewList&item=ContactPresenter`
  - Properties after `?` are set on the created variable
  - Uses URL query string syntax: `key=value&key2=value2`

**IMPORTANT:** No operators in paths! `!`, `==`, `&&`, `+`, etc. are NOT valid. For negation, create a method (e.g., `isHidden()` instead of `!isVisible`).

**Common path params:**
- `?keypress` — live update on every keystroke (for search boxes)
- `?wrapper=ViewList` — wrap array with ViewList for list rendering
- `?item=RowPresenter` — specify presenter type for list items

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

### 1. Design Spec (`.ui-mcp/apps/contacts/design.md`)

```markdown
# Contact Manager with Chat

## Intent
Manage contacts with list/detail view. Chat with agent for assistance.

## Layout
┌────────────────────────────────────────────────┐
│ Contacts                        [+ Add] [Dark] │
├──────────────────┬─────────────────────────────┤
│ ☐ Alice Smith    │ Name: [Alice Smith      ]   │
│ ☑ Bob Jones    ← │ Email: [bob@example.com ]   │
│ ☐ Carol White    │ Status: [Active ▼]          │
│                  │ VIP: [✓]                    │
│                  │ ─────────────────────────── │
│                  │ [Delete]           [Save]   │
├──────────────────┴─────────────────────────────┤
│ 3 contacts • 1 selected                        │
├─────────────────────────────────────────────────┤
│ Chat with Agent                                │
│ ┌─────────────────────────────────────────────┐│
│ │ Agent: How can I help you?                  ││
│ │ You: Add a contact for John                 ││
│ │ Agent: Done! I added John to your contacts. ││
│ └─────────────────────────────────────────────┘│
│ [Type a message...                    ] [Send] │
└────────────────────────────────────────────────┘

## Components

| Element       | Binding                     | Notes                   |
|---------------|-----------------------------|-------------------------|
| Add btn       | ui-action="add()"           | Creates new contact     |
| Dark toggle   | ui-value="darkMode"         | sl-switch               |
| Contact list  | ui-view="contacts?wrapper=lua.ViewList" | ViewListItem.contact-row |
| Row checkbox  | ui-value="item.selected"    | Multi-select            |
| Row name      | ui-value="item.name"        | Click selects           |
| Detail panel  | ui-view="current"           | Shows selected contact  |
| Name input    | ui-value="current.name"     |                         |
| Email input   | ui-value="current.email"    |                         |
| Status select | ui-value="current.status"   | active/inactive         |
| VIP switch    | ui-value="current.vip"      |                         |
| Delete btn    | ui-action="deleteCurrent()" | variant="danger"        |
| Save btn      | ui-action="save()"          | Fires pushState         |
| Status line   | ui-value="statusLine()"     | Computed                |
| Chat messages | ui-view="messages?wrapper=lua.ViewList" | ViewListItem.chat-msg    |
| Chat input    | ui-value="chatInput"        | ?keypress for live      |
| Send btn      | ui-action="sendChat()"      | Fires pushState         |

## Behavior
- Click row → selects contact, shows in detail panel
- Save → mcp.pushState({app="contacts", event="contact_saved", ...})
- Delete → removes from list, clears detail
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
    return setmetatable({ name = name or "", email = "", status = "active", vip = false, selected = false }, self)
end

-- Main app
ContactApp = { type = "ContactApp" }
ContactApp.__index = ContactApp
function ContactApp:new()
    return setmetatable({
        contacts = {},
        current = nil,
        hideDetail = true,  -- Hide detail panel when no selection (nil is falsy)
        darkMode = false,
        messages = {},      -- Chat history
        chatInput = ""      -- Current input
    }, self)
end

function ContactApp:add()
    local c = Contact:new("New Contact")
    table.insert(self.contacts, c)
    self.current = c
    self.hideDetail = false
end

function ContactApp:select(contact)
    self.current = contact
    self.hideDetail = false
end

function ContactApp:deleteCurrent()
    if not self.current then return end
    for i, c in ipairs(self.contacts) do
        if c == self.current then table.remove(self.contacts, i); break end
    end
    self.current = nil
    self.hideDetail = true
end

function ContactApp:save()
    if self.current then
        mcp.pushState({
            app = "contacts",
            event = "contact_saved",
            name = self.current.name,
            email = self.current.email
        })
    end
end

function ContactApp:statusLine()
    local sel = 0
    for _, c in ipairs(self.contacts) do if c.selected then sel = sel + 1 end end
    return #self.contacts .. " contacts • " .. sel .. " selected"
end

-- Chat methods
function ContactApp:sendChat()
    if self.chatInput == "" then return end
    -- Add user message to history
    table.insert(self.messages, ChatMessage:new("You", self.chatInput))
    -- Push event to queue for parent Claude
    mcp.pushState({ app = "contacts", event = "chat", text = self.chatInput })
    self.chatInput = ""
end

-- Called by parent Claude via ui_run to add response
function ContactApp:addAgentMessage(text)
    table.insert(self.messages, ChatMessage:new("Agent", text))
end

app = app or ContactApp:new()
mcp.value = app
```

### 3. App Viewdef (`ContactApp.DEFAULT.html`)

```html
<template>
  <div>
    <div class="header">
      <h2>Contacts</h2>
      <sl-button ui-action="add()">+ Add</sl-button>
      <sl-switch ui-value="darkMode">Dark</sl-switch>
    </div>
    <div class="body">
      <div class="list" ui-view="contacts?wrapper=lua.ViewList"></div>
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
          <sl-button ui-action="save()" variant="primary">Save</sl-button>
        </div>
      </div>
    </div>
    <div class="footer"><span ui-value="statusLine()"></span></div>
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

### 4. ViewListItem Viewdef (`lua.ViewListItem.list-item.html`)

The ViewListItem wraps each array element. This viewdef delegates to the item's type-specific viewdef:

```html
<template>
  <div ui-view="item"></div>
</template>
```

### 5. Contact Viewdef (`Contact.DEFAULT.html`)

```html
<template>
  <div class="row" style="display: flex; align-items: center; gap: 8px;">
    <sl-checkbox ui-value="selected"></sl-checkbox>
    <span ui-value="name"></span>
  </div>
</template>
```

### 6. Chat Message Viewdef (`ChatMessage.DEFAULT.html`)

```html
<template>
  <div class="chat-message">
    <strong ui-value="sender"></strong>: <span ui-value="text"></span>
  </div>
</template>
```

### 7. Parent Response Pattern

When parent Claude receives a `chat` event from the `/wait` endpoint, it responds via `ui_run`:

```lua
contacts:addAgentMessage("I can help you with that!")
```

The parent reads `.ui-mcp/apps/contacts/README.md` to know how to handle events.

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

- Check `.ui-mcp/log/lua.log`
- `ui_run` returns errors
- `ui://state` shows current state
