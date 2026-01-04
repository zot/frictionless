// Package mcp implements the Model Context Protocol server.
// CRC: crc-MCPResource.md
// Spec: mcp.md
// Sequence: seq-mcp-get-state.md
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	luajson "github.com/layeh/gopher-json"
	"github.com/mark3labs/mcp-go/mcp"
	lua "github.com/yuin/gopher-lua"
	"github.com/zot/ui-engine/cli"
)

func (s *Server) registerResources() {
	// ui://state (defaults to session 1)
	s.mcpServer.AddResource(mcp.NewResource("ui://state", "Current Session State",
		mcp.WithResourceDescription("Current JSON state of session 1 (Variable 1)"),
		mcp.WithMIMEType("application/json"),
	), s.handleGetStateResource)

	// ui://state/{sessionId}
	s.mcpServer.AddResource(mcp.NewResource("ui://state/{sessionId}", "Session State",
		mcp.WithResourceDescription("Current JSON state of the session (Variable 1)"),
		mcp.WithMIMEType("application/json"),
	), s.handleGetStateResource)

	// ui://{path} - Generic resource server for static content
	s.mcpServer.AddResource(mcp.NewResource("ui://{path}", "Static Resource",
		mcp.WithResourceDescription("Static documentation or pattern resource"),
	), s.handleGetStaticResource)

	// Explicitly register core docs for discovery
	s.mcpServer.AddResource(mcp.NewResource("ui://reference", "UI Platform Reference",
		mcp.WithResourceDescription("Main entry point for UI platform documentation"),
		mcp.WithMIMEType("text/markdown"),
	), s.handleGetStaticResource)

	s.mcpServer.AddResource(mcp.NewResource("ui://viewdefs", "Viewdef Syntax",
		mcp.WithResourceDescription("Guide to ui-* attributes and path syntax"),
		mcp.WithMIMEType("text/markdown"),
	), s.handleGetStaticResource)

	s.mcpServer.AddResource(mcp.NewResource("ui://lua", "Lua API Guide",
		mcp.WithResourceDescription("Lua API, class patterns, and global objects"),
		mcp.WithMIMEType("text/markdown"),
	), s.handleGetStaticResource)

	s.mcpServer.AddResource(mcp.NewResource("ui://mcp", "MCP Agent Guide",
		mcp.WithResourceDescription("Guide for AI agents to build apps"),
		mcp.WithMIMEType("text/markdown"),
	), s.handleGetStaticResource)

	// Prompt UI resources
	// Spec: prompt-ui.md
	s.mcpServer.AddResource(mcp.NewResource("ui://prompt/viewdefs", "Prompt Viewdef Locations",
		mcp.WithResourceDescription("Editable viewdefs for customizing permission prompt UI"),
		mcp.WithMIMEType("text/markdown"),
	), s.handleGetPromptViewdefsResource)

	s.mcpServer.AddResource(mcp.NewResource("ui://permissions/history", "Permission Decision History",
		mcp.WithResourceDescription("Log of recent user permission choices for pattern analysis"),
		mcp.WithMIMEType("application/json"),
	), s.handleGetPermissionsHistoryResource)

	// Debug resource for variable inspection
	s.mcpServer.AddResource(mcp.NewResource("ui://variables", "Variable Tree",
		mcp.WithResourceDescription("Topologically sorted array of all tracked variables with their IDs, parent IDs, types, values, and properties"),
		mcp.WithMIMEType("application/json"),
	), s.handleGetVariablesResource)
}

