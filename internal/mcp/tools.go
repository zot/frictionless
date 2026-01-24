// Package mcp implements the Model Context Protocol server.
// CRC: crc-MCPTool.md
// Spec: mcp.md
// Sequence: seq-mcp-lifecycle.md, seq-mcp-run.md
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		mcp.WithString("base_dir", mcp.Required(), mcp.Description("Absolute path to the UI working directory. Use {project}/.ui unless user specifies otherwise.")),
	), s.handleConfigure)

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

	// ui_status
	s.mcpServer.AddTool(mcp.NewTool("ui_status",
		mcp.WithDescription("Get current server status including browser connection count"),
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

	// ui_audit
	// Spec: specs/ui-audit.md
	s.mcpServer.AddTool(mcp.NewTool("ui_audit",
		mcp.WithDescription("Analyze an app for code quality violations (dead methods, viewdef issues)"),
		mcp.WithString("name", mcp.Required(), mcp.Description("App name to audit")),
	), s.handleAudit)
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

	// Start and create session (shared with process startup)
	// Spec: mcp.md Section 5.1 - ui_configure starts server, returns base URL
	baseURL, err := s.StartAndCreateSession()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Return structured response with URL (no session ID per spec)
	response := map[string]interface{}{
		"base_dir":       baseDir,
		"url":            baseURL,
		"install_needed": false,
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
	Suggestions      []string `json:"suggestions,omitempty"`
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

	// Project root is the grandparent of baseDir (e.g., .ui -> .)
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

	// 7. Check for optional external dependencies and add suggestions
	var suggestions []string
	codeSimplifierPath := filepath.Join(projectRoot, ".claude", "agents", "code-simplifier.md")
	if _, err := os.Stat(codeSimplifierPath); os.IsNotExist(err) {
		suggestions = append(suggestions, "Run `claude plugin install code-simplifier` to enable code simplification")
	}

	return &InstallResult{
		Installed:   installed,
		Skipped:     skipped,
		Suggestions: suggestions,
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
// Spec: mcp.md section 5.6
func (s *Server) handleInstall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check state - server must be running
	if s.state != Running {
		return mcp.NewToolResultError("ui_install requires the server to be running. Call ui_configure first."), nil
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

// setupMCPGlobal creates the mcp global object in Lua with Go functions attached.
func (s *Server) setupMCPGlobal(vendedID string) error {
	session := s.UiServer.GetLuaSession(vendedID)
	if session == nil {
		return fmt.Errorf("session %s not found", vendedID)
	}

	// Use SafeExecuteInSession to ensure proper executor context and session global
	_, err := s.SafeExecuteInSession(vendedID, func() (interface{}, error) {
		L := session.State

		// Create mcp table (instance)
		mcpTable := L.NewTable()
		L.SetGlobal("mcp", mcpTable)

		// Create MCP table (namespace for nested prototypes like MCP.AppMenuItem)
		// This allows mcp.lua to do: MCP.AppMenuItem = session:prototype(...)
		mcpNamespace := L.NewTable()
		L.SetGlobal("MCP", mcpNamespace)

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

		// mcp:app(appName) - load an app without displaying it
		// Returns the app global, or nil, errmsg
		L.SetField(mcpTable, "app", L.NewFunction(func(L *lua.LState) int {
			// Arg 1 is self (mcp table) when called with colon notation
			appName := L.CheckString(2)
			if appName == "" {
				L.Push(lua.LNil)
				L.Push(lua.LString("app name required"))
				return 2
			}

			// Sanitize app name to valid Lua identifier (camelCase)
			globalName := sanitizeAppName(appName)

			// Check if global exists for this app
			appVal := L.GetGlobal(globalName)
			if appVal == lua.LNil {
				// Load the app file via RequireLuaFile
				luaFile := appName + ".lua"
				if _, err := session.DirectRequireLuaFile(luaFile); err != nil {
					L.Push(lua.LNil)
					L.Push(lua.LString(fmt.Sprintf("failed to load app %s: %v", appName, err)))
					return 2
				}
				appVal = L.GetGlobal(globalName)
			}

			if appVal == lua.LNil {
				L.Push(lua.LNil)
				L.Push(lua.LString(fmt.Sprintf("app %s has no global '%s'", appName, globalName)))
				return 2
			}

			L.Push(appVal)
			return 1
		}))

		// mcp:display(appName) - load and display an app
		// Returns true, or nil, errmsg
		L.SetField(mcpTable, "display", L.NewFunction(func(L *lua.LState) int {
			// Arg 1 is self (mcp table) when called with colon notation
			appName := L.CheckString(2)
			if appName == "" {
				L.Push(lua.LNil)
				L.Push(lua.LString("app name required"))
				return 2
			}

			// Sanitize app name to valid Lua identifier (camelCase)
			globalName := sanitizeAppName(appName)

			// Check if global exists for this app
			appVal := L.GetGlobal(globalName)
			if appVal == lua.LNil {
				// Load the app file via RequireLuaFile
				luaFile := appName + ".lua"
				if _, err := session.DirectRequireLuaFile(luaFile); err != nil {
					L.Push(lua.LNil)
					L.Push(lua.LString(fmt.Sprintf("failed to load app %s: %v", appName, err)))
					return 2
				}
				appVal = L.GetGlobal(globalName)
			}

			// Assign to mcp.value to display
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
			mcpPort := s.mcpPort
			s.mu.RUnlock()

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
				L.SetField(result, "mcp_port", lua.LNumber(mcpPort))
				if s.getSessionCount != nil {
					L.SetField(result, "sessions", lua.LNumber(s.getSessionCount()))
				}
			}

			L.Push(result)
			return 1
		}))

		// Load mcp.lua if it exists to extend the mcp global
		// Use DirectRequireLuaFile to register for hot-loading
		// Spec: mcp.md Section 4.3 "Extension via mcp.lua"
		if _, err := session.DirectRequireLuaFile("mcp.lua"); err != nil {
			// Ignore "not found" errors - mcp.lua is optional
			if !os.IsNotExist(err) && !strings.Contains(err.Error(), "not found") {
				return nil, fmt.Errorf("failed to load mcp.lua: %w", err)
			}
		}

		// Load init.lua from each app directory if it exists
		// Use DirectRequireLuaFile to register for hot-loading
		// Path is relative to baseDir (e.g., "apps/myapp/init.lua")
		// Spec: mcp.md Section 4.3 "App Initialization (init.lua)"
		appsDir := filepath.Join(s.baseDir, "apps")
		if entries, err := os.ReadDir(appsDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					initAbsPath := filepath.Join(appsDir, entry.Name(), "init.lua")
					if _, err := os.Stat(initAbsPath); err == nil {
						// Path relative to baseDir for tracking
						initRelPath := filepath.Join("apps", entry.Name(), "init.lua")
						if _, err := session.DirectRequireLuaFile(initRelPath); err != nil {
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
		return mcp.NewToolResultError("no active session - call ui_configure first"), nil
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
		return mcp.NewToolResultError("no active session - call ui_configure first"), nil
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

// CRC: crc-MCPTool.md
// Spec: mcp.md (section 5.5)
func (s *Server) handleStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.mu.RLock()
	state := s.state
	url := s.url
	baseDir := s.baseDir
	mcpPort := s.mcpPort
	s.mu.RUnlock()

	result := map[string]interface{}{}

	// Always include bundled version from README.md
	// Spec: mcp.md section 5.5 - version is always present
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

	// Include base_dir when set
	if baseDir != "" {
		result["base_dir"] = baseDir
	}

	// Include url, mcp_port, and sessions when running
	if state == Running {
		result["url"] = url
		result["mcp_port"] = mcpPort
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

// handleDisplay loads and displays an app by calling mcp:display(name) in Lua.
func (s *Server) handleDisplay(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.state != Running {
		return mcp.NewToolResultError("server not running - call ui_configure first"), nil
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

	// Call mcp:display(name) in Lua
	code := fmt.Sprintf("return mcp:display(%q)", name)
	result, err := s.SafeExecuteInSession(sessionID, func() (interface{}, error) {
		return session.LoadCodeDirect("ui_display", code)
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("display failed: %v", err)), nil
	}

	// Check if display returned an error (returns nil, errorMsg on failure)
	if result == nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to display app: %s", name)), nil
	}

	// Wait for the display to complete by executing a no-op in the session
	done := make(chan struct{})
	s.SafeExecuteInSession(sessionID, func() (interface{}, error) {
		close(done)
		return nil, nil
	})
	<-done

	return mcp.NewToolResultText(fmt.Sprintf("Displayed app: %s", name)), nil
}

// handleAudit analyzes an app for code quality violations.
// CRC: crc-Auditor.md
// Seq: seq-audit.md
func (s *Server) handleAudit(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.mu.RLock()
	baseDir := s.baseDir
	s.mu.RUnlock()

	if baseDir == "" {
		return mcp.NewToolResultError("server not configured - call ui_configure first"), nil
	}

	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("arguments must be a map"), nil
	}

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("name must be a non-empty string"), nil
	}

	result, err := AuditApp(baseDir, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("audit failed: %v", err)), nil
	}

	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}

// HTTP Tool API handlers (Spec 2.5)
// These wrap the MCP tool handlers for HTTP access by spawned agents.

// apiResponse writes a JSON response for the Tool API.
func apiResponse(w http.ResponseWriter, result interface{}, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"result": result})
}

// apiError writes an error response.
func apiError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// parseJSONBody parses JSON from request body into a map.
func parseJSONBody(r *http.Request) (map[string]interface{}, error) {
	if r.Body == nil {
		return make(map[string]interface{}), nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return make(map[string]interface{}), nil
	}
	var args map[string]interface{}
	if err := json.Unmarshal(body, &args); err != nil {
		return nil, err
	}
	return args, nil
}

// callMCPHandler invokes an MCP handler and extracts the result.
func (s *Server) callMCPHandler(
	handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error),
	args map[string]interface{},
) (interface{}, error) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	result, err := handler(context.Background(), req)
	if err != nil {
		return nil, err
	}
	// Extract text content from result
	if result != nil && len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			// Try to parse as JSON first
			var jsonResult interface{}
			if err := json.Unmarshal([]byte(textContent.Text), &jsonResult); err == nil {
				return jsonResult, nil
			}
			return textContent.Text, nil
		}
	}
	return nil, nil
}

// handleAPIStatus handles GET /api/ui_status
func (s *Server) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		apiError(w, http.StatusMethodNotAllowed, "GET required")
		return
	}
	result, err := s.callMCPHandler(s.handleStatus, nil)
	apiResponse(w, result, err)
}

