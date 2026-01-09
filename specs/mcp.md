# Model Context Protocol (MCP) Server Specification

## 1. Overview
The UI platform provides a Model Context Protocol (MCP) server integration to allow AI assistants (like Claude) to control the application lifecycle, inspect state, and manipulate the runtime environment.

## 2. Transport & Modes

### 2.1 Transport Options

The MCP server supports two transport modes:

| Mode      | Command        | MCP Transport                  | Use Case                           |
|-----------|----------------|--------------------------------|------------------------------------|
| **Stdio** | `ui-mcp mcp`   | JSON-RPC 2.0 over stdin/stdout | AI agent integration (Claude Code) |
| **SSE**   | `ui-mcp serve` | Server-Sent Events over HTTP   | Standalone development/debugging   |

Both modes start an HTTP server with debug and API endpoints.

### 2.2 Stdio Mode (`mcp` command)

- **MCP Protocol:** JSON-RPC 2.0 over Standard Input (stdin) and Standard Output (stdout).
- **Encoding:** UTF-8.
- **Activation:** `ui-mcp mcp --dir <base_dir>`

**Output Hygiene:**
- **STDOUT:** Reserved EXCLUSIVELY for MCP JSON-RPC messages.
- **STDERR:** Used for all application logs, debug information, and runtime warnings.

**HTTP Server (random ports):**
Both modes start HTTP servers. In stdio mode, ports are selected randomly and written to `{base_dir}/ui-port` and `{base_dir}/mcp-port`. Endpoints on the MCP port:
- `GET /variables`: Interactive variable tree view
- `GET /state`: Current session state (JSON)
- `GET /wait`: Long-poll for mcp.state changes (see Section 8.2)

### 2.3 SSE Mode (`serve` command)

- **MCP Protocol:** JSON-RPC 2.0 over Server-Sent Events (HTTP).
- **Activation:** `ui-mcp serve --port <ui_port> --mcp-port <mcp_port> --dir <base_dir>`
- **Two-Port Design:**
  - UI Server port (default 8000): Serves HTML/JS and WebSocket connections
  - MCP Server port (default 8001): SSE transport plus debug endpoints

**MCP Server Endpoints:**
- `GET /sse`: SSE stream for MCP messages
- `POST /message`: Send MCP requests
- `GET /variables`: Interactive variable tree view
- `GET /state`: Current session state (JSON)
- `GET /wait`: Long-poll for mcp.state changes (see Section 8.2)

## 3. Server Lifecycle

The MCP server operates as a strict Finite State Machine (FSM).

### 3.1 States

| State            | HTTP Server Status | Configuration | Lua I/O    | Description                                                                                  |
|:-----------------|:-------------------|:--------------|:-----------|:---------------------------------------------------------------------------------------------|
| **UNCONFIGURED** | **Stopped**        | None          | Standard   | Initial state on process start. Only `ui_configure` is permitted.                            |
| **CONFIGURED**   | **Stopped**        | Loaded        | Redirected | Environment is prepped, logs are active, but no network port is bound. Ready for `ui_start`. |
| **RUNNING**      | **Active**         | Loaded        | Redirected | Server is listening on a port. All tools are fully operational.                              |

### 3.2 Transitions

**1. UNCONFIGURED -> CONFIGURED**
*   **Trigger:** Successful execution of `ui_configure`.
*   **Conditions:** `base_dir` is valid and writable.
*   **Effects:**
    *   Filesystem (logs, config) is initialized.
    *   Lua `print`, `stdout`, and `stderr` are redirected to log files.
    *   Internal config struct is populated.

**2. CONFIGURED -> RUNNING**
*   **Trigger:** Successful execution of `ui_start`.
*   **Conditions:** None (other than being in CONFIGURED state).
*   **Effects:**
    *   HTTP listener starts on ephemeral port.
    *   Background workers (SessionManager, etc.) are started.

### 3.3 State Invariants & Restrictions

*   **UNCONFIGURED:**
    *   Calling `ui_start`, `ui_run`, `ui_get_state`, etc. MUST fail with error: "Server not configured".
*   **CONFIGURED:**
    *   Calling `ui_configure` again IS permitted (re-configuration).
    *   Calling runtime tools (`ui_run`, etc.) MUST fail with error: "Server not started".
