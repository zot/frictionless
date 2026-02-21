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
	"sync"
	"syscall"
	"time"

	"github.com/zot/frictionless/internal/mcp"
	"github.com/zot/ui-engine/cli"
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
			// Handle install command
			// Spec: mcp.md Section 2.4
			if command == "install" {
				return true, runInstall(args)
			}
			// Handle theme command (file-based, no server needed)
			if command == "theme" {
				return true, runTheme(args)
			}
			return false, 0
		},
		CustomHelp: func() string {
			return `
Commands:
  mcp             Start MCP server on Stdio (for AI integration)
  serve           Start standalone server with HTTP UI and MCP endpoints
  install         Install skills and resources (without starting server)
  theme           Theme management (list, classes, audit)

Examples:
  frictionless mcp                                        Start MCP server (default: --dir .ui)
  frictionless serve --port 8000 --mcp-port 8001          Start standalone with UI on 8000, MCP on 8001
  frictionless install                                    Install skills and resources
  frictionless install --force                            Force reinstall even if up to date
  frictionless theme list                                 List available themes
  frictionless theme classes [THEME]                      Show semantic classes for a theme
  frictionless theme audit APP [THEME]                    Audit app's theme class usage`
		},
		CustomVersion: func() string {
			return "frictionless " + Version
		},
	}

	os.Exit(cli.RunWithHooks(os.Args[1:], hooks))
}

// runMCP runs the MCP server on Stdio.
func runMCP(args []string) int {
	os.Setenv("FRICTIONLESS_MCP", "true")
	// Load config using the same parser as serve command
	cfg, err := cli.Load(args)
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		return 1
	}

	// Enable hot-loading by default in MCP mode
	// Spec: mcp.md Section 4.0
	cfg.Lua.Hotload = true

	// Default dir to .ui if not specified
	// Spec: mcp.md Section 2.2
	if cfg.Server.Dir == "" {
		cfg.Server.Dir = ".ui"
	}

	// Track the current log file for reopening after log clearing
	var currentLogFile *os.File
	var logFileMu sync.Mutex

	// openLogFile opens (or reopens) the Go log file
	// Spec: mcp.md Section 5.1 - reopening Go log handles after clearing
	openLogFile := func() {
		logFileMu.Lock()
		defer logFileMu.Unlock()

		// Close existing file if any
		if currentLogFile != nil {
			currentLogFile.Close()
			currentLogFile = nil
		}

		logDir := filepath.Join(cfg.Server.Dir, "log")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return
		}
		logPath := filepath.Join(logDir, "mcp.log")
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return
		}
		currentLogFile = logFile
		os.Stderr = logFile
		log.SetOutput(logFile)
	}

	// Initial log file setup
	logToFile := false
	if cfg.Server.Dir != "" {
		openLogFile()
		logFileMu.Lock()
		logToFile = currentLogFile != nil
		logFileMu.Unlock()
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

	// Set callback to reopen Go log file after logs are cleared
	// Spec: mcp.md Section 5.1 - ui_configure clears logs
	if logToFile {
		mcpServer.SetOnClearLogs(openLogFile)
	}

	// Configure AFTER SetOnClearLogs so log file can be reopened after ClearLogs()
	// Spec: mcp.md Section 3.1 - Server auto-starts
	if cfg.Server.Dir != "" {
		if err := mcpServer.Configure(cfg.Server.Dir); err != nil {
			log.Printf("Failed to configure MCP server: %v", err)
			return 1
		}
	}

	// Auto-start: create session and start server
	// Must be called AFTER newMCPServer returns so mcpServer is assigned
	// Spec: mcp.md Section 3.1 - Server auto-starts
	// Sequence: seq-mcp-lifecycle.md (Scenario 1)
	if cfg.Server.Dir != "" {
		if _, err := mcpServer.StartAndCreateSession(); err != nil {
			log.Printf("Failed to start MCP server: %v", err)
			return 1
		}
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
		func() int {
			// Return count of active browser sessions
			return srv.GetSessions().Count()
		},
	)

	// Set root session provider so "/" uses the MCP session instead of creating a new one
	// Spec: mcp.md Section 3.3 - Root URL Session Binding
	srv.SetRootSessionProvider(func() string {
		return mcpServer.GetCurrentSessionID()
	})

	// Note: Configure() is called by runMCP AFTER SetOnClearLogs is set,
	// so the log file can be reopened after ClearLogs() deletes it.
	// StartAndCreateSession is also called AFTER newMCPServer returns
	// to avoid nil pointer in startFunc closure.

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
	return mcpServer
}

