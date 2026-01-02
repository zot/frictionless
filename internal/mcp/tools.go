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
	lua "github.com/yuin/gopher-lua"
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
		mcp.WithString("sessionId", mcp.Description("The vended session ID to open (defaults to '1')")),
		mcp.WithString("path", mcp.Description("The URL path to open (defaults to '/')")),
		mcp.WithBoolean("conserve", mcp.Description("Use conserve mode to prevent duplicate tabs (defaults to true)")),
	), s.handleOpenBrowser)

	// ui_run
	s.mcpServer.AddTool(mcp.NewTool("ui_run",
		mcp.WithDescription("Execute Lua code in a session context"),
		mcp.WithString("code", mcp.Required(), mcp.Description("Lua code to execute")),
		mcp.WithString("sessionId", mcp.Description("The vended session ID to run in (defaults to '1')")),
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

// Spec: mcp.md
// CRC: crc-MCPTool.md
// Sequence: seq-mcp-lifecycle.md
func (s *Server) handleConfigure(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("arguments must be a map"), nil
	}

	baseDir, ok := args["base_dir"].(string)
	if !ok {
		return mcp.NewToolResultError("base_dir must be a string"), nil
	}

	// 0. Stop current session if running (allows reconfiguration)
	if err := s.Stop(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to stop current session: %v", err)), nil
	}

	// 1. Directory Creation
	if err := os.MkdirAll(filepath.Join(baseDir, "log"), 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create directories: %v", err)), nil
	}

	// 2. Store log paths for session setup (applied when session is created)
	s.logPath = filepath.Join(baseDir, "log", "lua.log")
	s.errPath = filepath.Join(baseDir, "log", "lua-err.log")

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

	// 5. Agent File Installation (Spec: mcp.md section 5.1.1)
	// Sequence: seq-mcp-lifecycle.md (Scenario 1a)
	s.installAgentFiles(baseDir)

	return mcp.NewToolResultText(fmt.Sprintf("Server configured. Log files created at %s", filepath.Join(baseDir, "log"))), nil
}

// installAgentFiles checks for and installs bundled agent files.
// Spec: mcp.md section 5.1.1
// Sequence: seq-mcp-lifecycle.md (Scenario 1a)
func (s *Server) installAgentFiles(baseDir string) {
	// Bundled agent files to install
	agentFiles := []string{"ui-builder.md"}

	// Project root is parent of base_dir (e.g., .ui-mcp -> project root)
	projectRoot := filepath.Dir(baseDir)
	agentsDir := filepath.Join(projectRoot, ".claude", "agents")

	isBundled, _ := cli.IsBundled()

	for _, agentFile := range agentFiles {
		destPath := filepath.Join(agentsDir, agentFile)

		// Skip if file already exists
		if _, err := os.Stat(destPath); err == nil {
			continue
		}

		// Try to read from bundle
		bundlePath := filepath.Join("agents", agentFile)
		var content []byte
		var err error

		if isBundled {
			content, err = cli.BundleReadFile(bundlePath)
		} else {
			// Fallback: read from local agents/ directory (development mode)
			localPath := filepath.Join(filepath.Dir(baseDir), "agents", agentFile)
			content, err = os.ReadFile(localPath)
		}

		if err != nil || len(content) == 0 {
			s.cfg.Log(1, "Agent file not found in bundle: %s", bundlePath)
			continue
		}

		// Create directory if needed
		if err := os.MkdirAll(agentsDir, 0755); err != nil {
			s.cfg.Log(0, "Failed to create agents directory: %v", err)
			continue
		}

		// Write agent file
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			s.cfg.Log(0, "Failed to write agent file: %v", err)
			continue
		}

		s.cfg.Log(1, "Installed agent file: %s", destPath)

		// Send notification to AI agent (if MCP server is available)
		if s.mcpServer != nil {
			s.SendNotification("agent_installed", map[string]interface{}{
				"file": agentFile,
				"path": filepath.Join(".claude", "agents", agentFile),
			})
		}
	}
}

// Spec: mcp.md
// CRC: crc-MCPTool.md
// Sequence: seq-mcp-lifecycle.md
func (s *Server) handleStart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	baseURL, err := s.Start()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Create session - this triggers CreateLuaBackendForSession
	// Returns (session, vendedID, error) - session.ID has the UUID for URLs
	session, vendedID, err := s.UiServer.GetSessions().CreateSession()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create session: %v", err)), nil
	}

	// Store the vended ID for later cleanup
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
		return mcp.NewToolResultError(fmt.Sprintf("failed to setup mcp global: %v", err)), nil
	}

	// Return URL with session UUID so browser connects to the right session
	sessionURL := fmt.Sprintf("%s/%s", baseURL, session.ID)
	return mcp.NewToolResultText(sessionURL), nil
}

