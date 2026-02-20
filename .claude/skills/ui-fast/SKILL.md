---
name: ui-fast
description: Rapid UI prototyping with checkpointing. Use for quick iterations on ui-engine apps.
---

# UI Fast

Rapid prototyping for ui-engine apps. Make changes, checkpoint, try alternatives, rollback.

## On Skill Load

**IMMEDIATELY invoke `/ui-basics` using the Skill tool before doing anything else.** It covers the mental model, bindings, object model, and patterns. Without it, standard web patterns will lead you astray.

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

**Key behavior**: `save` is a no-op if nothing changed since last checkpoint. This allows a simple "save, work, save" rhythm.

**Example:** `checkpoint list` output shows numbered history — use the number with `rollback` or `diff`:
```
a2b9eeedc0 fix styling       # checkpoint 1 (most recent)
c07d7ee460 add search         # checkpoint 2
eb2185e694 initial state      # checkpoint 3
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

# Workflow

**Always checkpoint and show progress**, even for small fixes. The user is watching — without progress, the build looks frozen.

1. **Checkpoint** before starting
2. **Create progress steps** appropriate to the work (see below)
3. Write `{base_dir}/apps/<app>/icon.html` (if new app)
4. Check `.ui/patterns/` for reusable patterns (if building new features)
5. **Do the work**, updating progress as you go
6. **Verify** (see Fast verify below)
7. Run `.ui/mcp audit APP`
8. **Checkpoint** after finishing

## Progress Steps

Pick steps relevant to what you're changing:

| Change type | Steps |
|-------------|-------|
| CSS/styling fix | `{'Fix viewdefs'}` |
| Lua logic fix | `{'Fix code'}` |
| New feature / full build | `{'Fast requirements', 'Fast code', 'Fast viewdefs', 'Fast verify', 'Fast finish'}` |

**Fast verify**: Go through requirements item by item and verify the code/viewdefs meet them. Fix any gaps before finishing. Check behavioral requirements like scrolling, resizing, and hiding.

Create both MCP progress and Claude Code tasks (see MCP Methods in `/ui-basics`).

# Consolidation

Checkpoints are **ephemeral** - they exist during rapid prototyping, then get cleared after changes are incorporated into the design.

When switching to `/ui-thorough` or when user says "update the design":

1. `checkpoint save`
2. Review changes: `checkpoint list` and `checkpoint diff`
3. Update `requirements.md` and `design.md` to reflect the changes
4. Clear checkpoints: `checkpoint clear <app>`

The `/ui-thorough` skill automatically checks for existing checkpoints and consolidates them before proceeding.
