# UI Platform Reference

Build interactive UIs for rich two-way communication with users. The platform uses a **Server-Side UI** architecture: application state lives in Lua on the server, and the browser acts as a thin renderer.

## Quick Start for AI Agents

```
1. Design    → Plan UI, create .ui-mcp/design/ui-{name}.md spec
2. Configure → ui_configure(base_dir=".ui-mcp")
3. Start     → ui_start() → returns URL
4. Define    → ui_run(lua_code) → create classes
5. Template  → ui_upload_viewdef(type, ns, html)
6. Show      → ui_open_browser()
7. Listen    → mcp.notify() sends events back to you
8. Iterate   → Update state or viewdefs, user sees changes
```

## Two-Phase Workflow

**Phase 1: Design** — Before writing code:
- Read `.ui-mcp/patterns/` for established UI patterns
- Read `.ui-mcp/conventions/` for layout and terminology rules
- Create/update `.ui-mcp/design/ui-{name}.md` layout spec

**Phase 2: Build** — Implement the design:
- Configure, start, run Lua, upload viewdefs, open browser

See [AI Interaction Guide](ui://mcp) for details.

## Core Concepts

### Displaying Objects

Set `mcp.state` to display an object on screen:

```lua
mcp.state = MyForm:new()
```

**Key points**:
- `mcp.state` starts as `nil` (blank screen)
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
    <h3 ui-text="fullName()"></h3>
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
.ui-mcp/
├── lua/            # Lua source files
├── viewdefs/       # HTML templates
├── log/            # Runtime logs
├── design/         # UI layout specs (prevents drift)
├── patterns/       # Reusable UI patterns
├── conventions/    # Layout, terminology, preferences
└── library/        # Proven implementations
```
