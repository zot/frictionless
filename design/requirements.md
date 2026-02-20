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
- **R12:** Register mcp global in Lua with pushState, pollingEvents, app, display, status, waitTime, sessionId methods
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
- **R128:** Provide `mcp.sessionId` field containing the current external session ID (internal UUID)
- **R129:** Install bundled pattern files to `{base_dir}/patterns/` during `ui_install`

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
- **R125:** `checkpoint update APP [MSG]` saves current file state on a separate "updates" branch
- **R126:** `checkpoint local APP [MSG]` saves current file state on a separate "local" branch
- **R127:** The "updates" and "local" branches survive `checkpoint baseline` resets via fossil bundle export/import

### linkapp Script
- **R80:** `linkapp add APP` creates symlinks for app's lua and viewdefs
- **R81:** `linkapp remove APP` removes symlinks for app
- **R82:** `linkapp list` lists currently linked apps
- **R83:** Create `{base_dir}/lua/` and `{base_dir}/viewdefs/` directories if missing
- **R84:** Link `app.lua` to `lua/{app}.lua` using relative paths
- **R85:** Link app directory to `lua/{app}` using relative paths (enables `require("{app}.module")`)
- **R86:** Link all viewdef HTML files individually to `viewdefs/`
- **R87:** Remove command scans viewdefs/ for symlinks pointing to app and removes them

## Feature: Publisher Server
**Source:** specs/publisher.md

- **R88:** The publisher is a separate background process that binds to `localhost:25283`
- **R89:** `POST /publish/{topic}` accepts a JSON body and delivers it to all current subscribers of that topic, returning `{"listeners": N}`
- **R90:** `GET /subscribe/{topic}` long-polls until data is published (returns 200 with the JSON body) or times out after ~60s (returns 204 No Content)
- **R91:** `GET /` serves an install page with the bookmarklet link, instructions, and current topic/listener counts
- **R92:** All endpoints set CORS headers to allow any origin
- **R93:** Topics are implicit — created on first subscribe or publish, no registration required
- **R94:** Published messages wait up to 20ms for reconnecting subscribers before being dropped (fan-out with brief TTL)
- **R95:** After receiving data, a subscriber's long-poll returns; the client must reconnect to receive the next message

## Feature: Publisher Lifecycle
**Source:** specs/publisher.md

- **R96:** Each MCP server attempts to host the publisher on port 25283 at startup; first to bind wins
- **R97:** If the port is already bound, the attempt is silently ignored (another MCP server is hosting it)
- **R98:** The publisher lifecycle is tied to the MCP server that hosts it — no separate process or idle watchdog
- **R99:** (removed — idle watchdog no longer needed; publisher lives with MCP server)
- **R100:** (removed — no standalone `frictionless publisher` CLI subcommand)

## Feature: MCP Subscribe Integration
**Source:** specs/publisher.md

- **R101:** Lua apps can subscribe to topics via `mcp:subscribe(topic, handler)` where handler receives the parsed JSON data
- **R102:** `mcp:subscribe` runs a background goroutine that long-polls the publisher and calls the handler on each received message
- **R103:** If the long-poll connection fails, `mcp:subscribe` retries after a short delay (publisher is co-hosted)
- **R104:** After calling the handler, the goroutine immediately reconnects to continue listening
- **R105:** (inferred) The handler function executes in the Lua VM of the session that called `mcp:subscribe`

## Feature: Bookmarklet
**Source:** specs/publisher.md

- **R106:** The bookmarklet publishes to `/publish/scrape` with JSON containing `url` (location.href), `title` (document.title), and `text` (document.body.innerText, truncated to 50k chars)
- **R107:** On success, the bookmarklet updates the tab title to show `[Sent to N session(s)]`
- **R108:** On failure (publisher not running), the bookmarklet shows an alert
- **R109:** The install page (`GET /`) provides the bookmarklet as a draggable link for one-time setup
- **R110:** (inferred) Multiple MCP sessions can subscribe to the same topic and all receive a copy of published data

### Topic Favicons
- **R111:** The subscribe endpoint (`GET /subscribe/{topic}`) accepts an optional `favicon` query parameter containing a data URL
- **R112:** The publisher stores the favicon per topic; the most recent favicon supplied for a topic wins
- **R113:** The install page displays per-topic bookmarklet sections with the topic's favicon when available
- **R114:** `mcp:subscribe(topic, handler, opts)` accepts an optional third argument table with a `favicon` field (a data URL string)
- **R115:** The subscribe goroutine passes the favicon as a query parameter on its first long-poll request to the publisher

## Feature: CSP-Safe Relay
**Source:** specs/publisher.md

- **R116:** `GET /relay/{topic}` serves a self-contained HTML relay page that POSTs to `/publish/{topic}` same-origin
- **R117:** The relay page signals `window.opener` with `postMessage('ready', '*')` when loaded
- **R118:** The relay page listens for an incoming `message` event containing page data and POSTs it to `/publish/{topic}`
- **R119:** The relay page shows the result ("Sent to N session(s)") and auto-closes after 1.5 seconds
- **R120:** The relay page times out after 10 seconds if no data is received
- **R121:** The bookmarklet uses `window.open` to open the relay page instead of `fetch` to avoid CSP restrictions
- **R122:** The bookmarklet checks if `window.open` returned null (popup blocked) and alerts the user
- **R123:** The bookmarklet waits for a 'ready' signal from the relay page before sending data via `postMessage`
- **R124:** The bookmarklet uses origin checking on message events for security
