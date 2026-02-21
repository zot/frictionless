---
name: ui-thorough
description: Full UI development workflow from specs to auditing. Use for new apps or major features.
---

# UI Thorough

Complete workflow for ui-engine apps: requirements → design → code → audit → simplify.

## On Skill Load

**IMMEDIATELY invoke `/ui-basics` using the Skill tool before doing anything else.** It covers the mental model, bindings, object model, and patterns. Without it, standard web patterns will lead you astray.

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

   Use `checkpoint diff <app> 10` to see the full diff from earliest to current. Use `checkpoint diff <app> N` for specific checkpoints. The checkpoint messages (from `list`) provide context for each change.

2. **Update BOTH requirements.md AND design.md:**
   - Read current `requirements.md` and `design.md`
   - Update `requirements.md` first — add any new features that were prototyped
   - Update `design.md` to match — add data model fields, methods, viewdef changes
   - Use the diff output to ensure nothing is missed
   - **Both files must be updated** — the design must stay in sync with requirements

3. **Simplify the checkpointed code:**

   Run the code-simplifier agent on the app's Lua code:
   ```
   Task tool with subagent_type="code-simplifier"
   prompt: "Simplify the code in {base_dir}/apps/<app>/app.lua"
   ```

4. **Record as fast code gaps:**

   Since checkpoint work is rapid prototyping ("quick and dirty"), add a summary to `TESTING.md` under `## Gaps` with a `### Fast Code` subsection:

   Add a `### Fast Code` subsection under `## Gaps` in `TESTING.md` listing the methods/features added via checkpoints that may need review. This flags the code for later cleanup while preserving the feature.

5. **Clear checkpoints:**
   ```bash
   .ui/mcp checkpoint clear <app>
   ```

This ensures prototyping work is captured in the design before proceeding.

## Step 1: Create Progress Steps (IMMEDIATELY)

**Run BEFORE reading any files.** The user is watching — without this, the build looks frozen.

Steps: `{'Read requirements', 'Requirements', 'Design', 'Write code', 'Write viewdefs', 'Link and audit', 'Simplify', 'Set baseline'}`

Use `mcp:createTodos(STEPS, APP)` with the steps list and app name. The listed steps above all have matching entries in `UI_STEP_DEFS` (`apps/mcp/app.lua`) for tailored progress messages; new labels work but get generic ones. Also create Claude Code tasks (see MCP Methods in `/ui-basics`).

**Progress pattern:** At each step transition, call `mcp:startTodoStep(N)` and update Claude Code tasks. At the end, call `mcp:completeTodos()`.

## Step 2: Requirements

- Read `{base_dir}/apps/<app>/requirements.md` (create with prose if missing)
- Read `TESTING.md` if it exists — note Known Issues
- Update requirements if the task requires changes
- Ensure requirements are complete and unambiguous before designing

## Step 3: Design

- Check `.ui/patterns/` for reusable patterns
- Write `{base_dir}/apps/<app>/icon.html` with an emoji, `<sl-icon>`, or `<img>`
- Write `{base_dir}/apps/<app>/design.md`:
   - **Intent**: What the UI accomplishes
   - **Layout**: ASCII wireframe showing structure
   - **Data Model**: Tables of types, fields, descriptions
   - **Methods**: Actions each type performs
   - **ViewDefs**: Template files needed
   - **Events**: JSON examples with **complete handling instructions**

## Step 4: Write Code

Write `{base_dir}/apps/<app>/app.lua` — Lua classes and logic.

**Write code BEFORE viewdefs** — viewdefs may reference types that must exist first.

## Step 5: Write Viewdefs

- `viewdefs/<Type>.DEFAULT.html` — HTML templates
- `viewdefs/<Item>.list-item.html` — List item templates (if needed)

## Step 6: Link and Audit

```bash
.ui/mcp linkapp add <app>
.ui/mcp audit <app>
.ui/mcp theme audit <app>
```

The audit tool checks Lua code and viewdefs for common violations. Also do **AI-based checks**:
- Compare design.md against requirements.md
- Compare implementation against design.md

**Fix violations:**
1. Dead methods NOT in design.md → Delete from `app.lua`
2. Dead methods IN design.md → Record in `TESTING.md` under `## Gaps`
3. Other violations → Fix in code
4. Warnings (external methods) → OK to ignore

**Clear Empty Gaps:**
If there were gaps and they are all gone now, leave the `## Gaps` section empty, do not leave a comment.

## Step 7: Simplify

Use the `code-simplifier` agent:
```
Task tool with subagent_type="code-simplifier"
prompt: "Simplify the code in {base_dir}/apps/<app>/app.lua"
```

## Step 8: Set Baseline

Commit the audited code to the local branch and set a clean baseline for future `/ui-fast` iterations:

```bash
.ui/mcp checkpoint local <app> "thorough: <brief description>"
.ui/mcp checkpoint baseline <app>
```

The `local` commit preserves audited code on a permanent branch that survives baseline resets.
The `baseline` clears trunk for fresh `/ui-fast` checkpoints.

## Step 9: Complete

```bash
.ui/mcp run "mcp:completeTodos()"
.ui/mcp run "mcp:appUpdated('APP_NAME')"
.ui/mcp run "mcp:addAgentMessage('Done - built APP_NAME')"
```

---

# Examples

See `.ui/apps/app-console` for complete examples:
- `requirements.md` — Requirements spec
- `design.md` — Design spec
