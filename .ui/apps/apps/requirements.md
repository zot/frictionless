# Apps

Dashboard for discovering, launching, and tracking quality of ui-mcp apps. Acts as a command center for UI development with Claude.

## Architecture

**Lua-driven:** All app discovery, parsing, and state management happens in Lua. Claude only pushes state changes when it modifies something.

**App discovery:** Lua uses `mcp:status()` to get `base_dir`, then scans `{base_dir}/apps/*/requirements.md` to find all apps.

## App List

Display all apps found in the apps directory. For each app show:
- App name (directory name)
- Status: test count ("17/21"), "not built", "--" (no tests), or build progress
- Visual indicator (green if all passing, yellow if partial, gray if not built/no tests)
- Build progress shows hover help with the current phase name (e.g., "designing", "writing code")

A "+" button in the header opens the new app form.

Clicking an app selects it and shows details in the adjacent panel.

## App Details

When an app is selected, show:
- App name as header
- Description (first paragraph from requirements.md, parsed by Lua)
- Build progress and phase (when app is building) - shows progress bar and stage label
- Action buttons based on state:
  - Build (when no viewdefs) - sends build_request to Claude
  - Open (when has viewdefs) - opens the app in the embedded app view (disabled for "apps" and "mcp")
  - Test (when has app.lua)
  - Fix Issues (when has known issues)
- Test checklist from TESTING.md with checkboxes (read-only, parsed by Lua)
- Known Issues section (expandable)
- Fixed Issues section (collapsed by default)

## Embedded App View

When the "Open" button is clicked, the selected app replaces the detail panel (right side):
- App list remains visible on the left
- The embedded view displays `embeddedValue` directly (not an iframe)
- Header shows app name and close button `[X]`
- The chat panel remains visible below
- User can interact with the embedded app while still chatting with Claude

Clicking the close button `[X]` closes the embedded view and restores the normal detail panel.

## New App Form

When "+" is clicked, show a form instead of details:
- Name field (becomes directory name, kebab-case)
- Description textarea (what the app should do)
- Create button (Lua creates app directory and requirements.md, selects the new app)
- Cancel button (returns to app details or empty state)

**On Create (Lua):**
1. Create directory `{base_dir}/apps/{name}/`
2. Write `requirements.md` with title and description
3. Rescan to add app to list
4. Select the new app (shows Build button since no viewdefs)
5. Send `app_created` event to Claude with the app name and description

**On Create (Claude):**
When Claude receives the `app_created` event, it should:
1. Read the basic requirements.md that Lua created
2. Flesh out the requirements with proper structure and detail based on the description
3. Write the expanded requirements.md to disk
4. Use `ui_run` to call `apps:updateRequirements(name, content)` to populate the requirements textbox in the UI

## Chat Panel

Always visible at the bottom. User can chat with Claude about the selected app:
- Ask questions about the app
- Request actions (test, build, fix)
- General development discussion

Selected app provides context for the conversation.

**Layout:**
- Chat area should be vertically resizable (drag handle at top edge)
- Messages auto-scroll to bottom on new output
- Messages display in reverse order (newest at bottom, like Claude Code terminal)
- User messages prefixed with `>` character

**Quality Selector:**

A 3-position slider next to the Send button controls how modification requests are handled:

| Mode | Behavior |
|------|----------|
| Fast | Vibe code - just make the change directly, no skill, no phases |
| Thorough | Use ui-builder skill with full phases (design, audit, etc.) |
| Background | Spawn background agent (shows progress bar, non-blocking) |

Default is Fast for quickest feedback. User can switch to higher quality modes when needed.

## Build Progress

When Claude is building an app, Lua tracks progress state:
- Progress bar (0-100%)
- Stage label (designing, writing code, creating viewdefs, linking)

Claude pushes progress updates via `ui_run` calling `apps:onAppProgress()` when building.

## Events to Claude

Events are sent via `mcp.pushState()` and include `app` (the app name) and `event` (the event type).

**Note field:** Lua includes a `note` field in each event reminding Claude to understand the target app: `"note": "make sure you have understood the app at {base_dir}/apps/{APP}"`. Claude should read the app's requirements.md and design.md before taking action.

