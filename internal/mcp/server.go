// Package mcp implements the Model Context Protocol server.
// CRC: crc-MCPServer.md
// Spec: mcp.md
// Sequence: seq-mcp-lifecycle.md, seq-mcp-state-wait.md
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/server"
	lua "github.com/yuin/gopher-lua"
	"github.com/zot/ui-engine/cli"
)

// State represents the lifecycle state of the MCP server.
type State int

const (
	Unconfigured State = iota
	Configured
	Running
)

// Server implements an MCP server for AI integration.
type Server struct {
	mcpServer         *server.MCPServer
	cfg               *cli.Config
	UiServer          *cli.Server // UI engine server for ExecuteInSession
	viewdefs          *cli.ViewdefManager
	startFunc         func(port int) (string, error) // Callback to start HTTP server
	onViewdefUploaded func(typeName string)          // Callback when a viewdef is uploaded
	getSessionCount   func() int                     // Callback to get active session count

	mu              sync.RWMutex
	state           State
	baseDir         string
	url             string
	httpServer      *http.Server // HTTP server for debug endpoints (in stdio mode)
	mcpPort         int          // Port for MCP HTTP server (written to baseDir/mcp-port)
	uiPort          int          // Port for UI HTTP server (written to baseDir/ui-port)
	currentVendedID string       // Current session's vended ID (e.g., "1")
	logPath         string       // Path for Lua log file (set at configure time)
	errPath         string       // Path for Lua error log file (set at configure time)

	// State change waiting (mcp.state queue)
	stateWaiters   map[string][]chan struct{} // sessionID -> list of waiting channels
	stateQueue     map[string][]interface{}   // sessionID -> queued events
	stateWaitersMu sync.Mutex                 // Protects stateWaiters and stateQueue
}

// NewServer creates a new MCP server.
func NewServer(cfg *cli.Config, uiServer *cli.Server, viewdefs *cli.ViewdefManager, startFunc func(port int) (string, error), onViewdefUploaded func(typeName string), getSessionCount func() int) *Server {
	s := server.NewMCPServer("ui-server", "0.1.0")
	srv := &Server{
		mcpServer:         s,
		cfg:               cfg,
		UiServer:          uiServer,
		viewdefs:          viewdefs,
		startFunc:         startFunc,
		onViewdefUploaded: onViewdefUploaded,
		getSessionCount:   getSessionCount,
		state:             Unconfigured,
		stateWaiters:      make(map[string][]chan struct{}),
		stateQueue:        make(map[string][]interface{}),
	}
	srv.registerTools()
	srv.registerResources()
	return srv
}

// SafeExecuteInSession wraps ExecuteInSession with panic recovery to prevent crashes.
func (s *Server) SafeExecuteInSession(sessionID string, fn func() (interface{}, error)) (result interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			s.cfg.Log(0, "PANIC in ExecuteInSession: %v", r)
			err = fmt.Errorf("panic during execution: %v", r)
		}
	}()
	return s.UiServer.ExecuteInSession(sessionID, fn)
}

// ServeStdio starts the MCP server on Stdin/Stdout.
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcpServer)
}

// ServeSSE starts the MCP server as an SSE HTTP server on the given address.
// Spec: mcp.md Section 2.3
func (s *Server) ServeSSE(addr string) error {
	sseServer := server.NewSSEServer(s.mcpServer)

	// Wrap SSE server with our custom handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/variables", s.handleVariables)
	mux.HandleFunc("/state", s.handleState)
	mux.HandleFunc("/wait", s.handleWait)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sseServer.ServeHTTP(w, r)
	})

	s.cfg.Log(0, "Starting MCP SSE server on %s (/variables, /state, /wait)", addr)

	// Parse port from addr and write mcp-port file
	if _, portStr, err := net.SplitHostPort(addr); err == nil {
		if port, err := strconv.Atoi(portStr); err == nil {
			if err := s.WriteMCPPortFile(port); err != nil {
				s.cfg.Log(0, "Warning: failed to write mcp-port file: %v", err)
			}
		}
	}

	// Start HTTP server
	httpServer := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return httpServer.ListenAndServe()
}

