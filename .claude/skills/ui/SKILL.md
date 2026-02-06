---
name: ui
description: use when **running Frictionless UIs** or needing to understand UI structure (before modifying one)
---

# UI MCP

Foundation for building and running ui-engine UIs with Lua apps connected to widgets.

## CRITICAL: Before Handling ANY Event

**ALWAYS read the design file for the app that sent the event FIRST.** The `app` field tells you which design to read:

```
Event: {"app":"app-console", "event":"app_created", "name":"contacts", "context":"contacts", "note":"...contacts..."}
       ^^^^^^^^^^^^^^^^
Read:  .ui/apps/app-console/design.md  <-- ONLY use the "app" field
```

**WARNING:** Do NOT be misled by `context`, `note`, `name`, or other fields that mention other app names. These are data for handling the event—they do NOT change which design.md you read. The `app` field is the ONLY field that determines the design file.

The design.md explains how to handle each event type. Do NOT skip this step—even for events that seem obvious like `app_created`.

## Handler Dispatch Rule

**NEVER modify or create UI code directly.** When an event requires UI changes (build, fix, chat requests about UI), you MUST:

1. Check the `handler` field in the event (`"/ui-fast"` or `"/ui-thorough"`)
2. Invoke that skill using `Skill(skill: "ui-fast")` or `Skill(skill: "ui-thorough")`
3. Pass the event context to the skill

This applies to ALL UI modifications—requirements, design, code, and viewdefs. The `/ui` skill handles event loops and display; it delegates all UI changes to the handler skill.

## Simple Requests

When the user says `/ui`, show app-console as in Quick Start. Prefer Playwright if connected.

When the user says `show APP`, show APP as in Quick Start. Prefer Playwright if connected.

When the user says `events` it means to start the event loop in the foreground, but not use `.ui/mcp display` or `.ui/mcp browser` as in Quick Start.

## Helper Script Reference

The `.ui/mcp` script provides commands for interacting with the UI server:

**CRITICAL: URLs must NEVER include session IDs.** Always use `{url}/?conserve=true` (root URL). Session IDs in URLs will cause problems.

| Command | Description |
|---------|-------------|
| `.ui/mcp status` | Get server status (url, sessions, base_dir) |
| `.ui/mcp browser` | Open browser to `{url}/?conserve=true` |
| `.ui/mcp display APP` | Display APP in the browser |
| `.ui/mcp run 'lua code'` | Execute Lua code in session |
| `.ui/mcp event` | Wait for next UI event (120s timeout) |
| `.ui/mcp state` | Get current session state |
| `.ui/mcp variables` | Get current variable values |
| `.ui/mcp audit APP` | Run code quality audit on APP |
| `.ui/mcp progress APP PERCENT STAGE` | Report build progress |
| `.ui/mcp linkapp add\|remove APP` | Manage app symlinks |
| `.ui/mcp checkpoint CMD APP [MSG]` | Manage checkpoints (save/list/rollback/diff/clear) |
| `.ui/mcp theme list` | List available themes |
| `.ui/mcp theme classes [THEME]` | List semantic classes for theme |
| `.ui/mcp theme audit APP [THEME]` | Audit app's theme class usage |

## How to Process Events

The event script waits for user interactions and returns JSON:

**Always use the relative path `.ui/mcp`**:

```bash
.ui/mcp event
```

**NOT the absolute path `/home/X/.ui/mcp`:**

```bash
/home/bubba/work/.ui/mcp event # <-- wrong, this is inconvenient for the user's permissions and also uses more tokens
```

This returns one JSON array per line containing one or more events:
```json
[{"app":"claude-panel","event":"chat","text":"Hello"},{"app":"claude-panel","event":"action","action":"commit"}]
```

As soon as you receive events, create a task to restart the event loop using TaskCreate—but only if there isn't already a pending restart task. Stacking multiple restart tasks just creates confusion.

