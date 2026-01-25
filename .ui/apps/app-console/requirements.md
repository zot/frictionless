# App Console

Dashboard for discovering, launching, and tracking quality of frictionless apps. Acts as a command center for UI development with Claude.

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
  - Build (when no viewdefs) - sets progress to `0, "pondering"` then sends build_request to Claude
  - Open (when has viewdefs) - opens the app in the embedded app view (disabled for "app-console" and "mcp")
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
5. Start progress at "pondering, 0%" via `mcp:appProgress(name, 0, "pondering")`
6. Send `app_created` event to Claude with the app name and description

**On Create (Claude):**
When Claude receives the `app_created` event, it should show both progress (in the app list) and thinking messages (in the chat panel) while processing:

| Step | Progress | Thinking Message                  | Action                                                                        |
|------|----------|-----------------------------------|-------------------------------------------------------------------------------|
| 1    | 33%       | "Reading initial requirements..." | Read the basic requirements.md that Lua created                               |
| 2    | 66%      | "Fleshing out requirements..."    | Expand requirements with proper structure and detail based on the description |
| 3    | (clear)  | (final message)                   | Write the expanded requirements.md to disk, clear progress                    |

Use `mcp:appProgress(name, percent, stage)` for the progress bar, `appConsole:addAgentThinking(text)` for chat panel updates, then `appConsole:addAgentMessage(text)` for the final response. Call `appConsole:updateRequirements(name, content)` to populate the requirements in the UI, then `mcp:appProgress(name, nil, nil)` to clear the progress bar.

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

**IMPORTANT:** Claude MUST always respect the quality field. When `quality` is "thorough", Claude MUST invoke the `/ui-builder` skill and follow its full workflow (update design.md, audit, etc.). When `quality` is "fast", Claude can make direct edits without the skill. This is not optional - the quality setting is the user's explicit choice about how they want changes handled.

## Claude Code Todo List

The bottom panel displays Claude's current todo list alongside the Chat/Lua panel:

**Layout:**
- Bottom resizable section has two columns: Todo List (left) | Chat/Lua (right)
- Todo list is a narrow column (200px) showing current task status
- Todo items don't wrap; column scrolls horizontally if text overflows
- Collapse button hides the column horizontally (shrinks to icon-only 32px width)

**Display:**
- Show todo items with status indicators:
  - ⏳ pending (gray)
  - 🔄 in_progress (blue, highlighted)
  - ✓ completed (green, muted)
- The in_progress item is shown prominently at the top
- Completed items can be collapsed/hidden

**MCP Todo API (in init.lua):**

Two methods simplify todo management for background agents:

```lua
-- Create todo list at the start of a build
mcp:createTodos({'Read requirements', 'Requirements', 'Design', 'Write code', 'Write viewdefs', 'Link and audit', 'Simplify'})

-- Advance to step n (completes previous step, starts step n)
mcp:startTodoStep(2)  -- starts "Requirements", completes "Read requirements"
```

**`mcp:createTodos(steps)`** - Creates todo items from an array of step labels. Uses hardcoded progress percentages for ui-builder workflow:

| Step | Label | Progress | Thinking |
|------|-------|----------|----------|
| 1 | Read requirements | 5% | "Reading requirements..." |
| 2 | Requirements | 10% | "Updating requirements..." |
| 3 | Design | 20% | "Designing..." |
| 4 | Write code | 40% | "Writing code..." |
| 5 | Write viewdefs | 60% | "Writing viewdefs..." |
| 6 | Link and audit | 90% | "Auditing..." |
| 7 | Simplify | 95% | "Simplifying..." |

**`mcp:startTodoStep(n)`** - Advances to step n:
- Marks the previous in_progress step as completed
- Marks step n as in_progress
- Updates progress bar via `mcp:appProgress(currentApp, step.progress, step.label)`
- Updates statusLine with thinking message
- Requires `mcp:createTodos()` to have been called first

**`mcp:completeTodos()`** - Marks all steps complete and clears progress (call at end of build).

**Legacy API:** `mcp:setTodos(todos)` still works for full control over todo state.

## Lua Console

The bottom panel has Chat/Lua tabs. The Lua tab provides a REPL for executing Lua code:
- Output area shows command history and results
- Input textarea for multi-line Lua code
- Run button (or Ctrl+Enter) executes the code
- Clear button clears output history
- Clicking an output line copies it to the input area, focuses the input, and positions cursor at the end

Useful for debugging, inspecting app state, and testing Lua expressions.

## Build Progress

When Claude is building an app, Lua tracks progress state:
- Progress bar (0-100%)
- Stage label (designing, writing code, creating viewdefs, linking)

Claude pushes progress updates via `.ui/mcp run` calling `appConsole:onAppProgress()` when building.

## Events to Claude

Events are sent via `mcp.pushState()` and include `app` (the app name) and `event` (the event type).

**Note field:** Lua includes a `note` field in each event reminding Claude to understand the target app: `"note": "make sure you have understood the app at {base_dir}/apps/{APP}"`. Claude should read the app's requirements.md and design.md before taking action.

