-- App Console
-- Command center for UI development with Claude

-- App prototype (serves as namespace)
AppConsole = session:prototype("AppConsole", {
    _apps = EMPTY,
    _baseDir = "",  -- Cached base_dir from mcp:status()
    selected = EMPTY,
    showNewForm = false,
    newAppName = "",
    newAppDesc = "",
    messages = EMPTY,
    chatInput = "",
    embeddedApp = EMPTY,  -- Name of embedded app, or nil
    embeddedValue = EMPTY, -- App global loaded via mcp:app, or nil
    panelMode = "chat",   -- "chat" or "lua" (bottom panel mode)
    luaOutputLines = EMPTY,
    luaInput = "",
    chatQuality = 0,  -- 0=fast, 1=thorough, 2=background
    todos = EMPTY,           -- Claude Code todo list items
    todosCollapsed = false,  -- Whether todo column is collapsed
    _todoSteps = EMPTY,      -- Step definitions for createTodos/startTodoStep
    _currentStep = 0,        -- Current in_progress step (1-based), 0 if none
    _todoApp = EMPTY         -- App name for progress reporting
})

-- Hardcoded ui-builder step definitions
local UI_BUILDER_STEPS = {
    {label = "Read requirements", progress = 5, thinking = "Reading requirements..."},
    {label = "Requirements", progress = 10, thinking = "Updating requirements..."},
    {label = "Design", progress = 20, thinking = "Designing..."},
    {label = "Write code", progress = 40, thinking = "Writing code..."},
    {label = "Write viewdefs", progress = 60, thinking = "Writing viewdefs..."},
    {label = "Link and audit", progress = 90, thinking = "Auditing..."},
    {label = "Simplify", progress = 95, thinking = "Simplifying..."},
}

-- Nested prototype: Chat message model
AppConsole.ChatMessage = session:prototype("AppConsole.ChatMessage", {
    sender = "",
    text = "",
    style = "normal"  -- "normal" or "thinking"
})
local ChatMessage = AppConsole.ChatMessage

function ChatMessage:new(sender, text, style)
    return session:create(ChatMessage, { sender = sender, text = text, style = style or "normal" })
end

function ChatMessage:isUser()
    return self.sender == "You"
end

function ChatMessage:isThinking()
    return self.style == "thinking"
end

function ChatMessage:mutate()
    -- Initialize style for existing instances
    if self.style == nil then
        self.style = "normal"
    end
end

function ChatMessage:prefix()
    return self.sender == "You" and "> " or ""
end

-- Nested prototype: TodoItem model (Claude Code task)
AppConsole.TodoItem = session:prototype("AppConsole.TodoItem", {
    content = "",
    status = "pending",  -- "pending", "in_progress", or "completed"
    activeForm = ""
})
local TodoItem = AppConsole.TodoItem

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

function TodoItem:statusIcon()
    if self.status == "in_progress" then
        return "🔄"
    elseif self.status == "completed" then
        return "✓"
    else
        return "⏳"
    end
end

-- Nested prototype: Output line model (for Lua console)
AppConsole.OutputLine = session:prototype("AppConsole.OutputLine", {
    text = "",
    panel = EMPTY
})
local OutputLine = AppConsole.OutputLine

function OutputLine:copyToInput()
    -- Strip leading "> " from command lines when copying
    local text = self.text
    if text:match("^> ") then
        text = text:sub(3)
    end
    self.panel.luaInput = text
end

-- Nested prototype: Issue model
AppConsole.Issue = session:prototype("AppConsole.Issue", {
    number = 0,
    title = ""
})
local Issue = AppConsole.Issue

function Issue:new(num, title)
    return session:create(Issue, { number = num, title = title })
end

-- Nested prototype: Test item model
AppConsole.TestItem = session:prototype("AppConsole.TestItem", {
    text = "",
    status = "untested"  -- "passed", "failed", or "untested"
})
local TestItem = AppConsole.TestItem

function TestItem:new(text, status)
    return session:create(TestItem, { text = text, status = status or "untested" })
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