// runInstall installs skills and resources without starting the MCP server.
// Spec: mcp.md Section 2.4
func runInstall(args []string) int {
	// Parse --force flag
	force := false
	var filteredArgs []string
	for _, arg := range args {
		if arg == "--force" || arg == "-f" {
			force = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	// Load config for --dir flag
	cfg, err := cli.Load(filteredArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		return 1
	}

	// Default dir to .ui if not specified
	if cfg.Server.Dir == "" {
		cfg.Server.Dir = ".ui"
	}

	// Create a minimal MCP server just for installation
	srv := cli.NewServer(cfg)
	mcpServer := mcp.NewServer(
		cfg,
		srv,
		srv.GetViewdefManager(),
		nil, // No start function needed
		nil, // No session count needed
	)

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(cfg.Server.Dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create directory %s: %v\n", cfg.Server.Dir, err)
		return 1
	}

	// Set baseDir (needed for Install)
	mcpServer.SetBaseDir(cfg.Server.Dir)

	// Run installation
	result, err := mcpServer.Install(force)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Install failed: %v\n", err)
		return 1
	}

	// Print results
	if result.VersionSkipped {
		fmt.Printf("Skipped: installed version %s >= bundled version %s\n", result.InstalledVersion, result.BundledVersion)
		fmt.Println("Use --force to reinstall")
		return 0
	}

	if len(result.Installed) > 0 {
		fmt.Println("Installed:")
		for _, f := range result.Installed {
			fmt.Printf("  %s\n", f)
		}
	}
	if len(result.Skipped) > 0 {
		fmt.Println("Skipped (already exist):")
		for _, f := range result.Skipped {
			fmt.Printf("  %s\n", f)
		}
	}

	return 0
}

func printThemeClasses(classes []mcp.ThemeClass) {
	fmt.Println("Classes:")
	for _, c := range classes {
		fmt.Printf("  .%s\n", c.Name)
		fmt.Printf("    %s\n", c.Description)
		fmt.Printf("    Usage: %s\n", c.Usage)
		if len(c.Elements) > 0 {
			fmt.Printf("    Elements: %s\n", strings.Join(c.Elements, ", "))
		}
		fmt.Println()
	}
}

// runTheme handles theme management commands (file-based, no server needed).
func runTheme(args []string) int {
	// Default dir to .ui
	baseDir := ".ui"

	// Parse --dir flag
	var filteredArgs []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--dir" && i+1 < len(args) {
			baseDir = args[i+1]
			i++ // skip the value
		} else if strings.HasPrefix(args[i], "--dir=") {
			baseDir = strings.TrimPrefix(args[i], "--dir=")
		} else {
			filteredArgs = append(filteredArgs, args[i])
		}
	}

	if len(filteredArgs) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: frictionless theme <list|classes|audit> [options]")
		fmt.Fprintln(os.Stderr, "  theme list              List available themes")
		fmt.Fprintln(os.Stderr, "  theme classes [THEME]   Show semantic classes for a theme")
		fmt.Fprintln(os.Stderr, "  theme audit APP [THEME] Audit app's theme class usage")
		return 1
	}

	action := filteredArgs[0]
	actionArgs := filteredArgs[1:]

	switch action {
	case "list":
		themes, err := mcp.ListThemes(baseDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing themes: %v\n", err)
			return 1
		}
		current := mcp.GetCurrentTheme(baseDir)
		fmt.Printf("Themes: %s\n", strings.Join(themes, ", "))
		if current != "" {
			fmt.Printf("Current: %s\n", current)
		}
		return 0

	case "classes":
		theme := ""
		if len(actionArgs) > 0 {
			theme = actionArgs[0]
		}
		themeName, classes, err := mcp.ResolveThemeClasses(baseDir, theme)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting theme classes: %v\n", err)
			return 1
		}
		fmt.Printf("Theme: %s\n", themeName)
		if theme != "" {
			fm, _ := mcp.GetThemeClasses(baseDir, theme)
			if fm != nil && fm.Description != "" {
				fmt.Printf("Description: %s\n", fm.Description)
			}
		}
		fmt.Println()
		printThemeClasses(classes)
		return 0

	case "audit":
		if len(actionArgs) == 0 {
			fmt.Fprintln(os.Stderr, "Usage: frictionless theme audit APP [THEME]")
			return 1
		}
		app := actionArgs[0]
		theme := ""
		if len(actionArgs) > 1 {
			theme = actionArgs[1]
		}
		result, err := mcp.AuditAppTheme(baseDir, app, theme)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error auditing theme: %v\n", err)
			return 1
		}

		fmt.Printf("App: %s\n", result.App)
		fmt.Printf("Theme: %s\n", result.Theme)
		fmt.Printf("\nSummary: %d classes total, %d documented, %d undocumented\n",
			result.Summary.Total, result.Summary.Documented, result.Summary.Undocumented)

		if len(result.UndocumentedClasses) > 0 {
			fmt.Println("\nUndocumented classes:")
			for _, c := range result.UndocumentedClasses {
				fmt.Printf("  .%s (%s:%d)\n", c.Class, c.File, c.Line)
			}
		}

		if len(result.UnusedThemeClasses) > 0 {
			fmt.Println("\nUnused theme classes (not used by this app):")
			for _, c := range result.UnusedThemeClasses {
				fmt.Printf("  .%s\n", c)
			}
		}

		return 0

	default:
		fmt.Fprintf(os.Stderr, "Unknown theme action: %s\n", action)
		fmt.Fprintln(os.Stderr, "Usage: frictionless theme <list|classes|audit> [options]")
		return 1
	}
}

