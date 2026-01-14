# Test Design: MCP Integration

**CRC Cards**: crc-MCPServer.md, crc-MCPTool.md
**Sequences**: seq-mcp-lifecycle.md, seq-mcp-run.md, seq-mcp-create-session.md, seq-mcp-create-presenter.md, seq-mcp-state-wait.md

### Test: MCP Server Lifecycle
**Purpose**: Verify the lifecycle behavior of the MCP server.

**Scenarios**:
1.  **Startup (Auto-Start)**:
    - Server starts with `--dir` parameter.
    - Verify auto-install runs if README.md missing.
    - Verify server starts and is ready to accept requests.
    - Verify port files: `{base_dir}/ui-port` and `{base_dir}/mcp-port` exist.
    - Verify port files contain valid port numbers.
    - Check filesystem: `log/` directory created.
    - Check Lua I/O: `print()` output goes to `log/lua.log`.
    - Call `ui_run` -> Expect execution success.

2.  **Reconfiguration (ui_configure while Running)**:
    - Start server with active session.
    - Call `ui_configure` with new base_dir -> Expect Success.
    - Verify old session destroyed, HTTP server stopped.
    - Verify new server started with new base_dir.
    - Verify server running with new URL.
    - Call `ui_run` -> Expect execution success (new session).

### Test: Tool - ui_open_browser
**Purpose**: Verify browser launch logic.

**Scenarios**:
1.  **Standard Launch (Default)**:
    - Call `ui_open_browser` (no args or minimal args).
    - Verify OS "open" command invoked with correct URL containing `?conserve=true`.
2.  **Specific Path**:
    - Call with `path="/my/view"`.
    - Verify URL ends with `/my/view?conserve=true`.
3.  **Explicit Disable**:
    - Call with `conserve=false`.
    - Verify URL does NOT contain `?conserve=true`.

### Test: Tool - ui_run
**Purpose**: Verify Lua execution capabilities.

**Scenarios**:
1.  **Execute Code**:
    - Call `ui_run` with `return 1 + 1`.
    - Expect result `2`.
2.  **Session Access**:
    - Call `ui_run` accessing `session` global.
    - Expect valid access to session variables.
3.  **JSON Marshalling**:
    - Return a table `{a=1, b="text"}`.
    - Expect JSON object `{ "a": 1, "b": "text" }`.
4.  **Non-JSON Result**:
    - Return a function or userdata.
    - Expect `{"non-json": "..."}` wrapper.

### Test: Tool - ui_upload_viewdef
**Purpose**: Verify dynamic view definition.

**Scenarios**:
1.  **Upload Viewdef**:
    - Call `ui_upload_viewdef` with valid HTML.
    - Verify viewdef is added to store.
2.  **Push Update**:
    - Connect a mock frontend.
    - Call `ui_upload_viewdef`.
    - Verify mock frontend receives "viewdef" message.
3.  **Variable Refresh**:
    - Create a session with a variable of the viewdef's type.
    - Call `ui_upload_viewdef`.
    - Verify "update" message sent for that variable.

### Test: Tool - ui_status
**Purpose**: Verify status reporting.

**Scenarios**:
1.  **Server Status**:
    - Server auto-starts on initialization.
    - Call `ui_status`.
    - Expect `base_dir` field with configured path.
    - Expect `url` field with valid URL pattern.
    - Expect `sessions` field with numeric value.
    - Expect `version` field with semver string.

### Test: MCP frictionless UI creation
**Purpose**: Verify end-to-end workflow for on-the-fly UI creation. This represents the core value proposition of the MCP integration: allowing an AI agent to build tiny collaborative apps to facilitate two-way communication and collaboration with the user.

**Scenarios**:
1.  **Define Presenter & View**:
    - Call `ui_run` to define a new Lua presenter class (e.g., `MyApp`).
    - Call `ui_upload_viewdef` to provide the HTML template for `MyApp`.
2.  **Instantiate & Display**:
    - Call `ui_run` to instantiate `MyApp` and attach it to the session root (e.g., `app.agent_display = MyApp()`).
    - Verify via `ui_run` (inspection) that the app is attached.
3.  **Verify Rendering**:
    - (Mock) Frontend receives update for `app`.
    - (Mock) Frontend requests `MyApp` viewdef.
    - (Mock) Frontend renders `MyApp` using the uploaded template.
4.  **User Interaction**:
    - Simulate user interaction on the frontend (e.g., user types "Hello" into a field and clicks a button).
    - Protocol message sent to backend to update variable state or call a method.
5.  **State Inspection**:
    - AI Agent calls `ui_run` to check the current state of `app.agent_display`.
    - Verify that the Lua object reflects the user's input (e.g., `app.agent_display.userInput == "Hello"`).
6.  **Iterative Refinement**:
    - AI Agent, seeing the user's input, calls `ui_upload_viewdef` with *modified* HTML to provide feedback or the next step in the workflow.
    - Verify frontend receives immediate push update and re-renders.

### Test: Installation Check (ui_configure)
**Purpose**: Verify configuration checks for installation and returns hint.
**Sequence**: seq-mcp-lifecycle.md (Scenario 1)

**Scenarios**:
1.  **Install Needed (Files Missing)**:
    - Start with empty project root (no `.claude/skills/` directory).
    - Call `ui_configure` with `base_dir=".claude/ui"`.
    - Verify response includes `install_needed: true`.
    - Verify response includes hint about running `ui_install`.

2.  **Install Not Needed (Files Present)**:
    - Pre-create `.claude/skills/ui-builder/SKILL.md`.
    - Call `ui_configure` with `base_dir=".claude/ui"`.
    - Verify response does NOT include `install_needed: true`.