// handleGetStaticResource serves static documentation or pattern resources.
// Spec: mcp.md
// CRC: crc-MCPResource.md
func (s *Server) handleGetStaticResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := request.Params.URI
	path := strings.TrimPrefix(uri, "ui://")

	s.mu.RLock()
	baseDir := s.baseDir
	s.mu.RUnlock()

	// Clean path to prevent directory traversal
	cleanPath := filepath.Clean(path)
	if strings.HasPrefix(cleanPath, "..") {
		return nil, fmt.Errorf("Invalid resource path")
	}

	var content []byte
	var err error
	found := false

	// 1. Try file system if configured
	if baseDir != "" {
		fullPath := filepath.Join(baseDir, "resources", cleanPath+".md")
		// Try with .md extension first
		if _, err := os.Stat(fullPath); err == nil {
			content, err = os.ReadFile(fullPath)
			found = (err == nil)
		}

		if !found {
			// Try exact match
			fullPath = filepath.Join(baseDir, "resources", cleanPath)
			if _, err := os.Stat(fullPath); err == nil {
				content, err = os.ReadFile(fullPath)
				found = (err == nil)
			}
		}
	}

	// 2. Try bundle if not found in FS (or server not configured)
	if !found {
		// Try with .md extension
		content, err = cli.BundleReadFile("resources/" + cleanPath + ".md")
		if err != nil {
			// Try exact match
			content, err = cli.BundleReadFile("resources/" + cleanPath)
		}

		if err != nil {
			return nil, fmt.Errorf("Resource not found: %s", path)
		}
	}

	mimeType := "text/markdown"
	// Heuristic: if we requested a .md file or resolved to one, it's markdown.
	// But bundle.ReadFile doesn't return the resolved name.
	// We can assume markdown if we're not sure, or check the extension of cleanPath.
	// If cleanPath doesn't have .md, and we found it, it might be .md or plain.
	// Since all our core docs are .md, defaulting to markdown is reasonable.
	// But if cleanPath has .css or .js, we should respect it.
	ext := filepath.Ext(cleanPath)
	if ext != "" && ext != ".md" {
		mimeType = "text/plain" // Or specific types if we care
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: mimeType,
			Text:     string(content),
		},
	}, nil
}

func (s *Server) handleGetStateResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := request.Params.URI

	// Simple parsing of URI to get sessionId
	var sessionID string
	if uri == "ui://state" {
		sessionID = s.currentVendedID
	} else {
		n, err := fmt.Sscanf(uri, "ui://state/%s", &sessionID)
		if err != nil || n != 1 {
			return nil, fmt.Errorf("invalid URI format")
		}
	}

	session := s.UiServer.GetLuaSession(sessionID)
	if session == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Get state from Lua: mcp.state if set, else mcp.value
	L := session.State
	mcpTable := L.GetGlobal("mcp")
	if mcpTable.Type() != lua.LTTable {
		return nil, fmt.Errorf("mcp global not found")
	}

	// Try mcp.state first, then mcp.value
	stateValue := L.GetField(mcpTable, "state")
	if stateValue == lua.LNil {
		stateValue = L.GetField(mcpTable, "value")
	}

	// Convert Lua value to Go value using gopher-json
	goValue := luajson.DecodeValue(L, stateValue)

	// Marshal to JSON
	jsonVal, err := json.Marshal(goValue)
	if err != nil {
		return nil, fmt.Errorf("marshaling error: %v", err)
	}

	result := map[string]interface{}{
		"sessionId": sessionID,
		"value":     json.RawMessage(jsonVal),
	}

	content, _ := json.MarshalIndent(result, "", "  ")

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(content),
		},
	}, nil
}

// handleGetPromptViewdefsResource returns information about editable viewdefs for the prompt UI.
// Spec: prompt-ui.md
func (s *Server) handleGetPromptViewdefsResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	s.mu.RLock()
	baseDir := s.baseDir
	s.mu.RUnlock()

	content := `# UI MCP Prompt Viewdefs

Viewdefs are HTML templates that control how UI elements are rendered.
You can edit these files to customize the appearance and behavior of permission prompts.

## Locations

`
	if baseDir != "" {
		promptViewdef := filepath.Join(baseDir, "viewdefs", "Prompt.DEFAULT.html")
		content += fmt.Sprintf("- `%s` - Permission prompt dialog\n", promptViewdef)
		content += fmt.Sprintf("- `%s/viewdefs/Feedback.DEFAULT.html` - Default app UI (if customized)\n", baseDir)
	} else {
		content += "- `.claude/ui/viewdefs/Prompt.DEFAULT.html` - Permission prompt dialog\n"
		content += "- `.claude/ui/viewdefs/Feedback.DEFAULT.html` - Default app UI (if customized)\n"
	}

	content += `
## Editing Tips

- Use Shoelace components (sl-button, sl-dialog, sl-input, etc.)
- Use ` + "`ui-value=\"path\"`" + ` for data binding
- Use ` + "`ui-action=\"method()\"`" + ` for button actions
- Changes take effect on next render

## Example: Customizing the Prompt Dialog

` + "```html" + `
<template>
  <div class="prompt-overlay">
    <sl-dialog open label="Permission Request">
      <p ui-value="pendingPrompt.message"></p>
      <div ui-viewlist="pendingPrompt.options">
        <sl-button ui-action="respondToPrompt(_)">
          <span ui-value="label"></span>
        </sl-button>
      </div>
    </sl-dialog>
  </div>
</template>
` + "```" + `

When editing, you can add custom styling, rearrange options, add keyboard shortcuts,
or modify the dialog behavior to match your preferences.
`

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "text/markdown",
			Text:     content,
		},
	}, nil
}

