-- MCP Shell
-- Outer shell for all ui-mcp apps with app switcher menu
-- Note: MCP is the type of the server-created mcp object

local json = require("mcp.json")

-- Version comparison helper
local function compareVersions(current, latest)
    local function parse(v)
        local major, minor, patch = v:match("(%d+)%.(%d+)%.(%d+)")
        return tonumber(major) or 0, tonumber(minor) or 0, tonumber(patch) or 0
    end
    local cm, cn, cp = parse(current)
    local lm, ln, lp = parse(latest)
    if lm > cm then return true end
    if lm == cm and ln > cn then return true end
    if lm == cm and ln == cn and lp > cp then return true end
    return false
end

-- Fetch latest version from GitHub releases API
local function fetchLatestVersion()
    local handle = io.popen('curl -s --connect-timeout 5 --max-time 10 "https://api.github.com/repos/zot/frictionless/releases/latest" 2>/dev/null')
    if not handle then return nil end
    local content = handle:read("*a")
    handle:close()
    if not content or content == "" then return nil end
    local ok, data = pcall(json.decode, content)
    if ok and data and data.tag_name then
        return data.tag_name:gsub("^v", "")
    end
    return nil
end

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

-- Settings helpers (shared across apps via mcp global)
function mcp:readSettings()
    local status = self:status()
    if not status or not status.base_dir then return {} end
    local path = status.base_dir .. "/storage/settings.json"
    local f = io.open(path, "r")
    if not f then return {} end
    local content = f:read("*a")
    f:close()
    local ok, data = pcall(json.decode, content)
    if not ok or type(data) ~= "table" then return {} end
    return data
end

function mcp:writeSettings(data)
    local status = self:status()
    if not status or not status.base_dir then return end
    local dir = status.base_dir .. "/storage"
    os.execute('mkdir -p "' .. dir .. '"')
    local path = dir .. "/settings.json"
    local f = io.open(path, "w")
    if not f then return end
    f:write(json.encode(data))
    f:close()
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

-- Step definitions keyed by label for lookup in createTodos
local UI_STEP_DEFS = {
    ["Read requirements"]  = {progress = 5,  thinking = "Reading requirements..."},
    ["Requirements"]       = {progress = 10, thinking = "Updating requirements..."},
    ["Design"]             = {progress = 20, thinking = "Designing..."},
    ["Write code"]         = {progress = 40, thinking = "Writing code..."},
    ["Write viewdefs"]     = {progress = 60, thinking = "Writing viewdefs..."},
    ["Link and audit"]     = {progress = 85, thinking = "Auditing..."},
    ["Simplify"]           = {progress = 92, thinking = "Simplifying..."},
    ["Set baseline"]       = {progress = 98, thinking = "Setting baseline..."},
    ["Fast Design"]        = {progress = 20, thinking = "Designing..."},
    ["Fast code"]          = {progress = 40, thinking = "Writing code..."},
    ["Fast viewdefs"]      = {progress = 60, thinking = "Writing viewdefs..."},
    ["Fast verify"]        = {progress = 60, thinking = "Verifying requirements..."},
    ["Fast finish"]        = {progress = 80, thinking = "Finishing..."},
}

-- Nested prototype: Chat message model
MCP.ChatMessage = session:prototype("MCP.ChatMessage", {
    sender = "",
    text = "",
    style = "normal",  -- "normal" or "thinking"
    _thumbnails = EMPTY
})
local ChatMessage = MCP.ChatMessage

MCP.ChatThumbnail = session:prototype("MCP.ChatThumbnail", {
    uri = "",
    fullUri = "",
    filename = ""
})
local ChatThumbnail = MCP.ChatThumbnail

function ChatThumbnail:showFull()
    mcp.lightboxUri = self.fullUri ~= "" and self.fullUri or self.uri
end

function ChatMessage:new(sender, text, style, thumbnails)
    local msg = session:create(ChatMessage, { sender = sender, text = text, style = style or "normal" })
    msg._thumbnails = thumbnails or {}
    return msg
end

function ChatMessage:isUser()
    return self.sender == "You"
end

function ChatMessage:isThinking()
    return self.style == "thinking"
end

function ChatMessage:hasThumbnails()
    return self._thumbnails and #self._thumbnails > 0
end

function ChatMessage:noThumbnails()
    return not self:hasThumbnails()
end

function ChatMessage:chatThumbnails()
    return self._thumbnails
