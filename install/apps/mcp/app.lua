-- MCP Shell
-- Outer shell for all ui-mcp apps with app switcher menu
-- Note: MCP is the type of the server-created mcp object

local json = require('mcp.json')

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

-- Initialize properties for menu state, notifications, and build settings
mcp._availableApps = mcp._availableApps or {}
mcp.menuOpen = mcp.menuOpen or false
mcp._notifications = mcp._notifications or {}
mcp.buildMode = mcp.buildMode or "fast"  -- "fast" or "thorough"
mcp.runInBackground = mcp.runInBackground or false  -- foreground or background execution

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

-- Build mode toggle (fast/thorough)
function mcp:toggleBuildMode()
    if self.buildMode == "fast" then
        self.buildMode = "thorough"
    else
        self.buildMode = "fast"
    end
end

function mcp:isFastMode()
    return self.buildMode == "fast"
end

function mcp:isThoroughMode()
    return self.buildMode == "thorough"
end

function mcp:buildModeTooltip()
    if self.buildMode == "fast" then
        return "Fast mode (click to change)"
    else
        return "Thorough mode (click to change)"
    end
end

-- Background execution toggle
function mcp:toggleBackground()
    self.runInBackground = not self.runInBackground
end

function mcp:isBackground()
    return self.runInBackground
end

function mcp:isForeground()
    return not self.runInBackground
end

function mcp:backgroundTooltip()
    if self.runInBackground then
        return "Background (click to change)"
    else
        return "Foreground (click to change)"
    end
end

-- Generate HTML link for variables endpoint (opens in new tab)
-- Cached since the port doesn't change during a session
function mcp:variablesLinkHtml()
    if not self._variablesLinkHtml then
        local status = self:status()
        local port = status and status.mcp_port or 8000
        self._variablesLinkHtml = string.format('<a href="http://localhost:%d/variables" target="_blank" title="Variables"><sl-icon name="braces"></sl-icon></a>', port)
    end
    return self._variablesLinkHtml
end

-- Generate HTML link for help documentation (opens in new tab)
-- Cached since the port doesn't change during a session
function mcp:helpLinkHtml()
    if not self._helpLinkHtml then
        local status = self:status()
        local port = status and status.mcp_port or 8000
        self._helpLinkHtml = string.format('<a href="http://localhost:%d/api/resource/" target="_blank" title="Documentation"><sl-icon name="question-circle"></sl-icon></a>', port)
    end
    return self._helpLinkHtml
end

-- Get the kebab-case name of the current app from mcp.value.type
function mcp:currentAppName()
    if self.value and self.value.type then
        local typeName = self.value.type
        -- Convert PascalCase to kebab-case: "AppConsole" -> "app-console"
        return typeName:gsub("(%u)", "-%1"):lower():gsub("^-", "")
    end
    return nil
end

-- Check if current app has checkpoints
function mcp:currentAppHasCheckpoints()
    local name = self:currentAppName()
    if name and appConsole then
        local app = appConsole:findApp(name)
        if app then
            return app:hasCheckpoints()
        end
    end
    return false
end

function mcp:currentAppNoCheckpoints()
    return not self:currentAppHasCheckpoints()
end

function mcp:toolsTooltip()
    if self:currentAppHasCheckpoints() then
        return "Go to App - fast coded"
    else
        return "Go to App"
    end
end

-- Open tools panel (app-console) and select the current app
function mcp:openTools()
    local currentApp = self:currentAppName()

    mcp:display("app-console")

    -- Select the current app in app-console after it loads
    if currentApp and appConsole then
        local app = appConsole:findApp(currentApp)
        if app then
            appConsole:select(app)
        end
    end
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

-- Returns true if waiting for Claude to reconnect
function mcp:isWaiting()
    return self:waitTime() > 0
end


-- Check and notify if Claude appears disconnected (called on UI refresh)
-- Returns empty string for hidden span binding
function mcp:checkDisconnectNotify()
    local wt = self:waitTime()
    if wt == 0 then
        self._notifiedForDisconnect = false
    elseif not self._notifiedForDisconnect and wt > 5 and self:pendingEventCount() > 0 then
        self:notify("Claude might be busy. Use /ui events to reconnect.", "warning")
        self._notifiedForDisconnect = true
    end
    return ""
end

-- Override pushState to inject build mode and warn on long wait times (idempotent)
if not mcp._originalPushState then
   mcp._originalPushState = mcp.pushState
end

function mcp.pushState(event)
    -- Inject build settings
    if mcp.buildMode == "thorough" then
        event.handler = "/ui-thorough"
    else
        event.handler = "/ui-fast"
    end
    event.background = mcp.runInBackground

    -- Warn on long wait times
    local wt = mcp:waitTime()
    if wt == 0 then
        mcp._notifiedForDisconnect = false
    elseif not mcp._notifiedForDisconnect and wt > 5 then
        mcp:notify("Claude might be busy. Use /ui events to reconnect.", "warning")
        mcp._notifiedForDisconnect = true
    end
    return mcp._originalPushState(event)
end

-- Initialization
if not session.reloading then
    mcp:scanAvailableApps()
end
