# Lua API & Patterns Guide

The backend logic is written in Lua (GopherLua). This guide covers the idiomatic patterns for building UI logic.

## Defining Classes

Use the standard Lua table-with-metatable pattern for classes.

```lua
-- 1. Define the class table
MyForm = { type = "MyForm" }
MyForm.__index = MyForm

-- 2. Define the constructor
function MyForm:new(data)
    local obj = data or {}
    setmetatable(obj, self)
    obj.userInput = ""
    return obj
end

-- 3. Define methods
function MyForm:submit()
    mcp.notify("form_submitted", { value = self.userInput })
end
```

## Global Objects

### 1. `session`
Provides access to session-level services.
- `session:getApp()`: Returns the session's root application object (Variable 1).
- `session:createAppVariable(obj)`: Sets the initial app object.

### 2. `mcp` (AI Agents Only)
Provides display and communication for AI Agents.
- `mcp.value`: Set this to display an object on screen. Starts as `nil` (blank). The object must have a `type` field matching a viewdef.
- `mcp.notify(method, params)`: Sends a notification to the Agent.

## Change Detection

The platform uses **Automatic Change Detection**. You do not need to call `update()` or `notify()` when you change properties on a Lua table. The system detects modifications after every message batch and pushes changes to the frontend.

**Example:**
```lua
function MyForm:clear()
    -- These changes are automatically detected and sent to the browser
    self.userInput = ""
    self.error = nil
end
```

## Tips for AI Agents

- **Modules:** Use `require` to load standard libraries or other files.
- **Error Handling:** Errors in Lua code will be reported back through the `ui_run` tool.
- **Persistence:** Use the provided base directory (via `ui_configure`) if you need to read/write local files.
