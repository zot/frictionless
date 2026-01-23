# Apps - Design

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
| [Chat] [Lua]                             [⬆]  |
|------------------------------------------------|
| (Chat Panel - when Chat tab selected)          |
| Agent: Which app would you like to work on?    |
| You: Test the contacts app                     |
| [____________________________________] [Send]  |
|------------------------------------------------|
| (Lua Panel - when Lua tab selected)            |
| > print("hello")                               |
| hello                                          |
| [                                    ] [Run]   |
| [Clear]                                        |
+------------------------------------------------+
```

### Embedded App View (replaces top portion)
```
+------------------------------------------------+
|            [ Embedded App View ]               |
|                                                |
|           (displays embeddedValue)             |
|                                                |
+------------------------------------------------+
| ═══════════════ drag handle ════════════════  |
| Chat with Claude                         [⬆]  |
| Agent: Which app would you like to work on?    |
| You: Test the contacts app                     |
| [____________________________________] [Send]  |
+------------------------------------------------+
```

When "Open" is clicked, the selected app replaces the app list + details panel:
- Embedded view displays `embeddedValue` directly (loaded via `mcp:app(appName)`)
- Chat panel remains visible below
- `[⬆]` restore button appears in chat header (far right) to close embedded view

Legend:
- `[R]` = Refresh button
- `[+]` = New app button
- `>` = Selected app indicator
- `████░` = Build progress bar (80%) with hover tooltip showing stage name
- `--` = No tests
- `17/21` = Tests passing/total (green if all pass, yellow if partial)
- `v` = Expanded section, `>` = Collapsed section
- `[✓]` = Test passed, `[ ]` = Untested, `[✗]` = Test failed

**Build progress** (when app is building):
- Progress bar showing 0-100%
- Stage label (e.g., "designing", "writing code")
- Shown between description and action buttons

**Action buttons** (based on app state):
- `[Build]` — shown when app has no viewdefs (needsBuild)
- `[Open]` — shown when app has viewdefs (canOpen), disabled for "app-console" itself
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
| scanAppsFromDisk() | Full scan: get base_dir via mcp:status(), list apps/, parse each |
| rescanApp(name) | Rescan single app from disk |
| refresh() | Calls mcp:scanAvailableApps() then scanAppsFromDisk() |
| select(app) | Select an app, hide new form |
| openNewForm() | Show new app form, deselect current |
| cancelNewForm() | Hide new app form |
| createApp() | Create app dir, write requirements.md, rescan, select new app, send app_created event |
| sendChat() | Send chat event with selected app context |
| addAgentMessage(text) | Add agent message to chat |
| onAppProgress(name, progress, stage) | Update app build progress |
| onAppUpdated(name) | Calls rescanApp(name) |
| openEmbedded(name) | Call mcp:app(name), if not nil set embeddedValue and embeddedApp |
| closeEmbedded() | Clear embeddedApp/embeddedValue, restore normal view |
| hasEmbeddedApp() | Returns true if embeddedApp is set |
| noEmbeddedApp() | Returns true if embeddedApp is nil |
| updateRequirements(name, content) | Update an app's requirementsContent and rescan it from disk |
| showChatPanel() | Set panelMode to "chat" |
| showLuaPanel() | Set panelMode to "lua" |
| isChatPanel() | Returns panelMode == "chat" |
| isLuaPanel() | Returns panelMode == "lua" |
| notChatPanel() | Returns panelMode ~= "chat" |
| notLuaPanel() | Returns panelMode ~= "lua" |
| chatTabVariant() | Returns "primary" if chat, else "default" |
| luaTabVariant() | Returns "primary" if lua, else "default" |
| runLua() | Execute luaInput, append output to luaOutputLines |
| clearLuaOutput() | Clear luaOutputLines |
| qualityLabel() | Returns "Fast", "Thorough", or "Background" |
| qualityValue() | Returns "fast", "thorough", or "background" |
| setChatQuality(value) | Set quality from slider event (handles sl-input) |

### OutputLine

| Method | Description |
|--------|-------------|
| copyToInput() | Copy this line's text to apps.luaInput |

### Apps.AppInfo

| Method | Description |
|--------|-------------|
| selectMe() | Call apps:select(self) |
| isSelected() | Check if this app is selected |
| statusText() | Returns "17/21", "not built", "--", or build stage |
| statusVariant() | Returns "success", "warning", "neutral", "primary" |
| hasTests() | Returns true if tests array not empty |
| hasIssues() | Returns true if knownIssues not empty |
| isBuilding() | Returns true if buildProgress is not nil |
| canOpen() | Returns true if hasViewdefs |
| needsBuild() | Returns true if not hasViewdefs |
| toggleKnownIssues() | Toggle showKnownIssues |
| toggleFixedIssues() | Toggle showFixedIssues |
| toggleRequirements() | Toggle showRequirements |
| requirementsHidden() | Returns not showRequirements |
| requirementsIcon() | Returns chevron icon based on showRequirements |
| pushEvent(eventType, extra) | Push event with common fields (app, mcp_port, note) plus custom fields |
| requestBuild() | Call pushEvent("build_request", {target = self.name}) |
| requestTest() | Call pushEvent("test_request", {target = self.name}) |
| requestFix() | Call pushEvent("fix_request", {target = self.name}) |
| openApp() | Call apps:openEmbedded(self.name) to show in embedded view |
| isSelf() | Returns true if this is the "app-console" app itself |
| canOpenApp() | Returns canOpen() and not isSelf() |

### TestItem

| Method | Description |
|--------|-------------|
| isPassed() | Returns status == "passed" |
| isFailed() | Returns status == "failed" |
| isUntested() | Returns status == "untested" |
| icon() | Returns "✓", "✗", or " " based on status |

### ChatMessage

| Method | Description |
|--------|-------------|
| isUser() | Returns true if sender == "You" |
| prefix() | Returns "> " for user messages, "" for agent |

## ViewDefs

| File | Type | Purpose |
|------|------|---------|
| Apps.DEFAULT.html | Apps | Main layout with list, details, chat/lua panels |
| Apps.AppInfo.list-item.html | Apps.AppInfo | App row in list with status badge |
| Apps.TestItem.list-item.html | Apps.TestItem | Test checkbox row |
| Apps.Issue.list-item.html | Apps.Issue | Issue row |
| Apps.ChatMessage.list-item.html | Apps.ChatMessage | Chat message bubble |
| Apps.OutputLine.list-item.html | Apps.OutputLine | Clickable Lua output line |

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
| `app_created` | Flesh out requirements.md for the new app based on `description`, write to disk, then call `apps:updateRequirements(name, content)` to populate the UI |

**chat event handling:**

Respond to user message about app in `context`. Reply via `apps:addAgentMessage(text)`.

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
4. Clear progress and reply: `ui_run('mcp:appUpdated("{context}"); apps:addAgentMessage("Done - {brief description}")')`

**Background Agent Pattern (build_request):**

Spawn a background ui-builder agent:
```
Task(subagent_type="ui-builder", run_in_background=true, prompt="Build the {target} app at .ui/apps/{target}/")
```

Tell the ui-builder agent:
- Use `.ui/` scripts for MCP operations (they read the port from `.ui/mcp-port`)
- Report progress via `.ui/progress {target} {percent} "{stage}"`
- Call `.ui/run "mcp:appUpdated('{target}')"` on completion (triggers rescan)

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
    if apps then apps:onAppProgress(name, progress, stage) end
end

function mcp:appUpdated(name)
    mcp:scanAvailableApps()  -- rescan all apps from disk
    if apps then apps:onAppUpdated(name) end
end
```

This allows Claude to call `mcp:appProgress()` and `mcp:appUpdated()` without checking if apps is loaded. The `mcp:scanAvailableApps()` call ensures the MCP server's app list stays in sync with disk.

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
