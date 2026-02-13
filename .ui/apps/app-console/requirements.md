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

A "+" button in the header opens the new app form. A GitHub icon button opens the GitHub download form.

Clicking an app selects it and shows details in the adjacent panel.

## App Details

When an app is selected, show:
- App name as header
- Source URL row (for downloaded apps): GitHub URL with link icons to open source and readme in new tabs
- Description (first paragraph from requirements.md, parsed by Lua)
- Action buttons based on state (shown above requirements/progress sections):
  - Build (when no viewdefs) - sets progress to `0, "pondering"` then sends build_request to Claude
  - Show (when has viewdefs) - opens the app in the embedded app view (disabled for "app-console" and "mcp")
  - Make it thorough (N) (when has checkpoints) - shows count of pending changes, tooltip says "N pending changes", sends consolidate_request to invoke `/ui-thorough` skill
  - Test (when has app.lua)
  - Fix Issues (when has known issues)
  - Review Gaps (when has gaps) - sends review_gaps_request to invoke `/ui-thorough` skill
  - Analyze (when built) - sends analyze_request for full gap analysis even without existing gaps
  - Delete App (when not protected) - shows confirmation dialog, then removes the app entirely
- Requirements section (expandable, collapsed by default) - shows full requirements.md content
- Build progress and phase (when app is building) - shows progress bar and stage label
- Test checklist from TESTING.md with checkboxes (read-only, parsed by Lua)
- Known Issues section (expandable)
- Fixed Issues section (collapsed by default)

## Embedded App View

When the "Show" button is clicked, the selected app replaces the detail panel (right side):
- App list remains visible on the left
- The embedded view displays `embeddedValue` directly (not an iframe)
- Header shows app name and close button `[X]`
- The MCP shell's chat panel remains visible below
- User can interact with the embedded app while still chatting with Claude

Clicking the close button `[X]` closes the embedded view and restores the normal detail panel.

## GitHub Download

Download apps from GitHub repositories. The form validates the URL, lets the user inspect files for security, and downloads the app.

### GitHub Icon Button
A GitHub icon button in the header opens the download form (replaces the detail panel, like the new app form).

### URL Input
- Input field for GitHub tree URL (e.g., `https://github.com/user/repo/tree/main/apps/my-app`)
- "Investigate" button validates the URL and fetches directory contents
- Shows error if URL is invalid or directory doesn't contain a valid app

### Name Conflict Detection
When the user enters a URL, check if an app with the same name already exists:
- Parse app name from the URL path (last segment)
- Check if `{base_dir}/apps/{name}` directory exists
- If conflict exists, show a danger alert: "App 'name' already exists in .ui/apps/. Delete or rename it before downloading."
- Disable the Investigate button when there's a conflict

### Validation
A valid app directory must contain:
- `requirements.md`
- `design.md`
- `app.lua`
- `viewdefs/` directory

### File Inspection (Security Review)
After validation, show tabs for each file:
- `requirements.md`, `design.md`, `app.lua`, plus any other `.lua` files
- User must click each tab to mark it as "viewed"
- Unviewed tabs shown with warning variant (yellow)
- Viewed tabs shown with default variant
- Selected tab shown with primary variant

### Security Warnings for Lua Files
Lua files are analyzed for potentially dangerous code:
- **pushState calls** (orange): Count occurrences of `pushState` - these can send events to Claude
- **Dangerous calls** (red): Count occurrences of dangerous patterns:
  - Shell execution: `os.execute`, `io.popen`
  - Code loading: `dofile`, `load`, `loadfile`, `loadstring`
  - File operations: `io.open`, `io.input`
  - OS operations: `os.exit`, `os.remove`, `os.rename`, `os.tmpname`
  - Dynamic require: `require(variable)` (but NOT `require("constant")`)

Tab labels show warning counts: `app.lua (3 events, 2 danger)`

Tooltips explain the warnings:
- "N pushState call(s) - can send events to Claude"
- "N os.execute/io.popen call(s) - runs shell commands"

### Syntax Highlighting
In the file content viewer, highlight dangerous code:
- **Orange highlight**: Lines containing `pushState` calls (with orange left border)
- **Red highlight**: Lines containing `os.execute(` or `io.popen(` calls (with red left border)

