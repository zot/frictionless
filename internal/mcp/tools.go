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
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	lua "github.com/yuin/gopher-lua"
	"github.com/zot/ui-engine/cli"
)

func (s *Server) registerTools() {
	// ui_configure
	// Spec: mcp.md section 5.1
	s.mcpServer.AddTool(mcp.NewTool("ui_configure",
		mcp.WithDescription("Prepare the server environment and file system. Must be the first tool called."),
		mcp.WithString("base_dir", mcp.Required(), mcp.Description("Absolute path to the UI working directory. Use {project}/.claude/ui unless user specifies otherwise.")),
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

// Spec: mcp.md section 5.1
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

	// Stop current session if running (allows reconfiguration)
	if err := s.Stop(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to stop current session: %v", err)), nil
	}

	// Configure handles directory creation, log paths, and auto-install if README.md missing
	// Spec: mcp.md Section 3.1 - Startup Behavior
	if err := s.Configure(baseDir); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Return structured response
	response := map[string]interface{}{
		"status":   "configured",
		"base_dir": baseDir,
		"log_path": filepath.Join(baseDir, "log"),
	}

	responseJSON, _ := json.Marshal(response)
	return mcp.NewToolResultText(string(responseJSON)), nil
}

// parseReadmeVersion extracts the version from README.md.
// Looks for **Version: X.Y.Z** pattern near the top.
// Returns empty string if no version found.
func parseReadmeVersion(content []byte) string {
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "**Version:") && strings.HasSuffix(line, "**") {
			// Extract version from **Version: X.Y.Z**
			version := strings.TrimPrefix(line, "**Version:")
			version = strings.TrimSuffix(version, "**")
			return strings.TrimSpace(version)
		}
	}
	return ""
}

// InstallResult contains the results of an install operation.
type InstallResult struct {
	Installed        []string `json:"installed"`
	Skipped          []string `json:"skipped"`
	Appended         []string `json:"appended"`
	VersionSkipped   bool     `json:"version_skipped,omitempty"`
	BundledVersion   string   `json:"bundled_version,omitempty"`
	InstalledVersion string   `json:"installed_version,omitempty"`
	Hint             string   `json:"hint,omitempty"`
}

// installFile installs a single file from the bundle.
// bundlePath: path relative to bundle root (e.g., "resources/reference.md")
// destPath: absolute destination path
// mode: file permissions (0644 for regular files, 0755 for scripts)
// Returns: "installed", "skipped", or "" (file not found in bundle)
func (s *Server) installFile(bundlePath, destPath string, mode os.FileMode, force bool) (string, error) {
	// Skip if file exists (unless force)
	if _, err := os.Stat(destPath); err == nil && !force {
		return "skipped", nil
	}

	// Read content from bundle
	content, err := cli.BundleReadFile(bundlePath)
	if err != nil || len(content) == 0 {
		s.cfg.Log(1, "File not found in bundle: %s", bundlePath)
		return "", nil
	}

	// Create directory and write file
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for %s: %v", bundlePath, err)
	}
	if err := os.WriteFile(destPath, content, mode); err != nil {
		return "", fmt.Errorf("failed to write %s: %v", filepath.Base(destPath), err)
	}

	s.cfg.Log(1, "Installed: %s", destPath)
	return "installed", nil
}

