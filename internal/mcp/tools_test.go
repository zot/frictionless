// Package mcp tests for tool handlers
// Test: test-MCP.md
package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/zot/ui-engine/cli"
)

// createTestServer creates a server configured for testing with the given baseDir
// This is a minimal server without Lua session support (for install/ClearLogs tests)
func createTestServer(t *testing.T, baseDir string) *Server {
	t.Helper()
	cfg := cli.DefaultConfig()
	s := &Server{
		cfg:     cfg,
		baseDir: baseDir,
		state:   Running, // Server must be running for tools (ui_configure auto-starts)
	}
	return s
}

// createTestServerWithSession creates a full MCP server with a working Lua session
// This is needed for ui_run tests that execute Lua code
func createTestServerWithSession(t *testing.T) (*Server, func()) {
	t.Helper()

	// Create temp directory for base_dir
	tempDir := t.TempDir()
	baseDir := filepath.Join(tempDir, ".claude", "ui")
	os.MkdirAll(filepath.Join(baseDir, "log"), 0755)
	os.MkdirAll(filepath.Join(baseDir, "lua"), 0755)

	// Create README.md to skip auto-install (which requires bundled binary)
	os.WriteFile(filepath.Join(baseDir, "README.md"), []byte("# Test\n**Version: 99.0.0**\n"), 0644)

	// Create config
	cfg := cli.DefaultConfig()
	cfg.Server.Dir = baseDir

	// Create ui-engine server
	uiServer := cli.NewServer(cfg)

	// Create MCP server with mock startFunc (returns fake URL without starting HTTP)
	mcpServer := NewServer(
		cfg,
		uiServer,
		uiServer.GetViewdefManager(),
		func(port int) (string, error) {
			// Mock: return fake URL without actually starting HTTP server
			return "http://127.0.0.1:12345", nil
		},
		func() int { return 0 }, // getSessionCount
	)

	// Configure the server
	if err := mcpServer.Configure(baseDir); err != nil {
		t.Fatalf("Failed to configure server: %v", err)
	}

	// Start and create session (uses mock startFunc)
	if _, err := mcpServer.StartAndCreateSession(); err != nil {
		t.Fatalf("Failed to start and create session: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		// Nothing to clean up - tempDir is handled by t.TempDir()
	}

	return mcpServer, cleanup
}

// callHandleRun calls handleRun with the given code
func callHandleRun(s *Server, code string) (*mcp.CallToolResult, error) {
	args := map[string]interface{}{
		"code": code,
	}
	request := mcp.CallToolRequest{}
	request.Params.Arguments = args
	return s.handleRun(context.Background(), request)
}

// setupInstallSource creates source files in install/init/skills/ui-builder/ for testing
func setupInstallSource(t *testing.T, projectRoot string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(projectRoot, "install", "init", "skills", "ui-builder", name)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write source file %s: %v", name, err)
		}
	}
}

// callHandleInstall calls handleInstall with the given force parameter
func callHandleInstall(s *Server, force bool) (*mcp.CallToolResult, error) {
	args := map[string]interface{}{}
	if force {
		args["force"] = true
	}
	request := mcp.CallToolRequest{}
	request.Params.Arguments = args
	return s.handleInstall(context.Background(), request)
}

// TestInstallSkillFilesFreshInstall tests skill file installation when file is missing
func TestInstallSkillFilesFreshInstall(t *testing.T) {
	// Create temp directories
	tempDir := t.TempDir()
	projectRoot := tempDir
	baseDir := filepath.Join(projectRoot, ".claude", "ui")
	skillFile := filepath.Join(projectRoot, ".claude", "skills", "ui-builder", "SKILL.md")

	// Create base_dir
	os.MkdirAll(baseDir, 0755)

	// Create source files in install/init/skills/ui-builder/
	setupInstallSource(t, projectRoot, map[string]string{
		"SKILL.md": "# UI Builder Skill Test\n",
	})

	// Create server
	s := createTestServer(t, baseDir)

	// Verify file doesn't exist
	if _, err := os.Stat(skillFile); err == nil {
		t.Fatal("Skill file should not exist before installation")
	}

	// Run installation
	result, err := callHandleInstall(s, false)
	if err != nil {
		t.Fatalf("handleInstall returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handleInstall returned tool error: %v", result.Content)
	}

	// Verify file was created
	content, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("Skill file should exist after installation: %v", err)
	}
	if string(content) != "# UI Builder Skill Test\n" {
		t.Errorf("Skill file content mismatch: got %q", content)
	}
}

