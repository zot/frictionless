// Package publisher provides a pub/sub server on a fixed port, hosted in-process by the MCP server.
// Browser bookmarklets publish page content; MCP sessions subscribe via long-poll.
// CRC: crc-Publisher.md | Seq: seq-publisher-lifecycle.md, seq-publish-subscribe.md
package publisher

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	DefaultAddr = "localhost:25283"
	PollTimeout = 60 * time.Second
	PublishTTL  = 20 * time.Millisecond
	MaxBodySize = 1 << 20 // 1MB
)

// Publisher is a topic-based pub/sub HTTP server.
type Publisher struct {
	addr   string
	topics map[string]*topic
	mu     sync.Mutex
}

type topic struct {
	mu          sync.Mutex
	subscribers []chan json.RawMessage
	favicon     string // data URL, set by subscribers via query param
}

// New creates a Publisher bound to the given address.
func New(addr string) *Publisher {
	return &Publisher{
		addr:   addr,
		topics: make(map[string]*topic),
	}
}

// ListenAndServe starts the HTTP server. Blocks until the server shuts down.
func (p *Publisher) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/publish/", p.handlePublish)
	mux.HandleFunc("/subscribe/", p.handleSubscribe)
	mux.HandleFunc("/relay/", p.handleRelay)
	mux.HandleFunc("/", p.handleInstall)

	srv := &http.Server{
		Addr:    p.addr,
		Handler: corsMiddleware(mux),
	}

	// Try to bind — if port is taken, another MCP server is hosting the publisher
	ln, err := net.Listen("tcp", p.addr)
	if err != nil {
		return fmt.Errorf("bind %s: %w (another instance may be hosting it)", p.addr, err)
	}

	log.Printf("Publisher listening on %s", p.addr)
	return srv.Serve(ln)
}

