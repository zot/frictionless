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
  - MCP resources, viewdefs, and helper scripts to `{base_dir}/`
- Starts in CONFIGURED state (no UNCONFIGURED state)
- `ui_configure` optional—for reconfiguration only

### Versioning
- Source of truth: `README.md` (`**Version: X.Y.Z**`)
- CLI: `--version` flag or `version` subcommand (build-time ldflags)
- MCP: `ui_status` returns bundled version from README.md

### HTTP Endpoints (MCP port)
Debug and inspect runtime state:
- `GET /wait`: Long-poll for mcp.pushState() events
- `GET /variables`: Interactive variable tree view
- `GET /state`: Current session state JSON

### Build & Release System
Cross-platform binary builds via Makefile:
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
  - [ ] Lifecycle FSM (4 scenarios)
  - [ ] ui_open_browser (3 scenarios)
  - [ ] ui_run (4 scenarios)
  - [ ] ui_upload_viewdef (3 scenarios)
  - [ ] Frictionless UI Creation (6 scenarios)
- [ ] O2: Document frontend conserve mode SharedWorker requirements (spec 6.1)
- [ ] O3: Include current state in FSM error messages for debugging
