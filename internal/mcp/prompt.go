// Package mcp implements the Model Context Protocol server.
// CRC: crc-PromptManager.md
// Spec: prompt-ui.md
package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zot/ui-engine/cli"
)

// PromptOption represents a single option in a prompt.
type PromptOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// PromptResponse is the response from a prompt.
type PromptResponse struct {
	ID    string `json:"id"`
	Value string `json:"value"`
	Label string `json:"label"`
}

// pendingPrompt tracks a prompt waiting for response.
type pendingPrompt struct {
	responseChan chan PromptResponse
}

// PromptManager manages pending prompts and their response channels.
type PromptManager struct {
	mu      sync.Mutex
	pending map[string]*pendingPrompt
	server  *cli.Server
	runtime *cli.LuaRuntime
}

// NewPromptManager creates a new PromptManager.
func NewPromptManager(server *cli.Server, runtime *cli.LuaRuntime) *PromptManager {
	return &PromptManager{
		pending: make(map[string]*pendingPrompt),
		server:  server,
		runtime: runtime,
	}
}

// Prompt displays a prompt in the browser and blocks until response or timeout.
// sessionID is the vended session ID ("1", "2", etc.)
func (pm *PromptManager) Prompt(sessionID string, message string, options []PromptOption, timeout time.Duration) (*PromptResponse, error) {
	// Generate unique prompt ID
	id := uuid.New().String()

	// Create response channel
	pp := &pendingPrompt{
		responseChan: make(chan PromptResponse, 1),
	}

	pm.mu.Lock()
	pm.pending[id] = pp
	pm.mu.Unlock()

	defer func() {
		pm.mu.Lock()
		delete(pm.pending, id)
		pm.mu.Unlock()
	}()

	// Set prompt in Lua app state
	err := pm.setPromptInLua(sessionID, id, message, options)
	if err != nil {
		return nil, fmt.Errorf("failed to set prompt in Lua: %w", err)
	}

	// Wait for response or timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case resp := <-pp.responseChan:
		return &resp, nil
	case <-ctx.Done():
		// Clear the prompt on timeout
		pm.clearPromptInLua(sessionID)
		return nil, fmt.Errorf("prompt timeout after %v", timeout)
	}
}

// Respond is called when the user responds to a prompt.
// This is typically called from Lua via mcp.promptResponse().
func (pm *PromptManager) Respond(id, value, label string) error {
	pm.mu.Lock()
	pp, ok := pm.pending[id]
	pm.mu.Unlock()

	if !ok {
		return fmt.Errorf("prompt %s not found or already responded", id)
	}

	// Send response (non-blocking since channel is buffered)
	select {
	case pp.responseChan <- PromptResponse{ID: id, Value: value, Label: label}:
		return nil
	default:
		return fmt.Errorf("prompt %s already has a response", id)
	}
}

// setPromptInLua sets app.pendingPrompt and switches presenter to "Prompt".
func (pm *PromptManager) setPromptInLua(sessionID, promptID, message string, options []PromptOption) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in setPromptInLua: %v", r)
		}
	}()

	if pm.runtime == nil {
		return fmt.Errorf("Lua runtime not available")
	}

	// Build Lua code to set the prompt
	// Convert options to Lua table literal
	optionsLua := "{"
	for i, opt := range options {
		if i > 0 {
			optionsLua += ", "
		}
		optionsLua += fmt.Sprintf("{label = %q, value = %q}", opt.Label, opt.Value)
	}
	optionsLua += "}"

	code := fmt.Sprintf(`
		-- Set mcp.value to prompt data (value is bound in viewdef, triggers update)
		local promptId = %q
		local opts = %s
		local options = {}
		for _, opt in ipairs(opts) do
			local option = {
				type = "PromptOption",
				label = opt.label,
				value = opt.value,
				respond = function(self)
					mcp.promptResponse(promptId, self.value, self.label)
					mcp.value = nil
				end
			}
			table.insert(options, option)
		end
		mcp.value = {
			isPrompt = true,
			id = promptId,
			message = %q,
			options = options
		}
	`, promptID, optionsLua, message)

	// Execute in session context
	_, err = pm.server.ExecuteInSession(sessionID, func() (interface{}, error) {
		return pm.runtime.LoadCodeDirect("prompt-setup", code)
	})

	return err
}

// clearPromptInLua clears mcp.value (the prompt).
func (pm *PromptManager) clearPromptInLua(sessionID string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in clearPromptInLua: %v", r)
		}
	}()

	if pm.runtime == nil {
		return nil
	}

	code := `mcp.value = nil`

	_, err = pm.server.ExecuteInSession(sessionID, func() (interface{}, error) {
		return pm.runtime.LoadCodeDirect("prompt-clear", code)
	})

	return err
}
