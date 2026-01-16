---
name: ui
description: use when **running ui-mcp UIs** or needing to understand UI structure (before modifying one)
---

# UI MCP

Foundation for building and running ui-engine UIs with Lua apps connected to widgets.

## Getting base_dir and url

Always get `base_dir` and `url` from `ui_status` first. All paths below use `{base_dir}` as a placeholder. Use `{url}` exactly as returned (e.g., `http://127.0.0.1:34919`).

## Simple Requests

When the user says `show APP` as in Quick Start.

When the user says `events` it means simply to start the event loop, but not use `ui_display` or `ui_open_browser` as in Quick Start

## Quick Start: Show an Existing App

To display an app (e.g., `claude-panel`):

```
1. Read {base_dir}/apps/APP/design.md to learn:
   - The global variable name (e.g., `claudePanel`)
   - Event types and how to respond

2. Display and open:
   - ui_display("app-name")
   - ui_open_browser (or navigate to {url}/?conserve=true with Playwright)

3. IMMEDIATELY start the event loop:
   {base_dir}/event

   The UI will NOT respond to clicks until the event loop is running!

4. When events arrive, handle according to design.md, then restart the loop:
   - Parse JSON: [{"app":"apps","event":"select","name":"contacts"}]
   - Check design.md's "Events" section for how to handle each event type
   - Some events use `ui_run` directly: `ui_run('contacts:doSomething()')`
   - Some events require spawning a background agent (see below)
   - Restart: {base_dir}/event
```

**Background Agent Events:**
Some events require spawning a background Task agent. The design.md specifies this explicitly:
```
Task(subagent_type="ui-builder", run_in_background=true, prompt="Build the {target} app")
```
Look for patterns like "spawn background agent" or `Task(...)` in the design.md's event handling section. Background agents allow the event loop to continue while long-running work (builds, tests) executes.

**The event loop is NOT optional.** Without it, button clicks and form submissions are silently ignored.

## App Variable Convention

Each app defines a **global variable** for interacting with it via `ui_run`:

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

The server auto-starts when the MCP connection is established. Use `ui_status` to get `base_dir`, `url`, and `sessions` count.

**URL:** Always use `{url}/?conserve=true` to access the UI. The `conserve` parameter prevents duplicate browser tabs. The server binds the root URL to the MCP session automatically via a cookie - no session ID needed in the URL.

If you need to reconfigure (different base_dir), call `ui_configure({base_dir})` - this stops the current server and restarts with the new directory.

## Event Loop

The event script waits for user interactions and returns JSON:

```bash
{base_dir}/event
```

Returns one JSON array per line containing one or more events:
```json
[{"app":"claude-panel","event":"chat","text":"Hello"},{"app":"claude-panel","event":"action","action":"commit"}]
```

**Foreground event loop (recommended):**
1. Run `{base_dir}/event` with Bash (blocking, ~2 min timeout)
2. When it returns, parse JSON array. If non-empty, handle each event
   - use `ui_run` to alter app state and reflect changes to the user
3. Restart the event loop

This is the most responsive approach - events are handled immediately.

**Background event loop (alternative):**
Run `{base_dir}/event` in background if you need to do other work while waiting. Note: this adds latency since you must poll the output file.

**Exit codes:**
- 0 + empty output = timeout, no events (just restart)
- 0 + JSON output = events received
- 52 = server restarted (restart both server and event loop)

## Building or modifying UIs

**ALWAYS use the `/ui-builder` skill to create or modify UIs.** Do NOT use `ui_*` MCP tools directly for building.

Before invoking `/ui-builder`:
1. Create the app directory: `mkdir -p {base_dir}/apps/<app>`
2. Write requirements to `{base_dir}/apps/<app>/requirements.md`
3. Invoke `/ui-builder`: "Read `{base_dir}/apps/<app>/requirements.md` and build the app"

### Requirements Format

```markdown
# Descriptive Title

A short paragraph describing what the app does.

## Section 1
...
```

The first line is a descriptive title (e.g., "# Contact Manager"), followed by prose describing the app. See the `/ui-builder` skill's `examples/requirements.md` for a reference.

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
- `design.md`, `app.lua`, `viewdefs/` — use `/ui-builder` skill to modify

## Debugging

- Check `{base_dir}/log/lua.log` for Lua errors
- `ui://variables` resource shows full variable tree with IDs, parents, types, values
- `ui_run` returns error messages
- `ui://state` resource shows live state JSON
- `window.uiApp` contains the app object in the browser
  - `window.uiApp.store` shows all variables
