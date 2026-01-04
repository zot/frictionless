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

You can have multiple apps. Switch between them by assigning to `mcp.value`:

```lua
-- Define apps
contacts = ContactApp:new()
tasks = TaskApp:new()

-- Show contacts app
mcp.value = contacts

-- Later, switch to tasks app
mcp.value = tasks
```

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
4. Define    → ui_run(lua_code) → create classes
5. Template  → ui_upload_viewdef(type, ns, html)
6. Show      → ui_open_browser()
7. Listen    → mcp.notify() sends events back to you
8. Iterate   → Update state or viewdefs, user sees changes
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

Set `mcp.value` to display an object on screen:

```lua
mcp.value = MyForm:new()
```

**Key points**:
- `mcp.value` starts as `nil` (blank screen)
- The object MUST have a `type` field matching a viewdef
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

### Notifications

Send events from Lua back to the AI agent:

```lua
function Feedback:submit()
    mcp.notify("feedback_received", {
        rating = self.rating,
        comment = self.comment
    })
end
```

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
