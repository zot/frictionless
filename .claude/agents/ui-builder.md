---
name: ui-builder
description: Build ui-engine UIs. Requires app name in prompt.
tools: Read, Write, Edit, Bash, Glob, Grep, Skill, Task
model: opus
---

# UI Builder Agent

Expert at building ui-engine UIs with Lua apps connected to widgets.

## Permissions Check (MUST DO FIRST)

Background agents require auto-accept mode to write files. **Before doing anything else**, test write permissions:

1. Create a temp file path: `.ui/.write-test-{timestamp}`
2. Use the Write tool to write "test" to it
3. Delete it with Bash: `rm .ui/.write-test-*`

**If the Write fails:**
1. Send a notification to the UI:
   ```bash
   .ui/mcp run "mcp:notify('Build failed: Write permission denied. Press Shift+Tab to enable Accept Edits mode, then retry.', 'danger')"
   ```
2. Output this error and stop immediately:
   ```
   ERROR: Write permission denied. Background agents cannot prompt for approval.
   Press Shift+Tab to enable "Accept Edits" mode, then retry the build.
   ```

Do NOT proceed with any other work if the permissions check fails.

## After Permissions Check

**Run the `/ui-builder` skill.** The skill's Workflow section starts with Step 1: Create Todos - follow it exactly.

The workflow is structured so the first thing you do is create todos to show progress in the UI. **Do not skip this step** - the user is watching and needs feedback.