// setupMCPGlobal creates the mcp global object in Lua with Go functions attached.
func (s *Server) setupMCPGlobal(vendedID string) error {
	session := s.UiServer.GetLuaSession(vendedID)
	if session == nil {
		return fmt.Errorf("session %s not found", vendedID)
	}

	// Use SafeExecuteInSession to ensure proper executor context and session global
	_, err := s.SafeExecuteInSession(vendedID, func() (interface{}, error) {
		L := session.State

		// Create mcp table
		mcpTable := L.NewTable()
		L.SetGlobal("mcp", mcpTable)

		// Set type for viewdef resolution
		L.SetField(mcpTable, "type", lua.LString("MCP"))

		// Set value to nil initially
		L.SetField(mcpTable, "value", lua.LNil)

		// mcp.pushState(event) - push event to queue and signal waiters
		// Spec: mcp.md Section 8.1
		L.SetField(mcpTable, "pushState", L.NewFunction(func(L *lua.LState) int {
			event := L.CheckTable(1)

			// Convert Lua table to Go value
			goEvent := luaTableToGo(event)

			// Add to queue and signal waiters
			s.pushStateEvent(vendedID, goEvent)
			return 0
		}))

		// Register as app variable - this creates variable 1 in the tracker
		code := "session:createAppVariable(mcp)"
		if err := L.DoString(code); err != nil {
			return nil, fmt.Errorf("failed to create app variable: %w", err)
		}

		return nil, nil
	})
	return err
}

// luaTableToGo converts a Lua table to a Go map/slice.
func luaTableToGo(tbl *lua.LTable) interface{} {
	// Check if it's an array (sequential integer keys starting at 1)
	isArray := true
	maxIdx := 0
	tbl.ForEach(func(k, v lua.LValue) {
		if kn, ok := k.(lua.LNumber); ok {
			idx := int(kn)
			if idx > maxIdx {
				maxIdx = idx
			}
		} else {
			isArray = false
		}
	})

	if isArray && maxIdx > 0 {
		arr := make([]interface{}, maxIdx)
		tbl.ForEach(func(k, v lua.LValue) {
			if kn, ok := k.(lua.LNumber); ok {
				arr[int(kn)-1] = luaValueToGo(v)
			}
		})
		return arr
	}

	// It's a map
	m := make(map[string]interface{})
	tbl.ForEach(func(k, v lua.LValue) {
		if ks, ok := k.(lua.LString); ok {
			m[string(ks)] = luaValueToGo(v)
		}
	})
	return m
}

// luaValueToGo converts a Lua value to Go.
func luaValueToGo(v lua.LValue) interface{} {
	switch val := v.(type) {
	case lua.LString:
		return string(val)
	case lua.LNumber:
		return float64(val)
	case lua.LBool:
		return bool(val)
	case *lua.LTable:
		return luaTableToGo(val)
	case *lua.LNilType:
		return nil
	default:
		return nil
	}
}

// Spec: mcp.md
// CRC: crc-MCPTool.md
// Sequence: seq-mcp-lifecycle.md
func (s *Server) handleOpenBrowser(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("arguments must be a map"), nil
	}

	sessionID, ok := args["sessionId"].(string)
	if !ok || sessionID == "" {
		sessionID = s.currentVendedID
	}
	if sessionID == "" {
		return mcp.NewToolResultError("no active session - call ui_start first"), nil
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

	// Convert vended ID to internal session ID for URL
	// URLs should use internal session IDs, not vended IDs
	internalID := s.UiServer.GetSessions().GetInternalID(sessionID)
	if internalID == "" {
		return mcp.NewToolResultError(fmt.Sprintf("session %s not found", sessionID)), nil
	}

	// Construct URL: baseURL + "/" + internalSessionID + path
	fullURL := fmt.Sprintf("%s/%s%s", baseURL, internalID, path)
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

// Spec: mcp.md
// CRC: crc-MCPTool.md
// Sequence: seq-mcp-run.md
func (s *Server) handleRun(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
		sessionID = s.currentVendedID
	}
	if sessionID == "" {
		return mcp.NewToolResultError("no active session - call ui_start first"), nil
	}

	// Get the session for LoadCodeDirect
	session := s.UiServer.GetLuaSession(sessionID)
	if session == nil {
		return mcp.NewToolResultError(fmt.Sprintf("session %s not found", sessionID)), nil
	}

	// Use SafeExecuteInSession (sets Lua context, triggers afterBatch, recovers panics)
	result, err := s.SafeExecuteInSession(sessionID, func() (interface{}, error) {
		return session.LoadCodeDirect("mcp-run", code)
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

// Spec: mcp.md
// CRC: crc-MCPTool.md
func (s *Server) handleUploadViewdef(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
