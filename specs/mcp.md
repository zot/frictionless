# Model Context Protocol (MCP) Server Specification

## 1. Overview
The UI platform provides a Model Context Protocol (MCP) server integration to allow AI assistants (like Claude) to control the application lifecycle, inspect state, and manipulate the runtime environment.

## .ui/mcp notation equivalent to {basedir}/mcp
In rare circumstances, the `ui_status` tool will return a directory other than `.ui` for `{basedir}`. In those cases, use `{basedir}` instead of `.ui` in the below instructions.


## 1.1 Build & Release

The Makefile provides a `release` target that builds binaries for all supported platforms:

| Platform       | Architecture | Output File                          |
|----------------|--------------|--------------------------------------|
| Linux          | amd64        | `release/frictionless-linux-amd64`         |
| Linux          | arm64        | `release/frictionless-linux-arm64`         |
| macOS          | amd64        | `release/frictionless-darwin-amd64`        |
| macOS          | arm64        | `release/frictionless-darwin-arm64`        |
| Windows        | amd64        | `release/frictionless-windows-amd64.exe`   |

All binaries are built with `CGO_ENABLED=0` for static linking and include bundled assets.

## 1.2 Versioning

**Source of Truth:** `README.md` contains the canonical version in the format `**Version: X.Y.Z**`.

**CLI Version (`--version` flag or `version` subcommand):**
- Reports the version injected at build time via ldflags (`-X main.Version=$(VERSION)`)
- Falls back to "dev" if not set
- Format: `frictionless vX.Y.Z` (e.g., `frictionless v0.4.0`)

**MCP Version (`ui_status` tool):**
- Returns version from bundled `README.md` (see Section 5.5)

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
| **Stdio** | `frictionless mcp`   | JSON-RPC 2.0 over stdin/stdout | AI agent integration (Claude Code) |
| **SSE**   | `frictionless serve` | Server-Sent Events over HTTP   | Standalone development/debugging   |

Both modes start an HTTP server with debug and API endpoints.

### 2.2 Stdio Mode (`mcp` command)

- **MCP Protocol:** JSON-RPC 2.0 over Standard Input (stdin) and Standard Output (stdout).
- **Encoding:** UTF-8.
- **Activation:** `frictionless mcp --dir <base_dir>` (default: `{project}/.ui`)

**Output Hygiene:**
- **STDOUT:** Reserved EXCLUSIVELY for MCP JSON-RPC messages.
- **STDERR:** Used for all application logs, debug information, and runtime warnings.

**HTTP Server (random ports):**
Both modes start HTTP servers. In stdio mode, ports are selected randomly and written to `{base_dir}/ui-port` and `{base_dir}/mcp-port`. Endpoints on the MCP port:
- `GET /variables`: Interactive variable tree view
- `GET /state`: Current session state (JSON)
- `GET /wait`: Long-poll for mcp.state changes (see Section 8.4)
- `GET /api/resource/`: List resources directory (JSON for curl, HTML for browsers)
- `GET /api/resource/{path}`: Serve resource file (markdown rendered as HTML for browsers, raw for curl)
- `GET /app/{app}/readme`: Serve app's README.md as HTML (case-insensitive lookup, rendered via goldmark)

### 2.3 SSE Mode (`serve` command)

- **MCP Protocol:** JSON-RPC 2.0 over Server-Sent Events (HTTP).
- **Activation:** `frictionless serve --port <ui_port> --mcp-port <mcp_port> --dir <base_dir>` (default: `{project}/.ui`)
- **Two-Port Design:**
  - UI Server port (default 8000): Serves HTML/JS and WebSocket connections
  - MCP Server port (default 8001): SSE transport plus debug endpoints

**MCP Server Endpoints:**
- `GET /sse`: SSE stream for MCP messages
- `POST /message`: Send MCP requests
- `GET /variables`: Interactive variable tree view
- `GET /state`: Current session state (JSON)
- `GET /wait`: Long-poll for mcp.state changes (see Section 8.4)
- `GET /api/resource/`: List resources directory (JSON for curl, HTML for browsers)
- `GET /api/resource/{path}`: Serve resource file (markdown rendered as HTML for browsers, raw for curl)
- `GET /app/{app}/readme`: Serve app's README.md as HTML (case-insensitive lookup, rendered via goldmark)

### 2.4 Install Command (`install`)

Manually install bundled skills and resources without starting the MCP server.

