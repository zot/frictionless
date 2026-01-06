# Claude Panel Requirements

## Purpose

A universal panel for Claude Code showing project status and quick actions.

## Layout

```
┌───────────────────────────────────────────────────────────────────────────────┐
│ ┌─────────────────────────────┐  ┌──────────────────────────────────────────┐ │
│ │ Claude Panel                │  │ Chat                                     │ │
│ ├─────────────────────────────┤  ├──────────────────────────────────────────┤ │
│ │ [Commit]  [Test]  [Build]   │  │ ┌──────────────────────────────────────┐ │ │
│ ├─────────────────────────────┤  │ │ Agent: How can I help?               │ │ │
│ │ Status: Ready               │  │ │ You: Hello                           │ │ │
│ │ Branch: main                │  │ │ Agent: Hi there!                     │ │ │
│ │ Changed: 3 files            │  │ │                                      │ │ │
│ ├─────────────────────────────┤  │ │                                      │ │ │
│ │ ▼ Agents (4)                │  │ │                                      │ │ │
│ │   • ui-builder              │  │ │                                      │ │ │
│ │   • ui-learning             │  │ └──────────────────────────────────────┘ │ │
│ │ ▶ Commands (12)            │  ├──────────────────────────────────────────┤ │
│ │ ▶ Skills (3)               │  │ [Type a message...              ] [Send] │ │
│ └─────────────────────────────┘  └──────────────────────────────────────────┘ │
├───────────────────────────────────────────────────────────────────────────────┤
│ ▼ Lua Console                                                                 │
│ ┌───────────────────────────────────────────────────────────────────────────┐ │
│ │ Output:                                                                   │ │
│ │ > print("hello")                                                          │ │
│ │ hello                                                                     │ │
│ │ > claudePanel.status                                                      │ │
│ │ Ready                                                                     │ │
│ └───────────────────────────────────────────────────────────────────────────┘ │
│ ┌───────────────────────────────────────────────────────────────────────────┐ │
│ │ local x = 1                                                               │ │
│ │ local y = 2                                                               │ │
│ │ print(x + y)                                                              │ │
│ │ _                                                                         │ │
│ └───────────────────────────────────────────────────────────────────────────┘ │
│                                                                   [Run] [Clear]│
└───────────────────────────────────────────────────────────────────────────────┘
```

- No close button (panel is always visible)
- Left side: Claude Panel (320px fixed width)
- Right side: Chat panel (flexible width, fills remaining space)
- Bottom: Collapsible Lua console (full width, collapsed by default)

## Quick Actions

| Button | Event                               | Description              |
|--------|-------------------------------------|--------------------------|
| Commit | `{event:"action", action:"commit"}` | Stage and commit changes |
| Test   | `{event:"action", action:"test"}`   | Run test suite           |
| Build  | `{event:"action", action:"build"}`  | Run build command        |

## Status Section

| Field   | Source                            | Description                        |
|---------|-----------------------------------|------------------------------------|
| Status  | Lua state                         | "Loading" or "Ready"               |
| Branch  | `git branch --show-current`       | Current git branch                 |
| Changed | `git status --porcelain \| wc -l` | Count of changed files             |
| Events  | `mcp.eventQueueSize()`            | Count of pending events for Claude |

Display "Events: N pending" when N > 0, otherwise hidden.

## Tree Sections

Three collapsible sections:

### Agents
- Scan `.claude/agents/*.md` for project agents
- Include built-in agents: general-purpose, Explore, Plan, commit, etc.
- Click fires `{event:"invoke", type:"agent", name:"..."}`

### Commands
- Built-in slash commands (hardcoded list)
- Scan `.claude/commands/*.md` for custom commands
- Click fires `{event:"invoke", type:"command", name:"..."}`

### Skills
- Scan `.claude/skills/*/` directories
- Click fires `{event:"invoke", type:"skill", name:"..."}`

## Chat Panel

