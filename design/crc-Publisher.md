# Publisher

**Source Spec:** specs/publisher.md
**Requirements:** R88, R89, R90, R91, R92, R93, R94, R95, R96, R97, R98, R99, R100, R106, R107, R108, R109, R111, R112, R113

## Knows

- addr: Listen address (`localhost:25283`)
- topics: Map of topic name → Topic (created on demand)
- idleTimeout: Duration before auto-shutdown with zero connections (5 minutes)
- pollTimeout: Long-poll timeout before returning 204 (~60s)
- publishTTL: How long a publish waits for reconnecting subscribers (20ms)
- activeConns: Count of active long-poll connections
- mu: Mutex protecting topics and activeConns

## Does

- main: Parse CLI flags, create Publisher, call listenAndServe
- listenAndServe: Bind to addr, start HTTP server, start idle watchdog goroutine
- handlePublish: POST /publish/{topic} — parse JSON body, get or create topic, deliver to all waiting subscribers, return `{"listeners": N}`
- handleSubscribe: GET /subscribe/{topic}?favicon=... — increment activeConns, get or create topic, if `favicon` query param present store it on the topic, register channel, block until data arrives (return 200 with JSON) or pollTimeout (return 204), decrement activeConns
- handleInstall: GET / — serve HTML page with per-topic bookmarklet sections (each with its favicon if available), instructions, and live topic/listener counts
- handleCORS: Set `Access-Control-Allow-Origin: *` and handle OPTIONS preflight on all endpoints
- getTopic: Return existing topic or create new one
- idleWatchdog: Goroutine that periodically checks activeConns; shuts down server when zero connections persist for idleTimeout

## Topic

A lightweight inner struct (not a separate CRC — it's purely internal to Publisher).

### Knows
- name: Topic name string
- subscribers: Slice of channels waiting for data
- favicon: Data URL string (optional, set by subscribers via query param)

### Does
- addSubscriber: Append a channel, return it
- removeSubscriber: Remove a channel from the slice
- publish: Send data to all subscriber channels, return count; if no subscribers, wait publishTTL then retry once

## Collaborators

- MCPSubscribe: MCP servers connect as subscribers via long-poll
- Bookmarklet: Browser-side JavaScript publishes page content
- OS: Process lifecycle (bind port, exit on idle)

## Sequences

- seq-publisher-lifecycle.md: Start, idle watchdog, auto-shutdown
- seq-publish-subscribe.md: Data flow from publish to subscriber fan-out
