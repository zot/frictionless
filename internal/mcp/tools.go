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

	// ui_install
	// Spec: mcp.md section 5.7
	s.mcpServer.AddTool(mcp.NewTool("ui_install",
		mcp.WithDescription("Install bundled configuration files (skill files). Call after ui_configure when install_needed is true."),
		mcp.WithBoolean("force", mcp.Description("If true, overwrites existing files. Defaults to false.")),
	), s.handleInstall)

	// ui_display
	s.mcpServer.AddTool(mcp.NewTool("ui_display",
		mcp.WithDescription("Load and display an app by name. Loads from apps/{name}/app.lua if not already loaded."),
		mcp.WithString("name", mcp.Required(), mcp.Description("App name (e.g., 'claude-panel')")),
		mcp.WithString("sessionId", mcp.Description("Session ID (defaults to current session)")),
	), s.handleDisplay)
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

	// 5. Check if installation is needed (Spec: mcp.md section 5.1.1)
	// Sequence: seq-mcp-lifecycle.md (Scenario 1a)
	// Project root is the grandparent of baseDir (e.g., .claude/ui -> .)
	projectRoot := filepath.Dir(filepath.Dir(baseDir))
	skillFile := filepath.Join(projectRoot, ".claude", "skills", "ui-builder", "SKILL.md")
	installNeeded := false
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		installNeeded = true
	}

	// Return structured response with install_needed hint
	response := map[string]interface{}{
		"status":   "configured",
		"log_path": filepath.Join(baseDir, "log"),
	}
	if installNeeded {
		response["install_needed"] = true
		response["hint"] = "Run ui_install to install skill files"
	}

	responseJSON, _ := json.Marshal(response)
	return mcp.NewToolResultText(string(responseJSON)), nil
}

// handleInstall installs bundled configuration files.
// Spec: mcp.md section 5.7
func (s *Server) handleInstall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check state
	if s.state != Configured && s.state != Running {
		return mcp.NewToolResultError("ui_install requires CONFIGURED or RUNNING state. Call ui_configure first."), nil
	}

	// Parse force parameter
	force := false
	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		if f, ok := args["force"].(bool); ok {
			force = f
		}
	}

	// Project root is the grandparent of baseDir (e.g., .claude/ui -> .)
	// Spec: mcp.md section 5.7
	projectRoot := filepath.Dir(filepath.Dir(s.baseDir))
	installed := []string{}
	skipped := []string{}
	appended := []string{}

	// 1. Install skill files
	// Skills are installed from install/init/skills/ (dev) or bundled skills/ directory
	isBundled, _ := cli.IsBundled()
	skillsDir := filepath.Join(projectRoot, ".claude", "skills")

	// Skills to install: map of skill name to files
	skillsToInstall := map[string][]string{
		"ui": {"SKILL.md"},
		"ui-builder": {"SKILL.md", "examples/requirements.md", "examples/design.md", "examples/code.lua",
			"examples/ContactApp.DEFAULT.html", "examples/Contact.list-item.html", "examples/ChatMessage.list-item.html"},
	}

	for skillName, skillFiles := range skillsToInstall {
		for _, skillFile := range skillFiles {
			destPath := filepath.Join(skillsDir, skillName, skillFile)
			relPath := filepath.Join(".claude", "skills", skillName, skillFile)

			// Check if file exists
			exists := false
			if _, err := os.Stat(destPath); err == nil {
				exists = true
			}

			// Skip if exists and not forcing
			if exists && !force {
				skipped = append(skipped, relPath)
				continue
			}

			// Read from bundle or local
			bundlePath := filepath.Join("skills", skillName, skillFile)
			var content []byte
			var err error

			if isBundled {
				content, err = cli.BundleReadFile(bundlePath)
			} else {
				// Development mode: read from install/init/skills/
				localPath := filepath.Join(projectRoot, "install", "init", "skills", skillName, skillFile)
				content, err = os.ReadFile(localPath)
			}

			if err != nil || len(content) == 0 {
				s.cfg.Log(1, "Skill file not found: %s", bundlePath)
				continue
			}

			// Create directory and write file
			destDir := filepath.Dir(destPath)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to create skills directory: %v", err)), nil
			}
			if err := os.WriteFile(destPath, content, 0644); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to write %s: %v", skillFile, err)), nil
			}

			installed = append(installed, relPath)
			s.cfg.Log(1, "Installed skill file: %s", destPath)
		}
	}

	// Return result
	result := map[string]interface{}{
		"installed": installed,
		"skipped":   skipped,
		"appended":  appended,
	}
	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// Spec: mcp.md
