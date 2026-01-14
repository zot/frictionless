# Design: ui-mcp

**Source Spec:** specs/mcp.md

## Intent

MCP (Model Context Protocol) server for AI assistants to control browser-based UIs. Provides tools for session management, Lua execution, and browser interaction.

## Artifacts

### CRC Cards
- [x] crc-MCPServer.md → `internal/mcp/server.go`
- [x] crc-MCPResource.md → `internal/mcp/resources.go`
- [x] crc-MCPTool.md → `internal/mcp/tools.go`

### Sequences
- [x] seq-mcp-lifecycle.md → `internal/mcp/server.go`, `internal/mcp/tools.go`
- [x] seq-mcp-create-session.md → `internal/mcp/server.go`
- [x] seq-mcp-create-presenter.md → `internal/mcp/server.go`
- [x] seq-mcp-receive-event.md → `internal/mcp/tools.go`
- [x] seq-mcp-run.md → `internal/mcp/tools.go`
- [x] seq-mcp-get-state.md → `internal/mcp/resources.go`
- [x] seq-mcp-state-wait.md → `internal/mcp/server.go` (handleWait, pushStateEvent, drainStateQueue, hasPollingClients)

### Test Designs
- [ ] test-MCP.md → `tools_test.go`

## Systems

### MCP Integration System
AI assistant integration via Model Context Protocol
- crc-MCPServer.md, crc-MCPResource.md, crc-MCPTool.md
- seq-mcp-lifecycle.md, seq-mcp-create-session.md, seq-mcp-create-presenter.md
- seq-mcp-receive-event.md, seq-mcp-run.md, seq-mcp-get-state.md, seq-mcp-state-wait.md

### Transport System
Support multiple MCP transport modes:
- **Stdio** (`mcp` command): JSON-RPC 2.0 over stdin/stdout
- **SSE** (`serve` command): Server-Sent Events over HTTP
- **Install** (`install` command): Manual installation without MCP server
- **Default base_dir:** `{project}/.claude/ui` for all commands

### Startup Behavior
- Server uses `--dir` (defaults to `.claude/ui`)
- Auto-install if `{base_dir}` or `{base_dir}/README.md` missing:
  - Claude skills (`/ui`, `/ui-builder`) to `{project}/.claude/skills/`
  - Claude agents to `{project}/.claude/agents/`
  - Web frontend (html/*) to `{base_dir}/html/`
  - MCP resources, viewdefs, and helper scripts to `{base_dir}/`
- Auto-starts HTTP server
- `ui_configure` optional—triggers full reconfigure (stop, reinitialize, restart)

### Versioning
- Source of truth: `README.md` (`**Version: X.Y.Z**`)
- CLI: `--version` flag or `version` subcommand (build-time ldflags)
- MCP: `ui_status` returns bundled version from README.md

### HTTP Endpoints (MCP port)
Debug and inspect runtime state:
- `GET /wait`: Long-poll for mcp.pushState() events
- `GET /variables`: Interactive variable tree view
- `GET /state`: Current session state JSON

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

**Methods:**
| Method | Signature | Returns |
|--------|-----------|---------|
| `pushState` | `mcp.pushState(event: table)` | `nil` |
| `pollingEvents` | `mcp:pollingEvents()` | `boolean` |
| `display` | `mcp:display(appName: string)` | `true` or `nil, string` |
| `status` | `mcp:status()` | `table` (see below) |

**`mcp:status()` returns:**
| Field | Type | Description |
|-------|------|-------------|
| `version` | `string` | Semver (e.g., `"0.6.0"`) |
| `base_dir` | `string` | Path (e.g., `".claude/ui"`) |
| `url` | `string` | Server URL |
| `sessions` | `number` | Browser count |

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
- [ ] O1: Test coverage - only `tools_test.go` and `notify_test.go` exist
  - [ ] State Change Waiting (10 scenarios)
  - [ ] Lifecycle (startup, reconfigure)
  - [ ] ui_open_browser (3 scenarios)
  - [x] ui_run (6 tests: execute code, session access, JSON marshalling, non-JSON result, mcp global, no session)
  - [ ] ui_upload_viewdef (3 scenarios)
  - [ ] Frictionless UI Creation (6 scenarios)
  - [x] ClearLogs (5 tests: clears files, calls callback, handles missing dir, skips subdirs, no callback)
- [ ] O2: Document frontend conserve mode SharedWorker requirements (spec 6.1)
- [ ] O3: Install tests fail without bundled binary (`make build`)
  - [ ] TestInstallSkillFilesFreshInstall
  - [ ] TestInstallSkillFilesNoOpIfExists
  - [ ] TestInstallSkillFilesCreatesDirectory
  - [ ] TestInstallSkillFilesPathResolution
  - [ ] TestInstallForceOverwrites
