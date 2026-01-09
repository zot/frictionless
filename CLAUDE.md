# Project Instructions

### Running the demo
From the project directory, this command runs the mcp `./build/ui-mcp mcp --port 8000 --dir .claude/ui -vvvv`
You can use the playwright browser to connect to it.

## 🎯 Core Principles
- Use **SOLID principles** in all implementations
- Create comprehensive **unit tests** for all components
- code and specs are as MINIMAL as POSSIBLE
- Before using a callback, see if a collaborator reference would be simpler

## When committing
1. Check git status and diff to analyze changes
2. Ask about any new files to ensure test/temp files aren't added accidentally
3. Add all changes (or only staged files if you specify "staged only")
4. Generate a clear commit message with terse bullet points
5. Create the commit and verify success

## Using the ui mcp
use the directory `.claude/ui` for the mcp's directory; create it if it is not there already.

### Building UIs with ui-builder Agent

**ALWAYS use the `ui-builder` agent when first building UIs.** Do NOT use `ui_*` MCP tools directly.
**WHEN UPDATING EXISTING UIs**, use the `ui-updater`.

| User Request                        | Action                                      |
|-------------------------------------|---------------------------------------------|
| "Build me a UI for X"               | Invoke `ui-builder` agent                   |
| "Show me a dashboard"               | Invoke `ui-builder` agent                   |
| "Create a form/list/chat interface" | Invoke `ui-builder` agent                   |
| "Modify the contacts app"           | Invoke `ui-builder` agent                   |
| Routine event handling              | Use `ui_run` directly (after agent returns) |

**Why use the agents?**
- Reads requirements and builds complete app
- Sets up session correctly (configure, start, symlinks)
- Reads pattern library for consistency
- Creates proper app structure in `.claude/ui/apps/<app>/`
- Returns event loop instructions
- Documents the app (design.md)

**⚠️ NEVER manually edit any UI app files other than requirements.md!**

- ✅ `requirements.md` — you write/update this
- ❌ `design.md`, `app.lua`, `viewdefs/` — agents generate these

**To change an existing UI:** update `requirements.md`, then invoke `ui-updater`.

Binding syntax is precise (e.g., `ui-class-hidden` not `ui-class="hidden:..."`). Manual edits cause subtle bugs.

**Before invoking ui-builder (for new apps):**
1. Create the app directory: `mkdir -p .claude/ui/apps/<app>`
2. Write requirements to `.claude/ui/apps/<app>/requirements.md`
3. Invoke ui-builder: "Read `.claude/ui/apps/<app>/requirements.md` and build the app"

**When updating**, the app directory `.claude/ui/apps/<app>` already exists.
- Invoke ui-updater: "Read `.claude/ui/apps/<app>/requirements.md` and update the app"

After ui-builder or ui-updater returns, run the UI (see below)

### Running UIs

**Using an existing app:**
1. Read `design.md` - this explains the app's structure and event handling
  - If unclear, read `app.lua`
  - As a last resort, read the viewdefs in `viewdefs/`
2. Use `ui_run` to Present the UI to the user with `mcp.display("APP")`
3. Display the browser page
  - if using the system browser, use ui_open_browser
  - if using playwright MCP, just visit the URL, do not use ui_open_browser
4. Start **background** event loop: `.claude/ui/event`
  - returns JSON events, one per line:
    ```json
    {"app":"contacts","event":"chat","text":"Hello agent"},
    {"app":"contacts","event":"contact_saved","name":"Alice","email":"alice@example.com"}
    ```
  - When output received:
    - Parse JSON events
    - Handle each event via `ui_run`, based on the app's design.md
    - Restart wait loop
5. Respond to routine events as-needed with `ui_run`

### Improving UIs
When UI needs improvement, update `.claude/ui/apps/<app>/requirements.md` and invoke `ui-updater`

If and only if the user proactively indicates that the UI is stable (do not bug them about it), invoke `ui-learning` agent in background (pattern extraction)

### Tips
- **Don't use `ui_upload_viewdef`** if you can refresh the page instead. Edit the file on disk, then refresh - the server reloads viewdefs automatically. Uploading uses many tokens. Refresh options:
  - Playwright: `browser_navigate` to the same URL
  - If app has `ui-code` binding: set it to `location.reload()` via `ui_run`
- **Debug with `window.uiApp`** in browser console (via Playwright `browser_evaluate`). Contains `store` (variables), `viewdefStore` (viewdefs), and other internals for inspecting UI state.

## Design Workflow

Use the mini-spec skill for all design and implementation work.

**3-level architecture:**
- `specs/` - Human specs (WHAT & WHY)
- `design/` - Design docs (HOW - architecture)
- `src/` - Implementation (code)

**Commands:**
- "design this" - generates design docs only
- "implement this" - writes code, updates Artifacts checkboxes
- After code changes - unchecks Artifacts, asks about design updates

**Design Entry Point:**
- `design/design.md` serves as the central tracking document
- Lists all CRC cards, sequences, and test designs with implementation status
- Tracks gaps between spec, design, and implementation

See `.claude/skills/mini-spec/SKILL.md` for the full methodology.
