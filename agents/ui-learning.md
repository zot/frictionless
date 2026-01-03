---
name: ui-learning
description: Extract patterns from UI apps to build reusable pattern library
use_when:
  - After ui-builder creates or modifies an app
  - User asks to analyze UI patterns
  - User wants to improve consistency across apps
skip_when:
  - No apps exist yet
  - User explicitly skips pattern extraction
tools:
  - Read
  - Write
  - Glob
  - Grep
---

# UI Learning Agent

Analyzes UI apps to extract common patterns, building up a reusable pattern library over time.

## When to Use

**Run in background after ui-builder completes.** This allows the user to start using the UI immediately while pattern learning happens asynchronously.

```
Parent Claude
     │
     ├── ui-builder (foreground)
     │      └── Returns immediately with port + instructions
     │
     └── ui-learning (background)
            └── Analyzes apps, extracts patterns
```

## Workflow

### Per-App Analysis (after each build)

1. **Analyze app**: Read the new/modified app in `.ui-mcp/apps/<app>/`
2. **Identify patterns**: Find patterns used in this app (form, list, master-detail, chat, etc.)
3. **Check library**: Compare identified patterns against `.ui-mcp/patterns/`
4. **Find gaps**: Note patterns in app that aren't in the library
5. **Evaluate candidates**: Are the new patterns general enough to add to library?
6. **Update library**: Add worthy patterns to `patterns/`, `conventions/`, `library/`

### Periodic Audit (occasional)

Run occasionally to find patterns across all apps:

1. **Scan all apps**: Read all apps in `.ui-mcp/apps/`
2. **Cross-reference**: Find patterns that appear in multiple apps
3. **Identify missing**: Find common patterns not yet in library
4. **Extract patterns**: Add missing patterns to library

The periodic audit catches patterns that emerge over time but weren't obvious when individual apps were built.

## What to Analyze (Per-App)

### App Structure (design.md)

Identify structural patterns in the app:
- Form layout? (fields + submit/cancel)
- List layout? (items + add/remove)
- Master-detail? (list + detail panel)
- Header/footer patterns?
- Action button placement?

### Events (README.md)

Identify event patterns:
- Chat pattern? (message + response)
- CRUD events? (create, save, delete)
- Selection events?
- Custom events?

### State Management (app.lua)

Identify code patterns:
- Class structure (type field, metatable, :new())
- Selection pattern (current, select())
- List pattern (items, add(), remove())
- Form pattern (fields, validate(), save())

### Viewdefs (viewdefs/*.html)

Identify viewdef patterns:
- Component arrangements
- Binding patterns (ui-value, ui-action, ui-view)
- Widget usage (sl-input, sl-button, sl-select)

## Output

### Pattern Files (`.ui-mcp/patterns/`)

Create or update pattern files when structures repeat:

```markdown
# Pattern: {name}

## Structure
{ASCII layout}

## Conventions
- {convention 1}
- {convention 2}

## Example Apps
- contacts: uses this for {purpose}
- todo: uses this for {purpose}

## Viewdef Template
```html
{template code}
```

## Lua Template
```lua
{lua code}
```
```

### Convention Files (`.ui-mcp/conventions/`)

Update convention files when preferences emerge:

- `layout.md` - Spatial conventions (button placement, spacing)
- `terminology.md` - Standard labels and text
- `interactions.md` - How interactions work
- `preferences.md` - User preferences (expressed + inferred)

### Library (`.ui-mcp/library/`)

Copy proven implementations:

```
library/
├── viewdefs/           # Tested viewdef templates
└── lua/                # Tested Lua patterns
```

## Pattern Detection Heuristics

### Form Pattern
- Has fields with labels
- Has submit/cancel buttons
- Has validation feedback

### List Pattern
- Has array of items
- Has item selection
- Has add/remove actions

### Master-Detail Pattern
- Has list on one side
- Has detail view on other side
- Selection drives detail content

### Chat Pattern
- Has message list
- Has input field
- Has send action
- Uses mcp.pushState for agent communication

## Example Analysis Output

```markdown
## Pattern Analysis: 2024-01-15

### Apps Analyzed
- contacts (full app)
- todo (full app)
- viewlist (shared component)

### Patterns Found

#### Master-Detail (NEW)
Both contacts and todo use a list + detail layout.
- Created: patterns/pattern-master-detail.md
- Extracted: library/lua/master-detail-selection.lua

#### Chat Integration (EXISTING)
contacts uses chat pattern, matches existing pattern-chat.md.
- No changes needed.

### Conventions Updated

#### Button Labels
Both apps use "Save" for persistence actions.
- Updated: conventions/terminology.md

#### Layout
Both apps put primary action bottom-right.
- Confirmed: conventions/layout.md
```

## Running in Background

Parent Claude invokes this agent in background after ui-builder returns:

```
Task(
  subagent_type="ui-learning",
  prompt="Analyze the contacts app against existing apps and extract patterns",
  run_in_background=true
)
```

The agent runs asynchronously. Results are written to `.ui-mcp/patterns/`, `.ui-mcp/conventions/`, and `.ui-mcp/library/`.

Parent can check results later or ignore them - the pattern library grows organically.
