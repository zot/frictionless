# Test Design: App Activation and Loading
**Source:** crc-MCPServer.md, crc-FlibRuntime.md, crc-LinkappScript.md

Apps are activated by a `lua/{app}` directory link alone; the session tests
lay out `apps/<name>/app.lua` plus that link, with no `lua/<name>.lua`.

## Test: BuiltinApps reaches the ui-engine config
**Purpose:** flib passes its built-in app list through (R190)
**Input:** `flib.New(Config{Dir, BuiltinApps: [ark mcp]})`
**Expected:** the runtime's ui-engine config carries `[ark mcp]`
**Refs:** crc-FlibRuntime.md
**Code:** flib/flib_test.go
**Fire alarm:** drop the passthrough line. Red: `ui-engine BuiltinApps = []`.
**Inject:** flib/flib.go:New
**Pulled:** 2026-10-07 — rang: `ui-engine BuiltinApps = [], want [ark mcp]`

## Test: the mcp app's entry extends the mcp global
**Purpose:** the mcp app loads from `lua/mcp/app.lua` at session start (R183)
**Input:** `apps/mcp/app.lua` sets `mcp.extended`; only `lua/mcp` is linked
**Expected:** `mcp.extended` is `"yes"` in the session
**Refs:** crc-MCPServer.md, seq-mcp-lifecycle.md
**Code:** internal/mcp/activation_test.go
**Fire alarm:** load `"mcp.lua"` again instead of the entry file. Red: `mcp.extended = "null"`.
**Inject:** internal/mcp/tools.go:setupMCPGlobal
**Pulled:** 2026-10-07 — rang: `mcp.extended = "null"`

## Test: mcp:app loads an app's entry through its activation link
**Purpose:** on first request an app's entry is `lua/{app}/app.lua` (R189)
**Input:** `apps/demo/app.lua` defines global `demo`; only `lua/demo` is linked
**Expected:** `mcp:app('demo').loaded` is `"yes"`
**Refs:** crc-MCPServer.md
**Code:** internal/mcp/activation_test.go
**Fire alarm:** make appEntryFile return the old `<app>.lua`. Red: `mcp:app('demo')` is nil (indexing it fails).
**Inject:** internal/mcp/tools.go:appEntryFile
**Pulled:** 2026-10-07 — rang: `attempt to index a non-table object(nil) with key 'loaded'`

## Test: linkapp makes exactly one link and cleans up the old scheme
**Purpose:** add links `lua/{app}` and removes `lua/{app}.lua` and viewdef links into the app; list reports directory links; remove deletes the link (R184-R188)
**Input:** a temp base with the script, `apps/demo`, and old-scheme links
**Expected:** `lua/demo -> ../apps/demo`; old links gone; `list` prints `demo`; `remove` deletes `lua/demo`
**Refs:** crc-LinkappScript.md
**Code:** internal/mcp/activation_test.go
**Fire alarm:** (a) skip remove_old_links in add_app, red: old links survive; (b) list `*.lua` links as before, red: list reports nothing.
**Inject:** install/linkapp:add_app, install/linkapp:list_apps
**Pulled:** 2026-10-07 — rang: (a) `old-scheme link lua/demo.lua survived add`, (b) `list output "Active apps:\n" does not report demo`
