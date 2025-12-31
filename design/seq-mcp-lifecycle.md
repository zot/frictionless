# Sequence: MCP Server Lifecycle

**Source Spec:** specs/mcp.md

## Participants
- AI Agent: External AI assistant (MCP Client)
- MCPServer: Protocol handler and lifecycle FSM
- MCPTool: Tool handler dispatcher
- LuaRuntime: Embedded Lua VM manager
- HTTPServer: The UI platform's HTTP service
- OS: Operating system services (filesystem, browser launch)

## Scenario 1: Initial Configuration & Setup
The AI agent initializes the environment before starting the server.

```
     ┌────────┐             ┌─────────┐             ┌───────┐             ┌──────────┐           ┌────┐
     │AI Agent│             │MCPServer│             │MCPTool│             │LuaRuntime│           │ OS │
     └────┬───┘             └────┬────┘             └───┬───┘             └─────┬────┘           └─┬──┘
          │Call("ui_configure", {base_dir})             │                       │                  │
          │─────────────────────>│                      │                       │                  │
          │                      │ Handle("ui_configure")                       │                  │
          │                      │─────────────────────>│                       │                  │
          │                      │                      │ CreateDir(base_dir)   │                  │
          │                      │                      │─────────────────────────────────────────>│
          │                      │                      │                       │                  │
          │                      │                      │ RedirectIO(base_dir)  │                  │
          │                      │                      │──────────────────────>│                  │
          │                      │                      │                       │                  │
          │                      │                      │   LoadConfig()        │                  │
          │                      │                      │─────────────────────────────────────────>│
          │                      │                      │                       │                  │
          │                      │  SetState(CONFIGURED)│                       │                  │
          │                      │<─────────────────────│                       │                  │
          │                      │                      │                       │                  │
          │                      │                      │ InstallAgentFiles()   │                  │
          │                      │                      │─────────────────────────────────────────>│
          │                      │                      │         [if missing]  │                  │
          │                      │                      │                       │                  │
          │                      │ SendNotification("agent_installed")          │                  │
          │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                      │ [per installed file]  │                  │
          │                      │                      │                       │                  │
          │       Success        │                      │                       │                  │
          │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                      │                       │                  │
     ┌────┴───┐             ┌────┴────┐             ┌───┴───┐             ┌─────┴────┘           ┌─┴──┐
     │AI Agent│             │MCPServer│             │MCPTool│             │LuaRuntime│           │ OS │
     └────────┘             └─────────┘             └───┴───┘             └──────────┘           └────┘
```

## Scenario 1a: Agent File Installation (Detail)
Shows the conditional agent file extraction during configuration.

```
     ┌───────┐                    ┌────┐                    ┌──────┐
     │MCPTool│                    │ OS │                    │Bundle│
     └───┬───┘                    └─┬──┘                    └──┬───┘
         │                          │                          │
         │ projectRoot = parent(base_dir)                      │
         │────┐                     │                          │
         │    │                     │                          │
         │<───┘                     │                          │
         │                          │                          │
         │ Stat(".claude/agents/ui-builder.md")                │
         │─────────────────────────>│                          │
         │   [file exists?]         │                          │
         │                          │                          │
         │ ────────────────────────────────────────────────────│
         │ alt [file missing]       │                          │
         │ ─────────────────────────│──────────────────────────│
         │ │ MkdirAll(".claude/agents/")                       │
         │ │───────────────────────>│                          │
         │ │                        │                          │
         │ │ ReadFile("agents/ui-builder.md")                  │
         │ │───────────────────────────────────────────────────>
         │ │                        │           content        │
         │ │<──────────────────────────────────────────────────│
         │ │                        │                          │
         │ │ WriteFile(".claude/agents/ui-builder.md")         │
         │ │───────────────────────>│                          │
         │ │                        │                          │
         │ │ SendNotification("agent_installed", {...})        │
         │ │───────┐                │                          │
         │ │       │ to AI Agent    │                          │
         │ │<──────┘                │                          │
         │ ─────────────────────────│──────────────────────────│
         │                          │                          │
     ┌───┴───┐                    ┌─┴──┐                    ┌──┴───┐
     │MCPTool│                    │ OS │                    │Bundle│
     └───────┘                    └────┘                    └──────┘
```

**Notes:**
- Agent files installed to `.claude/agents/` relative to **parent** of `base_dir`
- `base_dir` is typically `.ui-mcp/`, so agents install to project root's `.claude/agents/`
- Notification params: `{"file": "ui-builder.md", "path": ".claude/agents/ui-builder.md"}`
- No-op if file already exists (no notification sent)

## Scenario 2: Server Startup
The AI agent starts the HTTP server after configuration.

```
     ┌────────┐             ┌─────────┐             ┌───────┐             ┌──────────┐           ┌────┐
     │AI Agent│             │MCPServer│             │MCPTool│             │HTTPServer│           │ OS │
     └────┬───┘             └────┬────┘             └───┬───┘             └─────┬────┘           └─┬──┘
          │ Call("ui_start")     │                      │                       │                  │
          │─────────────────────>│                      │                       │                  │
          │                      │  Handle("ui_start")  │                       │                  │
          │                      │─────────────────────>│                       │                  │
          │                      │                      │                       │                  │
          │                      │                      │ SelectPort(0)         │                  │
          │                      │                      │─────────────────────────────────────────>│
          │                      │                      │                       │                  │
          │                      │                      │    Start(port)        │                  │
          │                      │                      │──────────────────────>│                  │
          │                      │                      │                       │                  │
          │                      │   SetState(RUNNING)  │                       │                  │
          │                      │<─────────────────────│                       │                  │
          │                      │                      │                       │                  │
          │     Success(URL)     │                      │                       │                  │
          │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                      │                       │                  │
     ┌────┴───┐             ┌────┴────┐             ┌───┴───┐             ┌─────┴────┘           ┌─┴──┐
     │AI Agent│             │MCPServer│             │MCPTool│             │HTTPServer│           │ OS │
     └────────┘             └─────────┘             └───┴───┘             └──────────┘           └────┘
```

## Scenario 3: Opening Browser
The AI agent instructs the system to open a browser to the session.

```
     ┌────────┐             ┌─────────┐             ┌───────┐             ┌────┐
     │AI Agent│             │MCPServer│             │MCPTool│             │ OS │
     └────┬───┘             └────┬────┘             └───┬───┘             └─┬──┘
          │Call("ui_open_browser", {sessionId, conserve})                 │
          │─────────────────────>│                      │                   │
          │                      │Handle("ui_open_browser")                 │
          │                      │─────────────────────>│                   │
          │                      │                      │                   │
          │                      │                      │ ConstructURL()    │
          │                      │                      │────┐              │
          │                      │                      │    │              │
          │                      │                      │<───┘              │
          │                      │                      │                   │
          │                      │                      │  xdg-open(URL)    │
          │                      │                      │──────────────────>│
          │                      │                      │                   │
          │       Success        │                      │                   │
          │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│                      │                   │
     ┌────┴───┐             ┌────┴────┐             ┌───┴───┐             ┌─┴──┐
     │AI Agent│             │MCPServer│             │MCPTool│             │ OS │
     └────────┘             └─────────┘             └───┴───┘             └────┘
```
