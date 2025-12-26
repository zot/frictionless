# Test Design: MCP Integration

**CRC Cards**: crc-MCPServer.md, crc-MCPTool.md
**Sequences**: seq-mcp-lifecycle.md, seq-mcp-run.md, seq-mcp-create-session.md, seq-mcp-create-presenter.md

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
    - Call `ui_configure` -> Expect Error ("Cannot reconfigure while running").
    - Call `ui_run` -> Expect execution success.

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

### Test: MCP initialization
**Purpose**: Verify MCP server setup
- initialize() called by MCP client
- CRC: crc-MCPServer.md - "Does: initialize"
- Sequence: seq-mcp-lifecycle.md