### Scrollbar Trough Markers
A narrow trough next to the scrollbar shows the positions of all warnings in the file:
- Orange markers indicate pushState call locations
- Red markers indicate os.execute/io.popen locations
- Markers use DOM measurement (getBoundingClientRect) for accurate positioning
- Marker position aligns with scrollbar thumb when warning is at viewport top
- Consecutive warning lines are wrapped in group spans for accurate bounding boxes
- Helps users quickly see where warnings are in long files

### Warning Group Styling
Warning blocks (consecutive pushState lines or single os.execute lines) are visually grouped:
- Group spans wrap consecutive warning lines
- Left border bar in the warning color (orange for pushState, red for os.execute)
- Block display to ensure proper line grouping

### Safety Message
Before any tab is clicked, show a warning alert:
- "Review before approving"
- Instructions to click each tab and review contents
- Explains the color-coded highlights

### Approve Button
- Only enabled after ALL tabs have been viewed
- Downloads the app from GitHub:
  1. Fetch repository zip from GitHub
  2. Extract the app directory to `{base_dir}/apps/{name}/`
  3. Save the source URL to `{base_dir}/apps/{name}/source.txt`
  4. Create `original.fossil` baseline for local changes tracking
  5. Link the app using `.ui/mcp linkapp add {name}`
  6. Refresh the app list and select the new app

### Downloaded App Tracking
Downloaded apps are tracked for local modifications:
- `source.txt` stores the original GitHub URL
- `original.fossil` stores the original code state (copy of checkpoint.fossil baseline)
- Local changes are detected by comparing current state to original.fossil
- Apps with local changes show a pencil icon in the app list

### Cancel Button
Closes the form and clears all state.

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

Use `mcp:appProgress(name, percent, stage)` for the progress bar, `mcp:addAgentThinking(text)` for chat panel updates, then `mcp:addAgentMessage(text)` for the final response. Call `appConsole:updateRequirements(name, content)` to populate the requirements in the UI, then `mcp:appProgress(name, nil, nil)` to clear the progress bar.

## Delete App

Allow users to delete apps that are not protected. Protected apps include: `app-console`, `mcp`, `claude-panel`, `viewlist`.

**Delete Button:**
- Only shown for non-protected apps (apps not in the protected list)
- Clicking shows a confirmation dialog
- Located in the action buttons area

**Confirmation Dialog:**
- Shows warning message asking user to confirm deletion
- Has "Delete" and "Cancel" buttons
- Prevents accidental deletion

**On Delete (Lua):**
1. Set the app's global variables to nil (e.g., `contacts = nil`, `Contacts = nil`)
2. Unlink the app from lua/ and viewdefs/ directories
3. Delete the app directory recursively
4. Remove the app from the apps list
5. Clear selection

## Chat Panel, Todo List, Lua Console

These features live in the **MCP shell** (not app-console). The MCP shell provides:
- Chat panel for conversing with Claude (with selected app context from app-console)
- Todo list showing Claude's current build progress
- Lua console REPL for debugging

The chat panel's `sendChat()` reads `appConsole.selected` to provide app context when sending messages. The todo system calls `appConsole:onAppProgress()` to update build progress in the app list.

See the MCP shell's design.md for details on these features.

**Build Mode:**

Build mode is controlled globally via the status bar toggle (rocket=fast, diamond=thorough). The `handler` field is injected into every event automatically by the MCP layer.

See `/ui` skill's "Build Mode" section for how Claude should handle the `handler` field.

## Build Progress

When Claude is building an app, Lua tracks progress state:
- Progress bar (0-100%)
- Stage label (designing, writing code, creating viewdefs, linking)

Claude pushes progress updates via `.ui/mcp run` calling `mcp:appProgress()` which delegates to `appConsole:onAppProgress()` when building.

## Events to Claude

Events are sent via `mcp.pushState()` and include `app` (the app name) and `event` (the event type).

**Note field:** Lua includes a `note` field in each event reminding Claude to understand the target app: `"note": "make sure you have understood the app at {base_dir}/apps/{APP}"`. Claude should read the app's requirements.md and design.md before taking action.

