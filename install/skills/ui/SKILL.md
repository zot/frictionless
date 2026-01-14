---
name: ui
description: use when **running ui-mcp UIs** or needing to understand UI structure (before modifying one)
---

# UI MCP

Foundation for building and running ui-engine UIs with Lua apps connected to widgets.

## Quick Start: Show an Existing App

To display an app (e.g., `claude-panel`):

```
1. Read design.md to learn:
   - The global variable name (e.g., `claudePanel`)
   - Event types and how to respond

2. Ensure server is running (do all steps in parallel):
   - ui_status → if not "running", call ui_configure + ui_start
   - ui_display("app-name")
   - ui_open_browser (or navigate with Playwright)

3. Start event loop (foreground is most responsive):
   .claude/ui/event

4. When events arrive, handle via ui_run using the global variable:
   claudePanel:addAgentMessage("response")
```

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

Always ensure the server is running before displaying UIs:

```
ui_status → state?
  "unconfigured" → ui_configure(".claude/ui") → ui_start()
  "configured"   → ui_start()
  "running"      → ready!
```

**Shortcut:** Call `ui_status`, `ui_configure`, and `ui_start` as needed, then `ui_display` and `ui_open_browser` - all can be done in quick succession.

## Event Loop

The event script waits for user interactions and returns JSON:

```bash
.claude/ui/event
```

Returns one JSON array per line containing one or more events:
```json
[{"app":"claude-panel","event":"chat","text":"Hello"},{"app":"claude-panel","event":"action","action":"commit"}]
```

**Foreground event loop (recommended):**
1. Run `.claude/ui/event` with Bash (blocking, ~2 min timeout)
2. When it returns, parse JSON array. If non-empty, handle each event
   - use `ui_run` to alter app state and reflect changes to the user
3. Restart the event loop

This is the most responsive approach - events are handled immediately.

**Background event loop (alternative):**
Run `.claude/ui/event` in background if you need to do other work while waiting. Note: this adds latency since you must poll the output file.

**Exit codes:**
- 0 + empty output = timeout, no events (just restart)
- 0 + JSON output = events received
- 52 = server restarted (restart both server and event loop)

## Building or modifying UIs

**ALWAYS use the `/ui-builder` skill to create or modify UIs.** Do NOT use `ui_*` MCP tools directly for building.

Before invoking `/ui-builder`:
1. Create the app directory: `mkdir -p .claude/ui/apps/<app>`
2. Write requirements to `.claude/ui/apps/<app>/requirements.md`
3. Invoke `/ui-builder`: "Read `.claude/ui/apps/<app>/requirements.md` and build the app"

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
.claude/ui/
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

- Check `.claude/ui/log/lua.log` for Lua errors
- `ui://variables` resource shows full variable tree with IDs, parents, types, values
- `ui_run` returns error messages
- `ui://state` resource shows live state JSON
- `window.uiApp` contains the app object in the browser
  - `window.uiApp.store` shows all variables
