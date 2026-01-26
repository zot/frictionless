---
name: ui-thorough
description: Full UI development workflow from specs to auditing. Use for new apps or major features.
---

# UI Thorough

Complete workflow for ui-engine apps: requirements → design → code → audit → simplify.

## Prerequisites

**You must understand ui-engine before building apps.** It's an alien framework - standard web patterns don't apply.

- `/ui` — directory structure, running UIs, event loop
- `/ui-basics` — bindings, state management, patterns, widgets

**Read both skills if you haven't already this session.** Don't re-read if you already have.

---

# Workflow

Follow these steps in order.

## Step 0: Consolidate Existing Checkpoints

**Check for existing checkpoints from `/ui-fast` prototyping:**

```bash
.ui/mcp checkpoint list <app>
```

If checkpoints exist (output is not "No checkpoints for <app>"):

1. **Thoroughly explore what was done:**

   The checkpoints contain ALL the prototyping work. Use the checkpoint tools to understand it completely:

   ```bash
   # List all checkpoints - messages describe each change
   .ui/mcp checkpoint list <app>

   # Show full diff from earliest checkpoint to current state
   .ui/mcp checkpoint diff <app> 10

   # Show diff from a specific checkpoint (e.g., checkpoint 3)
   .ui/mcp checkpoint diff <app> 3
   ```

   Read the diff carefully. It shows exactly what code was added, modified, or removed. The checkpoint messages provide context for each change.

2. **Update BOTH requirements.md AND design.md:**
   - Read current `requirements.md` and `design.md`
   - Update `requirements.md` first — add any new features that were prototyped
   - Update `design.md` to match — add data model fields, methods, viewdef changes
   - Use the diff output to ensure nothing is missed
   - **Both files must be updated** — the design must stay in sync with requirements

3. **Record as fast code gaps:**

   Since checkpoint work is rapid prototyping ("quick and dirty"), add a summary to `TESTING.md` under `## Gaps` with a `### Fast Code` subsection:

   ```markdown
   ## Gaps

   ### Fast Code

   The following were added via `/ui-fast` checkpoints and may need review:

   - `hasCheckpoints()` - checkpoint detection with caching
   - `refreshCheckpoints()` - batch refresh for all apps
   - `requestConsolidate()` - consolidate_request event
   ```

   This flags the code for later review/cleanup while preserving the feature.

4. **Clear checkpoints:**
   ```bash
   .ui/mcp checkpoint clear <app>
   ```

This ensures prototyping work is captured in the design before proceeding.

## Step 1: Create Todos (IMMEDIATELY)

Extract the app name from your prompt.

**Run this command BEFORE reading any files:**

```bash
.ui/mcp run "mcp:createTodos({'Read requirements', 'Requirements', 'Design', 'Write code', 'Write viewdefs', 'Link and audit', 'Simplify'}, 'APP_NAME')"
```

This shows progress in the UI. The user is watching - without this, the build looks frozen.

Also create Claude Code tasks:
```
TaskCreate: "Read requirements" (activeForm: "Reading requirements...")
TaskCreate: "Update requirements" (activeForm: "Updating requirements...")
TaskCreate: "Design changes" (activeForm: "Designing...")
TaskCreate: "Write code" (activeForm: "Writing code...")
TaskCreate: "Write viewdefs" (activeForm: "Writing viewdefs...")
TaskCreate: "Link and audit" (activeForm: "Auditing...")
TaskCreate: "Simplify code" (activeForm: "Simplifying...")
```

## Step 2: Read Requirements

```bash
.ui/mcp run "mcp:startTodoStep(1)"
```

TaskUpdate("Read requirements": in_progress)

- Check `{base_dir}/apps/<app>/requirements.md`
- If it does not exist, create it with human-readable prose
- If `{base_dir}/apps/<app>/TESTING.md` exists, read it and note Known Issues

## Step 3: Update Requirements

```bash
.ui/mcp run "mcp:startTodoStep(2)"
```

TaskUpdate("Read requirements": completed), TaskUpdate("Update requirements": in_progress)

- If the task requires changes to requirements.md, update it now
- Ensure requirements are complete and unambiguous before designing

