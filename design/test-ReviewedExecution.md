# Test Design: Reviewed synchronous execution
**Source:** crc-MCPServer.md, crc-FlibRuntime.md

ui-engine treats a synchronous wait on a session's executor as dangerous, because a wait
from inside an executor deadlocks; with `AllowDangerousExecution` unset, a dangerous call
panics naming its site. `SafeExecuteInSession` and flib's `WithLuaUpdate` wait, but every
caller of both was reviewed as running off the executor (Item 12b of the coherency carve),
so both use the primitive with `dangerous` false. These tests run them with dangerous
execution disallowed, the default ui-engine is moving to.

## Test: SafeExecuteInSession works with dangerous execution disallowed
**Purpose:** the reviewed wrapper is not a dangerous call, so `ui_run` keeps working after the flip (R252)
**Input:** an MCP server session whose ui-engine config has `AllowDangerousExecution` unset; `ui_run` with `return 6 * 7`
**Expected:** the result is 42 and no error
**Refs:** crc-MCPServer.md
**Code:** internal/mcp/tools_test.go
**Alarm:** 1
**Fire alarm:** implement SafeExecuteInSession with `UiServer.ExecuteInSession` again; the dangerous call panics, SafeExecuteInSession recovers it, and ui_run returns "panic during execution: dangerous synchronous ExecuteInSession from ..."
**Inject:** internal/mcp/server.go:Server.SafeExecuteInSession
@alarm-hash: a0e8249398057217
**Pulled:** 2026-11-19 — rang: `ui_run = execution failed: panic during execution: dangerous synchronous ExecuteInSession from .../server.go:101`

## Test: WithLuaUpdate works with dangerous execution disallowed
**Purpose:** the reviewed active path is not a dangerous call, so embedders keep working after the flip (R250)
**Input:** a started flib runtime whose config has `AllowDangerousExecution` unset; `WithLuaUpdate` with a closure that sets a Lua global
**Expected:** no error and no panic; the global is set
**Refs:** crc-FlibRuntime.md
**Code:** flib/flib_test.go
**Alarm:** 2
**Fire alarm:** implement WithLuaUpdate with `uiServer.ExecuteInSession` again; the dangerous call panics naming its site, which the test recovers and reports
**Inject:** flib/flib.go:WithLuaUpdate
@alarm-hash: a1d8f4f6fc661ae0
**Pulled:** 2026-11-19 — rang: `WithLuaUpdate made a dangerous call: dangerous synchronous ExecuteInSession from .../flib.go:183`

## Not tested, deliberately
**R251, the caller contract.** Calling `WithLua` from code already on the executor
deadlocks rather than fails, so a test would hang. The review in crc-MCPServer.md and the
spec are its guard.
