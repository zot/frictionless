-- Apps Dashboard
-- Command center for UI development with Claude

-- Chat message model
ChatMessage = session:prototype("ChatMessage", {
    sender = "",
    text = ""
})

function ChatMessage:new(sender, text)
    return session:create(ChatMessage, { sender = sender, text = text })
end

function ChatMessage:isUser()
    return self.sender == "You"
end

function ChatMessage:prefix()
    return self.sender == "You" and "> " or ""
end

-- Issue model
Issue = session:prototype("Issue", {
    number = 0,
    title = ""
})

function Issue:new(num, title)
    return session:create(Issue, { number = num, title = title })
end

-- Test item model
TestItem = session:prototype("TestItem", {
    text = "",
    status = "untested"  -- "passed", "failed", or "untested"
})

function TestItem:new(text, status)
    return session:create(TestItem, { text = text, status = status or "untested" })
end

function TestItem:isPassed()
    return self.status == "passed"
end

function TestItem:isFailed()
    return self.status == "failed"
end

function TestItem:isUntested()
    return self.status == "untested"
end

function TestItem:icon()
    if self.status == "passed" then
        return "✓"
    elseif self.status == "failed" then
        return "✗"
    else
        return " "
    end
end

function TestItem:iconClass()
    if self.status == "passed" then
        return "passed"
    elseif self.status == "failed" then
        return "failed"
    else
        return "untested"
    end
end

-- App info model
AppInfo = session:prototype("AppInfo", {
    name = "",
    description = "",
    isBuilt = false,
    testsPassing = 0,
    testsTotal = 0,
    knownIssues = EMPTY,
    fixedIssues = EMPTY,
    tests = EMPTY,
    showKnownIssues = true,
    showFixedIssues = false,
    buildProgress = EMPTY,
    buildStage = EMPTY
})

function AppInfo:new(name)
    local app = session:create(AppInfo, { name = name })
    app.knownIssues = {}
    app.fixedIssues = {}
    app.tests = {}
    return app
end

function AppInfo:selectMe()
    apps:select(self)
end

function AppInfo:isSelected()
    return apps.selected == self
end

function AppInfo:statusText()
    if self.buildProgress then
        return self.buildStage or "building..."
    elseif not self.isBuilt then
        return "not built"
    elseif self.testsTotal == 0 then
        return "--"
    else
        return self.testsPassing .. "/" .. self.testsTotal
    end
end

function AppInfo:statusVariant()
    if self.buildProgress then
        return "primary"
    elseif not self.isBuilt then
        return "neutral"
    elseif self.testsTotal == 0 then
        return "neutral"
    elseif self.testsPassing == self.testsTotal then
        return "success"
    else
        return "warning"
    end
end

function AppInfo:hasIssues()
    return #self.knownIssues > 0
end

function AppInfo:noIssues()
    return #self.knownIssues == 0
end

function AppInfo:notBuilt()
    return not self.isBuilt
end

function AppInfo:knownIssueCount()
    return #self.knownIssues
end

function AppInfo:fixedIssueCount()
    return #self.fixedIssues
end

function AppInfo:hasFixedIssues()
    return #self.fixedIssues > 0
end

function AppInfo:noFixedIssues()
    return #self.fixedIssues == 0
end

function AppInfo:hasTests()
    return self.testsTotal > 0
end

function AppInfo:noTests()
    return self.testsTotal == 0
end

function AppInfo:isBuilding()
    return self.buildProgress ~= nil
end

function AppInfo:notBuilding()
    return self.buildProgress == nil
end

function AppInfo:toggleKnownIssues()
    self.showKnownIssues = not self.showKnownIssues
end

function AppInfo:toggleFixedIssues()
    self.showFixedIssues = not self.showFixedIssues
end

function AppInfo:knownIssuesHidden()
    return not self.showKnownIssues
end

function AppInfo:fixedIssuesHidden()
    return not self.showFixedIssues
end

function AppInfo:knownIssuesIcon()
    return self.showKnownIssues and "chevron-down" or "chevron-right"
end

function AppInfo:fixedIssuesIcon()
    return self.showFixedIssues and "chevron-down" or "chevron-right"
end

