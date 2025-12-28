// Package mcp implements the Model Context Protocol server.
// CRC: crc-MCPTool.md
// Spec: mcp.md
// Sequence: seq-mcp-lifecycle.md, seq-mcp-run.md
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/zot/ui-engine/cli"
)

func (s *Server) registerTools() {
	// ui_configure
	s.mcpServer.AddTool(mcp.NewTool("ui_configure",
		mcp.WithDescription("Prepare the server environment and file system. Must be the first tool called."),
		mcp.WithString("base_dir", mcp.Required(), mcp.Description("Absolute path to the project root directory")),
	), s.handleConfigure)

	// ui_start
	s.mcpServer.AddTool(mcp.NewTool("ui_start",
		mcp.WithDescription("Start the embedded HTTP UI server. Requires server to be Configured."),
	), s.handleStart)

	// ui_open_browser
	s.mcpServer.AddTool(mcp.NewTool("ui_open_browser",
		mcp.WithDescription("Open the system's default web browser to the UI session."),
		mcp.WithString("sessionId", mcp.Description("The session to open (defaults to '1')")),
		mcp.WithString("path", mcp.Description("The URL path to open (defaults to '/')")),
		mcp.WithBoolean("conserve", mcp.Description("Use conserve mode to prevent duplicate tabs (defaults to true)")),
	), s.handleOpenBrowser)

	// ui_run
	s.mcpServer.AddTool(mcp.NewTool("ui_run",
		mcp.WithDescription("Execute Lua code in a session context"),
		mcp.WithString("code", mcp.Required(), mcp.Description("Lua code to execute")),
		mcp.WithString("sessionId", mcp.Description("The session ID to run in (defaults to '1')")),
	), s.handleRun)

	// ui_upload_viewdef
	s.mcpServer.AddTool(mcp.NewTool("ui_upload_viewdef",
		mcp.WithDescription("Upload a dynamic view definition"),
		mcp.WithString("type", mcp.Required(), mcp.Description("Presenter type (e.g. 'MyPresenter')")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace (e.g. 'DEFAULT')")),
		mcp.WithString("content", mcp.Required(), mcp.Description("HTML content")),
	), s.handleUploadViewdef)

	// ui_status
	s.mcpServer.AddTool(mcp.NewTool("ui_status",
		mcp.WithDescription("Get current server status including lifecycle state and browser connection count"),
	), s.handleStatus)
}

func (s *Server) handleConfigure(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Spec: mcp.md
	// CRC: crc-MCPTool.md
	// Sequence: seq-mcp-lifecycle.md
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("arguments must be a map"), nil
	}

	baseDir, ok := args["base_dir"].(string)
	if !ok {
		return mcp.NewToolResultError("base_dir must be a string"), nil
	}

	// 1. Directory Creation
	if err := os.MkdirAll(filepath.Join(baseDir, "log"), 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create directories: %v", err)), nil
	}

	// 2. Runtime Setup (Lua I/O Redirection)
	logPath := filepath.Join(baseDir, "log", "lua.log")
	errPath := filepath.Join(baseDir, "log", "lua-err.log")
	if err := s.runtime.RedirectOutput(logPath, errPath); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to redirect Lua output: %v", err)), nil
	}

	// 3. State Transition
	if err := s.Configure(baseDir); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 4. Resource Extraction (Optional - only if resources dir is missing)
	resourcesDir := filepath.Join(baseDir, "resources")
	if _, err := os.Stat(resourcesDir); os.IsNotExist(err) {
		// Try to extract only the resources directory from bundle
		if isBundled, _ := cli.IsBundled(); isBundled {
			// List files in resources/ from bundle
			files, _ := cli.BundleListFiles("resources")
			if len(files) > 0 {
				os.MkdirAll(resourcesDir, 0755)
				for _, f := range files {
					content, _ := cli.BundleReadFile(f)
					os.WriteFile(filepath.Join(baseDir, f), content, 0644)
				}
			}
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("Server configured. Log files created at %s", filepath.Join(baseDir, "log"))), nil
}

