---
description: Capture page content from external websites using a draggable bookmarklet. Uses the publisher pub/sub server with a postMessage relay to bypass CSP restrictions.
---

# Bookmarklet Page Capture

Capture URL, title, and text from any website and deliver it to your app as a `page_received` event. Works on CSP-restricted sites (LinkedIn, etc.) via a relay page.

## When to Use

- Your app needs to ingest content from external web pages
- Users browse to pages and want to send them to the app with one click
- You need the rendered `innerText` (JS-rendered content, no HTML tags, works on authenticated pages)

## Architecture

```
Browser tab (any site)           Publisher server (:25283)         Lua session
─────────────────────           ──────────────────────────        ────────────
User clicks bookmarklet  ──→  Opens /relay/{topic} in new tab
                               Relay sends "ready" to opener
Opener posts {url,title,text}  Relay receives via postMessage
                               Relay POSTs to /publish/{topic}
                               Publisher fans out to subscribers  ──→  handler(data)
                                                                      pushState({event: "page_received", ...})
```

The relay page is needed because CSP on many sites blocks direct `fetch()` to localhost. Instead, the bookmarklet opens a same-origin relay tab on the publisher, which can freely POST.

## Implementation

### 1. init.lua — Subscribe to Topic

Create `apps/<app>/init.lua` to subscribe to a publisher topic. This file runs once on app load (not on hot-reload).

```lua
-- Subscribe to the "<app>" publisher topic for bookmarklet page captures.
-- When a user clicks the bookmarklet on a page, the content
-- arrives here and gets pushed as a page_received event for Claude to handle.
mcp:subscribe("<app>", function(data)
    mcp.pushState({
        app = "<app>",
        event = "page_received",
        url = data.url,
        title = data.title,
        text = data.text,
    })
end, {favicon = "data:image/svg+xml;base64,..."})
```

**Parameters:**
- Topic name: Use the app name (e.g., `"job-tracker"`)
- `favicon`: Optional base64 data URL shown on the publisher install page (`http://localhost:25283/`)
- The callback receives `{url, title, text}` — the page's URL, document title, and body innerText (up to 50KB)

### 2. Bookmarklet Link in Viewdef

Add a draggable bookmarklet link. Users drag it to their bookmarks bar.

**CSS:**
```css
.bookmarklet-toggle {
  font-size: 0.7rem;
  color: var(--term-text-dim);
  text-decoration: none;
  cursor: pointer;
}
.bookmarklet-toggle:hover {
  color: var(--term-accent);
}
.bookmarklet-section {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 16px;
  background: var(--term-bg);
  border-bottom: 1px solid var(--term-border);
}
.bookmarklet-hint {
  font-size: 0.8rem;
  color: var(--term-text-dim);
}
.bookmarklet-link {
  display: inline-block;
  padding: 3px 12px;
  background: var(--term-accent);
  color: #fff;
  border-radius: 4px;
  text-decoration: none;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: grab;
}
.bookmarklet-link:hover {
  filter: brightness(0.9);
}
```

**HTML:**
```html
<!-- Toggle link in header -->
<a class="bookmarklet-toggle" href="#" ui-event-click="toggleBookmarklet()">bookmarklet</a>

<!-- Collapsible section -->
<div class="bookmarklet-section" ui-class-hidden="isBookmarkletHidden()">
  <span class="bookmarklet-hint">Drag this to your bookmarks bar:</span>
  <a class="bookmarklet-link"
     href="javascript:void(function(){var d={url:location.href,title:document.title,text:document.body.innerText.slice(0,50000)};var w=window.open('http://localhost:25283/relay/TOPIC','_blank');if(!w){alert('Please allow popups for this site');return}window.addEventListener('message',function h(e){if(e.origin==='http://localhost:25283'&&e.data==='ready'){w.postMessage(d,'http://localhost:25283');window.removeEventListener('message',h)}});}())">
    Add to App
  </a>
  <span class="bookmarklet-hint">Browse to a page and click it.</span>
</div>
```

Replace `TOPIC` in the href with your app's topic name (must match init.lua).

### 3. Lua Toggle Methods

```lua
local MyApp = session:prototype("MyApp", {
    showBookmarklet = false,
    -- ...
})

function MyApp:toggleBookmarklet()
    self.showBookmarklet = not self.showBookmarklet
end

function MyApp:isBookmarkletHidden()
    return not self.showBookmarklet
end
```

### 4. Event Handling in design.md

Document the `page_received` event in your app's design:

```markdown
## Events

| Event | Action |
|-------|--------|
| `page_received` | Page content from bookmarklet. Extract structured data from text, prefill form. No fetching needed — text is pre-rendered `innerText` from the user's browser. |
```

The event payload is:
```json
{"app": "<app>", "event": "page_received", "url": "https://...", "title": "Page Title", "text": "...rendered page text..."}
```

**Key point:** The `text` field is `document.body.innerText` — already clean text with no HTML tags. No scraping or fetching is needed. This also captures JS-rendered content and works on authenticated pages since the bookmarklet runs in the user's browser session.

## How the Relay Works

The bookmarklet can't POST directly to localhost because many sites (LinkedIn, GitHub, etc.) have strict Content Security Policies. Instead:

1. **Bookmarklet** collects `{url, title, text}` and opens `http://localhost:25283/relay/{topic}` in a new tab
2. **Relay page** (served by the publisher) sends `"ready"` back to the opener via `postMessage`
3. **Bookmarklet** receives "ready" and posts the data to the relay via `postMessage`
4. **Relay page** receives the data and does a same-origin `fetch('/publish/{topic}', {body: data})`
5. **Publisher** fans the data out to all long-poll subscribers
6. **Subscriber** (init.lua) receives data and calls `pushState` to create the event
7. **Relay page** shows "Sent to N sessions" and auto-closes after 1.5s

## Publisher Install Page

The publisher also serves an install page at `http://localhost:25283/` that shows:
- A default "Send to Frictionless" bookmarklet
- A list of active topics with per-topic bookmarklets and subscriber counts
- Topic favicons (if provided via the `favicon` option in `mcp:subscribe`)

Users can visit this page directly to install bookmarklets instead of using the in-app toggle.

## Key Points

- **Topic name = app name** by convention
- **50KB text limit** — `innerText.slice(0, 50000)` keeps payloads reasonable
- **No fetching needed** — the text arrives pre-rendered, saving a network round-trip and handling auth/JS content
- **Popup blocker** — the bookmarklet alerts the user if the relay popup is blocked
- **init.lua vs app.lua** — use init.lua for the subscription so it runs once, not on every hot-reload
- **Publisher port** — hardcoded to `localhost:25283`, shared across all MCP instances (first one wins)