*   **RUNNING:**
    *   Calling `ui_start` again MUST fail with error: "Server already running".
    *   Calling `ui_configure` IS permitted: it destroys the current session, resets state to CONFIGURED, then proceeds with configuration. This allows session restart without process restart.

## 4. Lua Environment Integration

When in `--mcp` mode, the Lua runtime environment is modified to ensure compatibility with the stdio transport and enable hot-loading.

### 4.0 Hot-Loading

Hot-loading is **enabled by default** in MCP mode. This capability is provided by ui-engine. The MCP server sets `cfg.Lua.Hotload = true` on startup.

**How it works:** (see ui-engine/USAGE.md and ui-engine/HOT-LOADING.md for full details)
1. ui-engine watches Lua files in `{base_dir}/apps/*/` for changes
2. On file change, the file is re-executed in the session's Lua context
3. Prototypes declared with `session:prototype()` preserve their identity
4. Existing instances get new methods immediately
5. `mutate()` methods are called automatically for schema migrations

**Requirements for hot-loadable code:**
- Use `session:prototype(name, init)` instead of manual metatable setup
- Use `session:create(prototype, instance)` for instance tracking
- Guard app creation: `if not session:getApp() then ... end`

**Example:**
```lua
-- Declare prototype (preserves identity on reload)
Contact = session:prototype("Contact", {
    name = "",
    email = "",
})

-- Override :new() only when needed (default provided)
function Contact:new(data)
    local instance = session:create(Contact, data)
    return instance
end

-- Guard app creation
if not session:getApp() then
    session:createAppVariable(App:new())
end
```

**What hot-loading enables:**
- Edit methods → changes take effect immediately
- Add fields → inherited by existing instances via metatable
- Remove fields → automatically nil'd out on instances
- Add `mutate()` → called on all instances for migrations

### 4.1 I/O Redirection

- **`print(...)` Override:** The global `print` function is replaced with a version that:
    - Opens the log file (`{base_dir}/log/lua.log`) in **append** mode.
    - Seeks to the end of the file (to handle concurrent external edits/truncation).
    - Writes the formatted output.
    - Flushes the stream.
    - Closes the file handle (or effectively manages it) to allow external tools (e.g., `tail -f`) to read the log safely.
- **Standard Streams:**
    - `io.stdout` is redirected to `{base_dir}/log/lua.log`.
    - `io.stderr` is redirected to `{base_dir}/log/lua-err.log`.

### 4.2 Browser Update Mechanism

The MCP server delegates to the ui-server's `Server.ExecuteInSession` method for executing code within a session context. This method:

1. Queues the function through the session's executor (serializing with WebSocket messages)
2. Executes the function
3. Calls `afterBatch` to detect and push state changes to connected browsers
4. Returns the result

**Implication:** Any operation that needs to trigger a browser update can call `ExecuteInSession` with an empty function.

**Panic Recovery Requirement:** The MCP server MUST wrap `ExecuteInSession` with panic recovery (e.g., `SafeExecuteInSession`) to prevent Lua errors or panics from crashing the MCP process. Panics should be caught and returned as errors.

## 5. Tools

### 5.1 `ui_configure`
**Purpose:** Prepares the server environment and file system. This must be the first tool called.

**Parameters:**
- `base_dir` (string, required): Absolute path to the directory serving as the project root.

**Behavior:**
1.  **Directory Creation:**
    - Creates `base_dir` if it does not exist.
    - Creates a `log` subdirectory within `base_dir`.
2.  **Configuration Loading:**
    - Checks for existing configuration files in `base_dir`.
    - If found, loads them.
    - If not found, initializes default configuration suitable for the MCP environment.
3.  **Runtime Setup:**
    - Configures Lua I/O redirection as described in Section 4.
4.  **State Transition:** Moves server state from `Unconfigured` to `Configured`.

**Returns:**
- Success message indicating the configured directory and log paths.

### 5.1.1 Installation Check

During configuration, the MCP server checks if bundled files have been installed and prompts the agent to run `ui_install` if needed.

**Behavior:**
1. **Check:** Look for `.claude/agents/ui-builder.md` relative to the parent of `base_dir`.
2. **If missing:** Include in `ui_configure` response: `"install_needed": true, "hint": "Run ui_install to install agent files and CLAUDE.md instructions"`
3. **If present:** No action needed.

