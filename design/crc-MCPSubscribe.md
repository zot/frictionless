# MCPSubscribe

**Source Spec:** specs/publisher.md
**Requirements:** R101, R102, R103, R104, R105, R110, R114, R115

## Knows

- publisherAddr: Publisher address (`http://localhost:25283`)

## Does

- registerSubscribeMethod: Register `mcp:subscribe(topic, handler, opts)` on the mcp Lua global during setupMCPGlobal
- subscribe: Go function backing the Lua method — extracts optional favicon from opts table, starts a background goroutine for the given topic
- pollLoop: Goroutine that long-polls `GET /subscribe/{topic}?favicon=...` in a loop (favicon query param on first request only); on 200, parses JSON and calls handler via SafeExecuteInSession; on 204, reconnects; on connection error, retries after short delay
- callHandler: Execute the Lua handler function in the session context with the parsed data table

## Collaborators

- MCPServer: Provides SafeExecuteInSession for running Lua handlers in session context; hosts Publisher in-process
- Publisher: The publisher hosted by MCPServer, polled via HTTP on localhost:25283
- LuaRuntime: Lua VM where handler functions execute

## Sequences

- seq-publish-subscribe.md: MCP subscribe loop receives published data
- seq-publisher-lifecycle.md: MCP server hosts publisher at startup
