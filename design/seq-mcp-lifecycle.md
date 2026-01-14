# Sequence: MCP Server Lifecycle

**Source Spec:** specs/mcp.md

## Participants
- AI Agent: External AI assistant (MCP Client)
- MCPServer: Protocol handler and lifecycle FSM
- MCPTool: Tool handler dispatcher
- LuaRuntime: Embedded Lua VM manager
- HTTPServer: The UI platform's HTTP service
- OS: Operating system services (filesystem, browser launch)

## Scenario 1: Startup (Auto-Configure and Start)
Server configures and auto-starts using `--dir` (defaults to `.claude/ui`).

```
     ┌─────────┐             ┌───────┐             ┌──────────┐           ┌────┐
     │MCPServer│             │MCPTool│             │HTTPServer│           │ OS │
     └────┬────┘             └───┬───┘             └─────┬────┘           └─┬──┘
          │ Initialize(base_dir) │                       │                  │
          │─────────────────────>│                       │                  │
          │                      │ CreateDir(base_dir)   │                  │
          │                      │─────────────────────────────────────────>│
          │                      │                       │                  │
          │                      │ CheckReadme(base_dir) │                  │
          │                      │─────────────────────────────────────────>│
          │                      │      [exists?]        │                  │
          │                      │                       │                  │
          │                      │ [if missing] Install()│                  │
          │                      │─────────────────────────────────────────>│
          │                      │                       │                  │
          │                      │ SelectPorts(0, 0)     │                  │
          │                      │─────────────────────────────────────────>│
          │                      │     (ui, mcp ports)   │                  │
          │                      │                       │                  │
          │                      │   StartUI(uiPort)     │                  │
          │                      │──────────────────────>│                  │
          │                      │                       │                  │
          │                      │   StartMCP(mcpPort)   │                  │
          │                      │──────────────────────>│                  │
          │                      │                       │                  │
          │                      │ setupMCPGlobal()      │                  │
          │                      │────┐ [creates mcp table]                 │
          │                      │<───┘                  │                  │
          │                      │                       │                  │
          │                      │ loadMCPLua()          │                  │
          │                      │─────────────────────────────────────────>│
          │                      │ [load mcp.lua if exists]                 │
          │                      │                       │                  │
          │                      │ loadAppInitFiles()    │                  │
          │                      │─────────────────────────────────────────>│
          │                      │ [scan apps/*/init.lua]│                  │
          │                      │                       │                  │
          │                      │ WriteFile(ui-port)    │                  │
          │                      │─────────────────────────────────────────>│
          │                      │                       │                  │
          │                      │ WriteFile(mcp-port)   │                  │
          │                      │─────────────────────────────────────────>│
     ┌────┴────┐             ┌───┴───┐             ┌─────┴────┐           ┌─┴──┐
     │MCPServer│             │MCPTool│             │HTTPServer│           │ OS │
     └─────────┘             └───────┘             └──────────┘           └────┘
```

**Lua Loading Sequence (during startup):**
1. ui-engine loads `main.lua` (mcp global does NOT exist yet)
2. `setupMCPGlobal()` creates the `mcp` global with core methods
3. `loadMCPLua()` loads `{base_dir}/lua/mcp.lua` if it exists
4. `loadAppInitFiles()` scans `{base_dir}/apps/*/` and loads each `init.lua`

**Notes:**
- Auto-install runs if `{base_dir}/README.md` is missing
- Server auto-starts
- Port files written to `{base_dir}/ui-port` and `{base_dir}/mcp-port`
- `ui_configure` is optional—triggers full reconfigure if needed

## Scenario 1a: Bundled File Installation (ui_install)
Shows the installation of bundled files when agent calls `ui_install`.

```
     ┌────────┐             ┌─────────┐             ┌───────┐             ┌────┐             ┌──────┐
     │AI Agent│             │MCPServer│             │MCPTool│             │ OS │             │Bundle│
     └────┬───┘             └────┬────┘             └───┬───┘             └─┬──┘             └──┬───┘
          │ Call("ui_install", {force})                 │                   │                   │
          │─────────────────────>│                      │                   │                   │
          │                      │ Handle("ui_install") │                   │                   │
          │                      │─────────────────────>│                   │                   │
          │                      │                      │                   │                   │
          │                      │                      │ ReadBundledVersion│                   │
          │                      │                      │───────────────────────────────────────>│
          │                      │                      │      "0.1.0"      │                   │
          │                      │                      │<──────────────────────────────────────│
          │                      │                      │                   │                   │
          │                      │                      │ ReadInstalledVer  │                   │
          │                      │                      │──────────────────>│                   │
          │                      │                      │    "0.0.9" or nil │                   │
          │                      │                      │<─────────────────│                   │
          │                      │                      │                   │                   │
          │                      │                      │ [if bundled > installed OR force]     │
          │                      │                      │─ ─ ─ ─ ─ ─ ─ ─ ─ ─│─ ─ ─ ─ ─ ─ ─ ─ ─ ─│
          │                      │                      │ loop [each bundle]│                   │
          │                      │                      │─ ─ ─ ─ ─ ─ ─ ─ ─ ─│─ ─ ─ ─ ─ ─ ─ ─ ─ ─│
          │                      │                      │  │ ReadFile()     │                   │
          │                      │                      │  │───────────────────────────────────>│
          │                      │                      │  │                │       content     │
          │                      │                      │  │<──────────────────────────────────│
          │                      │                      │  │                │                   │
          │                      │                      │  │ WriteFile()    │                   │
          │                      │                      │  │───────────────>│                   │
          │                      │                      │─ ─ ─ ─ ─ ─ ─ ─ ─ ─│─ ─ ─ ─ ─ ─ ─ ─ ─ ─│
          │                      │                      │                   │                   │
          │  Success({installed, skipped, version_skipped, versions})       │                   │
          │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                      │                   │                   │
     ┌────┴───┐             ┌────┴────┐             ┌───┴───┐             ┌─┴──┐             ┌──┴───┐
     │AI Agent│             │MCPServer│             │MCPTool│             │ OS │             │Bundle│
     └────────┘             └─────────┘             └───────┘             └────┘             └──────┘
```

