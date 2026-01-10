## Using the ui mcp
Once a UI is built, run it, if that is the user's intent (see below)

### Running UIs

**Using an existing app:**
1. Read `design.md` - this explains the app's structure and event handling
  - If unclear, read `app.lua`
  - As a last resort, read the viewdefs in `viewdefs/`
2. Use the `ui_display("APP")` tool to present the UI to the user
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

### Tips
- **Don't use `ui_upload_viewdef`** because hotloading is enabled; just edit the file on disk.
- **Debug with `window.uiApp`** in browser console (via Playwright `browser_evaluate`). Contains `store` (variables), `viewdefStore` (viewdefs), and other internals for inspecting UI state.
