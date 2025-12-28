// Package mcp implements the Model Context Protocol server.
// CRC: crc-MCPServer.md
// Spec: mcp.md
// Sequence: seq-mcp-lifecycle.md, seq-mcp-notify.md
package mcp

import (
	"fmt"
	"sync"

	"github.com/mark3labs/mcp-go/server"
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
	startFunc         func(port int) (string, error) // Callback to start HTTP server
	onViewdefUploaded func(typeName string)          // Callback when a viewdef is uploaded
	getSessionCount   func() int                     // Callback to get active session count

	mu      sync.RWMutex
	state   State
	baseDir string
	url     string
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
		startFunc:         startFunc,
		onViewdefUploaded: onViewdefUploaded,
		getSessionCount:   getSessionCount,
		state:             Unconfigured,
	}
	srv.registerTools()
	srv.registerResources()
	return srv
}

// ServeStdio starts the MCP server on Stdin/Stdout.
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcpServer)
}

// ServeSSE starts the MCP server as an SSE HTTP server on the given address.
func (s *Server) ServeSSE(addr string) error {
	sseServer := server.NewSSEServer(s.mcpServer)
	s.cfg.Log(0, "Starting MCP SSE server on %s", addr)
	return sseServer.Start(addr)
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
