# Test Design: Permission Prompt System

**Source Spec:** prompt-ui.md

**CRC Cards:**
- crc-PromptManager.md
- crc-PromptHTTPServer.md
- crc-Server.md
- crc-PromptResponseCallback.md
- crc-PromptViewdef.md
- crc-PermissionHook.md
- crc-HookCLI.md
- crc-ViewdefsResource.md
- crc-PermissionHistoryResource.md

**Sequences:**
- seq-prompt-flow.md
- seq-prompt-server-startup.md
- seq-hook-install.md

---

## Unit Tests

### PromptManager Tests

**File:** `internal/prompt/manager_test.go`

| Scenario | Input | Expected |
|----------|-------|----------|
| CreatePrompt generates unique ID | `CreatePrompt()` | Returns non-empty ID and channel |
| CreatePrompt stores in pending map | `CreatePrompt()` | `pendingPrompts[id]` exists |
| Respond sends to channel | `Respond(id, value, label)` | Channel receives response |
| Respond removes from map | `Respond(id, value, label)` | `pendingPrompts[id]` no longer exists |
| Respond with invalid ID | `Respond("invalid", ...)` | Returns error, no panic |
| Cancel closes channel | `Cancel(id)` | Channel closed |
| Cancel removes from map | `Cancel(id)` | `pendingPrompts[id]` no longer exists |
| Concurrent CreatePrompt | Multiple goroutines call `CreatePrompt()` | All IDs unique, no race |
| Concurrent Respond | Create then respond from different goroutine | Response received correctly |

### PromptHTTPServer Tests

**File:** `internal/prompt/server_test.go`

| Scenario | Input | Expected |
|----------|-------|----------|
| Start binds to port | `Start()` | Port > 0, server listening |
| Start writes port file | `Start()` | `.ui-mcp/mcp-port` contains port number |
| Stop removes port file | `Stop()` | `.ui-mcp/mcp-port` deleted |
| POST /api/prompt valid request | JSON with message, options | Returns 200, blocks until response |
| POST /api/prompt missing message | JSON without message | Returns 400 |
| POST /api/prompt empty options | JSON with empty options array | Returns 400 |
| POST /api/prompt timeout | Request with timeout=1, no response | Returns 408 after timeout |
| POST /api/prompt invalid JSON | Malformed JSON body | Returns 400 |
| POST /api/prompt method not allowed | GET /api/prompt | Returns 405 |

### Server.Prompt Tests

**File:** `internal/prompt/prompt_test.go`

| Scenario | Input | Expected |
|----------|-------|----------|
| Prompt sets app.pendingPrompt | `Prompt(session, msg, opts)` | Lua `app.pendingPrompt` contains data |
| Prompt switches presenter | `Prompt(session, msg, opts)` | Lua `app._presenter == "Prompt"` |
| Prompt blocks until response | Call Prompt, then Respond | Prompt returns after Respond |
| Prompt returns response data | Respond with value/label | Prompt returns matching response |
| Prompt timeout | `Prompt()` with short timeout | Returns timeout error |

### PromptResponseCallback Tests

**File:** `internal/prompt/callback_test.go`

| Scenario | Input | Expected |
|----------|-------|----------|
| Callback registered in Lua | After registration | `_G.promptResponse` callable |
| Callback signals channel | Call from Lua | PromptManager.Respond called |
| Callback with invalid ID | Call with unknown ID | Error logged, no panic |

### HookCLI Tests

**File:** `cmd/hooks_test.go`

| Scenario | Input | Expected |
|----------|-------|----------|
| Install creates script | `hooks install` | `.claude/hooks/permission-ui.sh` exists |
| Install makes script executable | `hooks install` | Script has execute permission |
| Install updates settings.json | `hooks install` | PermissionRequest hook added |
| Install preserves existing settings | Existing settings.json | Other settings preserved |
| Install creates .claude if missing | No .claude dir | Directory created |
| Uninstall removes from settings | `hooks uninstall` | PermissionRequest hook removed |
| Uninstall preserves script | `hooks uninstall` | Script still exists |
| Uninstall --delete-script | `hooks uninstall --delete-script` | Script deleted |
| Status shows installed | Hook in settings | Output shows "installed" |
| Status shows not installed | No hook | Output shows "not installed" |
| Status shows server running | mcp-port file exists | Output shows "server running" |

