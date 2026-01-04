// Package mcp tests for tool handlers
// Test Design: test-MCP.md (Agent File Installation section)
package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zot/ui-engine/cli"
)

// TestInstallAgentFilesFreshInstall tests agent file installation when file is missing
func TestInstallAgentFilesFreshInstall(t *testing.T) {
	// Create temp directories
	tempDir := t.TempDir()
	baseDir := filepath.Join(tempDir, ".claude/ui")
	projectRoot := tempDir
	agentsDir := filepath.Join(projectRoot, ".claude", "agents")
	agentFile := filepath.Join(agentsDir, "ui-builder.md")

	// Create the source agents directory with test content
	sourceAgentsDir := filepath.Join(tempDir, "agents")
	os.MkdirAll(sourceAgentsDir, 0755)
	os.WriteFile(filepath.Join(sourceAgentsDir, "ui-builder.md"), []byte("# UI Builder Test\n"), 0644)

	// Create base_dir
	os.MkdirAll(baseDir, 0755)

	// Create server with default config
	cfg := cli.DefaultConfig()
	s := &Server{
		cfg: cfg,
	}

	// Verify file doesn't exist
	if _, err := os.Stat(agentFile); err == nil {
		t.Fatal("Agent file should not exist before installation")
	}

	// Run installation
	s.installAgentFiles(baseDir)

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
	baseDir := filepath.Join(tempDir, ".claude/ui")
	projectRoot := tempDir
	agentsDir := filepath.Join(projectRoot, ".claude", "agents")
	agentFile := filepath.Join(agentsDir, "ui-builder.md")

	// Create source with different content
	sourceAgentsDir := filepath.Join(tempDir, "agents")
	os.MkdirAll(sourceAgentsDir, 0755)
	os.WriteFile(filepath.Join(sourceAgentsDir, "ui-builder.md"), []byte("# New Content\n"), 0644)

	// Create destination with existing content
	os.MkdirAll(agentsDir, 0755)
	os.WriteFile(agentFile, []byte("# Existing Content\n"), 0644)

	os.MkdirAll(baseDir, 0755)

	cfg := cli.DefaultConfig()
	s := &Server{
		cfg: cfg,
	}

	// Run installation
	s.installAgentFiles(baseDir)

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
	baseDir := filepath.Join(tempDir, ".claude/ui")
	projectRoot := tempDir
	agentsDir := filepath.Join(projectRoot, ".claude", "agents")

	// Create source agents directory
	sourceAgentsDir := filepath.Join(tempDir, "agents")
	os.MkdirAll(sourceAgentsDir, 0755)
	os.WriteFile(filepath.Join(sourceAgentsDir, "ui-builder.md"), []byte("# Test\n"), 0644)

	os.MkdirAll(baseDir, 0755)

	// Verify .claude/agents doesn't exist
	if _, err := os.Stat(agentsDir); err == nil {
		t.Fatal("Agents directory should not exist before installation")
	}

	cfg := cli.DefaultConfig()
	s := &Server{
		cfg: cfg,
	}

	// Run installation
	s.installAgentFiles(baseDir)

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
	baseDir := filepath.Join(tempDir, "project", ".claude/ui")
	expectedAgentsDir := filepath.Join(tempDir, "project", ".claude", "agents")

	// Create source agents at the level above base_dir (parent = project)
	sourceAgentsDir := filepath.Join(tempDir, "project", "agents")
	os.MkdirAll(sourceAgentsDir, 0755)
	os.WriteFile(filepath.Join(sourceAgentsDir, "ui-builder.md"), []byte("# Test\n"), 0644)

	os.MkdirAll(baseDir, 0755)

	cfg := cli.DefaultConfig()
	s := &Server{
		cfg: cfg,
	}

	s.installAgentFiles(baseDir)

	// Verify agent installed to correct location
	agentFile := filepath.Join(expectedAgentsDir, "ui-builder.md")
	if _, err := os.Stat(agentFile); err != nil {
		t.Fatalf("Agent file should be at %s: %v", agentFile, err)
	}
}