func (s *Server) handleStart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Spec: mcp.md
	// CRC: crc-MCPTool.md
	// Sequence: seq-mcp-lifecycle.md
	url, err := s.Start()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(url), nil
}

func (s *Server) handleOpenBrowser(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Spec: mcp.md
	// CRC: crc-MCPTool.md
	// Sequence: seq-mcp-lifecycle.md
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("arguments must be a map"), nil
	}

	sessionID, ok := args["sessionId"].(string)
	if !ok || sessionID == "" {
		sessionID = "1"
	}

	path, ok := args["path"].(string)
	if !ok {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	conserve := true
	if c, ok := args["conserve"].(bool); ok {
		conserve = c
	}

	s.mu.RLock()
	baseURL := s.url
	state := s.state
	s.mu.RUnlock()

	if state != Running {
		return mcp.NewToolResultError("Server not running"), nil
	}

	// Construct URL: baseURL + "/" + sessionID + path
	fullURL := fmt.Sprintf("%s/%s%s", baseURL, sessionID, path)
	if conserve {
		if strings.Contains(fullURL, "?") {
			fullURL += "&conserve=true"
		} else {
			fullURL += "?conserve=true"
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", fullURL)
	case "darwin":
		cmd = exec.Command("open", fullURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", fullURL)
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unsupported platform: %s", runtime.GOOS)), nil
	}

	if err := cmd.Start(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open browser: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Opened %s", fullURL)), nil
}

func (s *Server) handleGetState(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Removed in favor of ui_run
	return mcp.NewToolResultError("Tool removed. Use ui_run to inspect state."), nil
}

func (s *Server) handleRun(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Spec: mcp.md
	// CRC: crc-MCPTool.md
	// Sequence: seq-mcp-run.md
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("arguments must be a map"), nil
	}

	code, ok := args["code"].(string)
	if !ok {
		return mcp.NewToolResultError("code must be a string"), nil
	}
	sessionID, ok := args["sessionId"].(string)
	if !ok || sessionID == "" {
		sessionID = "1"
	}

	// Use Server.ExecuteInSession (triggers afterBatch) wrapping Runtime.ExecuteInSession (sets Lua context)
	result, err := s.uiServer.ExecuteInSession(sessionID, func() (interface{}, error) {
		return s.runtime.ExecuteInSession(sessionID, func() (interface{}, error) {
			return s.runtime.LoadCodeDirect("mcp-run", code)
		})
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("execution failed: %v", err)), nil
	}

	// Marshal result
	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		// Fallback for non-serializable results
		fallback := map[string]string{
			"non-json": fmt.Sprintf("%v", result),
		}
		jsonResult, _ = json.Marshal(fallback)
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleUploadViewdef(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Spec: mcp.md
	// CRC: crc-MCPTool.md
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("arguments must be a map"), nil
	}

	typeName, ok := args["type"].(string)
	if !ok {
		return mcp.NewToolResultError("type must be a string"), nil
	}
	namespace, ok := args["namespace"].(string)
	if !ok {
		return mcp.NewToolResultError("namespace must be a string"), nil
	}
	content, ok := args["content"].(string)
	if !ok {
		return mcp.NewToolResultError("content must be a string"), nil
	}

	key := fmt.Sprintf("%s.%s", typeName, namespace)
	s.viewdefs.AddViewdef(key, content)

	// Notify server to refresh variables of this type
	if s.onViewdefUploaded != nil {
		s.onViewdefUploaded(typeName)
	}
	
	return mcp.NewToolResultText(fmt.Sprintf("Viewdef %s uploaded", key)), nil
}

// CRC: crc-MCPTool.md
// Spec: mcp.md (section 5.6)
func (s *Server) handleStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.mu.RLock()
	state := s.state
	url := s.url
	s.mu.RUnlock()

	result := map[string]interface{}{
		"state": stateToString(state),
	}

	if state == Running {
		result["url"] = url
		if s.getSessionCount != nil {
			result["sessions"] = s.getSessionCount()
		}
	}

	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal status: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}

func stateToString(state State) string {
	switch state {
	case Unconfigured:
		return "unconfigured"
	case Configured:
		return "configured"
	case Running:
		return "running"
	default:
		return "unknown"
	}
}