- **Activation:** `frictionless install [--dir <base_dir>] [--force]`
- **Default base_dir:** `{project}/.ui`
- **Behavior:** Same as `ui_install` MCP tool (see Section 5.5):
  - Installs Claude skills to `{project}/.claude/skills/`
  - Installs resources, viewdefs, and scripts to `{base_dir}/`
  - Uses version checking (skips if installed >= bundled unless `--force`)
  - Checks for optional dependencies and outputs suggestions (e.g., `code-simplifier` agent)

### 2.5 HTTP Tool API

All MCP tools are accessible via HTTP through the `.ui/mcp TOOL_NAME` command. This enables spawned agents and scripts to interact with the UI server using curl instead of requiring MCP protocol access.

**Commands:**

| command                              | Description                           |
|--------------------------------------|---------------------------------------|
| `.ui/mcp audit APP`                  | run code quality audit on APP         |
| `.ui/mcp browser`                    | open browser to UI session            |
| `.ui/mcp display APP`                | display APP in the browser            |
| `.ui/mcp event`                      | wait for next UI event (120s timeout) |
| `.ui/mcp linkapp add|remove APP`     | manage app symlinks                   |
| `.ui/mcp progress APP PERCENT STAGE` | report build progress                 |
| `.ui/mcp run 'lua code'`             | execute Lua code in session           |
| `.ui/mcp state`                      | get current session state             |
| `.ui/mcp status`                     | get server status                     |
| `.ui/mcp variables`                  | get current variable values           |

**Response Format:**
```json
{"result": ...}
```

Or on error:
```json
{"error": "error message"}
```

**Example Usage:**
```bash
# Get status
.ui/mcp status

# Execute Lua code
.ui/mcp run 'return testApp:addResponse(\"Hello!\")'

# Display an app
.ui/mcp display 'contacts'
```

## 3. Server Lifecycle

### 3.1 Startup Behavior

On startup, the server uses `--dir` (defaults to `.ui`) and automatically configures and starts:

1. **Auto-Install:** If `{base_dir}` does not exist OR `{base_dir}/README.md` does not exist, run `ui_install` automatically. This installs:
   - **Claude skills** (`/ui` and `/ui-builder`) to `{project}/.claude/skills/`
   - **Claude agents** to `{project}/.claude/agents/`
   - **MCP resources** (reference docs) to `{base_dir}/resources/`
   - **Standard viewdefs** to `{base_dir}/viewdefs/`
   - **Helper scripts** to `{base_dir}/`
2. **Auto-Start:** Server starts HTTP listeners
3. **Reconfiguration:** `ui_configure` can be called to reconfigure and restart with a different base_dir

### 3.2 Reconfiguration

Calling `ui_configure` while running triggers a full reconfigure:
*   **Trigger:** Successful execution of `ui_configure`.
*   **Conditions:** `base_dir` is valid and writable.
*   **Effects:**
    *   Current session is destroyed, HTTP server stops.
    *   Filesystem (logs, config) is re-initialized for new base_dir.
    *   HTTP listener restarts on new ephemeral port.
    *   Background workers are restarted.

### 3.3 Root URL Session Binding

**Problem:** ui-engine's default behavior creates a new session when a browser navigates to `/`. This is incorrect for MCP mode where a session with the `mcp` global already exists.

**Solution:** frictionless registers a root session provider that returns the current MCP session ID:

*   **Trigger:** Browser navigates to `http://127.0.0.1:PORT/`
*   **Behavior:** Server sets a `ui-session` cookie with the current session ID and serves index.html (no redirect)
*   **Effect:** Browser connects to the existing session containing the `mcp` global and any displayed app

**Session Cookie (`ui-session`):**
- Set by the server when serving index.html (both for `/` and `/{session-id}` paths)
- JavaScript client reads session ID from cookie (takes precedence over URL path)
- Allows URL to stay clean (`/`) while maintaining correct session binding

**Implementation Notes:**
- frictionless calls `SetRootSessionProvider` on the ui-engine server with a callback that returns the current session's internal ID
- The cookie is set with `HttpOnly: false` (JS needs to read it), `SameSite: Lax`, `Path: /`
- If no session exists (server not started), falls back to ui-engine's default behavior (create new session and redirect)

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

**Prototype Naming Convention:**

The `session:prototype(name, init)` function accepts arbitrary prototype names and does not consult global variables. The `name` becomes the prototype's `type` field, which is used for viewdef resolution. This allows apps to maintain their own namespaces with minimal global pollution.

Each app creates two globals:
- **Name** (PascalCase) — The app prototype, which also serves as a namespace for related prototypes
- **name** (camelCase) — The instance that frictionless uses to display the app

