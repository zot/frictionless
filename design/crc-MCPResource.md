# MCPResource

**Source Spec:** interfaces.md

## Responsibilities

### Knows
- name: Resource identifier
- description: Human-readable description
- mimeType: Content type of resource data

### Does
- listPresenterTypes: Return available presenter types and properties
- listViewdefs: Return available TYPE.VIEW viewdefs and their bindings
- getSessionState: Return current session variable tree
- getPendingMessages: Return queued user messages/requests
- getPresenterState: Return specific presenter data

## Collaborators

- MCPServer: Registers and invokes resources
- ViewdefStore: Viewdef listing
- VariableStore: Session state
- Session: Current session context

## Sequences

- seq-mcp-create-session.md: Resource queries during setup
- seq-mcp-receive-event.md: Pending message retrieval
