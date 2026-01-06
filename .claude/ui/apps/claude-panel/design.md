# Claude Panel Design

## Intent

A universal panel for Claude Code showing project status, quick actions, collapsible tree sections for discovering Agents/Commands/Skills, and a chat interface for interacting with the agent.

## Layout

```
+----------------------------------+------------------------------------+
| Claude Panel (320px)             | Chat                               |
+----------------------------------+------------------------------------+
| [Commit] [Test] [Build]          | +--------------------------------+ |
+----------------------------------+ | Agent: How can I help?         | |
| Status: Ready                    | | You: Hello                     | |
| Branch: main                     | | Agent: Hi there!               | |
| Changed: 3 files                 | |                                | |
+----------------------------------+ |                                | |
| > Agents (4)                     | +--------------------------------+ |
|   - ui-builder                   | +--------------------------------+ |
|   - ui-learning                  | | [Type message...]       [Send] | |
| > Commands (12)                  | +--------------------------------+ |
| > Skills (3)                     |                                    |
+----------------------------------+------------------------------------+
| v Lua Console                                                         |
+-----------------------------------------------------------------------+
| Output:                                                               |
| > print("hello")                                                      |
| hello                                                                 |
+-----------------------------------------------------------------------+
| local x = 1                                                           |
| print(x)                                                              |
|                                                          [Run] [Clear]|
+-----------------------------------------------------------------------+
```

## Data Model

### ClaudePanel (root)
| Field           | Type          | Description                      |
|-----------------|---------------|----------------------------------|
| status          | string        | "Loading" or "Ready"             |
| branch          | string        | Current git branch               |
| changedFiles    | number        | Count of changed files           |
| pendingEvents   | number        | Count of pending events for Claude (from mcp.eventQueueSize()) |
| sections        | TreeSection[] | Collapsible sections             |
| messages        | ChatMessage[] | Chat message history             |
| chatInput       | string        | Current chat input text          |
| jsCode          | string        | JavaScript to execute in browser |
| consoleExpanded | boolean       | Whether Lua console is expanded  |
| luaOutputLines  | OutputLine[]  | Console output as clickable lines|
| luaInput        | string        | Current code in input textarea   |

### TreeSection
| Field    | Type       | Description                        |
|----------|------------|------------------------------------|
| name     | string     | Section name (Agents/Commands/Skills) |
| expanded | boolean    | Whether section is expanded        |
| items    | TreeItem[] | Items in this section              |
| itemType | string     | "agent", "command", or "skill"     |

### TreeItem
| Field   | Type   | Description            |
|---------|--------|------------------------|
| name    | string | Item name              |
| section | ref    | Parent section ref     |

### ChatMessage
| Field  | Type   | Description          |
|--------|--------|----------------------|
| sender | string | "Agent" or "You"     |
| text   | string | Message content      |

### OutputLine
| Field  | Type   | Description                    |
|--------|--------|--------------------------------|
| text   | string | Line content                   |
| panel  | ref    | Reference to ClaudePanel       |

## Methods

### ClaudePanel
| Method               | Description                                       |
|----------------------|---------------------------------------------------|
| commitAction()       | Push action event for commit                      |
| testAction()         | Push action event for test                        |
| buildAction()        | Push action event for build                       |
| sendChat()           | Send chat message, push event, clear input        |
| addAgentMessage(t)   | Add agent response to messages                    |
| loadGitStatus()      | Load branch and changed file count                |
| discoverItems()      | Scan filesystem for agents/commands/skills        |
| toggleConsole()      | Toggle consoleExpanded state                      |
| runLua()             | Try "return " + input first, fallback to original, append result, clear input on success only |
| clearOutput()        | Clear luaOutput                                   |
| isConsoleCollapsed() | Return not consoleExpanded                        |
| copyToInput(line)    | Copy an output line to input for re-running       |

### OutputLine
| Method      | Description                              |
|-------------|------------------------------------------|
| copyToInput() | Copy this line's text to panel's luaInput |

### TreeSection
| Method          | Description                    |
|-----------------|--------------------------------|
| toggle()        | Toggle expanded state          |
| isCollapsed()   | Return not expanded            |
| itemCount()     | Return count with parens       |

### TreeItem
| Method   | Description                          |
|----------|--------------------------------------|
| invoke() | Push invoke event with type and name |

## ViewDefs

| File                       | Purpose                    |
|----------------------------|----------------------------|
| ClaudePanel.DEFAULT.html   | Main two-column layout     |
| TreeSection.list-item.html | Collapsible section header |
| TreeItem.list-item.html    | Clickable tree item        |
| ChatMessage.list-item.html | Chat message display       |
| OutputLine.list-item.html  | Clickable output line      |

## Events

### action
```json
{"app":"claude-panel","event":"action","action":"commit"}
{"app":"claude-panel","event":"action","action":"test"}
{"app":"claude-panel","event":"action","action":"build"}
```

### invoke
```json
{"app":"claude-panel","event":"invoke","type":"agent","name":"ui-builder"}
{"app":"claude-panel","event":"invoke","type":"command","name":"/help"}
{"app":"claude-panel","event":"invoke","type":"skill","name":"plantuml"}
```

### chat
```json
{"app":"claude-panel","event":"chat","text":"Hello agent"}
```

**Note:** Lua console executes code locally via `load()` - no event pushed to parent.

## Parent Response Patterns

**On action event:**
- commit: Run git add and commit workflow
- test: Run project test suite
- build: Run project build command

**On invoke event:**
- agent: Invoke the named agent
- command: Execute the slash command
- skill: Apply the skill

**On chat event:**
- Process message and respond with `app:addAgentMessage(response)`

## Styling

- Left panel: fixed 320px width
- Right panel: flexible, fills remaining space
- Section headers: clickable, show expand/collapse indicator
- Tree items: indented, hover highlight
- Chat messages: fixed height container (calc to fit viewport), vertical scroll overflow, auto-scroll on new messages
- Chat input: Enter key sends message (same as clicking Send button)
- Lua console: full width, collapsible, collapsed by default
- Lua output: monospace font, scrollable, max height ~150px, auto-scroll on new output, lines clickable to copy to input
- Lua input: monospace font, 4 lines, resizable textarea, clears on successful execution (retained on error)
- Ctrl+Enter in Lua input triggers Run
- Shoelace components for buttons, inputs, icons
- `ui-code="jsCode"` on container for dynamic JS execution
