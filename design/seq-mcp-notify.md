# Sequence: MCP Notifications

**Source Spec:** specs/mcp.md (Section 7)

## Participants
- LuaCode: User's Lua application code
- LuaRuntime: Embedded Lua VM manager
- MCPServer: Protocol handler
- AI Agent: External AI assistant (MCP Client)

## Scenario: Lua Code Sends Notification to AI Agent
When Lua code calls `mcp.notify(method, params)`, the notification is forwarded to the connected AI agent.

```
     +---------+             +------------+             +-----------+             +----------+
     | LuaCode |             | LuaRuntime |             | MCPServer |             | AI Agent |
     +----+----+             +-----+------+             +-----+-----+             +-----+----+
          |                        |                         |                         |
          | mcp.notify("event", {data})                      |                         |
          |----------------------->|                         |                         |
          |                        |                         |                         |
          |                        | SendNotification(method, params)                  |
          |                        |------------------------>|                         |
          |                        |                         |                         |
          |                        |                         | ConvertToMap(params)    |
          |                        |                         |-----+                   |
          |                        |                         |     |                   |
          |                        |                         |<----+                   |
          |                        |                         |                         |
          |                        |                         | SendNotificationToAllClients(method, params)
          |                        |                         |------------------------>|
          |                        |                         |                         |
     +----+----+             +-----+------+             +-----+-----+             +-----+----+
     | LuaCode |             | LuaRuntime |             | MCPServer |             | AI Agent |
     +---------+             +------------+             +-----------+             +----------+
```

## Wiring (Initialization)
The notification handler is wired during server startup:

```
     +---------+             +------------+             +-----------+
     |  main   |             | LuaRuntime |             | MCPServer |
     +----+----+             +-----+------+             +-----+-----+
          |                        |                         |
          | NewServer(...)         |                         |
          |------------------------------------------------->|
          |                        |                         |
          | SetNotificationHandler(mcpServer.SendNotification)
          |----------------------->|                         |
          |                        |                         |
          |                        | Store handler           |
          |                        |-----+                   |
          |                        |     |                   |
          |                        |<----+                   |
          |                        |                         |
     +----+----+             +-----+------+             +-----+-----+
     |  main   |             | LuaRuntime |             | MCPServer |
     +---------+             +------------+             +-----------+
```

## Implementation Notes

- Lua params table is converted to `map[string]interface{}` for JSON serialization
- Notifications are sent to all connected MCP clients (typically one AI agent)
- Wire format is JSON-RPC 2.0 notification (no `id` field, no response expected)
