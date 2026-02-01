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
- **R22:** Provide GET /app/{app}/readme endpoint returning app's README.md as HTML (case-insensitive lookup, rendered via goldmark)

## Feature: UI Audit
**Source:** specs/ui-audit.md

- **R23:** Detect dead methods (defined but never called from Lua code, viewdefs, or factory functions)
- **R24:** Recognize factory method pattern: methods created by local factory functions called at outer scope are not dead
- **R25:** Detect missing reloading guard on instance creation
- **R26:** Detect global name mismatch with app directory
- **R27:** Detect malformed HTML in viewdefs
- **R28:** Detect style tags in list-item viewdefs
- **R29:** Detect item. prefix in list-item bindings
- **R30:** Detect ui-action on non-button elements
- **R31:** Detect wrong hidden syntax (ui-class vs ui-class-hidden)
- **R32:** Detect ui-value on checkbox/switch elements
- **R33:** Detect operators in binding paths (excludes ui-namespace which is a viewdef namespace, not a path)
- **R34:** Detect missing Lua methods referenced in viewdefs
- **R35:** Return JSON with violations, warnings, and summary
- **R36:** Detect ui-value on sl-badge elements (must use span with ui-value inside badge)
- **R37:** Detect non-empty method args in paths (only `method()` or `method(_)` allowed)
- **R38:** Validate path syntax against grammar as final check

## Feature: Pluggable Themes
**Source:** specs/pluggable-themes.md

- **R39:** Store themes as CSS files in `.ui/html/themes/` with `.theme-{name}` class prefix
- **R40:** Parse theme metadata from CSS comments (`@theme`, `@class`, `@description`, `@usage`, `@elements`)
- **R41:** Provide `base.css` with `:root` fallbacks, transitions, and global styles
- **R42:** On server startup, inject `<!-- #frictionless -->` block into index.html with theme links
- **R43:** Remove existing frictionless block before injection (idempotent)
- **R44:** Scan `.ui/html/themes/*.css` for theme files (excluding base.css)
- **R45:** Generate theme restore script that reads localStorage and sets `<html>` class
- **R46:** `theme list` scans CSS files and parses metadata from comments
- **R47:** `theme classes [THEME]` parses `@class` annotations from CSS comments
- **R48:** `theme audit APP [THEME]` uses CSS files as source (same logic)
- **R49:** Install copies theme CSS files to `.ui/html/themes/`
- **R50:** Bundle multiple themes: lcars, clarity, midnight, ninja
- **R51:** MCP.DEFAULT.html uses only shell CSS wrapped in `@layer components`
- **R52:** Prefs app allows runtime theme switching with localStorage persistence
