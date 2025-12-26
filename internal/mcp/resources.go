// Package mcp implements the Model Context Protocol server.
// CRC: crc-MCPResource.md
// Spec: interfaces.md
// Sequence: seq-mcp-get-state.md
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/zot/ui-engine/internal/bundle"
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
		content, err = bundle.ReadFile("resources/" + cleanPath + ".md")
		if err != nil {
			// Try exact match
			content, err = bundle.ReadFile("resources/" + cleanPath)
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
		sessionID = "1"
	} else {
		n, err := fmt.Sscanf(uri, "ui://state/%s", &sessionID)
		if err != nil || n != 1 {
			return nil, fmt.Errorf("invalid URI format")
		}
	}

	session, ok := s.runtime.GetLuaSession(sessionID)
	if !ok {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	tracker := session.GetTracker()
	if tracker == nil {
		return nil, fmt.Errorf("tracker not found")
	}

	var result interface{}

	// 1. Try to use the explicit MCP state variable
	if session.McpStateID != 0 {
		v := tracker.GetVariable(session.McpStateID)
		if v != nil {
			val := v.NavigationValue()
			jsonVal, err := tracker.ToValueJSONBytes(val)
			if err == nil {
				result = map[string]interface{}{
					"id":         v.ID,
					"properties": v.Properties,
					"value":      json.RawMessage(jsonVal),
				}
			}
		}
	}

	// 2. Fallback to Variable 1 (App)
	if result == nil {
		v1 := tracker.GetVariable(1)
		if v1 == nil {
			return nil, fmt.Errorf("variable 1 not found")
		}

		val := v1.NavigationValue()
		jsonVal, err := tracker.ToValueJSONBytes(val)
		if err != nil {
			return nil, fmt.Errorf("marshaling error: %v", err)
		}

		result = map[string]interface{}{
			"id":         1,
			"properties": v1.Properties,
			"value":      json.RawMessage(jsonVal),
		}
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