**Version Checking:**
- Read `version` from bundled README.md (`**Version: X.Y.Z**`)
- Compare with installed version using semver comparison
- Install all files only if bundled > installed OR force=true
- Return `version_skipped: true` with both versions when skipping

**Bundled Files (from `install/` directory):**
| Source (in `install/`)           | Destination                              | Purpose                                |
|----------------------------------|------------------------------------------|----------------------------------------|
| `init/skills/ui/*`               | `{project}/.claude/skills/ui/*`          | `/ui` skill (running UIs)              |
| `init/skills/ui-builder/*`       | `{project}/.claude/skills/ui-builder/*`  | `/ui-builder` skill (building UIs)     |
| `resources/*`                    | `{base_dir}/resources/*`                 | MCP server resources                   |
| `viewdefs/*`                     | `{base_dir}/viewdefs/*`                  | Standard viewdefs (ViewList, etc.)     |
| `event`, `state`, `variables`, `linkapp` | `{base_dir}`                       | Scripts for easy MCP endpoint access   |

**Notes:**
- `{project}` is the parent of `base_dir` (e.g., if `base_dir` is `.claude/ui`, project is `.`)
- Skills are self-describing (no CLAUDE.md augmentation needed)
- `force=true` overwrites existing files

## Scenario 2: Opening Browser
The AI agent instructs the system to open a browser to the session.

```
     ┌────────┐             ┌─────────┐             ┌───────┐             ┌────┐
     │AI Agent│             │MCPServer│             │MCPTool│             │ OS │
     └────┬───┘             └────┬────┘             └───┬───┘             └─┬──┘
          │Call("ui_open_browser", {sessionId, conserve})                 │
          │─────────────────────>│                      │                   │
          │                      │Handle("ui_open_browser")                 │
          │                      │─────────────────────>│                   │
          │                      │                      │                   │
          │                      │                      │ ConstructURL()    │
          │                      │                      │────┐              │
          │                      │                      │    │              │
          │                      │                      │<───┘              │
          │                      │                      │                   │
          │                      │                      │  xdg-open(URL)    │
          │                      │                      │──────────────────>│
          │                      │                      │                   │
          │       Success        │                      │                   │
          │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                      │                   │
     ┌────┴───┐             ┌────┴────┐             ┌───┴───┐             ┌─┴──┐
     │AI Agent│             │MCPServer│             │MCPTool│             │ OS │
     └────────┘             └─────────┘             └───┴───┘             └────┘
```

## Scenario 3: Reconfiguration
The AI agent reconfigures the server with a different base directory.

```
     ┌────────┐             ┌─────────┐             ┌───────┐             ┌──────────┐           ┌────┐
     │AI Agent│             │MCPServer│             │MCPTool│             │HTTPServer│           │ OS │
     └────┬───┘             └────┬────┘             └───┬───┘             └─────┬────┘           └─┬──┘
          │Call("ui_configure", {base_dir})              │                       │                  │
          │─────────────────────>│                      │                       │                  │
          │                      │Handle("ui_configure")│                       │                  │
          │                      │─────────────────────>│                       │                  │
          │                      │                      │                       │                  │
          │                      │                      │  StopServer()         │                  │
          │                      │                      │──────────────────────>│                  │
          │                      │                      │                       │                  │
          │                      │                      │ DestroySession()      │                  │
          │                      │                      │──────────────────────>│                  │
          │                      │                      │                       │                  │
          │                      │                      │ CreateDir(base_dir)   │                  │
          │                      │                      │─────────────────────────────────────────>│
          │                      │                      │                       │                  │
          │                      │                      │ ClearLogs(base_dir)   │                  │
          │                      │                      │─────────────────────────────────────────>│
          │                      │                      │                       │                  │
          │                      │                      │ ReopenGoLogFile()     │                  │
          │                      │                      │─────────────────────────────────────────>│
          │                      │                      │                       │                  │
          │                      │                      │ [if missing] Install()│                  │
          │                      │                      │─────────────────────────────────────────>│
          │                      │                      │                       │                  │
          │                      │                      │ SelectPorts(0, 0)     │                  │
          │                      │                      │─────────────────────────────────────────>│
          │                      │                      │                       │                  │
          │                      │                      │   StartUI(uiPort)     │                  │
          │                      │                      │──────────────────────>│                  │
          │                      │                      │                       │                  │
          │                      │                      │   StartMCP(mcpPort)   │                  │
          │                      │                      │──────────────────────>│                  │
          │                      │                      │                       │                  │
          │  Success({base_dir, url})                   │                       │                  │
          │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                      │                       │                  │
     ┌────┴───┐             ┌────┴────┐             ┌───┴───┐             ┌─────┴────┐           ┌─┴──┐
     │AI Agent│             │MCPServer│             │MCPTool│             │HTTPServer│           │ OS │
     └────────┘             └─────────┘             └───────┘             └──────────┘           └────┘
```

**Notes:**
- Calling `ui_configure` triggers a full reconfigure
- Current session is destroyed, HTTP server stops
- Log files in `{base_dir}/log/` are cleared (deleted or truncated)
- Go log file handles (`mcp.log`) are reopened to point to fresh files
- Filesystem is re-initialized for new base_dir
- HTTP listeners restart on new ephemeral ports
- Returns `{base_dir, url, install_needed}` where url is `http://HOST:PORT` (no session ID)
