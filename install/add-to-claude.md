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
- Sets up session correctly (configure, start, symlinks)
- Reads pattern library for consistency
- Creates proper app structure in `.ui-mcp/apps/<app>/`
- Returns event loop instructions
- Documents the app (README.md, design.md)

**After ui-builder returns:**
1. Start background event loop: `./.ui-mcp/event`
2. Invoke `ui-learning` agent in background (pattern extraction)
3. Handle routine events directly with `ui_run`
4. Re-invoke `ui-builder` only when UI structure needs to change

See `agents/ui-builder.md` for full workflow.