### `chat`
User message with selected app as context. Respond conversationally.

**Payload:**
| Field | Description |
|-------|-------------|
| `text` | The user's message |
| `quality` | Quality level: "fast", "thorough", or "background" |
| `handler` | Skill to invoke: `null` (direct edit), `"/ui-builder"`, or `"background-ui-builder"` |
| `context` | Selected app name (if any) |
| `reminder` | Brief reminder to show todos and thinking messages |
| `note` | Path to app files for context |

**Interstitial thinking messages:** While working on a request, send progress updates via `appConsole:addAgentThinking(text)`. These:
- Appear in chat log styled differently (italic, muted)
- Update `mcp.statusLine` and set `mcp.statusClass = "thinking"` (orange bold-italic in MCP shell status bar)

Before sending a thinking message, check for new events first. If there's an event, handle it immediately and save the thinking message as a todo.

Use `appConsole:addAgentMessage(text)` for the final response (clears status bar).

**If the chat involves modifying an app:** Check the `handler` field and follow it exactly:
- `null` (fast quality) — Read app files at `{base_dir}/apps/{context}/`, make the change directly, reply via `appConsole:addAgentMessage()`
- `"/ui-builder"` (thorough quality) — **MUST invoke `/ui-builder` skill** with full phases (design update, code, viewdefs, audit, simplify). Do NOT skip phases or make direct edits.
- `"background-ui-builder"` (background quality) — Spawn background ui-builder agent using the **same prompt template as `build_request`** (see below). Include the user's `text` in the prompt so the agent knows what to do.

**The handler field reflects the user's quality choice and must be respected.** If handler is `/ui-builder`, Claude must use the skill even for "simple" changes.

### `build_request`
Build, complete, or update an app. **Spawn a background ui-builder agent** to handle this.

**Event payload:** `{app: "app-console", event: "build_request", target: "my-app"}`

**Note:** Lua already sets progress to `0, "pondering"` before sending this event, so the user sees immediate feedback when clicking Build.

**Prompt template for background build agent:**
```
Build the app "{target}" at .ui/apps/{target}/

## Progress Reporting

Use the `.ui/mcp` script for all MCP operations:

```bash
.ui/mcp progress {target} <percent> "<stage>"   # Report progress
.ui/mcp run "<lua code>"                        # Execute Lua
.ui/mcp audit {target}                          # Audit app
```

**Report progress at EACH phase of the /ui-builder skill:**

| Phase | Command |
|-------|---------|
| Starting | `.ui/mcp progress {target} 0 "starting..."` |
| Reading requirements | `.ui/mcp progress {target} 5 "reading requirements..."` |
| Updating requirements | `.ui/mcp progress {target} 10 "updating requirements..."` |
| Designing | `.ui/mcp progress {target} 20 "designing..."` |
| Writing code | `.ui/mcp progress {target} 40 "writing code..."` |
| Writing viewdefs | `.ui/mcp progress {target} 60 "writing viewdefs..."` |
| Linking | `.ui/mcp progress {target} 80 "linking..."` |
| Auditing | `.ui/mcp progress {target} 90 "auditing..."` |
| Simplifying | `.ui/mcp progress {target} 95 "simplifying..."` |
| Complete | `.ui/mcp progress {target} 100 "complete"` then `.ui/mcp run "mcp:appUpdated('{target}')"` |

## Instructions

**Run the /ui-builder skill and follow its full workflow.** The skill defines the phases, design spec format, auditing checks, and simplification steps. Do NOT skip phases.

The progress commands above correspond to the skill's phases. Send each progress update BEFORE starting that phase.

The user is watching the progress bar in the UI. Missing progress updates make it look frozen.
```

Then spawn: `Task(subagent_type="ui-builder", run_in_background=true, prompt=<above>)`

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

Claude uses `.ui/mcp run` to call these mcp methods.

### App Initialization (`init.lua`)

The app-console app provides `init.lua` which adds convenience methods to the `mcp` global:

```lua
function mcp:appProgress(name, progress, stage)
    if appConsole then appConsole:onAppProgress(name, progress, stage) end
end

function mcp:appUpdated(name)
    mcp:scanAvailableApps()  -- rescan all apps from disk
    if appConsole then appConsole:onAppUpdated(name) end
end
```

This allows Claude to call `mcp:appProgress()` and `mcp:appUpdated()` without needing to check if the apps-console is loaded. The `mcp:scanAvailableApps()` call ensures the MCP server's app list stays in sync with disk.

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
- `## Gaps` section with non-empty content = design/code mismatch indicator

## Gaps Indicator

When an app's TESTING.md has a non-empty `## Gaps` section, show a warning indicator. This signals that the design and code are out of sync (e.g., methods defined in design but not used, or vice versa).

- In the app list: show a ⚠ icon next to apps with gaps
- In app details: show a "Gaps" section header (similar to Known Issues) that expands to show the gaps content
