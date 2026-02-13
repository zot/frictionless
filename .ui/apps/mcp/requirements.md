# MCP

The outer shell for all frictionless apps. Displays the current app and provides navigation between apps.

## Architecture

**Uses the global `mcp` object directly.** This app does not create its own prototype — it renders the `mcp` object provided by the server. The `mcp` object has:
- `value` - the currently displayed app (set via `mcp:display(appName)`)
- `code` - JavaScript code to execute (for browser control)

## Layout

The shell is minimal chrome around the current app:
- Full viewport display of `mcp.value` (the current app)
- A 9-dot menu button in the top-right corner for app switching
- A status bar at the bottom showing `mcp.statusLine` in `mcp.statusColor`
- A hidden element with `ui-code="code"` for JavaScript execution

## App Switcher Menu

The 9-dot button (grid icon) in the top-right corner:
- Overlays the app content (always visible on top)
- Has a glow for easy visibility
- Opens a dropdown/popover menu with app icons
- Clicking an app calls `mcp:display(appName)` to switch
- Menu closes after selection

### Icon Grid Layout

Apps display as icons with names underneath:
- Each app's icon comes from its `icon.html` file (contains emoji, `<sl-icon>`, or `<img>`)
- App name displayed below each icon
- Icons arranged in rows of 3, Z formation (left-to-right, then next row)
- Clickable icon cards with hover effect

## Processing Indicator

Indicate processing state via a pulsating glow effect on the menu button:
- When `isWaiting()` returns true, the button enters a `.waiting` state
- The button border glows orange and pulses via CSS animation
- The grid icon dims to 30% opacity
- A wait time counter appears centered over the button
- The button remains clickable in waiting state
- When `isWaiting()` becomes false, the glow stops and counter disappears

### Wait Time Counter (Client-Local)

Display a counter showing seconds elapsed since waiting started:
- Entirely client-side JavaScript — no server round-trips needed
- JavaScript interval (200ms) reads timestamp from hidden element
- Counter displays as bold orange text with glow, centered in button
- Only shown when elapsed time exceeds 5 seconds
- Text clears when not waiting

### Busy Notification

When pushing an event via `mcp.pushState`:
- Idempotently override the global `pushState` function to add this behavior
- Before pushing, check `mcp:waitTime()`
- If waitTime exceeds 5 seconds, show a notification: "Claude might be busy. Use /ui events to reconnect."
- Use the "warning" variant for this notification
- Only show once per disconnect period (reset when Claude reconnects)

Additionally, on UI refresh, check if events are pending and Claude appears disconnected:
- If `waitTime() > 5` and there are pending events, show the same warning
- This catches interactions that don't go through pushState

## Available Apps

The menu should list apps discovered from the apps directory. Since mcp is Lua-driven, it should scan for available apps on load using the same filesystem pattern as the `apps` app.

## JavaScript Execution

A hidden element binds to `mcp.code` via `ui-code`:
- When `mcp.code` changes, the JavaScript is executed
- Used for browser control (close window, open URLs, etc.)
- The element is invisible (display: none or similar)

## Events

App switching is handled entirely in Lua via `mcp:display()`.

The chat panel's `sendChat()` sends `chat` events to Claude via `mcp.pushState()`, with the selected app context from `appConsole.selected` when available. The event uses `app: "app-console"` for compatibility with the app-console event handler.

## Status Bar

A status bar at the bottom of the viewport:
- Always visible
- Displays `mcp.statusLine` text with `mcp.statusClass` CSS class
- The `.thinking` class styles text as orange bold-italic
- Maintains consistent height even when empty
- Compact padding (6px horizontal)

### Status Bar Icons

Icons at the right edge of the status bar, grouped tightly together from left to right:

| Icon | Action | Description |
|------|--------|-------------|
| `{}` braces | variablesLinkHtml() | Opens `/variables` in new tab - shows variable tree |
| ❓ question mark | helpLinkHtml() | Opens `/api/resource/` in new tab - documentation |
| 🔧 tools | openTools() | Opens app-console, selects current app |
| 🚀/💎 | toggleBuildMode() | fast / thorough |
| ⏳/🔄 | toggleBackground() | foreground / background |
| 💬 chat-dots | togglePanel() | Toggle chat/lua/todo panel |