end

function ChatMessage:mutate()
    if self.style == nil then
        self.style = "normal"
    end
    if self._thumbnails == nil then
        self._thumbnails = {}
    end
end

function ChatMessage:prefix()
    return self.sender == "You" and "> " or ""
end

-- Nested prototype: TodoItem model (Claude Code task)
MCP.TodoItem = session:prototype("MCP.TodoItem", {
    content = "",
    status = "pending",  -- "pending", "in_progress", or "completed"
    activeForm = ""
})
local TodoItem = MCP.TodoItem

function TodoItem:displayText()
    if self.status == "in_progress" then
        return self.activeForm ~= "" and self.activeForm or self.content
    end
    return self.content
end

function TodoItem:isPending()
    return self.status == "pending"
end

function TodoItem:isInProgress()
    return self.status == "in_progress"
end

function TodoItem:isCompleted()
    return self.status == "completed"
end

local TODO_STATUS_ICONS = {
    in_progress = "🔄",
    completed = "✓",
    pending = "⏳"
}

function TodoItem:statusIcon()
    return TODO_STATUS_ICONS[self.status] or "⏳"
end

-- Nested prototype: Output line model (for Lua console)
MCP.OutputLine = session:prototype("MCP.OutputLine", {
    text = ""
})
local OutputLine = MCP.OutputLine

function OutputLine:copyToInput()
    local text = self.text
    if text:match("^> ") then
        text = text:sub(3)
    end
    mcp.luaInput = text
    mcp:focusLuaInput()
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

-- Nested prototype: Image attachment (pending image for chat send)
MCP.ImageAttachment = session:prototype("MCP.ImageAttachment", {
    path = "",
    filename = "",
    thumbnailUri = "",
    fullUri = "",
    _mcp = EMPTY
})
local ImageAttachment = MCP.ImageAttachment

function ImageAttachment:remove()
    if self._mcp then
        self._mcp:removeAttachment(self)
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

-- Chat/Lua/Todo panel state
mcp.panelOpen = mcp.panelOpen or false
mcp.messages = mcp.messages or {}
mcp.chatInput = mcp.chatInput or ""
mcp.panelMode = mcp.panelMode or "chat"  -- "chat" or "lua"
mcp.luaOutputLines = mcp.luaOutputLines or {}
mcp.luaInput = mcp.luaInput or ""
mcp._luaInputFocusTrigger = mcp._luaInputFocusTrigger or 0
mcp.todos = mcp.todos or {}
mcp.todosCollapsed = mcp.todosCollapsed or false
mcp._todoSteps = mcp._todoSteps or {}
mcp._currentStep = mcp._currentStep or 0
mcp._todoApp = mcp._todoApp or nil

-- Image attachment state
mcp._imageAttachments = mcp._imageAttachments or {}
mcp.imageUploadData = mcp.imageUploadData or ""
mcp.lightboxUri = mcp.lightboxUri or ""

-- Update check state
mcp.showUpdatePrefDialog = mcp.showUpdatePrefDialog or false
mcp.showUpdateConfirmDialog = mcp.showUpdateConfirmDialog or false
mcp.latestVersion = mcp.latestVersion or ""
mcp._isUpdating = mcp._isUpdating or false
mcp._updateNotificationDismissed = mcp._updateNotificationDismissed or false
mcp._needsUpdate = mcp._needsUpdate or false

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
    self.buildMode = self.buildMode == "fast" and "thorough" or "fast"
end

function mcp:isFastMode()
    return self.buildMode == "fast"
end

function mcp:isThoroughMode()
    return self.buildMode == "thorough"
end

function mcp:buildModeTooltip()
    return self.buildMode == "fast" and "Fast mode (click to change)" or "Thorough mode (click to change)"
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
    return self.runInBackground and "Background (click to change)" or "Foreground (click to change)"
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
    return self:currentAppHasCheckpoints() and "Go to App - fast coded" or "Go to App"
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

-- Warn once when Claude appears disconnected for > 5 seconds
-- requirePending: when true, only warn if there are pending events
function mcp:warnIfDisconnected(requirePending)
    local wt = self:waitTime()
    if wt == 0 then
        self._notifiedForDisconnect = false
    elseif not self._notifiedForDisconnect and wt > 5
           and (not requirePending or self:pendingEventCount() > 0) then
        self:notify("Claude might be busy. Use /ui events to reconnect.", "warning")
        self._notifiedForDisconnect = true
    end
