-- MCP Shell
-- Outer shell for all ui-mcp apps with app switcher menu
-- Note: MCP is the type of the server-created mcp object

-- Filesystem helpers

local function fileExists(path)
    local handle = io.open(path, "r")
    if not handle then return false end
    handle:close()
    return true
end

local function readFileTrimmed(path)
    local handle = io.open(path, "r")
    if not handle then return "" end
    local content = handle:read("*a")
    handle:close()
    if not content then return "" end
    return content:match("^%s*(.-)%s*$") or ""
end

local function listDirs(path)
    local dirs = {}
    local handle = io.popen('ls -1d "' .. path .. '"/*/ 2>/dev/null')
    if not handle then return dirs end
    for line in handle:lines() do
        local name = line:match("([^/]+)/$")
        if name and name ~= "" then
            table.insert(dirs, name)
        end
    end
    handle:close()
    return dirs
end

-- Nested prototype: Notification toast
MCP.Notification = session:prototype("MCP.Notification", {
    message = "",
    variant = "danger",
    _mcp = EMPTY
})
local Notification = MCP.Notification

function Notification:dismiss()
    if self._mcp then
        self._mcp:dismissNotification(self)
    end
end

-- Nested prototype: App menu item (wraps app info for list binding)
MCP.AppMenuItem = session:prototype("MCP.AppMenuItem", {
    _name = "",
    _iconHtml = "",
    _mcp = EMPTY
})
local AppMenuItem = MCP.AppMenuItem

function AppMenuItem:name()
    return self._name
end

function AppMenuItem:iconHtml()
    return self._iconHtml
end

function AppMenuItem:select()
    if self._mcp then
        self._mcp:selectApp(self._name)
    end
end

-- Extend the global mcp object with shell functionality
-- Note: mcp is created by the server, we just add methods and properties

-- Initialize properties for menu state and notifications
mcp._availableApps = mcp._availableApps or {}
mcp.menuOpen = mcp.menuOpen or false
mcp._notifications = mcp._notifications or {}

-- Scan for available apps (built apps with app.lua, excluding mcp shell)
function mcp:scanAvailableApps()
    local status = mcp:status()
    if not status or not status.base_dir then return end

    local appsPath = status.base_dir .. "/apps"
    self._availableApps = {}

    for _, name in ipairs(listDirs(appsPath)) do
        local appPath = appsPath .. "/" .. name
        if name ~= "mcp" and fileExists(appPath .. "/app.lua") then
            local item = session:create(AppMenuItem, {
                _name = name,
                _iconHtml = readFileTrimmed(appPath .. "/icon.html"),
                _mcp = self
            })
            table.insert(self._availableApps, item)
        end
    end
end

function mcp:availableApps()
    return self._availableApps
end

function mcp:toggleMenu()
    self.menuOpen = not self.menuOpen
end

function mcp:closeMenu()
    self.menuOpen = false
end

function mcp:menuHidden()
    return not self.menuOpen
end

function mcp:selectApp(name)
    mcp:display(name)
    self.menuOpen = false
end

function mcp:notify(message, variant)
    local notification = session:create(Notification, {
        message = message,
        variant = variant or "danger",
        _mcp = self
    })
    table.insert(self._notifications, notification)
end

function mcp:notifications()
    return self._notifications
end

function mcp:dismissNotification(notification)
    for i, n in ipairs(self._notifications) do
        if n == notification then
            table.remove(self._notifications, i)
            return
        end
    end
end

-- Returns seconds offset from UNIX epoch when wait started, or 0 if connected
function mcp:waitStartOffset()
    local wt = self:waitTime()
    if wt == 0 then return 0 end
    return math.floor(os.time() - wt)
end


-- Check and notify if Claude appears disconnected (called on UI refresh)
-- Returns empty string for hidden span binding
function mcp:checkDisconnectNotify()
    local wt = self:waitTime()
    if wt == 0 then
        self._notifiedForDisconnect = false
    elseif not self._notifiedForDisconnect and wt > 5 and self:pendingEventCount() > 0 then
        self:notify("Claude might be busy or not watching events", "warning")
        self._notifiedForDisconnect = true
    end
    return ""
end

-- Override pushState to warn on long wait times (idempotent)
function mcp:setupPushStateOverride()
    if mcp._pushStateOverridden then return end
    mcp._pushStateOverridden = true

    local originalPushState = mcp.pushState
    mcp.pushState = function(event)
        local wt = mcp:waitTime()
        if wt == 0 then
            mcp._notifiedForDisconnect = false
        elseif not mcp._notifiedForDisconnect and wt > 5 then
            mcp:notify("Claude might be busy or not watching events", "warning")
            mcp._notifiedForDisconnect = true
        end
        return originalPushState(event)
    end
end

-- Initialization
if not session.reloading then
    mcp:scanAvailableApps()
end
mcp:setupPushStateOverride()
