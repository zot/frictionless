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
- **R35:** Return JSON with violations, warnings, reminders, and summary
- **R36:** Detect ui-value on sl-badge elements (must use span with ui-value inside badge)
- **R37:** Detect non-empty method args in paths (only `method()` or `method(_)` allowed)
- **R38:** Validate path syntax against grammar as final check
- **R39:** Include behavioral reminders for checks that cannot be automated (min-height: 0, Cancel revert, slow function caching)

## Feature: Pluggable Themes
**Source:** specs/pluggable-themes.md

- **R40:** Store themes as CSS files in `.ui/html/themes/` with `.theme-{name}` class prefix
- **R41:** Parse theme metadata from CSS comments (`@theme`, `@class`, `@description`, `@usage`, `@elements`)
- **R42:** Provide `base.css` with `:root` fallbacks, transitions, and global styles
- **R43:** Inject `<!-- #frictionless -->` block into index.html with theme links at startup and whenever the file is overwritten by external processes
- **R44:** Remove existing frictionless block before injection (idempotent)
- **R45:** Scan `.ui/html/themes/*.css` for theme files (excluding base.css)
- **R46:** Generate theme restore script that reads localStorage and sets `<html>` class
- **R47:** `theme list` scans CSS files and parses metadata from comments
- **R48:** `theme classes [THEME]` parses `@class` annotations from CSS comments
- **R49:** `theme audit APP [THEME]` uses CSS files as source (same logic)
- **R50:** Install copies theme CSS files to `.ui/html/themes/`
- **R51:** Bundle multiple themes: lcars, clarity, midnight, ninja
- **R52:** MCP.DEFAULT.html uses only shell CSS wrapped in `@layer components`
- **R53:** Prefs app allows runtime theme switching with localStorage persistence

## Feature: Helper Scripts
**Source:** specs/helper-scripts.md

### mcp Script
- **R54:** Read MCP server port from `{base_dir}/mcp-port` file
- **R55:** Provide `--help` showing all available commands
- **R56:** `status` returns server status JSON via `/api/ui_status`
- **R57:** `browser` opens browser via `/api/ui_open_browser`
- **R58:** `display APP` displays app via `/api/ui_display`
- **R59:** `run 'lua'` executes Lua via `/api/ui_run`
- **R60:** `event` long-polls `/wait` with 120s timeout, returns JSON array of events
- **R61:** `event` tracks PID in `.eventpid` and kills previous watcher before starting
- **R62:** `state` returns session state JSON via `/state`
- **R63:** `variables` returns variable values via `/variables`
- **R64:** `progress APP PERCENT STAGE` updates build progress and thinking message
- **R65:** `audit APP` runs code quality audit via `/api/ui_audit`
- **R66:** `patterns` lists patterns with frontmatter from `{base_dir}/patterns/`
- **R67:** `theme` commands delegate to frictionless binary
- **R68:** Guard against recursion: `run` exits if `FRICTIONLESS_MCP=1` is set

### Checkpoint Commands
- **R69:** `checkpoint save APP [MSG]` saves current state (no-op if unchanged)
- **R70:** `checkpoint list APP` shows checkpoint history
- **R71:** `checkpoint rollback APP [N]` restores Nth checkpoint
- **R72:** `checkpoint diff APP [N]` shows diff from checkpoint N to current
- **R73:** `checkpoint clear APP` resets to baseline (alias for baseline)
- **R74:** `checkpoint baseline APP` sets current state as new baseline
- **R75:** `checkpoint count APP` returns number of checkpoints
- **R76:** Auto-download fossil SCM to `~/.claude/bin/fossil` if missing
- **R77:** Detect platform (Linux x86_64, macOS arm64/x86_64) for fossil download
- **R78:** Initialize per-app fossil repository in `{app_dir}/checkpoint.fossil`
- **R79:** Notify appConsole of checkpoint changes by resetting `_checkpointsTime`

### linkapp Script
- **R80:** `linkapp add APP` creates symlinks for app's lua and viewdefs
- **R81:** `linkapp remove APP` removes symlinks for app
- **R82:** `linkapp list` lists currently linked apps
- **R83:** Create `{base_dir}/lua/` and `{base_dir}/viewdefs/` directories if missing
- **R84:** Link `app.lua` to `lua/{app}.lua` using relative paths
- **R85:** Link app directory to `lua/{app}` using relative paths (enables `require("{app}.module")`)
- **R86:** Link all viewdef HTML files individually to `viewdefs/`
- **R87:** Remove command scans viewdefs/ for symlinks pointing to app and removes them