**Design Rationale:**
- Separates configuration from installation (cleaner lifecycle)
- Agent explicitly calls `ui_install` when needed
- See section 5.7 for full installation behavior

### 5.2 `ui_start`
**Purpose:** Starts the embedded HTTP UI server.

**Pre-requisites:**
- Server must be in the `Configured` state.
- Server must not already be `Running`.

**Behavior:**
1.  **Port Selection:** Selects random available ephemeral ports for UI and MCP servers.
2.  **Server Start:** Launches the HTTP servers on `127.0.0.1`.
3.  **Port File Creation:** Writes port numbers to files in `base_dir`:
    - `{base_dir}/ui-port` - The UI server port (serves HTML/JS/WebSocket)
    - `{base_dir}/mcp-port` - The MCP server port (serves /state, /wait, /variables endpoints)
4.  **State Transition:** Moves server state from `Configured` to `Running`.

**Returns:**
- The full base URL of the running UI server (e.g., `http://127.0.0.1:39482`).

### 5.3 `ui_run`
**Purpose:** Execute arbitrary Lua code within a session's context.

**Parameters:**
- `code` (string, required): The Lua code chunk to execute.
- `sessionId` (string, optional): The target session ID. Defaults to "1".

**Behavior:**
- Wraps execution in a `session` context, allowing direct access to session variables via the `session` global object.
- Attempts to marshal the execution result to JSON.
- **Browser Update:** After Lua execution, any state changes are automatically pushed to connected browsers.

**Example Usage:**
To get the first name of the first contact in an application:
```lua
return session:getApp().contacts[1].firstName
```

**Returns:**
- If successful: The JSON representation of the result.
- If not marshalable: A JSON object `{"non-json": "STRING_REPRESENTATION"}`.
- If execution fails: An error message.

### 5.4 `ui_upload_viewdef`
**Purpose:** Dynamically add or update a view definition.

**Parameters:**
- `type` (string, required): The presenter type (e.g., "MyPresenter").
- `namespace` (string, required): The view namespace (e.g., "DEFAULT").
- `content` (string, required): The HTML content of the view definition.

**Behavior:**
- Registers the view definition in the runtime's viewdef store.
- **Push Update:** If any frontends are currently connected to the server, the new view definition MUST be pushed to them immediately to ensure they can re-render affected components without a page reload.
- **Variable Refresh:** The server MUST identify all active variables in the session that match the `type` of the uploaded viewdef and send an empty update for them. This forces the frontend to re-request the variable state and re-render using the new view definition.

**Returns:**
- Confirmation message.

### 5.5 `ui_open_browser`
**Purpose:** Opens the system's default web browser to the UI session.

**Parameters:**
- `sessionId` (string, optional): The session to open. Defaults to "1".
- `path` (string, optional): The URL path to open. Defaults to "/".
- `conserve` (boolean, optional): If true, attempts to focus an existing tab or notifies the user instead of opening a duplicate session. Defaults to `true`.

**Behavior:**
- Constructs the full URL using the running server's port and session ID.
- **URL Pattern:** `http://127.0.0.1:PORT/SESSION-ID/PATH?conserve=true`
- **Default:** Always appends `?conserve=true` unless explicitly disabled, ensuring the SharedWorker coordination logic is engaged to prevent duplicate tabs.
- Invokes the system's default browser opener (e.g., `xdg-open`, `open`, or `start`).

**Returns:**
- Success message or error if the browser could not be launched.

## 6. Frontend Integration

### 6.1 Conserve Mode (`?conserve=true`)
To prevent cluttering the user's workspace with multiple tabs for the same session, the frontend must implement a "Conserve Mode" relying on a **SharedWorker**.

**Mechanism:**
1.  **SharedWorker Coordination:**
    - The frontend connects to a SharedWorker unique to the session/server origin.
    - **Initialization:** If the SharedWorker is not running, the presence of `?conserve=true` MUST trigger its creation and initialization.
    - The SharedWorker maintains a count or list of active clients (tabs/windows).
2.  **Detection:**
    - On load, if `?conserve=true` is present, the client queries the SharedWorker.
    - If the SharedWorker reports **other active clients** for this session:
        - **Action 1:** The new tab does **NOT** initialize the full UI application or WebSocket connection.
        - **Action 2:** Displays a minimal page with a "Session is already open" message and a "Close this page" link.
        - **Action 3:** Triggers a **Desktop Notification** (via the Web Notifications API) to alert the user: "Session [ID] is already active in another tab."
    - If no other clients are active, the tab proceeds to load normally.