// runServe runs the standalone server with HTTP UI and SSE MCP endpoints.
func runServe(args []string) int {
	os.Setenv("FRICTIONLESS_MCP", "true")
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

	// Default dir to .ui if not specified
	// Spec: mcp.md Section 2.3
	if cfg.Server.Dir == "" {
		cfg.Server.Dir = ".ui"
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

	// Configure before starting (serve mode doesn't redirect Go logs)
	if cfg.Server.Dir != "" {
		if err := mcpServer.Configure(cfg.Server.Dir); err != nil {
			log.Printf("Failed to configure MCP server: %v", err)
			return 1
		}
	}

	// Start UI HTTP server
	url, err := mcpServer.UiServer.StartAsync(cfg.Server.Port)
	if err != nil {
		log.Printf("Failed to start UI server: %v", err)
		return 1
	}
	log.Printf("UI server running at %s", url)

	// Auto-start: create session and start server
	// Spec: mcp.md Section 3.1 - Server auto-starts
	// Sequence: seq-mcp-lifecycle.md (Scenario 1)
	if cfg.Server.Dir != "" {
		if _, err := mcpServer.StartAndCreateSession(); err != nil {
			log.Printf("Failed to start MCP server: %v", err)
			return 1
		}
	}

	// Start MCP SSE server (blocks)
	mcpAddr := fmt.Sprintf(":%d", mcpPort)
	if err := mcpServer.ServeSSE(mcpAddr); err != nil {
		log.Printf("MCP SSE server error: %v", err)
		return 1
	}
	return 0
}