// StartHTTPServer starts a standalone HTTP server in stdio mode.
// Serves debug pages and state wait endpoint.
// Returns the port number.
// Spec: mcp.md Section 2.2
func (s *Server) StartHTTPServer() (int, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/variables", s.handleVariables)
	mux.HandleFunc("/state", s.handleState)
	mux.HandleFunc("/wait", s.handleWait)

	// Listen on random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, fmt.Errorf("failed to listen: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	s.mcpPort = port

	s.httpServer = &http.Server{
		Handler: mux,
	}

	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			s.cfg.Log(0, "HTTP server error: %v", err)
		}
	}()

	s.cfg.Log(0, "HTTP server listening on port %d (/variables, /state, /wait)", port)

	// Write mcp-port file
	if err := s.WriteMCPPortFile(port); err != nil {
		s.cfg.Log(0, "Warning: failed to write mcp-port file: %v", err)
	}

	return port, nil
}

// WriteMCPPortFile writes the MCP port to the mcp-port file in baseDir.
// Spec: mcp.md Section 5.2
func (s *Server) WriteMCPPortFile(port int) error {
	s.mu.Lock()
	s.mcpPort = port
	baseDir := s.baseDir
	s.mu.Unlock()

	if baseDir == "" {
		return fmt.Errorf("server not configured")
	}

	mcpPortFile := filepath.Join(baseDir, "mcp-port")
	return os.WriteFile(mcpPortFile, []byte(strconv.Itoa(port)), 0644)
}

// WriteUIPortFile writes the UI port to the ui-port file in baseDir.
// Spec: mcp.md Section 5.2
func (s *Server) WriteUIPortFile(port int) error {
	s.mu.Lock()
	s.uiPort = port
	baseDir := s.baseDir
	s.mu.Unlock()

	if baseDir == "" {
		return fmt.Errorf("server not configured")
	}

	uiPortFile := filepath.Join(baseDir, "ui-port")
	return os.WriteFile(uiPortFile, []byte(strconv.Itoa(port)), 0644)
}

// RemovePortFiles removes the mcp-port and ui-port files.
func (s *Server) RemovePortFiles() {
	s.mu.RLock()
	baseDir := s.baseDir
	s.mu.RUnlock()

	if baseDir != "" {
		os.Remove(filepath.Join(baseDir, "mcp-port"))
		os.Remove(filepath.Join(baseDir, "ui-port"))
	}
}

// ShutdownHTTPServer shuts down the standalone HTTP server.
func (s *Server) ShutdownHTTPServer(ctx context.Context) error {
	if s.httpServer != nil {
		s.RemovePortFiles()
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// Configure transitions the server to the Configured state.
// Spec: mcp.md
// CRC: crc-MCPServer.md
func (s *Server) Configure(baseDir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == Running {
		return fmt.Errorf("Cannot reconfigure while running")
	}

	s.baseDir = baseDir
	s.state = Configured

	return nil
}

// Start transitions the server to the Running state and starts the HTTP server.
// Spec: mcp.md
// CRC: crc-MCPServer.md
func (s *Server) Start() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == Unconfigured {
		return "", fmt.Errorf("Server not configured")
	}
	if s.state == Running {
		return "", fmt.Errorf("Server already running")
	}

	// Select random port (0)
	url, err := s.startFunc(0)
	if err != nil {
		return "", err
	}

	s.state = Running
	s.url = url
	return url, nil
}

// Stop destroys the current session and resets state to Configured.
// This allows reconfiguration without restarting the process.
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != Running {
		return nil // Nothing to stop
	}

	// Destroy the current session if we have one
	if s.currentVendedID != "" {
		sessions := s.UiServer.GetSessions()
		internalID := sessions.GetInternalID(s.currentVendedID)
		if internalID != "" {
			sessions.DestroySession(internalID)
		}
		s.currentVendedID = ""
	}

	// Reset state (keep baseDir for reconfiguration)
	s.state = Configured
	s.url = ""

	return nil
}

