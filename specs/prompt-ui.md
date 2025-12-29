# UI Prompt for Claude Code Permissions

## Goal
Replace Claude Code's terminal-based permission prompts with a browser UI.

## Background

Claude Code presents permission prompts in the terminal:
```
Bash(grep -r 'password' .)
1. Allow once
2. Always allow grep for this session
3. Deny
```

We want to show these as clickable buttons in a browser instead.

## Constraints

1. **No infinite permission loop**: The prompt mechanism cannot use MCP tools that themselves require permission. Solution: Use direct HTTP endpoint, not MCP tool.

2. **Port separation**: UI connections (browser WebSocket) and MCP connections (agent communication) use separate ports to avoid mixing concerns.

3. **Stdio MCP compatibility**: Must work when Claude connects via stdio MCP (the default), not just SSE mode.

## Architecture

### Two-Port Design

```
Port A (UI Server)              Port B (MCP Server)
├── GET /           (HTML/JS)   ├── SSE /sse (MCP transport)
├── WS /ws/:id      (browser)   └── POST /api/prompt (hook calls)
└── Static assets
```

### Request Flow

```
Claude Code needs permission
        ↓
PermissionRequest hook fires
        ↓
Hook script reads MCP port from .ui-mcp/mcp-port
        ↓
curl POST http://127.0.0.1:{mcp-port}/api/prompt
        ↓
MCP server signals UI server internally
        ↓
UI server sends "prompt" via WebSocket to browser
        ↓
Browser shows modal with buttons
        ↓
User clicks option
        ↓
Browser sends "promptResponse" via WebSocket
        ↓
UI server signals MCP server (unblocks /api/prompt)
        ↓
HTTP response returns to curl
        ↓
Hook script returns allow/deny decision to Claude
```

### Stdio Mode with Background HTTP

When running in stdio mode (default), ui-mcp:
1. Communicates with Claude via stdio for MCP protocol
2. Starts UI HTTP server on random port (already does this)
3. Starts MCP HTTP server on separate random port (new)
4. Writes ports to `.ui-mcp/ui-port` and `.ui-mcp/mcp-port`

Hook script reads `.ui-mcp/mcp-port` to know where to POST.

## HTTP API: POST /api/prompt

**Request:**
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

**Response:**
```json
{
  "value": "allow",
  "label": "Allow once"
}
```

**Behavior:**
- Blocks until user clicks an option or timeout
- On timeout, returns error
- `sessionId` specifies which browser session shows the prompt (default "1")

## Viewdef-Based Prompt UI (Dogfooding ui-engine)

Instead of custom WebSocket messages and hardcoded modal, use ui-engine's viewdef system.

### User-Customizable UI

The prompt viewdef lives at `.ui-mcp/viewdefs/Prompt.DEFAULT.html` - a location Claude Code can read and edit. This enables conversational refinement:

- "Make the permission buttons larger"
- "Add a dark theme to the prompt dialog"
- "Show the full command, not just the tool name"
- "Add a 'Remember for this directory' option"

Claude modifies the viewdef, changes take effect on next prompt. No pre-approval needed - if users don't want Claude refining the UI, they simply don't ask for changes.

### MCP Resource: Viewdef Locations

Add an MCP resource that informs Claude about editable viewdefs:

**Resource URI:** `ui://viewdefs`

**Content:**
```markdown
# UI MCP Viewdefs

Viewdefs are HTML templates that control how UI elements are rendered.
You can edit these files to customize the appearance and behavior.

## Locations

- `.ui-mcp/viewdefs/Prompt.DEFAULT.html` - Permission prompt dialog
- `.ui-mcp/viewdefs/Feedback.DEFAULT.html` - Default app UI (if customized)

## Editing Tips

- Use Shoelace components (sl-button, sl-dialog, sl-input, etc.)
- Use `ui-value="path"` for data binding
- Use `ui-action="method()"` for button actions
- Changes take effect on next render
```

This resource appears in Claude's context when ui-mcp is connected, informing it about customization options.

### MCP Resource: Permission History

Track user permission decisions to enable proactive refinement:

**Resource URI:** `ui://permissions/history`

**Content:** JSON log of recent permission decisions:
```json
{
  "decisions": [
    {"tool": "Bash", "command": "grep ...", "choice": "allow_session", "timestamp": "..."},
    {"tool": "Bash", "command": "git status", "choice": "allow", "timestamp": "..."},
    {"tool": "Read", "path": "/home/...", "choice": "allow_session", "timestamp": "..."}
  ]
}
```

**Resource Description (shown to Claude):**
```markdown
# Permission Decision History

This log shows recent user permission choices. Analyze patterns to proactively improve the permission UI:

- If user frequently clicks "Always allow X", suggest making it the default option
- If user always allows certain tool patterns, offer to add them to auto-allow
- If user adds custom options (via viewdef), note which ones get used
- Suggest viewdef refinements based on observed preferences

Example: "I notice you always allow `git` commands. Want me to update the prompt viewdef to show 'Always allow git' as the primary option?"
```

The hook script appends to `.ui-mcp/permissions.log` which this resource reads.

### Prompt Viewdef: `Prompt.DEFAULT.html`

```html
<template>
  <div class="prompt-overlay">
    <sl-dialog open label="Permission Request">
      <p ui-value="pendingPrompt.message"></p>
      <div ui-viewlist="pendingPrompt.options">
        <sl-button ui-action="respondToPrompt(_)" ui-attr-variant="pendingPrompt.options.0.value == value ? 'primary' : 'default'">
          <span ui-value="label"></span>
        </sl-button>
      </div>
    </sl-dialog>
  </div>
</template>
```

### App State for Prompts