-- Nested prototype: App info model
AppConsole.AppInfo = session:prototype("AppConsole.AppInfo", {
    name = "",
    description = "",
    requirementsContent = "",
    showRequirements = false,
    hasViewdefs = false,
    testsPassing = 0,
    testsTotal = 0,
    knownIssues = EMPTY,
    fixedIssues = EMPTY,
    tests = EMPTY,
    showKnownIssues = true,
    showFixedIssues = false,
    gapsContent = "",
    showGaps = false,
    buildProgress = EMPTY,
    buildStage = EMPTY
})
local AppInfo = AppConsole.AppInfo

function AppInfo:new(name)
    local app = session:create(AppInfo, { name = name })
    app.knownIssues = {}
    app.fixedIssues = {}
    app.tests = {}
    return app
end

function AppInfo:selectMe()
    appConsole:select(self)
end

function AppInfo:isSelected()
    return appConsole.selected == self
end

function AppInfo:statusText()
    if self.buildProgress then
        return self.buildStage or "building..."
    elseif not self.hasViewdefs then
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
    elseif not self.hasViewdefs then
        return "neutral"
    elseif self.testsTotal == 0 then
        return "neutral"
    elseif self.testsPassing == self.testsTotal then
        return "success"
    else
        return "warning"
    end
end

function AppInfo:noIssues()
    return #self.knownIssues == 0
end

function AppInfo:canOpen()
    return self.hasViewdefs
end

function AppInfo:needsBuild()
    return not self.hasViewdefs
end

function AppInfo:knownIssueCount()
    return #self.knownIssues
end

function AppInfo:fixedIssueCount()
    return #self.fixedIssues
end

function AppInfo:noFixedIssues()
    return #self.fixedIssues == 0
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

function AppInfo:hasGaps()
    return self.gapsContent ~= nil and self.gapsContent ~= ""
end

function AppInfo:noGaps()
    return not self:hasGaps()
end

function AppInfo:toggleGaps()
    self.showGaps = not self.showGaps
end

function AppInfo:gapsHidden()
    return not self.showGaps
end

function AppInfo:gapsIcon()
    return self.showGaps and "chevron-down" or "chevron-right"
end

function AppInfo:toggleRequirements()
    self.showRequirements = not self.showRequirements
end

function AppInfo:requirementsHidden()
    return not self.showRequirements
end

function AppInfo:requirementsIcon()
    return self.showRequirements and "chevron-down" or "chevron-right"
end

-- Push an event with common fields (app, mcp_port, note) plus custom fields
function AppInfo:pushEvent(eventType, extra)
    local status = mcp:status()
    local event = {
        app = "app-console",
        event = eventType,
        mcp_port = status.mcp_port,
        note = "make sure you have understood the app at " .. status.base_dir .. "/apps/" .. self.name
    }
    if extra then
        for k, v in pairs(extra) do
            event[k] = v
        end
    end
    mcp.pushState(event)
end

function AppInfo:requestBuild()
   self.buildProgress = 0
   self.buildStage = "pondering"
   self:pushEvent("build_request", { target = self.name })
end

function AppInfo:requestTest()
    self:pushEvent("test_request", { target = self.name })
end

function AppInfo:requestFix()
    self:pushEvent("fix_request", { target = self.name })
end

function AppInfo:openApp()
    print("[DEBUG] openApp called, self.name type:", type(self.name), "value:", tostring(self.name))
    if type(self.name) == "string" then
        appConsole:openEmbedded(self.name)
    else
        print("[DEBUG] ERROR: self.name is not a string!")
    end
end

function AppInfo:isSelf()
    return self.name == "app-console"
end

function AppInfo:isMcp()
    return self.name == "mcp"
end

function AppInfo:openButtonDisabled()
    return self:isSelf() or self:isMcp()
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

-- Check if directory exists and has files
local function dirHasFiles(path)
    local handle = io.popen('ls -1 "' .. path .. '" 2>/dev/null | head -1')
    if handle then
        local result = handle:read("*l")
        handle:close()
        return result ~= nil and result ~= ""
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

