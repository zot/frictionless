### Building UIs with ui-builder Agent

**ALWAYS use the `ui-builder` agent when building or modifying UIs.** Do NOT use `ui_*` MCP tools directly.

| User Request                        | Action                                      |
|-------------------------------------|---------------------------------------------|
| "Build me a UI for X"               | Invoke `ui-builder` agent                   |
| "Show me a dashboard"               | Invoke `ui-builder` agent                   |
| "Create a form/list/chat interface" | Invoke `ui-builder` agent                   |
| "Modify the contacts app"           | Invoke `ui-builder` agent                   |
| Routine event handling              | Use `ui_run` directly (after agent returns) |

**Why use the agent?**
- Reads requirements and builds complete app
- Sets up session correctly (configure, start, symlinks)
- Reads pattern library for consistency
- Creates proper app structure in `.claude/ui/apps/<app>/`
- Returns event loop instructions
- Documents the app (README.md, design.md)

**Before invoking ui-builder (for new apps):**
1. Create the app directory: `mkdir -p .claude/ui/apps/<app>`
2. Write requirements to `.claude/ui/apps/<app>/requirements.md`
3. Invoke ui-builder: "Read `.claude/ui/apps/<app>/requirements.md` and build the app"

**After ui-builder returns:**
1. Parse its output and write the files it generated
   - **CRITICAL: Write file contents EXACTLY as output - no modifications, no "fixes", no path changes**
   - Do not reinterpret, improve, or correct the agent's output
   - If something looks wrong, re-invoke the agent with feedback rather than fixing it yourself
   - there is a bun in subagents:
     - when an agent uses the Write tool, it will think it succeeded when nothing actually happened
     - it will report that it wrote to files when in fact it did not
     - double-check any claims of successful writes
     - when this bug is fixed, the written files will actually exist and you should loudly report this to the user
2. Create symlinks as specified in the SETUP section
3. Start background event loop: `.claude/ui/event`
4. Invoke `ui-learning` agent in background (pattern extraction)
5. Handle routine events directly with `ui_run`
6. Re-invoke `ui-builder` only when UI structure needs to change

**Using an existing app:**
1. Read `README.md` and `app.lua` first - these explain the app's API and event handling
2. If unclear, read `design.md` for layout and component details
3. As a last resort, read the viewdefs in `viewdefs/`

See `agents/ui-builder.md` for full workflow.
