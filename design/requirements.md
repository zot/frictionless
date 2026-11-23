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
- **~~R16:~~** (Retired T1 — see R183) Load mcp.lua extension file if present
- **R17:** Load app init.lua files on startup
- **R18:** Bind root URL to current session (session cookie)
- **R19:** Hot-reload Lua and viewdef files on change
- **R20:** Redirect Lua print/io to log files
- **R21:** Provide HTTP Tool API for spawned agents (GET/POST /api/*)
- **R22:** Provide GET /app/{app}/readme endpoint returning app's README.md as HTML (case-insensitive lookup, rendered via goldmark)
- **R128:** Provide `mcp.sessionId` field containing the current external session ID (internal UUID)
- **R129:** Install bundled pattern files to `{base_dir}/patterns/` during `ui_install`
- **R144:** MCP HTTP server serves static files from `{base_dir}/html/` as a catch-all for unmatched paths
- **R145:** `.md` files served via MCP HTTP port are rendered as HTML (via goldmark) for browser User-Agents, raw for curl
- **R252:** `SafeExecuteInSession` runs its work through ui-engine's `PrimitiveExecuteInSession` with `dangerous` false, keeping its panic recovery; every caller runs off the session's executor (HTTP and MCP tool handlers, startup and configuration, the subscription poll goroutine), so it never warns or panics when ui-engine's dangerous-execution default flips

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
- **R136:** Theme switching dynamically adds CSS `<link>` tag if the theme's stylesheet is not yet loaded
- **R137:** Provide `mcp:reinjectThemes()` Lua method that re-scans themes and updates index.html theme block
- **R138:** `theme classes` without a theme argument returns classes from all themes, deduplicated
- **R139:** `theme audit APP` without a theme argument audits against the all-themes class list
- **R140:** Theme CSS files define structural semantic classes (`sidebar-panel`, `content-panel`, `content-card`, `dock-panel`) for layout hooks
- **R141:** Stock app viewdefs use structural semantic classes on sidebar, content, card, and dock elements
- **R142:** Theme CSS `<link>` elements in index.html are cache-busted with `?v={modtime}` from file modification timestamps
- **R143:** App CSS in viewdefs is loaded via `<script>` that creates `<link>` elements with a `Date.now()` nonce for cache busting

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
- **R76:** Auto-download fossil SCM to `~/.claude/bin/fossil` if missing
- **R77:** Detect platform (Linux x86_64, macOS arm64/x86_64) for fossil download
- **R201:** `mcp checkpoint CMD APP [ARG]` ensures fossil is installed, then hands off to `frictionless checkpoint CMD APP [ARG]`, adding nothing to its output or exit status

### linkapp Script
- **~~R80:~~** (Retired T2 — see R184) `linkapp add APP` creates symlinks for app's lua and viewdefs
- **~~R81:~~** (Retired T3 — see R185) `linkapp remove APP` removes symlinks for app
- **~~R82:~~** (Retired T4 — see R186) `linkapp list` lists currently linked apps
- **~~R83:~~** (Retired T5 — see R187) Create `{base_dir}/lua/` and `{base_dir}/viewdefs/` directories if missing
- **~~R84:~~** (Retired T6 — no replacement) Link `app.lua` to `lua/{app}.lua` using relative paths
- **~~R85:~~** (Retired T9 — see R226) Link app directory to `lua/{app}` using relative paths (enables `require("{app}.module")`)
- **~~R86:~~** (Retired T7 — no replacement) Link all viewdef HTML files individually to `viewdefs/`
- **~~R87:~~** (Retired T8 — see R188) Remove command scans viewdefs/ for symlinks pointing to app and removes them

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

## Feature: Debug Pages
**Source:** specs/debug-pages.md

- **R130:** MCP port `/variables` and `/state` serve static HTML files from `{base_dir}/html/` instead of embedding HTML in Go code
- **R131:** Static debug pages are bundled in `install/html/` and installed to `.ui/html/` on install/upgrade
- **R132:** Variable browser click-to-copy: clicking any cell except Path copies the variable's full JSON to clipboard
- **R133:** Click-to-copy shows a brief toast notification confirming the copy
- **R134:** MCP port `/state` returns raw JSON (not embedded HTML) for script consumption
- **R135:** MCP port `/variables` redirects to the UI port's variable browser

## Feature: Rich Messages
**Source:** specs/rich-messages.md

- **R146:** Provide `renderMarkdownFragment(content)` Go function that converts markdown to an HTML fragment via goldmark (no page wrapper)
- **R147:** Expose markdown rendering to Lua as `mcp:renderMarkdown(text)` returning an HTML string
- **R148:** ChatMessage prototype has an `html` field for rendered/rich HTML content
- **R149:** ChatMessage viewdef displays `html` field (via `ui-html`) when present, plain `text` (via `ui-value`) otherwise
- **R150:** `mcp:addRichMessage(html)` creates an agent message with pre-built HTML content
- **R151:** `mcp:highlightLink(elementId, label)` returns an anchor tag HTML string that calls `window.uiApp.highlight()` on click
- **R152:** All chat messages (user and agent) render `text` as markdown into `html` at creation time via `mcp:renderMarkdown()`
- **R153:** Raw `text` field is preserved as source alongside rendered `html`
- **R154:** Goldmark renders with default settings (no `html.WithUnsafe()`) so raw HTML in markdown input is escaped

## Feature: Reconfigure Waiter Notification
**Source:** specs/mcp.md

- **R155:** When `Stop()` is called during reconfiguration, it pushes a `server_reconfigured` event to the state queue for the current session before destroying it, so clients blocked on `/wait` unblock immediately

## Feature: Project-Agnostic Skills
**Source:** specs/mcp.md

- **R156:** Skill files use `{cmd}` as a placeholder for the Frictionless command, making skills project-agnostic
- **R157:** `ui_install` patches the project's CLAUDE.md to declare what `{cmd}` resolves to (e.g., `.ui/mcp`)
- **R158:** The `{cmd}` declaration is inserted at the top of CLAUDE.md (after any front matter)
- **R159:** If the declaration section already exists in CLAUDE.md, it is updated rather than duplicated
- **R160:** If CLAUDE.md does not exist, it is created with just the declaration section

## Feature: Library Embedding
**Source:** specs/mcp.md

- **R161:** The `flib` package provides a public API for embedding Frictionless as a Go library; internal packages remain unexposed
- **R162:** `flib.New(Config)` creates a runtime with ui-engine server, MCP server, and cleanup worker without starting anything
- **R163:** `Runtime.Configure()` prepares the base directory and runs auto-install if needed
- **R164:** `Runtime.Start()` starts the UI HTTP server and creates a Lua session with the `mcp` global; returns the base URL
- **R165:** `Runtime.RegisterAPI(mux)` mounts Frictionless API handlers (`/api/*`, `/wait`, `/state`, `/variables`) on an external `http.ServeMux`
- **R166:** `Runtime.StartAPI()` starts a standalone HTTP API server and returns its port; alternative to `RegisterAPI` when no external listener exists
- **R167:** `Runtime.Shutdown(ctx)` gracefully shuts down both UI and API servers
- **R168:** `Server.RegisterAPIRoutes(mux)` extracts route registration from `StartHTTPServer` so both standalone and embedded modes share the same handler set
- **R177:** `flib.Config.Project` sets an explicit project directory for skill installation; when empty, `{project}` is derived from `{base_dir}` as before
- **R178:** `Server.ProjectDir` is the internal field that `Install()` uses; `flib` sets it from `Config.Project`
- **~~R179:~~** (Retired T10 — see R249) `Runtime.WithLua(fn)` executes a closure in the Lua executor goroutine (thread-safe, passive path) without triggering afterBatch; flib resolves the vendedID so callers never need session IDs
- **R180:** `flib.Config.Port` sets a preferred HTTP port for the ui-engine server (0 = auto-select, default); enables downstream binaries to persist the port across restarts so browsers reconnect automatically
- **R181:** `Runtime.UIHandleFunc(pattern, handler)` registers a custom HTTP handler on the UI server's mux by delegating to `ui-engine HTTPEndpoint.HandleFunc`; enables downstream binaries to add routes to the UI port
- **R182:** When `Config.Dir` is non-empty, `New` derives the ui-engine backend control socket as `{Dir}/ui.sock` instead of the shared `/tmp/ui.sock` default, so each embedded runtime owns a private control socket and multiple embedded runtimes no longer collide on the fixed default
- **~~R200:~~** (Retired T11 — see R250) `Runtime.WithLuaUpdate(fn)` is the active twin of `WithLua`: it executes the closure in the Lua executor goroutine through the ui-engine server's `ExecuteInSession`, which runs afterBatch when a browser is connected, so the closure's Lua state changes reach the browser at once; flib resolves the vendedID and reports the same errors as `WithLua`
- **R249:** `Runtime.WithLua(fn)` executes a closure on the session's executor (thread-safe: ui-engine runs a session's Lua only there; passive path) without triggering afterBatch; flib resolves the vendedID so callers never need session IDs
- **R250:** `Runtime.WithLuaUpdate(fn)` is the active twin of `WithLua`: it executes the closure on the session's executor through ui-engine's `PrimitiveExecuteInSession` with `dangerous` false, which runs afterBatch when a browser is connected, so the closure's Lua state changes reach the browser at once; flib resolves the vendedID and reports the same errors as `WithLua`
- **R251:** A caller of `WithLua` or `WithLuaUpdate` does its I/O before the call and never calls either from code already running on the session's executor, which would wait for itself
- **R253:** `Runtime.OnSessionSetup(fn)` registers a per-session setup hook; flib runs every registered hook, in registration order, for each session it creates, at startup and again when reconfiguring replaces the session
- **R254:** A hook runs on the session's executor at the end of flib's own session setup, after the `mcp` global and Frictionless's host operations exist, and receives that session's `cli.LuaRig`, so it can call `DefineHostOp`
- **R255:** A hook registered after `Start` also runs once at once, for the session that already exists
- **R256:** An error from a hook fails that session's setup as an error in flib's own setup does; for the run at registration, `OnSessionSetup` returns the error

## Feature: Install Manifest
**Source:** specs/mcp.md

- **R169:** `ui_install` records a SHA-256 hash for each installed file in an install manifest stored in `{base_dir}/storage/settings.json` under the `installManifest` key
- **R170:** Each manifest entry has a collision policy (`skip`, `overwrite`, or `backup`) and a group identifier
- **R171:** On upgrade, `ui_install` compares the on-disk file hash to the manifest hash to detect user modifications
- **R172:** If a file is user-modified, `ui_install` applies the file's collision policy: `skip` leaves it, `overwrite` replaces it, `backup` saves it as `{file}.bak` then replaces
- **R173:** Default collision policies are assigned by group (e.g. skills→skip, engine→overwrite)
- **R174:** A missing manifest (pre-manifest install or first install) is treated as empty — all existing files are considered unmodified for backward compatibility
- **R175:** `force=true` overrides all collision policies and overwrites every file, updating manifest hashes
- **R176:** The install result includes `user_modified` and `backed_up` arrays alongside existing `installed` and `skipped` arrays

### Multi-File Theme Injection
- **R195:** `InjectThemeBlockInFile(baseDir, filePath)` patches any HTML file with the theme block, replacing content between markers if present
- **R196:** `InjectAllThemeBlocks(baseDir)` scans `html/*.html` for files containing frictionless markers and patches them all
- **R197:** Theme CSS watcher calls `InjectAllThemeBlocks` instead of `InjectThemeBlock` to update all HTML files on CSS changes
- **R198:** `flib.ThemeBlock(baseDir)` returns the generated theme HTML block string for in-memory injection by embedders
- **R199:** `flib.InjectAllThemeBlocks(baseDir)` exposes disk-patching for embedders to call at startup

## Feature: App Activation and Loading
**Source:** specs/mcp.md, specs/helper-scripts.md, specs/applink.md

- **R183:** After creating the `mcp` global, Go loads the mcp app's entry `{base_dir}/lua/mcp/app.lua` if it exists
- **R184:** `linkapp add APP` activates an app by linking its directory to `lua/{app}`
- **R185:** `linkapp remove APP` deactivates an app by removing its `lua/{app}` link
- **R186:** `linkapp list` reports the directory links in `lua/` as the active apps
- **R187:** `linkapp add` creates `{base_dir}/lua/` if missing
- **R188:** `linkapp add` and `remove` delete the app's old-scheme links if present: `lua/{app}.lua`, and any `viewdefs/` symlink pointing into the app's `viewdefs/`
- **R189:** `mcp:app(name)` and `mcp:display(name)` load an app's entry file `{base_dir}/lua/{app}/app.lua` the first time the app is asked for; apps load on request, never at session start
- **R190:** `flib.Config.BuiltinApps` is passed to the ui-engine as its built-in app list; empty means no app is built-in
- **R191:** `flib.Runtime.ViewdefCollisions()` returns the running ui-engine's live viewdef-name collisions, unchanged
- **R226:** On Unix an activation link is a symbolic link with the relative target `../apps/{app}`, so a moved `{base_dir}` keeps every link resolving
- **R227:** On Windows an activation link is a directory junction, made with no administrator privilege, no Developer Mode, and no other program run
- **R228:** At server start, an activation link that no longer resolves and whose app directory exists under `apps/` is made again; one whose app is gone is reported and left alone; repair never creates a link that was not there
- **R229:** Telling an activation link from a real directory treats both a symbolic link and a junction as a link
- **R230:** Activating an app whose directory does not exist under `apps/` is an error naming the path; activating an app already linked to its own directory does nothing; a link there pointing elsewhere is replaced
- **R231:** Deactivating refuses a real directory at `lua/{app}`: it is an error and nothing is removed; deactivating an app that is not active does nothing
- **R232:** Deleting an app deactivates it, then removes `apps/{app}`; an empty name or one containing a path separator or `..` is an error and nothing is touched; anything at `apps/{app}` but a real directory is an error and nothing is removed
- **R233:** `frictionless linkapp add|remove|list [APP]` activates, deactivates or lists apps; `mcp linkapp` hands off to it
- **R234:** `ui.linkApp(name)`, `ui.unlinkApp(name)` and `ui.deleteApp(name)` run synchronously, return nothing on success, and raise an error naming what went wrong otherwise
- **R235:** No separate `linkapp` script is installed

## Feature: Checkpoints
**Source:** specs/checkpoint.md

- **R69:** `checkpoint save APP [MSG]` saves current state (no-op if unchanged)
- **R70:** `checkpoint list APP` shows checkpoint history
- **R71:** `checkpoint rollback APP [N]` restores Nth checkpoint
- **R72:** `checkpoint diff APP [N]` shows diff from checkpoint N to current
- **R73:** `checkpoint clear APP` resets to baseline (alias for baseline)
- **R74:** `checkpoint baseline APP` sets current state as new baseline
- **R75:** `checkpoint count APP` returns number of checkpoints
- **R78:** Initialize per-app fossil repository in `{app_dir}/checkpoint.fossil`
- **R79:** Notify appConsole of checkpoint changes by resetting `_checkpointsTime`
- **R125:** `checkpoint update APP [MSG]` saves current file state on a separate "updates" branch
- **R126:** `checkpoint local APP [MSG]` saves current file state on a separate "local" branch
- **R127:** The "updates" and "local" branches survive `checkpoint baseline` resets via fossil bundle export/import
- **R192:** `checkpoint baseline` sets the old repository aside rather than deleting it, and discards it only after the new baseline is committed and every preserved branch is imported and present
- **R193:** If any baseline step fails, the old repository and checkout are restored byte-identical, the failure is reported on stderr, and the command exits non-zero
- **R194:** Preserved branches are detected by exact branch name, not substring
- **R202:** The `checkpoint` package (`github.com/zot/frictionless/checkpoint`) is importable outside Frictionless, so ark can use it
- **R203:** The frictionless binary takes `frictionless checkpoint CMD APP [ARG]`, with APP a directory under `{base_dir}/apps/`
- **R204:** A failing checkpoint command prints the reason on stderr and exits non-zero; success and failure are told apart by exit status alone
- **R205:** The app-console notify after a changing command (save, rollback, clear, baseline, update, local) is best effort and never changes the exit status
- **R206:** `Find` returns `~/.claude/bin/fossil` when present, then fossil on PATH (`fossil.exe` on Windows), and `ErrNoFossil` otherwise
- **~~R207:~~** (Retired T12 — see R257) When fossil is missing, the error says how to install it, with the download for each platform
- **R257:** When fossil is missing, the error's install steps are for the running platform only (`runtime.GOOS`, and `GOARCH` where the download differs): Linux x86_64, macOS arm64 and macOS x86_64 get shell commands that fetch that platform's archive into `~/.claude/bin` and make it executable
- **R258:** On Windows the install steps are PowerShell commands that fetch the 64-bit zip and expand it into `%USERPROFILE%\.claude\bin`
- **R259:** On a platform with no prebuilt fossil (Linux on arm64, among others) the install steps name the download page and where to put the binary
- **R208:** The `App` methods return values and print nothing; `Command` does the printing
- **R209:** `Command(args, appsDir, out, onChange)` runs one checkpoint command and calls `onChange`, when non-nil, after a command that changes checkpoints
- **R210:** Every fossil step is checked; a failure stops the command with an error naming the fossil command and carrying fossil's output
- **R211:** Before acting, the package verifies that fossil's repository and checkout root for the app directory are the app's own, and stops otherwise
- **R212:** A checkout is opened with `--nested` and `--keep`; a checkout bound to the wrong repository is reported, never replaced
- **R213:** Switching to the "updates" or "local" branch and back to trunk never changes the working tree
- **R214:** Fossil's `mtime-changes` setting is off on every opened checkout
- **R215:** `Changed` reports whether the app's files differ from the tip of trunk in `{app_dir}/original.fossil`: a file changed, added or removed
- **R216:** `Changed` ignores fossil's own files (`checkpoint.fossil`, `original.fossil`, the checkout database) and the files the checkpoint repository ignores
- **R217:** `Changed` is false for an app without `original.fossil`
- **R218:** `App.Status` reports an app's trunk checkpoint count, whether it has `original.fossil`, and whether it has changed since install
- **R219:** `Status` names each repository it reads (`-R`), and never opens, verifies or configures a checkout or writes anything
- **R220:** `Status` reports a count of 0 for an app without a repository
- **R221:** `App.Install` sets a baseline, then copies `checkpoint.fossil` to `original.fossil`; a failed baseline copies nothing, and a failed copy leaves no partial `original.fossil`
- **R222:** The checkpoint ignore-glob includes `original.fossil`
- **R223:** The host defines `ui.checkpointStatus` and `ui.checkpointInstall` as host operations where it builds each session's `mcp` global, so flib embedders get them too
- **R224:** `ui.checkpointStatus(target, field, names)` delivers one entry per name, holding `count`, `downloaded` and `changed`, or `error` when that app's status could not be read; one unreadable app does not hide the others
- **R225:** `ui.checkpointInstall(target, field, name, cb)` runs `Install` for the named app and delivers true

## Feature: Installing Apps from GitHub
**Source:** specs/github-install.md

- **R236:** A GitHub tree URL `https://github.com/{user}/{repo}/tree/{branch}/{path}` names one directory in one branch; the app's name is the last segment of `{path}`; a URL of any other shape is an error
- **R237:** `ui.githubInvestigate(target, field, url, cb)` fetches the listing of the app's directory and of its `viewdefs/` directory, each a list of `{name, type}` entries with type `file` or `dir`
- **R238:** `ui.githubInvestigate` fetches the text of `requirements.md`, `design.md`, every `.lua` file in the app's directory and every `.html` file in its `viewdefs/`, keyed by path relative to the app directory
- **R239:** `ui.githubInvestigate` writes `target[field] = {files, viewdefs, contents}`; a file that cannot be fetched is absent from `contents` without being an error; an unparseable URL or an unfetchable directory listing is an error; its deadline is 60 seconds
- **R240:** `ui.githubInstall(target, field, url, name, cb)` is an error, changing nothing, when `apps/{name}` already exists
- **R241:** `ui.githubInstall` extracts only the archive entries under the app's directory (the archive's one top-level directory, then `{path}`), with that prefix removed
- **R242:** `ui.githubInstall` fails the install when an entry's path would land outside the app's directory
- **R243:** `ui.githubInstall` refuses an archive larger than 100 MB, or one whose files total more than 100 MB
- **R244:** `ui.githubInstall` extracts into a dot-prefixed temporary directory under `apps/` and moves it into place only when everything is extracted; any failure removes the temporary directory, so `apps/{name}` is either the complete app or absent
- **R245:** `ui.githubInstall` writes `target[field] = true` on success and any failure to the error field; its deadline is 120 seconds
- **R246:** `ui.githubInvestigate` and `ui.githubInstall` return at once and never run another program
- **R247:** `ui.githubInvestigate` and `ui.githubInstall` use GitHub's public endpoints without credentials
- **R248:** `ui.githubInstall` refuses a name that is empty or contains a path separator or `..`, changing nothing, so the app lands directly under `apps/`