**Remember:** Read the design.md for the `app` field (see "Before Handling ANY Event" above). Other fields like `context` or `note` provide data but do NOT change which design file you read.

### Running in Foreground or Background

**Foreground event loop (recommended):**
1. Run `.ui/mcp event` with Bash (blocking, ~2 min timeout)
2. When it returns, parse JSON array. If non-empty, handle each event
   - use `.ui/mcp run` to alter app state and reflect changes to the user
3. Restart the event loop

This is the most responsive approach - events are handled immediately.

**Background event loop (alternative):**
Run `.ui/mcp event` in background if you need to do other work while waiting. Note: this adds latency since you must poll the output file.

**CRITICAL: Kill previous event listener before restarting.**
If the event call runs in background or times out, the old listener may still be running. **Always use TaskStop to kill the previous task before starting a new `.ui/mcp event`.**

```
1. Track the task_id from `.ui/mcp event` (returned when running in background or via TaskOutput)
2. Before restarting: TaskStop(task_id=<previous_task_id>)
3. Then start new: `.ui/mcp event`
```

Failure to kill the old listener means it will consume events intended for the new one.

**Exit codes:**
- 0 + empty output = timeout, no events (kill old task, restart)
- 0 + JSON output = events received
- 52 = server restarted (restart both server and event loop)

## Quick Start: Show an Existing App

To display an app (e.g., `claude-panel`):

```
1. Read {base_dir}/apps/APP/design.md to learn:
   - The global variable name (e.g., `claudePanel`)
   - Event types and how to respond

2. Display and open:
   - use `.ui/mcp run` to see if `mcp.value.type == AppName` (PascalCase version of app-name)
     - if it's not already in mcp.value, use `.ui/mcp display app-name`
   - If Playwright is connected:
     - use browser_evaluate with `function: "() => window.location.href"` to get current URL
     - if URL does NOT start with {url}/, then browser_navigate to `{url}/mark-playwright.html?url={url}/?conserve=true`
       This marks `localStorage.playwright = "true"` so the UI can detect Playwright, then redirects to the app.
       **CRITICAL: ONLY use mark-playwright.html when navigating via Playwright.** Never use it for `.ui/mcp browser` or any non-Playwright path.
     - do not wait for playwright page to display
     - do not check the current state or take a snapshot
   - Otherwise, run `.ui/mcp browser` (do NOT use mark-playwright.html here)
   - `.ui/mcp status` to verify sessions > 0 (confirms browser connected)

3. IMMEDIATELY start the event loop:
   .ui/mcp event

   The UI will NOT respond to clicks until the event loop is running!

4. When events arrive, handle according to design.md, then restart the loop:
   - Parse JSON: [{"app":"app-console","event":"select","name":"contacts"}]
   - Check design.md's "Events" section for how to handle each event type
   - **For UI changes**: Check event's `handler` field and invoke that skill (see Handler Dispatch Rule above)
   - Simple state changes: use `.ui/mcp run` directly
   - Some events require spawning a background agent (see Build Settings below)
   - Restart: `.ui/mcp event`
```

**Background Agent Events:**
Some events require spawning a background Task agent. The design.md specifies this explicitly. Look for patterns like "spawn background agent" or `Task(...)` in the design.md's event handling section. Background agents allow the event loop to continue while long-running work (builds, tests) executes.

**The event loop is NOT optional.** Without it, button clicks and form submissions are silently ignored.

## Build Settings

Every event has two fields injected automatically by the MCP layer based on the status bar toggles:

| Toggle | Field | Values |
|--------|-------|--------|
| Build mode (rocket/diamond) | `handler` | `"/ui-fast"` or `"/ui-thorough"` |
| Execution (hourglass/arrows) | `background` | `false` or `true` |

### Handler Dispatch (MANDATORY)

**DO NOT skip this.** When handling events that involve UI changes:

1. **Invoke the skill named in `handler`** — use `Skill(skill: "ui-fast")` or `Skill(skill: "ui-thorough")`
2. **Check `background`** — if true, run as background agent via Task tool

