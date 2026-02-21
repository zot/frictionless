# Design: frictionless

**Source Spec:** specs/mcp.md

## Intent

MCP (Model Context Protocol) server for AI assistants to control browser-based UIs. Provides tools for session management, Lua execution, and browser interaction.

## Artifacts

### CRC Cards
- [x] crc-MCPServer.md → `internal/mcp/server.go`
- [x] crc-MCPResource.md → `internal/mcp/resources.go`
- [x] crc-MCPTool.md → `internal/mcp/tools.go`
- [x] crc-Auditor.md → `internal/mcp/audit.go`
- [x] crc-ThemeManager.md → `internal/mcp/theme.go`
- [x] crc-MCPScript.md → `install/mcp`
- [x] crc-CheckpointManager.md → `install/mcp`
- [x] crc-LinkappScript.md → `install/linkapp`
- [x] crc-Publisher.md → `internal/publisher/publisher.go`
- [x] crc-MCPSubscribe.md → `internal/mcp/subscribe.go`

### Sequences
- [x] seq-mcp-lifecycle.md → `internal/mcp/server.go`, `internal/mcp/tools.go`
- [x] seq-mcp-create-session.md → `internal/mcp/server.go`
- [x] seq-mcp-receive-event.md → `internal/mcp/tools.go`
- [x] seq-mcp-run.md → `internal/mcp/tools.go`
- [x] seq-mcp-get-state.md → `internal/mcp/resources.go`
- [x] seq-mcp-state-wait.md → `internal/mcp/server.go`
- [x] seq-audit.md → `internal/mcp/audit.go`, `internal/mcp/tools.go`
- [x] seq-theme-inject.md → `internal/mcp/theme.go`, `internal/mcp/server.go`
- [x] seq-theme-list.md → `internal/mcp/theme.go`
- [x] seq-theme-audit.md → `internal/mcp/theme.go`, `internal/mcp/tools.go`
- [x] seq-publisher-lifecycle.md → `internal/publisher/publisher.go`, `internal/mcp/subscribe.go`
- [x] seq-publish-subscribe.md → `internal/publisher/publisher.go`, `internal/mcp/subscribe.go`

### UI Layouts
- [x] ui-variable-browser.md → `install/html/variables.html`

### Test Designs
- [ ] test-MCP.md → `internal/mcp/tools_test.go`
- [x] test-Auditor.md → `internal/mcp/audit_test.go`

## Systems

### MCP Integration System
AI assistant integration via Model Context Protocol
- crc-MCPServer.md, crc-MCPResource.md, crc-MCPTool.md
- seq-mcp-lifecycle.md, seq-mcp-create-session.md
- seq-mcp-receive-event.md, seq-mcp-run.md, seq-mcp-get-state.md, seq-mcp-state-wait.md

### Publisher System
Shared pub/sub server for browser-to-MCP data flow (bookmarklets, external tools)
- crc-Publisher.md, crc-MCPSubscribe.md
- seq-publisher-lifecycle.md, seq-publish-subscribe.md

### Transport System
Support multiple MCP transport modes:
- **Stdio** (`mcp` command): JSON-RPC 2.0 over stdin/stdout
- **SSE** (`serve` command): Server-Sent Events over HTTP
- **Install** (`install` command): Manual installation without MCP server
- **Default base_dir:** `{project}/.ui` for all commands
- Publisher (port 25283) is hosted in-process by the MCP server — no separate command