### `chat`
User message with selected app as context. Respond conversationally.

**If the chat involves modifying an app:** Check the `quality` field:
- `fast` — Read app files at `{base_dir}/apps/{context}/`, make the change directly, reply via `apps:addAgentMessage()`
- `thorough` — Use `/ui-builder` skill inline with full phases
- `background` — Spawn background ui-builder agent (shows progress bar)

### `build_request`
Build, complete, or update an app. **Spawn a background ui-builder agent** to handle this.

**Event payload:** `{app: "apps", event: "build_request", target: "my-app", mcp_port: 37067}`

Lua includes `mcp_port` from `mcp:status()` so Claude can spawn the agent directly:
```
Task(subagent_type="ui-builder", run_in_background=true, prompt="MCP port is {mcp_port}. Build the {target} app at .ui/apps/{target}/")
```

Before spawning the agent, use `ui_run` to update app progress with (APP, 0%, "thinking...")

Tell the ui-builder agent:
- Use the HTTP API (curl) since background agents don't have MCP tool access
- Report progress via `curl -s -X POST http://127.0.0.1:{mcp_port}/api/ui_run -d 'mcp:appProgress("{name}", {progress}, "{stage}")'`
- Call `mcp:appUpdated("{name}")` when done (triggers rescan)

**Why background?** Building takes time. A background agent lets Claude continue responding to chat while the build runs. The progress bar shows real-time status.

### `test_request`
Run ui-testing on an app. Can also use a background agent pattern.

### `fix_request`
Fix known issues in an app. Can also use a background agent pattern.

## Data Flow

### Lua Responsibilities:

**On load and refresh:**
1. Call `mcp:status()` to get `base_dir`
2. Scan `{base_dir}/apps/` for directories with `requirements.md`
3. For each app, parse:
   - `requirements.md` → name, description
   - `viewdefs/` presence → built status (has viewdefs = can be opened)
   - `TESTING.md` → test counts, checklist, issues

**On app creation:**
1. Create `{base_dir}/apps/{name}/` directory
2. Write `requirements.md` with `# {Name}` title and description
3. Rescan to add app to list
4. Select the new app (user can click Build to trigger build_request)

**Embedded app view:**
- Track `embeddedApp` state (name of currently embedded app, or nil)
- Track `embeddedValue` state (the app global loaded via `mcp:app`)
- `openEmbedded(name)` - call `mcp:app(name)`, if not nil store in `embeddedValue` and set `embeddedApp`
- `closeEmbedded()` - clear `embeddedApp` and `embeddedValue`, restore app list + details view

### Claude Responsibilities (via mcp methods):
- `mcp:appProgress(name, progress, stage)` - call during build to update progress
- `mcp:appUpdated(name)` - call after modifying app files to trigger rescan

Claude uses `ui_run` to call these mcp methods.

### App Initialization (`init.lua`)

The apps app provides `init.lua` which adds convenience methods to the `mcp` global:

```lua
function mcp:appProgress(name, progress, stage)
    if apps then apps:onAppProgress(name, progress, stage) end
end

function mcp:appUpdated(name)
    mcp:scanAvailableApps()  -- rescan all apps from disk
    if apps then apps:onAppUpdated(name) end
end
```

This allows Claude to call `mcp:appProgress()` and `mcp:appUpdated()` without needing to check if the apps dashboard is loaded. The `mcp:scanAvailableApps()` call ensures the MCP server's app list stays in sync with disk.

## Refresh

A refresh button triggers Lua to rescan all apps and update the display. The refresh also calls `mcp:scanAvailableApps()` to keep the MCP server's app list in sync with disk.

## File Parsing (Lua)

**requirements.md:**
- First paragraph (text before first blank line) = description

**TESTING.md:**
- `- [ ]` = untested
- `- [✓]` = passed
- `- [✗]` = failed
- Status shows "passed/total" (e.g., "17/21")
- `### N.` under "Known Issues" = open bugs
- `### N.` under "Fixed Issues" = resolved bugs
