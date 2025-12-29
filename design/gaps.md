# Gap Analysis

**Date:** 2025-12-29
**CRC Cards:** 13 | **Sequences:** 10 | **UI Specs:** 1 | **Test Designs:** 2

## Summary

**Status:** Green
**Type A (Critical):** 0
**Type B (Quality):** 0
**Type C (Enhancements):** 1

---

## Type C Issues (Enhancements)

### C1: Implementation Pending

**Issue:** Permission prompt system implementation not started.

**Expected Files:**
- `internal/prompt/manager.go` - PromptManager
- `internal/prompt/server.go` - PromptHTTPServer
- `internal/prompt/prompt.go` - Server.Prompt() extension
- `internal/prompt/callback.go` - _G.promptResponse callback
- `.ui-mcp/viewdefs/Prompt.DEFAULT.html` - Prompt viewdef
- `cmd/hooks.go` - HookCLI subcommands
- `internal/mcp/resources.go` - MCP resources (viewdefs, permissions history)

**Recommendation:** Implement following test-Prompt.md specifications.

**Status:** Open (by design - implementation phase)

---

## Coverage Summary

**CRC Responsibilities Coverage:**

| System | CRC Cards | Notes |
|--------|-----------|-------|
| MCP Integration | 3 | Partially implemented |
| Permission Prompt | 7 | Design complete, implementation pending |
| Hook Management | 1 | Design complete |
| MCP Resources | 2 | Design complete |

**Sequences Coverage:** 10/10 (100%)
- MCP: 7 sequences
- Prompt: 3 sequences (flow, startup, hook install)

**Test Designs Coverage:** 2/2 (100%)
- test-MCP.md - MCP functionality
- test-Prompt.md - Permission prompt system

**UI Specs Coverage:** 1/1 (100%)
- ui-prompt-modal.md - Prompt viewdef specification

---

## Artifact Verification

### Sequence References Valid
- **Status:** PASS
- All CRC cards reference existing sequence files
- All sequences reference existing participants

### Complex Behaviors Have Sequences
- **Status:** PASS
- seq-prompt-flow.md: Full prompt lifecycle
- seq-prompt-server-startup.md: Server initialization
- seq-hook-install.md: CLI hook installation

### Collaborator Format Valid
- **Status:** PASS
- All collaborators reference CRC card names or external components
- No orphan references

### Architecture Updated
- **Status:** PASS
- All 13 CRC cards appear in architecture.md
- 4 systems defined: MCP Integration, Permission Prompt, Hook Management, MCP Resources

### Traceability Updated
- **Status:** PASS
- Level 1-2 mapping complete for both specs
- Level 2-3 implementation checkboxes ready for tracking

### Test Designs Exist
- **Status:** PASS
- test-MCP.md and test-Prompt.md cover all testable components

---

## Design Changes from Previous Version

**Replaced custom WebSocket protocol with viewdef approach:**
- Removed: crc-PromptModal.md (used custom WS messages)
- Added: crc-PromptViewdef.md (uses variable binding)
- Added: crc-Server.md (Prompt() method)
- Added: crc-PromptResponseCallback.md (Lua-Go bridge)

**Added new components:**
- crc-HookCLI.md (hooks install/uninstall/status)
- crc-ViewdefsResource.md (ui://viewdefs)
- crc-PermissionHistoryResource.md (ui://permissions/history)
- seq-hook-install.md (CLI installation flow)

**Key design principle:** No custom protocol messages. Uses ui-engine's existing viewdef and variable binding system.

---

## Quality Checklist

**Completeness:**
- [x] All CRC cards analyzed (13)
- [x] All sequences analyzed (10)
- [x] All UI specs analyzed (1)
- [x] Test designs cover all systems

**Artifact Verification:**
- [x] Sequence references valid
- [x] Complex behaviors have sequences
- [x] Collaborators are CRC card names
- [x] All CRCs in architecture.md
- [x] All CRCs in traceability.md
- [x] Test designs exist for testable components

**Consistency:**
- [x] Viewdef-based approach consistent across all artifacts
- [x] No references to deprecated custom protocol

---

## Recommended Implementation Order

1. **PromptManager** - Core channel/timeout management
2. **PromptResponseCallback** - Lua-Go bridge
3. **Server.Prompt()** - Orchestration method
4. **PromptHTTPServer** - HTTP endpoint
5. **Prompt.DEFAULT.html** - Viewdef template
6. **MCP Resources** - ui://viewdefs, ui://permissions/history
7. **HookCLI** - Installation commands