### Startup Behavior
- Server uses `--dir` (defaults to `.ui`)
- Auto-install if `{base_dir}` or `{base_dir}/README.md` missing:
  - Claude skills (`/ui`, `/ui-basics`, `/ui-fast`, `/ui-thorough`, `/ui-testing`) to `{project}/.claude/skills/`
  - Apps (app-console, claude-panel, mcp, viewlist) to `{base_dir}/apps/`
  - Web frontend (html/*) to `{base_dir}/html/`
  - MCP resources to `{base_dir}/resources/`
  - Lua files and viewdef symlinks to `{base_dir}/lua/` and `{base_dir}/viewdefs/`
  - Helper scripts (mcp, linkapp) to `{base_dir}/`
- Auto-starts HTTP server
- `ui_configure` optional—triggers full reconfigure (stop, reinitialize, restart)

### Versioning
- Source of truth: `README.md` (`**Version: X.Y.Z**`)
- CLI: `--version` flag or `version` subcommand (build-time ldflags)
- MCP: `ui_status` returns bundled version from README.md

### HTTP Endpoints (MCP port)
Debug and inspect runtime state:
- `GET /wait`: Long-poll for mcp.pushState() events
- `GET /variables`: Redirects to UI port `/variables` (static HTML served from `{base_dir}/html/variables.html`)
- `GET /state`: Redirects to UI port `/state`

Tool API (Spec 2.5) - enables curl access for spawned agents:
- `GET /api/ui_status`: Get server status
- `POST /api/ui_run`: Execute Lua code
- `POST /api/ui_display`: Load and display an app
- `POST /api/ui_configure`: Reconfigure server
- `POST /api/ui_install`: Install bundled files
- `POST /api/ui_open_browser`: Open browser to UI
- `POST /api/ui_audit`: Audit app for code quality violations
- `GET /api/resource/`: List resources directory (JSON for curl, HTML for browsers)
- `GET /api/resource/{path}`: Serve resource file (markdown rendered as HTML via goldmark for browsers, raw for curl)
- `GET /app/{app}/readme`: Serve app's README.md as HTML (case-insensitive lookup, rendered via goldmark)

### Lua Loading Sequence (Spec 4.2)
During startup, Go executes:
1. ui-engine loads `main.lua` (mcp global does NOT exist yet)
2. `setupMCPGlobal()` creates the `mcp` global with core methods
3. `loadMCPLua()` loads `{base_dir}/lua/mcp.lua` if it exists
4. `loadAppInitFiles()` scans `{base_dir}/apps/*/` and loads each `init.lua`

App init files can register metadata with the mcp shell since `mcp` global exists.

### Lua `mcp` Global Object (Spec 4.3)
Registered by `setupMCPGlobal` in each session:

**Properties:**
| Property | Type | Description |
|----------|------|-------------|
| `type` | `string` | Always `"MCP"` |
| `value` | `table\|nil` | Current app value |
| `sessionId` | `string` | External session ID (internal UUID) |

**Methods:**
| Method | Signature | Returns |
|--------|-----------|---------|
| `pushState` | `mcp.pushState(event: table)` | `nil` |
| `pollingEvents` | `mcp:pollingEvents()` | `boolean` |
| `waitTime` | `mcp:waitTime()` | `number` (seconds since agent last responded, 0 if connected) |
| `app` | `mcp:app(appName: string)` | `app` or `nil, errmsg` |
| `display` | `mcp:display(appName: string)` | `true` or `nil, errmsg` |
| `status` | `mcp:status()` | `table` (see below) |
| `subscribe` | `mcp:subscribe(topic: string, handler: function)` | `nil` |
| `reinjectThemes` | `mcp:reinjectThemes()` | `true` or `nil, errmsg` |

**`mcp:status()` returns:**
| Field | Type | Description |
|-------|------|-------------|
| `version` | `string` | Semver (e.g., `"0.6.0"`) |
| `base_dir` | `string` | Path (e.g., `".ui"`) |
| `url` | `string` | Server URL |
| `mcp_port` | `number` | MCP server port |
| `sessions` | `number` | Browser count |

### Lua `session` Object (Spec 4.0)
Provided by ui-engine, used by apps for prototype management:

**Prototype Methods:**
| Method | Signature | Description |
|--------|-----------|-------------|
| `prototype` | `session:prototype(name, init, parent?)` | Create/update a prototype with optional inheritance |
| `create` | `session:create(proto, instance)` | Create a tracked instance |

**Prototype Helper:**
| Function | Signature | Description |
|----------|-----------|-------------|
| `metaTostring` | `session.metaTostring(obj)` | If obj has a `tostring` method, calls it; otherwise uses Lua's `tostring()` |

**Automatic `__tostring` Wiring:**
- `main.lua` wraps `session:prototype()` to set `prototype.__tostring = session.metaTostring` on every prototype
- Required because GopherLua doesn't inherit `__tostring` through metatables properly
- Enables `print(obj)` and `tostring(obj)` to call `obj:tostring()` if defined
- Falls back to the object's type for objects without a `tostring` method

### Build & Release System
Cross-platform binary builds via Makefile:
- `make cache`: Extracts web assets from ui-engine-bundled, copies html/* to install/html/
- `make build`: Builds binary, bundles install/ directory into binary
- `make release`: Builds for Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64)
- Output in `release/` directory with platform-specific naming
- Version from README.md (`**Version: X.Y.Z**`)
- `ui_status` returns bundled version; `ui_install` uses semver comparison

## Gaps

### Spec→Design
*None*

### Design→Code
*None*

### Code→Design
*None*

### Oversights
- [x] O1: `session:prototype()` automatic `__tostring` wiring ~~requires ui-engine change~~
  - Fixed: main.lua wraps `session:prototype()` to set `__tostring` on every prototype
  - GopherLua doesn't inherit `__tostring` through metatables, so explicit setting is required
- [ ] O2: Test coverage - only `tools_test.go` and `notify_test.go` exist
  - [ ] State Change Waiting (10 scenarios)
  - [ ] Lifecycle (startup, reconfigure)
  - [ ] ui_open_browser (3 scenarios)
  - [x] ui_run (6 tests: execute code, session access, JSON marshalling, non-JSON result, mcp global, no session)
  - [ ] Frictionless UI Creation (6 scenarios)
  - [x] ClearLogs (5 tests: clears files, calls callback, handles missing dir, skips subdirs, no callback)
  - [x] ui_audit (27 tests via temp fixtures: R34 badge/R35 method args/R36 path syntax)
- [ ] O3: Document frontend conserve mode SharedWorker requirements (spec 6.1)
- [ ] O4: Install tests fail without bundled binary (`make build`)
  - [ ] TestInstallSkillFilesFreshInstall
  - [ ] TestInstallSkillFilesNoOpIfExists
  - [ ] TestInstallSkillFilesCreatesDirectory
  - [ ] TestInstallSkillFilesPathResolution
  - [ ] TestInstallForceOverwrites
- [ ] O5: Pluggable themes - remaining implementation
  - [x] Theme CSS files (base.css, lcars.css, clarity.css, midnight.css, ninja.css)
  - [x] Server startup injection to index.html
  - [x] CSS `@` annotation parser for theme metadata
  - [x] Updated theme CLI commands (list, classes, audit)
  - [x] Prefs app for theme switching with localStorage persistence
  - [ ] Theme test coverage (ParseThemeCSS, InjectThemeBlock, AuditAppTheme)
  - [x] All-themes class listing (GetAllThemeClasses, no-arg classes/audit)
  - [x] Structural semantic classes (@class declarations in all themes, brume rules)
  - [x] Stock app viewdef semantic class usage
- [ ] O6: Publisher test coverage
  - [ ] Publisher server (publish, subscribe, fan-out, TTL wait, idle shutdown)
  - [ ] MCPSubscribe (poll loop, ensurePublisher, handler dispatch)
