# App Console - Design

## Intent

Command center for UI development with Claude. Browse apps, see testing status, launch actions (build/test/fix), create new apps, and chat with Claude about selected apps.

## Layout

### Normal View (app list + details)
```
+------------------+-----------------------------+
| Apps      [R][+] | contacts                    |
|------------------|  A contact manager with...  |
| > contacts 17/21 | [Open] [Test] [Fix Issues]  |
|   tasks    5/5   |-----------------------------|
|   my-app   ████░ | Tests (17/21)               |
|   new-app  --    | [✓] Badge shows count       |
|                  | [ ] Delete removes contact  |
|                  | [✗] Edit saves changes      |
|                  |-----------------------------|
|                  | v Known Issues (2)          |
|                  |   1. Status dropdown broken |
|                  |   2. Cancel doesn't revert  |
|                  | > Fixed Issues (1)          |
+------------------+-----------------------------+
| ═══════════════ drag handle ════════════════  |
| Todos [▼]    | [Chat] [Lua]              [⬆]  |
|--------------|----------------------------------|
| 🔄 Reading   | (Chat Panel - when Chat tab)    |
| ⏳ Update    | Agent: Which app would you like |
| ⏳ Test      | You: Test the contacts app      |
| ✓ Design     | [____________________________]  |
|              |                          [Send] |
+------------------------------------------------+
```

### Embedded App View (in detail area)
```
+------------------+-----------------------------+
| Apps      [R][+] |        [contacts]     [X]   |
|------------------|-----------------------------+
| > contacts 17/21 |                             |
|   tasks    5/5   |   [ Embedded App View ]     |
|   my-app   ████░ |                             |
|   new-app  --    |   (displays embeddedValue)  |
|                  |                             |
|                  |                             |
+------------------+-----------------------------+
| ═══════════════ drag handle ════════════════  |
| Todos [▼]    | [Chat] [Lua]                   |
| ...          | ...                             |
+------------------------------------------------+
```

When "Open" is clicked, the embedded app replaces the detail panel:
- App list remains visible on the left
- Embedded view displays `embeddedValue` in the detail area
- Header shows app name and close button `[X]`
- Chat panel remains visible below

Legend:
- `[R]` = Refresh button
- `[+]` = New app button
- `>` = Selected app indicator
- `████░` = Build progress bar (80%) with hover tooltip showing stage name
- `--` = No tests
- `17/21` = Tests passing/total (green if all pass, yellow if partial)
- `v` = Expanded section, `>` = Collapsed section
- `[✓]` = Test passed, `[ ]` = Untested, `[✗]` = Test failed
- Todo: `🔄` = in_progress (blue), `⏳` = pending (gray), `✓` = completed (green/muted)
- `[▼]` = Collapse todos button

**Build progress** (when app is building):
- Progress bar showing 0-100%
- Stage label (e.g., "designing", "writing code")
- Shown between description and action buttons

**Action buttons** (based on app state):
- `[Build]` — shown when app has no viewdefs (needsBuild)
- `[Open]` — shown when app has viewdefs (canOpen), disabled for "app-console" and "mcp"
- `[Test]` — shown when app has viewdefs
- `[Fix Issues]` — shown when app has known issues

### New App Form (replaces details)
```
+------------------+-----------------------------+
| Apps      [R][+] | New App                     |
|------------------|                             |
|   contacts 17/21 | Name: [_______________]     |
|   tasks    5/5   |                             |
|                  | Description:                |
|                  | [                         ] |
|                  |                             |
|                  | [Cancel]         [Create]   |
+------------------+-----------------------------+
```

### Status Badge Colors

| Condition | Color | Variant |
|-----------|-------|---------|
| All tests passing | Green | success |
| Some tests failing | Yellow | warning |
| Not built / No tests | Gray | neutral |
| Building | Blue | primary |

## Data Model

### Apps (main app)

| Field | Type | Description |
|-------|------|-------------|
| _apps | AppInfo[] | All discovered apps |
| selected | AppInfo | Currently selected app |
| showNewForm | boolean | Show new app form vs details |
| newAppName | string | Name input for new app |
| newAppDesc | string | Description input for new app |
| messages | ChatMessage[] | Chat history |
| chatInput | string | Current chat input |
| embeddedApp | string | Name of embedded app, or nil |
| embeddedValue | object | App global loaded via mcp:app, or nil |
| panelMode | string | "chat" or "lua" (bottom panel mode) |
| luaOutputLines | OutputLine[] | Lua console output history |
| luaInput | string | Current Lua code input |
| chatQuality | number | 0=fast, 1=thorough, 2=background |
| todos | TodoItem[] | Claude Code todo list items |
| todosCollapsed | boolean | Whether todo column is collapsed |

### TodoItem (Claude Code task)

