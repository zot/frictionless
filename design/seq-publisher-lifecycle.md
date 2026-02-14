# Sequence: Publisher Lifecycle

**Requirements:** R96, R97, R98

## MCP Server Hosts Publisher at Startup

```
MCPServer.Start()                    Publisher
    |                                    |
    |-- tryStartPublisher (goroutine):   |
    |   pub := publisher.New(addr)       |
    |   pub.ListenAndServe()             |
    |------------------------------------+-- bind localhost:25283
    |                                    |-- listening
    |                                    |
    |-- (continue MCP startup)           |
```

If port 25283 is already bound (another MCP server has it), ListenAndServe
returns an error and the goroutine exits silently. The subscribe poll loop
connects to whichever MCP server holds the port.

## Publisher Shutdown

```
MCPServer                            Publisher
    |                                    |
    |-- Stop() or process exit           |
    |                                    |-- server closes
    |                                    |-- port released
    |                                    |
                                    Other MCP servers' pollLoops
                                         |-- connection error
                                         |-- retry after delay
                                         |-- (next MCP server to start grabs port)
```

No idle watchdog, no forked processes. Publisher lives and dies with its MCP server.