**Always respect these fields.** They reflect the user's explicit choices via the UI toggles. Ignoring the handler means ignoring user preferences.

### Execution Mode

| `background` | Behavior |
|--------------|----------|
| `false` | Run in foreground (blocks event loop) |
| `true` | Run as background agent (event loop continues) |

For background execution:
```
Task(subagent_type="ui-builder", run_in_background=true, prompt="invoke {handler} skill...")
```

### `/ui-fast`

Rapid iteration with checkpointing:
1. Checkpoint current state before changes
2. Make edits directly
3. Hot-reload shows results immediately
4. User can rollback if needed

### `/ui-thorough`

Full workflow with progress feedback:
- Requirements → Design → Code → Viewdefs → Audit → Simplify
- Shows step-by-step progress in UI

## App Variable Convention

Each app defines a **global variable** for interacting with it via `.ui/mcp run`:

| App Name       | Global Variable | Example Call                        |
|----------------|-----------------|-------------------------------------|
| `claude-panel` | `claudePanel`   | `claudePanel:addAgentMessage("Hi")` |
| `contacts`     | `contacts`      | `contacts:addContact(name, email)`  |
| `ma-luba`      | `maLuba`        | `maLuba:someMethod()`               |

**Convention:** kebab-case app name → camelCase variable. The global variable is exactly the camelCase conversion of the app directory name (no "App" suffix).

Find the variable by looking at the bottom of `app.lua` for the instance creation:
```lua
if not session.reloading then
    claudePanel = ClaudePanel:new()  -- <-- this is the global variable
end
```

## Server Lifecycle

The server auto-starts when the MCP connection is established. Use `.ui/mcp status` to get `sessions` count and `url`.

**URL:** Always use `{url}/?conserve=true` to access the UI. The `conserve` parameter prevents duplicate browser tabs. **NEVER include session IDs in URLs** — they will cause problems. The server binds the root URL to the MCP session automatically via a cookie.

**Verifying connection:** After navigating with Playwright, call `.ui/mcp status` and check that `sessions > 0`. This confirms the browser connected without needing artificial waits.

## Building or Modifying UIs

**Always use `/ui-fast` or `/ui-thorough`** — never edit UI files directly from the `/ui` skill. When handling events, check the `handler` field to determine which skill to invoke. When the user asks for changes outside an event loop, ask which mode they prefer or default to `/ui-fast` for small changes.

**Before building a new app:**
1. Create the app directory: `mkdir -p {base_dir}/apps/<app>`
2. Write requirements to `{base_dir}/apps/<app>/requirements.md`

### Requirements Format

```markdown
# Descriptive Title

A short paragraph describing what the app does.

## Section 1
...
```

The first line is a descriptive title (e.g., "# Contact Manager"), followed by prose describing the app. See `.claude/skills/ui-builder/examples/requirements.md` for a reference.

## Directory Structure

```
{base_dir}/
├── apps/<app>/           # App source files
│   ├── requirements.md   # What to build (you write this)
│   ├── design.md         # How it works (generated)
│   ├── app.lua           # Lua code (generated)
│   └── viewdefs/         # HTML templates (generated)
├── lua/                  # Symlinks to app lua files
├── viewdefs/             # Symlinks to app viewdefs
├── event                 # Event wait script
└── log/                  # Runtime logs
```

## File Ownership

- `requirements.md` — you write/update this
- `design.md`, `app.lua`, `viewdefs/` — **MUST use `/ui-fast` or `/ui-thorough` skill** (never edit directly from `/ui`)

## Debugging

- Check `{base_dir}/log/lua.log` for Lua errors
- `ui://variables` resource shows full variable tree with IDs, parents, types, values
- `.ui/mcp run` returns error messages
- `ui://state` resource shows live state JSON
- `window.uiApp` contains the app object in the browser
  - `window.uiApp.store` shows all variables