// TestInstallSkillFilesNoOpIfExists tests no installation when file already exists
func TestInstallSkillFilesNoOpIfExists(t *testing.T) {
	// Create temp directories
	tempDir := t.TempDir()
	projectRoot := tempDir
	baseDir := filepath.Join(projectRoot, ".claude", "ui")
	skillsDir := filepath.Join(projectRoot, ".claude", "skills", "ui-builder")
	skillFile := filepath.Join(skillsDir, "SKILL.md")

	// Create base_dir
	os.MkdirAll(baseDir, 0755)

	// Create source with different content
	setupInstallSource(t, projectRoot, map[string]string{
		"SKILL.md": "# New Content\n",
	})

	// Create destination with existing content
	os.MkdirAll(skillsDir, 0755)
	os.WriteFile(skillFile, []byte("# Existing Content\n"), 0644)

	// Create server
	s := createTestServer(t, baseDir)

	// Run installation (force=false)
	result, err := callHandleInstall(s, false)
	if err != nil {
		t.Fatalf("handleInstall returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handleInstall returned tool error: %v", result.Content)
	}

	// Verify file was NOT overwritten
	content, _ := os.ReadFile(skillFile)
	if string(content) != "# Existing Content\n" {
		t.Errorf("Existing file should not be overwritten: got %q", content)
	}
}

// TestInstallSkillFilesCreatesDirectory tests directory creation
func TestInstallSkillFilesCreatesDirectory(t *testing.T) {
	// Create temp directories
	tempDir := t.TempDir()
	projectRoot := tempDir
	baseDir := filepath.Join(projectRoot, ".claude", "ui")
	skillsDir := filepath.Join(projectRoot, ".claude", "skills")

	// Create base_dir only (not skills dir)
	os.MkdirAll(baseDir, 0755)

	// Create source skills
	setupInstallSource(t, projectRoot, map[string]string{
		"SKILL.md": "# Test\n",
	})

	// Verify .claude/skills doesn't exist
	if _, err := os.Stat(skillsDir); err == nil {
		t.Fatal("Skills directory should not exist before installation")
	}

	// Create server
	s := createTestServer(t, baseDir)

	// Run installation
	_, err := callHandleInstall(s, false)
	if err != nil {
		t.Fatalf("handleInstall returned error: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(skillsDir); err != nil {
		t.Fatalf("Skills directory should be created: %v", err)
	}
}

// TestInstallSkillFilesPathResolution tests that path is resolved correctly
func TestInstallSkillFilesPathResolution(t *testing.T) {
	// Create temp directories simulating real structure
	tempDir := t.TempDir()

	// Simulate: /project/.ui as base_dir
	// Skill should install to: /project/.claude/skills/ui-builder/
	projectRoot := filepath.Join(tempDir, "project")
	baseDir := filepath.Join(projectRoot, ".claude", "ui")
	expectedSkillFile := filepath.Join(projectRoot, ".claude", "skills", "ui-builder", "SKILL.md")

	// Create base_dir
	os.MkdirAll(baseDir, 0755)

	// Create source skills at install/init/skills/ui-builder/ relative to projectRoot
	setupInstallSource(t, projectRoot, map[string]string{
		"SKILL.md": "# Test\n",
	})

	// Create server
	s := createTestServer(t, baseDir)

	// Run installation
	_, err := callHandleInstall(s, false)
	if err != nil {
		t.Fatalf("handleInstall returned error: %v", err)
	}

	// Verify skill installed to correct location
	if _, err := os.Stat(expectedSkillFile); err != nil {
		t.Fatalf("Skill file should be at %s: %v", expectedSkillFile, err)
	}
}

// TestInstallRequiresBaseDir tests that install fails if baseDir is not set
func TestInstallRequiresBaseDir(t *testing.T) {
	cfg := cli.DefaultConfig()
	s := &Server{
		cfg:     cfg,
		state:   Running,
		baseDir: "", // Empty baseDir should cause failure
	}

	result, err := callHandleInstall(s, false)
	if err != nil {
		t.Fatalf("handleInstall returned error: %v", err)
	}
	if !result.IsError {
		t.Error("handleInstall should return error when baseDir is not set")
	}
}

// TestInstallForceOverwrites tests that force=true overwrites existing files
func TestInstallForceOverwrites(t *testing.T) {
	// Create temp directories
	tempDir := t.TempDir()
	projectRoot := tempDir
	baseDir := filepath.Join(projectRoot, ".claude", "ui")
	skillsDir := filepath.Join(projectRoot, ".claude", "skills", "ui-builder")
	skillFile := filepath.Join(skillsDir, "SKILL.md")

	// Create base_dir
	os.MkdirAll(baseDir, 0755)

	// Create source with new content
	setupInstallSource(t, projectRoot, map[string]string{
		"SKILL.md": "# New Content\n",
	})

	// Create destination with existing content
	os.MkdirAll(skillsDir, 0755)
	os.WriteFile(skillFile, []byte("# Old Content\n"), 0644)

	// Create server
	s := createTestServer(t, baseDir)

	// Run installation with force=true
	result, err := callHandleInstall(s, true)
	if err != nil {
		t.Fatalf("handleInstall returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handleInstall returned tool error: %v", result.Content)
	}

	// Verify file WAS overwritten
	content, _ := os.ReadFile(skillFile)
	if string(content) != "# New Content\n" {
		t.Errorf("File should be overwritten with force=true: got %q", content)
	}
}

// ============================================================================
// ClearLogs Tests
// Test Design: test-MCP.md (ClearLogs section)
// Spec: mcp.md Section 5.1 - ui_configure clears logs
// ============================================================================

// TestClearLogsClearsAllFiles tests that ClearLogs removes all files in log directory
func TestClearLogsClearsAllFiles(t *testing.T) {
	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, "log")

	// Create log directory with some files
	os.MkdirAll(logDir, 0755)
	os.WriteFile(filepath.Join(logDir, "mcp.log"), []byte("go log content"), 0644)
	os.WriteFile(filepath.Join(logDir, "lua.log"), []byte("lua log content"), 0644)
	os.WriteFile(filepath.Join(logDir, "lua-err.log"), []byte("lua error content"), 0644)

	// Create server
	s := createTestServer(t, tempDir)

	// Verify files exist before clearing
	entries, _ := os.ReadDir(logDir)
	if len(entries) != 3 {
		t.Fatalf("Expected 3 log files before clearing, got %d", len(entries))
	}

	// Clear logs
	err := s.ClearLogs()
	if err != nil {
		t.Fatalf("ClearLogs returned error: %v", err)
	}

	// Verify all files were removed
	entries, _ = os.ReadDir(logDir)
	if len(entries) != 0 {
		t.Errorf("Expected 0 log files after clearing, got %d", len(entries))
	}
}