| Field | Type | Description |
|-------|------|-------------|
| content | string | Task description (shown for pending/completed) |
| status | string | "pending", "in_progress", or "completed" |
| activeForm | string | Present tense form (shown for in_progress) |

### AppInfo (app metadata)

| Field | Type | Description |
|-------|------|-------------|
| name | string | Directory name |
| description | string | First paragraph from requirements.md |
| requirementsContent | string | Full requirements.md content |
| showRequirements | boolean | Expand requirements section (default false) |
| hasViewdefs | boolean | Has viewdefs/ directory |
| tests | TestItem[] | Test checklist from TESTING.md |
| testsPassing | number | Count of passing tests |
| testsTotal | number | Total test count |
| knownIssues | Issue[] | Open issues from TESTING.md |
| fixedIssues | Issue[] | Resolved issues from TESTING.md |
| showKnownIssues | boolean | Expand known issues section |
| showFixedIssues | boolean | Expand fixed issues section (default false) |
| gapsContent | string | Content of ## Gaps section from TESTING.md |
| showGaps | boolean | Expand gaps section (default false) |
| buildProgress | number | 0-100 or nil |
| buildStage | string | Current build stage or nil |

### Issue

| Field | Type | Description |
|-------|------|-------------|
| number | number | Issue number |
| title | string | Issue title/summary |

### TestItem

| Field | Type | Description |
|-------|------|-------------|
| text | string | Test description |
| status | string | "passed", "failed", or "untested" |

### ChatMessage

| Field | Type | Description |
|-------|------|-------------|
| sender | string | "You" or "Agent" |
| text | string | Message content |
| style | string | "normal" (default) or "thinking" for interstitial progress |

### OutputLine

| Field | Type | Description |
|-------|------|-------------|
| text | string | Line content |
| panel | ref | Reference to Apps for copyToInput |

## Chat Panel Features

### Resizable
- Drag handle at top edge of chat panel
- Drag up to increase height, down to decrease
- Minimum height: 120px, Maximum: 60vh
- Initial height: 200px

### Auto-scroll
- Messages container uses `scrollOnOutput` variable property
- New messages automatically scroll into view

### User Message Styling
- User messages prefixed with `>` character
- CSS class `user-message` applied for distinct styling

## Methods

### Apps

| Method | Description |
|--------|-------------|
| apps() | Returns _apps (for binding) |
| findApp(name) | Find app by name in _apps |
| scanAppsFromDisk() | Full scan: get base_dir via mcp:status(), list apps/, parse each |
| rescanApp(name) | Rescan single app from disk |
| refresh() | Calls mcp:scanAvailableApps() then scanAppsFromDisk() |
| select(app) | Select an app, hide new form |
| openNewForm() | Show new app form, deselect current |
| cancelNewForm() | Hide new app form |
| createApp() | Create app dir, write requirements.md, rescan, select new app, send app_created event |
| sendChat() | Send chat event with selected app context |
| onAppProgress(name, progress, stage) | Update app build progress |
| onAppUpdated(name) | Calls rescanApp(name) |
| openEmbedded(name) | Call mcp:app(name), if not nil set embeddedValue and embeddedApp |
| closeEmbedded() | Clear embeddedApp/embeddedValue, restore normal view |
| hasEmbeddedApp() | Returns true if embeddedApp is set |
| noEmbeddedApp() | Returns true if embeddedApp is nil |
| showDetail() | Returns true if selected and not showNewForm |
| hideDetail() | Returns not showDetail() |
| showPlaceholder() | Returns true if no selection and not showNewForm |
| hidePlaceholder() | Returns not showPlaceholder() |
| hideNewForm() | Returns not showNewForm |
| showChatPanel() | Set panelMode to "chat" |
| showLuaPanel() | Set panelMode to "lua" |
| notChatPanel() | Returns panelMode ~= "chat" |
| notLuaPanel() | Returns panelMode ~= "lua" |
| chatTabVariant() | Returns "primary" if chat, else "default" |
| luaTabVariant() | Returns "primary" if lua, else "default" |
| runLua() | Execute luaInput, append output to luaOutputLines |
| clearLuaOutput() | Clear luaOutputLines |
| qualityLabel() | Returns "Fast", "Thorough", or "Background" |
| qualityValue() | Returns "fast", "thorough", or "background" |
| setChatQuality() | Update quality from slider event |
| toggleTodos() | Toggle todosCollapsed state |

**External methods (called by Claude via ui_run/mcp methods):**

| Method | Description |
|--------|-------------|
| addAgentMessage(text) | Add agent message to chat, clear mcp.statusLine |
| addAgentThinking(text) | Add thinking message to chat, update mcp.statusLine |
| updateRequirements(name, content) | Update an app's requirementsContent |
| setTodos(todos) | Replace todo list with new items |

