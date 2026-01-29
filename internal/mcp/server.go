// Package mcp implements the Model Context Protocol server.
// CRC: crc-MCPServer.md
// Spec: mcp.md
// Sequence: seq-mcp-lifecycle.md, seq-mcp-state-wait.md
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
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

// State represents the internal state of the MCP server (not exposed externally).
// Spec: mcp.md Section 3.1 - Server auto-starts.
type State int

const (
	Configured State = iota // Internal state during configuration (not exposed)
	Running                 // Server is running and accepting connections
)

// Server implements an MCP server for AI integration.
type Server struct {
	mcpServer         *server.MCPServer
	cfg               *cli.Config
	UiServer          *cli.Server // UI engine server for ExecuteInSession
	viewdefs          *cli.ViewdefManager
	startFunc       func(port int) (string, error) // Callback to start HTTP server
	getSessionCount func() int                     // Callback to get active session count
	onClearLogs       func()                         // Callback to reopen Go log file after clearing logs

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

	// Wait time tracking (Spec: mcp.md Section 8.3)
	waitStartTime time.Time // When agent last responded (updated on /wait return)
}

// NewServer creates a new MCP server.
func NewServer(cfg *cli.Config, uiServer *cli.Server, viewdefs *cli.ViewdefManager, startFunc func(port int) (string, error), getSessionCount func() int) *Server {
	s := server.NewMCPServer("ui-server", "0.1.0")
	srv := &Server{
		mcpServer:       s,
		cfg:             cfg,
		UiServer:        uiServer,
		viewdefs:        viewdefs,
		startFunc:       startFunc,
		getSessionCount: getSessionCount,
		state:           Configured, // Initial internal state before ui_configure is called
		stateWaiters:    make(map[string][]chan struct{}),
		stateQueue:      make(map[string][]interface{}),
		waitStartTime:   time.Now(), // Spec: mcp.md Section 8.3
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

	// Tool API endpoints (Spec 2.5)
	mux.HandleFunc("/api/ui_status", s.handleAPIStatus)
	mux.HandleFunc("/api/ui_run", s.handleAPIRun)
	mux.HandleFunc("/api/ui_display", s.handleAPIDisplay)
	mux.HandleFunc("/api/ui_configure", s.handleAPIConfigure)
	mux.HandleFunc("/api/ui_install", s.handleAPIInstall)
	mux.HandleFunc("/api/ui_open_browser", s.handleAPIOpenBrowser)
	mux.HandleFunc("/api/ui_audit", s.handleAPIAudit)
	mux.HandleFunc("/api/ui_theme", s.handleAPITheme)
	mux.HandleFunc("/api/resource/", s.handleAPIResource)

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

	s.cfg.Log(0, "HTTP server listening on port %d (/variables, /state, /wait, /api/*)", port)

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

// Configure prepares the server environment (directories, auto-install).
// Called by handleConfigure after Stop() to allow reconfiguration.
// Auto-installs if README.md is missing (Spec: mcp.md Section 3.1).
// CRC: crc-MCPServer.md
// Sequence: seq-mcp-lifecycle.md (Scenario 1)
func (s *Server) Configure(baseDir string) error {
	s.mu.Lock()
	s.baseDir = baseDir
	s.state = Configured // Temporary state during configuration
	s.mu.Unlock()        // Release lock before I/O operations

	// Create base directory and log directory
	if err := os.MkdirAll(filepath.Join(baseDir, "log"), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Clear existing log files and reopen Go log handles
	// Spec: mcp.md Section 5.1 - ui_configure clears logs
	if err := s.ClearLogs(); err != nil {
		s.cfg.Log(1, "Warning: failed to clear logs: %v", err)
	}

	// Store log paths for session setup
	s.logPath = filepath.Join(baseDir, "log", "lua.log")
	s.errPath = filepath.Join(baseDir, "log", "lua-err.log")

	// Auto-install if README.md is missing
	// Spec: mcp.md Section 3.1 - Startup Behavior
	readmePath := filepath.Join(baseDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		s.cfg.Log(1, "README.md not found, running auto-install")
		if _, installErr := s.Install(false); installErr != nil {
			return fmt.Errorf("auto-install failed: %w", installErr)
		}
	}

	return nil
}

// SetBaseDir sets the base directory without running auto-install.
// Used by the install command which handles installation separately.
func (s *Server) SetBaseDir(baseDir string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.baseDir = baseDir
}

// SetOnClearLogs sets a callback to be called after logs are cleared.
// Used by main.go to reopen the Go log file handle.
// CRC: crc-MCPServer.md
func (s *Server) SetOnClearLogs(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onClearLogs = fn
}

// ClearLogs clears all log files in the log directory.
// After clearing, calls the onClearLogs callback to allow reopening Go log file handles.
// Spec: mcp.md Section 5.1 - ui_configure clears logs
// CRC: crc-MCPServer.md
func (s *Server) ClearLogs() error {
	s.mu.RLock()
	baseDir := s.baseDir
	callback := s.onClearLogs
	s.mu.RUnlock()

	if baseDir == "" {
		return nil // No base directory configured yet
	}

	logDir := filepath.Join(baseDir, "log")
	entries, err := os.ReadDir(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Log directory doesn't exist, nothing to clear
		}
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	// Delete all files in the log directory
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}
		path := filepath.Join(logDir, entry.Name())
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			s.cfg.Log(1, "Warning: failed to remove log file %s: %v", path, err)
		}
	}

	// Call callback to reopen Go log file handles
	if callback != nil {
		callback()
	}

	return nil
}

