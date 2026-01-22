# UI Platform Reference

Build interactive UIs for rich two-way communication with users. The platform uses a **Server-Side UI** architecture: application state lives in Lua on the server, and the browser acts as a thin renderer.

## Tiny Apps

The core concept is building **tiny apps** — small, purpose-built UIs for specific interactions. Instead of text-only conversation, the agent creates visual interfaces.

**Use cases:**
- **Prototyping** — Agent and user collaborate on UI wireframes for production apps
- **Testing** — Create mock UIs for testing workflows
- **Structured input** — Forms, selections, ratings, file picks
- **Data presentation** — Lists, tables, comparisons, dashboards
- **Multi-step workflows** — Wizards, confirmations, progress tracking
- **Real-time collaboration** — Editing, previewing, validation
- **Claude Apps** — Persistent UIs for interacting with Claude:
  - Launch panels with buttons for design, implement, analyze gaps
  - Project dashboards showing available commands, agents, skills
  - Status displays for background tasks and build progress

### App Structure

Each app lives in `.ui/apps/<app>/`:

```
.ui/apps/contacts/
├── app.lua              # Lua classes and logic
├── viewdefs/            # HTML templates
│   ├── ContactApp.DEFAULT.html
│   └── Contact.list-item.html
├── README.md            # Events, state, methods (for agent)
└── design.md            # UI layout spec (prevents drift)
```

### Multiple Apps

You can have multiple apps. The agent uses `ui_display("appName")` to show them:

```lua
-- Define app prototypes (each serves as its own namespace)
Contacts = session:prototype("Contacts", {})
Tasks = session:prototype("Tasks", {})

-- Guard instance creation
if not session.reloading then
    contacts = Contacts:new()
    tasks = Tasks:new()
end
```

The agent calls `ui_display("contacts")` or `ui_display("tasks")` to show the desired app.

### Event-Driven Communication

Users interact with the UI, which pushes events to the agent:

```
┌─────────┐    click/type    ┌─────────┐   mcp.pushState()   ┌─────────┐
│  User   │ ◄──────────────► │   UI    │ ──────────────────► │  Agent  │
└─────────┘   Lua responds   └─────────┘                     └────┬────┘
     ▲                                                            │
     │                       ┌─────────┐      ui_run()            │
     └───────────────────────│ Browser │ ◄────────────────────────┘
           sees changes      └─────────┘   updates state
```

**Lua side** — Push events to the queue:
```lua
function Contacts:sendChat()
    mcp.pushState({ app = "contacts", event = "chat", text = self.chatInput })
end
```

**Agent side** — Poll for events via `/wait` endpoint:
```bash
curl "http://127.0.0.1:PORT/wait?timeout=30"
# Returns: [{"app":"contacts","event":"chat","text":"hello"}]
```

### Pattern Library

As you build apps, common patterns emerge. Store reusable patterns in:
- `.ui/patterns/` — UI patterns (forms, lists, master-detail)
- `.ui/conventions/` — Layout rules, terminology, preferences
- `.ui/library/` — Proven implementations to copy from

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  AI Agent (Claude)                                          │
│                                                             │
│  Decides WHEN to use UI for user communication              │
│  Instructs UI Agent on WHAT to build                        │
│  Receives notifications from user interactions              │
└─────────────────────┬───────────────────────────────────────┘
                      │ MCP tools + resources
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  UI MCP Server (Go)                                         │
│                                                             │
│  Lifecycle:          auto-starts, ui_configure (optional)   │
│  Code execution:     ui_run(lua_code)                       │
│  Browser:            ui_open_browser()                      │
│                                                             │
│  Resources:          ui://reference, ui://lua, ui://mcp     │
│  State:              ui://state                             │
└─────────────────────┬───────────────────────────────────────┘
                      │ HTTP + WebSocket
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  Browser UI                                                 │
│                                                             │
│  Renders viewdefs, binds to Lua state                       │
│  User interactions → mcp.pushState() → AI Agent             │
└─────────────────────────────────────────────────────────────┘
```

## Server Lifecycle

The server auto-starts when the MCP connection is established:

```
STARTUP ──auto-configure──► RUNNING ◄──ui_configure──┐
                               │                      │
                               └──────────────────────┘
                                  (reconfigure)
```

- **Auto-start:** Server uses `--dir` flag (defaults to `.ui`) and starts automatically
- **Reconfigure:** Call `ui_configure(base_dir)` to restart with a different directory

## MCP Tools

| Tool | Purpose |
|------|---------|
| `ui_status` | Get server state, URL, and connection count |
| `ui_run` | Execute Lua code in session context |
| `ui_display` | Load and display an app by name |
| `ui_open_browser` | Open browser to session (with conserve mode) |
| `ui_configure` | Reconfigure with different base directory (optional) |
| `ui_install` | Install/update bundled skills and resources |

## MCP Resources

| Resource | Content |
|----------|---------|
| `ui://reference` | This document - quick start and core concepts |
| `ui://lua` | Lua API and patterns |
| `ui://viewdefs` | Viewdef syntax reference |
| `ui://mcp` | Agent workflow guide |
| `ui://state` | Live session state (JSON) |

