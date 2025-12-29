// Package main provides the MCP-enabled UI server.
// This extends remote-ui with MCP support via the Hooks interface.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
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
			// Handle hooks command for Claude Code integration
			if command == "hooks" {
				return true, runHooks(args)
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
  ui-mcp mcp --dir .ui-mcp                          Start MCP server with working directory
  ui-mcp serve --port 8000 --mcp-port 8001 --dir .  Start standalone with UI on 8000, MCP on 8001
  ui-mcp hooks install                              Install permission UI hook`
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

	// Configure MCP server with baseDir (needed for port file)
	if cfg.Server.Dir != "" {
		if err := mcpServer.Configure(cfg.Server.Dir); err != nil {
			log.Printf("Failed to configure MCP server: %v", err)
			return 1
		}
	}

	// Start prompt server for /api/prompt endpoint (used by permission hooks)
	// Spec: prompt-ui.md
	promptPort, err := mcpServer.StartPromptServer()
	if err != nil {
		log.Printf("Failed to start prompt server: %v", err)
		return 1
	}

	// Write port file so hooks can find the prompt API
	if cfg.Server.Dir != "" {
		if err := mcpServer.WritePortFile(promptPort); err != nil {
			log.Printf("Warning: failed to write port file: %v", err)
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
		mcpServer.ShutdownPromptServer(ctx)
		srv.Shutdown(ctx)
		os.Exit(0)
	}()

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

// runHooks manages Claude Code permission hooks.
// Spec: prompt-ui.md
// CRC: crc-HookCLI.md
func runHooks(args []string) int {
	if len(args) == 0 {
		fmt.Println("Usage: ui-mcp hooks <install|uninstall|status>")
		return 1
	}

	switch args[0] {
	case "install":
		return installHook()
	case "uninstall":
		return uninstallHook()
	case "status":
		return hookStatus()
	default:
		fmt.Printf("Unknown hooks subcommand: %s\n", args[0])
		fmt.Println("Usage: ui-mcp hooks <install|uninstall|status>")
		return 1
	}
}

// Hook script content
const hookScript = `#!/bin/bash
# Permission UI hook for Claude Code
# Generated by ui-mcp hooks install
# Spec: prompt-ui.md
set -e

# Read hook input from stdin
input=$(cat)

# Extract details from hook data
tool_name=$(echo "$input" | jq -r '.tool_name // "unknown tool"')
tool_input=$(echo "$input" | jq -c '.tool_input // {}')
message="Claude wants to use: $tool_name"

# Read MCP port
MCP_PORT=$(cat .ui-mcp/mcp-port 2>/dev/null || echo "")
if [ -z "$MCP_PORT" ]; then
  # UI MCP not running, fall back to terminal
  exit 0
fi

# Build prompt request
request=$(jq -n \
  --arg msg "$message" \
  '{
    "message": $msg,
    "options": [
      {"label": "Allow once", "value": "allow"},
      {"label": "Always allow", "value": "allow_session"},
      {"label": "Deny", "value": "deny"}
    ],
    "timeout": 60
  }')

# Call prompt API
response=$(curl -s -X POST "http://127.0.0.1:$MCP_PORT/api/prompt" \
  -H "Content-Type: application/json" \
  -d "$request" 2>/dev/null || echo '{"error": "failed"}')

# Check for error
if echo "$response" | jq -e '.error' >/dev/null 2>&1; then
  # API call failed, fall back to terminal
  exit 0
fi

# Parse response
value=$(echo "$response" | jq -r '.value')
label=$(echo "$response" | jq -r '.label')

# Log decision for pattern analysis
log_entry=$(jq -n \
  --arg tool "$tool_name" \
  --argjson input "$tool_input" \
  --arg choice "$value" \
  --arg ts "$(date -Iseconds)" \
  '{tool: $tool, input: $input, choice: $choice, timestamp: $ts}')
echo "$log_entry" >> .ui-mcp/permissions.log

# Return hook decision
case "$value" in
  "allow"|"allow_session")
    echo '{"decision": "allow"}'
    ;;
  "deny")
    echo '{"decision": "deny", "message": "User denied permission"}'
    ;;
  *)
    # Timeout or error - let terminal prompt handle it
    exit 0
    ;;