// handleAPIRun handles POST /api/ui_run
func (s *Server) handleAPIRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apiError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	args, err := parseJSONBody(r)
	if err != nil {
		apiError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	result, err := s.callMCPHandler(s.handleRun, args)
	apiResponse(w, result, err)
}

// handleAPIDisplay handles POST /api/ui_display
func (s *Server) handleAPIDisplay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apiError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	args, err := parseJSONBody(r)
	if err != nil {
		apiError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	result, err := s.callMCPHandler(s.handleDisplay, args)
	apiResponse(w, result, err)
}

// handleAPIConfigure handles POST /api/ui_configure
func (s *Server) handleAPIConfigure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apiError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	args, err := parseJSONBody(r)
	if err != nil {
		apiError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	result, err := s.callMCPHandler(s.handleConfigure, args)
	apiResponse(w, result, err)
}

// handleAPIInstall handles POST /api/ui_install
func (s *Server) handleAPIInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apiError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	args, err := parseJSONBody(r)
	if err != nil {
		apiError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	result, err := s.callMCPHandler(s.handleInstall, args)
	apiResponse(w, result, err)
}

// handleAPIOpenBrowser handles POST /api/ui_open_browser
func (s *Server) handleAPIOpenBrowser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apiError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	args, err := parseJSONBody(r)
	if err != nil {
		apiError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	result, err := s.callMCPHandler(s.handleOpenBrowser, args)
	apiResponse(w, result, err)
}

// handleAPIAudit handles POST /api/ui_audit
func (s *Server) handleAPIAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apiError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	args, err := parseJSONBody(r)
	if err != nil {
		apiError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	result, err := s.callMCPHandler(s.handleAudit, args)
	apiResponse(w, result, err)
}
