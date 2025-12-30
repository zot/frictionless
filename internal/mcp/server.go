// Package mcp implements the Model Context Protocol server.
// CRC: crc-MCPServer.md
// Spec: mcp.md
// Sequence: seq-mcp-lifecycle.md, seq-mcp-notify.md
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
	uiServer          *cli.Server                    // UI engine server for ExecuteInSession
	runtime           *cli.LuaRuntime
	viewdefs          *cli.ViewdefManager
	promptManager     *PromptManager                 // Manages browser-based prompts
	startFunc         func(port int) (string, error) // Callback to start HTTP server
	onViewdefUploaded func(typeName string)          // Callback when a viewdef is uploaded
	getSessionCount   func() int                     // Callback to get active session count

	mu         sync.RWMutex
	state      State
	baseDir    string
	url        string
	httpServer *http.Server // HTTP server for /api/prompt (in stdio mode)
	mcpPort    int          // Port for MCP HTTP server (written to .ui-mcp/mcp-port)
}

// NewServer creates a new MCP server.
func NewServer(cfg *cli.Config, uiServer *cli.Server, runtime *cli.LuaRuntime, viewdefs *cli.ViewdefManager, startFunc func(port int) (string, error), onViewdefUploaded func(typeName string), getSessionCount func() int) *Server {
	s := server.NewMCPServer("ui-server", "0.1.0")
	srv := &Server{
		mcpServer:         s,
		cfg:               cfg,
		uiServer:          uiServer,
		runtime:           runtime,
		viewdefs:          viewdefs,
		promptManager:     NewPromptManager(uiServer, runtime),
		startFunc:         startFunc,
		onViewdefUploaded: onViewdefUploaded,
		getSessionCount:   getSessionCount,
		state:             Unconfigured,
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
	return s.uiServer.ExecuteInSession(sessionID, fn)
}

// ServeStdio starts the MCP server on Stdin/Stdout.
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcpServer)
}

// ServeSSE starts the MCP server as an SSE HTTP server on the given address.
// It also adds the /api/prompt endpoint for browser-based permission prompts.
// Spec: prompt-ui.md
func (s *Server) ServeSSE(addr string) error {
	sseServer := server.NewSSEServer(s.mcpServer)

	// Wrap SSE server with our custom handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/api/prompt", s.handlePrompt)
	mux.HandleFunc("/debug/variables", s.handleDebugVariables)
	mux.HandleFunc("/debug/state", s.handleDebugState)
	mux.Handle("/", sseServer) // SSE server handles MCP traffic

	s.cfg.Log(0, "Starting MCP SSE server on %s (with /api/prompt, /debug/*)", addr)

	// Start HTTP server
	httpServer := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return httpServer.ListenAndServe()
}

// StartPromptServer starts a standalone HTTP server for /api/prompt in stdio mode.
// This allows hook scripts to call the prompt API while MCP runs over stdio.
// Also serves debug pages at /debug/variables and /debug/state.
// Returns the port number.
// Spec: prompt-ui.md
func (s *Server) StartPromptServer() (int, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/prompt", s.handlePrompt)
	mux.HandleFunc("/debug/variables", s.handleDebugVariables)
	mux.HandleFunc("/debug/state", s.handleDebugState)

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
			s.cfg.Log(0, "Prompt server error: %v", err)
		}
	}()

	s.cfg.Log(0, "Prompt server listening on port %d", port)
	return port, nil
}

// WritePortFile writes the MCP port to the mcp-port file in baseDir.
// Spec: prompt-ui.md
func (s *Server) WritePortFile(port int) error {
	s.mu.RLock()
	baseDir := s.baseDir
	s.mu.RUnlock()

	if baseDir == "" {
		return fmt.Errorf("server not configured")
	}

	portFile := filepath.Join(baseDir, "mcp-port")
	return os.WriteFile(portFile, []byte(strconv.Itoa(port)), 0644)
}

// RemovePortFile removes the mcp-port file.
func (s *Server) RemovePortFile() {
	s.mu.RLock()
	baseDir := s.baseDir
	s.mu.RUnlock()

	if baseDir != "" {
		os.Remove(filepath.Join(baseDir, "mcp-port"))
	}
}

// PromptRequest is the JSON request body for POST /api/prompt.
// Spec: prompt-ui.md
type PromptRequest struct {
	Message   string         `json:"message"`
	Options   []PromptOption `json:"options"`
	SessionID string         `json:"sessionId,omitempty"`
	Timeout   int            `json:"timeout,omitempty"` // seconds
}

// PromptResponseBody is the JSON response for POST /api/prompt.
type PromptResponseBody struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// handlePrompt handles POST /api/prompt requests from hook scripts.
// Spec: prompt-ui.md
// CRC: crc-PromptManager.md
func (s *Server) handlePrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Defaults
	if req.SessionID == "" {
		req.SessionID = "1"
	}
	if req.Timeout <= 0 {
		req.Timeout = 60
	}

	s.cfg.Log(1, "Prompt request: session=%s message=%q options=%d timeout=%ds",
		req.SessionID, req.Message, len(req.Options), req.Timeout)

	if s.promptManager == nil {
		http.Error(w, "Prompt manager not available", http.StatusServiceUnavailable)
		return
	}

	// Call prompt manager (blocks until user responds or timeout)
	timeout := time.Duration(req.Timeout) * time.Second
	resp, err := s.promptManager.Prompt(req.SessionID, req.Message, req.Options, timeout)
	if err != nil {
		s.cfg.Log(1, "Prompt error: %v", err)
		http.Error(w, err.Error(), http.StatusRequestTimeout)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PromptResponseBody{
		Value: resp.Value,
		Label: resp.Label,
	})
}

// ShutdownPromptServer shuts down the standalone prompt HTTP server.
func (s *Server) ShutdownPromptServer(ctx context.Context) error {
	if s.httpServer != nil {
		s.RemovePortFile()
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

// SendNotification sends an MCP notification to the client.
// Called by Lua runtime when mcp.notify(method, params) is invoked.
// Sequence: seq-mcp-notify.md
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

// handleDebugVariables renders a debug page with a variable tree.
func (s *Server) handleDebugVariables(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		sessionID = "1"
	}

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

// handleDebugState renders a debug page with the session state JSON.
func (s *Server) handleDebugState(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		sessionID = "1"
	}

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
	// Use SafeExecuteInSession to safely access the Lua state
	return s.SafeExecuteInSession(sessionID, func() (interface{}, error) {
		L := s.runtime.State
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

// escapeHTML escapes special HTML characters.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
