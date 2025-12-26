// Package main provides the MCP-enabled UI server.
// This extends remote-ui with MCP support via the Hooks interface.
package main

import (
	"os"

	"github.com/zot/ui-engine/cli"
)

func main() {
	hooks := &cli.Hooks{
		BeforeDispatch: func(command string, args []string) (bool, int) {
			// Handle mcp command
			if command == "mcp" {
				return true, runMCP(args)
			}
			return false, 0
		},
		CustomHelp: func() string {
			return `
MCP Commands:
  mcp             Start MCP server on Stdio (for AI integration)

MCP Examples:
  ui-mcp mcp      Start MCP server for Claude/AI integration`
		},
		CustomVersion: func() string {
			return "MCP Extension v0.1.0"
		},
	}

	os.Exit(cli.RunWithHooks(os.Args[1:], hooks))
}

// runMCP runs the MCP server on Stdio.
// TODO: Wire up the full MCP server implementation.
func runMCP(args []string) int {
	// For now, just print that MCP is not yet fully integrated
	// The full implementation will:
	// 1. Load config from args
	// 2. Create the Lua runtime
	// 3. Create the MCP server
	// 4. Start on Stdio
	println("MCP server starting on Stdio...")
	println("(Full implementation in progress)")
	return 0
}
