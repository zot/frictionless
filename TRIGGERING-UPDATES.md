# Triggering Updates from MCP

When MCP's `ui_run` executes Lua code that modifies app state, changes must be pushed to the browser. The ui-engine now provides `Server.ExecuteInSession` which handles this automatically.

## How It Works

`Server.ExecuteInSession(vendedID, fn)`:
1. Queues the function through the session's executor (serializes with WebSocket messages)
2. Executes the function
3. Calls `afterBatch` to detect and push changes to the browser
4. Returns the result

## Usage in MCP

Instead of:
```go
result, err := s.runtime.ExecuteInSession(sessionID, func() (interface{}, error) {
    return s.runtime.LoadCodeDirect("mcp-run", code)
})
s.server.TriggerUpdates(sessionID)  // separate call, now removed
```

Use:
```go
result, err := s.server.ExecuteInSession(sessionID, func() (interface{}, error) {
    return s.runtime.ExecuteInSession(sessionID, func() (interface{}, error) {
        return s.runtime.LoadCodeDirect("mcp-run", code)
    })
})
// No separate call needed - afterBatch is called automatically
```

## Why Two Nested ExecuteInSession Calls?

- **`Server.ExecuteInSession`**: Serializes with WebSocket messages via session executor, calls `afterBatch` after
- **`Runtime.ExecuteInSession`**: Sets up Lua session context (`session` global), serializes Lua VM access

Both are needed because:
1. The session executor ensures MCP and WebSocket operations don't interleave
2. The Lua executor ensures thread-safe Lua VM access and session context setup

## Key Points

- `Server.ExecuteInSession` uses `SvcSync` (blocks until complete) - appropriate for MCP which needs results
- WebSocket message handling uses `Svc` (async) - appropriate for fire-and-forget message processing
- Both go through the same per-session executor channel, ensuring serialization
