# Sequence: Publish and Subscribe

**Requirements:** R89, R90, R94, R95, R101, R102, R104, R105, R106, R107, R110

## Normal Flow: Bookmarklet → Publisher → MCP Sessions

```
Bookmarklet              Publisher                MCP-A (pollLoop)      MCP-B (pollLoop)
    |                        |                        |                      |
    |                        |<-- GET /subscribe/scrape (blocking)           |
    |                        |<-- GET /subscribe/scrape (blocking) ----------|
    |                        |                        |                      |
    |-- POST /publish/scrape |                        |                      |
    |   {url, title, text}   |                        |                      |
    |----------------------->|                        |                      |
    |                        |-- fan-out:             |                      |
    |                        |   send to MCP-A ch --->|                      |
    |                        |   send to MCP-B ch --->|---------------------->
    |                        |                        |                      |
    |<-- {"listeners": 2} ---|                        |                      |
    |                        |                        |                      |
    |-- update tab title     |                   200 + JSON             200 + JSON
    |   "[Sent to 2          |                        |                      |
    |    sessions]"          |-- callHandler:         |                      |
    |                        |   SafeExecuteInSession |                      |
    |                        |   handler(data) in Lua |                      |
    |                        |                        |                      |
    |                        |                   reconnect              reconnect
    |                        |<-- GET /subscribe/scrape                      |
    |                        |<-- GET /subscribe/scrape ---------------------|
```

## Publish with No Subscribers (TTL Wait)

```
Bookmarklet              Publisher
    |                        |
    |-- POST /publish/scrape |
    |----------------------->|
    |                        |-- no subscribers
    |                        |-- wait 20ms (publishTTL)
    |                        |-- still no subscribers
    |                        |
    |<-- {"listeners": 0} ---|
    |                        |
    |-- update tab title     |
    |   "[Sent to 0          |
    |    sessions]"          |
```

## Long-Poll Timeout

```
Publisher                MCP-A (pollLoop)
    |                        |
    |<-- GET /subscribe/scrape
    |                        |
    |   ... 60s pass ...     |
    |                        |
    |-- 204 No Content ----->|
    |                        |
    |                   reconnect
    |<-- GET /subscribe/scrape
```