end

-- Check and notify if Claude appears disconnected (called on UI refresh)
-- Returns empty string for hidden span binding
function mcp:checkDisconnectNotify()
    self:warnIfDisconnected(true)
    return ""
end

-- Chat panel toggle
function mcp:togglePanel()
    self.panelOpen = not self.panelOpen
end

function mcp:panelHidden()
    return not self.panelOpen
end

function mcp:panelIcon()
    return self.panelOpen and "chat-dots-fill" or "chat-dots"
end

-- Chat panel mode
function mcp:showChatPanel()
    self.panelMode = "chat"
end

function mcp:showLuaPanel()
    self.panelMode = "lua"
end

function mcp:notChatPanel()
    return self.panelMode ~= "chat"
end

function mcp:notLuaPanel()
    return self.panelMode ~= "lua"
end

function mcp:chatTabVariant()
    return self.panelMode == "chat" and "primary" or "default"
end

function mcp:luaTabVariant()
    return self.panelMode == "lua" and "primary" or "default"
end

-- Chat messaging
function mcp:sendChat()
    if self.chatInput == "" and #self._imageAttachments == 0 then return end

    -- Collect image paths and thumbnails
    local imagePaths = nil
    local thumbnails = nil
    if #self._imageAttachments > 0 then
        imagePaths = {}
        thumbnails = {}
        for _, att in ipairs(self._imageAttachments) do
            table.insert(imagePaths, att.path)
            table.insert(thumbnails, session:create(ChatThumbnail, {
                uri = att.thumbnailUri,
                fullUri = att.fullUri,
                filename = att.filename
            }))
        end
    end
    table.insert(self.messages, ChatMessage:new("You", self.chatInput, nil, thumbnails))

    local currentApp = self:currentAppName() or "app-console"
    local status = self:status()

    local event = {
        app = currentApp,
        event = "chat",
        text = self.chatInput,
        images = imagePaths,
        reminder = "Show todos and thinking messages while working. **IMPORTANT:** respond with UI chat messages **AND** in the main Claude Code Console!",
        mcp_port = status and status.mcp_port or nil,
        note = status and ("make sure you have understood the app at " .. status.base_dir .. "/apps/" .. currentApp) or nil
    }

    mcp.pushState(event)
    self.chatInput = ""
    self._imageAttachments = {}  -- Clear attachments (files persist for agent)
end

-- Image attachment handling
function mcp:processImageUpload()
    if self.imageUploadData == "" then return end
    local payload = self.imageUploadData
    self.imageUploadData = ""

    local ok, data = pcall(json.decode, payload)
    if not ok or not data then return end

    local filename = data.filename or "image.png"
    local base64Data = data.base64 or ""
    local thumbnailUri = data.thumbnail or ""
    local fullUri = data.fullUri or ""

    if base64Data == "" then return end

    -- Write base64 to temp file, decode to output
    local status = self:status()
    local uploadDir = (status and status.base_dir or "/tmp") .. "/storage/uploads"
    os.execute('mkdir -p "' .. uploadDir .. '"')

    local ext = filename:match("%.(%w+)$") or "png"
    local outPath = uploadDir .. "/img-" .. os.time() .. "-" .. math.random(10000) .. "." .. ext

    local tmpB64 = os.tmpname()
    local f = io.open(tmpB64, "wb")
    if not f then return end
    f:write(base64Data)
    f:close()
    os.execute('base64 -d < "' .. tmpB64 .. '" > "' .. outPath .. '"')
    os.remove(tmpB64)

    table.insert(self._imageAttachments, session:create(ImageAttachment, {
        path = outPath,
        filename = filename,
        thumbnailUri = thumbnailUri,
        fullUri = fullUri,
        _mcp = self
    }))
end

function mcp:imageAttachments()
    return self._imageAttachments
end

function mcp:hasImages()
    return #self._imageAttachments > 0
end

function mcp:noImages()
    return #self._imageAttachments == 0
end

function mcp:removeAttachment(att)
    for i, a in ipairs(self._imageAttachments) do
        if a == att then
            table.remove(self._imageAttachments, i)
            if att.path ~= "" then os.remove(att.path) end
            break
        end
    end
end

function mcp:clearAttachments()
    for _, att in ipairs(self._imageAttachments) do
        if att.path ~= "" then os.remove(att.path) end
    end
    self._imageAttachments = {}