### 5.6 `ui_status`
**Purpose:** Returns the current status of the MCP server including lifecycle state and browser connection status.

**Parameters:** None.

**Behavior:**
- Reports the current server state (UNCONFIGURED, CONFIGURED, or RUNNING).
- If RUNNING, reports the server URL and number of connected browser sessions.

**Returns:**
- JSON object with status information:
  - `state`: Current lifecycle state ("unconfigured", "configured", or "running")
  - `base_dir`: Configured base directory (only if configured or running)
  - `url`: Server URL (only if running)
  - `sessions`: Number of active browser sessions (only if running)

**Example Response:**
```json
{
  "state": "running",
  "base_dir": ".claude/ui",
  "url": "http://127.0.0.1:39482",
  "sessions": 1
}
```

### 5.7 `ui_install`
**Purpose:** Installs bundled configuration files to enable full ui-mcp integration.

**Parameters:**
- `force` (boolean, optional): If true, overwrites existing files. Defaults to `false`.

**Bundled Files:**

| Source (bundled)              | Destination                               | Purpose                                 | Exclude |
|-------------------------------|-------------------------------------------|-----------------------------------------|---------|
| `init/add-to-claude.md`       | `{project}/CLAUDE.md` (appended)          | Instructions for using ui-builder agent |         |
| `init/agents/ui-builder.md`   | `{project}/.claude/agents/ui-builder.md`  | UI building agent                       | yes     |
| `init/agents/ui-learning.md`  | `{project}/.claude/agents/ui-learning.md` | Pattern extraction agent                | yes     |
| `init/skills/*`               | `{project}/.claude/skills/*`              | UI builder skills                       |         |
| `resources/*`                 | `{base_dir}/resources/*`                  | MCP server resources                    |         |
| `viewdefs/*`                  | `{base_dir}/viewdefs/*`                   | standard viewdefs, like ViewList's      |         |
| `event`, `state`, `variables` | `{base_dir}`                              | scripts for easy MCP endpoint access    |         |

**Note:** The agents are currently disabled due to a bug that prevents subagents from accessing files. Users should invoke the skills directly until this is resolved.

**Path Resolution:**
- `{project}` is the parent of `base_dir` (e.g., if `base_dir` is `.claude/ui`, project is `.`)
- Creates `.claude/` and `.claude/agents/` directories if they don't exist

**Behavior:**
1. **Check State:** Must be in CONFIGURED or RUNNING state.
2. **CLAUDE.md Handling:**
   - If `CLAUDE.md` doesn't exist in project root: create with bundled content
   - If exists and ui-builder instructions are not in CLAUDE.md: append bundled content with separator
   - If exists and ui-builder instructions are in CLAUDE.md and `force=false`: ignore
   - If exists and ui-builder instructions are in CLAUDE.md and `force=true`: replace ui-builder instructions with bundled content
3. **Agent Files:**
   - If file doesn't exist: install from bundle
   - If exists and `force=false`: skip (no-op)
   - If exists and `force=true`: overwrite

**Returns:**
- JSON object listing installed files:
```json
{
  "installed": [".claude/agents/ui-builder.md", ".claude/agents/ui-learning.md"],
  "skipped": [],
  "appended": ["CLAUDE.md"]
}
```

**Design Rationale:**
- Separates installation from configuration (user controls when files are added)
- CLAUDE.md append behavior preserves existing project instructions
- Agent files are only overwritten with explicit `force=true`
- Enables easy updates: `ui_install(force=true)` reinstalls latest bundled versions

## 7. Resources

MCP Resources provide read access to state and documentation.

### 7.1 State Resources

| URI          | Description                           |
|--------------|---------------------------------------|
| `ui://state` | Current JSON state of the MCP session |

