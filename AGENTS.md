# Agent Architecture Discussion

This document explores the agent design for the UI Platform MCP—specifically how AI agents create, present, operate, and modify UIs for two-way communication with users.

---

## Architecture Decision: Single UI Agent

The architecture has two distinct agents with different roles:

```
┌─────────────────┐
│    AI Agent     │  ← Main intelligence (Claude, etc.)
│  (the client)   │    Decides when/what to communicate with user
└────────┬────────┘
         │ invokes with instructions
         ▼
┌─────────────────┐
│    UI Agent     │  ← MCP-based sub-agent
│  (the builder)  │    Builds/modifies UIs per instructions
└────────┬────────┘
         │ renders
         ▼
┌─────────────────┐
│   Browser UI    │  ← What the user sees and interacts with
└────────┬────────┘
         │ interactions (clicks, input)
         ▼
┌─────────────────┐
│  mcp.notify()   │  ← Flows back to AI Agent
└─────────────────┘
```

**Key insight**: The AI Agent is the *client* of the UI Agent. The AI Agent decides it needs to communicate something to the user, then instructs the UI Agent to build an appropriate interface. The user interacts with the resulting UI, and notifications flow back to the AI Agent.

**Decision**: A single UI Agent handles all UI concerns. The MCP's tools are already well-factored (`ui_configure`, `ui_run`, `ui_upload_viewdef`, etc.), so one agent can handle the workflow without becoming monolithic. If complexity demands it later, sub-agents can be extracted.

---

## The `.ui-mcp/` Directory

The user configures the MCP to use `.ui-mcp/` as the base directory. This directory serves multiple purposes:

### Current Structure

```
.ui-mcp/
├── html/           # Static HTML files (index.html, etc.)
├── viewdefs/       # Viewdef templates (*.html)
├── lua/            # Lua source files (main.lua, etc.)
├── log/            # Runtime logs (lua.log, etc.)
└── resources/      # Documentation served via ui:// protocol
```

### Complete Directory Structure

Based on our discussions, here is the consolidated directory structure:

```
.ui-mcp/
│
├── html/               # Static HTML files
│   └── index.html
│
├── viewdefs/           # Viewdef templates
│   └── *.html
│
├── lua/                # Lua source files
│   └── main.lua
│
├── log/                # Runtime logs (gitignored)
│   └── lua.log
│
├── design/             # UI layout specifications (prevents drift)
│   ├── ui-main.md          # Main layout spec
│   ├── ui-*.md             # Per-UI layout specs
│   └── history.md          # Change log of UI iterations
│
├── patterns/           # Reusable UI patterns (cross-session consistency)
│   ├── pattern-form.md
│   ├── pattern-list.md
│   ├── pattern-dialog.md
│   └── pattern-wizard.md
│
├── conventions/        # Established conventions (cross-session consistency)
│   ├── layout.md           # Spatial conventions
│   ├── terminology.md      # Standard labels and text
│   ├── interactions.md     # How interactions work
│   └── preferences.md      # User preferences (expressed + inferred)
│
├── library/            # Proven implementations
│   ├── viewdefs/           # Tested viewdef templates
│   └── lua/                # Tested Lua patterns
│
├── state/              # Persistent state (optional)
│   └── session.json        # Last session state
│
└── resources/          # Documentation served via ui:// protocol
```

**What's gitignored**: `log/`, `state/` (ephemeral runtime data)

**What's tracked**: `design/`, `patterns/`, `conventions/`, `library/` (accumulated knowledge)

### Design Files: Auto-Generated with Specific Schema

The UI Agent auto-generates `design/ui-*.md` files using a specific schema:

```markdown
# {UI Name} Layout

## Intent
{Why this UI exists, what it accomplishes}

## Layout
{ASCII art showing structure}

## Components
| Element | Binding | Notes |
|---------|---------|-------|
| ...     | ...     | ...   |

## Behavior
- {Interaction rules}
```

This schema ensures:
- Intent is documented (prevents drift)
- Visual structure is explicit (prevents layout changes)
- Components are inventoried (prevents accidental removal)
- Behavior is specified (prevents interaction changes)

---

## Agent Workflow: Two-Phase Approach

The UI Agent uses a **two-phase workflow**—not separate agents, but two distinct modes of operation. This mirrors the CRC framework's "design first, then implement" discipline.

### Why Not Multiple Agents?

