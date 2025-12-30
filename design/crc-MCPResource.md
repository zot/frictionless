# MCPResource

**Source Spec:** mcp.md

## Responsibilities

### Knows
- name: Resource identifier (URI pattern)
- description: Human-readable description
- mimeType: Content type of resource data
- baseDir: Server base directory for file access

### Does
- getSessionState: Return current session state as JSON (ui://state, ui://state/{sessionId})
- getVariables: Return all tracked variables in topological order (ui://variables)
- getStaticResource: Serve static documentation from resources/ dir (ui://{path})
- getPromptViewdefs: Return markdown about editable prompt viewdefs (ui://prompt/viewdefs)
- getPermissionsHistory: Return JSON of recent permission decisions (ui://permissions/history)

## Collaborators

- MCPServer: Registers and invokes resources
- LuaRuntime: Accesses mcp.state/mcp.value for state resources
- VariableTracker: Provides variable tree for debug resource
- OS: Reads static files and permissions.log

## Sequences

- seq-mcp-get-state.md: State resource queries
