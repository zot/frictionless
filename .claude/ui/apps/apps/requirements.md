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

A "+" button in the header opens the new app form.

Clicking an app selects it and shows details in the adjacent panel.

## App Details

When an app is selected, show:
- App name as header
- Description (first paragraph from requirements.md, parsed by Lua)
- Action buttons based on state:
  - Build (when no viewdefs) - sends build_request to Claude
  - Open (when has viewdefs) - uses `mcp.display(appName)` directly to switch apps
  - Test (when has app.lua)
  - Fix Issues (when has known issues)
- Test checklist from TESTING.md with checkboxes (read-only, parsed by Lua)
- Known Issues section (expandable)
- Fixed Issues section (collapsed by default)

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

## Build Progress

When Claude is building an app, Lua tracks progress state:
- Progress bar (0-100%)
- Stage label (designing, writing code, creating viewdefs, linking)

Claude pushes progress updates via `ui_run` calling `apps:onAppProgress()` when building.

## Events to Claude

- `chat` - User message with selected app as context
- `build_request` - Use `/ui-builder` skill to build, complete, or update the app
- `test_request` - Run ui-testing on an app
- `fix_request` - Fix known issues in an app

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
    if apps then apps:onAppUpdated(name) end
end
```

This allows Claude to call `mcp:appProgress()` and `mcp:appUpdated()` without needing to check if the apps dashboard is loaded.

## Refresh

A refresh button triggers Lua to rescan all apps and update the display.

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