## Step 4: Design

```bash
.ui/mcp run "mcp:startTodoStep(3)"
```

TaskUpdate("Update requirements": completed), TaskUpdate("Design changes": in_progress)

- Check `{base_dir}/patterns/` for reusable patterns
- Write `{base_dir}/apps/<app>/icon.html` with an emoji, `<sl-icon>`, or `<img>`
- Write `{base_dir}/apps/<app>/design.md`:
   - **Intent**: What the UI accomplishes
   - **Layout**: ASCII wireframe showing structure
   - **Data Model**: Tables of types, fields, descriptions
   - **Methods**: Actions each type performs
   - **ViewDefs**: Template files needed
   - **Events**: JSON examples with **complete handling instructions**

## Step 5: Write Code

```bash
.ui/mcp run "mcp:startTodoStep(4)"
```

TaskUpdate("Design changes": completed), TaskUpdate("Write code": in_progress)

Write `{base_dir}/apps/<app>/app.lua` — Lua classes and logic.

**Write code BEFORE viewdefs** — viewdefs may reference types that must exist first.

## Step 6: Write Viewdefs

```bash
.ui/mcp run "mcp:startTodoStep(5)"
```

TaskUpdate("Write code": completed), TaskUpdate("Write viewdefs": in_progress)

- `viewdefs/<Type>.DEFAULT.html` — HTML templates
- `viewdefs/<Item>.list-item.html` — List item templates (if needed)

## Step 7: Link and Audit

```bash
.ui/mcp run "mcp:startTodoStep(6)"
```

TaskUpdate("Write viewdefs": completed), TaskUpdate("Link and audit": in_progress)

**Create symlinks:**
```bash
.ui/mcp linkapp add <app>
```

**Automated audit:**
```bash
.ui/mcp audit $APP
```

The tool checks Lua code AND viewdefs for:
- Dead methods (defined but never called)
- Missing `session.reloading` guard
- Global variable name mismatch
- `<style>` blocks in list-item viewdefs
- `item.` prefix in list-item viewdefs
- `ui-action` on non-buttons
- `ui-class="hidden:..."` (should use `ui-class-hidden`)
- `ui-value` on checkboxes/switches
- Operators in binding paths
- HTML parse errors

**AI-based checks** (require reading comprehension):
- Compare design.md against requirements.md
- Compare implementation against design.md
- Check for missing `min-height: 0` on scrollable flex children
- Check that Cancel buttons revert changes

**Fix violations:**
1. Dead methods NOT in design.md → Delete from `app.lua`
2. Dead methods IN design.md → Record in `TESTING.md` under `## Gaps`
3. Other violations → Fix in code
4. Warnings (external methods) → OK to ignore

## Step 8: Simplify

```bash
.ui/mcp run "mcp:startTodoStep(7)"
```

TaskUpdate("Link and audit": completed), TaskUpdate("Simplify code": in_progress)

Use the `code-simplifier` agent:
```
Task tool with subagent_type="code-simplifier"
prompt: "Simplify the code in {base_dir}/apps/<app>/app.lua"
```

## Step 9: Complete

```bash
.ui/mcp run "mcp:completeTodos()"
.ui/mcp run "mcp:appUpdated('APP_NAME')"
.ui/mcp run "if appConsole then appConsole:addAgentMessage('Done - built APP_NAME') end"
```

TaskUpdate("Simplify code": completed)

---

# Preventing Drift

During iterative modifications, features can accidentally disappear:

1. **Before modifying** — Read the design spec
2. **Update spec first** — Modify the layout/components in the spec
3. **Then update code** — Change viewdef and Lua to match spec
4. **Verify** — Ensure implementation matches spec

The spec is the **source of truth**.

---

# Behavior Location

| Location | Use For | Trade-offs |
|----------|---------|------------|
| **Lua** | All behavior whenever possible | Simple, fast, responsive |
| **Claude** | Complex logic, external APIs | Slow (event loop latency) |
| **JavaScript** | Browser APIs, DOM tricks | Last resort |

---

# Examples

See `.claude/skills/ui-builder/examples/` for complete examples:
- `requirements.md` — Requirements spec
- `design.md` — Design spec