// Install installs bundled configuration files.
// This is the core install logic used by both Configure (auto-install) and handleInstall (MCP tool).
// Spec: mcp.md section 5.7
func (s *Server) Install(force bool) (*InstallResult, error) {
	if s.baseDir == "" {
		return nil, fmt.Errorf("server not configured (baseDir not set)")
	}

	// Require bundled binary for install
	if bundled, _ := cli.IsBundled(); !bundled {
		return nil, fmt.Errorf("install requires a bundled binary (use 'make build')")
	}

	// Project root is the grandparent of baseDir (e.g., .claude/ui -> .)
	projectRoot := filepath.Dir(filepath.Dir(s.baseDir))

	// Version checking: compare bundled README version with installed version
	var bundledVersion, installedVersion string
	if bundledContent, err := cli.BundleReadFile("README.md"); err == nil {
		bundledVersion = parseReadmeVersion(bundledContent)
	}
	if installedContent, err := os.ReadFile(filepath.Join(s.baseDir, "README.md")); err == nil {
		installedVersion = parseReadmeVersion(installedContent)
	}

	// Skip if installed version >= bundled version (unless force)
	if !force && installedVersion != "" && bundledVersion != "" {
		if compareSemver(installedVersion, bundledVersion) >= 0 {
			return &InstallResult{
				VersionSkipped:   true,
				BundledVersion:   bundledVersion,
				InstalledVersion: installedVersion,
				Hint:             "Use force=true to reinstall",
			}, nil
		}
	}

	var installed, skipped []string

	// Helper to track install results
	track := func(relPath, status string) {
		switch status {
		case "installed":
			installed = append(installed, relPath)
		case "skipped":
			skipped = append(skipped, relPath)
		}
	}

	// 1. Install skills and agents to {project}/.claude/{category}/
	claudeFiles := []struct{ category, file string }{
		{"skills/ui", "SKILL.md"},
		{"skills/ui-builder", "SKILL.md"},
		{"skills/ui-builder", "examples/requirements.md"},
		{"skills/ui-builder", "examples/design.md"},
		{"skills/ui-builder", "examples/app.lua"},
		{"skills/ui-builder", "examples/viewdefs/ContactApp.DEFAULT.html"},
		{"skills/ui-builder", "examples/viewdefs/Contact.list-item.html"},
		{"skills/ui-builder", "examples/viewdefs/ChatMessage.list-item.html"},
		{"agents", "ui-builder.md"},
	}
	for _, f := range claudeFiles {
		bundlePath := filepath.Join(f.category, f.file)
		destPath := filepath.Join(projectRoot, ".claude", f.category, f.file)
		relPath := filepath.Join(".claude", f.category, f.file)
		status, err := s.installFile(bundlePath, destPath, 0644, force)
		if err != nil {
			return nil, err
		}
		track(relPath, status)
	}

	// 2. Install resources to {base_dir}/resources/
	for _, file := range []string{"reference.md", "viewdefs.md", "lua.md", "mcp.md"} {
		bundlePath := filepath.Join("resources", file)
		destPath := filepath.Join(s.baseDir, "resources", file)
		status, err := s.installFile(bundlePath, destPath, 0644, force)
		if err != nil {
			return nil, err
		}
		track(bundlePath, status)
	}

	// 3. Install viewdefs to {base_dir}/viewdefs/
	for _, file := range []string{"lua.ViewList.DEFAULT.html", "lua.ViewListItem.list-item.html", "MCP.DEFAULT.html"} {
		bundlePath := filepath.Join("viewdefs", file)
		destPath := filepath.Join(s.baseDir, "viewdefs", file)
		status, err := s.installFile(bundlePath, destPath, 0644, force)
		if err != nil {
			return nil, err
		}
		track(bundlePath, status)
	}

	// 4. Install scripts to {base_dir}/ (executable)
	for _, file := range []string{"event", "state", "variables", "linkapp"} {
		destPath := filepath.Join(s.baseDir, file)
		status, err := s.installFile(file, destPath, 0755, force)
		if err != nil {
			return nil, err
		}
		track(file, status)
	}

	// 5. Install html files to {base_dir}/html/ (dynamically discovered from bundle)
	htmlFiles, _ := cli.BundleListFiles("html")
	for _, bundlePath := range htmlFiles {
		fileName := filepath.Base(bundlePath)
		destPath := filepath.Join(s.baseDir, "html", fileName)
		relPath := filepath.Join("html", fileName)
		status, err := s.installFile(bundlePath, destPath, 0644, force)
		if err != nil {
			return nil, err
		}
		track(relPath, status)
	}

	// 6. Install README.md to {base_dir}/
	status, err := s.installFile("README.md", filepath.Join(s.baseDir, "README.md"), 0644, force)
	if err != nil {
		return nil, err
	}
	track("README.md", status)

	return &InstallResult{
		Installed: installed,
		Skipped:   skipped,
	}, nil
}

