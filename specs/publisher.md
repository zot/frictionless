# Publisher

A shared pub/sub server that lets browser bookmarklets send page content to Frictionless. Runs on a fixed port so bookmarklets always know where to send data, even though MCP sessions use random ports.

## Why

Claude scrapes job posting URLs via WebFetch/Playwright, but this often fails — sites block bots, render dynamically, or require authentication. The user's browser already has the fully rendered page. We need a way to get that content into Frictionless without a browser extension (which would require broad "read all your data" permissions).

A bookmarklet is user-initiated and only reads the current page. The problem is knowing where to send the data — each MCP session gets a random port, and there can be multiple sessions.

## The Publisher

A separate background process on port **25283** (T9 for "claud"). It's a simple topic-based pub/sub server.

- Publishers POST JSON to `/publish/{topic}`
- Subscribers long-poll GET `/subscribe/{topic}`
- When data arrives, all current subscribers get a copy (fan-out)
- No persistence, no history — if nobody's listening, data is dropped

### Starting and Stopping

The publisher starts on demand. When an MCP server tries to subscribe to a topic and the connection fails, it spawns `frictionless publisher` as a detached background process. If the port is already taken, that's fine — one is already running.

It shuts down automatically after being idle (e.g. 5 minutes) with zero connections. Each long-poll counts as a connection, and the idle timer resets on any request.

### Endpoints

**POST /publish/{topic}** — send data to all subscribers of a topic. Returns `{"listeners": N}`.

**GET /subscribe/{topic}** — long-poll. Blocks until data arrives (returns the JSON) or times out after ~60s (returns 204). Client reconnects to keep listening.

**GET /** — install page with the bookmarklet link, instructions, and current topic/listener info.

All endpoints allow CORS `*` since bookmarklets run on arbitrary sites.

### Topics

Topics are implicit — created on first use. No registration, no configuration. After delivering data, subscribers must reconnect to get the next message.

Published messages have a short TTL (20ms) — if no subscribers are connected when data arrives, it waits briefly before dropping. This gives subscribers a grace window to reconnect between long-poll cycles.

## MCP Integration

Apps subscribe to topics via Lua:

```lua
mcp:subscribe("scrape", function(data)
    mcp.pushState({
        app = "job-tracker",
        event = "page_scrape",
        url = data.url,
        title = data.title,
        text = data.text,
    })
end)
```

Under the hood, `mcp:subscribe(topic, handler)` runs a background goroutine that long-polls the publisher. If the connection fails, it tries to start the publisher and retries. On receiving data, it calls the handler and immediately reconnects.

## Bookmarklet

The user drags a bookmarklet to their bookmarks bar (one-time setup via the install page at `localhost:25283`). Clicking it on any page sends the page content to the publisher:

```javascript
javascript:void(fetch('http://localhost:25283/publish/scrape',{
  method:'POST',
  headers:{'Content-Type':'application/json'},
  body:JSON.stringify({
    url:location.href,
    title:document.title,
    text:document.body.innerText.slice(0,50000)
  })
}).then(r=>r.json()).then(d=>{
  let n=d.listeners||0;
  document.title='[Sent to '+n+' session'+(n!=1?'s':'')+'] '+document.title
}).catch(()=>alert('Frictionless publisher not running')))
```

It captures three things:
- **url** — `location.href`
- **title** — `document.title`
- **text** — `document.body.innerText`, truncated to 50k chars

`innerText` is the key insight — it includes JS-rendered content, clean text without HTML, and works on authenticated pages the user is logged into.

Feedback: on success, the tab title shows `[Sent to N session(s)]`. On failure, an alert says the publisher isn't running.

## Topic Favicons

When an app subscribes to a topic, it can supply a favicon (a data URL). The publisher stores this per-topic and uses it on the install page.

The subscribe endpoint accepts an optional `favicon` query parameter:

```
GET /subscribe/job-tracker?favicon=data:image/svg+xml;base64,...
```

The publisher stores the favicon with the topic. If multiple subscribers provide different favicons for the same topic, the most recent one wins.

The install page shows a per-topic bookmarklet section. Each topic with a favicon gets its own draggable bookmarklet link with the favicon displayed next to it. Topics without favicons appear in a plain list as before.

On the MCP side, `mcp:subscribe` accepts an optional third argument:

```lua
mcp:subscribe("job-tracker", handler, {favicon = "data:image/svg+xml;base64,..."})
```

The favicon is read from the app's `favicon.svg` file (created as part of the app's favicon support). The subscribe goroutine passes it as a query parameter on its first long-poll request to the publisher.

## Beyond Scraping

The topic model is generic. Any app can subscribe to any topic:
- `scrape` — page content for any app
- `clipboard` — clipboard data between browser and apps
- `notify` — push notifications from external tools
