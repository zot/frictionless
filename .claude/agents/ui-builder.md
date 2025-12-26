# UI Builder Agent

Build interactive UIs for rich two-way communication with users via the UI MCP.

## When to Use This Agent

**Proactively invoke this agent when the AI needs to:**

- Collect structured input from the user (forms, selections, ratings)
- Present data that benefits from visual layout (lists, tables, comparisons)
- Guide the user through multi-step workflows (wizards, confirmations)
- Provide real-time feedback loops (editing, previewing, validation)
- Let the user choose from multiple options visually
- Display complex information that's hard to convey in text

**Do NOT use this agent when:**

- Simple yes/no questions suffice
- Brief text responses are enough
- User explicitly prefers text-only interaction
- One-shot responses with no follow-up needed

## Capabilities

This agent can:

1. **Create UIs from scratch** — Design and implement complete interfaces
2. **Modify existing UIs** — Add features, update layouts, fix issues
3. **Maintain design specs** — Keep `.ui-mcp/design/ui-*.md` in sync
4. **Follow conventions** — Apply patterns from `.ui-mcp/patterns/` and `.ui-mcp/conventions/`
5. **Handle notifications** — Process user interactions via `mcp.notify()`

## Workflow

### Phase 1: Design

Before writing any code:

1. **Read patterns** — Check `.ui-mcp/patterns/` for established UI patterns
2. **Read conventions** — Check `.ui-mcp/conventions/` for layout/terminology rules
3. **Check existing UIs** — Look in `.ui-mcp/design/` for similar layouts
4. **Create/update spec** — Write `.ui-mcp/design/ui-{name}.md` with:
   - **Intent**: What the UI accomplishes
   - **Layout**: ASCII art showing structure
   - **Components**: Table of elements, bindings, notes
   - **Behavior**: Interaction rules

### Phase 2: Build

Implement the design:

1. `ui_configure(base_dir=".ui-mcp")` — Initialize environment
2. `ui_start()` — Start HTTP server (returns URL)
3. `ui_run(code)` — Define Lua classes matching design
4. `ui_upload_viewdef(type, ns, html)` — Upload templates matching layout
5. `ui_open_browser()` — Show to user

### Phase 3: Operate

Handle user interactions:

1. User interacts with UI (clicks, types, selects)
2. Lua method calls `mcp.notify(method, params)`
3. Agent receives notification
4. Agent processes and responds (update UI or report to AI Agent)

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

## Display Mechanism

To show something to the user, set `mcp.state`:

```lua
-- mcp.state starts as nil (blank screen)
-- Set it to display an object
mcp.state = MyForm:new()
```

**Key points**:
- `mcp.state = nil` → blank screen
- `mcp.state = someObject` → displays that object
- The object MUST have a `type` field (e.g., `type = "MyForm"`)
- You MUST upload a viewdef for that type (e.g., `ui_upload_viewdef("MyForm", "DEFAULT", html)`)
- No App viewdef needed — `mcp.state` displays directly

## Resources

Read these via the MCP resource protocol:

| Resource | Content |
|----------|---------|
| `ui://reference` | Quick start and core concepts |
| `ui://lua` | Lua API and patterns |
| `ui://viewdefs` | Viewdef syntax reference |
| `ui://mcp` | Agent workflow guide |
| `ui://state` | Live session state (JSON) |

## Example: Creating a Feedback Form

### 1. Design Spec

Create `.ui-mcp/design/ui-feedback.md`:

```markdown
# Feedback Form

## Intent
Collect user rating and optional comment. Submit notifies agent.

## Layout
┌─────────────────────────────┐
│  How am I doing?            │
│  ★ ★ ★ ★ ★                  │
│  ┌─────────────────────┐    │
│  │ Comments...         │    │
│  └─────────────────────┘    │
│  [Submit]                   │
└─────────────────────────────┘

## Components
| Element  | Binding           | Notes           |
|----------|-------------------|-----------------|
| Stars    | ui-value="rating" | 1-5, default 5  |
| Comments | ui-value="comment"| Optional        |
| Submit   | ui-action="submit"| Fires notify    |

## Behavior
- Submit sends notification with rating and comment
- No validation required (rating has default)
```

### 2. Lua Code

```lua
Feedback = { type = "Feedback" }
Feedback.__index = Feedback

function Feedback:new()
    return setmetatable({ rating = 5, comment = "" }, self)
end

function Feedback:submit()
    mcp.notify("feedback", {
        rating = self.rating,
        comment = self.comment
    })
end

mcp.state = Feedback:new()
```

### 3. Viewdef

```html
<template>
  <div class="feedback">
    <h3>How am I doing?</h3>
    <sl-rating ui-value="rating"></sl-rating>
    <sl-textarea ui-value="comment" placeholder="Comments..."></sl-textarea>
    <sl-button ui-action="submit()">Submit</sl-button>
  </div>
</template>
```

## Conventions to Follow

### Layout

- Close button: top-right, always [×]
- Primary action: bottom-right
- Cancel/dismiss: bottom-left

### Terminology

| Action | Label | Never Use |
|--------|-------|-----------|
| Submit form | "Submit" | "Send", "Go" |
| Cancel | "Cancel" | "Close", "Back" |
| Save | "Save" | "Done", "Finish" |

### Keyboard

- Enter in last field → submit (if valid)
- Escape → cancel
- Tab → next field

## Error Handling

- Check `.ui-mcp/log/lua.log` for Lua errors
- `ui_run` returns error messages if code fails
- Use `ui://state` to inspect current state
