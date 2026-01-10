# UI Platform MCP Overview

The UI Platform provides a Model Context Protocol (MCP) server that enables AI agents to build, display, and modify "tiny apps" for rich two-way communication with users.

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
│  User interactions → mcp.pushState() → /wait → AI Agent    │
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

## MCP Resources

| Resource | Content |
|----------|---------|
| `ui://reference` | Quick start and core concepts |
| `ui://lua` | Lua API and patterns |
| `ui://viewdefs` | Viewdef syntax reference |
| `ui://mcp` | Agent workflow guide |
| `ui://state` | Live session state (JSON) |

## Agent Workflow

See [AGENTS.md](AGENTS.md) for detailed agent architecture. Summary:

```
┌─────────────────────────────────────────────────┐
│              DESIGN PHASE                        │
│  Read patterns/conventions                       │
│  Plan UI structure                               │
│  Create/update design spec (ui-*.md)             │
└─────────────────────┬───────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│              BUILD PHASE                         │
│  ui_configure + ui_start                         │
│  ui_run (Lua classes, idempotent pattern)        │
│  Write viewdefs (hot-loaded from disk)           │
│  ui_display("varName")                           │
│  ui_open_browser                                 │
└─────────────────────────────────────────────────┘
```

## Working Directory Structure

When the AI uses `.claude/ui/` as the base directory:

```
.claude/ui/
├── html/           # Static HTML
├── viewdefs/       # Viewdef templates
├── lua/            # Lua source files
├── log/            # Runtime logs
├── design/         # UI layout specs (prevents drift)
├── patterns/       # Reusable UI patterns
├── conventions/    # Established conventions
└── library/        # Proven implementations
```

## Quick Example

```lua
-- Define a feedback form (hot-loadable)
Feedback = session:prototype("Feedback", {
    rating = 5,
    comment = ""
})

function Feedback:submit()
    mcp.pushState({ app = "feedback", event = "submit", rating = self.rating, comment = self.comment })
end

-- Guard instance creation (idempotent)
if not session.reloading then
    feedback = Feedback:new()
end
```

The agent then calls `ui_display("feedback")` to show it in the browser.

```html
<!-- Viewdef for Feedback -->
<template>
  <div class="feedback-form">
    <sl-rating ui-value="rating"></sl-rating>
    <sl-textarea ui-value="comment" placeholder="Comments..."></sl-textarea>
    <sl-button ui-action="submit()">Submit</sl-button>
  </div>
</template>
```

## Related Documentation

- [AGENTS.md](AGENTS.md) — Agent architecture, two-phase workflow, drift prevention
- [ARCHITECTURE.md](ARCHITECTURE.md) — Core platform architecture (variables, wrappers, ViewList)
- [specs/mcp.md](specs/mcp.md) — Formal MCP specification
