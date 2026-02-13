# Claude Panel Design

## Intent

A universal panel for Claude Code showing project status, quick actions, and collapsible tree sections for discovering Agents/Commands/Skills.

## Layout

```
+----------------------------------+
| Claude Panel                     |
+----------------------------------+
| [Commit] [Test] [Build]          |
+----------------------------------+
| Status: Ready                    |
| Branch: main                     |
| Changed: 3 files                 |
+----------------------------------+
| > Agents (4)                     |
|   - ui-builder                   |
|   - ui-learning                  |
| > Commands (12)                  |
| > Skills (3)                     |
+----------------------------------+
```

## Data Model

### ClaudePanel (root)
| Field           | Type          | Description                      |
|-----------------|---------------|----------------------------------|
| status          | string        | "Loading" or "Ready"             |
| branch          | string        | Current git branch               |
| changedFiles    | number        | Count of changed files           |
| sections        | TreeSection[] | Collapsible sections             |
| jsCode          | string        | JavaScript to execute in browser |

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

## Methods

### ClaudePanel
| Method               | Description                                       |
|----------------------|---------------------------------------------------|
| commitAction()       | Push action event for commit                      |
| testAction()         | Push action event for test                        |
| buildAction()        | Push action event for build                       |
| loadGitStatus()      | Load branch and changed file count                |
| discoverItems()      | Scan filesystem for agents/commands/skills        |

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
| ClaudePanel.DEFAULT.html   | Main single-column layout  |
| TreeSection.list-item.html | Collapsible section header |
| TreeItem.list-item.html    | Clickable tree item        |

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

## Parent Response Patterns

**On action event:**
- commit: Run git add and commit workflow
- test: Run project test suite
- build: Run project build command

**On invoke event:**
- agent: Invoke the named agent
- command: Execute the slash command
- skill: Apply the skill

## Styling

Inherits terminal aesthetic from MCP shell via CSS variables:
- Dark backgrounds: `--term-bg`, `--term-bg-elevated`, `--term-bg-panel`
- Orange accent: `--term-accent`, `--term-accent-glow`
- Typography: `--term-mono` (JetBrains Mono), `--term-sans` (Space Grotesk)
- Glow effects on interactive elements

Layout:
- Single column layout, fills available space
- Section headers: clickable, chevron icon rotates when collapsed
- Tree items: indented, hover shows `--term-bg-hover`

Components:
- Shoelace components with dark theme overrides (::part selectors)
- `ui-code="jsCode"` on container for dynamic JS execution

**Note:** Connection status is shown by MCP shell's unified indicator, not in this panel.
