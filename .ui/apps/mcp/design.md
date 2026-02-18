# MCP - Design

## Intent

Outer shell for all frictionless apps. Displays the current app full-viewport and provides a floating app switcher menu in the top-right corner.

## Layout

```
+--------------------------------------------------+
|                                     [:::] or (o) <- 9-dot menu or spinner
|                                                  |
|              Current App (mcp.value)             |
|                  (full viewport)                 |
|                                                  |
|                                                  |
+--------------------------------------------------+
| ═══════════ drag handle ══════════════════════  | <- chat panel (toggleable)
| Todos [▼][🗑] | [Chat] [Lua]              [Clear]|
|--------------|-----------------------------------|
| 🔄 Reading   | Agent: Reading the design...     |
| ⏳ Update    | You: Build the contacts app      |
| ✓ Design     | [____________________________]   |
|              |                          [Send]  |
+--------------------------------------------------+
| Status: mcp.statusLine        [{} ❓ 🔧 🚀 ⏳ 💬] | <- status bar
+--------------------------------------------------+
[hidden: ui-code element]
```

Chat panel is between the app content and status bar. Toggled by the 💬 icon (rightmost in status bar).

### Status Bar

Fixed at the bottom of the viewport, always visible. Compact horizontal padding (6px). Displays `mcp.statusLine` text with `mcp.statusClass` CSS class applied. The `.thinking` class styles text as orange bold-italic.

At the right end of the status bar are icons grouped tightly together in a `.mcp-status-toggles` container (no gap between icons):

| Icon | Action | Description |
|------|--------|-------------|
| `{}` braces | variablesLinkHtml() | Opens `/variables` in new tab (variable tree) |
| ❓ question mark | helpLinkHtml() | Opens `/api/resource/` in new tab (documentation) |
| 🔧 tools | openTools() | Opens app-console, selects current app |
| 🚀/💎 | toggleBuildMode() | fast / thorough |
| ⏳/🔄 | toggleBackground() | foreground / background |
| 💬 chat-dots | togglePanel() | Toggle chat/lua/todo panel (rightmost) |

Icon styling: minimal padding (2px vertical, 3px horizontal), no gap between icons. Click triggers action. Hover shows dynamic tooltip.