| App Directory | Prototype/Namespace | Instance Variable |
|---------------|---------------------|-------------------|
| `contacts`    | `Contacts`          | `contacts`        |
| `tasks`       | `Tasks`             | `tasks`           |
| `my-cool-app` | `MyCoolApp`         | `myCoolApp`       |

**Nested Prototypes:**

Other prototypes are assigned to fields on the app prototype and registered with dotted names:

```lua
-- Register Contact as a nested prototype under Contacts
Contacts.Contact = session:prototype('Contacts.Contact', {
    name = "",
    email = "",
})
local Contact = Contacts.Contact  -- Local shortcut for cleaner method declarations

function Contact:new(data)
    return session:create(Contact, data)
end

function Contact:fullName()
    return self.name
end
```

This pattern:
- Keeps the global namespace clean (only `Contacts` and `contacts` are global)
- Groups related prototypes under the app namespace
- Allows local shortcuts for cleaner code within the app file

**Prototype Inheritance:**

Prototypes can inherit from other prototypes using the optional third parameter to `session:prototype()`:

```lua
-- Define a base prototype
Animal = session:prototype('Animal', { name = "" })
function Animal:speak() return "..." end

-- Inherit from Animal
Dog = session:prototype('Dog', { breed = "" }, Animal)
function Dog:speak() return "Woof!" end
```

Instances of `Dog` inherit methods from `Animal` through the prototype chain.

**The Object Prototype:**

`main.lua` defines an `Object` prototype as the default base for all prototypes. It provides:

```lua
-- Returns "a <Type>" or "an <Type>" with correct article
function Object:tostring()
    local t = self.type or "Object"
    local first = t:sub(1,1):lower()
    local article = first:match("[aeiou]") and "an" or "a"
    return article .. " " .. t
end
```

Apps can use this for debugging and display:
```lua
local contact = Contact:new({ name = "Alice" })
print(contact:tostring())  -- "a Contact"

local item = Item:new()
print(item:tostring())     -- "an Item"
```

**`session.metaTostring(obj)`:**

A helper function that enables Lua's `tostring()` to work with prototype methods:

```lua
-- Checks if obj has a "tostring" property (or inherited one) that's a function
-- If so, calls obj:tostring()
-- Otherwise, falls back to Lua's built-in tostring(obj)
local str = session.metaTostring(contact)  -- calls contact:tostring() if defined
```

**Automatic `__tostring` Setup:**

When `session:prototype(name, init)` creates a prototype, it automatically sets:
```lua
prototype.__tostring = session.metaTostring
```

This means instances can be printed directly with Lua's `tostring()` and `print()`:
```lua
local contact = Contact:new({ name = "Alice" })
print(contact)           -- "a Contact" (via __tostring → metaTostring → contact:tostring())
print(tostring(contact)) -- "a Contact"
```

**Complete App Example:**

```lua
-- Declare app prototype (serves as namespace)
Contacts = session:prototype("Contacts", {
    _contacts = {},
    selectedContact = nil,
})

-- Nested prototype with dotted name
Contacts.Contact = session:prototype('Contacts.Contact', {
    name = "",
    email = "",
})
local Contact = Contacts.Contact

function Contact:new(data)
    return session:create(Contact, data)
end

function Contacts:new()
    return session:create(Contacts, {
        _contacts = {},
    })
end

function Contacts:addContact(name, email)
    local contact = Contact:new({ name = name, email = email })
    table.insert(self._contacts, contact)
    return contact
end

-- Guard instance creation (idempotent)
if not session.reloading then
    contacts = Contacts:new()
end
```

The agent then uses `ui_display("contacts")` to show it in the browser (loads `apps/contacts/app.lua` and displays `contacts`).

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
4. Go scans `{base_dir}/apps/*/` and loads `init.lua` from each app directory if it exists

**Note:** Since `main.lua` runs before the `mcp` global is created, `mcp.lua` must be loaded by Go code after creating the mcp table, not by main.lua.

#### App Initialization (`init.lua`)

Apps can provide `{base_dir}/apps/{app}/init.lua` to run initialization code at startup, before the app is displayed.

**Use cases:**
- Register app metadata with the mcp shell
- Set up shared prototypes or utilities
- Pre-load data or configuration

**Example `apps/contacts/init.lua`:**
```lua
-- Register this app with the mcp shell
if mcp.registerApp then
    mcp:registerApp("contacts", {
        name = "Contacts",
        description = "Contact manager"
    })
end
```

**Available at init time:**
- `mcp` global with all methods
- `session` for creating prototypes

**Not available:**
- Browser connection (app not displayed yet)

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
| `sessionId` | string | The current external session ID (internal UUID, not the vended "1"). |

#### Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `pushState` | `mcp.pushState(event)` | Push an event table to the state queue. Signals waiting HTTP clients. See Section 8.1. |
| `pollingEvents` | `mcp:pollingEvents()` | Returns `true` if an agent is connected to `/wait`. See Section 8.2. |
| `waitTime` | `mcp:waitTime()` | Returns seconds since agent last responded, or 0 if connected. See Section 8.3. |
| `app` | `mcp:app(appName)` | Load an app without displaying it. Returns the app global, or `nil, errmsg`. |
| `display` | `mcp:display(appName)` | Load and display an app. Returns `true`, or `nil, errmsg`. |
| `status` | `mcp:status()` | Returns the current MCP server status as a table. See below. |

#### `mcp:status()`

**Purpose:** Returns the current MCP server status, equivalent to the `ui_status` tool response.

**Returns:** A table with the following fields:

| Field      | Lua Type | Description                                      |
|------------|----------|--------------------------------------------------|
| `version`  | `string` | Semver string (e.g., `"0.6.0"`)                  |
| `base_dir` | `string` | Absolute or relative path (e.g., `".ui"`) |
| `url`      | `string` | Server URL (e.g., `"http://127.0.0.1:39482"`)    |
| `mcp_port` | `number` | MCP server port (e.g., `8001`)                   |
| `sessions` | `number` | Integer count of connected browsers              |

**Example:**
```lua
local status = mcp:status()
print("Server running at " .. status.url)
print("MCP port: " .. status.mcp_port)
print("Connected browsers: " .. status.sessions)
```

## 5. Tools

### 5.1 `ui_configure`
**Purpose:** Configure and start the UI server. Optional—server auto-configures at startup using `--dir` (defaults to `.ui`).

**Parameters:**
- `base_dir` (string, required): Absolute path to the UI working directory. **Use `{project}/.ui` unless the user explicitly requests a different location.**

**Behavior:**
1.  **Stop Existing Server:** If already running, stops current HTTP server and destroys session.
2.  **Directory Creation:**
    - Creates `base_dir` if it does not exist.
    - Creates a `log` subdirectory within `base_dir`.
    - **Clears all existing log files** in the `log` subdirectory (deletes or truncates).
    - **Reopens Go log file handles** (`mcp.log`) to point to the cleared/new files.
3.  **Auto-Install:** If `{base_dir}/README.md` does not exist, runs `ui_install` automatically.
4.  **Configuration Loading:**
    - Checks for existing configuration files in `base_dir`.
    - If found, loads them.
    - If not found, initializes default configuration.
5.  **Runtime Setup:**
    - Configures Lua I/O redirection as described in Section 4.
6.  **Port Selection:** Selects random available ephemeral ports for UI and MCP servers.
7.  **Server Start:** Launches the HTTP servers on `127.0.0.1`.
8.  **Port File Creation:** Writes port numbers to files in `base_dir`:
    - `{base_dir}/ui-port` - The UI server port (serves HTML/JS/WebSocket)
    - `{base_dir}/mcp-port` - The MCP server port (serves /state, /wait, /variables endpoints)

**Returns:**
- JSON object with configuration details including the server URL:
```json
{
  "base_dir": "/path/to/.ui",
  "url": "http://127.0.0.1:39482",
  "install_needed": false
}
```

### 5.2 `ui_run`
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

### 5.3 `ui_open_browser`
**Purpose:** Opens the system's default web browser to the UI session.

**Parameters:**
- `sessionId` (string, optional): The session to open. Defaults to "1".
- `path` (string, optional): The URL path to open. Defaults to "/".
- `conserve` (boolean, optional): If true, attempts to focus an existing tab or notifies the user instead of opening a duplicate session. Defaults to `true`.

**Behavior:**
- Constructs the full URL using the running server's port.
- **URL Pattern:** `http://127.0.0.1:{PORT}{PATH}?conserve=true`
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

### 5.4 `ui_status`
**Purpose:** Returns the current status of the MCP server including browser connection status.

**Parameters:** None.

**Behavior:**
- Reports the current server status and connection information.

**Returns:**
- JSON object with status information:
  - `version`: Bundled version from README.md
  - `base_dir`: Configured base directory
  - `url`: Server URL
  - `mcp_port`: MCP server port number
  - `sessions`: Number of active browser sessions

**Example Response:**
```json
{
  "version": "0.1.0",
  "base_dir": ".ui",
  "url": "http://127.0.0.1:39482",
  "mcp_port": 8001,
  "sessions": 1
}
```

### 5.5 `ui_install`
**Purpose:** Installs bundled configuration files to enable full frictionless integration.

