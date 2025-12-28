// Package main provides the MCP-enabled UI server.
// This extends remote-ui with MCP support via the Hooks interface.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zot/ui-engine/cli"
	"github.com/zot/ui-mcp/internal/mcp"
)

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

MCP Examples:
  ui-mcp mcp --dir .ui-mcp                          Start MCP server with working directory
  ui-mcp serve --port 8000 --mcp-port 8001 --dir .  Start standalone with UI on 8000, MCP on 8001`
		},
		CustomVersion: func() string {
			return "MCP Extension v0.1.0"
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

	// Ensure logs go to Stderr to keep Stdout clean for MCP protocol
	log.SetOutput(os.Stderr)

	// Create the ui-engine server
	srv := cli.NewServer(cfg)

	// Start cleanup worker
	srv.StartCleanupWorker(time.Hour)

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
		os.Exit(0)
	}()

	// Create the MCP server with callbacks into the ui-engine server
	mcpServer := mcp.NewServer(
		cfg,
		srv,
		srv.GetLuaRuntime(),
		srv.GetViewdefManager(),
		func(port int) (string, error) {
			// Start HTTP server async and return URL
			return srv.StartAsync(port)
		},
		func(typeName string) {
			// Called when a viewdef is uploaded - could trigger refresh
			cfg.Log(2, "Viewdef uploaded for type: %s", typeName)
		},
		func() int {
			// Return count of active browser sessions
			return srv.GetSessions().Count()
		},
	)

	// Wire up Lua mcp.notify() to send MCP notifications
	srv.GetLuaRuntime().SetNotificationHandler(mcpServer.SendNotification)

	// Serve MCP on Stdio (blocks until done)
	if err := mcpServer.ServeStdio(); err != nil {
		log.Printf("MCP server error: %v", err)
		return 1
	}

	return 0
}

// runServe runs the standalone server with HTTP UI and SSE MCP endpoints.
func runServe(args []string) int {
	// Parse serve-specific flags
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8000, "Port for HTTP UI server")
	mcpPort := fs.Int("mcp-port", 8001, "Port for MCP SSE server")
	dir := fs.String("dir", ".", "Working directory for ui-mcp")
	verbose := fs.Int("v", 0, "Verbosity level (0-4)")

	if err := fs.Parse(args); err != nil {
		log.Printf("Failed to parse flags: %v", err)
		return 1
	}

	// Build config args for cli.Load
	configArgs := []string{
		"--dir", *dir,
		"--port", fmt.Sprintf("%d", *port),
	}
	for i := 0; i < *verbose; i++ {
		configArgs = append(configArgs, "-v")
	}

	cfg, err := cli.Load(configArgs)
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		return 1
	}

	// Create the ui-engine server
	srv := cli.NewServer(cfg)
	srv.StartCleanupWorker(time.Hour)

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
		os.Exit(0)
	}()

	// Create MCP server
	mcpServer := mcp.NewServer(
		cfg,
		srv,
		srv.GetLuaRuntime(),
		srv.GetViewdefManager(),
		func(p int) (string, error) {
			// In serve mode, UI server is already started on fixed port
			return fmt.Sprintf("http://127.0.0.1:%d", *port), nil
		},
		func(typeName string) {
			cfg.Log(2, "Viewdef uploaded for type: %s", typeName)
		},
		func() int {
			return srv.GetSessions().Count()
		},
	)

	// Wire up Lua notifications
	srv.GetLuaRuntime().SetNotificationHandler(mcpServer.SendNotification)

	// Auto-configure MCP server with base dir
	if err := mcpServer.Configure(*dir); err != nil {
		log.Printf("Failed to configure MCP server: %v", err)
		return 1
	}

	// Start UI HTTP server
	url, err := srv.StartAsync(*port)
	if err != nil {
		log.Printf("Failed to start UI server: %v", err)
		return 1
	}
	log.Printf("UI server running at %s", url)

	// Start MCP SSE server (blocks)
	mcpAddr := fmt.Sprintf(":%d", *mcpPort)
	if err := mcpServer.ServeSSE(mcpAddr); err != nil {
		log.Printf("MCP SSE server error: %v", err)
		return 1
	}

	return 0
}
