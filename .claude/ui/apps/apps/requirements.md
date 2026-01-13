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
  - Build (when no app.lua)
  - Open (when has app.lua)
  - Test (when has app.lua)
  - Fix Issues (when has known issues)
- Test checklist from TESTING.md with checkboxes (read-only, parsed by Lua)
- Known Issues section (expandable)
- Fixed Issues section (collapsed by default)

## New App Form

When "+" is clicked, show a form instead of details:
- Name field (becomes directory name, kebab-case)
- Description textarea (what the app should do)
- Create button (sends to Claude to build)
- Cancel button (returns to app details or empty state)

## Chat Panel

Always visible at the bottom. User can chat with Claude about the selected app:
- Ask questions about the app
- Request actions (test, build, fix)
- General development discussion

Selected app provides context for the conversation.

## Build Progress

When Claude is building an app, Lua tracks progress state:
- Progress bar (0-100%)
- Stage label (designing, writing code, creating viewdefs, linking)

Claude pushes progress updates via `mcp.pushState()` when building.

## Events to Claude

- `chat` - User message with selected app as context
- `build_request` - Build an unbuilt app
- `test_request` - Run ui-testing on an app
- `fix_request` - Fix known issues in an app
- `create_app` - Create new app with name and description

## Data Flow

### Lua Responsibilities (on load and refresh):
1. Call `mcp:status()` to get `base_dir`
2. Scan `{base_dir}/apps/` for directories with `requirements.md`
3. For each app, parse:
   - `requirements.md` → name, description
   - `app.lua` presence → built status
   - `TESTING.md` → test counts, checklist, issues

### Claude Responsibilities (push only):
- `mcp.pushState({type="app_progress", app=name, progress=N, stage=S})` during build
- `mcp.pushState({type="app_updated", app=name})` after any file changes
- `mcp.pushState({type="app_created", app=name})` after creating new app

Lua listens for these events and re-parses affected apps.

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
