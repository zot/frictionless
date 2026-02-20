-- Variable browser panel
-- Loaded via require("mcp.variables") from app.lua

MCP.VariableBrowser = session:prototype("MCP.VariableBrowser", {
    active = false,
    _shell = EMPTY,
})
local VariableBrowser = MCP.VariableBrowser

function VariableBrowser:new(shell)
    local instance = session:create(VariableBrowser, { _shell = shell })
    return instance
end

function VariableBrowser:variablesUrl()
    if not self._variablesUrl then
        local status = self._shell:status()
        local port = status and status.url or ("http://localhost:" .. (status and status.mcp_port or 8000))
        self._variablesUrl = port .. "/" .. self._shell.sessionId .. "/variables.json"
    end
    return self._variablesUrl
end

function VariableBrowser:activate()
    self.active = true
end

function VariableBrowser:deactivate()
    self.active = false
end

function VariableBrowser:popOut()
    local status = self._shell:status()
    local port = status and status.url or ("http://localhost:" .. (status and status.mcp_port or 8000))
    local url = port .. "/variables"
    self._popOutCounter = (self._popOutCounter or 0) + 1
    self._popOutCode = string.format("window.open(%q, '_blank') // %d", url, self._popOutCounter)
end
