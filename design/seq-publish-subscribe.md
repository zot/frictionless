# Sequence: Publish and Subscribe

**Requirements:** R89, R90, R94, R95, R101, R102, R104, R105, R106, R107, R110, R116, R117, R118, R119, R120, R121, R122, R123, R124

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

## CSP-Safe Relay Flow: Bookmarklet → Relay Page → Publisher → MCP Sessions

```
Bookmarklet              Relay Page (tab)         Publisher              MCP-A (pollLoop)
    |                        |                        |                       |
    |                        |                   GET /subscribe/scrape        |
    |                        |                        |<----------------------|
    |                        |                        |                       |
    |-- window.open ---------->                       |                       |
    |   /relay/scrape        |                        |                       |
    |                        |-- postMessage('ready') |                       |
    |<-----------------------|   to opener            |                       |
    |                        |                        |                       |
    |-- postMessage(data) -->|                        |                       |
    |   {url,title,text}     |                        |                       |
    |                        |-- POST /publish/scrape |                       |
    |                        |   (same-origin)------->|                       |
    |                        |                        |-- fan-out:            |
    |                        |                        |   send to MCP-A ch -->|
    |                        |                        |                       |
    |                        |<-- {"listeners": 1} ---|                  200 + JSON
    |                        |                        |                       |
    |                        |-- show "Sent to 1      |                  reconnect
    |                        |   session"             |                       |
    |                        |-- auto-close (1.5s)    |                       |
```