// compareSemver compares two semantic versions.
// Returns: -1 if a < b, 0 if a == b, 1 if a > b
func compareSemver(a, b string) int {
	parseVersion := func(v string) (int, int, int) {
		parts := strings.Split(v, ".")
		major, minor, patch := 0, 0, 0
		if len(parts) >= 1 {
			major, _ = strconv.Atoi(parts[0])
		}
		if len(parts) >= 2 {
			minor, _ = strconv.Atoi(parts[1])
		}
		if len(parts) >= 3 {
			patch, _ = strconv.Atoi(parts[2])
		}
		return major, minor, patch
	}

	aMajor, aMinor, aPatch := parseVersion(a)
	bMajor, bMinor, bPatch := parseVersion(b)

	if aMajor != bMajor {
		if aMajor < bMajor {
			return -1
		}
		return 1
	}
	if aMinor != bMinor {
		if aMinor < bMinor {
			return -1
		}
		return 1
	}
	if aPatch != bPatch {
		if aPatch < bPatch {
			return -1
		}
		return 1
	}
	return 0
}

// handleInstall installs bundled configuration files.
// Spec: mcp.md section 5.7
func (s *Server) handleInstall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check state - server must be configured
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

	// Call the Install method
	result, err := s.Install(force)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
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

		// mcp:status() - get current server status
		// Spec: mcp.md Section 4.3
		L.SetField(mcpTable, "status", L.NewFunction(func(L *lua.LState) int {
			// Note: Called as mcp:status() but we ignore the self argument
			result := L.NewTable()

			// Get server state
			s.mu.RLock()
			state := s.state
			url := s.url
			baseDir := s.baseDir
			s.mu.RUnlock()

			L.SetField(result, "state", lua.LString(stateToString(state)))
			L.SetField(result, "base_dir", lua.LString(baseDir))

			// Get bundled version (same logic as handleStatus)
			isBundled, _ := cli.IsBundled()
			var bundledContent []byte
			var err error
			if isBundled {
				bundledContent, err = cli.BundleReadFile("README.md")
			} else {
				bundledContent, err = os.ReadFile(filepath.Join("install", "README.md"))
			}
			if err == nil {
				if version := parseReadmeVersion(bundledContent); version != "" {
					L.SetField(result, "version", lua.LString(version))
				}
			}

			// Add running-only fields
			if state == Running {
				L.SetField(result, "url", lua.LString(url))
				if s.getSessionCount != nil {
					L.SetField(result, "sessions", lua.LNumber(s.getSessionCount()))
				}
			}

			L.Push(result)
			return 1
		}))

		// Load mcp.lua if it exists to extend the mcp global
		// Spec: mcp.md Section 4.3 "Extension via mcp.lua"
		mcpLuaPath := filepath.Join(s.baseDir, "lua", "mcp.lua")
		if _, err := os.Stat(mcpLuaPath); err == nil {
			if err := L.DoFile(mcpLuaPath); err != nil {
				return nil, fmt.Errorf("failed to load mcp.lua: %w", err)
			}
		}

		// Load init.lua from each app directory if it exists
		// Spec: mcp.md Section 4.3 "App Initialization (init.lua)"
		appsDir := filepath.Join(s.baseDir, "apps")
		if entries, err := os.ReadDir(appsDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					initPath := filepath.Join(appsDir, entry.Name(), "init.lua")
					if _, err := os.Stat(initPath); err == nil {
						if err := L.DoFile(initPath); err != nil {
							return nil, fmt.Errorf("failed to load %s/init.lua: %w", entry.Name(), err)
						}
					}
				}
			}
		}

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

	// Always include bundled version from README.md
	// Spec: mcp.md section 5.6 - version is always present
	isBundled, _ := cli.IsBundled()
	var bundledContent []byte
	var err error
	if isBundled {
		bundledContent, err = cli.BundleReadFile("README.md")
	} else {
		// Development mode: read from install/README.md
		bundledContent, err = os.ReadFile(filepath.Join("install", "README.md"))
	}
	if err == nil {
		if version := parseReadmeVersion(bundledContent); version != "" {
			result["version"] = version
		}
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
	case Configured:
		return "configured"
	case Running:
		return "running"
	default:
		return "unknown"
	}
}
