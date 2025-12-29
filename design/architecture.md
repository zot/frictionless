# Architecture

**Entry point to the design - shows how design elements are organized into logical systems**

**Sources**: mcp.md, prompt-ui.md

---

## MCP Integration System

**Purpose**: AI assistant integration via Model Context Protocol

**Design Elements:**
- crc-MCPServer.md
- crc-MCPResource.md
- crc-MCPTool.md
- seq-mcp-lifecycle.md
- seq-mcp-create-session.md
- seq-mcp-create-presenter.md
- seq-mcp-receive-event.md
- seq-mcp-run.md
- seq-mcp-get-state.md
- seq-mcp-notify.md
- test-MCP.md

---

## Permission Prompt System

**Purpose**: Browser-based UI for Claude Code permission prompts via viewdefs

**Design Elements:**
- crc-PromptManager.md
- crc-PromptHTTPServer.md
- crc-Server.md
- crc-PromptResponseCallback.md
- crc-Prompt.md (Lua model)
- crc-PromptOption.md (Lua model)
- crc-PromptViewdef.md
- crc-PermissionHook.md
- seq-prompt-flow.md
- seq-prompt-server-startup.md
- ui-prompt-modal.md
- test-Prompt.md

---

## Hook Management System

**Purpose**: CLI for installing/managing Claude Code permission hooks

**Design Elements:**
- crc-HookCLI.md
- seq-hook-install.md

---

## MCP Resources

**Purpose**: Expose UI metadata and history to AI agents

**Design Elements:**
- crc-ViewdefsResource.md
- crc-PermissionHistoryResource.md

---

## Integration with ui-engine

This MCP server integrates with the ui-engine project:
- Uses `internal/lua/runtime.go` for Lua execution
- Creates sessions via ui-engine's session management
- Accesses variable state through ui-engine's variable store
- Delivers viewdefs through ui-engine's viewdef system
- Extends Server with Prompt() method for permission prompts

---

*This file serves as the architectural "main program" - start here to understand the design structure*