// handleGetPermissionsHistoryResource returns the permission decision log.
// Spec: prompt-ui.md
func (s *Server) handleGetPermissionsHistoryResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	s.mu.RLock()
	baseDir := s.baseDir
	s.mu.RUnlock()

	var decisions []map[string]interface{}

	if baseDir != "" {
		logFile := filepath.Join(baseDir, "permissions.log")
		content, err := os.ReadFile(logFile)
		if err == nil {
			// Parse JSONL format (one JSON object per line)
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				var entry map[string]interface{}
				if err := json.Unmarshal([]byte(line), &entry); err == nil {
					decisions = append(decisions, entry)
				}
			}
		}
	}

	// Limit to last 50 decisions
	if len(decisions) > 50 {
		decisions = decisions[len(decisions)-50:]
	}

	result := map[string]interface{}{
		"decisions": decisions,
		"analysis_hints": `Analyze patterns to proactively improve the permission UI:
- If user frequently clicks "Always allow X", suggest making it the default option
- If user always allows certain tool patterns, offer to add them to auto-allow
- If user adds custom options (via viewdef), note which ones get used
- Suggest viewdef refinements based on observed preferences

Example: "I notice you always allow git commands. Want me to update the prompt viewdef to show 'Always allow git' as the primary option?"`,
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(resultJSON),
		},
	}, nil
}

// handleGetVariablesResource returns all variables in topological order.
func (s *Server) handleGetVariablesResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	variables, err := s.getDebugVariables("1")
	if err != nil {
		return nil, err
	}

	resultJSON, _ := json.MarshalIndent(variables, "", "  ")

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(resultJSON),
		},
	}, nil
}

// getDebugVariables returns all variables in topological order for a session.
// This is used by both the MCP resource and the debug page.
func (s *Server) getDebugVariables(sessionID string) ([]cli.DebugVariable, error) {
	session := s.UiServer.GetLuaSession(sessionID)
	if session == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	tracker := session.GetTracker()
	if tracker == nil {
		return nil, fmt.Errorf("tracker not found")
	}

	// Get all variables and sort topologically (parents before children)
	allVars := tracker.Variables()

	// Build map for quick lookup
	varMap := make(map[int64]*cli.Variable)
	for _, v := range allVars {
		varMap[v.ID] = v
	}

	// Topological sort: BFS from roots
	var sorted []*cli.Variable
	visited := make(map[int64]bool)

	// Start with root variables (parentID == 0)
	queue := make([]*cli.Variable, 0)
	for _, v := range allVars {
		if v.ParentID == 0 {
			queue = append(queue, v)
		}
	}

	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]

		if visited[v.ID] {
			continue
		}
		visited[v.ID] = true
		sorted = append(sorted, v)

		// Add children to queue
		for _, childID := range v.ChildIDs {
			if child := varMap[childID]; child != nil && !visited[childID] {
				queue = append(queue, child)
			}
		}
	}

	// Add any orphans (shouldn't happen, but be safe)
	for _, v := range allVars {
		if !visited[v.ID] {
			sorted = append(sorted, v)
		}
	}

	// Convert to DebugVariable
	result := make([]cli.DebugVariable, len(sorted))
	for i, v := range sorted {
		info := cli.DebugVariable{
			ID:         v.ID,
			ParentID:   v.ParentID,
			Type:       v.Properties["type"],
			Path:       v.Properties["path"],
			Properties: v.Properties,
			ChildIDs:   v.ChildIDs,
		}

		// Get value - convert to interface{} for JSON serialization
		if v.Value != nil {
			info.Value = tracker.ToValueJSON(v.Value)
		}

		result[i] = info
	}

	return result, nil
}
