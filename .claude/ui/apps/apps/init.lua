-- Apps dashboard initialization
-- Adds convenience methods to mcp global for Claude to report build progress

function mcp:appProgress(name, progress, stage)
    if apps then apps:onAppProgress(name, progress, stage) end
end

function mcp:appUpdated(name)
    if apps then apps:onAppUpdated(name) end
end
