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

## Capabilities

This agent can:

1. **Create UIs from scratch** — Design and implement complete interfaces
2. **Modify existing UIs** — Add features, update layouts, fix issues
3. **Maintain design specs** — Keep `.ui-mcp/design/ui-*.md` in sync
4. **Follow conventions** — Apply patterns from `.ui-mcp/patterns/` and `.ui-mcp/conventions/`
5. **Handle notifications** — Process user interactions via `mcp.notify()`

## Workflow

1. **Design**: Check `.ui-mcp/patterns/`, `.ui-mcp/conventions/`, create `.ui-mcp/design/ui-{name}.md`
   - **Intent**: What the UI accomplishes
   - **Layout**: ASCII art showing structure
   - **Components**: Table of elements, bindings, notes
   - **Behavior**: Interaction rules
2. **Build**: `ui_configure` → `ui_start` → `ui_run` → `ui_upload_viewdef` → `ui_open_browser`
3. **Operate**: User interacts → Lua calls `mcp.notify()` → Agent processes

## Directory Structure

```
.ui-mcp/
├── lua/            # Lua source files
├── viewdefs/       # HTML templates
├── log/            # Runtime logs (lua.log for debugging)
│
├── design/         # UI layout specs (SOURCE OF TRUTH)
│   └── ui-*.md         # Per-UI ASCII layouts
│
├── patterns/       # Reusable UI patterns
│   ├── pattern-form.md
│   └── pattern-list.md
│
├── conventions/    # Established conventions
│   ├── layout.md       # Spatial rules
│   ├── terminology.md  # Standard labels
│   └── preferences.md  # User preferences
│
└── library/        # Proven implementations
    ├── viewdefs/
    └── lua/
```

## Preventing Drift

During iterative modifications, features can accidentally disappear. To prevent this:

1. **Before modifying** — Read the design spec (`.ui-mcp/design/ui-*.md`)
2. **Update spec first** — Modify the layout/components in the spec
3. **Then update code** — Change viewdef and Lua to match spec
4. **Verify** — Ensure implementation matches spec

The spec is the **source of truth**. If it says a close button exists, don't remove it.

## State Management (Critical)

**Keep app objects in globals to preserve state:**

```lua
myApp = myApp or MyApp:new()  -- Create once, reuse
mcp.state = myApp             -- Display

-- Reset: myApp = MyApp:new(); mcp.state = myApp
```

**Why globals?**
- `mcp.state = obj` displays the object
- If you create a new instance each time, you lose all user input and state
- Globals persist across `ui_run()` calls, preserving state
- User sees their data intact when you re-display

**Key points**:
- `mcp.state = nil` → blank screen
- `mcp.state = someObject` → displays that object
- The object MUST have a `type` field (e.g., `type = "MyApp"`)
- You MUST upload a viewdef for that type
- Changes to the object automatically sync to the browser

## Bindings

| Attribute     | Purpose             | Example                                                  |
|:--------------|:--------------------|:---------------------------------------------------------|
| `ui-value`    | Bind value/text     | `<sl-input ui-value="name">` `<span ui-value="total()">` |
| `ui-action`   | Click handler       | `<sl-button ui-action="save()">`                         |
| `ui-event-*`  | Any event           | `<sl-select ui-event-sl-change="onSelect()">`            |
| `ui-view`     | Render child object | `<div ui-view="selected">`                               |
| `ui-viewlist` | Render array        | `<div ui-viewlist="items">`                              |
| `ui-attr-*`   | HTML attribute      | `<sl-alert ui-attr-open="hasError">`                     |
| `ui-class-*`  | CSS class toggle    | `<div ui-class-active="isOn">`                           |
| `ui-style-*`  | CSS style           | `<div ui-style-color="color">`                           |

**Paths:** `property`, `nested.path`, `method()`, `method(_)`, `items[0]` (0-indexed in HTML, 1-indexed in Lua)
- `method(_)` uses the update value as the argument
**Params:** `ui-value="query?keypress"` (live), `ui-view="x?wrapper=Presenter"`

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
<!-- Hide --> <div ui-class-hidden="!show">Content</div>
<!-- Alert --> <sl-alert ui-attr-open="err" variant="danger"><span ui-value="msg"></span></sl-alert>
<!-- Child --> <div ui-view="selectedItem"></div>
```

## Lists

**App viewdef:**
```html
<div ui-viewlist="items" ui-namespace="item-row"></div>
```

**Item viewdef (`lua.ViewListItem.item-row.html`):**
```html
<template>
  <div><span ui-value="item.name"></span><sl-icon-button name="x" ui-action="remove()"></sl-icon-button></div>
