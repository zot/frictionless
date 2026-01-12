// Package main provides the MCP-enabled UI server.
// This extends remote-ui with MCP support via the Hooks interface.
// Spec: mcp.md
// Sequence: seq-mcp-lifecycle.md
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/zot/ui-engine/cli"
	"github.com/zot/ui-mcp/internal/mcp"
)

// Version is set at build time via ldflags
// Spec: mcp.md Section 1.2
var Version = "dev"

func main() {
	hooks := &cli.Hooks{
		BeforeDispatch: func(command string, args []string) (bool, int) {
			// Handle mcp command
			if command == "mcp" {
				return true, runMCP(args)
			}
			// Handle serve command (standalone mode with HTTP MCP)
			if command == "serve" {
				return true, runServe(args)
			}
			return false, 0
		},
		CustomHelp: func() string {
			return `
MCP Commands:
  mcp             Start MCP server on Stdio (for AI integration)
  serve           Start standalone server with HTTP UI and MCP endpoints
  hooks           Manage Claude Code permission hooks

Hook Commands:
  hooks install   Install permission UI hook for Claude Code
  hooks uninstall Remove permission UI hook
  hooks status    Check hook installation status

MCP Examples:
  ui-mcp mcp                                        Start MCP server (default: --dir .claude/ui)
  ui-mcp serve --port 8000 --mcp-port 8001          Start standalone with UI on 8000, MCP on 8001
  ui-mcp hooks install                              Install permission UI hook`
		},
		CustomVersion: func() string {
			return "ui-mcp " + Version
		},
	}

	os.Exit(cli.RunWithHooks(os.Args[1:], hooks))
}

// runMCP runs the MCP server on Stdio.
func runMCP(args []string) int {
	// Load config using the same parser as serve command
	cfg, err := cli.Load(args)
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		return 1
	}

	// Enable hot-loading by default in MCP mode
	// Spec: mcp.md Section 4.0
	cfg.Lua.Hotload = true

	// Default dir to .claude/ui if not specified
	// Spec: mcp.md Section 2.2
	if cfg.Server.Dir == "" {
		cfg.Server.Dir = ".claude/ui"
	}

	logToFile := false
	// Redirect stderr to a log file for debugging
	if cfg.Server.Dir != "" {
		logDir := filepath.Join(cfg.Server.Dir, "log")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Printf("Warning: failed to create log directory: %v", err)
		} else {
			logPath := filepath.Join(logDir, "mcp.log")
			logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				log.Printf("Warning: failed to open mcp.log: %v", err)
			} else {
				logToFile = true
				os.Stderr = logFile
				log.SetOutput(logFile)
			}
		}
	}

	if !logToFile {
		// Ensure logs go to Stderr to keep Stdout clean for MCP protocol
		log.SetOutput(os.Stderr)
	}

	var mcpServer *mcp.Server
	mcpServer = newMCPServer(cfg, func(port int) (string, error) {
		// Start HTTP server async and return URL
		return mcpServer.UiServer.StartAsync(port)
	})
	if mcpServer == nil {
		return 1
	}
	// Start HTTP server for debug endpoints and state wait
	// Spec: mcp.md Section 2.2
	httpPort, err := mcpServer.StartHTTPServer()
	if err != nil {
		log.Printf("Failed to start HTTP server: %v", err)
		return 1
	}

	// Write MCP port file so scripts can find the HTTP API
	// Spec: mcp.md Section 5.2
	if cfg.Server.Dir != "" {
		if err := mcpServer.WriteMCPPortFile(httpPort); err != nil {
			log.Printf("Warning: failed to write mcp-port file: %v", err)
		}
	}

	// Serve MCP on Stdio (blocks until done)
	if err := mcpServer.ServeStdio(); err != nil {
		log.Printf("MCP server error: %v", err)
		return 1
	}
	return 0
}

// Create the MCP server with callbacks into the ui-engine server
func newMCPServer(cfg *cli.Config, fn func(p int) (string, error)) *mcp.Server {
	// Create the ui-engine server
	srv := cli.NewServer(cfg)

	// Start cleanup worker
	srv.StartCleanupWorker(time.Hour)

	mcpServer := mcp.NewServer(
		cfg,
		srv,
		srv.GetViewdefManager(),
		fn,
		func(typeName string) {
			// Called when a viewdef is uploaded - could trigger refresh
			cfg.Log(2, "Viewdef uploaded for type: %s", typeName)
		},
		func() int {
			// Return count of active browser sessions
			return srv.GetSessions().Count()
		},
	)
	if cfg.Server.Dir != "" {
		if err := mcpServer.Configure(cfg.Server.Dir); err != nil {
			log.Printf("Failed to configure MCP server: %v", err)
			return nil
		}
	}
	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		//mcpServer.ShutdownPromptServer(ctx)
		srv.Shutdown(ctx)
		os.Exit(0)
	}()
	return mcpServer
}

// runServe runs the standalone server with HTTP UI and SSE MCP endpoints.
func runServe(args []string) int {
	// Extract --mcp-port from args (not part of standard cli.Load flags)
	mcpPort := 8001
	var filteredArgs []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--mcp-port" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%d", &mcpPort)
			i++ // skip the value
		} else if strings.HasPrefix(args[i], "--mcp-port=") {
			fmt.Sscanf(args[i], "--mcp-port=%d", &mcpPort)
		} else {
			filteredArgs = append(filteredArgs, args[i])
		}
	}

	// Use cli.Load for standard flags (handles -vvvv expansion)
	cfg, err := cli.Load(filteredArgs)
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		return 1
	}

	// Enable hot-loading by default in serve mode
	// Spec: mcp.md Section 4.0
	cfg.Lua.Hotload = true

	// Default dir to .claude/ui if not specified
	// Spec: mcp.md Section 2.3
	if cfg.Server.Dir == "" {
		cfg.Server.Dir = ".claude/ui"
	}
	// Default port to 8000 if not specified
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8000
	}

	mcpServer := newMCPServer(cfg, func(p int) (string, error) {
		// In serve mode, UI server is already started on fixed port
		return fmt.Sprintf("http://127.0.0.1:%d", cfg.Server.Port), nil
	})
	if mcpServer == nil {
		return 1
	}

	// Start UI HTTP server
	url, err := mcpServer.UiServer.StartAsync(cfg.Server.Port)
	if err != nil {
		log.Printf("Failed to start UI server: %v", err)
		return 1
	}
	log.Printf("UI server running at %s", url)

	// Start MCP SSE server (blocks)
	mcpAddr := fmt.Sprintf(":%d", mcpPort)
	if err := mcpServer.ServeSSE(mcpAddr); err != nil {
		log.Printf("MCP SSE server error: %v", err)
		return 1
	}
	return 0
}