**Parameters:**
- `force` (boolean, optional): If true, overwrites existing files regardless of version. Defaults to `false`.

**Version Checking:**

The README.md contains a semantic version near the top:
```markdown
# frictionless

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

Skills installed to `{project}/.claude/skills/`:
```
.claude/skills/ui/SKILL.md
.claude/skills/ui-basics/SKILL.md
.claude/skills/ui-fast/SKILL.md
.claude/skills/ui-thorough/SKILL.md
.claude/skills/ui-testing/SKILL.md
.claude/skills/ui-testing/TESTING-TEMPLATE.md
```

Resources installed to `{base_dir}/resources/`:
```
resources/intro.md
resources/reference.md
resources/viewdefs.md
resources/lua.md
resources/mcp.md
resources/ui_audit.md
```

Apps installed to `{base_dir}/apps/` (each with app.lua, viewdefs/, design.md, requirements.md):
```
apps/app-console/
apps/claude-panel/
apps/mcp/
apps/viewlist/
```

Viewdefs installed to `{base_dir}/viewdefs/` (symlinks to app viewdefs, bundled as symlinks):
```
viewdefs/lua.ViewList.DEFAULT.html
viewdefs/lua.ViewListItem.list-item.html
viewdefs/MCP.DEFAULT.html
viewdefs/MCP.AppMenuItem.list-item.html
viewdefs/MCP.Notification.list-item.html
viewdefs/AppConsole.*.html (7 files)
viewdefs/ClaudePanel.*.html (4 files)
```

Scripts installed to `{base_dir}/` (executable):
```
mcp
linkapp
```

Lua files installed to `{base_dir}/lua/` (symlinks to app lua files):
```
lua/main.lua
lua/mcp.lua
lua/app-console.lua
lua/claude-panel.lua
```

HTML files installed to `{base_dir}/html/` (dynamically discovered from bundle):
```
html/index.html
html/main-*.js
```

Patterns installed to `{base_dir}/patterns/` (dynamically discovered from bundle):
```
patterns/js-to-lua-bridge.md
patterns/edit-cancel.md
...
```

Documentation installed to `{base_dir}`:
```
README.md
```

**Path Resolution:**
- `{project}` is the parent of `base_dir` (e.g., if `base_dir` is `.ui`, project is `.`)
- Creates `.claude/` and `.claude/skills/` directories if they don't exist

**Behavior:**
1. **Check State:** Server must be running.
2. **Skill/Resource Files:**
   - If file doesn't exist: install from bundle
   - If exists and `force=false`: skip (no-op)
   - If exists and `force=true`: overwrite

**Returns:**
- JSON object listing installed files:
```json
{
  "installed": [".claude/skills/ui-builder/SKILL.md", ".ui/resources/reference.md"],
  "skipped": [],
  "appended": [],
  "suggestions": ["Run `claude plugin install code-simplifier` to enable code simplification"]
}
```

**External Dependency Checks:**

After installation, `ui_install` checks for optional external dependencies and includes suggestions in the response:

| Dependency | Check | Suggestion |
|------------|-------|------------|
| `code-simplifier` agent | `{project}/.claude/agents/code-simplifier.md` exists | `Run 'claude plugin install code-simplifier' to enable code simplification` |

Suggestions are informational only and do not affect the success of installation.

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

The `/variables` HTTP endpoint renders the same data as an interactive variable browser. The MCP shell includes an embedded variable browser viewdef (`MCP.VariableBrowser.DEFAULT.html`) that extends the upstream implementation with Impact and Max Impact columns.

**Upstream reference:** `ui-engine/internal/server/variables_html.go` — the standalone variable browser page served at `/{session}/variables`. Changes to columns, time parsing, or the JSON API in that file should be synced to our viewdef.

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

### 8.3 Wait Time Tracking

**Purpose:** Track how long the UI has been waiting for the agent to respond.

**Server Behavior:**
1. On startup, the server records a timestamp (`waitStartTime`).
2. Whenever the `/wait` endpoint returns (either with events or timeout), the server updates `waitStartTime` to the current time.
3. When the `/wait` endpoint is connected (a client is actively waiting), `waitStartTime` is conceptually "now".

**Lua API:**
```lua
-- Returns seconds since agent last responded
local seconds = mcp:waitTime()

-- Returns 0 if agent is currently connected to /wait
if mcp:pollingEvents() then
    assert(mcp:waitTime() == 0)
end
```

**Use Case:**
UI can display a "waiting for Claude..." indicator that shows elapsed time, helping users understand that the agent is processing their request.

### 8.4 HTTP Wait Endpoint

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

### 8.5 Agent Integration Pattern

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
