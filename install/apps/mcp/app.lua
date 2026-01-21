-- MCP Shell
-- Outer shell for all Frictionless apps with app switcher menu
-- Note: MCP is the type of the server-created mcp object

-- Filesystem helpers (same pattern as apps app)

local function fileExists(path)
    local handle = io.open(path, "r")
    if handle then
        handle:close()
        return true
    end
    return false
end

local function listDirs(path)
    local dirs = {}
    local handle = io.popen('ls -1d "' .. path .. '"/*/ 2>/dev/null')
    if handle then
        for line in handle:lines() do
            local name = line:match("([^/]+)/$")
            if name and name ~= "" then
                table.insert(dirs, name)
            end
        end
        handle:close()
    end
    return dirs
end

-- Nested prototype: App menu item (wraps app name for list binding)
-- Namespace under MCP since mcp is the global instance of type MCP
MCP.AppMenuItem = session:prototype("MCP.AppMenuItem", {
    _name = "",
    _mcp = EMPTY
})
local AppMenuItem = MCP.AppMenuItem

function AppMenuItem:new(name, mcpRef)
    local item = session:create(AppMenuItem, { _name = name, _mcp = mcpRef })
    return item
end

function AppMenuItem:name()
    return self._name
end

function AppMenuItem:select()
    if self._mcp then
        self._mcp:selectApp(self._name)
    end
end

-- Extend the global mcp object with shell functionality
-- Note: mcp is created by the server, we just add methods and properties

-- Add properties for menu state
if not mcp._availableApps then
    mcp._availableApps = {}
end
if mcp.menuOpen == nil then
    mcp.menuOpen = false
end

-- Scan for available apps (built apps with app.lua)
function mcp:scanAvailableApps()
    local status = mcp:status()
    if not status or not status.base_dir then
        return
    end

    local appsPath = status.base_dir .. "/apps"
    local appDirs = listDirs(appsPath)

    self._availableApps = {}
    for _, name in ipairs(appDirs) do
        local appPath = appsPath .. "/" .. name
        -- Only include apps that are built (have app.lua)
        -- Exclude "mcp" - it's the shell, not a user app
        if name ~= "mcp" and fileExists(appPath .. "/app.lua") then
            table.insert(self._availableApps, AppMenuItem:new(name, self))
        end
    end
end

-- Return available apps for binding
function mcp:availableApps()
    return self._availableApps
end

-- Toggle menu visibility
function mcp:toggleMenu()
    self.menuOpen = not self.menuOpen
end

-- Close menu
function mcp:closeMenu()
    self.menuOpen = false
end

-- Check if menu is hidden (for ui-class-hidden)
function mcp:menuHidden()
    return not self.menuOpen
end

-- Select an app from the menu
function mcp:selectApp(name)
    mcp:display(name)
    self.menuOpen = false
end

-- Scan apps on initial load
if not session.reloading then
    mcp:scanAvailableApps()
end