The braces and question mark icons use `ui-html` to generate anchor tags that open in new tabs. They are styled in purple (#bb88ff) with a brighter hover state (#dd99ff).

The chat-dots icon is the rightmost status bar item. It shows `chat-dots-fill` when open, `chat-dots` when closed.

### Build Settings Toggles

Two toggle icons on the right edge of the status bar, grouped tightly together:

| Toggle | Icons | Values |
|--------|-------|--------|
| Build mode | 🚀 rocket / 💎 diamond | fast / thorough |
| Execution | ⏳ hourglass / 🔄 arrows | foreground / background |

- Icons have minimal padding (3px horizontal) for a compact appearance
- Click toggles between states
- Hover shows tooltip describing current value with "(click to change)"

## Notifications

Agents can display notifications to alert users of important events (errors, warnings, info):

- `mcp:notify(message, variant)` - Show a notification toast
- `variant` can be: "danger" (red), "warning" (yellow), "success" (green), "primary" (blue), "neutral" (gray)
- Default variant is "danger" (most notifications are errors)
- Notifications appear as Shoelace alerts, auto-dismiss after 5 seconds
- Multiple notifications stack vertically
- Each notification has a close button for manual dismissal
- Notifications appear in top-right corner, below the menu button

## Chat Panel

A toggleable panel between the app content and status bar. The chat-dots icon in the status bar opens/closes it.

**Layout:**
- Resizable via drag handle at top edge (min 120px, max 60vh, initial 220px)
- Two columns: Todo List (left, 200px collapsible) | Chat/Lua (right)
- Messages auto-scroll to bottom on new output (`scrollOnOutput`)

**Chat tab:**
- Messages display with user messages prefixed with `>`
- Thinking messages shown italic/muted
- Input field + Send button (or Enter key)
- `sendChat()` reads `appConsole.selected` for app context when sending events

**Lua tab:**
- REPL for executing Lua code
- Output area shows command history and results
- Input textarea for multi-line code (Ctrl+Enter to run)
- Clicking an output line copies it to input and focuses

**Todo column:**
- Shows Claude's current build progress (⏳ pending, 🔄 in-progress, ✓ completed)
- Collapse button shrinks to 32px icon-only width
- Clear button removes all todos

### MCP Chat/Todo API

Methods available to Claude via `.ui/mcp run`:

```lua
mcp:addAgentMessage(text)        -- Add agent message, clear status
mcp:addAgentThinking(text)       -- Add thinking message, update status
mcp:createTodos(steps, appName)  -- Create todos from step labels
mcp:startTodoStep(n)             -- Complete previous, start step n
mcp:completeTodos()              -- Mark all complete, clear progress
mcp:appProgress(name, pct, stage) -- Update app build progress
mcp:appUpdated(name)             -- Trigger rescan after file changes
```

`createTodos` uses hardcoded step definitions mapping labels to progress percentages. Unknown labels get auto-calculated percentages.

`startTodoStep` also calls `appConsole:onAppProgress()` to update the build progress bar in the app list.

## Styling

### Terminal Aesthetic

The MCP shell uses a retro-futuristic terminal theme that all child apps inherit:

**Color Palette (CSS Variables):**
- `--term-bg`: #0a0a0f (deep dark background)
- `--term-bg-elevated`: #12121a (raised surfaces)
- `--term-bg-hover`: #1a1a24 (hover states)
- `--term-border`: #2a2a3a (subtle borders)
- `--term-text`: #e0e0e8 (primary text)
- `--term-text-dim`: #8888a0 (secondary text)
- `--term-text-muted`: #5a5a70 (tertiary text)
- `--term-accent`: #E07A47 (orange accent)
- `--term-accent-glow`: rgba(224, 122, 71, 0.4) (glow effect)
- `--term-success`: #4ade80 (green)
- `--term-warning`: #fbbf24 (yellow)
- `--term-danger`: #f87171 (red)

**Typography:**
- `--term-mono`: JetBrains Mono, Fira Code, Consolas (monospace)
- `--term-sans`: Space Grotesk, system-ui (headings/UI)

**Visual Effects:**
- Scan line overlay on shell background
- Glow effects on interactive elements
- Orange accent color for focused/active states

### Layout Guidelines

- No padding or margins around the app content
- The menu button should have subtle styling with hover glow
- Menu button should not interfere with app content interaction
- Status bar has terminal styling with prompt character (`>`)
