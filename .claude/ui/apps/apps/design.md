# Apps - Design

## Intent

Command center for UI development with Claude. Browse apps, see testing status, launch actions (build/test/fix), create new apps, and chat with Claude about selected apps.

## Layout

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
| Chat                                           |
| Agent: Which app would you like to work on?    |
| You: Test the contacts app                     |
| [____________________________________] [Send]  |
+------------------------------------------------+
```

Legend:
- `[R]` = Refresh button
- `[+]` = New app button
- `>` = Selected app indicator
- `████░` = Build progress bar (80%)
- `--` = No tests
- `17/21` = Tests passing/total (green if all pass, yellow if partial)
- `v` = Expanded section, `>` = Collapsed section
- `[✓]` = Test passed, `[ ]` = Untested, `[✗]` = Test failed

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

### AppsApp (main app)

| Field | Type | Description |
|-------|------|-------------|
| _apps | AppInfo[] | All discovered apps |
| selected | AppInfo | Currently selected app |
| showNewForm | boolean | Show new app form vs details |
| newAppName | string | Name input for new app |
| newAppDesc | string | Description input for new app |
| messages | ChatMessage[] | Chat history |
| chatInput | string | Current chat input |

### AppInfo (app metadata)

| Field | Type | Description |
|-------|------|-------------|
| name | string | Directory name |
| description | string | First paragraph from requirements.md |
| isBuilt | boolean | Has app.lua |
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

## Methods

### AppsApp

| Method | Description |
|--------|-------------|
| apps() | Returns _apps (for binding) |
| scanAppsFromDisk() | Full scan: get base_dir via mcp:status(), list apps/, parse each |
| rescanApp(name) | Rescan single app from disk |
| refresh() | Calls scanAppsFromDisk() (Lua-driven) |
| select(app) | Select an app, hide new form |
| openNewForm() | Show new app form, deselect current |
| cancelNewForm() | Hide new app form |
| createApp() | Send create_app event, close form |
| sendChat() | Send chat event with selected app context |
| addAgentMessage(text) | Add agent message to chat |
| onAppProgress(name, progress, stage) | Update app build progress |
| onAppUpdated(name) | Calls rescanApp(name) |
| onAppCreated(name) | Calls rescanApp(name) |

### AppInfo

| Method | Description |
|--------|-------------|
| selectMe() | Call apps:select(self) |
| isSelected() | Check if this app is selected |
| statusText() | Returns "17/21", "not built", "--", or build stage |
| statusVariant() | Returns "success", "warning", "neutral", "primary" |
| hasTests() | Returns true if tests array not empty |
| hasIssues() | Returns true if knownIssues not empty |
| isBuilding() | Returns true if buildProgress is not nil |
| toggleKnownIssues() | Toggle showKnownIssues |
| toggleFixedIssues() | Toggle showFixedIssues |
| requestBuild() | Send build_request event |
| requestTest() | Send test_request event |
| requestFix() | Send fix_request event |
| openApp() | Open app in browser |

### TestItem

| Method | Description |
|--------|-------------|
| isPassed() | Returns status == "passed" |
| isFailed() | Returns status == "failed" |
| isUntested() | Returns status == "untested" |
| icon() | Returns "✓", "✗", or " " based on status |

## ViewDefs

| File | Type | Purpose |
|------|------|---------|
| AppsApp.DEFAULT.html | AppsApp | Main layout with list, details, chat |
| AppInfo.list-item.html | AppInfo | App row in list with status badge |
| TestItem.list-item.html | TestItem | Test checkbox row |
| Issue.list-item.html | Issue | Issue row |
| ChatMessage.list-item.html | ChatMessage | Chat message bubble |

## Events

### From UI to Claude

```json
{"app": "apps", "event": "chat", "text": "...", "context": "contacts"}
{"app": "apps", "event": "build_request", "target": "my-app"}
{"app": "apps", "event": "test_request", "target": "contacts"}
{"app": "apps", "event": "fix_request", "target": "contacts"}
{"app": "apps", "event": "create_app", "name": "my-app", "description": "..."}
```

### From Claude to UI (via mcp.pushState)

```json
{"type": "app_progress", "app": "my-app", "progress": 40, "stage": "writing code"}
{"type": "app_updated", "app": "contacts"}
{"type": "app_created", "app": "my-app"}
```

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