### MCP Resource Tests

**File:** `internal/mcp/resources_test.go`

| Scenario | Input | Expected |
|----------|-------|----------|
| ui://viewdefs returns content | Read resource | Markdown with viewdef locations |
| ui://permissions/history reads log | permissions.log exists | JSON with decisions array |
| ui://permissions/history empty log | No log file | Empty decisions array |

---

## Integration Tests

### HTTP to Lua Flow

**File:** `internal/prompt/integration_test.go`

| Scenario | Setup | Action | Expected |
|----------|-------|--------|----------|
| Full prompt flow | Server running, browser connected | POST /api/prompt, trigger Lua callback | HTTP returns response |
| Prompt to specific session | Two sessions connected | POST with sessionId=2 | Only session 2 sees prompt |
| No browser connected | Server running, no WS | POST /api/prompt | Times out (no recipient) |
| Multiple concurrent prompts | Server running, browser connected | Two POSTs simultaneously | Both resolve correctly |

### Server Startup Integration

**File:** `internal/mcp/server_test.go` (extend existing)

| Scenario | Action | Expected |
|----------|--------|----------|
| ui_start creates prompt server | Call ui_start tool | PromptHTTPServer running |
| ui_start writes both port files | Call ui_start tool | Both ui-port and mcp-port files exist |
| ui_start registers callback | Call ui_start tool | _G.promptResponse callable in Lua |
| Shutdown stops prompt server | Start then shutdown | Port file removed |

---

## End-to-End Tests

### Hook Script Tests

**File:** `tests/e2e/hook_test.sh` (shell script tests)

| Scenario | Setup | Action | Expected |
|----------|-------|--------|----------|
| Hook reads port file | Start server | Run hook with test input | Hook calls correct port |
| Hook falls back without server | No server running | Run hook | Exits with 0 (fallback) |
| Hook returns allow decision | Server + browser, click allow | Run hook | Output: `{"decision": "allow"}` |
| Hook returns deny decision | Server + browser, click deny | Run hook | Output: `{"decision": "deny", "message": "..."}` |
| Hook logs decision | Complete flow | Check permissions.log | Entry appended |
| Hook handles timeout | Server running, no click | Run hook with short timeout | Exits with 0 (fallback) |

### Browser Integration Tests (Playwright)

**File:** `tests/e2e/prompt_e2e_test.go`

| Scenario | Setup | Action | Expected |
|----------|-------|--------|----------|
| Prompt viewdef renders | POST /api/prompt | Check browser | sl-dialog visible with message |
| Options rendered as buttons | POST with 3 options | Check browser | 3 sl-button elements |
| Click button resolves prompt | Click first button | HTTP response | Returns matching value/label |
| Presenter switches back | Complete prompt flow | Check browser | Previous viewdef restored |
| Viewdef customization | Modify Prompt.DEFAULT.html | Trigger prompt | Custom styling visible |

---

## Test Data

### Sample Prompt Request
```json
{
  "message": "Claude wants to run: grep -r 'password' .",
  "options": [
    {"label": "Allow once", "value": "allow"},
    {"label": "Always allow grep", "value": "allow_session"},
    {"label": "Deny", "value": "deny"}
  ],
  "sessionId": "1",
  "timeout": 60
}
```

### Sample Prompt Response
```json
{
  "value": "allow",
  "label": "Allow once"
}
```

### Sample Lua State After Prompt
```lua
app.pendingPrompt = {
  id = "abc123",
  message = "Claude wants to run: grep -r 'password' .",
  options = {
    {label = "Allow once", value = "allow"},
    {label = "Always allow grep", value = "allow_session"},
    {label = "Deny", value = "deny"}
  }
}
app._presenter = "Prompt"
```

### Sample Hook Input
```json
{
  "tool_name": "Bash",
  "tool_input": {"command": "grep -r 'password' ."}
}
```

### Sample Hook Output (allow)
```json
{"decision": "allow"}
```

### Sample Hook Output (deny)
```json
{"decision": "deny", "message": "User denied permission"}
```

### Sample Permission Log Entry
```json
{"tool": "Bash", "input": {"command": "grep -r 'password' ."}, "choice": "allow_session", "timestamp": "2025-01-01T12:00:00Z"}
```
