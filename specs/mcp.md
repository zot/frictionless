# Model Context Protocol (MCP) Server Specification

## 1. Overview
The UI platform provides a Model Context Protocol (MCP) server integration to allow AI assistants (like Claude) to control the application lifecycle, inspect state, and manipulate the runtime environment.

## 1.1 Build & Release

The Makefile provides a `release` target that builds binaries for all supported platforms:

| Platform       | Architecture | Output File                          |
|----------------|--------------|--------------------------------------|
| Linux          | amd64        | `release/ui-mcp-linux-amd64`         |
| Linux          | arm64        | `release/ui-mcp-linux-arm64`         |
| macOS          | amd64        | `release/ui-mcp-darwin-amd64`        |
| macOS          | arm64        | `release/ui-mcp-darwin-arm64`        |
| Windows        | amd64        | `release/ui-mcp-windows-amd64.exe`   |

All binaries are built with `CGO_ENABLED=0` for static linking and include bundled assets.

## 1.2 Versioning

**Source of Truth:** `README.md` contains the canonical version in the format `**Version: X.Y.Z**`.

**CLI Version (`--version` flag or `version` subcommand):**
- Reports the version injected at build time via ldflags (`-X main.Version=$(VERSION)`)
- Falls back to "dev" if not set
- Format: `ui-mcp vX.Y.Z` (e.g., `ui-mcp v0.4.0`)

**MCP Version (`ui_status` tool):**
- Returns version from bundled `README.md` (see Section 5.6)

**Build-time Injection:**
```makefile
VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"
```

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
- **Activation:** `ui-mcp mcp --dir <base_dir>` (default: `{project}/.claude/ui`)

**Output Hygiene:**
- **STDOUT:** Reserved EXCLUSIVELY for MCP JSON-RPC messages.
- **STDERR:** Used for all application logs, debug information, and runtime warnings.

**HTTP Server (random ports):**
Both modes start HTTP servers. In stdio mode, ports are selected randomly and written to `{base_dir}/ui-port` and `{base_dir}/mcp-port`. Endpoints on the MCP port:
- `GET /variables`: Interactive variable tree view
- `GET /state`: Current session state (JSON)
- `GET /wait`: Long-poll for mcp.state changes (see Section 8.3)

### 2.3 SSE Mode (`serve` command)

- **MCP Protocol:** JSON-RPC 2.0 over Server-Sent Events (HTTP).
- **Activation:** `ui-mcp serve --port <ui_port> --mcp-port <mcp_port> --dir <base_dir>` (default: `{project}/.claude/ui`)
- **Two-Port Design:**
  - UI Server port (default 8000): Serves HTML/JS and WebSocket connections
  - MCP Server port (default 8001): SSE transport plus debug endpoints

**MCP Server Endpoints:**
- `GET /sse`: SSE stream for MCP messages
- `POST /message`: Send MCP requests
- `GET /variables`: Interactive variable tree view
- `GET /state`: Current session state (JSON)
- `GET /wait`: Long-poll for mcp.state changes (see Section 8.3)

### 2.4 Install Command (`install`)

Manually install bundled skills and resources without starting the MCP server.

- **Activation:** `ui-mcp install [--dir <base_dir>] [--force]`
- **Default base_dir:** `{project}/.claude/ui`
- **Behavior:** Same as `ui_install` MCP tool (see Section 5.7):
  - Installs Claude skills to `{project}/.claude/skills/`
  - Installs resources, viewdefs, and scripts to `{base_dir}/`
  - Uses version checking (skips if installed >= bundled unless `--force`)

## 3. Server Lifecycle

The MCP server operates as a strict Finite State Machine (FSM).

### 3.1 Startup Behavior

On startup, the server uses `--dir` (defaults to `.claude/ui`) and automatically configures:

1. **Auto-Install:** If `{base_dir}` does not exist OR `{base_dir}/README.md` does not exist, run `ui_install` automatically. This installs:
   - **Claude skills** (`/ui` and `/ui-builder`) to `{project}/.claude/skills/`
   - **Claude agents** to `{project}/.claude/agents/`
   - **MCP resources** (reference docs) to `{base_dir}/resources/`
   - **Standard viewdefs** to `{base_dir}/viewdefs/`
   - **Helper scripts** to `{base_dir}/`