end

-- Update check methods

function mcp:checkForUpdates()
    local current = self:currentVersion()
    if not current then return end
    local latest = fetchLatestVersion()
    if not latest then return end
    self.latestVersion = latest
    self._needsUpdate = compareVersions(current, latest)
    -- Persist to settings
    local settings = self:readSettings()
    settings.needsUpdate = self._needsUpdate
    if self._needsUpdate then
        settings.latestVersion = latest
    end
    self:writeSettings(settings)
end

function mcp:showUpdatePreferenceDialog()
    self.showUpdatePrefDialog = true
end

function mcp:setUpdatePreference(enabled)
    self.showUpdatePrefDialog = false
    local settings = self:readSettings()
    settings.checkUpdate = enabled
    self:writeSettings(settings)
    if enabled then
        self:checkForUpdates()
    end
end

function mcp:getUpdatePreference()
    local settings = self:readSettings()
    return settings.checkUpdate == true
end

function mcp:currentVersion()
    local status = self:status()
    return status and status.version
end

function mcp:noUpdateAvailable()
    return not self._needsUpdate
end

function mcp:updateAvailable()
    return self._needsUpdate
end

function mcp:hideUpdateNotification()
    return not self._needsUpdate or self._updateNotificationDismissed or self._isUpdating
end

function mcp:dismissUpdateNotification()
    self._updateNotificationDismissed = true
end

function mcp:startUpdate()
    self.showUpdateConfirmDialog = true
end

function mcp:cancelUpdate()
    self.showUpdateConfirmDialog = false
end

function mcp:confirmUpdate()
    self.showUpdateConfirmDialog = false
    self._isUpdating = true
    self._updateNotificationDismissed = true

    -- Detect platform for instructions
    local uname_s = ""
    local uname_m = ""
    local h = io.popen("uname -s 2>/dev/null")
    if h then uname_s = h:read("*a"):match("^%s*(.-)%s*$") or ""; h:close() end
    h = io.popen("uname -m 2>/dev/null")
    if h then uname_m = h:read("*a"):match("^%s*(.-)%s*$") or ""; h:close() end

    local event = {
        type = "update",
        action = "perform-update",
        currentVersion = self:currentVersion(),
        latestVersion = self.latestVersion,
        platform = uname_s,
        architecture = uname_m,
        releaseUrl = "https://github.com/zot/frictionless/releases/download/v" .. self.latestVersion,
        instructions = [[
1. Detect platform (uname -s) and architecture (uname -m)
2. Download the appropriate binary from releaseUrl:
   - Linux x86_64: frictionless-linux-amd64
   - Linux aarch64: frictionless-linux-arm64
   - Darwin arm64: frictionless-darwin-arm64
   - Darwin x86_64: frictionless-darwin-amd64
3. Replace the current frictionless binary (which frictionless)
4. Make it executable (chmod +x)
5. The MCP server will restart automatically with the new binary
6. Call the ui_update tool (NOT install --force) for smart file update
7. If conflicts returned, offer to merge each one
8. Notify the user that update is complete
]]
    }
    mcp.pushState(event)
end

function mcp:isUpdating()
    return self._isUpdating
end

function mcp:notUpdating()
    return not self._isUpdating
end

function mcp:lightboxVisible()
    return self.lightboxUri ~= ""
end

function mcp:hideLightbox()
    self.lightboxUri = ""
end

function mcp:addAgentMessage(text)
    table.insert(self.messages, ChatMessage:new("Agent", text))
    self.statusLine = ""
    self.statusClass = ""
end

function mcp:addAgentThinking(text)
    table.insert(self.messages, ChatMessage:new("Agent", text, "thinking"))
    self.statusLine = text
    self.statusClass = "thinking"
end

function mcp:clearChat()
    self.messages = {}
end

function mcp:clearPanel()
    if self.panelMode == "chat" then
        self:clearChat()
    else
        self:clearLuaOutput()
    end
end

-- Lua console
function mcp:appendOutput(text)
    table.insert(self.luaOutputLines, session:create(OutputLine, { text = text }))
end

