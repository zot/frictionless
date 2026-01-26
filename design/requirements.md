# Requirements

## Feature: MCP Server
**Source:** specs/mcp.md

- **R1:** Support stdio transport (JSON-RPC 2.0 over stdin/stdout) for AI agent integration
- **R2:** Support SSE transport (Server-Sent Events over HTTP) for standalone development
- **R3:** Auto-install bundled files if base_dir or README.md missing
- **R4:** Provide ui_configure tool to reconfigure and restart server
- **R5:** Provide ui_run tool to execute Lua code in session context
- **R6:** Provide ui_open_browser tool with conserve mode to prevent duplicate tabs
- **R7:** Provide ui_status tool returning version, base_dir, url, mcp_port, sessions
- **R8:** Provide ui_install tool with version checking and force option
- **R9:** Expose state via MCP resources (ui://state, ui://variables)
- **R10:** Provide HTTP endpoints for debugging (/state, /variables, /wait)
- **R11:** Support HTTP long-polling for state changes via /wait endpoint
- **R12:** Register mcp global in Lua with pushState, pollingEvents, app, display, status, waitTime methods
- **R13:** Track wait time: server records timestamp on start, updates when /wait returns
- **R14:** Provide mcp:waitTime() Lua method returning seconds since agent last responded
- **R15:** Return 0 from waitTime when agent is currently connected to /wait
- **R16:** Load mcp.lua extension file if present
- **R17:** Load app init.lua files on startup
- **R18:** Bind root URL to current session (session cookie)
- **R19:** Hot-reload Lua and viewdef files on change
- **R20:** Redirect Lua print/io to log files
- **R21:** Provide HTTP Tool API for spawned agents (GET/POST /api/*)

## Feature: UI Audit
**Source:** specs/ui-audit.md

- **R22:** Detect dead methods (defined but never called)
- **R23:** Detect missing reloading guard on instance creation
- **R24:** Detect global name mismatch with app directory
- **R25:** Detect malformed HTML in viewdefs
- **R26:** Detect style tags in list-item viewdefs
- **R27:** Detect item. prefix in list-item bindings
- **R28:** Detect ui-action on non-button elements
- **R29:** Detect wrong hidden syntax (ui-class vs ui-class-hidden)
- **R30:** Detect ui-value on checkbox/switch elements
- **R31:** Detect operators in binding paths (excludes ui-namespace which is a viewdef namespace, not a path)
- **R32:** Detect missing Lua methods referenced in viewdefs
- **R33:** Return JSON with violations, warnings, and summary
