# Test Design: App Activation and Loading
**Source:** crc-MCPServer.md, crc-FlibRuntime.md

Apps are activated by a `lua/{app}` directory link alone; the session tests
lay out `apps/<name>/app.lua` plus that link, with no `lua/<name>.lua`.

## Test: BuiltinApps reaches the ui-engine config
**Purpose:** flib passes its built-in app list through (R190)
**Input:** `flib.New(Config{Dir, BuiltinApps: [ark mcp]})`
**Expected:** the runtime's ui-engine config carries `[ark mcp]`
**Refs:** crc-FlibRuntime.md
**Code:** flib/flib_test.go
**Alarm:** 1
**Fire alarm:** drop the passthrough line. Red: `ui-engine BuiltinApps = []`.
**Inject:** flib/flib.go:New
**Pulled:** 2026-10-07 — rang: `ui-engine BuiltinApps = [], want [ark mcp]`

## Test: ViewdefCollisions passes the ui-engine's live list through
**Purpose:** an embedder can read the running server's collisions (R191)
**Input:** `flib.New(Config{Dir})` over a site whose apps `aa` and `zz` both ship `T.DEFAULT.html`
**Expected:** one collision, `T.DEFAULT`, `aa` over `zz`
**Refs:** crc-FlibRuntime.md
**Code:** flib/flib_test.go

## Test: WithLuaUpdate reports a missing session the way WithLua does
**Purpose:** before Start there is no session; the active path fails with WithLua's error and never runs the closure (R200)
**Input:** `flib.New(Config{Dir})`, no Start; `WithLuaUpdate` with a closure that records being called
**Expected:** error `no active session`; closure not called
**Refs:** crc-FlibRuntime.md
**Code:** flib/flib_test.go
**Alarm:** 2
**Fire alarm:** drop the empty-vendedID check. Red: the error becomes `session  not found` instead of `no active session`.
**Inject:** flib/flib.go:WithLuaUpdate
@alarm-hash: 089112a39e4f1891
**Pulled:** 2026-11-16 — rang: `WithLuaUpdate() error = session  not found, want "no active session"`

## Test: WithLuaUpdate pushes the closure's changes to a connected browser
**Purpose:** the active path runs afterBatch, so a change made in the closure reaches the browser with no further browser message (R200)
**Input:** a started runtime; a websocket client on `/ws/{session}` that creates variable 2 watching `mcp.probe` (path `probe` under variable 1); then `WithLuaUpdate` sets `mcp.probe`
**Expected:** within 2s the client receives an update carrying the new value
**Refs:** crc-FlibRuntime.md
**Code:** flib/flib_test.go
**Alarm:** 3
**Fire alarm:** run the closure through the Lua session's own ExecuteInSession (WithLua's passive path) instead of the ui-engine server's. Red: `browser received "" after WithLuaUpdate`.
**Inject:** flib/flib.go:WithLuaUpdate
@alarm-hash: 089112a39e4f1891
**Pulled:** 2026-11-16 — rang: `browser received "" after WithLuaUpdate, want an update carrying "pushed-by-update"`

## Test: the mcp app's entry extends the mcp global
**Purpose:** the mcp app loads from `lua/mcp/app.lua` at session start (R183)
**Input:** `apps/mcp/app.lua` sets `mcp.extended`; only `lua/mcp` is linked
**Expected:** `mcp.extended` is `"yes"` in the session
**Refs:** crc-MCPServer.md, seq-mcp-lifecycle.md
**Code:** internal/mcp/activation_test.go
**Alarm:** 4
**Fire alarm:** load `"mcp.lua"` again instead of the entry file. Red: `mcp.extended = "null"`.
**Inject:** internal/mcp/tools.go:setupMCPGlobal
@alarm-hash: 635bba224462911a
**Pulled:** 2026-11-19 — re-pulled after setupMCPGlobal gained defineAppLinkOps: the old mcp.lua load restored; `mcp.extended = "null"` *Earlier —* 2026-11-19 — rang: setupMCPGlobal loaded "mcp.lua" instead of the entry file; `mcp.extended = "null"` (re-pulled after the checkpoint host-op hook landed in setupMCPGlobal) *Earlier —* 2026-10-07 — rang: `mcp.extended = "null"`
## Test: mcp:app loads an app's entry through its activation link
**Purpose:** on first request an app's entry is `lua/{app}/app.lua` (R189)
**Input:** `apps/demo/app.lua` defines global `demo`; only `lua/demo` is linked
**Expected:** `mcp:app('demo').loaded` is `"yes"`
**Refs:** crc-MCPServer.md
**Code:** internal/mcp/activation_test.go
**Alarm:** 5
**Fire alarm:** make appEntryFile return the old `<app>.lua`. Red: `mcp:app('demo')` is nil (indexing it fails).
**Inject:** internal/mcp/tools.go:appEntryFile
**Pulled:** 2026-10-07 — rang: `attempt to index a non-table object(nil) with key 'loaded'`
