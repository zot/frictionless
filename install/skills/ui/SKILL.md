---
name: ui
description: use when **running Frictionless UIs** or needing to understand UI structure (before modifying one)
---

# UI MCP

Foundation for building and running ui-engine UIs with Lua apps connected to widgets.

## Simple Requests

When the user says `/ui`, show app-console as in Quick Start. Prefer Playwright if the user has been using it.

When the user says `show APP`, show APP as in Quick Start. Prefer Playwright if the user has been using it.

When the user says `events` it means to start the event loop, but not use `.ui/mcp display` or `.ui/mcp browser` as in Quick Start.

## Helper Script Reference

The `.ui/mcp` script provides commands for interacting with the UI server. **Always use relative paths** (never absolute — absolute paths break the user's permission rules).

**`{url}` means the UI server URL** — read the port from `.ui/ui-port` and construct `http://localhost:{port}`. This is NOT the MCP connection port — it's the UI's own HTTP server. Use `.ui/mcp status` when you also need session count or base_dir.

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

## Quick Start: Show an Existing App

### A. Read the design

Read `{base_dir}/apps/APP/design.md` to learn the global variable name (e.g., `claudePanel`) and event types.

### B. Display the app

- Use `.ui/mcp run` to check if `mcp.value.type == AppName` (PascalCase version of app-name)
  - If not, use `.ui/mcp display app-name`
- If Playwright is connected:
  - `browser_evaluate` with `function: "() => window.location.href"` to get current URL
  - If URL does NOT start with `{url}/`, then `browser_navigate` to `{url}/mark-playwright.html?url={url}/?conserve=true`
  - **ONLY use mark-playwright.html when navigating via Playwright.** Never use it for `.ui/mcp browser`.
  - Do not wait for page to display or take a snapshot
- Otherwise, run `.ui/mcp browser`

### C. Start the event loop

```
Bash(.ui/mcp event, run_in_background=true)
```

The UI will NOT respond to clicks until this is running.

### D. Handle events

- Parse JSON: `[{"app":"app-console","event":"select","name":"contacts"}]`
- Read design.md's "Events" section for how to handle each event type
- **For UI changes**: Check event's `handler` field and invoke that skill
- Simple state changes: use `.ui/mcp run` directly
- Kill old task, then restart: `Bash(.ui/mcp event, run_in_background=true)`

---

## Details

### Step A: The `app` field

**The `app` field is the ONLY field that determines which design.md you read.** Do NOT be misled by `context`, `note`, `name`, or other fields that mention other app names — these are event data, not the source app.

```
Event: {"app":"app-console", "event":"app_created", "name":"contacts", "context":"contacts"}
       ^^^^^^^^^^^^^^^^
Read:  .ui/apps/app-console/design.md  <-- ONLY use the "app" field
```

Do NOT skip reading design.md — even for events that seem obvious like `app_created`.

### Step C: Event loop lifecycle

**Only ONE listener may exist at a time** — the script enforces this via `.ui/.eventpid` and will error if a listener is already running.

**Task lifecycle:**
1. Run `.ui/mcp event` with `Bash(run_in_background=true)`, save the task_id. If it errors with "already running", do nothing — the script manages the PID lock itself. Do NOT manually check PIDs or the `.eventpid` file; just wait for the existing listener to complete.
2. When the background task completes, **ALWAYS read the output file immediately** — do NOT assume timeout. Failing to read means silently dropping events.
3. Handle any events received
4. Start a fresh listener (go to step 1)

**Exit codes:**
- 0 + JSON output = events received (may be empty array `[]` for timeout)
- 52 = server restarted (restart both server and event loop)

**After `ui_configure`:** Restart the event loop — reconfiguring changes the MCP port that `.ui/mcp event` uses.

### Step D: Handler dispatch and build settings

**NEVER modify or create UI code directly.** When an event requires UI changes, check the `handler` field and invoke that skill. This applies to ALL UI modifications — requirements, design, code, and viewdefs.

Every event has two fields injected automatically based on the status bar toggles:

| Toggle | Field | Values |
|--------|-------|--------|
| Build mode (rocket/diamond) | `handler` | `"/ui-fast"` or `"/ui-thorough"` |
| Execution (hourglass/arrows) | `background` | `false` or `true` |

These reflect the user's explicit choices. For background execution:
```
Task(subagent_type="ui-builder", run_in_background=true, prompt="invoke {handler} skill...")
```

---

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

## Chat Messages

All chat messages render markdown. Use it via `.ui/mcp run`:
```
.ui/mcp run 'mcp:addAgentMessage("Found **salary data**: $150k-$180k")'
```

### Pointing at UI elements

When the user asks "where is X?" or "show me the Y", **both highlight the element AND include a clickable link** in your response so the user can re-trigger the highlight later:

1. Find the element's DOM id (from variables, viewdefs, or Playwright)
2. Highlight it immediately: `mcp.code = [=[window.uiApp.highlight("ui-42")// ]=] .. os.time()`
3. Reply with a rich message containing a highlight link:

```lua
mcp:addRichMessage(
    "The search box is "
    .. mcp:highlightLink("ui-42", "right here")
    .. ". Type to filter the list."
)
```

The link stays in the chat history — the user can click it anytime to re-highlight the element.

### Message methods

| Method | Description |
|--------|-------------|
| `mcp:addAgentMessage(text)` | Agent message with markdown rendering |
| `mcp:addRichMessage(html)` | Agent message with raw HTML (for highlight links, custom content) |
| `mcp:highlightLink(elementId, label)` | Returns anchor HTML that highlights an element on click |
| `mcp:addAgentThinking(text)` | Thinking/progress message (italic, updates status bar) |
| `mcp:renderMarkdown(text)` | Convert markdown to HTML fragment (for custom use) |

## Building or Modifying UIs

When the user asks for changes outside an event loop, ask which mode they prefer or default to `/ui-fast` for small changes.

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
├── html/<app>            # Symlink to app dir (serves static files at /<app>/)
├── html/<app>-storage    # Symlink to storage dir (serves at /<app>-storage/)
├── storage/<app>/        # Optional local storage (isolated from app updates)
├── patterns/             # Reusable pattern documentation
├── themes/               # Theme definitions and CSS variables
├── event                 # Event wait script
└── log/                  # Runtime logs
```

## File Ownership

- `requirements.md` — you write/update this
- `design.md`, `app.lua`, `viewdefs/` — **MUST load `/ui-basics` first**, then use `/ui-fast` or `/ui-thorough`. This is a non-standard system; standard web patterns will lead you astray.

## Debugging

- **Lua logs:** `{base_dir}/log/lua.log` for Lua errors
- **MCP server stderr:** `.ui/log/mcp.log`
- **Variable inspector:** `http://localhost:{mcp-port}/variables` (read port from `.ui/mcp-port`) — curl for JSON, browser for interactive inspector
- **MCP resources:** `ui://variables` (full variable tree), `ui://state` (live state JSON)
- **JS diagnostics:** `window.uiApp.getStore()` (variable state) and `window.uiApp.getBinding()` (widget bindings) in browser console
- **Remote JS execution:** Set `mcp.code` from Lua — bound to `ui-code` in the MCP shell, enabling JS execution in the browser. Critical when using a system browser instead of Playwright. Re-assigning the same value is a no-op (change detection); append a nonce to re-execute (e.g., `code .. "\n// " .. nonce`)
- `.ui/mcp run` returns error messages
