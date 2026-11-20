# Test Design: Per-session setup hooks
**Source:** crc-FlibRuntime.md, crc-MCPServer.md

An embedder registers a hook with `OnSessionSetup`; flib runs it for every session it
creates, on the session's executor, with the session's `cli.LuaRig`. ark asked for this
(`ark-flib-hostop-access`) to define host operations in place of `ui.exec`, so the tests
define a real host operation through the hook and call it from Lua.

## Test: a hook defines a host operation the session can call
**Purpose:** hooks run for a created session, in registration order, on its executor, with its rig, so `DefineHostOp` works (R253, R254)
**Input:** an MCP server with two hooks registered before session creation: the first records "one", the second records "two" and defines a host operation `hookOp` returning "from-hook"; the session is created; Lua calls `ui.hookOp(t, "out", cb)`
**Expected:** the record reads "one,two", and `t.out` is "from-hook" after the callback
**Refs:** crc-MCPServer.md
**Code:** internal/mcp/session_setup_test.go
**Alarm:** 1
**Fire alarm:** drop the hook loop from setupMCPGlobal; the record stays empty and `ui.hookOp` is nil
**Inject:** internal/mcp/tools.go:Server.setupMCPGlobal
@alarm-hash: 66d23c4a920cb0ef
**Pulled:** 2026-11-20 — rang: `hooks ran as "", want "one,two" (registration order, R253)`

## Test: hooks run again when reconfiguring replaces the session
**Purpose:** every created session gets the hooks, not only the first (R253)
**Input:** a hook that counts its runs; the session is created, then replaced the way `handleConfigure` replaces it (`StartAndCreateSession` again)
**Expected:** the hook ran twice, and the new session has its effect
**Refs:** crc-MCPServer.md
**Code:** internal/mcp/session_setup_test.go

## Test: a hook registered after Start runs for the existing session
**Purpose:** registering after `Start` is as good as before it (R255)
**Input:** a started flib runtime; `OnSessionSetup` with a hook that sets a Lua global
**Expected:** `OnSessionSetup` returns nil and the global is set in the current session
**Refs:** crc-FlibRuntime.md
**Code:** flib/flib_test.go
**Alarm:** 2
**Fire alarm:** only keep the hook for later sessions, without running it at registration; the global stays unset
**Inject:** flib/flib.go:OnSessionSetup
@alarm-hash: d321d049bf22b336
**Pulled:** 2026-11-20 — rang: `lateHook = nil after OnSessionSetup, want true`

## Test: a hook's error reaches the caller
**Purpose:** an error from a hook run at registration is returned by `OnSessionSetup` (R256)
**Input:** a started flib runtime; `OnSessionSetup` with a hook that returns an error
**Expected:** `OnSessionSetup` returns that error
**Refs:** crc-FlibRuntime.md, crc-MCPServer.md
**Code:** flib/flib_test.go
**Alarm:** 3
**Fire alarm:** ignore the hook's error in RunSessionSetupHook; OnSessionSetup returns nil
**Inject:** internal/mcp/server.go:Server.RunSessionSetupHook
@alarm-hash: 722fd51a8388dda0
**Pulled:** 2026-11-20 — rang: `OnSessionSetup = <nil>, want the hook's error (R256)`
