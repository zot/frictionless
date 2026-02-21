// Package mcp implements the Model Context Protocol server.
// CRC: crc-MCPTool.md
// Spec: mcp.md
// Sequence: seq-mcp-lifecycle.md, seq-mcp-run.md
package mcp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	"github.com/yuin/goldmark"
	lua "github.com/yuin/gopher-lua"
	"github.com/zot/ui-engine/cli"
)

func (s *Server) registerTools() {
	// ui_configure
	// Spec: mcp.md section 5.1
	s.mcpServer.AddTool(mcp.NewTool("ui_configure",
		mcp.WithDescription("Reconfigure and restart the UI server with a different base directory. Optional—server auto-configures at startup using --dir (defaults to .ui)."),
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
		mcp.WithDescription("Install bundled configuration files (skill files). Server must be running (auto-starts on MCP connection)."),
		mcp.WithBoolean("force", mcp.Description("If true, overwrites existing files. Defaults to false.")),
	), s.handleInstall)

	// ui_update
	s.mcpServer.AddTool(mcp.NewTool("ui_update",
		mcp.WithDescription("Smart update: installs new bundled files, overwrites unmodified files, detects conflicts for user-modified files. Uses hash-based manifest for conflict detection."),
	), s.handleUpdate)

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

	// ui_theme
	s.mcpServer.AddTool(mcp.NewTool("ui_theme",
		mcp.WithDescription("Theme management: list available themes, get semantic classes, audit app theme usage"),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action: list, classes, audit")),
		mcp.WithString("theme", mcp.Description("Theme name (defaults to current theme)")),
		mcp.WithString("app", mcp.Description("App name (required for audit action)")),
	), s.handleTheme)
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
// Returns: "installed", "skipped", or "" (file not found in bundle)
func (s *Server) installFile(bundlePath, destPath string, mode os.FileMode, force bool, fileInfoMap map[string]cli.BundleFileInfo) (string, error) {
	fileExists := false
	if _, err := os.Stat(destPath); err == nil {
		if !force {
			return "skipped", nil
		}
		fileExists = true
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for %s: %v", bundlePath, err)
	}

	// Handle symlinks
	if info, ok := fileInfoMap[bundlePath]; ok && info.IsSymlink {
		if fileExists {
			os.Remove(destPath)
		}
		if err := os.Symlink(info.SymlinkTarget, destPath); err != nil {
			return "", fmt.Errorf("failed to create symlink %s: %v", filepath.Base(destPath), err)
		}
		s.cfg.Log(1, "Installed symlink: %s -> %s", destPath, info.SymlinkTarget)
		return "installed", nil
	}

	// Handle regular files
	content, err := cli.BundleReadFile(bundlePath)
	if err != nil || len(content) == 0 {
		s.cfg.Log(1, "File not found in bundle: %s", bundlePath)
		return "", nil
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

	if bundled, _ := cli.IsBundled(); !bundled {
		return nil, fmt.Errorf("install requires a bundled binary (use 'make build')")
	}

	projectRoot := filepath.Dir(filepath.Dir(s.baseDir))

	// Check versions to skip unnecessary reinstalls
	bundledVersion := readBundledVersion()
	installedVersion := readInstalledVersion(s.baseDir)

	if !force && installedVersion != "" && bundledVersion != "" &&
		compareSemver(installedVersion, bundledVersion) >= 0 {
		return &InstallResult{
			VersionSkipped:   true,
			BundledVersion:   bundledVersion,
			InstalledVersion: installedVersion,
			Hint:             "Use force=true to reinstall",
		}, nil
	}

	// Build file info map for symlink detection
	fileInfoMap := buildFileInfoMap()

	var installed, skipped []string
	track := func(relPath, status string) {
		if status == "installed" {
			installed = append(installed, relPath)
		} else if status == "skipped" {
			skipped = append(skipped, relPath)
		}
	}

	// installBundleFiles installs files from a bundle directory to a destination
	installBundleFiles := func(bundleDir, destDir string, mode os.FileMode) error {
		files, _ := cli.BundleListFiles(bundleDir)
		for _, bundlePath := range files {
			fileName := filepath.Base(bundlePath)
			destPath := filepath.Join(destDir, fileName)
			relPath := filepath.Join(bundleDir, fileName)
			status, err := s.installFile(bundlePath, destPath, mode, force, fileInfoMap)
			if err != nil {
				return err
			}
			track(relPath, status)
		}
		return nil
	}

	// 1. Install skills to {project}/.claude/skills/
	skillFiles := []struct{ category, file string }{
		{"skills/ui", "SKILL.md"},
		{"skills/ui-basics", "SKILL.md"},
		{"skills/ui-fast", "SKILL.md"},
		{"skills/ui-thorough", "SKILL.md"},
		{"skills/ui-testing", "SKILL.md"},
		{"skills/ui-testing", "TESTING-TEMPLATE.md"},
		{"skills/ui-themer", "SKILL.md"},
	}
	for _, f := range skillFiles {
		bundlePath := filepath.Join(f.category, f.file)
		destPath := filepath.Join(projectRoot, ".claude", f.category, f.file)
		relPath := filepath.Join(".claude", f.category, f.file)
		status, err := s.installFile(bundlePath, destPath, 0644, force, fileInfoMap)
		if err != nil {
			return nil, err
		}
		track(relPath, status)
	}

	// 2. Install resources to {base_dir}/resources/
	for _, file := range []string{"intro.md", "reference.md", "viewdefs.md", "lua.md", "mcp.md", "ui_audit.md"} {
		bundlePath := filepath.Join("resources", file)
		destPath := filepath.Join(s.baseDir, "resources", file)
		status, err := s.installFile(bundlePath, destPath, 0644, force, fileInfoMap)
		if err != nil {
			return nil, err
		}
		track(bundlePath, status)
	}

	// 3. Install apps to {base_dir}/apps/ (recursively, includes symlinks)
	appFiles, _ := cli.BundleListFilesRecursive("apps")
	for _, bundlePath := range appFiles {
		destPath := filepath.Join(s.baseDir, bundlePath)
		status, err := s.installFile(bundlePath, destPath, 0644, force, fileInfoMap)
		if err != nil {
			return nil, err
		}
		track(bundlePath, status)
	}

	// 4. Install viewdefs to {base_dir}/viewdefs/
	if err := installBundleFiles("viewdefs", filepath.Join(s.baseDir, "viewdefs"), 0644); err != nil {
		return nil, err
	}

	// 5. Install lua files to {base_dir}/lua/
	if err := installBundleFiles("lua", filepath.Join(s.baseDir, "lua"), 0644); err != nil {
		return nil, err
	}

	// 6. Install scripts to {base_dir}/ (executable)
	for _, file := range []string{"mcp", "linkapp"} {
		destPath := filepath.Join(s.baseDir, file)
		status, err := s.installFile(file, destPath, 0755, force, fileInfoMap)
		if err != nil {
			return nil, err
		}
		track(file, status)
	}

	// 7. Install html files to {base_dir}/html/
	if err := installBundleFiles("html", filepath.Join(s.baseDir, "html"), 0644); err != nil {
		return nil, err
	}

	// 8. Install README.md to {base_dir}/
	status, err := s.installFile("README.md", filepath.Join(s.baseDir, "README.md"), 0644, force, fileInfoMap)
	if err != nil {
		return nil, err
	}
	track("README.md", status)

	// 9. Install themes to {base_dir}/html/themes/
	themeFiles, _ := cli.BundleListFiles("html/themes")
	for _, bundlePath := range themeFiles {
		destPath := filepath.Join(s.baseDir, bundlePath)
		status, err := s.installFile(bundlePath, destPath, 0644, force, fileInfoMap)
		if err != nil {
			return nil, err
		}
		track(bundlePath, status)
	}

	// 10. Install patterns to {base_dir}/patterns/
	// CRC: crc-MCPTool.md
	if err := installBundleFiles("patterns", filepath.Join(s.baseDir, "patterns"), 0644); err != nil {
		return nil, err
	}

	// 11. Check for optional external dependencies
	var suggestions []string
	codeSimplifierPath := filepath.Join(projectRoot, ".claude", "agents", "code-simplifier.md")
	if _, err := os.Stat(codeSimplifierPath); os.IsNotExist(err) {
		suggestions = append(suggestions, "Run `claude plugin install code-simplifier` to enable code simplification")
	}

	// 12. Write install manifest with file hashes
	manifest, _ := readManifest(s.baseDir)
	if manifest == nil {
		manifest = &InstallManifest{Files: make(map[string]string)}
	}
	manifest.Version = bundledVersion
	for _, relPath := range installed {
		// .claude/ paths are relative to projectRoot, everything else to baseDir
		var absPath string
		if strings.HasPrefix(relPath, ".claude/") || strings.HasPrefix(relPath, ".claude\\") {
			absPath = filepath.Join(projectRoot, relPath)
		} else {
			absPath = filepath.Join(s.baseDir, relPath)
		}
		if hash, err := computeFileHash(absPath); err == nil {
			manifest.Files[relPath] = hash
		}
	}
	if err := writeManifest(s.baseDir, manifest); err != nil {
		s.cfg.Log(1, "Warning: failed to write install manifest: %v", err)
	}

	return &InstallResult{
		Installed:   installed,
		Skipped:     skipped,
		Suggestions: suggestions,
	}, nil
}

// readBundledVersion extracts version from the bundled README.md.
func readBundledVersion() string {
	content, err := cli.BundleReadFile("README.md")
	if err != nil {
		return ""
	}
	return parseReadmeVersion(content)
}

// readInstalledVersion extracts version from the installed README.md.
func readInstalledVersion(baseDir string) string {
	content, err := os.ReadFile(filepath.Join(baseDir, "README.md"))
	if err != nil {
		return ""
	}
	return parseReadmeVersion(content)
}

// buildFileInfoMap creates a lookup map for bundle file metadata (used for symlink detection).
func buildFileInfoMap() map[string]cli.BundleFileInfo {
	allFiles, err := cli.BundleListFilesWithInfo()
	if err != nil {
		return make(map[string]cli.BundleFileInfo)
	}
	fileInfoMap := make(map[string]cli.BundleFileInfo, len(allFiles))
	for _, f := range allFiles {
		fileInfoMap[f.Name] = f
	}
	return fileInfoMap
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

// InstallManifest records the SHA256 hashes of installed files for smart update.
type InstallManifest struct {
	Version string            `json:"version"`
	Files   map[string]string `json:"files"`
}

// computeFileHash returns "sha256:<hex>" for the file at path, or error.
func computeFileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(h[:]), nil
}

// readManifest reads the install manifest from storage/install-manifest.json.
func readManifest(baseDir string) (*InstallManifest, error) {
	data, err := os.ReadFile(filepath.Join(baseDir, "storage", "install-manifest.json"))
	if err != nil {
		return nil, err
	}
	var m InstallManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// writeManifest writes the install manifest to storage/install-manifest.json.
func writeManifest(baseDir string, m *InstallManifest) error {
	dir := filepath.Join(baseDir, "storage")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "install-manifest.json"), data, 0644)
}

// UpdateResult contains the results of an update operation.
type UpdateResult struct {
	Updated    []string         `json:"updated"`
	Skipped    []string         `json:"skipped"`
	Conflicts  []UpdateConflict `json:"conflicts,omitempty"`
	NewVersion string           `json:"new_version,omitempty"`
}

// UpdateConflict describes a file that couldn't be updated because the user modified it.
type UpdateConflict struct {
	Path        string `json:"path"`
	Reason      string `json:"reason"`
	OldHash     string `json:"old_hash"`
	CurrentHash string `json:"current_hash"`
	MergePath   string `json:"merge_path,omitempty"`
}

// Update performs a smart update using the install manifest for conflict detection.
// Files unchanged by the user are overwritten; modified files get a .merge copy.
func (s *Server) Update() (*UpdateResult, error) {
	if s.baseDir == "" {
		return nil, fmt.Errorf("server not configured (baseDir not set)")
	}

	if bundled, _ := cli.IsBundled(); !bundled {
		return nil, fmt.Errorf("update requires a bundled binary (use 'make build')")
	}

	manifest, err := readManifest(s.baseDir)
	if err != nil || manifest == nil {
		// No manifest — fall back to force install
		result, err := s.Install(true)
		if err != nil {
			return nil, err
		}
		return &UpdateResult{
			Updated:    result.Installed,
			Skipped:    result.Skipped,
			NewVersion: result.BundledVersion,
		}, nil
	}

	projectRoot := filepath.Dir(filepath.Dir(s.baseDir))
	bundledVersion := readBundledVersion()
	fileInfoMap := buildFileInfoMap()

	var updated, skipped []string
	var conflicts []UpdateConflict

	// updateFile handles a single file update with conflict detection.
	updateFile := func(bundlePath, destPath, relPath string, mode os.FileMode) error {
		// Check if file exists on disk
		_, statErr := os.Stat(destPath)
		fileExists := statErr == nil

		if !fileExists {
			// File doesn't exist → install it
			status, err := s.installFile(bundlePath, destPath, mode, true, fileInfoMap)
			if err != nil {
				return err
			}
			if status == "installed" {
				updated = append(updated, relPath)
				// Update manifest hash
				if hash, err := computeFileHash(destPath); err == nil {
					manifest.Files[relPath] = hash
				}
			}
			return nil
		}

		// File exists — check hash against manifest
		manifestHash, inManifest := manifest.Files[relPath]
		currentHash, hashErr := computeFileHash(destPath)

		if !inManifest || hashErr != nil {
			// Not in manifest (new file from older install) → treat as user-modified, skip
			skipped = append(skipped, relPath)
			return nil
		}

		if currentHash == manifestHash {
			// Hash matches manifest → user hasn't changed it → safe to overwrite
			status, err := s.installFile(bundlePath, destPath, mode, true, fileInfoMap)
			if err != nil {
				return err
			}
			if status == "installed" {
				updated = append(updated, relPath)
				if hash, err := computeFileHash(destPath); err == nil {
					manifest.Files[relPath] = hash
				}
			}
			return nil
		}

		// Hash differs → user modified → write .merge file
		mergeDir := destPath + ".merge"
		mergePath := relPath + ".merge"

		// For symlinks, just record conflict without merge file
		if info, ok := fileInfoMap[bundlePath]; ok && info.IsSymlink {
			conflicts = append(conflicts, UpdateConflict{
				Path:        relPath,
				Reason:      "user_modified",
				OldHash:     manifestHash,
				CurrentHash: currentHash,
			})
			return nil
		}

		content, err := cli.BundleReadFile(bundlePath)
		if err != nil || len(content) == 0 {
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(mergeDir), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(mergeDir, content, mode); err != nil {
			return err
		}

		conflicts = append(conflicts, UpdateConflict{
			Path:        relPath,
			Reason:      "user_modified",
			OldHash:     manifestHash,
			CurrentHash: currentHash,
			MergePath:   mergePath,
		})
		return nil
	}

	// Same file iteration order as Install()

	// 1. Skills → {project}/.claude/skills/
	skillFiles := []struct{ category, file string }{
		{"skills/ui", "SKILL.md"},
		{"skills/ui-basics", "SKILL.md"},
		{"skills/ui-fast", "SKILL.md"},
		{"skills/ui-thorough", "SKILL.md"},
		{"skills/ui-testing", "SKILL.md"},
		{"skills/ui-testing", "TESTING-TEMPLATE.md"},
		{"skills/ui-themer", "SKILL.md"},
	}
	for _, f := range skillFiles {
		bundlePath := filepath.Join(f.category, f.file)
		destPath := filepath.Join(projectRoot, ".claude", f.category, f.file)
		relPath := filepath.Join(".claude", f.category, f.file)
		if err := updateFile(bundlePath, destPath, relPath, 0644); err != nil {
			return nil, err
		}
	}

	// 2. Resources
	for _, file := range []string{"intro.md", "reference.md", "viewdefs.md", "lua.md", "mcp.md", "ui_audit.md"} {
		bundlePath := filepath.Join("resources", file)
		destPath := filepath.Join(s.baseDir, "resources", file)
		if err := updateFile(bundlePath, destPath, bundlePath, 0644); err != nil {
			return nil, err
		}
	}

	// 3. Apps (recursive)
	appFiles, _ := cli.BundleListFilesRecursive("apps")
	for _, bundlePath := range appFiles {
		destPath := filepath.Join(s.baseDir, bundlePath)
		if err := updateFile(bundlePath, destPath, bundlePath, 0644); err != nil {
			return nil, err
		}
	}

	// 4. Viewdefs
	vdFiles, _ := cli.BundleListFiles("viewdefs")
	for _, bundlePath := range vdFiles {
		destPath := filepath.Join(s.baseDir, bundlePath)
		if err := updateFile(bundlePath, destPath, filepath.Join("viewdefs", filepath.Base(bundlePath)), 0644); err != nil {
			return nil, err
		}
	}

	// 5. Lua files
	luaFiles, _ := cli.BundleListFiles("lua")
	for _, bundlePath := range luaFiles {
		destPath := filepath.Join(s.baseDir, bundlePath)
		if err := updateFile(bundlePath, destPath, filepath.Join("lua", filepath.Base(bundlePath)), 0644); err != nil {
			return nil, err
		}
	}

	// 6. Scripts (executable)
	for _, file := range []string{"mcp", "linkapp"} {
		destPath := filepath.Join(s.baseDir, file)
		if err := updateFile(file, destPath, file, 0755); err != nil {
			return nil, err
		}
	}

	// 7. HTML files
	htmlFiles, _ := cli.BundleListFiles("html")
	for _, bundlePath := range htmlFiles {
		destPath := filepath.Join(s.baseDir, bundlePath)
		if err := updateFile(bundlePath, destPath, filepath.Join("html", filepath.Base(bundlePath)), 0644); err != nil {
			return nil, err
		}
	}

	// 8. README.md
	if err := updateFile("README.md", filepath.Join(s.baseDir, "README.md"), "README.md", 0644); err != nil {
		return nil, err
	}

	// 9. Themes
	themeFiles, _ := cli.BundleListFiles("html/themes")
	for _, bundlePath := range themeFiles {
		destPath := filepath.Join(s.baseDir, bundlePath)
		if err := updateFile(bundlePath, destPath, bundlePath, 0644); err != nil {
			return nil, err
		}
	}

	// Write updated manifest
	manifest.Version = bundledVersion
	if err := writeManifest(s.baseDir, manifest); err != nil {
		s.cfg.Log(1, "Warning: failed to write install manifest: %v", err)
	}

	return &UpdateResult{
		Updated:    updated,
		Skipped:    skipped,
		Conflicts:  conflicts,
		NewVersion: bundledVersion,
	}, nil
}

// handleUpdate handles the ui_update MCP tool.
func (s *Server) handleUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.state != Running {
		return mcp.NewToolResultError("ui_update requires the server to be running"), nil
	}

	result, err := s.Update()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// handleAPIUpdate handles POST /api/ui_update
func (s *Server) handleAPIUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apiError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	args, err := parseJSONBody(r)
	if err != nil {
		apiError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	_ = args // no arguments needed
	result, err := s.callMCPHandler(s.handleUpdate, nil)
	apiResponse(w, result, err)
}

// handleInstall installs bundled configuration files.
// Spec: mcp.md section 5.6
func (s *Server) handleInstall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check state - server must be running
	if s.state != Running {
		return mcp.NewToolResultError("ui_install requires the server to be running - server may not have started correctly"), nil
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

		// mcp:waitTime() - seconds since agent last responded, or 0 if connected
		// Spec: mcp.md Section 8.3
		L.SetField(mcpTable, "waitTime", L.NewFunction(func(L *lua.LState) int {
			// Note: Called as mcp:waitTime() but we ignore the self argument
			L.Push(lua.LNumber(s.getWaitTime(vendedID)))
			return 1
		}))

		// mcp.sessionId - the current external session ID
		// CRC: crc-MCPTool.md
		L.SetField(mcpTable, "sessionId", lua.LString(s.UiServer.GetSessions().GetInternalID(vendedID)))

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

		// mcp:reinjectThemes() - re-scan themes and update index.html
		// CRC: crc-MCPServer.md
		L.SetField(mcpTable, "reinjectThemes", L.NewFunction(func(L *lua.LState) int {
			if err := InjectThemeBlock(s.baseDir); err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			L.Push(lua.LTrue)
			return 1
		}))

		// Register mcp:subscribe(topic, handler) for publisher integration
		// CRC: crc-MCPSubscribe.md
		s.registerSubscribeMethod(vendedID, mcpTable)

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
				// Use os.Stat to follow symlinks (entry.IsDir() doesn't)
				entryPath := filepath.Join(appsDir, entry.Name())
				if info, err := os.Stat(entryPath); err != nil || !info.IsDir() {
					continue
				}
				initPath := filepath.Join("apps", entry.Name(), "init.lua")
				if _, err := os.Stat(filepath.Join(s.baseDir, initPath)); err == nil {
					if _, err := session.DirectRequireLuaFile(initPath); err != nil {
						return nil, fmt.Errorf("failed to load %s/init.lua: %w", entry.Name(), err)
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
		return mcp.NewToolResultError("no active session - server may not have started correctly"), nil
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

	// Construct URL: baseURL + path (no session ID - cookie handles session binding)
	fullURL := fmt.Sprintf("%s%s", baseURL, path)
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
		return mcp.NewToolResultError("no active session - server may not have started correctly"), nil
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
		return mcp.NewToolResultError("server not running - server may not have started correctly"), nil
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
		return mcp.NewToolResultError("server not configured - server may not have started correctly"), nil
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

// handleTheme handles theme management operations.
func (s *Server) handleTheme(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.mu.RLock()
	baseDir := s.baseDir
	s.mu.RUnlock()

	if baseDir == "" {
		return mcp.NewToolResultError("server not configured"), nil
	}

	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("arguments must be a map"), nil
	}

	action, ok := args["action"].(string)
	if !ok || action == "" {
		return mcp.NewToolResultError("action is required"), nil
	}

	theme, _ := args["theme"].(string)
	app, _ := args["app"].(string)

	var result interface{}
	var err error

	switch action {
	case "list":
		listResult, listErr := ListThemesWithInfo(baseDir)
		if listErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("listing themes: %v", listErr)), nil
		}
		result = listResult

	case "classes":
		// CRC: crc-ThemeManager.md | Seq: seq-theme-audit.md
		themeName, classes, classErr := ResolveThemeClasses(baseDir, theme)
		if classErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("getting theme classes: %v", classErr)), nil
		}
		result = ThemeClassesResult{Theme: themeName, Classes: classes}

	case "audit":
		// CRC: crc-ThemeManager.md | Seq: seq-theme-audit.md
		if app == "" {
			return mcp.NewToolResultError("app is required for audit action"), nil
		}
		result, err = AuditAppTheme(baseDir, app, theme)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("auditing theme: %v", err)), nil
		}

	default:
		return mcp.NewToolResultError(fmt.Sprintf("unknown action: %s (use: list, classes, audit)", action)), nil
	}

	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshaling result: %v", err)), nil
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

