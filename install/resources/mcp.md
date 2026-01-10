# MCP Interaction Guide for AI Agents

Use this platform to build "tiny apps" for rich two-way communication with the user. Instead of text-only interaction, you can create visual interfaces with forms, lists, and interactive components.

## When to Use UI

**Proactively consider building a UI when:**

- Collecting structured input (forms, selections, ratings, file picks)
- Presenting data that benefits from layout (lists, tables, comparisons)
- Multi-step workflows (wizards, confirmations, progress tracking)
- Real-time feedback loops (editing, previewing, validation)
- User needs to choose from multiple options
- Complex information that's hard to convey in text

**Stick with text when:**

- Simple yes/no questions
- Brief information delivery
- User explicitly prefers text
- One-shot responses with no follow-up

## Server Lifecycle

The MCP server has three states:

```
UNCONFIGURED ──ui_configure──► CONFIGURED ──ui_start──► RUNNING
                                    ▲                       │
                                    └───ui_configure────────┘
                                      (restarts session)
```

| State | What Works |
|-------|------------|
| UNCONFIGURED | Only `ui_configure` |
| CONFIGURED | `ui_configure` again (reconfigure) |
| RUNNING | `ui_run`, `ui_upload_viewdef`, `ui_open_browser`, `ui_configure` (restarts session) |

**Always call `ui_configure` first**, then `ui_start`.

## Two-Phase Workflow

### Phase 1: Design

Before writing code, understand and plan:

1. **Read existing patterns** — Check `.claude/ui/patterns/` for established UI patterns
2. **Read conventions** — Check `.claude/ui/conventions/` for layout and terminology rules
3. **Check for similar UIs** — Look in `.claude/ui/design/` for existing layout specs
4. **Create/update design spec** — Write `.claude/ui/design/ui-{name}.md` with:
   - Intent (what this UI accomplishes)
   - ASCII layout (visual structure)
   - Components table (element, binding, notes)
   - Behavior (interaction rules)

### Phase 2: Build

Implement the design:

1. `ui_configure(base_dir=".claude/ui")` — Set up environment
2. `ui_start()` — Start HTTP server (returns URL)
3. `ui_run(code)` — Define Lua classes with `session:prototype()`, guard with `session:getApp()`
4. Write viewdefs to `apps/<app>/viewdefs/` — Templates are hot-loaded
5. `ui_open_browser()` — Open browser to user

## Directory Structure

Use `.claude/ui/` as your base directory:

```
.claude/ui/
├── lua/            # Lua source (main.lua loaded automatically)
├── viewdefs/       # Viewdef templates
├── log/            # Runtime logs (check lua.log for errors)
│
├── design/         # UI layout specs (prevents drift)
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
    ├── viewdefs/       # Tested templates
    └── lua/            # Tested code
```

## Collaborative Loop

1. **Show UI** — Write viewdefs to `apps/<app>/viewdefs/` (hot-loaded)
2. **User interacts** — Clicks, types, selects
3. **Receive event** — Lua calls `mcp.pushState(event)`, agent polls `/wait`
4. **Inspect state** — Read `ui://state` to see current values
5. **Update UI** — Edit Lua or viewdef files (hot-loaded automatically)
6. **Repeat** — Continue the conversation visually

## Preventing Drift

During iterative changes, features can accidentally disappear. To prevent this:

1. **Before modifying** — Read the design spec (`.claude/ui/design/ui-*.md`)
2. **Update spec first** — Add/change components in the spec
3. **Then update code** — Modify viewdef and Lua to match
4. **Verify** — Check that implementation matches spec

The spec is the source of truth. If it says a close button exists, don't remove it.

## Example: Feedback Form

### Design Spec (`.claude/ui/design/ui-feedback.md`)

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
```

### Lua Code

```lua
-- Hot-loadable class definition
Feedback = session:prototype("Feedback", {
    rating = 5,
    comment = ""
})

function Feedback:submit()
    mcp.pushState({
        app = "feedback",
        event = "submit",
        rating = self.rating,
        comment = self.comment
    })
end

-- Guard instance creation (idempotent)
if not session.reloading then
    feedback = Feedback:new()
end
```

The agent then calls `ui_display("feedback")` to show it in the browser.

### Viewdef

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

## Best Practices

- **Use `session:prototype()`** — Define hot-loadable classes
- **Guard instance creation** — `if not session.reloading then ... end`
- **Use `EMPTY`** — For optional fields that start nil but need mutation tracking
- **Instance naming** — lowercase camelCase matching app directory (e.g., `feedback` for `feedback/` app)
- **Atomic viewdefs** — One type per viewdef, keep them focused
- **Informative events** — Include enough context in `mcp.pushState()` params
- **Use hot-loading** — Edit files directly; changes auto-refresh in browser
- **Check logs** — Read `.claude/ui/log/lua.log` when debugging
- **Follow conventions** — Read `.claude/ui/conventions/` before creating UI
- **Update specs** — Keep `.claude/ui/design/ui-*.md` in sync with implementation
