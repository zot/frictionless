# Sequence: MCP Create Session

**Source Spec:** interfaces.md
**Use Case:** AI assistant creating a UI session via MCP

## Participants

- AIClient: AI assistant (e.g., Claude)
- MCPServer: MCP protocol handler
- MCPTool: Tool executor
- SessionManager: Session management
- VariableStore: Variable storage

## Sequence

```
     AIClient             MCPServer               MCPTool           SessionManager         VariableStore
        |                      |                      |                      |                      |
        |---tools/call-------->|                      |                      |                      |
        |   (create_session)   |                      |                      |                      |
        |                      |                      |                      |                      |
        |                      |---dispatch---------->|                      |                      |
        |                      |   (create_session)   |                      |                      |
        |                      |                      |                      |                      |
        |                      |                      |---createSession()--->|                      |
        |                      |                      |                      |                      |
        |                      |                      |                      |---generateId()------>|
        |                      |                      |                      |                      |
        |                      |                      |                      |---createRootVar()--->|
        |                      |                      |                      |                      |
        |                      |                      |                      |                      |---create(1)-->
        |                      |                      |                      |                      |
        |                      |                      |                      |<--session------------|
        |                      |                      |                      |                      |
        |                      |                      |<--{sessionId, url}---|                      |
        |                      |                      |                      |                      |
        |                      |<--result-------------|                      |                      |
        |                      |                      |                      |                      |
        |<--{sessionId,--------|                      |                      |                      |
        |    url: "http://..."}|                      |                      |                      |
        |                      |                      |                      |                      |
        |     [AI can now share URL with user]        |                      |                      |
        |                      |                      |                      |                      |
```

## Notes

- AI calls MCP tool to create session
- Session ID and full URL returned
- AI can share URL with user to open browser
- Session persists until cleanup
- AI can now create presenters and viewdefs