function AppInfo:requestBuild()
    mcp.pushState({
        app = "apps",
        event = "build_request",
        target = self.name
    })
end

function AppInfo:requestTest()
    mcp.pushState({
        app = "apps",
        event = "test_request",
        target = self.name
    })
end

function AppInfo:requestFix()
    mcp.pushState({
        app = "apps",
        event = "fix_request",
        target = self.name
    })
end

function AppInfo:openApp()
    mcp.display(self.name)
end

-- Filesystem helpers for Lua-driven app discovery

-- Read file contents, returns nil if file doesn't exist
local function readFile(path)
    local handle = io.open(path, "r")
    if not handle then return nil end
    local content = handle:read("*a")
    handle:close()
    return content
end

-- Check if file exists
local function fileExists(path)
    local handle = io.open(path, "r")
    if handle then
        handle:close()
        return true
    end
    return false
end

-- List directories in a path
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

-- Parse requirements.md: extract first paragraph as description
local function parseRequirements(content)
    if not content then return "" end
    -- Skip title line (starts with #)
    local afterTitle = content:gsub("^#[^\n]*\n+", "")
    -- Get first paragraph (text before blank line)
    local firstPara = afterTitle:match("^([^\n]+)")
    return firstPara or ""
end

-- Parse TESTING.md: extract tests, known issues, fixed issues
local function parseTesting(content)
    local result = {
        tests = {},
        knownIssues = {},
        fixedIssues = {},
        testsPassing = 0,
        testsTotal = 0
    }
    if not content then return result end

    -- Parse test checklist items
    -- - [ ] = untested, - [✓] = passed, - [✗] = failed
    for status, text in content:gmatch("%- %[([^%]]+)%] ([^\n]+)") do
        local testStatus = "untested"
        if status == "✓" or status == "x" or status == "X" then
            testStatus = "passed"
            result.testsPassing = result.testsPassing + 1
        elseif status == "✗" then
            testStatus = "failed"
        end
        result.testsTotal = result.testsTotal + 1
        table.insert(result.tests, { text = text, status = testStatus })
    end

    -- Parse Known Issues section
    local knownSection = content:match("## Known Issues.-\n(.-)\n## ")
        or content:match("## Known Issues.-\n(.*)$")
    if knownSection then
        for num, title in knownSection:gmatch("### (%d+)%.%s*([^\n]+)") do
            table.insert(result.knownIssues, { number = tonumber(num), title = title })
        end
    end

    -- Parse Fixed Issues section
    local fixedSection = content:match("## Fixed Issues.-\n(.-)\n## ")
        or content:match("## Fixed Issues.-\n(.*)$")
    if fixedSection then
        for num, title in fixedSection:gmatch("### (%d+)%.%s*([^\n]+)") do
            table.insert(result.fixedIssues, { number = tonumber(num), title = title })
        end
    end

    return result
end

-- Main app
AppsApp = session:prototype("AppsApp", {
    _apps = EMPTY,
    _baseDir = "",  -- Cached base_dir from mcp:status()
    selected = EMPTY,
    showNewForm = false,
    newAppName = "",
    newAppDesc = "",
    messages = EMPTY,
    chatInput = ""
})

function AppsApp:new(instance)
    instance = session:create(AppsApp, instance)
    instance._apps = instance._apps or {}
    instance.messages = instance.messages or {}
    return instance
end

-- Return apps list for binding
function AppsApp:apps()
    return self._apps
end

-- Find app by name
function AppsApp:findApp(name)
    for _, app in ipairs(self._apps) do
        if app.name == name then
            return app
        end
    end
    return nil
end

-- Scan apps from disk (Lua-driven discovery)
-- Uses mcp:status() to get base_dir, then scans apps/ directory
function AppsApp:scanAppsFromDisk()
    -- Get base_dir from mcp:status()
    local status = mcp:status()
    if not status or not status.base_dir then
        return  -- Can't scan without base_dir
    end
    self._baseDir = status.base_dir

    local appsPath = self._baseDir .. "/apps"
    local appDirs = listDirs(appsPath)
    local selectedName = self.selected and self.selected.name or nil

    -- Build new apps list
    self._apps = {}
    for _, name in ipairs(appDirs) do
        local appPath = appsPath .. "/" .. name

        -- Check for requirements.md (required for an app)
        local reqPath = appPath .. "/requirements.md"
        local reqContent = readFile(reqPath)
        if reqContent then
            local app = AppInfo:new(name)

            -- Parse requirements.md for description
            app.description = parseRequirements(reqContent)

            -- Check if built (has app.lua)
            app.isBuilt = fileExists(appPath .. "/app.lua")

            -- Parse TESTING.md if exists
            local testContent = readFile(appPath .. "/TESTING.md")
            local testData = parseTesting(testContent)

            app.testsPassing = testData.testsPassing
            app.testsTotal = testData.testsTotal

            -- Populate tests
            app.tests = {}
            for _, t in ipairs(testData.tests) do
                table.insert(app.tests, TestItem:new(t.text, t.status))
            end

            -- Populate issues
            app.knownIssues = {}
            for _, issue in ipairs(testData.knownIssues) do
                table.insert(app.knownIssues, Issue:new(issue.number, issue.title))
            end

            app.fixedIssues = {}
            for _, issue in ipairs(testData.fixedIssues) do
                table.insert(app.fixedIssues, Issue:new(issue.number, issue.title))
            end

            table.insert(self._apps, app)
        end
    end

    -- Restore selection if it still exists
    if selectedName then
        self.selected = self:findApp(selectedName)
    end
end

-- Rescan a single app from disk
function AppsApp:rescanApp(name)
    if not self._baseDir or self._baseDir == "" then
        -- No base_dir cached, do full scan
        self:scanAppsFromDisk()
        return
    end

    local appPath = self._baseDir .. "/apps/" .. name
    local reqContent = readFile(appPath .. "/requirements.md")

    if not reqContent then
        -- App doesn't exist, remove from list
        for i, app in ipairs(self._apps) do
            if app.name == name then
                table.remove(self._apps, i)
                if self.selected and self.selected.name == name then
                    self.selected = nil
                end
                break
            end
        end
        return
    end

    -- Find or create app
    local app = self:findApp(name)
    if not app then
        app = AppInfo:new(name)
        table.insert(self._apps, app)
    end

    -- Update app data
    app.description = parseRequirements(reqContent)
    app.isBuilt = fileExists(appPath .. "/app.lua")

    -- Parse TESTING.md
    local testContent = readFile(appPath .. "/TESTING.md")
    local testData = parseTesting(testContent)

    app.testsPassing = testData.testsPassing
    app.testsTotal = testData.testsTotal

    -- Clear build progress since we've rescanned
    app.buildProgress = nil
    app.buildStage = nil

    -- Rebuild tests list
    app.tests = {}
    for _, t in ipairs(testData.tests) do
        table.insert(app.tests, TestItem:new(t.text, t.status))
    end

    -- Rebuild issues lists
    app.knownIssues = {}
    for _, issue in ipairs(testData.knownIssues) do
        table.insert(app.knownIssues, Issue:new(issue.number, issue.title))
    end

    app.fixedIssues = {}
    for _, issue in ipairs(testData.fixedIssues) do
        table.insert(app.fixedIssues, Issue:new(issue.number, issue.title))
    end
end

-- Set entire apps list (deprecated: use scanAppsFromDisk instead)
-- Kept for backwards compatibility if Claude pushes data
function AppsApp:setApps(appDataList)
    self._apps = {}
    for _, data in ipairs(appDataList) do
        local app = AppInfo:new(data.name)
        app.description = data.description or ""
        app.isBuilt = data.isBuilt or false
        app.testsPassing = data.testsPassing or 0
        app.testsTotal = data.testsTotal or 0

        -- Populate tests with status
        app.tests = {}
        if data.tests then
            for _, t in ipairs(data.tests) do
                table.insert(app.tests, TestItem:new(t.text, t.status))
            end
        end

        -- Populate issues
        app.knownIssues = {}
        if data.knownIssues then
            for _, issue in ipairs(data.knownIssues) do
                table.insert(app.knownIssues, Issue:new(issue.number, issue.title))
            end
        end

        app.fixedIssues = {}
        if data.fixedIssues then
            for _, issue in ipairs(data.fixedIssues) do
                table.insert(app.fixedIssues, Issue:new(issue.number, issue.title))
            end
        end

        table.insert(self._apps, app)
    end

    -- Update selected if it was pointing to an app
    if self.selected then
        self.selected = self:findApp(self.selected.name)
    end
end

-- Add a single app (used during create flow)
function AppsApp:addApp(name)
    local app = AppInfo:new(name)
    table.insert(self._apps, app)
    return app
end

-- Set build progress for an app (legacy, use onAppProgress)
function AppsApp:setBuildProgress(name, progress, stage)
    local app = self:findApp(name)
    if app then
        app.buildProgress = progress
        app.buildStage = stage
    end
end

-- Handle app progress event from Claude
function AppsApp:onAppProgress(name, progress, stage)
    local app = self:findApp(name)
    if app then
        app.buildProgress = progress
        app.buildStage = stage
    end
end

-- Handle app updated event from Claude (re-parse single app)
function AppsApp:onAppUpdated(name)
    -- Rescan just this app from disk
    self:rescanApp(name)
end

-- Refresh: rescan all apps from disk (Lua-driven)
function AppsApp:refresh()
    self:scanAppsFromDisk()
end

-- Select an app
function AppsApp:select(app)
    self.selected = app
    self.showNewForm = false
end

-- Show new app form
function AppsApp:openNewForm()
    self.showNewForm = true
    self.newAppName = ""
    self.newAppDesc = ""
end

-- Cancel new app form
function AppsApp:cancelNewForm()
    self.showNewForm = false
    self.newAppName = ""
    self.newAppDesc = ""
end

-- Create new app (Lua creates directory and requirements.md)
function AppsApp:createApp()
    if self.newAppName == "" then return end
    if self._baseDir == "" then
        -- Need base_dir to create files
        local status = mcp:status()
        if not status or not status.base_dir then return end
        self._baseDir = status.base_dir
    end

    local name = self.newAppName
    local desc = self.newAppDesc
    local appPath = self._baseDir .. "/apps/" .. name

    -- 1. Create app directory
    os.execute('mkdir -p "' .. appPath .. '"')

    -- 2. Write requirements.md with title and description
    -- Convert kebab-case to Title Case for the heading
    local title = name:gsub("-", " "):gsub("(%a)([%w]*)", function(first, rest)
        return first:upper() .. rest
    end)
    local reqContent = "# " .. title .. "\n\n" .. desc .. "\n"
    local handle = io.open(appPath .. "/requirements.md", "w")
    if handle then
        handle:write(reqContent)
        handle:close()
    end

    -- 3. Rescan to add app to list
    self:rescanApp(name)
    local app = self:findApp(name)

    -- 4. Send build_request to Claude
    if app then
        app.buildProgress = 0
        app.buildStage = "starting..."
        self.selected = app

        mcp.pushState({
            app = "apps",
            event = "build_request",
            target = name
        })
    end

    self.showNewForm = false
    self.newAppName = ""
    self.newAppDesc = ""
end

-- Send chat message
function AppsApp:sendChat()
    if self.chatInput == "" then return end

    table.insert(self.messages, ChatMessage:new("You", self.chatInput))

    mcp.pushState({
        app = "apps",
        event = "chat",
        text = self.chatInput,
        context = self.selected and self.selected.name or nil
    })

    self.chatInput = ""
end

-- Add agent message (called by Claude)
function AppsApp:addAgentMessage(text)
    table.insert(self.messages, ChatMessage:new("Agent", text))
end

-- Check if detail panel should show
function AppsApp:showDetail()
    return self.selected ~= nil and not self.showNewForm
end

-- Check if detail panel should hide
function AppsApp:hideDetail()
    return not self:showDetail()
end

-- Check if placeholder should show
function AppsApp:showPlaceholder()
    return self.selected == nil and not self.showNewForm
end

-- Check if placeholder should hide
function AppsApp:hidePlaceholder()
    return not self:showPlaceholder()
end

-- Check if new form should hide
function AppsApp:hideNewForm()
    return not self.showNewForm
end

-- Initialize
if not session.reloading then
    apps = AppsApp:new()

    -- Scan apps from disk (Lua-driven discovery)
    apps:scanAppsFromDisk()
end