// StartAndCreateSession starts the UI server and creates a session with mcp global.
// This is called both on process startup (auto-start) and by ui_configure (reconfiguration).
// Spec: mcp.md Section 3.1 - Server auto-starts
// Sequence: seq-mcp-lifecycle.md (Scenario 1)
func (s *Server) StartAndCreateSession() (string, error) {
	// Start the UI HTTP server
	baseURL, err := s.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start server: %w", err)
	}

	// Extract UI port from URL and write ui-port file
	uiPort, err := parsePortFromURL(baseURL)
	if err != nil {
		s.cfg.Log(0, "Warning: failed to parse UI port from URL %s: %v", baseURL, err)
	} else {
		if err := s.WriteUIPortFile(uiPort); err != nil {
			s.cfg.Log(0, "Warning: failed to write ui-port file: %v", err)
		}
	}

	// Create session - this triggers CreateLuaBackendForSession
	// Returns (session, vendedID, error)
	_, vendedID, err := s.UiServer.GetSessions().CreateSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	// Store the vended ID for later use
	s.currentVendedID = vendedID

	// Apply Lua I/O redirection to the new session (if paths were set at configure time)
	if s.logPath != "" && s.errPath != "" {
		luaSession := s.UiServer.GetLuaSession(vendedID)
		if luaSession != nil {
			if err := luaSession.RedirectOutput(s.logPath, s.errPath); err != nil {
				s.cfg.Log(0, "Warning: failed to redirect Lua output: %v", err)
			}
		}
	}

	// Set up mcp global in Lua with Go functions
	if err := s.setupMCPGlobal(vendedID); err != nil {
		return "", fmt.Errorf("failed to setup mcp global: %w", err)
	}

	// Return base URL without session ID
	// Spec: mcp.md Section 5.1 - url is http://HOST:PORT (no session ID)
	// Browser uses cookie-based session binding (Section 3.3)
	return baseURL, nil
}

// GetCurrentSessionID returns the internal session ID for the current MCP session.
// Used by the root session provider to serve "/" without creating a new session.
// Spec: mcp.md Section 3.3 - Root URL Session Binding
func (s *Server) GetCurrentSessionID() string {
	s.mu.RLock()
	vendedID := s.currentVendedID
	s.mu.RUnlock()

	if vendedID == "" {
		return ""
	}

	return s.UiServer.GetSessions().GetInternalID(vendedID)
}

// Start transitions the server to the Running state and starts the HTTP server.
// Called by handleConfigure after Configure() completes.
// Spec: mcp.md
// CRC: crc-MCPServer.md
func (s *Server) Start() (string, error) {
	s.mu.Lock()
	if s.state == Running {
		s.mu.Unlock()
		return "", fmt.Errorf("Server already running")
	}
	s.mu.Unlock() // Release before calling startFunc to avoid deadlock

	// Select random port (0)
	url, err := s.startFunc(0)
	if err != nil {
		return "", err
	}

	// Update state after successful start
	s.mu.Lock()
	s.state = Running
	s.url = url
	s.mu.Unlock()

	return url, nil
}

// Stop destroys the current session and resets state.
// This allows reconfiguration via ui_configure.
func (s *Server) Stop() error {
	s.mu.Lock()
	if s.state != Running {
		s.mu.Unlock()
		return nil // Nothing to stop
	}

	// Capture session info while holding lock
	vendedID := s.currentVendedID
	s.mu.Unlock() // Release before calling DestroySession to avoid deadlock

	// Destroy the session outside the lock (may trigger callbacks)
	if vendedID != "" {
		sessions := s.UiServer.GetSessions()
		internalID := sessions.GetInternalID(vendedID)
		if internalID != "" {
			sessions.DestroySession(internalID)
		}
	}

	// Update state after destruction completes
	s.mu.Lock()
	s.currentVendedID = ""
	s.state = Configured
	s.url = ""
	s.mu.Unlock()

	return nil
}