function mcp:runLua()
    if self.luaInput == "" then return end

    self:appendOutput("> " .. self.luaInput)

    local code = self.luaInput

    local fn
    if code:match("^%s*return%s") then
        fn = loadstring(code)
    else
        fn = loadstring("return " .. code) or loadstring(code)
    end

    if not fn then
        local _, err = loadstring(code)
        self:appendOutput("Syntax error: " .. tostring(err))
        return
    end

    local ok, result = pcall(fn)
    if ok then
        if result ~= nil then
            self:appendOutput(tostring(result))
        end
        self.luaInput = ""
    else
        self:appendOutput("Error: " .. tostring(result))
    end
end

function mcp:clearLuaOutput()
    self.luaOutputLines = {}
end

function mcp:focusLuaInput()
    self._luaInputFocusTrigger = string.format([[
        var input = document.getElementById('lua-input');
        if (input) {
            input.focus();
            setTimeout(function() {
                var textarea = input.shadowRoot && input.shadowRoot.querySelector('textarea');
                if (textarea) {
                    var len = textarea.value.length;
                    textarea.setSelectionRange(len, len);
                }
            }, 0);
        }
        // %d
    ]], os.time())
end

-- Todo management
function mcp:setTodos(todos)
    self.todos = {}
    for _, t in ipairs(todos or {}) do
        local item = session:create(TodoItem, {
            content = t.content or "",
            status = t.status or "pending",
            activeForm = t.activeForm or ""
        })
        table.insert(self.todos, item)
    end
end

function mcp:toggleTodos()
    self.todosCollapsed = not self.todosCollapsed
end

function mcp:hasTodos()
    return self.todos and #self.todos > 0
end

function mcp:createTodos(steps, appName)
    self._todoApp = appName
    self._currentStep = 0
    self._todoSteps = {}
    self.todos = {}

    for _, label in ipairs(steps or {}) do
        local known = UI_STEP_DEFS[label]
        local stepDef = known
            and {label = label, progress = known.progress, thinking = known.thinking}
            or  {label = label, progress = #self._todoSteps * 15 + 10, thinking = label .. "..."}
        table.insert(self._todoSteps, stepDef)

        local item = session:create(TodoItem, {
            content = label,
            status = "pending",
            activeForm = stepDef.thinking
        })
        table.insert(self.todos, item)
    end
end

function mcp:startTodoStep(n)
    if n < 1 or n > #self._todoSteps then return end

    if self._currentStep > 0 and self._currentStep <= #self.todos then
        self.todos[self._currentStep].status = "completed"
    end

    self._currentStep = n
    local step = self._todoSteps[n]

    if n <= #self.todos then
        self.todos[n].status = "in_progress"
    end

    if self._todoApp and appConsole then
        appConsole:onAppProgress(self._todoApp, step.progress, step.thinking:gsub("%.%.%.$", ""))
    end
    self:addAgentThinking(step.thinking)
end

function mcp:completeTodos()
    for _, todo in ipairs(self.todos or {}) do
        todo.status = "completed"
    end
    if self._todoApp and appConsole then
        appConsole:onAppProgress(self._todoApp, nil, nil)
    end
    self._currentStep = 0
end

function mcp:clearTodos()
    self.todos = {}
    self._todoSteps = {}
    self._currentStep = 0
    self._todoApp = nil
    self.statusLine = ""
    self.statusClass = ""
end

function mcp:appProgress(name, progress, stage)
    if appConsole then appConsole:onAppProgress(name, progress, stage) end
end

function mcp:appUpdated(name)
    mcp:scanAvailableApps()
    if appConsole then appConsole:onAppUpdated(name) end
end

-- Override pushState to inject build mode and warn on long wait times (idempotent)
if not mcp._originalPushState then
   mcp._originalPushState = mcp.pushState
end

function mcp.pushState(event)
    event.handler = mcp.buildMode == "thorough" and "/ui-thorough" or "/ui-fast"
    event.background = mcp.runInBackground
    mcp:warnIfDisconnected(false)
    return mcp._originalPushState(event)
end

-- Initialization
if not session.reloading then
    mcp:scanAvailableApps()

    -- Update check startup logic
    local settings = mcp:readSettings()
    if settings.checkUpdate == nil then
        -- First run: show preference dialog
        mcp:showUpdatePreferenceDialog()
    elseif settings.checkUpdate then
        -- Restore cached needsUpdate from settings
        if settings.needsUpdate then
            mcp._needsUpdate = true
            mcp.latestVersion = settings.latestVersion or ""
        end
        -- Check for updates (runs curl in background-ish, may block briefly)
        mcp:checkForUpdates()
    end
end
