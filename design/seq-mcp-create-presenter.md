# Sequence: MCP Create Presenter

**Source Spec:** interfaces.md
**Use Case:** AI creating presenter and viewdef via MCP

## Participants

- AIClient: AI assistant
- MCPTool: Tool executor
- ViewdefStore: Viewdef storage
- VariableStore: Variable storage
- LuaRuntime: Lua execution (if logic provided)

## Sequence

```
     AIClient              MCPTool            ViewdefStore          VariableStore           LuaRuntime
        |                      |                      |                      |                      |
        |---create_viewdef---->|                      |                      |                      |
        |   (TYPE.VIEW, html)  |                      |                      |                      |
        |                      |                      |                      |                      |
        |                      |---store(T.V,html)--->|                      |                      |
        |                      |                      |                      |                      |
        |                      |                      |---parseBindings()--->|                      |
        |                      |                      |                      |                      |
        |                      |                      |---queueUpdate------->|                      |
        |                      |                      |   (for var 1)        |                      |
        |                      |                      |                      |                      |
        |<--success------------|                      |                      |                      |
        |                      |                      |                      |                      |
        |---load_presenter_--->|                      |                      |                      |
        |   logic(lua_code)    |                      |                      |                      |
        |                      |                      |                      |                      |
        |                      |---loadCode()-------->|                      |                      |
        |                      |                      |--------------------------------------load-->|
        |                      |                      |                      |                      |
        |                      |                      |                      |<--type registered----|
        |                      |                      |                      |                      |
        |<--success------------|                      |                      |                      |
        |                      |                      |                      |                      |
        |---create_presenter-->|                      |                      |                      |
        |   (type, props)      |                      |                      |                      |
        |                      |                      |                      |                      |
        |                      |---create(props,----->|                      |                      |
        |                      |    {type:T})         |---create()---------->|                      |
        |                      |                      |                      |                      |
        |                      |                      |---notifyWatchers---->|                      |
        |                      |                      |                      |                      |
        |<--{varId}------------|                      |                      |                      |
        |                      |                      |                      |                      |
```

## Notes

- AI creates viewdef with HTML and ui-* bindings
- Lua code optional for presenter logic
- Presenter created with type and initial properties
- Viewdef delivered to frontend via variable 1
- AI can create viewdefs on-the-fly without registration
- No compile step needed - immediate display
