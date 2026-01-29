---
name: ui
description: use when **running Frictionless UIs** or needing to understand UI structure (before modifying one)
---

# UI MCP

Foundation for building and running ui-engine UIs with Lua apps connected to widgets.

## Getting base_dir and url

Always get `base_dir` and `url` from `ui_status` first. All paths below use `{base_dir}` as a placeholder. Use `{url}` exactly as returned (e.g., `http://127.0.0.1:34919`).

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

```bash
.ui/mcp event
```

Returns one JSON array per line containing one or more events:
```json
[{"app":"claude-panel","event":"chat","text":"Hello"},{"app":"claude-panel","event":"action","action":"commit"}]
```

Make sure you have read the design file for the event.

### Which Design File to Read

**Example event:**
```json
{"app":"app-console", "context":"contacts", "note":"...contacts", "event":"chat", "text":"hello"}
```

**Read:** `{base_dir}/apps/app-console/design.md` (from `app` field)
**NOT:** `{base_dir}/apps/contacts/design.md` (ignore `context` and `note`)

The `app` field identifies which app's design.md to read. Other fields like `context` or `note` provide data for event handling but do NOT change which design file you read.

You must not skip reading that app's design unless you have already read it in this conversation.

### Running in Foreground or Background

**Foreground event loop (recommended):**
1. Run `.ui/mcp event` with Bash (blocking, ~2 min timeout)
2. When it returns, parse JSON array. If non-empty, handle each event
   - use `.ui/mcp run` to alter app state and reflect changes to the user
3. Restart the event loop

This is the most responsive approach - events are handled immediately.

**Background event loop (alternative):**
Run `.ui/mcp event` in background if you need to do other work while waiting. Note: this adds latency since you must poll the output file.

**Exit codes:**
- 0 + empty output = timeout, no events (just restart)
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
     - if URL does NOT start with {url}/, then browser_navigate to {url}/?conserve=true
     - do not wait for playwright page to display
     - do not check the current state or take a snapshot
   - Otherwise, run `.ui/mcp browser`
   - `.ui/mcp status` to verify sessions > 0 (confirms browser connected)

3. IMMEDIATELY start the event loop:
   .ui/mcp event

   The UI will NOT respond to clicks until the event loop is running!

4. When events arrive, handle according to design.md, then restart the loop:
   - Parse JSON: [{"app":"app-console","event":"select","name":"contacts"}]
   - Check design.md's "Events" section for how to handle each event type
   - Some events use `.ui/mcp run` directly: `.ui/mcp run 'contacts:doSomething()'`
   - Some events require spawning a background agent (see below)
   - Restart: `.ui/mcp event`
```

**Background Agent Events:**
Some events require spawning a background Task agent. The design.md specifies this explicitly:
```
Task(subagent_type="ui-builder", run_in_background=true, prompt="Build the {target} app")
```
Look for patterns like "spawn background agent" or `Task(...)` in the design.md's event handling section. Background agents allow the event loop to continue while long-running work (builds, tests) executes.

**The event loop is NOT optional.** Without it, button clicks and form submissions are silently ignored.

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

The server auto-starts when the MCP connection is established. Use `ui_status` to get `base_dir`, `url`, and `sessions` count.

**URL:** Always use `{url}/?conserve=true` to access the UI. The `conserve` parameter prevents duplicate browser tabs. **NEVER include session IDs in URLs** — they will cause problems. The server binds the root URL to the MCP session automatically via a cookie.

**Verifying connection:** After navigating with Playwright, call `.ui/mcp status` and check that `sessions > 0`. This confirms the browser connected without needing artificial waits.

If you need to reconfigure (different base_dir), run `.ui/mcp configure {base_dir}` - this stops the current server and restarts with the new directory.

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
- `.ui/mcp run` returns error messages
- `ui://state` resource shows live state JSON
- `window.uiApp` contains the app object in the browser
  - `window.uiApp.store` shows all variables