## Quick Start for AI Agents

```
1. Status    → ui_status() → get base_dir and url (server auto-started)
2. Design    → Plan UI in apps/<app>/design.md
3. Code      → Write Lua in apps/<app>/app.lua (hot-loaded)
4. Template  → Write viewdefs to apps/<app>/viewdefs/ (hot-loaded)
5. Display   → ui_display("app-name") → load and show the app
6. Browser   → ui_open_browser() or navigate to {url}/?conserve=true
7. Listen    → Poll /wait endpoint for mcp.pushState() events
8. Iterate   → Edit files, changes appear instantly (hot-loading)
```

## Two-Phase Workflow

**Phase 1: Design** — Before writing code:
- Read `.ui/patterns/` for established UI patterns
- Read `.ui/conventions/` for layout and terminology rules
- Create/update `apps/<app>/design.md` layout spec

**Phase 2: Build** — Implement the design:
- Write Lua code, create viewdefs, display app, open browser

See [AI Interaction Guide](ui://mcp) for details.

## Core Concepts

### Displaying Objects

The agent uses `ui_display("appName")` to show a Lua variable in the browser:

```lua
-- Define app prototype (serves as namespace)
MyForm = session:prototype("MyForm", {
    userInput = "",
    error = EMPTY,  -- EMPTY: starts nil, but tracked for mutation
})

-- Guard instance creation (idempotent)
if not session.reloading then
    myForm = MyForm:new()
end
```

The agent then calls `ui_display("my-form")` to show it.

**Key points**:
- The `type` field is set automatically by `session:prototype()` from the name argument
- The `type` is used for viewdef resolution (e.g., `MyForm` → `MyForm.DEFAULT.html`)
- Use `session:prototype()` with arbitrary names (does not consult globals)
- Naming: `Name` (PascalCase) for prototype, `name` (camelCase) for instance
- Guard with `if not session.reloading` for idempotency
- Use `EMPTY` for optional fields that start nil
- Nested prototypes: `Name.SubType = session:prototype('Name.SubType', ...)` → `Name.SubType.list-item.html`
- Inspect current state via `ui://state`

### Presenters and Domain Objects

- **Domain Objects** — Pure data: `Contact`, `Task`, `Order`
- **Presenters** — UI wrappers that add state and behavior: `ContactPresenter` with `isEditing`, `delete()`, `save()`

Keep data clean in domain objects. Put interaction logic in presenters.

### Viewdefs

HTML templates that define how objects render:

```html
<template>
  <div class="contact-card">
    <h3 ui-value="fullName()"></h3>
    <sl-input ui-value="email" label="Email"></sl-input>
    <sl-button ui-action="save()">Save</sl-button>
  </div>
</template>
```

Viewdefs are matched by the object's `type` property and namespace.

### Change Detection

Changes to Lua tables are automatically detected and pushed to the browser. No manual update calls needed:

```lua
function MyForm:clear()
    self.name = ""       -- Automatically synced
    self.email = ""      -- Automatically synced
end
```

### Hot-Loading

Edit files in your IDE and see changes instantly — no server restart or manual refresh needed.

**Supported files:**
- **Lua files** (`.lua`) — Code re-executes, browser updates automatically
- **Viewdef files** (`.html`) — Templates reload, components re-render

**How it works:**
1. Save a file in `.ui/apps/<app>/`
2. ui-engine detects the change
3. Lua is re-executed or viewdef is reloaded
4. Browser automatically reflects changes

**Requirements for Lua hot-loading:**
- Use `session:prototype()` for class definitions
- Use `session:create()` for instance creation
- Guard instance creation: `if not session.reloading then ... end`
- Use `EMPTY` for optional fields that start nil but need mutation tracking

This enables rapid iteration: edit code, save, see results immediately.

### Events (UI → Agent)

Send events from Lua back to the AI agent via `mcp.pushState()`:

```lua
function Feedback:submit()
    mcp.pushState({
        app = "feedback",
        event = "submit",
        rating = self.rating,
        comment = self.comment
    })
end
```

The agent polls for events via the `/wait` HTTP endpoint. Events are returned as a JSON array.

## Detailed Guides

- [Viewdef Syntax](ui://viewdefs) — `ui-*` attributes, path syntax, lists
- [Lua API & Patterns](ui://lua) — Classes, globals, change detection
- [AI Interaction Guide](ui://mcp) — Workflow, lifecycle, best practices

## Directory Structure

```
.ui/
├── apps/           # App source of truth
│   └── <app>/          # Each app has its own directory
│       ├── app.lua
│       ├── viewdefs/
│       ├── README.md
│       └── design.md
├── lua/            # Symlinks to apps/<app>/*.lua
├── viewdefs/       # Symlinks to apps/<app>/viewdefs/*
├── log/            # Runtime logs
├── patterns/       # Reusable UI patterns
├── conventions/    # Layout, terminology, preferences
└── library/        # Proven implementations
```