2. **Auto-Configure:** Server starts in CONFIGURED state with the base_dir ready
3. **Reconfiguration:** `ui_configure` can be called to change base_dir if needed

### 3.2 States

| State          | HTTP Server Status | Configuration | Lua I/O    | Description                                                                 |
|:---------------|:-------------------|:--------------|:-----------|:----------------------------------------------------------------------------|
| **CONFIGURED** | **Stopped**        | Loaded        | Redirected | Initial state after startup. Environment is prepped, ready for `ui_start`. |
| **RUNNING**    | **Active**         | Loaded        | Redirected | Server is listening on a port. All tools are fully operational.            |

### 3.3 Transitions

**1. CONFIGURED -> RUNNING**
*   **Trigger:** Successful execution of `ui_start`.
*   **Conditions:** None (other than being in CONFIGURED state).
*   **Effects:**
    *   HTTP listener starts on ephemeral port.
    *   Background workers (SessionManager, etc.) are started.

**2. CONFIGURED -> CONFIGURED (reconfigure)**
*   **Trigger:** Successful execution of `ui_configure`.
*   **Conditions:** `base_dir` is valid and writable.
*   **Effects:**
    *   Filesystem (logs, config) is re-initialized for new base_dir.

**3. RUNNING -> CONFIGURED (reconfigure)**
*   **Trigger:** Successful execution of `ui_configure`.
*   **Effects:**
    *   Current session is destroyed, HTTP server stops.
    *   Re-initializes for new base_dir.

### 3.4 State Invariants & Restrictions

*   **CONFIGURED:**
    *   Calling runtime tools (`ui_run`, etc.) MUST fail with error: "Server not started".
*   **RUNNING:**
    *   Calling `ui_start` again MUST fail with error: "Server already running".

## 4. Lua Environment Integration

When in `--mcp` mode, the Lua runtime environment is modified to ensure compatibility with the stdio transport and enable hot-loading.

### 4.0 Hot-Loading

Hot-loading is **enabled by default** in MCP mode. This capability is provided by ui-engine. The MCP server sets `cfg.Lua.Hotload = true` on startup.

**Supported file types:**
- **Lua files** (`.lua`) — Code is re-executed in the session's Lua context
- **Viewdef files** (`.html`) — Templates are reloaded and pushed to connected browsers

**How it works:**
1. ui-engine watches files in `{base_dir}/apps/*/` for changes
2. On Lua file change:
   - The file is re-executed in the session's Lua context
   - Prototypes declared with `session:prototype()` preserve their identity
   - Existing instances get new methods immediately
   - `mutate()` methods are called automatically for schema migrations
   - **Browser automatically updates** — state changes are pushed to connected browsers
3. On viewdef file change:
   - The template is reloaded from disk
   - The new viewdef is pushed to all connected browsers
   - Components using that viewdef re-render immediately

**Automatic UI updates:** After any hot-loaded change (Lua or viewdef), the browser automatically reflects the changes. No manual refresh or `ui_run` call is needed.

**Requirements for hot-loadable Lua code (idempotent pattern):**
- Use `session:prototype(name, init)` instead of manual metatable setup
- Use `session:create(prototype, instance)` for instance tracking (called by default `:new()`)
- Guard instance creation with `if not session.reloading then ... end`
- Instance name should be lowercase camelCase matching app directory

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

-- Guard instance creation (idempotent)
if not session.reloading then
    contact = Contact:new()
