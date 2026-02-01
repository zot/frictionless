package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

// TestAuditRealAppConsole verifies factory-created methods are not flagged as dead.
// This integration test runs against the real app-console app.
func TestAuditRealAppConsole(t *testing.T) {
	baseDir := "../../.ui"
	appLuaPath := filepath.Join(baseDir, "apps", "app-console", "app.lua")

	if _, err := os.Stat(appLuaPath); os.IsNotExist(err) {
		t.Skip("Skipping: app-console not found at " + appLuaPath)
	}

	result, err := AuditApp(baseDir, "app-console")
	if err != nil {
		t.Fatalf("AuditApp error: %v", err)
	}

	// Methods created by makeCollapsible factory should not be flagged
	factoryMethods := map[string]bool{
		"AppInfo:toggleKnownIssues":    true,
		"AppInfo:knownIssuesHidden":    true,
		"AppInfo:knownIssuesIcon":      true,
		"AppInfo:toggleFixedIssues":    true,
		"AppInfo:fixedIssuesHidden":    true,
		"AppInfo:fixedIssuesIcon":      true,
		"AppInfo:toggleGaps":           true,
		"AppInfo:gapsHidden":           true,
		"AppInfo:gapsIcon":             true,
		"AppInfo:toggleRequirements":   true,
		"AppInfo:requirementsHidden":   true,
		"AppInfo:requirementsIcon":     true,
	}

	for _, v := range result.Violations {
		if v.Type == "dead_method" && factoryMethods[v.Detail] {
			t.Errorf("Factory-created method incorrectly flagged as dead: %s", v.Detail)
		}
	}

	t.Logf("Audit complete: %d violations, %d dead methods", len(result.Violations), result.Summary.DeadMethods)
	for _, v := range result.Violations {
		t.Logf("  %s: %s (%s)", v.Type, v.Detail, v.Location)
	}
}
