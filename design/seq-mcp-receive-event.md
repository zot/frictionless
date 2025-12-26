# Sequence: MCP Receive Event

**Source Spec:** interfaces.md
**Use Case:** AI receiving user interaction events via MCP

## Participants

- User: User interacting with UI
- Frontend: Browser frontend
- ProtocolHandler: Message processor
- LuaRuntime: Lua presenter logic
- MCPServer: MCP protocol handler
- AIClient: AI assistant

## Sequence

```
     User               Frontend          ProtocolHandler          LuaRuntime            MCPServer             AIClient
        |                      |                      |                      |                      |                      |
        |---click button------>|                      |                      |                      |                      |
        |   (ui-action)        |                      |                      |                      |                      |
        |                      |                      |                      |                      |                      |
        |                      |---update(var,------->|                      |                      |                      |
        |                      |    {action:name,     |                      |                      |                      |
        |                      |     values:{...}})   |                      |                      |                      |
        |                      |                      |                      |                      |                      |
        |                      |                      |---isLuaPresenter?--->|                      |                      |
        |                      |                      |                      |                      |                      |
        |                      |          [if Lua presenter]                 |                      |                      |
        |                      |                      |---callMethod-------->|                      |                      |
        |                      |                      |                      |                      |                      |
        |                      |                      |                      |---execute()--------->|                      |
        |                      |                      |                      |                      |                      |
        |                      |                      |<--result-------------|                      |                      |
        |                      |                      |                      |                      |                      |
        |                      |          [if needs AI response]             |                      |                      |
        |                      |                      |---queueEvent-------->|                      |                      |
        |                      |                      |                      |---notify------------>|                      |
        |                      |                      |                      |                      |                      |
        |                      |                      |                      |                      |---notification------>|
        |                      |                      |                      |                      |   (user_event)       |
        |                      |                      |                      |                      |                      |
        |                      |                      |                      |                      |     [AI processes]   |
        |                      |                      |                      |                      |<--tools/call---------|
        |                      |                      |                      |                      |   (update_presenter) |
        |                      |                      |                      |                      |                      |
        |                      |<--update-------------|                      |                      |                      |
        |                      |                      |                      |                      |                      |
        |<--UI update----------|                      |                      |                      |                      |
        |                      |                      |                      |                      |                      |
```

## Notes

- User action triggers ui-action or value update
- Lua logic can handle locally if defined
- Events can be queued for AI processing
- AI receives notification of pending events
- AI can query events via resources
- AI responds with UI updates
- Two-way conversation loop enabled