end
```

The agent then uses `ui_display("contact")` to show it in the browser.

**What hot-loading enables:**
- Edit Lua methods → changes take effect immediately, UI updates
- Edit viewdef HTML → browser re-renders with new template
- Add fields → inherited by existing instances via metatable
- Remove fields → automatically nil'd out on instances
- Add `mutate()` → called on all instances for migrations

**Development workflow:** Edit files in your IDE, save, and see changes instantly in the browser. No need to restart the server or manually refresh.

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

### 4.3 The `mcp` Global Object

A global `mcp` table is created in each Lua session to provide MCP-specific functionality.

#### Extension via `mcp.lua`

Projects can extend the `mcp` object with custom shell functionality (e.g., app menus, global UI chrome) by providing `{base_dir}/lua/mcp.lua`.

**Loading sequence:**
1. ui-engine loads `main.lua` (mcp global does NOT exist yet)
2. Go creates the `mcp` global with core methods (`display`, `status`, `pushState`, etc.)
3. Go loads `{base_dir}/lua/mcp.lua` if it exists, extending the mcp global

**Note:** Since `main.lua` runs before the `mcp` global is created, `mcp.lua` must be loaded by Go code after creating the mcp table, not by main.lua.

**Example `mcp.lua`:**
```lua
-- Add app menu functionality to mcp global
function mcp:toggleMenu()
    self.menuOpen = not self.menuOpen
end

function mcp:availableApps()
    -- Return list of apps for menu
end
```

#### Properties

| Property | Type | Description |
|----------|------|-------------|
| `type` | string | Always `"MCP"`. Used for viewdef resolution. |
| `value` | any | The current app value displayed in the browser. Set via `mcp:display()` or direct assignment. Initially `nil`. |

#### Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `pushState` | `mcp.pushState(event)` | Push an event table to the state queue. Signals waiting HTTP clients. See Section 8.1. |
| `pollingEvents` | `mcp:pollingEvents()` | Returns `true` if an agent is connected to `/wait`. See Section 8.2. |
| `display` | `mcp:display(appName)` | Load and display an app. Returns `true` on success, or `nil, error` on failure. |
| `status` | `mcp:status()` | Returns the current MCP server status as a table. See below. |

#### `mcp:status()`

**Purpose:** Returns the current MCP server status, equivalent to the `ui_status` tool response.

**Returns:** A table with the following fields:

| Field      | Lua Type | Presence     | Description                                      |
|------------|----------|--------------|--------------------------------------------------|
| `state`    | `string` | Always       | `"configured"` or `"running"`                    |
| `version`  | `string` | Always       | Semver string (e.g., `"0.6.0"`)                  |
| `base_dir` | `string` | Always       | Absolute or relative path (e.g., `".claude/ui"`) |
| `url`      | `string` | Running only | Server URL (e.g., `"http://127.0.0.1:39482"`)    |
| `sessions` | `number` | Running only | Integer count of connected browsers              |

Fields marked "Running only" are `nil` when `state == "configured"`.

**Example:**
```lua
local status = mcp:status()
if status.state == "running" then
    print("Server running at " .. status.url)
    print("Connected browsers: " .. status.sessions)
end
```

## 5. Tools

### 5.1 `ui_configure`
**Purpose:** Reconfigure the server to use a different base directory. Optional—server auto-configures at startup using `--dir` (defaults to `.claude/ui`).

**Parameters:**
- `base_dir` (string, required): Absolute path to the UI working directory. **Use `{project}/.claude/ui` unless the user explicitly requests a different location.**

**Behavior:**
1.  **Directory Creation:**
    - Creates `base_dir` if it does not exist.
    - Creates a `log` subdirectory within `base_dir`.
2.  **Auto-Install:** If `{base_dir}/README.md` does not exist, runs `ui_install` automatically.
3.  **Configuration Loading:**
    - Checks for existing configuration files in `base_dir`.
    - If found, loads them.
    - If not found, initializes default configuration.
4.  **Runtime Setup:**
    - Configures Lua I/O redirection as described in Section 4.
5.  **State:** Remains in or transitions to CONFIGURED state.

**Returns:**
- Success message indicating the configured directory and log paths.

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
- Reports the current server state (CONFIGURED or RUNNING).
- If RUNNING, reports the server URL and number of connected browser sessions.

**Returns:**
- JSON object with status information:
  - `state`: Current lifecycle state ("configured" or "running")
  - `version`: Bundled version from README.md (always present)
  - `base_dir`: Configured base directory (always present)
  - `url`: Server URL (only if running)
  - `sessions`: Number of active browser sessions (only if running)

**Example Response:**
```json
{
  "state": "running",
  "version": "0.1.0",
  "base_dir": ".claude/ui",
  "url": "http://127.0.0.1:39482",
  "sessions": 1
}
```

### 5.7 `ui_install`
**Purpose:** Installs bundled configuration files to enable full ui-mcp integration.

**Parameters:**
- `force` (boolean, optional): If true, overwrites existing files regardless of version. Defaults to `false`.

**Version Checking:**

The README.md contains a semantic version near the top:
```markdown
# ui-mcp