- Messages area: scrollable list of chat messages
  - **Fixed height container** that fits within viewport (do not stretch off page)
  - Overflow scrolls vertically
  - **Auto-scrolls** to show newest messages (scrollOnOutput)
- Each message shows sender ("Agent" or "You") and text
- Input field at bottom (always visible, no button)
- **Pressing Enter sends message and clears input**
- Send fires `{event:"chat", text:"..."}`
- Parent Claude responds by calling `app:addAgentMessage(text)`

## Lua Console

Collapsible panel at the bottom for interactive Lua execution.

### Layout
- **Header**: Clickable "Lua Console" text with expand/collapse indicator
- **Output panel**: Scrollable area showing command history and results, auto-scrolls to bottom (scrollOnOutput)
- **Input panel**: 4-line textarea for entering Lua code
- **Buttons**: [Run] and [Clear] at bottom right

### Behavior
- Collapsed by default
- Click header to expand/collapse
- **Run button** or **Ctrl+Enter** executes code locally via `loadstring()`
- If input doesn't start with `return`, try prepending `return ` first (for expressions like `3+4`)
- If prepending `return` causes syntax error, execute original input (for statements like `x = 5`)
- Code executes in the app's Lua environment (has access to `claudePanel`, `mcp`, etc.)
- Executed code and results append to output panel
- Output prefixes commands with `>` and shows results/errors below
- **Input clears after successful execution** (on error, input is retained for correction)
- **Clicking a line in output** copies it to the input (for re-running or editing previous commands)
- **Clear button** clears the output panel
- **No event pushed to parent** - execution is entirely local

### Data Model

| Field          | Type    | Description                        |
|----------------|---------|------------------------------------|
| consoleExpanded | boolean | Whether console is expanded        |
| luaOutput      | string  | Accumulated output text            |
| luaInput       | string  | Current code in input textarea     |

### Methods

| Method          | Description                                              |
|-----------------|----------------------------------------------------------|
| toggleConsole() | Toggle consoleExpanded state                             |
| runLua()        | Execute luaInput via load(), capture result, append output |
| clearOutput()   | Clear luaOutput                                          |

## Events

Events pushed to parent Claude for handling:

| Event    | Payload                                                                         | When                    |
|----------|---------------------------------------------------------------------------------|-------------------------|
| `action` | `{app:"claude-panel", event:"action", action:"commit\|test\|build"}`            | Quick action clicked    |
| `invoke` | `{app:"claude-panel", event:"invoke", type:"agent\|command\|skill", name:"..."}` | Tree item clicked       |
| `chat`   | `{app:"claude-panel", event:"chat", text:"..."}`                                | User sends chat message |

**Note:** Lua console does NOT push events - it executes code locally.

## JavaScript Execution

The panel exposes a `jsCode` property bound via `ui-code` attribute, allowing dynamic JavaScript execution in the browser.

| Property | Purpose |
|----------|---------|
| `jsCode` | Set to JavaScript code string to execute in browser |

**Usage:**
```lua
claudePanel.jsCode = "document.querySelector('.quick-actions sl-button:nth-child(2)').style.display = 'none'"
```

This enables the parent Claude to dynamically manipulate the DOM, hide/show elements, or perform any browser-side operations.

## Behaviors

- Sections collapse/expand on header click
- Items within expanded sections are clickable
- Git status loads on startup via `io.popen`
- Agent/skill discovery via filesystem on startup
- Chat input clears after sending
- JavaScript execution via `jsCode` property
- Lua console collapsed by default, expands on header click
- Lua code execution via Run button or Ctrl+Enter

## Discovery Details

**Agents:**
- Project: `.claude/agents/*.md`
- User: `~/.claude/agents/*.md`
- Built-in: hardcoded list (general-purpose, Explore, Plan, commit, etc.)

**Commands:**
- Built-in: hardcoded list (/help, /clear, /compact, /commit, etc.)
- Project: `.claude/commands/*.md`
- User: `~/.claude/commands/*.md`

**Skills:**
- Project: `.claude/skills/*/` (directories with skill files)
- User: `~/.claude/skills/*/`