Multiple agents would be warranted if we had:
- Long-running background tasks (we don't)
- Parallel independent work (we don't)
- Different capability requirements (we don't—it's all file reading + code gen + tool calls)

Instead, a single agent with a disciplined workflow.

### Phase 1: Design

**Goal**: Understand context, plan changes, update specifications.

```
┌─────────────────────────────────────────────────┐
│                 DESIGN PHASE                     │
├─────────────────────────────────────────────────┤
│                                                  │
│  1. Read context                                 │
│     ├── patterns/*.md (how things should look)  │
│     ├── conventions/*.md (established rules)    │
│     └── design/ui-*.md (existing layouts)       │
│                                                  │
│  2. Plan the UI                                  │
│     ├── Which pattern applies? (form/list/etc)  │
│     ├── What components are needed?             │
│     └── What behaviors are required?            │
│                                                  │
│  3. Update design spec                           │
│     └── Create/modify design/ui-{name}.md       │
│         with Intent, Layout, Components,        │
│         Behavior sections                        │
│                                                  │
└─────────────────────────────────────────────────┘
```

**Key rule**: Do not write code until the design spec reflects the intended result.

### Phase 2: Build

**Goal**: Implement the design using MCP tools.

```
┌─────────────────────────────────────────────────┐
│                 BUILD PHASE                      │
├─────────────────────────────────────────────────┤
│                                                  │
│  1. Configure (if needed)                        │
│     └── ui_configure(base_dir=".ui-mcp")        │
│                                                  │
│  2. Implement                                    │
│     ├── Write Lua class matching design spec    │
│     ├── Write viewdef matching layout spec      │
│     ├── ui_run() to load Lua code               │
│     └── ui_upload_viewdef() to load template    │
│                                                  │
│  3. Present                                      │
│     ├── ui_start() to launch server             │
│     └── ui_open_browser() to show user          │
│                                                  │
│  4. Verify                                       │
│     └── Check that implementation matches spec  │
│                                                  │
└─────────────────────────────────────────────────┘
```

**Key rule**: Implementation must match the design spec. If it doesn't, fix the implementation or update the spec.

### Operating and Modifying

After initial build, the agent enters an operate/modify loop:

```
┌─────────────────────────────────────────────────┐
│              OPERATE/MODIFY LOOP                 │
├─────────────────────────────────────────────────┤
│                                                  │
│  On user interaction (mcp.notify):               │
│  ├── Process the notification                    │
│  ├── Update state via ui_run() if needed        │
│  └── Report back to AI Agent                    │
│                                                  │
│  On modification request from AI Agent:          │
│  ├── DESIGN PHASE: Update design spec first     │
│  └── BUILD PHASE: Implement the change          │
│                                                  │
└─────────────────────────────────────────────────┘
```

**Key rule**: Modifications always go through both phases. Never modify code without updating the spec first (or immediately after for exploratory changes).

---

## Preventing UI Drift

### The Problem

During long-running sessions with iterative UI changes ("vibe coding"), drift occurs:

1. AI Agent asks UI Agent to create UI with Feature A
2. User finds Feature A helpful (maybe AI didn't explicitly specify it)
3. AI Agent asks UI Agent to add Feature B
4. UI Agent modifies the UI, Feature A disappears or moves unexpectedly
5. User: "Where did that helpful thing go?"

The UI Agent lacks persistent memory of *what exists and why*. Each modification is made from current state + new instruction, without a specification to preserve.

### The Solution: ASCII Layout Specs

Borrow from the CRC framework's `ui-*.md` files—ASCII layouts that document UI structure:

```
.ui-mcp/
├── design/
│   ├── ui-main.md          # Main layout specification
│   ├── ui-feedback-form.md # Feedback form layout
│   └── ui-settings.md      # Settings panel layout
```

Example `ui-feedback-form.md`:

```markdown
# Feedback Form Layout

## Intent
Collect user rating and comments. Submit triggers notification to AI Agent.

## Layout

┌─────────────────────────────────────┐
│ Feedback                        [×] │  <- title bar with close
├─────────────────────────────────────┤
│                                     │
│  How would you rate this?           │
│  ┌─────────────────────────────┐    │
│  │ ★ ★ ★ ★ ★  (1-5 stars)     │    │  <- rating widget
│  └─────────────────────────────┘    │
│                                     │
│  Comments (optional):               │
│  ┌─────────────────────────────┐    │
│  │                             │    │
│  │                             │    │  <- text area
│  └─────────────────────────────┘    │
│                                     │
│  [Cancel]              [Submit]     │  <- action buttons
│                                     │
└─────────────────────────────────────┘

## Components

| Element     | Binding           | Notes                    |
|-------------|-------------------|--------------------------|
| Title       | static            | "Feedback"               |
| Close [×]   | on-click="close"  | Dismisses without submit |
| Stars       | ui-value="rating" | 1-5, default 3           |
| Comments    | ui-value="comment"| Optional text            |
| Cancel      | on-click="close"  | Dismisses without submit |
| Submit      | on-click="submit" | Triggers mcp.notify      |

## Behavior

- Submit disabled until rating selected
- Close/Cancel restore previous state (don't save partial input)
- Submit calls `self:submit()` which fires `mcp.notify("feedback", {...})`
```

### How This Prevents Drift

1. **Before modifying UI**: UI Agent reads the layout spec
2. **During modification**: UI Agent updates spec first, then code
3. **After modification**: Spec reflects new reality

The spec is the *source of truth*. If AI Agent says "add a 'skip' button", the UI Agent:
1. Reads `ui-feedback-form.md`
2. Adds `[Skip]` to the layout and components table
3. Updates the viewdef and Lua to match

If the spec says the close button exists, the UI Agent won't accidentally remove it.

### Spec-First vs Code-First

Two workflows the UI Agent can follow:

**Spec-First** (recommended for planned changes):
1. AI Agent gives instruction to UI Agent
2. UI Agent updates layout spec
3. UI Agent modifies viewdef/Lua to match spec
4. UI Agent verifies implementation matches spec

**Code-First** (for quick/exploratory changes):
1. AI Agent gives instruction, UI Agent makes quick change
2. AI Agent reviews result (via screenshot or state inspection)
3. If good: UI Agent updates spec to reflect new reality
4. If not: UI Agent reverts change

### Integration with CRC Framework

These UI specs are analogous to:
- `ui-*.md` in CRC → `ui-*.md` in `.ui-mcp/design/`
- CRC cards → Component definitions in layout specs
- Sequence diagrams → Behavior sections describing interactions

The UI Agent should treat `.ui-mcp/design/ui-*.md` files the same way the CRC framework treats design files:
- Read before modifying
- Update when structure changes
- Use as source of truth for what exists

---

## Cross-Session Consistency

### The Problem

Over many sessions, the AI creates many different UIs. Without shared conventions:
- A "feedback form" looks different each time
- Close buttons appear in different corners
- Submit buttons have different labels ("OK", "Submit", "Done", "Send")
- Lists scroll/paginate differently
- The user must re-learn each UI

Users develop expectations. Consistency reduces cognitive load and builds trust.

### The Solution: Design System in `.ui-mcp/`

```
.ui-mcp/
├── patterns/           # Reusable UI patterns
│   ├── pattern-form.md
│   ├── pattern-list.md
│   ├── pattern-dialog.md
│   └── pattern-wizard.md
│
├── conventions/        # Established conventions
│   ├── layout.md           # Spatial conventions
│   ├── components.md       # Component behavior conventions
│   ├── interactions.md     # How interactions work
│   └── terminology.md      # Standard labels and text
│
└── library/            # Proven implementations
    ├── viewdefs/           # Tested viewdef templates
    └── lua/                # Tested Lua patterns
```

### Pattern Files

`patterns/pattern-form.md`:

```markdown
# Form Pattern

## Structure

┌─────────────────────────────────────┐
│ {title}                         [×] │
├─────────────────────────────────────┤
│                                     │
│  {fields...}                        │
│                                     │
├─────────────────────────────────────┤
│  [Cancel]              [{primary}]  │  <- action bar always at bottom
└─────────────────────────────────────┘

## Conventions

- Title bar: title left, close button right
- Fields: label above input, full width
- Action bar: cancel left, primary action right
- Primary button: affirmative verb ("Submit", "Save", "Send")
- Cancel: always labeled "Cancel", never "Close" or "Back"

## Keyboard

- Enter in last field → submit (if valid)
- Escape → cancel
- Tab → next field

## Validation

- Inline errors below fields
- Primary button disabled until valid
- Error messages in red, start with field name
```

### Convention Files

`conventions/layout.md`:

```markdown
# Layout Conventions

## Window Chrome

- Close button: always top-right, always [×]
- Title: always top-left, sentence case
- No minimize/maximize (these are ephemeral UIs)

## Action Placement

- Primary action: bottom-right
- Cancel/dismiss: bottom-left
- Destructive actions: require confirmation

## Spacing

- Consistent padding: 16px edges, 8px between elements
- Group related fields visually

## Responsive

- Single column for narrow views
- Max width 600px for forms
```

`conventions/terminology.md`:

```markdown
# Terminology Conventions

## Button Labels

| Action          | Label      | Never Use          |
|-----------------|------------|--------------------|
| Confirm/proceed | "OK"       | "Okay", "Yes"      |
| Submit form     | "Submit"   | "Send", "Go"       |
| Save changes    | "Save"     | "Done", "Finish"   |
| Cancel          | "Cancel"   | "Close", "Back", "Exit" |
| Delete          | "Delete"   | "Remove", "Trash"  |
| Add item        | "Add"      | "Create", "New", "+" |

## Messages

- Confirmations: "Are you sure you want to {action}?"
- Success: "{Thing} {action}ed." (e.g., "Settings saved.")
- Error: "Couldn't {action}. {reason}."

## Placeholders

- Input hints: "Enter {thing}..."
- Empty states: "No {things} yet."
```

### How the UI Agent Uses This

**Before creating any UI**:
1. Read relevant `patterns/*.md` files
2. Read `conventions/*.md` files
3. Check `library/` for existing implementations

**When creating a new UI**:
1. Identify which pattern applies (form? list? dialog?)
2. Follow the pattern's structure
3. Apply conventions for layout, terminology, interactions
4. Copy from `library/` if a similar implementation exists

**When user expresses preference**:
1. Update relevant convention file
2. Example: User says "I prefer 'Done' over 'Submit'" → update `terminology.md`
3. Future UIs follow the updated convention

### Building the Design System Over Time

The design system grows organically:

1. **Session 1**: AI creates first form UI
   - UI Agent creates `patterns/pattern-form.md` documenting the structure
   - Saves working viewdef to `library/viewdefs/form-basic.html`

2. **Session 5**: User says "I liked how that list worked"
   - UI Agent documents the list in `patterns/pattern-list.md`
   - Notes what the user liked in `conventions/interactions.md`

3. **Session 12**: User says "Why does this form look different?"
   - UI Agent checks `patterns/pattern-form.md`, finds inconsistency
   - Fixes the form, reinforces convention

4. **Session 20**: New form needed
   - UI Agent reads patterns and conventions
   - Produces consistent form immediately
   - User's muscle memory works

### User Preference Tracking

`conventions/preferences.md`:

```markdown
# User Preferences

Learned from interactions. AI Agent updates this when user expresses preferences.

## Expressed Preferences

- 2024-01-15: "I prefer darker backgrounds" → added to style conventions
- 2024-01-18: "Always show me a cancel button" → added to form pattern
- 2024-02-01: "I don't like popups" → prefer inline UI over modal

## Inferred Preferences (tentative)

- User often resizes list views to be taller → may prefer more items visible
- User rarely uses keyboard shortcuts → de-emphasize keyboard hints
```

---

## Resolved Questions

1. **Single vs. multiple agents**: Single UI Agent, operated by the AI Agent as a sub-agent. (See Architecture Decision above.)

2. **Directory structure**: Defined. Includes `design/`, `patterns/`, `conventions/`, `library/` for accumulated knowledge. (See Complete Directory Structure above.)

3. **Design file format**: Auto-generated by UI Agent using specific schema (Intent, Layout, Components, Behavior sections).

4. **Version control**: Gitignore `log/` and `state/`. Track `design/`, `patterns/`, `conventions/`, `library/`.

5. **Persistence across sessions**: User preferences in `conventions/preferences.md`. Design knowledge in `patterns/` and `conventions/`. UI state optionally in `state/session.json`.

6. **Single-user**: This is a Claude CLI enhancement, not a multi-user system. One user per session.

7. **Security**: Localhost-only by default is sufficient. No immediate need for access controls.

8. **Resource limits**: Not needed at this time.

9. **Bootstrap**: Grow from scratch. Minimal bundle includes:
   - ViewList viewdefs (for array rendering)
   - Minimal app shell HTML
   - `main.lua` (in `web/lua/`, bundled via Makefile)

   The design system (`patterns/`, `conventions/`) grows organically as UIs are created.

---

## Next Steps

- [x] Decide on single vs. multiple agent approach
- [x] Define the directory structure for `.ui-mcp/`
- [ ] Create agent definition file (`.claude/agents/ui-builder.md`)
- [ ] Document workflow in resources
- [ ] Create initial pattern files (`pattern-form.md`, `pattern-list.md`)
- [ ] Create initial convention files (`layout.md`, `terminology.md`)

---

*See also: [PLAN.md](PLAN.md) for overall delivery roadmap*
