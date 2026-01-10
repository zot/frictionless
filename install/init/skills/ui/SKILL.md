---
name: ui
description: use when **running ui-mcp UIs** or needing to understand UI structure
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

3. Start event loop in background:
   .claude/ui/event

4. When events arrive, handle via ui_run using the global variable:
   claudePanel:addAgentMessage("response")
```

## App Variable Convention

Each app defines a **global variable** for interacting with it via `ui_run`:

| App Name       | Global Variable | Example Call                          |
|----------------|-----------------|---------------------------------------|
| `claude-panel` | `claudePanel`   | `claudePanel:addAgentMessage("Hi")`   |
| `contacts`     | `contactsApp`   | `contactsApp:addContact(name, email)` |
| `my-app`       | `myApp`         | `myApp:someMethod()`                  |

**Convention:** kebab-case app name → camelCase variable (check `app.lua` to confirm).

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

**Event loop pattern:**
1. Start `.claude/ui/event` in background
2. When it completes, read output file
3. If output is empty, just restart the loop (timeout with no events)
4. Otherwise parse JSON array, handle each event with `ui_run`
5. Restart the event loop

**Exit codes:**
- 0 + empty output = timeout, no events (just restart)
- 0 + JSON output = events received
- 52 = server restarted (restart both server and event loop)

## Building UIs

**ALWAYS use the `/ui-builder` skill to create or modify UIs.** Do NOT use `ui_*` MCP tools directly for building.

Before invoking `/ui-builder`:
1. Create the app directory: `mkdir -p .claude/ui/apps/<app>`
2. Write requirements to `.claude/ui/apps/<app>/requirements.md`
3. Invoke `/ui-builder`: "Read `.claude/ui/apps/<app>/requirements.md` and build the app"

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
- `ui_run` returns error messages
- `ui://state` resource shows live state JSON
- Browser console: `window.uiApp.store` shows all variables

## Resources

| Resource         | Content                          |
|------------------|----------------------------------|
| `ui://reference` | Quick start and overview         |
| `ui://lua`       | Lua API (session, mcp, etc.)     |
| `ui://viewdefs`  | Viewdef syntax and bindings      |
| `ui://state`     | Live state JSON (for debugging)  |
