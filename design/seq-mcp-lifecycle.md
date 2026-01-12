# Sequence: MCP Server Lifecycle

**Source Spec:** specs/mcp.md

## Participants
- AI Agent: External AI assistant (MCP Client)
- MCPServer: Protocol handler and lifecycle FSM
- MCPTool: Tool handler dispatcher
- LuaRuntime: Embedded Lua VM manager
- HTTPServer: The UI platform's HTTP service
- OS: Operating system services (filesystem, browser launch)

## Scenario 1: Initial Configuration & Setup
The AI agent initializes the environment before starting the server.

```
     ┌────────┐             ┌─────────┐             ┌───────┐             ┌──────────┐           ┌────┐
     │AI Agent│             │MCPServer│             │MCPTool│             │LuaRuntime│           │ OS │
     └────┬───┘             └────┬────┘             └───┬───┘             └─────┬────┘           └─┬──┘
          │Call("ui_configure", {base_dir})             │                       │                  │
          │─────────────────────>│                      │                       │                  │
          │                      │ Handle("ui_configure")                       │                  │
          │                      │─────────────────────>│                       │                  │
          │                      │                      │ CreateDir(base_dir)   │                  │
          │                      │                      │─────────────────────────────────────────>│
          │                      │                      │                       │                  │
          │                      │                      │ RedirectIO(base_dir)  │                  │
          │                      │                      │──────────────────────>│                  │
          │                      │                      │                       │                  │
          │                      │                      │   LoadConfig()        │                  │
          │                      │                      │─────────────────────────────────────────>│
          │                      │                      │                       │                  │
          │                      │  SetState(CONFIGURED)│                       │                  │
          │                      │<─────────────────────│                       │                  │
          │                      │                      │                       │                  │
          │                      │                      │ CheckInstallNeeded()  │                  │
          │                      │                      │─────────────────────────────────────────>│
          │                      │                      │     [check file]      │                  │
          │                      │                      │                       │                  │
          │  Success + install_needed hint              │                       │                  │
          │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                      │                       │                  │
     ┌────┴───┐             ┌────┴────┐             ┌───┴───┐             ┌─────┴────┘           ┌─┴──┐
     │AI Agent│             │MCPServer│             │MCPTool│             │LuaRuntime│           │ OS │
     └────────┘             └─────────┘             └───┴───┘             └──────────┘           └────┘
```

**Notes:**
- Configuration checks if bundled files are installed
- Returns `install_needed: true` hint if `.claude/skills/ui-builder/SKILL.md` is missing
- Agent explicitly calls `ui_install` to install files (separated from configure)

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

## Scenario 2: Server Startup
The AI agent starts the HTTP server after configuration.

```
     ┌────────┐             ┌─────────┐             ┌───────┐             ┌──────────┐           ┌────┐
     │AI Agent│             │MCPServer│             │MCPTool│             │HTTPServer│           │ OS │
     └────┬───┘             └────┬────┘             └───┬───┘             └─────┬────┘           └─┬──┘
          │ Call("ui_start")     │                      │                       │                  │
          │─────────────────────>│                      │                       │                  │
          │                      │  Handle("ui_start")  │                       │                  │
          │                      │─────────────────────>│                       │                  │
          │                      │                      │                       │                  │
          │                      │                      │ SelectPorts(0, 0)     │                  │
          │                      │                      │─────────────────────────────────────────>│
          │                      │                      │     (ui, mcp ports)   │                  │
          │                      │                      │                       │                  │
          │                      │                      │   StartUI(uiPort)     │                  │
          │                      │                      │──────────────────────>│                  │
          │                      │                      │                       │                  │
          │                      │                      │   StartMCP(mcpPort)   │                  │
          │                      │                      │──────────────────────>│                  │
          │                      │                      │                       │                  │
          │                      │                      │ WriteFile(ui-port)    │                  │
          │                      │                      │─────────────────────────────────────────>│
          │                      │                      │                       │                  │
          │                      │                      │ WriteFile(mcp-port)   │                  │
          │                      │                      │─────────────────────────────────────────>│
          │                      │                      │                       │                  │
          │                      │   SetState(RUNNING)  │                       │                  │
          │                      │<─────────────────────│                       │                  │
          │                      │                      │                       │                  │
          │     Success(URL)     │                      │                       │                  │
          │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                      │                       │                  │
     ┌────┴───┐             ┌────┴────┐             ┌───┴───┐             ┌─────┴────┘           ┌─┴──┐
     │AI Agent│             │MCPServer│             │MCPTool│             │HTTPServer│           │ OS │
     └────────┘             └─────────┘             └───┴───┘             └──────────┘           └────┘
```

**Notes:**
- Port files written to `{base_dir}/ui-port` and `{base_dir}/mcp-port`
- UI server serves HTML/JS and WebSocket connections
- MCP server serves /state, /wait, /variables endpoints

## Scenario 3: Opening Browser
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