// CRC: crc-MCPTool.md
// Sequence: seq-mcp-lifecycle.md
func (s *Server) handleStart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	baseURL, err := s.Start()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Extract UI port from URL and write ui-port file
	// Spec: mcp.md Section 5.2
	uiPort, err := parsePortFromURL(baseURL)
	if err != nil {
		s.cfg.Log(0, "Warning: failed to parse UI port from URL %s: %v", baseURL, err)
	} else {
		if err := s.WriteUIPortFile(uiPort); err != nil {
			s.cfg.Log(0, "Warning: failed to write ui-port file: %v", err)
		}
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

		// mcp:pollingEvents() - check if agent is connected to /wait endpoint
		// Spec: mcp.md Section 8.2
		L.SetField(mcpTable, "pollingEvents", L.NewFunction(func(L *lua.LState) int {
			// Note: Called as mcp:pollingEvents() but we ignore the self argument
			if s.hasPollingClients(vendedID) {
				L.Push(lua.LTrue)
			} else {
				L.Push(lua.LFalse)
			}
			return 1
		}))

		// mcp:display(appName) - load and display an app
		// Checks for global (sanitized to camelCase), if not found loads from lua/appName.lua (symlink)
		L.SetField(mcpTable, "display", L.NewFunction(func(L *lua.LState) int {
			appName := L.CheckString(1)
			if appName == "" {
				L.Push(lua.LNil)
				L.Push(lua.LString("app name required"))
				return 2
			}

			// Sanitize app name to valid Lua identifier (camelCase)
			// "claude-panel" -> "claudePanel"
			globalName := sanitizeAppName(appName)

			// Check if global exists for this app
			appVal := L.GetGlobal(globalName)
			if appVal == lua.LNil {
				// Load the app file via RequireLuaFile (uses unified load tracker)
				// Apps are symlinked: apps/<app>/app.lua -> lua/<app>.lua
				luaFile := appName + ".lua"
				if _, err := session.DirectRequireLuaFile(luaFile); err != nil {
					L.Push(lua.LNil)
					L.Push(lua.LString(fmt.Sprintf("failed to load app %s: %v", appName, err)))
					return 2
				}
				// Get the global after loading
				appVal = L.GetGlobal(globalName)
			}

			// Assign to mcp.value
			if appVal != lua.LNil {
				L.SetField(mcpTable, "value", appVal)
			}

			L.Push(lua.LTrue)
			return 1
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

// sanitizeAppName converts an app name to a valid Lua identifier in camelCase.
// - Ensures the name starts with a lowercase letter
// - Converts snake-case/kebab-case to camelCase: "claude-panel" -> "claudePanel"
func sanitizeAppName(name string) string {
	if name == "" {
		return name
	}

	var result strings.Builder
	capitalizeNext := false

	for i, r := range name {
		if r == '-' || r == '_' {
			capitalizeNext = true
			continue
		}
		if i == 0 {
			// Ensure starts with lowercase
			result.WriteRune(toLower(r))
		} else if capitalizeNext {
			result.WriteRune(toUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + 32
	}
	return r
}

func toUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 32
	}
	return r
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

	// Include base_dir when configured or running
	if state == Configured || state == Running {
		result["base_dir"] = s.baseDir
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

// handleDisplay loads and displays an app by calling mcp.display(name) in Lua.
func (s *Server) handleDisplay(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.state != Running {
		return mcp.NewToolResultError("server not running - call ui_start first"), nil
	}

	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("arguments must be a map"), nil
	}

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("name must be a non-empty string"), nil
	}

	sessionID, ok := args["sessionId"].(string)
	if !ok || sessionID == "" {
		sessionID = s.currentVendedID
	}
	if sessionID == "" {
		return mcp.NewToolResultError("no active session"), nil
	}

	// Get the session for LoadCodeDirect
	session := s.UiServer.GetLuaSession(sessionID)
	if session == nil {
		return mcp.NewToolResultError(fmt.Sprintf("session %s not found", sessionID)), nil
	}

	// Call mcp.display(name) in Lua - the common implementation
	log := func(name, str string) {
		f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err) // i'm simplifying it here. you can do whatever you want.
		}
		defer f.Close()
		f.WriteString(str + "\n")
	}
	code := fmt.Sprintf("return mcp.display(%q)", name)
	result, err := s.SafeExecuteInSession(sessionID, func() (interface{}, error) {
		log("/tmp/bubba", "load code")
		return session.LoadCodeDirect("ui_display", code)
		//return session.LoadCode("ui_display", code)
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("display failed: %v", err)), nil
	}

	// Check if display returned an error (returns nil, errorMsg on failure)
	if result == nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to display app: %s", name)), nil
	}
	log("/tmp/bubba", "after load code, before sync")
	c := make(chan bool)
	s.SafeExecuteInSession(sessionID, func() (interface{}, error) {
		log("/tmp/bubba", "sync")
		close(c)
		return nil, nil
	})
	<-c
	log("/tmp/bubba", "finished displaying")
	return mcp.NewToolResultText(fmt.Sprintf("Displayed app: %s", name)), nil
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