**Version: 0.1.0**
```

Installation behavior:
1. Read the `version` from bundled README.md
2. If installed README.md exists, read its `version`
3. **Install all bundled files if:**
   - No installed version exists, OR
   - Bundled version > installed version (semver comparison), OR
   - `force=true`
4. Skip installation if installed version >= bundled version (unless `force=true`)
5. Return `version_skipped: true` and both versions when skipping due to version

**Install Manifest:**

Skills and agents installed to `{project}/.claude/`:
```
.claude/skills/ui/SKILL.md
.claude/skills/ui-builder/SKILL.md
.claude/skills/ui-builder/examples/requirements.md
.claude/skills/ui-builder/examples/design.md
.claude/skills/ui-builder/examples/app.lua
.claude/skills/ui-builder/examples/viewdefs/ContactApp.DEFAULT.html
.claude/skills/ui-builder/examples/viewdefs/Contact.list-item.html
.claude/skills/ui-builder/examples/viewdefs/ChatMessage.list-item.html
.claude/agents/ui-builder.md
```

Resources installed to `{base_dir}/resources/`:
```
resources/reference.md
resources/viewdefs.md
resources/lua.md
resources/mcp.md
```

Viewdefs installed to `{base_dir}/viewdefs/`:
```
viewdefs/lua.ViewList.DEFAULT.html
viewdefs/lua.ViewListItem.list-item.html
viewdefs/MCP.DEFAULT.html
```

Scripts installed to `{base_dir}/` (executable):
```
event
state
variables
linkapp
```

Lua entry point installed to `{base_dir}/lua/`:
```
lua/main.lua
```

HTML files installed to `{base_dir}/html/` (dynamically discovered from bundle):
```
html/index.html
html/main-*.js
html/worker-*.js
```

Documentation installed to `{base_dir}`:
```
README.md
```

**Path Resolution:**
- `{project}` is the parent of `base_dir` (e.g., if `base_dir` is `.claude/ui`, project is `.`)
- Creates `.claude/`, `.claude/skills/`, and `.claude/agents/` directories if they don't exist

**Behavior:**
1. **Check State:** Must be in CONFIGURED or RUNNING state.
2. **Skill/Resource Files:**
   - If file doesn't exist: install from bundle
   - If exists and `force=false`: skip (no-op)
   - If exists and `force=true`: overwrite

**Returns:**
- JSON object listing installed files:
```json
{
  "installed": [".claude/skills/ui-builder/SKILL.md", ".claude/ui/resources/reference.md"],
  "skipped": [],
  "appended": []
}
```

**Design Rationale:**
- Separates installation from configuration (user controls when files are added)
- Skill files are self-describing (no CLAUDE.md augmentation needed)
- Skill files are only overwritten with explicit `force=true`
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

### 8.2 `mcp:pollingEvents()`

**Purpose:** Check whether an agent is actively polling for events via the `/wait` endpoint.

**Lua API:**
```lua
-- Check if a client is connected to the /wait endpoint
if mcp:pollingEvents() then
    print("Agent is listening for events")
end
```

**Returns:**
- `true` if there is at least one client currently connected to the `/wait` endpoint.
- `false` otherwise.

**Use Case:**
This allows Lua code to conditionally enable event-driven features or provide visual feedback (e.g., a status indicator) based on whether an agent is actively monitoring for events.

### 8.3 HTTP Wait Endpoint

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

### 8.4 Agent Integration Pattern

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