// TestClearLogsCallsCallback tests that ClearLogs invokes the onClearLogs callback
func TestClearLogsCallsCallback(t *testing.T) {
	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, "log")

	// Create log directory with a file
	os.MkdirAll(logDir, 0755)
	os.WriteFile(filepath.Join(logDir, "mcp.log"), []byte("content"), 0644)

	// Create server
	s := createTestServer(t, tempDir)

	// Track callback invocation
	callbackCalled := false
	s.SetOnClearLogs(func() {
		callbackCalled = true
	})

	// Clear logs
	err := s.ClearLogs()
	if err != nil {
		t.Fatalf("ClearLogs returned error: %v", err)
	}

	// Verify callback was called
	if !callbackCalled {
		t.Error("onClearLogs callback should have been called")
	}
}

// TestClearLogsHandlesMissingDirectory tests that ClearLogs handles missing log directory
func TestClearLogsHandlesMissingDirectory(t *testing.T) {
	tempDir := t.TempDir()
	// Don't create log directory

	// Create server
	s := createTestServer(t, tempDir)

	// Clear logs should not error on missing directory
	err := s.ClearLogs()
	if err != nil {
		t.Errorf("ClearLogs should not error on missing directory: %v", err)
	}
}

// TestClearLogsSkipsSubdirectories tests that ClearLogs only removes files, not subdirs
func TestClearLogsSkipsSubdirectories(t *testing.T) {
	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, "log")
	subDir := filepath.Join(logDir, "subdir")

	// Create log directory with a file and a subdirectory
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(logDir, "mcp.log"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(subDir, "nested.log"), []byte("nested"), 0644)

	// Create server
	s := createTestServer(t, tempDir)

	// Clear logs
	err := s.ClearLogs()
	if err != nil {
		t.Fatalf("ClearLogs returned error: %v", err)
	}

	// Verify file was removed but subdirectory remains
	if _, err := os.Stat(filepath.Join(logDir, "mcp.log")); err == nil {
		t.Error("mcp.log should have been removed")
	}
	if _, err := os.Stat(subDir); err != nil {
		t.Error("Subdirectory should not be removed")
	}
	if _, err := os.Stat(filepath.Join(subDir, "nested.log")); err != nil {
		t.Error("Files in subdirectory should not be removed")
	}
}

