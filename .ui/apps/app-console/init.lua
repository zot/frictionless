-- App Console initialization
-- Adds convenience methods to mcp global for Claude to report build progress

function mcp:appProgress(name, progress, stage)
    if appConsole then appConsole:onAppProgress(name, progress, stage) end
end

function mcp:appUpdated(name)
    mcp:scanAvailableApps()  -- rescan all apps from disk
    if appConsole then appConsole:onAppUpdated(name) end
end

-- Legacy API: full control over todo state
function mcp:setTodos(todos)
    if appConsole then appConsole:setTodos(todos) end
end

-- Simplified API: create todos from step labels
function mcp:createTodos(steps, appName)
    if appConsole then appConsole:createTodos(steps, appName) end
end

-- Advance to step n (completes previous, starts n)
function mcp:startTodoStep(n)
    if appConsole then appConsole:startTodoStep(n) end
end

-- Mark all complete, clear progress
function mcp:completeTodos()
    if appConsole then appConsole:completeTodos() end
end

-- Test function for hot-loading verification
function mcp:hotLoadTest()
    return "hot-load works!"
end
