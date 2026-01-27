---
name: ui-fast
description: Rapid UI prototyping with checkpointing. Use for quick iterations on ui-engine apps.
---

# UI Fast

Rapid prototyping for ui-engine apps. Make changes, checkpoint, try alternatives, rollback.

## Prerequisites

**You must understand ui-engine before making changes.** It's an alien framework - standard web patterns don't apply.

- `/ui` — directory structure, running UIs, event loop
- `/ui-basics` — bindings, state management, patterns, widgets

**Read both skills if you haven't already this session.** Don't re-read if you already have.

---

# Checkpointing

Checkpoints use fossil SCM under the hood, managed via `.ui/mcp checkpoint` commands. Fossil is auto-downloaded to `~/.claude/bin` on first use.

## Commands

```bash
# Save a checkpoint (creates repo on first use, no-op if no changes)
.ui/mcp checkpoint save <app> "description"

# List checkpoints
.ui/mcp checkpoint list <app>

# Show diff from nth checkpoint to current (1 = most recent)
.ui/mcp checkpoint diff <app> [n]

# Rollback to nth checkpoint
.ui/mcp checkpoint rollback <app> <n>

# Clear all checkpoints (deletes fossil repo)
.ui/mcp checkpoint clear <app>
```

## Workflow

```
.ui/mcp checkpoint save <app> "before: add search"
[make changes to app.lua / viewdefs]
.ui/mcp checkpoint save <app> "add search"

.ui/mcp checkpoint save <app> "before: fix styling"  # no-op if no changes
[make changes]
.ui/mcp checkpoint save <app> "fix styling"
```

**Key behavior**: `save` is a no-op if nothing changed since last checkpoint. This allows a simple "save, work, save" rhythm.

## Rollback

```bash
# List checkpoints to see history
.ui/mcp checkpoint list <app>
# Output:
# a2b9eeedc0 fix styling
# c07d7ee460 add search
# eb2185e694 initial state

# Rollback to checkpoint 2 (restores "add search" state)
.ui/mcp checkpoint rollback <app> 2
```

## Diff

```bash
# Show what changed since checkpoint 2
.ui/mcp checkpoint diff <app> 2
```

---

# User Commands

| User says | Action |
|-----------|--------|
| "add search" | `checkpoint save` before, make change, `checkpoint save` after |
| "rollback" or "undo" | `checkpoint list`, ask which to restore, `checkpoint rollback` |
| "rollback to 2" | `checkpoint rollback <app> 2` |
| "show checkpoints" | `checkpoint list <app>` |
| "clear checkpoints" | `checkpoint clear <app>` |

---

# Reporting progress

**Run this command BEFORE reading any files:**

```bash
.ui/mcp run "mcp:createTodos({'Fast requirements', 'Fast code', 'Fast viewdefs', 'Fast finish'}, 'APP_NAME')"
```

This shows progress in the UI. The user is watching - without this, the build looks frozen.

Also create Claude Code tasks:
```
TaskCreate: "Fast requirements" (activeForm: "Reading requirements...")
TaskCreate: "Fast code" (activeForm: "Writing code...")
TaskCreate: "Fast viewdefs" (activeForm: "Writing viewdefs...")
TaskCreate: "Fast finish" (activeForm: "Finishing...")
```



# Consolidation

Checkpoints are **ephemeral** - they exist during rapid prototyping, then get cleared after changes are incorporated into the design.

When switching to `/ui-thorough` or when user says "update the design":

1. Review changes: `checkpoint list` and `checkpoint diff`
2. Update `requirements.md` and `design.md` to reflect the changes
3. Clear checkpoints: `checkpoint clear <app>`

The `/ui-thorough` skill automatically checks for existing checkpoints and consolidates them before proceeding.