-- Parse TESTING.md: extract tests, known issues, fixed issues, gaps
local function parseTesting(content)
    local result = {
        tests = {},
        knownIssues = {},
        fixedIssues = {},
        gapsContent = "",
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

    -- Parse Gaps section (design/code mismatch indicator)
    local gapsSection = content:match("## Gaps.-\n(.-)\n## ")
        or content:match("## Gaps.-\n(.*)$")
    if gapsSection then
        -- Trim whitespace and check if non-empty
        local trimmed = gapsSection:gsub("^%s+", ""):gsub("%s+$", "")
        if trimmed ~= "" then
            result.gapsContent = trimmed
        end
    end

    return result
end

-- Main app methods
function AppConsole:new(instance)
    instance = session:create(AppConsole, instance)
    instance._apps = instance._apps or {}
    instance.messages = instance.messages or {}
    instance.luaOutputLines = instance.luaOutputLines or {}
    return instance
end

-- Hot-load mutation: initialize new fields on existing instances
function AppConsole:mutate()
    if self.chatQuality == nil then
        self.chatQuality = 0
    end
    if self.todos == nil then
        self.todos = {}
    end
    if self.todosCollapsed == nil then
        self.todosCollapsed = false
    end
    if self._todoSteps == nil then
        self._todoSteps = {}
    end
    if self._currentStep == nil then
        self._currentStep = 0
    end
end

-- Return apps list for binding
function AppConsole:apps()
    return self._apps
end

-- Find app by name
function AppConsole:findApp(name)
    for _, app in ipairs(self._apps) do
        if app.name == name then
            return app
        end
    end
    return nil
end

-- Scan apps from disk (Lua-driven discovery)
-- Uses mcp:status() to get base_dir, then scans apps/ directory
function AppConsole:scanAppsFromDisk()
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
            app.requirementsContent = reqContent

            -- Check if built (has viewdefs directory with files)
            app.hasViewdefs = dirHasFiles(appPath .. "/viewdefs")

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

            -- Set gaps content
            app.gapsContent = testData.gapsContent or ""

            table.insert(self._apps, app)
        end
    end

    -- Restore selection if it still exists
    if selectedName then
        self.selected = self:findApp(selectedName)
    end
end

-- Rescan a single app from disk
function AppConsole:rescanApp(name)
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
    app.requirementsContent = reqContent
    app.hasViewdefs = dirHasFiles(appPath .. "/viewdefs")

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

    -- Set gaps content
    app.gapsContent = testData.gapsContent or ""
end

-- Handle app progress event from Claude
function AppConsole:onAppProgress(name, progress, stage)
    local app = self:findApp(name)
    if app then
        app.buildProgress = progress
        app.buildStage = stage
    end
end

-- Handle app updated event from Claude (re-parse single app)
function AppConsole:onAppUpdated(name)
    -- Rescan just this app from disk
    self:rescanApp(name)
end

-- Refresh: rescan all apps from disk (Lua-driven)
function AppConsole:refresh()
    mcp:scanAvailableApps()  -- sync MCP server's app list with disk
    self:scanAppsFromDisk()
end

-- Select an app
function AppConsole:select(app)
    self.selected = app
    self.showNewForm = false
end

-- Show new app form
function AppConsole:openNewForm()
    self.showNewForm = true
    self.newAppName = ""
    self.newAppDesc = ""
end

-- Cancel new app form
function AppConsole:cancelNewForm()
    self.showNewForm = false
    self.newAppName = ""
    self.newAppDesc = ""
end

-- Create new app (Lua creates directory and requirements.md)
function AppConsole:createApp()
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

    -- 4. Select the new app (user can click Build to trigger build_request)
    local app = self:findApp(name)
    if app then
        self.selected = app
    end

    self.showNewForm = false

    -- 5. Start progress at "pondering, 0%"
    mcp:appProgress(name, 0, "pondering")

    -- 6. Send app_created event to Claude for requirements fleshing out
    app:pushEvent("app_created", { name = name, description = desc })

    self.newAppName = ""
    self.newAppDesc = ""
end

-- Quality setting methods
function AppConsole:qualityLabel()
    local labels = {"Fast", "Thorough", "Background"}
    return labels[self.chatQuality + 1]
end

function AppConsole:qualityValue()
    local values = {"fast", "thorough", "background"}
    return values[self.chatQuality + 1]
end

function AppConsole:qualityHandler()
    local handlers = {nil, "/ui-builder", "background-ui-builder"}
    return handlers[self.chatQuality + 1]
end

-- Set quality from slider (captures sl-input events)
function AppConsole:setChatQuality()
   print("QUALITY: "..tostring(self.chatQuality))
    -- self.chatQuality = tonumber(value) or 0
end

-- Send chat message
function AppConsole:sendChat()
    if self.chatInput == "" then return end

    table.insert(self.messages, ChatMessage:new("You", self.chatInput))

    local reminder = "Show todos and thinking messages while working"
    local handler = self:qualityHandler()
    if self.selected then
        self.selected:pushEvent("chat", { text = self.chatInput, context = self.selected.name, quality = self:qualityValue(), handler = handler, reminder = reminder })
    else
        mcp.pushState({ app = "app-console", event = "chat", text = self.chatInput, quality = self:qualityValue(), handler = handler, reminder = reminder })
    end

    self.chatInput = ""
end

-- Add agent message (called by Claude)
function AppConsole:addAgentMessage(text)
    table.insert(self.messages, ChatMessage:new("Agent", text))
    mcp.statusLine = ""  -- Clear status bar when sending a real message
    mcp.statusClass = ""
end

-- Add thinking/progress message (called by Claude while working)
-- text: appears in chat log, status: appears in MCP status bar (orange bold-italic)
function AppConsole:addAgentThinking(text)
    table.insert(self.messages, ChatMessage:new("Agent", text, "thinking"))
    mcp.statusLine = text
    mcp.statusClass = "thinking"
end

-- Check if detail panel should show
function AppConsole:showDetail()
    return self.selected ~= nil and not self.showNewForm
end

-- Check if detail panel should hide
function AppConsole:hideDetail()
    return not self:showDetail()
end

-- Check if placeholder should show
function AppConsole:showPlaceholder()
    return self.selected == nil and not self.showNewForm
end

-- Check if placeholder should hide
function AppConsole:hidePlaceholder()
    return not self:showPlaceholder()
end

-- Check if new form should hide
function AppConsole:hideNewForm()
    return not self.showNewForm
end

-- Open app in embedded view
function AppConsole:openEmbedded(name)
    -- Handle case where name might be an AppInfo object instead of string
    if type(name) == "table" then
        if type(name.name) == "string" then
            name = name.name
        else
            print("Could not extract name from app to open")
            -- Can't extract a string name, bail out
            return
        end
    end
    if type(name) ~= "string" or name == "" then
        return  -- Invalid argument, bail out
    end
    local appValue = mcp:app(name)
    if appValue then
        self.embeddedValue = appValue
        self.embeddedApp = name
    end
end

-- Close embedded view and restore normal layout
function AppConsole:closeEmbedded()
    self.embeddedApp = nil
    self.embeddedValue = nil
end

-- Check if an app is embedded
function AppConsole:hasEmbeddedApp()
    return self.embeddedApp ~= nil
end

-- Check if no app is embedded
function AppConsole:noEmbeddedApp()
    return self.embeddedApp == nil
end

-- Update an app's requirements content (called by Claude after fleshing out)
function AppConsole:updateRequirements(name, content)
    local app = self:findApp(name)
    if app then
        app.requirementsContent = content
        -- Also update description from first paragraph
        app.description = parseRequirements(content)
    end
end

-- Panel mode methods (Chat vs Lua)
function AppConsole:showChatPanel()
    self.panelMode = "chat"
end

function AppConsole:showLuaPanel()
    self.panelMode = "lua"
end

function AppConsole:notChatPanel()
    return self.panelMode ~= "chat"
end

function AppConsole:notLuaPanel()
    return self.panelMode ~= "lua"
end

function AppConsole:chatTabVariant()
    return self.panelMode == "chat" and "primary" or "default"
end

function AppConsole:luaTabVariant()
    return self.panelMode == "lua" and "primary" or "default"
end

-- Lua console methods
function AppConsole:runLua()
    if self.luaInput == "" then return end

    -- Add command line to output
    local cmdLine = session:create(OutputLine, { text = "> " .. self.luaInput, panel = self })
    table.insert(self.luaOutputLines, cmdLine)

    local code = self.luaInput
    local fn, err

    -- If input doesn't start with "return", try prepending it (for expressions)
    if not code:match("^%s*return%s") then
        fn, err = loadstring("return " .. code)
    end

    -- If that failed or wasn't tried, use original code
    if not fn then
        fn, err = loadstring(code)
    end

    if fn then
        local ok, result = pcall(fn)
        if ok then
            if result ~= nil then
                local resultLine = session:create(OutputLine, { text = tostring(result), panel = self })
                table.insert(self.luaOutputLines, resultLine)
            end
            -- Clear input on success
            self.luaInput = ""
        else
            -- Runtime error - keep input for correction
            local errLine = session:create(OutputLine, { text = "Error: " .. tostring(result), panel = self })
            table.insert(self.luaOutputLines, errLine)
        end
    else
        -- Syntax error - keep input for correction
        local errLine = session:create(OutputLine, { text = "Syntax error: " .. tostring(err), panel = self })
        table.insert(self.luaOutputLines, errLine)
    end
end

function AppConsole:clearLuaOutput()
    self.luaOutputLines = {}
end

function AppConsole:clearChat()
    self.messages = {}
end

function AppConsole:clearPanel()
    if self.panelMode == "chat" then
        self:clearChat()
    else
        self:clearLuaOutput()
    end
end

-- Todo list methods
function AppConsole:setTodos(todos)
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

function AppConsole:toggleTodos()
    self.todosCollapsed = not self.todosCollapsed
end

-- Create todos from step labels (simplified API)
function AppConsole:createTodos(steps, appName)
    self._todoApp = appName
    self._currentStep = 0
    self._todoSteps = {}
    self.todos = {}

    for _, label in ipairs(steps or {}) do
        -- Look up step definition from UI_BUILDER_STEPS by label
        local stepDef = nil
        for _, def in ipairs(UI_BUILDER_STEPS) do
            if def.label == label then
                stepDef = def
                break
            end
        end
        -- Default if not found in predefined steps
        if not stepDef then
            stepDef = {label = label, progress = #self._todoSteps * 15 + 10, thinking = label .. "..."}
        end
        table.insert(self._todoSteps, stepDef)

        -- Create TodoItem (all pending initially)
        local item = session:create(TodoItem, {
            content = stepDef.label,
            status = "pending",
            activeForm = stepDef.thinking
        })
        table.insert(self.todos, item)
    end
end

-- Advance to step n (completes previous, starts n)
function AppConsole:startTodoStep(n)
    if n < 1 or n > #self._todoSteps then return end

    -- Complete previous step
    if self._currentStep > 0 and self._currentStep <= #self.todos then
        self.todos[self._currentStep].status = "completed"
    end

    -- Start new step
    self._currentStep = n
    local step = self._todoSteps[n]

    if n <= #self.todos then
        self.todos[n].status = "in_progress"
    end

    -- Update progress bar
    if self._todoApp then
        self:onAppProgress(self._todoApp, step.progress, step.thinking:gsub("%.%.%.$", ""))
    end

    -- Update thinking message
    self:addAgentThinking(step.thinking)
end

-- Mark all steps complete, clear progress
function AppConsole:completeTodos()
    -- Mark all steps completed
    for _, todo in ipairs(self.todos or {}) do
        todo.status = "completed"
    end
    -- Clear progress bar
    if self._todoApp then
        self:onAppProgress(self._todoApp, nil, nil)
    end
    self._currentStep = 0
end

-- Clear all todos and reset step state
function AppConsole:clearTodos()
    self.todos = {}
    self._todoSteps = {}
    self._currentStep = 0
    self._todoApp = nil
end

-- Idempotent instance creation
if not session.reloading then
    appConsole = AppConsole:new()

    -- Scan apps from disk (Lua-driven discovery)
    appConsole:scanAppsFromDisk()
end