// handleAPITheme handles POST /api/ui_theme
func (s *Server) handleAPITheme(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apiError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	args, err := parseJSONBody(r)
	if err != nil {
		apiError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	result, err := s.callMCPHandler(s.handleTheme, args)
	apiResponse(w, result, err)
}

// handleAPIResource handles GET /api/resource/ and /api/resource/{path}
// Serves files from {base_dir}/resources/ with directory listing support
func (s *Server) handleAPIResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		apiError(w, http.StatusMethodNotAllowed, "GET required")
		return
	}

	s.mu.RLock()
	baseDir := s.baseDir
	s.mu.RUnlock()

	// Extract path after /api/resource/
	reqPath := strings.TrimPrefix(r.URL.Path, "/api/resource/")
	reqPath = filepath.Clean(reqPath)

	// Prevent directory traversal
	if strings.Contains(reqPath, "..") {
		apiError(w, http.StatusBadRequest, "invalid path")
		return
	}

	resourceDir := filepath.Join(baseDir, "resources")
	fullPath := filepath.Join(resourceDir, reqPath)

	// Ensure path is within resources directory
	if !strings.HasPrefix(fullPath, resourceDir) {
		apiError(w, http.StatusBadRequest, "invalid path")
		return
	}

	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		apiError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		apiError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if info.IsDir() {
		// Directory listing
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			apiError(w, http.StatusInternalServerError, err.Error())
			return
		}

		type dirEntry struct {
			Name  string `json:"name"`
			IsDir bool   `json:"is_dir"`
			Size  int64  `json:"size,omitempty"`
		}

		var listing []dirEntry
		for _, entry := range entries {
			// Skip hidden files
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			de := dirEntry{
				Name:  entry.Name(),
				IsDir: entry.IsDir(),
			}
			if !entry.IsDir() {
				if fi, err := entry.Info(); err == nil {
					de.Size = fi.Size()
				}
			}
			listing = append(listing, de)
		}

		// Return JSON for curl, HTML for browsers
		userAgent := r.Header.Get("User-Agent")
		if strings.Contains(userAgent, "curl") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"path":    reqPath,
				"entries": listing,
			})
		} else {
			// HTML directory listing
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <title>Resources: /%s</title>
  <style>
    body { font-family: system-ui, sans-serif; max-width: 800px; margin: 2rem auto; padding: 0 1rem; }
    h1 { font-size: 1.5rem; color: #333; }
    ul { list-style: none; padding: 0; }
    li { padding: 0.5rem 0; border-bottom: 1px solid #eee; }
    a { color: #0066cc; text-decoration: none; }
    a:hover { text-decoration: underline; }
    .dir { font-weight: bold; }
    .dir::after { content: "/"; }
    .size { color: #666; font-size: 0.875rem; margin-left: 1rem; }
  </style>
</head>
<body>
  <h1>Resources: /%s</h1>
  <ul>
`, reqPath, reqPath)
			for _, entry := range listing {
				entryPath := entry.Name
				if reqPath != "" && reqPath != "." {
					entryPath = reqPath + "/" + entry.Name
				}
				if entry.IsDir {
					fmt.Fprintf(w, `    <li><a href="/api/resource/%s" class="dir">%s</a></li>`+"\n", entryPath, entry.Name)
				} else {
					fmt.Fprintf(w, `    <li><a href="/api/resource/%s">%s</a><span class="size">%d bytes</span></li>`+"\n", entryPath, entry.Name, entry.Size)
				}
			}
			fmt.Fprintf(w, `  </ul>
</body>
</html>
`)
		}
	} else {
		// Serve file
		content, err := os.ReadFile(fullPath)
		if err != nil {
			apiError(w, http.StatusInternalServerError, err.Error())
			return
		}

		ext := filepath.Ext(fullPath)
		userAgent := r.Header.Get("User-Agent")
		isCurl := strings.Contains(userAgent, "curl")

		// Render markdown as HTML for browsers
		if ext == ".md" && !isCurl {
			var buf bytes.Buffer
			if err := goldmark.Convert(content, &buf); err != nil {
				apiError(w, http.StatusInternalServerError, err.Error())
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <title>%s</title>
  <style>
    body { font-family: system-ui, sans-serif; max-width: 800px; margin: 2rem auto; padding: 0 1rem; line-height: 1.6; }
    h1, h2, h3 { color: #333; }
    code { background: #f4f4f4; padding: 0.2em 0.4em; border-radius: 3px; }
    pre { background: #f4f4f4; padding: 1rem; overflow-x: auto; border-radius: 4px; }
    pre code { background: none; padding: 0; }
    a { color: #0066cc; }
    table { border-collapse: collapse; width: 100%%; }
    th, td { border: 1px solid #ddd; padding: 0.5rem; text-align: left; }
    th { background: #f4f4f4; }
  </style>
</head>
<body>
%s
</body>
</html>
`, filepath.Base(fullPath), buf.String())
			return
		}

		// Set content type based on extension
		switch ext {
		case ".md":
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		case ".json":
			w.Header().Set("Content-Type", "application/json")
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		default:
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}

		w.Write(content)
	}
}

// handleAppReadme handles GET /app/{app}/readme
// Serves the app's README.md as HTML (case-insensitive file lookup, rendered via goldmark)
// CRC: crc-MCPServer.md
func (s *Server) handleAppReadme(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		apiError(w, http.StatusMethodNotAllowed, "GET required")
		return
	}

	s.mu.RLock()
	baseDir := s.baseDir
	s.mu.RUnlock()

	// Extract app name from path: /app/{app}/readme
	path := strings.TrimPrefix(r.URL.Path, "/app/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "readme" {
		apiError(w, http.StatusBadRequest, "expected /app/{app}/readme")
		return
	}
	appName := parts[0]

	// Prevent directory traversal
	if strings.Contains(appName, "..") || strings.Contains(appName, "/") {
		apiError(w, http.StatusBadRequest, "invalid app name")
		return
	}

	appDir := filepath.Join(baseDir, "apps", appName)

	// Check app directory exists
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		apiError(w, http.StatusNotFound, "app not found")
		return
	}

	// Case-insensitive search for readme.md
	entries, err := os.ReadDir(appDir)
	if err != nil {
		apiError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var readmePath string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if name == "readme.md" {
			readmePath = filepath.Join(appDir, entry.Name())
			break
		}
	}

	if readmePath == "" {
		apiError(w, http.StatusNotFound, "readme.md not found")
		return
	}

	// Read and render markdown
	content, err := os.ReadFile(readmePath)
	if err != nil {
		apiError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(content, &buf); err != nil {
		apiError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <title>%s - README</title>
  <style>
    body { font-family: system-ui, sans-serif; max-width: 800px; margin: 2rem auto; padding: 0 1rem; line-height: 1.6; }
    h1, h2, h3 { color: #333; }
    code { background: #f4f4f4; padding: 0.2em 0.4em; border-radius: 3px; }
    pre { background: #f4f4f4; padding: 1rem; overflow-x: auto; border-radius: 4px; }
    pre code { background: none; padding: 0; }
    table { border-collapse: collapse; width: 100%%; }
    th, td { border: 1px solid #ddd; padding: 0.5rem; text-align: left; }
    th { background: #f4f4f4; }
  </style>
</head>
<body>
%s
</body>
</html>
`, appName, buf.String())
}