### `chat`
User message with selected app as context. Respond conversationally.

**Payload:**
| Field | Description |
|-------|-------------|
| `text` | The user's message |
| `handler` | Skill to invoke: `"/ui-fast"` or `"/ui-thorough"` (injected by MCP) |
| `background` | Whether to run as background agent (injected by MCP) |
| `context` | Selected app name (if any) |
| `reminder` | Brief reminder to show todos and thinking messages |
| `note` | Path to app files for context |

**Handler dispatch:** See `/ui` skill's "Build Mode" section for how to handle the `handler` field. Always respect it.

**Interstitial thinking messages:** While working on a request, send progress updates via `mcp:addAgentThinking(text)`. These:
- Appear in the MCP shell's chat log styled differently (italic, muted)
- Update `mcp.statusLine` and set `mcp.statusClass = "thinking"` (orange bold-italic in MCP shell status bar)

Before sending a thinking message, check for new events first. If there's an event, handle it immediately and save the thinking message as a todo.

Use `mcp:addAgentMessage(text)` for the final response (clears status bar).

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

### `consolidate_request`
Invoke the `/ui-thorough` skill to integrate checkpointed changes from `/ui-fast` into requirements.md and design.md. Always invokes `/ui-thorough` regardless of the build mode toggle. After consolidation, checkpoints are cleared.

### `review_gaps_request`
Review the `## Gaps` section in the target app's TESTING.md. For each gap item:
1. Verify the feature is properly documented in design.md
2. If documented correctly, remove the item from gaps
3. If not documented, add documentation to design.md then remove from gaps

After processing, the `## Gaps` section should be **empty** (no placeholder text). Always invokes `/ui-thorough` regardless of the build mode toggle.

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

The app-console's `init.lua` provides a hot-load test function. All MCP convenience methods (`mcp:appProgress`, `mcp:appUpdated`, `mcp:addAgentMessage`, `mcp:createTodos`, etc.) now live directly in the MCP shell's `app.lua`.

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

## Checkpoint Indicator

Apps can have checkpoints from `/ui-fast` prototyping sessions. Show checkpoint status in the app list:

- Rocket icon (🚀) next to apps WITH checkpoints (rapid prototyping in progress)
- Gem icon (💎) next to apps WITHOUT checkpoints (stable/thorough)

The icon is shown in the app list item, before the app name.

## Local Changes Indicator

Downloaded apps (from GitHub) track local modifications:

- Pencil icon (✏️) next to apps with local changes vs original download
- No icon if app matches original state
- Tooltip: "Modified since download"

The icon appears after the checkpoint icon in the app list item.

## Make it Thorough Button

When an app has checkpoints, show a "Make it thorough" button in the action buttons area. This button:

- Only appears when the app has checkpoints (`hasCheckpoints()`)
- Sends a `consolidate_request` event to Claude
- Claude invokes `/ui-thorough` to integrate the checkpointed changes into both requirements.md and design.md
- After consolidation, checkpoints are cleared

This allows users to prototype quickly with `/ui-fast`, then consolidate changes into proper documentation when ready.

## Styling

This app inherits the terminal aesthetic from the MCP shell, using CSS variables:

**Color Palette:**
- `--term-bg`: deep dark background (#0a0a0f)
- `--term-bg-elevated`: raised surfaces (#12121a)
- `--term-bg-hover`: hover states (#1a1a24)
- `--term-border`: subtle borders (#2a2a3a)
- `--term-text`: primary text (#e0e0e8)
- `--term-text-dim`: secondary text (#8888a0)
- `--term-accent`: orange accent (#E07A47)
- `--term-accent-glow`: glow effects (rgba(224, 122, 71, 0.4))

**Typography:**
- `--term-mono`: JetBrains Mono, Fira Code, Consolas (code)
- `--term-sans`: Space Grotesk, system-ui (headings)

**Component Overrides:**
All Shoelace components require dark theme overrides via `::part()` selectors:
- Inputs, textareas, selects: dark backgrounds, dim borders, orange focus glow
- Buttons: elevated backgrounds, orange hover states
- Badges: accent colors with glow
- Alerts: dark backgrounds, color-coded borders

**Selection States:**
Selected app in list shows orange left border accent.