// SendNotification sends an MCP notification to the client.
// Called by Lua runtime when mcp.notify(method, params) is invoked.
func (s *Server) SendNotification(method string, params interface{}) {
	// Convert params to map[string]any for the MCP library
	paramsMap, _ := params.(map[string]interface{})
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
func parsePortFromURL(urlStr string) (int, error) {
	lastColon := strings.LastIndex(urlStr, ":")
	if lastColon == -1 {
		return 0, fmt.Errorf("no port in URL: %s", urlStr)
	}

	portStr := urlStr[lastColon+1:]
	if slashIdx := strings.Index(portStr, "/"); slashIdx >= 0 {
		portStr = portStr[:slashIdx]
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port in URL %s: %w", urlStr, err)
	}
	return port, nil
}

// escapeHTML escapes special HTML characters using the standard library.
func escapeHTML(s string) string {
	return html.EscapeString(s)
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
	events := s.stateQueue[sessionID]
	s.stateQueue[sessionID] = nil
	s.stateWaitersMu.Unlock() // Release before calling into ui-engine to avoid deadlock

	// Trigger UI update after draining (see mcp.md Section 4.1)
	if len(events) > 0 {
		s.SafeExecuteInSession(sessionID, func() (interface{}, error) { return nil, nil })
	}

	return events
}

// hasPollingClients returns true if there are clients waiting on the /wait endpoint.
// Spec: mcp.md Section 8.2
func (s *Server) hasPollingClients(sessionID string) bool {
	s.stateWaitersMu.Lock()
	defer s.stateWaitersMu.Unlock()

	return len(s.stateWaiters[sessionID]) > 0
}

// getWaitTime returns seconds since agent last responded, or 0 if currently connected.
// Spec: mcp.md Section 8.3
func (s *Server) getWaitTime(sessionID string) float64 {
	if s.hasPollingClients(sessionID) {
		return 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Since(s.waitStartTime).Seconds()
}

// writeEventsJSON writes events as JSON. Returns true if events were written.
func writeEventsJSON(w http.ResponseWriter, events []interface{}) bool {
	if len(events) == 0 {
		return false
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
	return true
}

// respondWithEvents drains the queue, updates waitStartTime, and writes response.
// Used by handleWait to consolidate the response logic for signal and timeout cases.
func (s *Server) respondWithEvents(w http.ResponseWriter, sessionID string) {
	s.mu.Lock()
	s.waitStartTime = time.Now()
	s.mu.Unlock()

	if !writeEventsJSON(w, s.drainStateQueue(sessionID)) {
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleWait handles GET /wait - long-poll for state changes on the current session.
// Spec: mcp.md Section 8.3
// CRC: crc-MCPServer.md
func (s *Server) handleWait(w http.ResponseWriter, r *http.Request) {
	sessionID := s.currentVendedID
	if sessionID == "" {
		http.Error(w, "No active session", http.StatusNotFound)
		return
	}

	if s.UiServer.GetLuaSession(sessionID) == nil {
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
	} else if timeout > 120 {
		timeout = 120
	}

	// Check if there are already events queued
	if writeEventsJSON(w, s.drainStateQueue(sessionID)) {
		s.mu.Lock()
		s.waitStartTime = time.Now()
		s.mu.Unlock()
		return
	}

	// Create and register a channel for this waiter
	waiterCh := make(chan struct{}, 1)
	s.stateWaitersMu.Lock()
	s.stateWaiters[sessionID] = append(s.stateWaiters[sessionID], waiterCh)
	s.stateWaitersMu.Unlock()

	// Ensure cleanup on all exit paths
	// Seq: seq-mcp-state-wait.md Scenario 7
	defer func() {
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

		// Trigger UI refresh so pollingEvents() status updates after disconnect
		s.SafeExecuteInSession(sessionID, func() (interface{}, error) { return nil, nil })
	}()

	// Trigger UI refresh so pollingEvents() status updates
	s.SafeExecuteInSession(sessionID, func() (interface{}, error) { return nil, nil })
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.SafeExecuteInSession(sessionID, func() (interface{}, error) { return nil, nil })
	}()

	// Wait for signal or timeout
	select {
	case <-waiterCh:
		s.respondWithEvents(w, sessionID)
	case <-time.After(time.Duration(timeout) * time.Second):
		s.respondWithEvents(w, sessionID)
	case <-r.Context().Done():
		// Client disconnected - update waitStartTime so waitTime() resets
		s.mu.Lock()
		s.waitStartTime = time.Now()
		s.mu.Unlock()
	}
}
