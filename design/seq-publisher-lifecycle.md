# Sequence: Publisher Lifecycle

**Requirements:** R96, R97, R98, R99, R100

## Start on Demand

```
MCPSubscribe                        OS                          Publisher
    |                                |                              |
    |-- pollLoop: GET /subscribe --->|                              |
    |<- connection refused ----------|                              |
    |                                |                              |
    |-- ensurePublisher:             |                              |
    |   exec "frictionless publisher"|                              |
    |------------------------------->|                              |
    |                                |-- spawn detached process --->|
    |                                |                              |-- bind localhost:25283
    |                                |                              |-- start idleWatchdog
    |                                |                              |-- listening
    |                                |                              |
    |-- sleep 500ms                  |                              |
    |-- retry GET /subscribe ------->|----------------------------->|
    |                                |                   connected  |
```

If port is already bound (another publisher running), the spawn fails silently.
The retry succeeds because the existing publisher accepts the connection.

## Idle Auto-Shutdown

```
Publisher                     idleWatchdog
    |                              |
    |                              |-- check activeConns
    |                              |   activeConns == 0
    |                              |-- start idle timer
    |                              |
    |-- handleSubscribe            |
    |   activeConns++              |
    |                              |-- check activeConns
    |                              |   activeConns > 0
    |                              |-- reset idle timer
    |                              |
    |-- subscriber disconnects     |
    |   activeConns--              |
    |                              |-- check activeConns
    |                              |   activeConns == 0
    |                              |-- start idle timer
    |                              |
    |                              |-- 5 minutes elapsed
    |                              |   activeConns still 0
    |                              |-- shutdown server
    |-- exit                       |
```