```lua
-- When prompt is needed:
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

-- When user clicks:
function app:respondToPrompt(option)
  -- Callback to Go via mcp namespace
  mcp.promptResponse(self.pendingPrompt.id, option.value, option.label)
  self.pendingPrompt = nil
  self._presenter = "Feedback"  -- or previous presenter
end
```

### Go-Lua Bridge

```go
// Register callback on mcp table (already exists for other MCP functions)
mcpTable.RawSetString("promptResponse", L.NewFunction(func(L *lua.LState) int {
    id := L.CheckString(1)
    value := L.CheckString(2)
    label := L.CheckString(3)
    promptManager.Respond(id, value, label)
    return 0
}))
```

### Flow

1. `/api/prompt` called → Go creates prompt ID, sets up response channel
2. Go calls Lua to set `app.pendingPrompt` and switch presenter
3. Browser renders Prompt viewdef via normal variable binding
4. User clicks button → `respondToPrompt(_)` action fires
5. Lua calls `_G.promptResponse()` → Go callback signals channel
6. `/api/prompt` unblocks and returns response

## Hook Script

Location: `.claude/hooks/permission-ui.sh`

```bash
#!/bin/bash
set -e

# Read hook input from stdin
input=$(cat)

# Extract details from hook data
tool_name=$(echo "$input" | jq -r '.tool_name // "unknown tool"')
tool_input=$(echo "$input" | jq -c '.tool_input // {}')
message="Claude wants to use: $tool_name"

# Read MCP port
MCP_PORT=$(cat .ui-mcp/mcp-port 2>/dev/null || echo "")
if [ -z "$MCP_PORT" ]; then
  # UI MCP not running, fall back to terminal
  exit 0
fi

# Build prompt request
request=$(jq -n \
  --arg msg "$message" \
  '{
    "message": $msg,
    "options": [
      {"label": "Allow once", "value": "allow"},
      {"label": "Always allow", "value": "allow_session"},
      {"label": "Deny", "value": "deny"}
    ],
    "timeout": 60
  }')

# Call prompt API
response=$(curl -s -X POST "http://127.0.0.1:$MCP_PORT/api/prompt" \
  -H "Content-Type: application/json" \
  -d "$request")

# Parse response
value=$(echo "$response" | jq -r '.value')
label=$(echo "$response" | jq -r '.label')

# Log decision for pattern analysis
log_entry=$(jq -n \
  --arg tool "$tool_name" \
  --argjson input "$tool_input" \
  --arg choice "$value" \
  --arg ts "$(date -Iseconds)" \
  '{tool: $tool, input: $input, choice: $choice, timestamp: $ts}')
echo "$log_entry" >> .ui-mcp/permissions.log

# Return hook decision
case "$value" in
  "allow"|"allow_session")
    echo '{"decision": "allow"}'
    ;;
  "deny")
    echo '{"decision": "deny", "message": "User denied permission"}'
    ;;
  *)
    # Timeout or error - let terminal prompt handle it
    exit 0
    ;;
esac
```

## Hook Management CLI

Easy install/uninstall via ui-mcp commands:

```bash
# Install permission UI hook
./build/ui-mcp hooks install

# Uninstall hook (reverts to terminal prompts)
./build/ui-mcp hooks uninstall

# Check hook status
./build/ui-mcp hooks status
```

### `hooks install`

1. Creates `.claude/hooks/permission-ui.sh` with the hook script
2. Makes it executable
3. Updates `.claude/settings.json` to add PermissionRequest hook
4. Prints success message with usage instructions

### `hooks uninstall`

1. Removes PermissionRequest hook from `.claude/settings.json`
2. Optionally deletes `.claude/hooks/permission-ui.sh` (with `--delete-script` flag)
3. Prints confirmation

### `hooks status`

Shows:
- Whether hook is installed in settings.json
- Whether hook script exists
- Whether ui-mcp server is running (MCP port file exists)

## Claude Code Hook Configuration (Manual)

If preferred, manually edit `.claude/settings.json`:
```json
{
  "hooks": {
    "PermissionRequest": [
      {
        "type": "command",
        "command": ".claude/hooks/permission-ui.sh",
        "timeout": 120
      }
    ]
  }
}
```

## Implementation Plan

### Phase 1: ui-engine changes

1. Add `PromptManager` to track pending prompts with response channels
2. Add `Prompt(sessionID, message, options, timeout)` method to Server
3. Register `_G.promptResponse` Lua callback that signals channels
4. No protocol changes needed - uses existing variable binding

### Phase 2: ui-mcp changes

1. Start MCP HTTP server on separate port (even in stdio mode)
2. Write port to `.ui-mcp/mcp-port`
3. Add `POST /api/prompt` endpoint that calls ui-engine's `Prompt()`
4. Add `Prompt.DEFAULT.html` viewdef to cache/viewdefs
5. Add `ui://viewdefs` and `ui://permissions/history` MCP resources

### Phase 3: Lua app support

1. Add `respondToPrompt(option)` method to app prototype
2. Handle presenter switching when prompt shown/dismissed

### Phase 4: Hook CLI

1. Add `hooks` subcommand with `install`, `uninstall`, `status`
2. Embed hook script in binary or generate it
3. Handle `.claude/settings.json` JSON manipulation

### Phase 5: Integration testing

1. Test end-to-end flow with actual Claude Code
2. Verify fallback to terminal when server not running
3. Test timeout handling

## Testing

1. Start ui-mcp: `./build/ui-mcp mcp --port 8000 --dir .ui-mcp`
2. Open browser to displayed URL
3. Verify `.ui-mcp/mcp-port` exists
4. Test prompt API directly: `curl -X POST http://127.0.0.1:{port}/api/prompt ...`
5. Configure hook and test with actual permission prompt
