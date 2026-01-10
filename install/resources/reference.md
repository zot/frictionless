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

Each app lives in `.claude/ui/apps/<app>/`:

```
.claude/ui/apps/contacts/
├── app.lua              # Lua classes and logic
├── viewdefs/            # HTML templates
│   ├── ContactApp.DEFAULT.html
│   └── Contact.list-item.html
├── README.md            # Events, state, methods (for agent)
└── design.md            # UI layout spec (prevents drift)
```

### Multiple Apps

You can have multiple apps. The agent uses `ui_display("varName")` to show them:

```lua
-- Define apps
ContactApp = session:prototype("ContactApp", {})
TaskApp = session:prototype("TaskApp", {})

-- Guard instance creation
if not session.reloading then
    contacts = ContactApp:new()
    tasks = TaskApp:new()
end
```

The agent calls `ui_display("contacts")` or `ui_display("tasks")` to show the desired app.

### Event-Driven Communication

Users interact with the UI, which pushes events to the agent:

```
┌─────────┐    click/type    ┌─────────┐   mcp.pushState()   ┌─────────┐
│  User   │ ───────────────► │   UI    │ ──────────────────► │  Agent  │
└─────────┘                  └─────────┘                     └────┬────┘
     ▲                                                            │
     │                       ┌─────────┐      ui_run()            │
     └───────────────────────│ Browser │ ◄────────────────────────┘
           sees changes      └─────────┘   updates state
```

**Lua side** — Push events to the queue:
```lua
function ContactApp:sendChat()
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
- `.claude/ui/patterns/` — UI patterns (forms, lists, master-detail)
- `.claude/ui/conventions/` — Layout rules, terminology, preferences
- `.claude/ui/library/` — Proven implementations to copy from

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
│  Lifecycle:          ui_configure → ui_start                │
│  Code execution:     ui_run(lua_code)                       │
│  UI templates:       ui_upload_viewdef(type, ns, html)      │
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

The MCP server operates as a Finite State Machine:

```
UNCONFIGURED ──ui_configure──► CONFIGURED ──ui_start──► RUNNING
     │                              │      ◄──────────────────┘
     │ Only ui_configure allowed    │       ui_configure
     │                              │       (restarts session)
```

## MCP Tools

| Tool | Purpose |
|------|---------|
| `ui_configure` | Set base directory, initialize filesystem |
| `ui_start` | Start HTTP server on ephemeral port |
| `ui_run` | Execute Lua code in session context |
| `ui_upload_viewdef` | Upload HTML template for a type |
| `ui_open_browser` | Open browser to session (with conserve mode) |
| `ui_install` | Install agent files and CLAUDE.md instructions |
| `ui_status` | Get server state and connection count |

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
1. Design    → Plan UI, create .claude/ui/design/ui-{name}.md spec
2. Configure → ui_configure(base_dir=".claude/ui")
3. Start     → ui_start() → returns URL
4. Define    → ui_run(lua_code) → create classes with session:prototype()
5. Template  → Write viewdefs to apps/<app>/viewdefs/ (hot-loaded)
6. Browser   → ui_open_browser()
7. Listen    → Poll /wait endpoint for mcp.pushState() events
8. Iterate   → Edit files, changes appear instantly (hot-loading)
```

## Two-Phase Workflow

**Phase 1: Design** — Before writing code:
- Read `.claude/ui/patterns/` for established UI patterns
- Read `.claude/ui/conventions/` for layout and terminology rules
- Create/update `.claude/ui/design/ui-{name}.md` layout spec

**Phase 2: Build** — Implement the design:
- Configure, start, run Lua, upload viewdefs, open browser

See [AI Interaction Guide](ui://mcp) for details.

## Core Concepts

### Displaying Objects

The agent uses `ui_display("varName")` to show a Lua variable in the browser:

```lua
-- Define with session:prototype for hot-loading
MyForm = session:prototype("MyForm", {
    userInput = "",
    error = EMPTY,  -- EMPTY: starts nil, but tracked for mutation
})

-- Guard instance creation (idempotent)
if not session.reloading then
    myForm = MyForm:new()
end
```

The agent then calls `ui_display("myForm")` to show it.

**Key points**:
- The object MUST have a `type` field (set automatically by `session:prototype()`)
- Use `session:prototype()` for hot-loadable classes
- Guard with `if not session.reloading` for idempotency
- Instance name = lowercase camelCase matching app directory
- Use `EMPTY` for optional fields that start nil
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
1. Save a file in `.claude/ui/apps/<app>/`
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
.claude/ui/
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
