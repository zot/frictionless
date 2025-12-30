# Model Context Protocol (MCP) Server Specification

## 1. Overview
The UI platform provides a Model Context Protocol (MCP) server integration to allow AI assistants (like Claude) to control the application lifecycle, inspect state, and manipulate the runtime environment.

## 2. Transport & Hygiene

### 2.1 Transport
- **Protocol:** JSON-RPC 2.0 over Standard Input (stdin) and Standard Output (stdout).
- **Encoding:** UTF-8.

### 2.2 Output Hygiene
- **STDOUT (Standard Output):** Reserved EXCLUSIVELY for MCP JSON-RPC messages.
- **STDERR (Standard Error):** Used for all application logs, debug information, and runtime warnings.
- **Conditional Activation:** These restrictions and the stdio transport are only active when the application is started with the `--mcp` flag. Without this flag, the application behaves normally (logging to configured outputs, potentially using stdout).

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

When in `--mcp` mode, the Lua runtime environment is modified to ensure compatibility with the stdio transport:

- **`print(...)` Override:** The global `print` function is replaced with a version that:
    - Opens the log file (`{base_dir}/log/lua.log`) in **append** mode.
    - Seeks to the end of the file (to handle concurrent external edits/truncation).
    - Writes the formatted output.
    - Flushes the stream.
    - Closes the file handle (or effectively manages it) to allow external tools (e.g., `tail -f`) to read the log safely.
- **Standard Streams:**
    - `io.stdout` is redirected to `{base_dir}/log/lua.log`.
    - `io.stderr` is redirected to `{base_dir}/log/lua-err.log`.

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

### 5.2 `ui_start`
**Purpose:** Starts the embedded HTTP UI server.

**Pre-requisites:**
- Server must be in the `Configured` state.
- Server must not already be `Running`.

**Behavior:**
1.  **Port Selection:** Selects a random available ephemeral port (binding to port 0).
2.  **Server Start:** Launches the HTTP server on `127.0.0.1`.
3.  **State Transition:** Moves server state from `Configured` to `Running`.

**Returns:**
- The full base URL of the running server (e.g., `http://127.0.0.1:39482`).

### 5.3 `ui_run`
**Purpose:** Execute arbitrary Lua code within a session's context.

**Parameters:**
- `code` (string, required): The Lua code chunk to execute.
- `sessionId` (string, optional): The target session ID. Defaults to "1".

**Behavior:**
- Wraps execution in a `session` context, allowing direct access to session variables via the `session` global object.
- Attempts to marshal the execution result to JSON.

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
  - `url`: Server URL (only if running)
  - `sessions`: Number of active browser sessions (only if running)

**Example Response:**
```json
{
  "state": "running",
  "url": "http://127.0.0.1:39482",
  "sessions": 1
}
```

## 7. Notifications

The MCP server supports server-to-client notifications, enabling Lua code to communicate events back to the AI agent.

### 7.1 `mcp.notify(method, params)`
**Purpose:** Sends a JSON-RPC 2.0 notification from Lua code to the connected MCP client (AI agent).

**Lua API:**
```lua
mcp.notify(method, params)
```

**Parameters:**
- `method` (string): The notification method name (e.g., "feedback", "user_action").
- `params` (table, optional): Structured data for the notification.

**Behavior:**
- Converts Lua table to JSON object.
- Sends JSON-RPC 2.0 notification to the MCP client via stdout.

**Example:**
```lua
-- User submits feedback form
function Feedback:submit()
    mcp.notify("feedback", {
        rating = self.rating,
        comment = self.comment
    })
end
```

**Wire Format:**
```json
{"jsonrpc":"2.0","method":"feedback","params":{"rating":4,"comment":"Great!"}}
```

### 7.2 Use Cases
- **Form submission:** User fills form → Lua handler calls `mcp.notify` → Agent receives data
- **User actions:** Button clicks, selections, navigation events
- **State changes:** When application state changes that the agent should know about

## 8. Resources

MCP Resources provide read access to state and documentation.

### 8.1 State Resources

| URI | Description |
|-----|-------------|
| `ui://state` | Current JSON state of session 1 (Variable 1) |
| `ui://state/{sessionId}` | Current JSON state of the specified session |

**Example Response (ui://state):**
```json
{
  "type": "MyApp",
  "title": "My Application",
  "items": [...]
}
```

### 8.2 Debug Resources

| URI | Description |
|-----|-------------|
| `ui://variables` | Topologically sorted array of all tracked variables for session 1 |

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

The `/debug/variables` HTTP endpoint renders the same data as an interactive HTML tree using Shoelace components.

### 8.2 Documentation Resources

| URI | Description |
|-----|-------------|
| `ui://reference` | Main entry point for UI platform documentation |
| `ui://viewdefs` | Guide to ui-* attributes and path syntax |
| `ui://lua` | Lua API, class patterns, and global objects |
| `ui://mcp` | Guide for AI agents to build apps |

### 8.3 Static Resources

| URI | Description |
|-----|-------------|
| `ui://{path}` | Generic resource for static content in the resources directory |

Files in `{base_dir}/resources/` are accessible via `ui://{filename}` (e.g., `ui://patterns/form.md`).