**Example Response (ui://state):**
```json
{
  "type": "MyApp",
  "title": "My Application",
  "items": [...]
}
```

### 7.2 Variable Resources

| URI              | Description                                                             |
|------------------|-------------------------------------------------------------------------|
| `ui://variables` | Topologically sorted array of all tracked variables for the MCP session |

Each variable includes: id, parentId, type, path, value, properties, and childIds.

**Example Response (ui://variables):**
```json
[
  {
    "id": 1,
    "parentId": 0,
    "type": "MCP",
    "value": {"obj": 2},
    "properties": {"type": "MCP"},
    "childIds": []
  }
]
```

The `/variables` HTTP endpoint renders the same data as an interactive HTML tree using Shoelace components.

### 7.3 Documentation Resources

| URI              | Description                                    |
|------------------|------------------------------------------------|
| `ui://reference` | Main entry point for UI platform documentation |
| `ui://viewdefs`  | Guide to ui-* attributes and path syntax       |
| `ui://lua`       | Lua API, class patterns, and global objects    |
| `ui://mcp`       | Guide for AI agents to build apps              |

### 7.4 Static Resources

| URI           | Description                                                    |
|---------------|----------------------------------------------------------------|
| `ui://{path}` | Generic resource for static content in the resources directory |

Files in `{base_dir}/resources/` are accessible via `ui://{filename}` (e.g., `ui://patterns/form.md`).

## 8. State Change Waiting

Since some MCP clients (including Claude Code) do not support receiving server-to-client notifications, an alternative mechanism is provided for UI-to-agent communication via HTTP long-polling.

### 8.1 `mcp.pushState(event)`

**Purpose:** Push events to a queue that the agent can read via HTTP long-poll.

**Lua API:**
```lua
-- Push an event to the queue (signals waiting agent)
mcp.pushState({ app = "contacts", event = "chat", text = "hello" })

-- Push multiple events
mcp.pushState({ app = "contacts", event = "button", id = "save" })
mcp.pushState({ app = "contacts", event = "button", id = "cancel" })
```

**Behavior:**
- Events are queued internally and waiting HTTP clients are signaled immediately.
- When the wait endpoint responds, it atomically returns all queued events and clears the queue.
- This ensures no events are lost between the read and subsequent writes.
- Queue contents readable via `ui://state` MCP resource.

### 8.2 HTTP Wait Endpoint

**Endpoint:** `GET /wait`

**Implementation:** Added to HTTP mux in `internal/mcp/server.go` (both `ServeSSE` and `StartHTTPServer`). Uses the server's current session.

**Query Parameters:**
- `timeout` (integer, optional): Maximum wait time in seconds. Default: 30. Max: 120.

**Behavior:**
1. Blocks until events are pushed via `mcp.pushState()` or timeout expires.
2. Atomically drains the queue and returns accumulated events.
3. Returns the events as a JSON array.
4. Returns HTTP 204 (No Content) on timeout (no events).
5. Returns HTTP 404 if session does not exist.
6. **Triggers UI update** after draining the queue by calling `SafeExecuteInSession` with an empty function (see Section 4.1). This ensures UIs monitoring the event queue refresh.

**Example Request:**
```
GET /wait?timeout=30 HTTP/1.1
Host: localhost:39482
```

**Example Response (events queued):**
```
HTTP/1.1 200 OK
Content-Type: application/json

[{"app":"contacts","event":"chat","text":"hello"},{"app":"contacts","event":"button","id":"save"}]
```

**Example Response (timeout):**
```
HTTP/1.1 204 No Content
```

### 8.3 Agent Integration Pattern

Agents can monitor for state changes using a background shell script:

**Script:** `scripts/wait-for-state.sh`

```bash
#!/bin/bash
# Outputs one JSON object per line when mcp.state events arrive.
# Exits when server shuts down.
BASE_URL="${1:?Usage: wait-for-state.sh <base_url> [timeout]}"
TIMEOUT="${2:-30}"

while true; do
    RESPONSE=$(curl -s -w "\n%{http_code}" "${BASE_URL}/wait?timeout=${TIMEOUT}" 2>/dev/null)
    [ $? -ne 0 ] && exit 0  # Server disconnected

    HTTP_CODE=$(echo "$RESPONSE" | tail -1)
    BODY=$(echo "$RESPONSE" | sed '$d')

    case "$HTTP_CODE" in
        200) echo "$BODY" | jq -c '.[]' ;;  # Output each event on its own line
        204) continue ;;                     # Timeout, retry
        *)   exit 1 ;;                       # Other error
    esac
done
```

**Agent Workflow:**
1. Start script in background: `Bash(run_in_background=true)`
2. Continue with other work
3. Check `TaskOutput` periodically or when expecting user input
4. Parse JSON events from script output