The braces and question mark icons use `ui-html` to generate `<a>` tags with `target="_blank"`. They are styled in purple (#bb88ff) with brighter hover (#dd99ff) via `.mcp-build-mode-toggle a` CSS rules. The HTML is cached since the port doesn't change during a session.

### Tools Icon

The tools icon shows the current app in app-console when clicked. If the current app has checkpoints (fast code), the icon glows orange via CSS `filter: drop-shadow()`. The tooltip shows "Go to App" normally or "Go to App - fast coded" when checkpoints exist.

### Processing Indicator

When the agent event loop is not connected to `/wait` (Claude is processing):
- The menu button enters a `.waiting` state via `ui-class-waiting="isWaiting()"`
- A pulsating orange glow effect animates around the button (CSS `button-pulse` animation)
- The grid icon dims to 30% opacity
- A wait time counter appears centered over the button
- The button remains fully clickable during wait state

### Wait Time Counter (Client-Local JavaScript)

Client-side JavaScript manages the counter display without server round-trips:
- A `<script>` block with `setInterval(200ms)` reads timestamp from hidden element
- Server provides `waitStartOffset()` - UNIX timestamp when wait started, or 0 if connected
- Counter calculates elapsed seconds client-side
- Counter shows seconds elapsed, empty when <= 5 seconds
- Bold orange text with glow, centered in button (`.mcp-wait-counter`)
- CSS controls visibility via parent `.waiting` class

### pushState Override

On load, idempotently override the global `pushState` function to:

1. **Inject build settings:**
   - `event.handler` = `"/ui-fast"` or `"/ui-thorough"`
   - `event.background` = `true` or `false`

2. **Warn on long wait times:**
   - If `mcp:waitTime() > 5` and not already notified, show warning notification
   - Resets `_notifiedForDisconnect` when Claude reconnects (waitTime returns to 0)

### Disconnect Check (checkDisconnectNotify)

Called on UI refresh via hidden span binding. Warns if:
- `waitTime() > 5` seconds AND
- `pendingEventCount() > 0` (events are waiting) AND
- Not already notified this disconnect period

This catches the case where user interacts with UI but Claude isn't listening.

### Menu Open State (Icon Grid)

```
+--------------------------------------------------+
|                                        [:::] <- 9-dot menu
|                            +-------------------+ |
|              Current App   | [icon] [icon] [icon]|
|                            |  app1   app2   app3 |
|                            | [icon] [icon] [icon]|
|                            |  app4   app5   app6 |
|                            +-------------------+ |
+--------------------------------------------------+
```

Icons arranged in rows of 3 (Z formation: left-to-right, then next row). Each cell shows the app's icon (from icon.html) with the app name below it.

## Data Model

### MCP (extends global mcp object)

The global `mcp` object is provided by the server. This app adds:

| Field | Type | Description |
|-------|------|-------------|
| value | object | Currently displayed app (set by mcp:display) |
| code | string | JavaScript to execute via ui-code |
| _availableApps | string[] | List of discovered app names |
| menuOpen | boolean | Whether app menu is visible |
| statusLine | string | Status text to display (server-provided) |
| statusClass | string | CSS class for status bar styling (e.g., "thinking") |
| _notifications | Notification[] | Active notification toasts |
| buildMode | string | "fast" or "thorough" - global build mode setting |
| runInBackground | boolean | Whether to run builds in background |
| _notifiedForDisconnect | boolean | Whether disconnect warning has been shown (prevents duplicate notifications) |
| panelOpen | boolean | Whether chat/lua/todo panel is visible |
| messages | ChatMessage[] | Chat message history |
| chatInput | string | Current chat input text |
| _imageAttachments | ImageAttachment[] | Pending image attachments for next chat send |
| imageUploadData | string | JS-to-Lua bridge: JSON payload from drop/paste (cleared after processing) |
| lightboxUri | string | Full-resolution data URI for the lightbox preview (empty = hidden) |
| panelMode | string | "chat" or "lua" (bottom panel tab) |
| luaOutputLines | OutputLine[] | Lua console output history |
| luaInput | string | Current Lua code input |
| _luaInputFocusTrigger | string | JS code to focus Lua input (changes trigger ui-code) |
| todos | TodoItem[] | Claude Code todo list items |
| todosCollapsed | boolean | Whether todo column is collapsed |
| _todoSteps | table[] | Step definitions for createTodos/startTodoStep |
| _currentStep | number | Current in_progress step (1-based), 0 if none |
| _todoApp | string | App name for progress reporting during todo steps |
| showUpdatePrefDialog | boolean | Whether first-run update preference dialog is visible |
| showUpdateConfirmDialog | boolean | Whether update confirmation dialog is visible |
| latestVersion | string | Latest version string from GitHub releases API |
| _isUpdating | boolean | Whether an update is currently in progress |
| _updateNotificationDismissed | boolean | Whether the update notification banner was dismissed this session |
| _needsUpdate | boolean | Whether a newer version is available |

## Methods

### MCP (added to global mcp)

| Method | Description |
|--------|-------------|
| availableApps() | Returns _availableApps for binding |
| toggleMenu() | Toggle menuOpen state |
| closeMenu() | Set menuOpen to false |
| menuHidden() | Returns not menuOpen (for ui-class-hidden) |
| selectApp(name) | Call mcp:display(name), close menu |
| scanAvailableApps() | Scan apps/ directory for available apps |
| pollingEvents() | Server-provided: true if agent is connected to /wait endpoint |
| waitTime() | Server-provided: seconds since last agent connection to /wait |
| isWaiting() | Returns true if waitTime() > 0 (for ui-class-waiting binding) |
| pendingEventCount() | Server-provided: number of events waiting to be processed |
| waitStartOffset() | Returns UNIX timestamp when wait started, or 0 if connected (for client-side counter) |
| checkDisconnectNotify() | Check if Claude appears disconnected and show warning notification if needed |
| notify(message, variant) | Show a notification toast (variant: danger, warning, success, primary, neutral) |
| notifications() | Returns _notifications for binding |
| dismissNotification(n) | Remove notification from list |
| variablesLinkHtml() | Returns cached HTML anchor for /variables endpoint (opens in new tab) |
| helpLinkHtml() | Returns cached HTML anchor for /api/resource/ (opens in new tab) |
| openTools() | Display app-console and select the current app |
| currentAppName() | Returns kebab-case name of current app from mcp.value.type |
| currentAppHasCheckpoints() | Returns true if current app has checkpoints (via appConsole:findApp) |
| currentAppNoCheckpoints() | Returns not currentAppHasCheckpoints() |
| toolsTooltip() | Returns "Go to App - fast coded" if checkpoints, else "Go to App" |
| toggleBuildMode() | Toggle between "fast" and "thorough" modes |
| isFastMode() | Returns true if buildMode is "fast" |
| isThoroughMode() | Returns true if buildMode is "thorough" |
| buildModeTooltip() | Returns tooltip text for current mode |
| toggleBackground() | Toggle between foreground and background execution |
| isBackground() | Returns true if runInBackground is true |
| isForeground() | Returns true if runInBackground is false |
| backgroundTooltip() | Returns tooltip text for current execution mode |
| togglePanel() | Toggle panelOpen state |
| panelHidden() | Returns not panelOpen |
| panelIcon() | Returns "chat-dots-fill" if open, "chat-dots" if closed |
| showChatPanel() | Set panelMode to "chat" |
| showLuaPanel() | Set panelMode to "lua" |
| notChatPanel() | Returns panelMode ~= "chat" |
| notLuaPanel() | Returns panelMode ~= "lua" |
| chatTabVariant() | Returns "primary" if chat, else "default" |
| luaTabVariant() | Returns "primary" if lua, else "default" |
| sendChat() | Send chat event with current app as target; includes `images` array if attachments present |
| processImageUpload() | Bridge trigger: parse JSON from imageUploadData, decode base64, write to storage/uploads/, add to _imageAttachments |
| imageAttachments() | Returns _imageAttachments for binding |
| hasImages() | Returns true if _imageAttachments is non-empty |
| noImages() | Returns true if _imageAttachments is empty (for ui-class-hidden) |
| removeAttachment(att) | Remove attachment from list, delete file |
| clearAttachments() | Remove all attachments, delete files |
| lightboxVisible() | Returns true if lightboxUri is non-empty |
| hideLightbox() | Clears lightboxUri to close the lightbox |
| addAgentMessage(text) | Add agent message to chat, clear statusLine/statusClass |
| addAgentThinking(text) | Add thinking message to chat, update statusLine/statusClass |
| clearChat() | Clear messages list |
| clearPanel() | Clear chat or lua output based on panelMode |
| runLua() | Execute luaInput, append output to luaOutputLines |
| clearLuaOutput() | Clear luaOutputLines |
| focusLuaInput() | Set _luaInputFocusTrigger to JS that focuses input |
| setTodos(todos) | Replace todo list with new items (legacy API) |
| toggleTodos() | Toggle todosCollapsed state |
| hasTodos() | Returns true if todos is non-empty |
| createTodos(steps, appName) | Create todos from step labels with progress percentages |
| startTodoStep(n) | Complete previous step, start step n, update progress/thinking |
| completeTodos() | Mark all steps complete, clear progress bar |
| clearTodos() | Clear todos list and reset step state |
| appProgress(name, progress, stage) | Update app build progress (delegates to appConsole:onAppProgress) |
| appUpdated(name) | Rescan apps and delegate to appConsole:onAppUpdated |
| readSettings() | Read and parse `.ui/storage/settings.json`, returns empty table if missing/invalid |
| writeSettings(data) | Write settings table as JSON to `.ui/storage/settings.json`, creates storage/ dir if needed |
| checkForUpdates() | Fetch latest version from GitHub, compare with current, persist result to settings |
| showUpdatePreferenceDialog() | Show the first-run update preference dialog |
| setUpdatePreference(enabled) | Save checkUpdate preference to settings; triggers checkForUpdates() if enabled |
| getUpdatePreference() | Read checkUpdate boolean from settings |
| currentVersion() | Returns version string from mcp:status().version |
| noUpdateAvailable() | Returns not _needsUpdate |
| updateAvailable() | Returns _needsUpdate |
| hideUpdateNotification() | Returns true if notification should be hidden (no update, dismissed, or updating) |
| dismissUpdateNotification() | Set _updateNotificationDismissed to true |
| startUpdate() | Show the update confirmation dialog |
| cancelUpdate() | Hide the update confirmation dialog |
| confirmUpdate() | Emit pushState event with platform detection and update instructions, set _isUpdating |
| isUpdating() | Returns _isUpdating |
| notUpdating() | Returns not _isUpdating |

### MCP.Notification (notification toast)

| Field | Type | Description |
|-------|------|-------------|
| message | string | Notification text |
| variant | string | Shoelace alert variant (danger, warning, success, primary, neutral) |
| _mcp | ref | Reference to mcp for dismiss callback |

| Method | Description |
|--------|-------------|
| dismiss() | Calls mcp:dismissNotification(self) |

### MCP.AppMenuItem (wrapper for app info)

| Field | Type | Description |
|-------|------|-------------|
| _name | string | App directory name |
| _iconHtml | string | HTML content from icon.html |
| _mcp | ref | Reference to mcp for callbacks |

| Method | Description |
|--------|-------------|
| name() | Returns the app name |
| iconHtml() | Returns the icon HTML content |
| select() | Calls mcp:selectApp(self._name) |

### MCP.ChatMessage (chat message)

| Field | Type | Description |
|-------|------|-------------|
| sender | string | "You" or "Agent" |
| text | string | Message content |
| style | string | "normal" (default) or "thinking" for interstitial progress |
| _thumbnails | ChatThumbnail[] | Thumbnail images attached to this message |

| Method | Description |
|--------|-------------|
| new(sender, text, style, thumbnails) | Create a new ChatMessage with optional thumbnail list |
| isUser() | Returns true if sender == "You" |
| isThinking() | Returns true if style == "thinking" |
| hasThumbnails() | Returns true if _thumbnails is non-empty |
| noThumbnails() | Returns true if _thumbnails is empty |
| chatThumbnails() | Returns _thumbnails for binding |
| mutate() | Initialize style and _thumbnails if nil (hot-load migration) |
| prefix() | Returns "> " for user messages, "" for agent |

### MCP.ChatThumbnail (image thumbnail in chat message)

| Field | Type | Description |
|-------|------|-------------|
| uri | string | Thumbnail data URI (150px JPEG) for display |
| fullUri | string | Full-resolution data URI for lightbox preview |
| filename | string | Original filename |

| Method | Description |
|--------|-------------|
| showFull() | Sets mcp.lightboxUri to fullUri (or uri as fallback) to open lightbox |

### MCP.TodoItem (build progress item)

| Field | Type | Description |
|-------|------|-------------|
| content | string | Task description (shown for pending/completed) |
| status | string | "pending", "in_progress", or "completed" |
| activeForm | string | Present tense form (shown for in_progress) |

| Method | Description |
|--------|-------------|
| displayText() | Returns activeForm if in_progress, else content |
| isPending() | Returns status == "pending" |
| isInProgress() | Returns status == "in_progress" |
| isCompleted() | Returns status == "completed" |
| statusIcon() | Returns "🔄" for in_progress, "⏳" for pending, "✓" for completed |

### MCP.ImageAttachment (pending image for chat)

| Field | Type | Description |
|-------|------|-------------|
| path | string | File path on disk (in storage/uploads/) |
| filename | string | Original filename from drop/paste |
| thumbnailUri | string | Small data URI (JPEG, max 150px) for preview display |
| fullUri | string | Full-resolution data URI for lightbox preview |
| _mcp | ref | Reference to mcp for remove callback |

| Method | Description |
|--------|-------------|
| remove() | Calls mcp:removeAttachment(self) |

### MCP.OutputLine (Lua console output)

| Field | Type | Description |
|-------|------|-------------|
| text | string | Line content |

| Method | Description |
|--------|-------------|
| copyToInput() | Copy text to mcp.luaInput, focus input |

## Local Helper Functions

Module-level functions (not methods on mcp):

| Function | Description |
|----------|-------------|
| compareVersions(current, latest) | Semver comparison: parses "major.minor.patch" from both strings, returns true if latest > current |
| fetchLatestVersion() | Runs `curl` against `https://api.github.com/repos/zot/frictionless/releases/latest` with 5s connect / 10s max timeout, parses JSON for `tag_name`, strips leading "v", returns version string or nil |

## Update Check System

### Settings Persistence

`mcp:readSettings()` and `mcp:writeSettings(data)` provide shared settings storage at `{base_dir}/storage/settings.json`. They are defined on the global `mcp` object so any app can use them (e.g., the prefs app reads/writes the `checkUpdate` preference).

- `readSettings()` opens the file, JSON-decodes it, returns the table (or `{}` on any failure)
- `writeSettings(data)` creates the `storage/` directory if needed via `mkdir -p`, then writes JSON

### First-Run Flow

During initialization (`if not session.reloading`):
1. Read settings via `mcp:readSettings()`
2. If `settings.checkUpdate == nil` (no preference saved yet), call `mcp:showUpdatePreferenceDialog()` to open the first-run dialog
3. The dialog has Yes/No buttons that call `mcp:setUpdatePreference(true/false)`
4. `setUpdatePreference` saves the preference and, if enabled, immediately calls `checkForUpdates()`

### Update Check Flow

When `settings.checkUpdate == true` on startup:
1. Restore cached state: if `settings.needsUpdate` is true, set `mcp._needsUpdate = true` and `mcp.latestVersion` from settings (so the star/notification appear immediately without waiting for network)
2. Call `mcp:checkForUpdates()` which:
   - Gets current version from `mcp:status().version`
   - Calls `fetchLatestVersion()` (curl to GitHub API)
   - Compares with `compareVersions(current, latest)`
   - Sets `mcp.latestVersion` and `mcp._needsUpdate`
   - Persists `needsUpdate` and `latestVersion` to settings

### Notification / Star / Dialog UI

**Orange star indicator** (`.mcp-update-star`):
- Positioned at top-right of menu button container (absolute, top: -4px, right: -4px)
- Shown when `updateAvailable()` returns true, hidden via `noUpdateAvailable()`
- Pulses via `star-pulse` CSS animation (opacity 1 → 0.6 → 1, 2s)
- Click calls `startUpdate()` to open the confirm dialog

**Update notification banner** (`.mcp-update-notification`):
- Fixed position below menu area (top: 72px, right: 12px)
- Uses `sl-alert` with variant "primary" and download icon
- Shows latest version number
- Hidden when `hideUpdateNotification()` returns true (no update OR dismissed OR updating)
- **Update** button calls `startUpdate()`, **Later** button calls `dismissUpdateNotification()`

**First-run preference dialog** (`sl-dialog`):
- Bound to `showUpdatePrefDialog` via `ui-attr-open`
- Footer buttons: No → `setUpdatePreference(false)`, Yes → `setUpdatePreference(true)`

**Update confirmation dialog** (`sl-dialog`):
- Bound to `showUpdateConfirmDialog` via `ui-attr-open`
- Body shows latest version and current version
- Footer buttons: Cancel → `cancelUpdate()`, Update → `confirmUpdate()`

### Update Execution Flow

`confirmUpdate()`:
1. Closes the dialog, sets `_isUpdating = true`, dismisses the notification
2. Detects platform via `uname -s` and architecture via `uname -m`
3. Emits a `pushState` event with:
   - `type = "update"`, `action = "perform-update"`
   - `currentVersion`, `latestVersion`
   - `platform`, `architecture`
   - `releaseUrl` pointing to the specific GitHub release tag
   - `instructions` with step-by-step binary download/replace/chmod/restart/update procedure

### Progress Indicator

While `_isUpdating` is true:
- The normal status bar text is hidden (`ui-class-hidden="isUpdating()"`)
- A replacement span shows "Updating to vX.Y.Z..." with an indeterminate `sl-progress-bar`
- This span is hidden when `notUpdating()` returns true

## Chat Panel Features

### Image Drag & Drop

The entire chat panel is a drop zone. Images can also be pasted via Ctrl+V.

**UX flow:**
1. User drags image over chat panel → border glow, "Drop image here" overlay
2. User drops → JS reads file via FileReader, generates thumbnail (max 150px JPEG) and keeps full data URI, sends JSON via `updateValue` bridge
3. Lua receives via `processImageUpload()` (priority=high trigger), decodes base64, writes to `{base_dir}/storage/uploads/img-{time}-{rand}.{ext}`
4. Thumbnail preview row appears above text input with [x] remove buttons
5. User types optional text, hits Send
6. `sendChat()` creates ChatThumbnail objects (with uri + fullUri), attaches to ChatMessage, includes `images: ["/path/to/file"]` in event, clears attachments (files persist for agent)
7. Chat output shows thumbnail images below the message text (not text labels)
8. Clicking a chat thumbnail calls `ChatThumbnail:showFull()` which sets `mcp.lightboxUri` to display the full-resolution image in a fixed overlay

**Bridge payload** (JSON via `imageUploadData`):
```json
{"filename": "screenshot.png", "mime": "image/png", "base64": "...", "thumbnail": "data:image/jpeg;base64,...", "fullUri": "data:image/png;base64,..."}
```

**Lightbox:**
- Fixed overlay (`.image-lightbox`) inside `.mcp-shell`, bound to `lightboxVisible()` / `hideLightbox()`
- Image source bound to `mcp.lightboxUri` — only sent to browser when a thumbnail is clicked
- Close via click on overlay (`ui-event-click="hideLightbox()"`) or Escape key (JS triggers overlay click)

**Viewdef structure:**
- Drop overlay: absolute-positioned div shown via CSS `.drag-over` class on panel
- Image preview: horizontal row of thumbnails above input, hidden when no images
- Bridge: hidden span with id `image-bridge` + `processImageUpload()?priority=low` trigger
- JS: `<script>` block with drag/drop, paste, FileReader, thumbnail generation, updateValue calls

### Resizable
- Drag handle at top edge (JavaScript mousedown/mousemove/mouseup)
- Min height: 120px, Max: 60vh, Initial: 220px
- CSS `flex-shrink: 0` prevents panel from being squished

### Auto-scroll
- Chat messages and Lua output use `scrollOnOutput` variable property
- New content automatically scrolls into view

### Todo Step Definitions

Hardcoded step definitions map labels to progress percentages:

```lua
local UI_THOROUGH_STEPS = {
    {label = "Read requirements", progress = 5, thinking = "Reading requirements..."},
    {label = "Requirements", progress = 10, thinking = "Updating requirements..."},
    {label = "Design", progress = 20, thinking = "Designing..."},
    {label = "Write code", progress = 40, thinking = "Writing code..."},
    {label = "Write viewdefs", progress = 60, thinking = "Writing viewdefs..."},
    {label = "Link and audit", progress = 85, thinking = "Auditing..."},
    {label = "Simplify", progress = 92, thinking = "Simplifying..."},
    {label = "Set baseline", progress = 98, thinking = "Setting baseline..."},
    -- Also includes Fast-prefixed variants
}
```

Unknown labels get auto-calculated percentages based on position.

## ViewDefs

| File | Type | Purpose |
|------|------|---------|
| MCP.DEFAULT.html | MCP | Shell with app view, chat panel, menu button, icon grid, notifications, status bar, update star/dialogs/notification/progress |
| MCP.AppMenuItem.list-item.html | MCP.AppMenuItem | Icon card with icon HTML and name below |
| MCP.Notification.list-item.html | MCP.Notification | Toast notification with message and close button |
| MCP.ChatMessage.list-item.html | MCP.ChatMessage | Chat message with prefix, text, and optional thumbnail gallery |
| MCP.ChatThumbnail.list-item.html | MCP.ChatThumbnail | Clickable thumbnail image (opens lightbox on click) |
| MCP.TodoItem.list-item.html | MCP.TodoItem | Todo item with status icon and text |
| MCP.OutputLine.list-item.html | MCP.OutputLine | Clickable Lua output line (copies to input) |
| MCP.ImageAttachment.list-item.html | MCP.ImageAttachment | Thumbnail preview with [x] remove button |

## Events

App switching is handled entirely in Lua via `mcp:display()`.

The chat panel's `sendChat()` sends events via `mcp.pushState()`:

```json
{"app": "contacts", "event": "chat", "text": "...", "images": ["/path/to/img.png"], "mcp_port": 37067, "note": "...", "reminder": "Show todos and thinking messages while working"}
```

The `images` field is an array of file paths, only present when the user attached images. The agent can read these files with its Read tool.

The `app` field is set to `currentAppName()` (the currently displayed app), so the event is routed to the correct app's design.md for handling. Falls back to `"app-console"` if no app is displayed.

## App Discovery (Lua)

On load, scan `{base_dir}/apps/` for directories containing `app.lua` (built apps only). Store names in `mcp._availableApps`.