esac
`

func installHook() int {
	// Create .claude/hooks directory
	hooksDir := ".claude/hooks"
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		log.Printf("Failed to create hooks directory: %v", err)
		return 1
	}

	// Write hook script
	hookPath := filepath.Join(hooksDir, "permission-ui.sh")
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		log.Printf("Failed to write hook script: %v", err)
		return 1
	}

	// Update settings.json
	settingsPath := ".claude/settings.json"
	settings := make(map[string]interface{})

	// Read existing settings if present
	if data, err := os.ReadFile(settingsPath); err == nil {
		json.Unmarshal(data, &settings)
	}

	// Add PermissionRequest hook (new format with matchers)
	// See: https://code.claude.com/docs/en/hooks
	hooksConfig, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		hooksConfig = make(map[string]interface{})
	}

	hooksConfig["PermissionRequest"] = []map[string]interface{}{
		{
			"matcher": map[string]interface{}{}, // Match all permission requests
			"hooks": []map[string]interface{}{
				{
					"type":    "command",
					"command": ".claude/hooks/permission-ui.sh",
					"timeout": 120,
				},
			},
		},
	}
	settings["hooks"] = hooksConfig

	// Write updated settings
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal settings: %v", err)
		return 1
	}

	// Ensure .claude directory exists
	if err := os.MkdirAll(".claude", 0755); err != nil {
		log.Printf("Failed to create .claude directory: %v", err)
		return 1
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		log.Printf("Failed to write settings: %v", err)
		return 1
	}

	fmt.Println("Permission UI hook installed successfully!")
	fmt.Println("")
	fmt.Println("Hook script: .claude/hooks/permission-ui.sh")
	fmt.Println("Settings updated: .claude/settings.json")
	fmt.Println("")
	fmt.Println("Make sure ui-mcp is running with --dir .ui-mcp when using Claude Code.")
	return 0
}

func uninstallHook() int {
	// Update settings.json to remove PermissionRequest hook
	settingsPath := ".claude/settings.json"
	settings := make(map[string]interface{})

	// Read existing settings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		fmt.Println("No settings.json found - hook not installed")
		return 0
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		log.Printf("Failed to parse settings: %v", err)
		return 1
	}

	// Remove PermissionRequest hook
	if hooksConfig, ok := settings["hooks"].(map[string]interface{}); ok {
		delete(hooksConfig, "PermissionRequest")
		if len(hooksConfig) == 0 {
			delete(settings, "hooks")
		}
	}

	// Write updated settings
	data, err = json.MarshalIndent(settings, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal settings: %v", err)
		return 1
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		log.Printf("Failed to write settings: %v", err)
		return 1
	}

	fmt.Println("Permission UI hook uninstalled.")
	fmt.Println("Hook script left at .claude/hooks/permission-ui.sh (use --delete-script to remove)")
	return 0
}

func hookStatus() int {
	fmt.Println("Permission UI Hook Status")
	fmt.Println("==========================")

	// Check settings.json
	settingsPath := ".claude/settings.json"
	hookInstalled := false

	if data, err := os.ReadFile(settingsPath); err == nil {
		var settings map[string]interface{}
		if json.Unmarshal(data, &settings) == nil {
			if hooksConfig, ok := settings["hooks"].(map[string]interface{}); ok {
				if _, ok := hooksConfig["PermissionRequest"]; ok {
					hookInstalled = true
				}
			}
		}
	}

	if hookInstalled {
		fmt.Println("Hook in settings.json: INSTALLED")
	} else {
		fmt.Println("Hook in settings.json: NOT INSTALLED")
	}

	// Check hook script
	hookPath := ".claude/hooks/permission-ui.sh"
	if _, err := os.Stat(hookPath); err == nil {
		fmt.Println("Hook script exists:    YES")
	} else {
		fmt.Println("Hook script exists:    NO")
	}

	// Check if ui-mcp is running (port file exists)
	portFile := ".ui-mcp/mcp-port"
	if data, err := os.ReadFile(portFile); err == nil {
		fmt.Printf("UI MCP running:        YES (port %s)\n", string(data))
	} else {
		fmt.Println("UI MCP running:        NO")
	}

	return 0
}