### Test: Bundled File Installation (ui_install)
**Purpose**: Verify bundled files are installed via ui_install tool.
**Sequence**: seq-mcp-lifecycle.md (Scenario 1a)

**Scenarios**:
1.  **Fresh Install (Files Missing)**:
    - Start with empty project root.
    - Server auto-starts (auto-install if README.md missing).
    - Call `ui_install` explicitly if needed.
    - Verify all bundled files created:
      - `{project}/.claude/skills/*`
      - `{base_dir}/resources/*`
      - `{base_dir}/viewdefs/*`
      - `{base_dir}/event`, `state`, `variables` scripts
    - Verify response lists installed files.

2.  **No-Op (Files Exist, force=false)**:
    - Pre-create bundled files with custom content.
    - Call `ui_install` without force.
    - Verify file content unchanged.
    - Verify response lists files as skipped.

3.  **Force Overwrite (force=true)**:
    - Pre-create bundled files with custom content.
    - Call `ui_install` with `force=true`.
    - Verify files overwritten with bundled content.
    - Verify response lists files as installed.

4.  **Path Resolution**:
    - Set `base_dir="/project/.claude/ui"`.
    - Call `ui_install`.
    - Verify project files installed to `/project/.claude/`.
    - Verify base_dir files installed to `/project/.claude/ui/`.

### Test: MCP initialization
**Purpose**: Verify MCP server setup
- initialize() called by MCP client
- CRC: crc-MCPServer.md - "Does: initialize"
- Sequence: seq-mcp-lifecycle.md

### Test: State Change Waiting
**Purpose**: Verify HTTP long-poll mechanism for UI-to-agent communication with queue semantics.
**Sequence**: seq-mcp-state-wait.md

**Scenarios**:
1.  **Wait Success (Single Event)**:
    - Start server.
    - Make GET request to `/wait?timeout=5`.
    - In another goroutine, push event via `mcp.pushState({app="test", event="click"})`.
    - Verify request returns 200 with JSON array `[{"app":"test","event":"click"}]`.

2.  **Wait Timeout (Empty Queue)**:
    - Start server.
    - Make GET request to `/wait?timeout=1`.
    - Do not push any events to mcp.state.
    - Verify request returns 204 No Content after ~1 second.

3.  **No Active Session**:
    - (N/A - server always has an active session after startup)

4.  **Multiple Waiters**:
    - Start server.
    - Make two concurrent GET requests to `/wait?timeout=10`.
    - Push event via `mcp.pushState({app="test", event="broadcast"})`.
    - Verify BOTH requests return 200 with the same JSON array.

5.  **Client Disconnect**:
    - Start server.
    - Make GET request to `/wait?timeout=30`.
    - Cancel request (client disconnect) before timeout.
    - Verify server cleans up waiter without error.
    - Verify subsequent wait requests work normally.

6.  **Multiple Events Accumulated**:
    - Start server.
    - Make GET request to `/wait?timeout=10`.
    - Push two events in sequence:
      - `mcp.pushState({app="c", event="btn", id="save"})`
      - `mcp.pushState({app="c", event="btn", id="cancel"})`
    - Verify request returns 200 with array containing both events in order.

7.  **Atomic Queue Swap**:
    - Push event before waiting: `mcp.pushState({app="pre", event="queued"})`.
    - Make GET request to `/wait?timeout=5`.
    - Verify request returns immediately with the pre-queued event.
    - Verify mcp.state is now empty after response.
    - Push new event: `mcp.pushState({app="post", event="new"})`.
    - Make another GET request.
    - Verify only the new event is returned (not the old one).

8.  **Events Queued Before Wait**:
    - Push event: `mcp.pushState({app="x", event="early"})`.
    - Make GET request to `/wait?timeout=10`.
    - Verify request returns immediately (does not wait for timeout).
    - Verify response contains the queued event.

9.  **Timeout Parameter Bounds**:
    - Request with `timeout=0` -> Use default (30s).
    - Request with `timeout=200` -> Clamp to max (120s).
    - Request with `timeout=-5` -> Use default (30s).

10. **App Field in Events**:
    - Push events from different "apps":
      - `mcp.pushState({app="contacts", event="select", id=1})`
      - `mcp.pushState({app="chat", event="message", text="hi"})`
    - Verify both events returned with correct app fields.

### Test: mcp.pushState() Lua API
**Purpose**: Verify mcp.pushState() function and queue behavior.
**Sequence**: seq-mcp-state-wait.md

**Scenarios**:
1.  **Initial State (Empty Queue)**:
    - Verify `mcp.state` is `{}` (empty table) on session start.
    - Verify `#mcp.state == 0`.

2.  **Push Single Event**:
    - Execute `mcp.pushState({app="test", event="click"})`.
    - Verify `#mcp.state == 1`.
    - Verify `mcp.state[1].app == "test"`.
    - Verify any waiting HTTP clients are notified.

3.  **Push Multiple Events**:
    - Execute `mcp.pushState({app="a", event="e1"})`.
    - Execute `mcp.pushState({app="b", event="e2"})`.
    - Verify `#mcp.state == 2`.
    - Verify events are in insertion order.
    - Verify notification sent on each insert.

4.  **Read via Resource**:
    - Execute `mcp.pushState({app="test", event="value"})`.
    - Read `ui://state` MCP resource.
    - Verify resource returns current queue contents as JSON array.

5.  **Queue Cleared After Wait Response**:
    - Push events to queue via mcp.pushState().
    - Trigger wait response (via GET /wait).
    - Verify mcp.state is now empty `{}`.
    - Verify new events can be pushed to fresh queue.
