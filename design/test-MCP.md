# Test Design: MCP Integration

**CRC Cards**: crc-MCPServer.md, crc-MCPTool.md
**Sequences**: seq-mcp-lifecycle.md, seq-mcp-run.md, seq-mcp-create-session.md, seq-mcp-create-presenter.md, seq-mcp-state-wait.md

### Test: MCP Server Lifecycle
**Purpose**: Verify the FSM behavior of the MCP server.

**Scenarios**:
1.  **Initial State (Unconfigured)**:
    - Verify server starts in UNCONFIGURED state.
    - Call `ui_start` -> Expect Error ("Server not configured").
    - Call `ui_run` -> Expect Error ("Server not configured").
    - Call `ui_configure` with valid path -> Expect Success.

2.  **Configured State**:
    - Post-configuration, verify state is CONFIGURED.
    - Check filesystem: `log/` directory created.
    - Check Lua I/O: `print()` output goes to `log/lua.log`.
    - Call `ui_run` -> Expect Error ("Server not started").
    - Call `ui_start` -> Expect Success (returns URL).

3.  **Running State**:
    - Verify state is RUNNING.
    - Call `ui_start` -> Expect Error ("Server already running").
    - Call `ui_run` -> Expect execution success.

4.  **Session Restart (ui_configure while Running)**:
    - Start in RUNNING state with active session.
    - Call `ui_configure` -> Expect Success (old session destroyed).
    - Verify state is CONFIGURED.
    - Call `ui_start` -> Expect Success (new session created).

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

### Test: Agent File Installation
**Purpose**: Verify bundled agent files are installed during configuration.
**Sequence**: seq-mcp-lifecycle.md (Scenario 1a)

**Scenarios**:
1.  **Fresh Install (File Missing)**:
    - Start with empty project root (no `.claude/agents/` directory).
    - Call `ui_configure` with `base_dir=".ui-mcp"`.
    - Verify `.claude/agents/ui-builder.md` is created.
    - Verify `agent_installed` notification sent with params:
      `{"file": "ui-builder.md", "path": ".claude/agents/ui-builder.md"}`

2.  **No-Op (File Exists)**:
    - Pre-create `.claude/agents/ui-builder.md` with custom content.
    - Call `ui_configure`.
    - Verify file content unchanged (not overwritten).
    - Verify NO `agent_installed` notification sent.

3.  **Directory Creation**:
    - Start with project root that has no `.claude/` directory.
    - Call `ui_configure`.
    - Verify `.claude/agents/` directory created.
    - Verify `ui-builder.md` file created inside.

4.  **Path Resolution**:
    - Set `base_dir="/project/.ui-mcp"`.
    - Call `ui_configure`.
    - Verify agent file installed to `/project/.claude/agents/ui-builder.md`.
    - (Ensures parent directory resolution is correct)

5.  **Reconfiguration (File Already Installed)**:
    - Call `ui_configure` (installs agent file, notification sent).
    - Call `ui_configure` again (session restart).
    - Verify NO second `agent_installed` notification.

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
    - Start server in RUNNING state.
    - Make GET request to `/wait?timeout=5`.
    - In another goroutine, push event via `mcp.pushState({app="test", event="click"})`.
    - Verify request returns 200 with JSON array `[{"app":"test","event":"click"}]`.

2.  **Wait Timeout (Empty Queue)**:
    - Start server in RUNNING state.
    - Make GET request to `/wait?timeout=1`.
    - Do not push any events to mcp.state.
    - Verify request returns 204 No Content after ~1 second.

3.  **No Active Session**:
    - Server in CONFIGURED state (not RUNNING).
    - Make GET request to `/wait?timeout=5`.
    - Verify request returns 404 Not Found.

4.  **Multiple Waiters**:
    - Start server in RUNNING state.
    - Make two concurrent GET requests to `/wait?timeout=10`.
    - Push event via `mcp.pushState({app="test", event="broadcast"})`.
    - Verify BOTH requests return 200 with the same JSON array.

5.  **Client Disconnect**:
    - Start server in RUNNING state.
    - Make GET request to `/wait?timeout=30`.
    - Cancel request (client disconnect) before timeout.
    - Verify server cleans up waiter without error.
    - Verify subsequent wait requests work normally.

6.  **Multiple Events Accumulated**:
    - Start server in RUNNING state.
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