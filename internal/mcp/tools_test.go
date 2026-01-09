// Package mcp tests for tool handlers
// Test Design: test-MCP.md (Agent File Installation section)
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
func createTestServer(t *testing.T, baseDir string) *Server {
	t.Helper()
	cfg := cli.DefaultConfig()
	s := &Server{
		cfg:     cfg,
		baseDir: baseDir,
		state:   Configured, // Must be configured to run install
	}
	return s
}

// setupInstallSource creates source files in install/agents/ for testing
func setupInstallSource(t *testing.T, projectRoot string, files map[string]string) {
	t.Helper()
	installAgentsDir := filepath.Join(projectRoot, "install", "agents")
	if err := os.MkdirAll(installAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create install/agents dir: %v", err)
	}
	for name, content := range files {
		path := filepath.Join(installAgentsDir, name)
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

// TestInstallAgentFilesFreshInstall tests agent file installation when file is missing
func TestInstallAgentFilesFreshInstall(t *testing.T) {
	// Create temp directories
	tempDir := t.TempDir()
	projectRoot := tempDir
	baseDir := filepath.Join(projectRoot, ".claude", "ui")
	agentsDir := filepath.Join(projectRoot, ".claude", "agents")
	agentFile := filepath.Join(agentsDir, "ui-builder.md")

	// Create base_dir
	os.MkdirAll(baseDir, 0755)

	// Create source files in install/agents/
	setupInstallSource(t, projectRoot, map[string]string{
		"ui-builder.md":  "# UI Builder Test\n",
		"ui-learning.md": "# UI Learning Test\n",
	})

	// Create server
	s := createTestServer(t, baseDir)

	// Verify file doesn't exist
	if _, err := os.Stat(agentFile); err == nil {
		t.Fatal("Agent file should not exist before installation")
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
	content, err := os.ReadFile(agentFile)
	if err != nil {
		t.Fatalf("Agent file should exist after installation: %v", err)
	}
	if string(content) != "# UI Builder Test\n" {
		t.Errorf("Agent file content mismatch: got %q", content)
	}
}

// TestInstallAgentFilesNoOpIfExists tests no installation when file already exists
func TestInstallAgentFilesNoOpIfExists(t *testing.T) {
	// Create temp directories
	tempDir := t.TempDir()
	projectRoot := tempDir
	baseDir := filepath.Join(projectRoot, ".claude", "ui")
	agentsDir := filepath.Join(projectRoot, ".claude", "agents")
	agentFile := filepath.Join(agentsDir, "ui-builder.md")

	// Create base_dir
	os.MkdirAll(baseDir, 0755)

	// Create source with different content
	setupInstallSource(t, projectRoot, map[string]string{
		"ui-builder.md":  "# New Content\n",
		"ui-learning.md": "# UI Learning\n",
	})

	// Create destination with existing content
	os.MkdirAll(agentsDir, 0755)
	os.WriteFile(agentFile, []byte("# Existing Content\n"), 0644)

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
	content, _ := os.ReadFile(agentFile)
	if string(content) != "# Existing Content\n" {
		t.Errorf("Existing file should not be overwritten: got %q", content)
	}
}

// TestInstallAgentFilesCreatesDirectory tests directory creation
func TestInstallAgentFilesCreatesDirectory(t *testing.T) {
	// Create temp directories
	tempDir := t.TempDir()
	projectRoot := tempDir
	baseDir := filepath.Join(projectRoot, ".claude", "ui")
	agentsDir := filepath.Join(projectRoot, ".claude", "agents")

	// Create base_dir only (not agents dir)
	os.MkdirAll(baseDir, 0755)

	// Create source agents
	setupInstallSource(t, projectRoot, map[string]string{
		"ui-builder.md":  "# Test\n",
		"ui-learning.md": "# Learning\n",
	})

	// Verify .claude/agents doesn't exist
	if _, err := os.Stat(agentsDir); err == nil {
		t.Fatal("Agents directory should not exist before installation")
	}

	// Create server
	s := createTestServer(t, baseDir)

	// Run installation
	_, err := callHandleInstall(s, false)
	if err != nil {
		t.Fatalf("handleInstall returned error: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(agentsDir); err != nil {
		t.Fatalf("Agents directory should be created: %v", err)
	}
}

// TestInstallAgentFilesPathResolution tests that path is resolved correctly
func TestInstallAgentFilesPathResolution(t *testing.T) {
	// Create temp directories simulating real structure
	tempDir := t.TempDir()

	// Simulate: /project/.claude/ui as base_dir
	// Agent should install to: /project/.claude/agents/
	projectRoot := filepath.Join(tempDir, "project")
	baseDir := filepath.Join(projectRoot, ".claude", "ui")
	expectedAgentsDir := filepath.Join(projectRoot, ".claude", "agents")

	// Create base_dir
	os.MkdirAll(baseDir, 0755)

	// Create source agents at install/agents/ relative to projectRoot
	setupInstallSource(t, projectRoot, map[string]string{
		"ui-builder.md":  "# Test\n",
		"ui-learning.md": "# Learning\n",
	})

	// Create server
	s := createTestServer(t, baseDir)

	// Run installation
	_, err := callHandleInstall(s, false)
	if err != nil {
		t.Fatalf("handleInstall returned error: %v", err)
	}

	// Verify agent installed to correct location
	agentFile := filepath.Join(expectedAgentsDir, "ui-builder.md")
	if _, err := os.Stat(agentFile); err != nil {
		t.Fatalf("Agent file should be at %s: %v", agentFile, err)
	}
}

// TestInstallRequiresConfiguredState tests that install fails if not configured
func TestInstallRequiresConfiguredState(t *testing.T) {
	cfg := cli.DefaultConfig()
	s := &Server{
		cfg:   cfg,
		state: Unconfigured,
	}

	result, err := callHandleInstall(s, false)
	if err != nil {
		t.Fatalf("handleInstall returned error: %v", err)
	}
	if !result.IsError {
		t.Error("handleInstall should return error when not configured")
	}
}

// TestInstallForceOverwrites tests that force=true overwrites existing files
func TestInstallForceOverwrites(t *testing.T) {
	// Create temp directories
	tempDir := t.TempDir()
	projectRoot := tempDir
	baseDir := filepath.Join(projectRoot, ".claude", "ui")
	agentsDir := filepath.Join(projectRoot, ".claude", "agents")
	agentFile := filepath.Join(agentsDir, "ui-builder.md")

	// Create base_dir
	os.MkdirAll(baseDir, 0755)

	// Create source with new content
	setupInstallSource(t, projectRoot, map[string]string{
		"ui-builder.md":  "# New Content\n",
		"ui-learning.md": "# UI Learning\n",
	})

	// Create destination with existing content
	os.MkdirAll(agentsDir, 0755)
	os.WriteFile(agentFile, []byte("# Old Content\n"), 0644)

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
	content, _ := os.ReadFile(agentFile)
	if string(content) != "# New Content\n" {
		t.Errorf("File should be overwritten with force=true: got %q", content)
	}
}