// TestClearLogsNoCallbackIfNotSet tests that ClearLogs works without callback
func TestClearLogsNoCallbackIfNotSet(t *testing.T) {
	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, "log")

	// Create log directory with a file
	os.MkdirAll(logDir, 0755)
	os.WriteFile(filepath.Join(logDir, "mcp.log"), []byte("content"), 0644)

	// Create server without setting callback
	s := createTestServer(t, tempDir)

	// Clear logs should work without panic
	err := s.ClearLogs()
	if err != nil {
		t.Fatalf("ClearLogs returned error: %v", err)
	}

	// Verify file was removed
	if _, err := os.Stat(filepath.Join(logDir, "mcp.log")); err == nil {
		t.Error("mcp.log should have been removed")
	}
}

// ============================================================================
// ui_run Tests
// Test Design: test-MCP.md (Tool - ui_run section)
// Spec: mcp.md Section 5.2 - ui_run
// ============================================================================

// TestRunExecuteCode tests basic Lua code execution
func TestRunExecuteCode(t *testing.T) {
	s, cleanup := createTestServerWithSession(t)
	defer cleanup()

	// Execute simple arithmetic
	result, err := callHandleRun(s, "return 1 + 1")
	if err != nil {
		t.Fatalf("handleRun returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handleRun returned tool error: %v", result.Content)
	}

	// Check result contains "2"
	text := getTextContent(result)
	if text != "2" {
		t.Errorf("Expected result '2', got %q", text)
	}
}

// TestRunSessionAccess tests accessing session global
func TestRunSessionAccess(t *testing.T) {
	s, cleanup := createTestServerWithSession(t)
	defer cleanup()

	// Access session global - should not error
	result, err := callHandleRun(s, "return session ~= nil")
	if err != nil {
		t.Fatalf("handleRun returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handleRun returned tool error: %v", result.Content)
	}

	// Session should exist
	text := getTextContent(result)
	if text != "true" {
		t.Errorf("Expected session to exist (true), got %q", text)
	}
}

// TestRunJSONMarshalling tests that tables are marshalled to JSON
func TestRunJSONMarshalling(t *testing.T) {
	s, cleanup := createTestServerWithSession(t)
	defer cleanup()

	// Return a table
	result, err := callHandleRun(s, `return {a=1, b="text"}`)
	if err != nil {
		t.Fatalf("handleRun returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handleRun returned tool error: %v", result.Content)
	}

	// Result should be JSON object
	text := getTextContent(result)
	if text == "" {
		t.Fatal("Expected non-empty result")
	}
	// Should contain the keys (order may vary in JSON)
	if !contains(text, `"a"`) || !contains(text, `"b"`) {
		t.Errorf("Expected JSON with keys 'a' and 'b', got %q", text)
	}
	if !contains(text, "1") || !contains(text, `"text"`) {
		t.Errorf("Expected JSON with values 1 and 'text', got %q", text)
	}
}

// TestRunNonJSONResult tests that Lua functions convert to null in JSON
// Note: Lua functions are converted to nil in Go, which marshals to "null"
// The non-json wrapper is only triggered for Go types that fail JSON marshalling
func TestRunNonJSONResult(t *testing.T) {
	s, cleanup := createTestServerWithSession(t)
	defer cleanup()

	// Return a function - converts to nil in Go, marshals to "null"
	result, err := callHandleRun(s, "return function() end")
	if err != nil {
		t.Fatalf("handleRun returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handleRun returned tool error: %v", result.Content)
	}

	// Functions convert to null (nil in Go)
	text := getTextContent(result)
	if text != "null" {
		t.Errorf("Expected 'null' for function, got %q", text)
	}
}

// TestRunMCPGlobalAccess tests that mcp global is available
func TestRunMCPGlobalAccess(t *testing.T) {
	s, cleanup := createTestServerWithSession(t)
	defer cleanup()

	// Access mcp global
	result, err := callHandleRun(s, "return mcp.type")
	if err != nil {
		t.Fatalf("handleRun returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handleRun returned tool error: %v", result.Content)
	}

	// mcp.type should be "MCP"
	text := getTextContent(result)
	if !contains(text, "MCP") {
		t.Errorf("Expected mcp.type to be 'MCP', got %q", text)
	}
}

// TestRunNoActiveSession tests error when no session exists
func TestRunNoActiveSession(t *testing.T) {
	// Create server without session
	cfg := cli.DefaultConfig()
	s := &Server{
		cfg:   cfg,
		state: Running,
		// No currentVendedID set
	}

	result, err := callHandleRun(s, "return 1")
	if err != nil {
		t.Fatalf("handleRun returned error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error when no active session")
	}
}

// getTextContent extracts text content from a tool result
func getTextContent(result *mcp.CallToolResult) string {
	if len(result.Content) == 0 {
		return ""
	}
	// Content is []ContentBlock, first one should be text
	if textContent, ok := result.Content[0].(mcp.TextContent); ok {
		return textContent.Text
	}
	return ""
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