</template>
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
mcp.state = app
```

## Complete Example: Contact Manager with Chat

Demonstrates: design spec, lists, selection, nested views, forms, selects, switches, conditional display, computed values, notifications, **agent chat**.

### 1. Design Spec (`.ui-mcp/design/ui-contacts.md`)

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
| Contact list  | ui-viewlist="contacts"      | namespace="contact-row" |
| Row checkbox  | ui-value="item.selected"    | Multi-select            |
| Row name      | ui-value="item.name"        | Click selects           |
| Detail panel  | ui-view="current"           | Shows selected contact  |
| Name input    | ui-value="current.name"     |                         |
| Email input   | ui-value="current.email"    |                         |
| Status select | ui-value="current.status"   | active/inactive         |
| VIP switch    | ui-value="current.vip"      |                         |
| Delete btn    | ui-action="deleteCurrent()" | variant="danger"        |
| Save btn      | ui-action="save()"          | Fires notify            |
| Status line   | ui-value="statusLine()"     | Computed                |
| Chat messages | ui-viewlist="messages"      | namespace="chat-msg"    |
| Chat input    | ui-value="chatInput"        | ?keypress for live      |
| Send btn      | ui-action="sendChat()"      | Notifies agent          |

## Behavior
- Click row → selects contact, shows in detail panel
- Save → mcp.notify("contact_saved", {contact})
- Delete → removes from list, clears detail
- No selection → hide detail panel (ui-class-hidden)
- Send chat → mcp.notify("chat", {text}) → agent responds via ui_run
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
        darkMode = false,
        messages = {},      -- Chat history
        chatInput = ""      -- Current input
    }, self)
end

function ContactApp:add()
    local c = Contact:new("New Contact")
    table.insert(self.contacts, c)
    self.current = c
end

function ContactApp:select(contact)
    self.current = contact
end

function ContactApp:deleteCurrent()
    if not self.current then return end
    for i, c in ipairs(self.contacts) do
        if c == self.current then table.remove(self.contacts, i); break end
    end
    self.current = nil
end

function ContactApp:save()
    if self.current then mcp.notify("contact_saved", { name = self.current.name, email = self.current.email }) end
end

function ContactApp:statusLine()
    local sel = 0
    for _, c in ipairs(self.contacts) do if c.selected then sel = sel + 1 end end
    return #self.contacts .. " contacts • " .. sel .. " selected"
end

function ContactApp:hasCurrent() return self.current ~= nil end

-- Chat methods
function ContactApp:sendChat()
    if self.chatInput == "" then return end
    -- Add user message to history
    table.insert(self.messages, ChatMessage:new("You", self.chatInput))
    -- Notify agent with the message
    mcp.notify("chat", { text = self.chatInput })
    self.chatInput = ""
end

-- Called by agent via ui_run to add response
function ContactApp:addAgentMessage(text)
    table.insert(self.messages, ChatMessage:new("Agent", text))
end

app = app or ContactApp:new()
mcp.state = app
```

### 3. App Viewdef (`ContactApp.DEFAULT.html`)

```html
<template>
  <div ui-class-dark="darkMode">
    <div class="header">
      <h2>Contacts</h2>
      <sl-button ui-action="add()">+ Add</sl-button>
      <sl-switch ui-value="darkMode">Dark</sl-switch>
    </div>
    <div class="body">
      <div class="list" ui-viewlist="contacts" ui-namespace="contact-row"></div>
      <div class="detail" ui-class-hidden="!hasCurrent()">
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
      <div class="chat-messages" ui-viewlist="messages" ui-namespace="chat-msg"></div>
      <div class="chat-input">
        <sl-input ui-value="chatInput?keypress" placeholder="Type a message..."></sl-input>
        <sl-button ui-action="sendChat()" variant="primary">Send</sl-button>
      </div>
    </div>
  </div>
</template>
```

### 4. Item Viewdef (`lua.ViewListItem.contact-row.html`)

```html
<template>
  <div class="row" ui-action="item.select()" ui-class-current="item == list.current">
    <sl-checkbox ui-value="item.selected"></sl-checkbox>
    <span ui-value="item.name"></span>
  </div>
</template>
```

### 5. Chat Message Viewdef (`lua.ViewListItem.chat-msg.html`)

```html
<template>
  <div class="chat-message" ui-class-agent="item.sender == 'Agent'">
    <strong ui-value="item.sender"></strong>: <span ui-value="item.text"></span>
  </div>
</template>
```

### 6. Agent Response Pattern

When the agent receives a `chat` notification, respond via `ui_run`:

```lua
app:addAgentMessage("I can help you with that!")
```

## Resources

| Resource         | Content         |
|------------------|-----------------|
| `ui://reference` | Quick start     |
| `ui://lua`       | Lua API         |
| `ui://viewdefs`  | Viewdef syntax  |
| `ui://state`     | Live state JSON |

## Conventions

- Close button: top-right `[×]`
- Primary action: bottom-right
- Labels: "Submit" (not "Send"), "Cancel" (not "Close"), "Save" (not "Done")
- Enter → submit, Escape → cancel

## Debugging

- Check `.ui-mcp/log/lua.log`
- `ui_run` returns errors
- `ui://state` shows current state
