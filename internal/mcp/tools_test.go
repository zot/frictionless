// Package mcp tests for tool handlers
// Test Design: test-MCP.md (Skill File Installation section)
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
		state:   Running, // Server must be running for tools (ui_configure auto-starts)
	}
	return s
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

	// Simulate: /project/.claude/ui as base_dir
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