### TodoItem

| Method | Description |
|--------|-------------|
| displayText() | Returns activeForm if in_progress, else content |
| isPending() | Returns status == "pending" |
| isInProgress() | Returns status == "in_progress" |
| isCompleted() | Returns status == "completed" |
| statusIcon() | Returns "🔄" for in_progress, "⏳" for pending, "✓" for completed |

### OutputLine

| Method | Description |
|--------|-------------|
| copyToInput() | Copy this line's text to appConsole.luaInput |

### AppConsole.AppInfo

| Method | Description |
|--------|-------------|
| selectMe() | Call appConsole:select(self) |
| isSelected() | Check if this app is selected |
| statusText() | Returns "17/21", "not built", "--", or build stage |
| statusVariant() | Returns "success", "warning", "neutral", "primary" |
| noTests() | Returns true if testsTotal == 0 |
| noIssues() | Returns true if knownIssues is empty |
| noFixedIssues() | Returns true if fixedIssues is empty |
| hasGaps() | Returns true if gapsContent is non-empty |
| noGaps() | Returns true if gapsContent is empty |
| toggleGaps() | Toggle showGaps |
| gapsHidden() | Returns not showGaps |
| gapsIcon() | Returns chevron icon based on showGaps |
| isBuilding() | Returns true if buildProgress is not nil |
| notBuilding() | Returns true if buildProgress is nil |
| canOpen() | Returns true if hasViewdefs |
| needsBuild() | Returns true if not hasViewdefs |
| toggleKnownIssues() | Toggle showKnownIssues |
| toggleFixedIssues() | Toggle showFixedIssues |
| toggleRequirements() | Toggle showRequirements |
| knownIssuesHidden() | Returns not showKnownIssues |
| fixedIssuesHidden() | Returns not showFixedIssues |
| requirementsHidden() | Returns not showRequirements |
| knownIssuesIcon() | Returns chevron icon based on showKnownIssues |
| fixedIssuesIcon() | Returns chevron icon based on showFixedIssues |
| requirementsIcon() | Returns chevron icon based on showRequirements |
| knownIssueCount() | Returns #knownIssues |
| fixedIssueCount() | Returns #fixedIssues |
| pushEvent(eventType, extra) | Push event with common fields (app, mcp_port, note) plus custom fields |
| requestBuild() | Call pushEvent("build_request", {target = self.name}) |
| requestTest() | Call pushEvent("test_request", {target = self.name}) |
| requestFix() | Call pushEvent("fix_request", {target = self.name}) |
| openApp() | Call appConsole:openEmbedded(self.name) to show in embedded view |
| isSelf() | Returns true if this is the "app-console" app itself |
| isMcp() | Returns true if this is the "mcp" app |
| openButtonDisabled() | Returns true if Open button should be disabled (app-console or mcp) |

### TestItem

| Method | Description |
|--------|-------------|
| icon() | Returns "✓", "✗", or " " based on status |
| iconClass() | Returns "passed", "failed", or "untested" for CSS styling |

### ChatMessage

| Method | Description |
|--------|-------------|
| isUser() | Returns true if sender == "You" |
| isThinking() | Returns true if style == "thinking" |
| prefix() | Returns "> " for user messages, "" for agent |

## ViewDefs

| File | Type | Purpose |
|------|------|---------|
| AppConsole.DEFAULT.html | AppConsole | Main layout with list, details, chat/lua panels |
| AppConsole.AppInfo.list-item.html | AppConsole.AppInfo | App row in list with status badge |
| AppConsole.TestItem.list-item.html | AppConsole.TestItem | Test checkbox row |
| AppConsole.Issue.list-item.html | AppConsole.Issue | Issue row |
| AppConsole.ChatMessage.list-item.html | AppConsole.ChatMessage | Chat message bubble |
| AppConsole.OutputLine.list-item.html | AppConsole.OutputLine | Clickable Lua output line |
| AppConsole.TodoItem.list-item.html | AppConsole.TodoItem | Todo item with status icon |

## Events

### From UI to Claude

```json
{"app": "app-console", "event": "chat", "text": "...", "context": "contacts", "quality": "fast", "note": "make sure you have understood the app at /path/apps/contacts"}
{"app": "app-console", "event": "build_request", "target": "my-app", "mcp_port": 37067, "note": "make sure you have understood the app at /path/apps/my-app"}
{"app": "app-console", "event": "test_request", "target": "contacts", "mcp_port": 37067, "note": "make sure you have understood the app at /path/apps/contacts"}
{"app": "app-console", "event": "fix_request", "target": "contacts", "mcp_port": 37067, "note": "make sure you have understood the app at /path/apps/contacts"}
{"app": "app-console", "event": "app_created", "name": "my-app", "description": "A brief description", "mcp_port": 37067, "note": "make sure you have understood the app at /path/apps/my-app"}
```

