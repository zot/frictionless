# MCPResource

**Source Spec:** mcp.md
**Requirements:** R9, R10

## Responsibilities

### Knows
- name: Resource identifier (URI pattern)
- description: Human-readable description
- mimeType: Content type of resource data
- baseDir: Server base directory for file access

### Does
- getSessionState: Return current session state as JSON (ui://state, uses server's currentVendedID)
- getVariables: Return all tracked variables in topological order (ui://variables)
- getStaticResource: Serve static documentation from resources/ dir (ui://{path})

## Collaborators

- MCPServer: Registers and invokes resources
- LuaRuntime: Accesses mcp.state/mcp.value for state resources
- VariableTracker: Provides variable tree for debug resource
- OS: Reads static files

## Sequences

- seq-mcp-get-state.md: State resource queries