// handlePublish delivers a JSON body to all subscribers of a topic.
// POST /publish/{topic}
func (p *Publisher) handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/publish/")
	if name == "" {
		http.Error(w, "topic name required", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, MaxBodySize))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	p.mu.Lock()
	t := p.getTopic(name)
	p.mu.Unlock()

	n := t.publish(body)

	// If no subscribers, wait briefly for reconnecting ones
	if n == 0 {
		time.Sleep(PublishTTL)
		n = t.publish(body)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"listeners":%d}`, n)
}

// handleSubscribe long-polls until data is published to the topic.
// GET /subscribe/{topic}
func (p *Publisher) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/subscribe/")
	if name == "" {
		http.Error(w, "topic name required", http.StatusBadRequest)
		return
	}

	p.mu.Lock()
	t := p.getTopic(name)
	p.mu.Unlock()

	// Store favicon if provided (most recent wins)
	if fav := r.URL.Query().Get("favicon"); fav != "" {
		t.mu.Lock()
		t.favicon = fav
		t.mu.Unlock()
	}

	ch := t.addSubscriber()
	defer t.removeSubscriber(ch)

	select {
	case data := <-ch:
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	case <-time.After(PollTimeout):
		w.WriteHeader(http.StatusNoContent)
	case <-r.Context().Done():
		// Client disconnected
	}
}

// handleInstall serves the bookmarklet install page.
// GET /
func (p *Publisher) handleInstall(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	p.mu.Lock()
	topicInfo := p.topicSummary()
	p.mu.Unlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, installPageHTML, bookmarkletJS, topicInfo)
}

// handleRelay serves a CSP-safe relay page that receives data via postMessage and POSTs same-origin.
// GET /relay/{topic}
// CRC: crc-Publisher.md | Seq: seq-publish-subscribe.md
func (p *Publisher) handleRelay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/relay/")
	if name == "" {
		http.Error(w, "topic name required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, relayPageHTML, name)
}

// getTopic returns an existing topic or creates a new one. Caller must hold p.mu.
func (p *Publisher) getTopic(name string) *topic {
	t, ok := p.topics[name]
	if !ok {
		t = &topic{}
		p.topics[name] = t
	}
	return t
}

// topicSummary returns an HTML snippet of current topics with per-topic bookmarklets. Caller must hold p.mu.
func (p *Publisher) topicSummary() string {
	if len(p.topics) == 0 {
		return "<p>No active topics.</p>"
	}
	var sb strings.Builder
	for name, t := range p.topics {
		t.mu.Lock()
		n := len(t.subscribers)
		fav := t.favicon
		t.mu.Unlock()

		sb.WriteString("<div class=\"topic\">")
		if fav != "" {
			fmt.Fprintf(&sb, "<img class=\"topic-icon\" src=\"%s\" width=\"20\" height=\"20\">", fav)
		}
		fmt.Fprintf(&sb, "<strong>%s</strong> — %d listener(s)", name, n)

		// Per-topic bookmarklet
		topicJS := strings.ReplaceAll(bookmarkletTpl, "TOPIC", name)
		fmt.Fprintf(&sb, " &nbsp; <a class=\"bookmarklet small\" href=\"%s\">Send to %s</a>", topicJS, name)
		sb.WriteString("</div>")
	}
	return sb.String()
}

// addSubscriber creates a buffered channel and registers it for the topic.
func (t *topic) addSubscriber() chan json.RawMessage {
	ch := make(chan json.RawMessage, 1)
	t.mu.Lock()
	t.subscribers = append(t.subscribers, ch)
	t.mu.Unlock()
	return ch
}

// removeSubscriber unregisters a channel from the topic.
func (t *topic) removeSubscriber(ch chan json.RawMessage) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, s := range t.subscribers {
		if s == ch {
			t.subscribers = append(t.subscribers[:i], t.subscribers[i+1:]...)
			return
		}
	}
}

// publish sends data to all current subscribers. Returns count delivered.
func (t *topic) publish(data json.RawMessage) int {
	t.mu.Lock()
	subs := make([]chan json.RawMessage, len(t.subscribers))
	copy(subs, t.subscribers)
	t.mu.Unlock()

	n := 0
	for _, ch := range subs {
		select {
		case ch <- data:
			n++
		default:
			// Channel full, skip
		}
	}
	return n
}

// corsMiddleware adds CORS headers to all responses.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// bookmarkletTpl is a bookmarklet template with TOPIC as a placeholder for the topic name.
// Uses window.open + postMessage relay to bypass CSP restrictions on sites like LinkedIn.
// CRC: crc-Publisher.md | Seq: seq-publish-subscribe.md
const bookmarkletTpl = `javascript:void(function(){var d={url:location.href,title:document.title,text:document.body.innerText.slice(0,50000)};var w=window.open('http://localhost:25283/relay/TOPIC','_blank');if(!w){alert('Please allow popups for this site');return}window.addEventListener('message',function h(e){if(e.origin==='http://localhost:25283'&&e.data==='ready'){w.postMessage(d,'http://localhost:25283');window.removeEventListener('message',h)}});}())`

// bookmarkletJS is the default bookmarklet for the "page" topic.
var bookmarkletJS = strings.ReplaceAll(bookmarkletTpl, "TOPIC", "page")

// installPageHTML is the bookmarklet install page. %s format verbs: bookmarkletJS, topicInfo.
const installPageHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Frictionless Publisher</title>
<style>
  body { font-family: system-ui, sans-serif; max-width: 640px; margin: 2em auto; padding: 0 1em; color: #e0e0e8; background: #0a0a0f; }
  a.bookmarklet { display: inline-block; padding: 0.6em 1.2em; background: #E07A47; color: #fff; border-radius: 6px; text-decoration: none; font-weight: 600; cursor: grab; }
  a.bookmarklet:hover { background: #c56a3a; }
  code { background: #1a1a24; padding: 0.15em 0.4em; border-radius: 3px; font-size: 0.9em; }
  h1 { color: #E07A47; }
  h2 { color: #8888a0; }
  .topics { margin-top: 2em; padding-top: 1em; border-top: 1px solid #2a2a3a; }
  .topic { margin: 0.8em 0; padding: 0.5em 0; }
  a.bookmarklet.small { padding: 0.3em 0.8em; font-size: 0.85em; }
  .topic-icon { vertical-align: middle; margin-right: 8px; }
</style>
</head>
<body>
<h1>Frictionless Publisher</h1>
<p>Drag this to your bookmarks bar:</p>
<p><a class="bookmarklet" href="%s">Send to Frictionless</a></p>
<h2>How it works</h2>
<ol>
<li>Click the bookmarklet on any page</li>
<li>A small relay tab opens briefly to send the data</li>
<li>All connected Frictionless sessions receive the data</li>
<li>The relay tab shows how many sessions received it, then auto-closes</li>
</ol>
<p>The bookmarklet captures <code>innerText</code> — rendered text including JS content, no HTML tags, works on authenticated pages.</p>
<div class="topics">
<h2>Active Topics</h2>
%s
</div>
</body>
</html>`

// relayPageHTML is the CSP-safe relay page template. %q format verb: topic name.
// CRC: crc-Publisher.md | Seq: seq-publish-subscribe.md
const relayPageHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Sending to Frictionless…</title>
<style>
  body { font-family: system-ui, sans-serif; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; background: #0a0a0f; color: #e0e0e8; }
  .card { text-align: center; padding: 2em; }
  h2 { color: #E07A47; margin: 0 0 0.5em; }
  .status { font-size: 1.1em; color: #8888a0; }
  .ok { color: #5cb85c; }
  .err { color: #d9534f; }
</style>
</head>
<body>
<div class="card">
  <h2>Frictionless</h2>
  <div class="status" id="status">Waiting for page data…</div>
</div>
<script>
(function(){
  var topic = %q;
  var status = document.getElementById('status');
  var timeout = setTimeout(function(){
    status.textContent = 'Timed out — no data received.';
    status.className = 'status err';
  }, 10000);

  if (window.opener) { window.opener.postMessage('ready', '*'); }

  window.addEventListener('message', function(evt) {
    if (!evt.data || !evt.data.url) return;
    clearTimeout(timeout);
    status.textContent = 'Sending…';

    fetch('/publish/' + topic, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify(evt.data)
    })
    .then(function(resp) { return resp.json(); })
    .then(function(result) {
      var count = result.listeners || 0;
      status.textContent = 'Sent to ' + count + ' session' + (count !== 1 ? 's' : '') + '.';
      status.className = 'status ok';
      setTimeout(function() { window.close(); }, 1500);
    })
    .catch(function() {
      status.textContent = 'Failed to send data.';
      status.className = 'status err';
    });
  });
})();
</script>
</body>
</html>`