Lua includes `mcp_port` from `mcp:status()` in action events so Claude can spawn agents directly. Each event also includes a `note` field with the full path to the app, reminding Claude to read the app's docs before acting.

### Claude Event Handling

| Event | Handler |
|-------|---------|
| `chat` | Handle based on `quality` field - see below |
| `build_request` | Spawn background ui-builder agent using `mcp_port` from event |
| `test_request` | Spawn background agent to run `/ui-testing` on app in `target` |
| `fix_request` | Spawn background agent to read TESTING.md and fix issues using `/ui-builder` |
| `app_created` | Flesh out requirements.md for the new app based on `description`, write to disk, then call `appConsole:updateRequirements(name, content)` to populate the UI |

**chat event payload:**

| Field | Description |
|-------|-------------|
| `text` | The user's message |
| `quality` | Quality level: "fast", "thorough", or "background" |
| `context` | Selected app name (if any) |
| `reminder` | Brief reminder to show todos and thinking messages |
| `note` | Path to app files for context |

**chat event handling:**

Respond to user message about app in `context`. Reply via `appConsole:addAgentMessage(text)`.

**Interstitial thinking messages:** While working on a request, send progress updates via `appConsole:addAgentThinking(text)`. These:
- Appear in chat log styled differently (italic, muted)
- Update `mcp.statusLine` and set `mcp.statusClass = "thinking"` (orange bold-italic in MCP status bar)

**Before sending a thinking message:** Check for new events first. If there's an event, handle it immediately and save the thinking message as a todo to send after. This prevents missing user events while sending status updates.

Use for:
- "Let me look at the code..."
- "Found the issue, fixing it now..."
- "Reading the design spec..."

Then send the final response via `appConsole:addAgentMessage(text)` (which clears `mcp.statusLine`).

**If the chat involves modifying an app:** Check the `quality` field:

| Quality | Behavior |
|---------|----------|
| `fast` | Direct edit - see below |
| `thorough` | Use `/ui-builder` skill inline with full phases |
| `background` | Spawn background ui-builder agent (shows progress bar) |

**Fast quality (direct edit):**

1. Set progress: `ui_run('mcp:appProgress("{context}", 0, "editing")')`
2. Read the target app at `{base_dir}/apps/{context}/`:
   - `design.md` - understand app structure and data model
   - `app.lua` - for Lua code changes
   - `viewdefs/*.html` - for UI/styling changes
3. Make the requested change using Edit tool
4. Clear progress and reply: `ui_run('mcp:appUpdated("{context}"); appConsole:addAgentMessage("Done - {brief description}")')`

**Background Agent Pattern (build_request):**

Use `mcp_port` from the event payload:
```
Task(subagent_type="ui-builder", run_in_background=true, prompt="MCP port is {mcp_port}. Build the {target} app at .ui/apps/{target}/")
```

Before spawning the agent, use `ui_run` to update app progress with (APP, 0%, "thinking...")

Tell the ui-builder agent:
- Use HTTP API (curl) since background agents don't have MCP tool access
- Report progress via `curl -s -X POST http://127.0.0.1:{mcp_port}/api/ui_run -d 'mcp:appProgress("{name}", {progress}, "{stage}")'`
- Call `mcp:appUpdated("{name}")` on completion (triggers rescan)

Background agents allow Claude to continue handling chat while builds run.

### From Claude to UI (via mcp methods)

Claude uses `ui_run` to call mcp convenience methods:
```lua
mcp:appProgress("my-app", 40, "writing code")
mcp:appUpdated("contacts")
```

## App Initialization (`init.lua`)

The app-console app provides `init.lua` which adds convenience methods to the `mcp` global:

```lua
function mcp:appProgress(name, progress, stage)
    if appConsole then appConsole:onAppProgress(name, progress, stage) end
end

function mcp:appUpdated(name)
    mcp:scanAvailableApps()  -- rescan all apps from disk
    if appConsole then appConsole:onAppUpdated(name) end
end

function mcp:setTodos(todos)
    if appConsole then appConsole:setTodos(todos) end
end
```

This allows Claude to call `mcp:appProgress()`, `mcp:appUpdated()`, and `mcp:setTodos()` without checking if app-console is loaded. The `mcp:scanAvailableApps()` call ensures the MCP server's app list stays in sync with disk.

## File Parsing (Lua)

### requirements.md
- First paragraph (text before first blank line) = description

### TESTING.md
- `- [ ]` = untested
- `- [✓]` = passed
- `- [✗]` = failed
- Count passed/total for status display
- `### N.` under "Known Issues" = open bugs
- `### N.` under "Fixed Issues" = resolved bugs
- `## Gaps` section content = design/code mismatch (show ⚠ if non-empty)