// SendNotification sends an MCP notification to the client.
// Called by Lua runtime when mcp.notify(method, params) is invoked.
func (s *Server) SendNotification(method string, params interface{}) {
	// Convert params to map[string]any for the MCP library
	var paramsMap map[string]any
	if params != nil {
		if m, ok := params.(map[string]interface{}); ok {
			paramsMap = make(map[string]any, len(m))
			for k, v := range m {
				paramsMap[k] = v
			}
		}
	}

	s.cfg.Log(2, "Sending notification: method=%s params=%v", method, paramsMap)
	s.mcpServer.SendNotificationToAllClients(method, paramsMap)
}

// handleVariables renders a page with a variable tree for the current session.
// Spec: mcp.md Section 2.2
func (s *Server) handleVariables(w http.ResponseWriter, r *http.Request) {
	sessionID := s.currentVendedID

	variables, err := s.getDebugVariables(sessionID)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html>
<head>
  <title>Debug: Variables</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.19.1/cdn/themes/light.css" />
  <script type="module" src="https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.19.1/cdn/shoelace-autoloader.js"></script>
  <style>
    body { font-family: system-ui, sans-serif; padding: 20px; max-width: 1200px; margin: 0 auto; }
    h1 { color: #333; }
    .error { color: red; padding: 10px; background: #fee; border-radius: 4px; }
    .tree-container { margin-top: 20px; }
    sl-tree { --indent-size: 20px; }
    sl-tree-item::part(item) { padding: 4px 0; }
    .var-id { color: #666; font-size: 0.9em; margin-right: 8px; }
    .var-type { color: #0066cc; font-weight: bold; margin-right: 8px; }
    .var-path { color: #666; font-style: italic; margin-right: 8px; }
    .var-value { color: #228b22; font-family: monospace; font-size: 0.9em; }
    .var-props { color: #888; font-size: 0.8em; margin-left: 16px; }
    .refresh-btn { margin-bottom: 16px; }
    pre { background: #f5f5f5; padding: 10px; border-radius: 4px; overflow-x: auto; }
  </style>
</head>
<body>
  <h1>Variable Tree - Session ` + sessionID + `</h1>
  <sl-button class="refresh-btn" onclick="location.reload()">
    <sl-icon slot="prefix" name="arrow-clockwise"></sl-icon>
    Refresh
  </sl-button>
`

	if err != nil {
		html += `<div class="error">Error: ` + err.Error() + `</div>`
	} else if len(variables) == 0 {
		html += `<div class="error">No variables found for session ` + sessionID + `</div>`
	} else {
		html += `<div class="tree-container"><sl-tree>`
		html += s.renderVariableTree(variables)
		html += `</sl-tree></div>`

		jsonBytes, _ := json.MarshalIndent(variables, "", "  ")
		html += `<h2>Raw JSON</h2><pre>` + escapeHTML(string(jsonBytes)) + `</pre>`
	}

	html += `</body></html>`
	w.Write([]byte(html))
}

// handleState renders a page with the session state JSON for the current session.
// Spec: mcp.md Section 2.2
func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	sessionID := s.currentVendedID

	stateData, err := s.getDebugState(sessionID)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html>
<head>
  <title>Debug: State</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.19.1/cdn/themes/light.css" />
  <script type="module" src="https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.19.1/cdn/shoelace-autoloader.js"></script>
  <style>
    body { font-family: system-ui, sans-serif; padding: 20px; max-width: 1200px; margin: 0 auto; }
    h1 { color: #333; }
    .error { color: red; padding: 10px; background: #fee; border-radius: 4px; }
    .refresh-btn { margin-bottom: 16px; }
    pre { background: #f5f5f5; padding: 16px; border-radius: 4px; overflow-x: auto; font-size: 14px; line-height: 1.5; }
  </style>
</head>
<body>
  <h1>Session State - Session ` + sessionID + `</h1>
  <sl-button class="refresh-btn" onclick="location.reload()">
    <sl-icon slot="prefix" name="arrow-clockwise"></sl-icon>
    Refresh
  </sl-button>
`

	if err != nil {
		html += `<div class="error">Error: ` + err.Error() + `</div>`
	} else if stateData == nil {
		html += `<div class="error">No state found for session ` + sessionID + `</div>`
	} else {
		jsonBytes, err := json.MarshalIndent(stateData, "", "  ")
		if err != nil {
			html += `<div class="error">Error formatting JSON: ` + err.Error() + `</div>`
		} else {
			html += `<h2>State JSON</h2><pre>` + escapeHTML(string(jsonBytes)) + `</pre>`
		}
	}

	html += `</body></html>`
	w.Write([]byte(html))
}

// getDebugState returns the state for a session (mcp.state or mcp.value).
func (s *Server) getDebugState(sessionID string) (interface{}, error) {
	session := s.UiServer.GetLuaSession(sessionID)
	if session == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Use SafeExecuteInSession to safely access the Lua state
	return s.SafeExecuteInSession(sessionID, func() (interface{}, error) {
		L := session.State
		mcpTable := L.GetGlobal("mcp")
		if mcpTable.Type() != lua.LTTable {
			return nil, fmt.Errorf("mcp global not found")
		}

		stateValue := L.GetField(mcpTable, "state")
		if stateValue == lua.LNil {
			stateValue = L.GetField(mcpTable, "value")
		}

		return cli.LuaToGo(stateValue), nil
	})
}

// renderVariableTree renders variables as nested sl-tree-item elements.
func (s *Server) renderVariableTree(variables []cli.DebugVariable) string {
	varMap := make(map[int64]cli.DebugVariable)
	for _, v := range variables {
		varMap[v.ID] = v
	}

	var roots []int64
	for _, v := range variables {
		if v.ParentID == 0 {
			roots = append(roots, v.ID)
		}
	}

	var result strings.Builder
	for _, rootID := range roots {
		s.renderVariableNode(&result, varMap, rootID)
	}
	return result.String()
}

// renderVariableNode renders a single variable and its children.
func (s *Server) renderVariableNode(sb *strings.Builder, varMap map[int64]cli.DebugVariable, varID int64) {
	v, ok := varMap[varID]
	if !ok {
		return
	}

	valueStr := ""
	if v.Value != nil {
		valueBytes, _ := json.Marshal(v.Value)
		valueStr = string(valueBytes)
		if len(valueStr) > 100 {
			valueStr = valueStr[:100] + "..."
		}
	}

	label := `<span class="var-id">#` + fmt.Sprintf("%d", v.ID) + `</span>`
	if v.Type != "" {
		label += `<span class="var-type">` + v.Type + `</span>`
	}
	if v.Path != "" {
		label += `<span class="var-path">` + v.Path + `</span>`
	}
	if valueStr != "" {
		label += `<span class="var-value">` + escapeHTML(valueStr) + `</span>`
	}

	hasChildren := len(v.ChildIDs) > 0

	if hasChildren {
		sb.WriteString(`<sl-tree-item expanded>`)
	} else {
		sb.WriteString(`<sl-tree-item>`)
	}
	sb.WriteString(label)

	if len(v.Properties) > 0 {
		sb.WriteString(`<div class="var-props">`)
		first := true
		for k, val := range v.Properties {
			if k == "type" || k == "path" {
				continue
			}
			if !first {
				sb.WriteString(", ")
			}
			sb.WriteString(k + "=" + val)
			first = false
		}
		sb.WriteString(`</div>`)
	}

	for _, childID := range v.ChildIDs {
		s.renderVariableNode(sb, varMap, childID)
	}

	sb.WriteString(`</sl-tree-item>`)
}

// parsePortFromURL extracts the port number from a URL like "http://localhost:8080".
func parsePortFromURL(url string) (int, error) {
	// Find the last colon (port separator)
	lastColon := strings.LastIndex(url, ":")
	if lastColon == -1 {
		return 0, fmt.Errorf("no port in URL: %s", url)
	}
	portStr := url[lastColon+1:]
	// Remove any path after the port
	if slashIdx := strings.Index(portStr, "/"); slashIdx >= 0 {
		portStr = portStr[:slashIdx]
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port in URL %s: %w", url, err)
	}
	return port, nil
}

// escapeHTML escapes special HTML characters.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// pushStateEvent adds an event to the queue and signals waiting clients.
// Called from Lua via mcp.pushState().
// Spec: mcp.md Section 8.1
func (s *Server) pushStateEvent(sessionID string, event interface{}) {
	s.stateWaitersMu.Lock()
	defer s.stateWaitersMu.Unlock()

	// Add to queue
	s.stateQueue[sessionID] = append(s.stateQueue[sessionID], event)

	// Signal all waiters for this session
	if waiters, ok := s.stateWaiters[sessionID]; ok {
		for _, ch := range waiters {
			select {
			case ch <- struct{}{}:
			default:
				// Channel already signaled or closed
			}
		}
		// Clear waiters - they'll re-register if they want to wait again
		delete(s.stateWaiters, sessionID)
	}
}

// drainStateQueue atomically returns and clears the event queue for a session.
// Triggers UI update so UIs monitoring the event queue refresh.
// Spec: mcp.md Section 8.2
func (s *Server) drainStateQueue(sessionID string) []interface{} {
	s.stateWaitersMu.Lock()
	defer s.stateWaitersMu.Unlock()

	events := s.stateQueue[sessionID]
	s.stateQueue[sessionID] = nil

	// Trigger UI update after draining (see mcp.md Section 4.1)
	if len(events) > 0 {
		s.SafeExecuteInSession(sessionID, func() (interface{}, error) { return nil, nil })
	}

	return events
}

// handleWait handles GET /wait - long-poll for state changes on the current session.
// Spec: mcp.md Section 8.2
// CRC: crc-MCPServer.md
func (s *Server) handleWait(w http.ResponseWriter, r *http.Request) {
	// Use the distinguished session (currentVendedID)
	sessionID := s.currentVendedID
	if sessionID == "" {
		http.Error(w, "No active session", http.StatusNotFound)
		return
	}

	// Check session exists
	session := s.UiServer.GetLuaSession(sessionID)
	if session == nil {
		http.NotFound(w, r)
		return
	}

	// Parse timeout (default 30s, max 120s)
	timeout := 30
	if t := r.URL.Query().Get("timeout"); t != "" {
		if parsed, err := strconv.Atoi(t); err == nil {
			timeout = parsed
		}
	}
	if timeout < 1 {
		timeout = 1
	}
	if timeout > 120 {
		timeout = 120
	}

	// Check if there are already events queued
	events := s.drainStateQueue(sessionID)
	if len(events) > 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
		return
	}

	// Create a channel for this waiter
	waiterCh := make(chan struct{}, 1)

	// Register the waiter
	s.stateWaitersMu.Lock()
	s.stateWaiters[sessionID] = append(s.stateWaiters[sessionID], waiterCh)
	s.stateWaitersMu.Unlock()

	// Wait for signal or timeout
	select {
	case <-waiterCh:
		// Signaled - drain and return events
		events := s.drainStateQueue(sessionID)
		if len(events) > 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(events)
			return
		}
		// No events (shouldn't happen, but handle gracefully)
		w.WriteHeader(http.StatusNoContent)

	case <-time.After(time.Duration(timeout) * time.Second):
		// Timeout - check one more time for events
		events := s.drainStateQueue(sessionID)
		if len(events) > 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(events)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case <-r.Context().Done():
		// Client disconnected
		return
	}

	// Unregister this waiter (cleanup)
	s.stateWaitersMu.Lock()
	if waiters, ok := s.stateWaiters[sessionID]; ok {
		for i, ch := range waiters {
			if ch == waiterCh {
				s.stateWaiters[sessionID] = append(waiters[:i], waiters[i+1:]...)
				break
			}
		}
	}
	s.stateWaitersMu.Unlock()
}
