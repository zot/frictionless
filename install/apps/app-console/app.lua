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
    embeddedApp = EMPTY,  -- Name of embedded app, or nil
    embeddedValue = EMPTY, -- App global loaded via mcp:app, or nil
    github = EMPTY  -- GitHubDownloader instance
})

-- HTML escape helper for safe display
local function escapeHtml(str)
    return str:gsub("&", "&amp;"):gsub("<", "&lt;"):gsub(">", "&gt;"):gsub('"', "&quot;")
end

-- Dangerous Lua patterns for security review
local DANGER_PATTERNS = {
    -- Shell execution
    {pattern = "os%.execute%s*%(", label = "os.execute"},
    {pattern = "io%.popen%s*%(", label = "io.popen"},
    -- Code loading
    {pattern = "dofile%s*%(", label = "dofile"},
    {pattern = "loadfile%s*%(", label = "loadfile"},
    {pattern = "loadstring%s*%(", label = "loadstring"},
    {pattern = "load%s*%(", label = "load"},
    -- File operations
    {pattern = "io%.open%s*%(", label = "io.open"},
    {pattern = "io%.input%s*%(", label = "io.input"},
    -- OS operations
    {pattern = "os%.exit%s*%(", label = "os.exit"},
    {pattern = "os%.remove%s*%(", label = "os.remove"},
    {pattern = "os%.rename%s*%(", label = "os.rename"},
    {pattern = "os%.tmpname%s*%(", label = "os.tmpname"},
}

-- Pattern for require with non-constant string (dangerous dynamic require)
-- Matches: require(var), require(func()), require("str" .. var), etc.
-- Does NOT match: require("constant") or require('constant')
local function isDangerousRequire(line)
    local requireMatch = line:match("require%s*%((.-)%)")
    if not requireMatch then return false end
    -- Safe if it's just a quoted string constant
    if requireMatch:match("^%s*[\"'][^\"']+[\"']%s*$") then return false end
    return true
end

-- Nested prototype: GitHub tab button (for GitHub form)
AppConsole.GitHubTab = session:prototype("AppConsole.GitHubTab", {
    filename = "",  -- Display name and tab key (e.g., "requirements.md", "app.lua")
    pushStateCount = 0,  -- Count of pushState calls (for Lua files)
    dangerCount = 0,  -- Count of dangerous calls (os.execute, io.open, etc.)
    _contentHtml = nil,  -- Cached HTML content (generated once on first click)
    _totalLines = 0,  -- Total line count (for trough markers)
    _warningLines = nil,  -- {line=n, type="pushstate"|"danger"} array
})
local GitHubTab = AppConsole.GitHubTab

function GitHubTab:selectMe()
    appConsole.github:selectTab(self.filename)
end

function GitHubTab:isSelected()
    return appConsole.github.activeTab == self.filename
end

function GitHubTab:buttonVariant()
    if self:isSelected() then return "primary" end
    if appConsole.github:isTabViewed(self.filename) then return "default" end
    return "warning"  -- Unviewed tabs shown as warning to draw attention
end

-- Build a warnings list from pushState/danger counts
-- labelFn(count, type) returns the string for each warning
local function collectWarnings(tab, labelFn)
    local warnings = {}
    if tab.pushStateCount > 0 then
        table.insert(warnings, labelFn(tab.pushStateCount, "pushState"))
    end
    if tab.dangerCount > 0 then
        table.insert(warnings, labelFn(tab.dangerCount, "danger"))
    end
    return warnings
end

function GitHubTab:buttonLabel()
    local warnings = collectWarnings(self, function(count, wtype)
        if wtype == "pushState" then return count .. " events" end
        return count .. " danger"
    end)
    if #warnings == 0 then return self.filename end
    return self.filename .. " (" .. table.concat(warnings, ", ") .. ")"
end

function GitHubTab:isLuaFile()
    return self.filename:match("%.lua$") ~= nil
end

function GitHubTab:tooltipText()
    if not self:isLuaFile() then return self.filename end
    local warnings = collectWarnings(self, function(count, wtype)
        if wtype == "pushState" then return count .. " pushState call(s) - can send events to Claude" end
        return count .. " dangerous call(s) - shell, file, or code loading"
    end)
    if #warnings == 0 then return "Lua file - no dangerous calls found" end
    return table.concat(warnings, "\n")
end

-- Load and cache content (called once on first selection)
function GitHubTab:loadContent()
    if self._contentHtml ~= nil then return end  -- Already loaded

    local content = appConsole.github:fetchFile(self.filename)
    if not content then
        self._contentHtml = "Failed to load " .. self.filename
        return
    end

    -- generateHtml computes pushStateCount and dangerCount for Lua files
    self._contentHtml = self:generateHtml(content)
end

-- Generate HTML with highlighting for Lua files (pushState and shell commands).
-- Also computes pushStateCount and dangerCount during the line iteration.
function GitHubTab:generateHtml(content)
    if not self:isLuaFile() then
        local lineCount = 1
        for _ in content:gmatch("\n") do lineCount = lineCount + 1 end
        self._totalLines = lineCount
        self._warningLines = {}
        return escapeHtml(content)
    end

    -- Helper to check if line matches any danger pattern
    local function isDangerLine(line)
        for _, pat in ipairs(DANGER_PATTERNS) do
            if line:match(pat.pattern) then return true end
        end
        return isDangerousRequire(line)
    end

    -- Lua files: highlight pushState blocks and dangerous calls
    local lines = {}
    local warningLines = {}
    local inPushState = false
    local braceDepth = 0
    local lineNum = 0
    local pushCount = 0
    local dangerCount = 0

    for line in (content .. "\n"):gmatch("([^\n]*)\n") do
        lineNum = lineNum + 1
        local escaped = escapeHtml(line)

        local wasInPushState = inPushState
        if not inPushState and line:match("pushState%s*%(") then
            inPushState = true
            braceDepth = 0
            pushCount = pushCount + 1
        end

        if inPushState then
            for _ in line:gmatch("{") do braceDepth = braceDepth + 1 end
            for _ in line:gmatch("}") do braceDepth = braceDepth - 1 end
            local innerSpan = '<span class="pushstate-highlight-line">' .. escaped .. '</span>'
            if not wasInPushState then
                innerSpan = '<span class="pushstate-group">' .. innerSpan
            end
            if braceDepth <= 0 then
                innerSpan = innerSpan .. '</span>'
                inPushState = false
            end
            escaped = innerSpan
            table.insert(warningLines, {line = lineNum, type = "pushstate"})
        elseif isDangerLine(line) then
            escaped = '<span class="osexecute-group"><span class="osexecute-highlight-line">' .. escaped .. '</span></span>'
            table.insert(warningLines, {line = lineNum, type = "danger"})
            dangerCount = dangerCount + 1
        end

        table.insert(lines, escaped)
    end

    self._totalLines = lineNum
    self._warningLines = warningLines
    self.pushStateCount = pushCount
    self.dangerCount = dangerCount
    return table.concat(lines, "\n")
end

function GitHubTab:contentHtml()
    return self._contentHtml or ""
end

-- Nested prototype: GitHub downloader (owns all GitHub download state)
AppConsole.GitHubDownloader = session:prototype("AppConsole.GitHubDownloader", {
    visible = false,    -- Whether form is visible
    url = "",           -- URL input
    validated = false,  -- Whether URL has been validated
    error = "",         -- Error message if invalid
    activeTab = "",     -- Currently selected tab filename
    tabs = EMPTY,       -- GitHubTab instances
    viewedTabs = EMPTY, -- Tracks which tabs user has clicked
    markerRefresh = "", -- JS code to trigger trough marker positioning
    _conflict = false,  -- Whether app name conflicts with existing app
    _conflictCheckTime = 0, -- Last time conflict was checked
    _markerCounter = 0  -- Counter for markerRefresh changes
})
local GitHubDownloader = AppConsole.GitHubDownloader

function GitHubDownloader:new()
    local instance = session:create(GitHubDownloader, {})
    instance.tabs = {}
    instance.viewedTabs = {}
    return instance
end

function GitHubDownloader:show()
    self.visible = true
    appConsole.showNewForm = false
    appConsole.selected = nil
end

function GitHubDownloader:hide()
    self.visible = false
end

function GitHubDownloader:cancel()
    self.visible = false
    self.url = ""
    self.validated = false
    self.error = ""
    self.activeTab = ""
    self.tabs = {}
    self.viewedTabs = {}
end

function GitHubDownloader:isHidden()
    return not self.visible
end

function GitHubDownloader:hasUrl()
    return self.url ~= nil and self.url ~= ""
end

function GitHubDownloader:parseUrl()
    if not self:hasUrl() then return nil end
    local user, repo, branch, path = self.url:match("github%.com/([^/]+)/([^/]+)/tree/([^/]+)/?(.*)")
    if not user then return nil end
    return { user = user, repo = repo, branch = branch, path = path or "" }
end

function GitHubDownloader:getAppName()
    local parsed = self:parseUrl()
    if not parsed or parsed.path == "" then return nil end
    return parsed.path:match("([^/]+)$")
end

function GitHubDownloader:checkConflict()
    local now = os.time()
    if (now - self._conflictCheckTime) < 1 then
        return self._conflict
    end
    self._conflictCheckTime = now

    local appName = self:getAppName()
    if not appName then
        self._conflict = false
        return false
    end

    local status = mcp:status()
    local baseDir = status and status.base_dir
    if not baseDir then
        self._conflict = false
        return false
    end

    local targetDir = baseDir .. "/apps/" .. appName
    local handle = io.popen('test -d "' .. targetDir .. '" && echo "exists"')
    if handle then
        local result = handle:read("*a")
        handle:close()
        self._conflict = result:match("exists") ~= nil
    else
        self._conflict = false
    end
    return self._conflict
end

function GitHubDownloader:hasConflict()
    return self:checkConflict()
end

function GitHubDownloader:investigateDisabled()
    return self:hasConflict()
end

function GitHubDownloader:conflictMessage()
    local appName = self:getAppName()
    if appName and self:hasConflict() then
        return "App '" .. appName .. "' already exists in .ui/apps/. Delete or rename it before downloading."
    end
    return ""
end

function GitHubDownloader:showConflict()
    return self:hasConflict() and self:hasUrl()
end

function GitHubDownloader:hideConflict()
    return not self:showConflict()
end

function GitHubDownloader:fetchFile(filename)
    local parsed = self:parseUrl()
    if not parsed then return nil end

    local path = parsed.path
    if path ~= "" then path = path .. "/" end

    local rawUrl = string.format(
        "https://raw.githubusercontent.com/%s/%s/%s/%s%s",
        parsed.user, parsed.repo, parsed.branch, path, filename
    )

    local handle = io.popen('curl -sL "' .. rawUrl .. '"')
    if handle then
        local content = handle:read("*a")
        handle:close()
        if content and content ~= "" and not content:match("^404:") then
            return content
        end
    end
    return nil
end

function GitHubDownloader:listDir()
    local parsed = self:parseUrl()
    if not parsed then return {} end

    local apiUrl = string.format(
        "https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
        parsed.user, parsed.repo, parsed.path, parsed.branch
    )

    local handle = io.popen('curl -sL "' .. apiUrl .. '"')
    if not handle then return {} end

    local content = handle:read("*a")
    handle:close()

    local files = {}
    for name, ftype in content:gmatch('"name"%s*:%s*"([^"]+)"[^}]-"type"%s*:%s*"([^"]+)"') do
        table.insert(files, { name = name, type = ftype })
    end
    return files
end

function GitHubDownloader:investigate()
    self.validated = false
    self.error = ""
    self.activeTab = ""
    self.tabs = {}
    self.viewedTabs = {}

    local parsed = self:parseUrl()
    if not parsed then
        self.error = "Invalid GitHub URL. Expected format: https://github.com/user/repo/tree/branch/path"
        return
    end

    local files = self:listDir()
    if #files == 0 then
        self.error = "Could not fetch directory contents. Check the URL and try again."
        return
    end

    local hasRequirements, hasDesign, hasAppLua, hasViewdefs = false, false, false, false
    local extraLua = {}

    for _, f in ipairs(files) do
        if f.name == "requirements.md" and f.type == "file" then hasRequirements = true
        elseif f.name == "design.md" and f.type == "file" then hasDesign = true
        elseif f.name == "app.lua" and f.type == "file" then hasAppLua = true
        elseif f.name == "viewdefs" and f.type == "dir" then hasViewdefs = true
        elseif f.name:match("%.lua$") and f.type == "file" and f.name ~= "app.lua" then
            table.insert(extraLua, f.name)
        end
    end

    local missing = {}
    if not hasRequirements then table.insert(missing, "requirements.md") end
    if not hasDesign then table.insert(missing, "design.md") end
    if not hasAppLua then table.insert(missing, "app.lua") end
    if not hasViewdefs then table.insert(missing, "viewdefs directory") end

    if #missing > 0 then
        self.error = "This is not a valid app directory, it does not have: " .. table.concat(missing, ", ")
        return
    end

    self.validated = true
    self.tabs = {
        session:create(GitHubTab, { filename = "requirements.md" }),
        session:create(GitHubTab, { filename = "design.md" }),
        session:create(GitHubTab, { filename = "app.lua" }),
    }
    for _, name in ipairs(extraLua) do
        table.insert(self.tabs, session:create(GitHubTab, { filename = name }))
    end
end

function GitHubDownloader:selectTab(filename)
    self.activeTab = filename
    self.viewedTabs[filename] = true
    local tab = self:getActiveTab()
    if tab then tab:loadContent() end
    self._markerCounter = self._markerCounter + 1
    self.markerRefresh = "element.positionMarkers() // " .. self._markerCounter
end

function GitHubDownloader:getActiveTab()
    for _, tab in ipairs(self.tabs) do
        if tab.filename == self.activeTab then
            return tab
        end
    end
    return nil
end

function GitHubDownloader:isTabViewed(filename)
    return self.viewedTabs[filename] == true
end

function GitHubDownloader:allTabsViewed()
    for _, tab in ipairs(self.tabs) do
        if not self.viewedTabs[tab.filename] then
            return false
        end
    end
    return #self.tabs > 0
end

function GitHubDownloader:approveDisabled()
    return not self:allTabsViewed()
end

function GitHubDownloader:hasError()
    return self.error ~= nil and self.error ~= ""
end

function GitHubDownloader:noError()
    return not self:hasError()
end

function GitHubDownloader:hideTabs()
    return not self.validated
end

function GitHubDownloader:hasContent()
    local tab = self:getActiveTab()
    return tab and tab._contentHtml ~= nil
end

function GitHubDownloader:noContent()
    return not self:hasContent()
end

function GitHubDownloader:showSafetyMessage()
    return self.validated and not self:hasContent()
end

function GitHubDownloader:hideSafetyMessage()
    return not self:showSafetyMessage()
end

function GitHubDownloader:contentHtml()
    local tab = self:getActiveTab()
    return tab and tab:contentHtml() or ""
end

-- Returns empty; JS in viewdef positions markers by measuring actual span positions
function GitHubDownloader:troughMarkersHtml()
    return ""
end

function GitHubDownloader:noTroughMarkers()
    local tab = self:getActiveTab()
    return not tab or not tab._warningLines or #tab._warningLines == 0
end

function GitHubDownloader:approve()
    if not self.validated then
        self.error = "Please investigate the URL first to validate it's a valid app."
        return
    end

    local parsed = self:parseUrl()
    if not parsed then
        self.error = "Invalid GitHub URL."
        return
    end

    local appName = parsed.path:match("([^/]+)$")
    if not appName or appName == "" then
        self.error = "Could not determine app name from URL."
        return
    end

    local status = mcp:status()
    local baseDir = status and status.base_dir
    if not baseDir then
        self.error = "Could not determine base directory."
        return
    end

    local targetDir = baseDir .. "/apps/" .. appName

    if self:hasConflict() then
        self.error = "App '" .. appName .. "' already exists. Delete it first to reinstall."
        return
    end

    local tempDir = "/tmp/github-download-" .. os.time()
    local zipUrl = string.format(
        "https://github.com/%s/%s/archive/refs/heads/%s.zip",
        parsed.user, parsed.repo, parsed.branch
    )

    local cmd = string.format(
        'mkdir -p "%s" && ' ..
        'curl -sL "%s" -o "%s/repo.zip" && ' ..
        'unzip -q "%s/repo.zip" -d "%s" && ' ..
        'mv "%s/%s-%s/%s" "%s" && ' ..
        'rm -rf "%s"',
        tempDir,
        zipUrl, tempDir,
        tempDir, tempDir,
        tempDir, parsed.repo, parsed.branch, parsed.path, targetDir,
        tempDir
    )

    local handle = io.popen(cmd .. ' 2>&1; echo "EXIT:$?"')
    if handle then
        local output = handle:read("*a")
        handle:close()

        local exitCode = output:match("EXIT:(%d+)")
        if exitCode == "0" then
            -- Link the app to make it available
            os.execute('.ui/mcp linkapp add ' .. appName)

            -- Save the source URL for reference
            local sourceFile = io.open(targetDir .. "/source.txt", "w")
            if sourceFile then
                sourceFile:write(self.url .. "\n")
                sourceFile:close()
            end

            -- Create original.fossil to track that this was downloaded
            -- and to enable local changes detection
            os.execute('.ui/mcp checkpoint baseline ' .. appName)
            local originalCmd = string.format(
                'cp "%s/checkpoint.fossil" "%s/original.fossil"',
                targetDir, targetDir
            )
            os.execute(originalCmd)

            self:cancel()
            appConsole:refresh()
            local newApp = appConsole:findApp(appName)
            if newApp then
                appConsole:select(newApp)
            end
        else
            self.error = "Failed to download app. Check the URL and try again."
        end
    else
        self.error = "Failed to execute download command."
    end
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

local TEST_STATUS_ICONS = {
    passed = "✓",
    failed = "✗",
    untested = " "
}

function TestItem:icon()
    return TEST_STATUS_ICONS[self.status] or " "
end

function TestItem:iconClass()
    return self.status
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
    buildStage = EMPTY,
    confirmDelete = false,
    _isDownloaded = false,  -- Has original.fossil (downloaded from GitHub)
    _hasLocalChanges = false,  -- Has local modifications vs original
    sourceUrl = "",  -- GitHub URL from source.txt
    readmePath = ""  -- Path to readme file (case insensitive)
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
    if self.buildProgress then return self.buildStage or "building..." end
    if not self.hasViewdefs then return "not built" end
    if self.testsTotal == 0 then return "--" end
    return self.testsPassing .. "/" .. self.testsTotal
end

function AppInfo:statusVariant()
    if self.buildProgress then return "primary" end
    if not self.hasViewdefs or self.testsTotal == 0 then return "neutral" end
    if self.testsPassing == self.testsTotal then return "success" end
    return "warning"
end

function AppInfo:noIssues()
    return #self.knownIssues == 0
end

function AppInfo:canOpen()
    return self.hasViewdefs
end

function AppInfo:needsBuild()
    return not self:canOpen()
end

AppInfo.isBuilt = AppInfo.canOpen

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
    return not self:isBuilding()
end

-- Helper to create toggle/hidden/icon methods for collapsible sections
local function makeCollapsible(proto, fieldName)
    local showField = "show" .. fieldName
    proto["toggle" .. fieldName] = function(self)
        self[showField] = not self[showField]
    end
    proto[fieldName:sub(1,1):lower() .. fieldName:sub(2) .. "Hidden"] = function(self)
        return not self[showField]
    end
    proto[fieldName:sub(1,1):lower() .. fieldName:sub(2) .. "Icon"] = function(self)
        return self[showField] and "chevron-down" or "chevron-right"
    end
end

makeCollapsible(AppInfo, "KnownIssues")
makeCollapsible(AppInfo, "FixedIssues")
makeCollapsible(AppInfo, "Gaps")
makeCollapsible(AppInfo, "Requirements")

function AppInfo:hasGaps()
    return self.gapsContent ~= nil and self.gapsContent ~= ""
end

function AppInfo:noGaps()
    return not self:hasGaps()
end

function AppInfo:hasCheckpoints()
    -- Trigger batch refresh if cache is stale
    local now = os.time()
    if not appConsole._checkpointsTime or (now - appConsole._checkpointsTime) >= 1 then
        appConsole:refreshCheckpoints()
    end
    return self._hasCheckpoints or false
end

function AppInfo:checkpointIcon()
    return self:hasCheckpoints() and "rocket" or "gem"
end

function AppInfo:localChangesIcon()
    return self:hasLocalChanges() and "pencil" or ""
end

function AppInfo:showLocalChangesIcon()
    return self:hasLocalChanges()
end

function AppInfo:hideLocalChangesIcon()
    return not self:hasLocalChanges()
end

function AppInfo:checkpointCount()
    self:hasCheckpoints()  -- triggers refresh if needed
    return self._checkpointCount or 0
end

function AppInfo:checkpointTooltip()
    local count = self:checkpointCount()
    return count .. " pending change" .. (count == 1 and "" or "s")
end

function AppInfo:shouldPulsate()
    -- Clear the flag when todos are empty
    if not mcp:hasTodos() then
        self._consolidatePending = false
    end
    return self._consolidatePending
end

function AppInfo:consolidateButtonText()
    local count = self:checkpointCount()
    return "Make it thorough (" .. count .. ")"
end

function AppInfo:isDownloaded()
    return self._isDownloaded or false
end

function AppInfo:hasLocalChanges()
    return self._hasLocalChanges or false
end

function AppInfo:hasSourceUrl()
    return self.sourceUrl and self.sourceUrl ~= ""
end

function AppInfo:noSourceUrl()
    return not self:hasSourceUrl()
end

function AppInfo:hasReadme()
    return self.readmePath and self.readmePath ~= ""
end

function AppInfo:noReadme()
    return not self:hasReadme()
end

-- Generate HTML link for readme (opens in new tab via MCP endpoint)
function AppInfo:readmeLinkHtml()
    if not self:hasReadme() then return "" end
    local status = mcp:status()
    local port = status and status.mcp_port or 8000
    return string.format('<a href="http://localhost:%d/app/%s/readme" target="_blank" title="View readme"><sl-icon name="file-text"></sl-icon></a>', port, self.name)
end

function AppInfo:openReadme()
    if self:hasReadme() then
        os.execute('xdg-open "' .. self.readmePath .. '" 2>/dev/null &')
    end
end

function AppInfo:openSourceUrl()
    if self:hasSourceUrl() then
        os.execute('xdg-open "' .. self.sourceUrl .. '" 2>/dev/null &')
    end
end

function AppInfo:noLocalChanges()
    return not self:hasLocalChanges()
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

function AppInfo:noCheckpoints()
    return not self:hasCheckpoints()
end

function AppInfo:requestConsolidate()
    self._consolidatePending = true
    self:pushEvent("consolidate_request", { target = self.name })
end

function AppInfo:requestReviewGaps()
    self:pushEvent("review_gaps_request", { target = self.name })
end

function AppInfo:requestAnalyze()
    self:pushEvent("analyze_request", { target = self.name })
end

function AppInfo:openApp()
    appConsole:openEmbedded(self.name)
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

-- Protected apps that cannot be deleted
local PROTECTED_APPS = {
    ["app-console"] = true,
    ["mcp"] = true,
    ["claude-panel"] = true,
    ["viewlist"] = true
}

function AppInfo:isProtected()
    return PROTECTED_APPS[self.name] or false
end

function AppInfo:requestDelete()
    self.confirmDelete = true
end

function AppInfo:cancelDelete()
    self.confirmDelete = false
end

function AppInfo:hideDeleteConfirm()
    return not self.confirmDelete
end

-- Convert kebab-case to camelCase (e.g., "my-app" -> "myApp")
local function toCamelCase(name)
    return name:gsub("%-(%l)", function(c) return c:upper() end)
end

-- Convert kebab-case to PascalCase (e.g., "my-app" -> "MyApp")
local function toPascalCase(name)
    return toCamelCase(name):gsub("^%l", string.upper)
end

function AppInfo:confirmDeleteApp()
    if self:isProtected() then return end

    local name = self.name
    local baseDir = mcp:status().base_dir
    local protoName = toPascalCase(name)

    -- Clear global variables for this app
    _G[protoName] = nil
    _G[toCamelCase(name)] = nil

    -- Remove prototype and all nested prototypes (e.g., Contacts.Contact, Contacts.ChatMessage)
    session:removePrototype(protoName, true)

    -- Unlink and delete the app
    os.execute('.ui/mcp linkapp remove "' .. name .. '"')
    os.execute('rm -rf "' .. baseDir .. '/apps/' .. name .. '"')

    -- Remove from app list and clear selection
    for i, app in ipairs(appConsole._apps) do
        if app.name == name then
            table.remove(appConsole._apps, i)
            break
        end
    end
    appConsole.selected = nil

    mcp:scanAvailableApps()
end

-- Populate app from parsed test data (DRY helper)
function AppInfo:populateFromTestData(testData)
    self.testsPassing = testData.testsPassing
    self.testsTotal = testData.testsTotal

    self.tests = {}
    for _, t in ipairs(testData.tests) do
        table.insert(self.tests, TestItem:new(t.text, t.status))
    end

    self.knownIssues = {}
    for _, issue in ipairs(testData.knownIssues) do
        table.insert(self.knownIssues, Issue:new(issue.number, issue.title))
    end

    self.fixedIssues = {}
    for _, issue in ipairs(testData.fixedIssues) do
        table.insert(self.fixedIssues, Issue:new(issue.number, issue.title))
    end

    self.gapsContent = testData.gapsContent or ""
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

-- Construct GitHub readme URL from a repo URL
local function findReadme(repoUrl)
    if not repoUrl or repoUrl == "" then return nil end
    local user, repo, branch, path = repoUrl:match("github%.com/([^/]+)/([^/]+)/tree/([^/]+)/?(.*)")
    if not user or not repo or not branch then return nil end
    if path and path ~= "" then
        return string.format("https://github.com/%s/%s/blob/%s/%s/README.md", user, repo, branch, path)
    end
    return string.format("https://github.com/%s/%s/blob/%s/README.md", user, repo, branch)
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

    -- Parse issue sections (Known Issues and Fixed Issues share same format)
    local function parseIssueSection(sectionName)
        local issues = {}
        local section = content:match("## " .. sectionName .. ".-\n(.-)\n## ")
            or content:match("## " .. sectionName .. ".-\n(.*)$")
        if section then
            for num, title in section:gmatch("### (%d+)%.%s*([^\n]+)") do
                table.insert(issues, { number = tonumber(num), title = title })
            end
        end
        return issues
    end

    result.knownIssues = parseIssueSection("Known Issues")
    result.fixedIssues = parseIssueSection("Fixed Issues")

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
    instance.github = GitHubDownloader:new()
    return instance
end

-- Hot-load mutation: initialize new fields on existing instances
function AppConsole:mutate()
    if not self.github then
        self.github = GitHubDownloader:new()
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

            -- Read source URL for downloaded apps
            local sourceContent = readFile(appPath .. "/source.txt")
            if sourceContent then
                app.sourceUrl = sourceContent:gsub("%s+$", "")  -- trim trailing whitespace
                app.readmePath = findReadme(app.sourceUrl)
            end

            -- Parse TESTING.md and populate test data
            local testContent = readFile(appPath .. "/TESTING.md")
            app:populateFromTestData(parseTesting(testContent))

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

    -- Read source URL for downloaded apps
    local sourceContent = readFile(appPath .. "/source.txt")
    if sourceContent then
        app.sourceUrl = sourceContent:gsub("%s+$", "")  -- trim trailing whitespace
        app.readmePath = findReadme(app.sourceUrl)
    else
        app.sourceUrl = ""
        app.readmePath = ""
    end

    -- Clear build progress since we've rescanned
    app.buildProgress = nil
    app.buildStage = nil

    -- Parse TESTING.md and populate test data
    local testContent = readFile(appPath .. "/TESTING.md")
    app:populateFromTestData(parseTesting(testContent))
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
    self:rescanApp(name)
end

-- Refresh: rescan all apps from disk (Lua-driven)
function AppConsole:refresh()
    mcp:scanAvailableApps()  -- sync MCP server's app list with disk
    self:scanAppsFromDisk()
end

-- Batch refresh checkpoint status for all apps (cached for 1 second)
function AppConsole:refreshCheckpoints()
    local status = mcp:status()
    local baseDir = status and status.base_dir
    if not baseDir then
        for _, app in ipairs(self._apps) do
            app._hasCheckpoints = false
            app._checkpointCount = 0
            app._isDownloaded = false
            app._hasLocalChanges = false
        end
        self._checkpointsTime = os.time()
        return
    end

    local fossilBin = os.getenv("HOME") .. "/.claude/bin/fossil"
    for _, app in ipairs(self._apps) do
        local appDir = baseDir .. "/apps/" .. app.name

        local cmd = baseDir .. "/mcp checkpoint count " .. app.name .. " 2>/dev/null"
        local countHandle = io.popen(cmd)
        local count = 0
        if countHandle then
            count = tonumber(countHandle:read("*a")) or 0
            countHandle:close()
        end
        app._hasCheckpoints = count > 0
        app._checkpointCount = count

        -- Check if this is a downloaded app (has original.fossil)
        local originalCheck = io.open(appDir .. "/original.fossil", "r")
        if originalCheck then
            originalCheck:close()
            app._isDownloaded = true

            -- Check for local changes vs original
            local diffCmd = string.format(
                'cd "%s" && "%s" diff --from original --brief 2>/dev/null | head -1',
                appDir, fossilBin
            )
            local diffHandle = io.popen(diffCmd)
            if diffHandle then
                local diffOutput = diffHandle:read("*a")
                diffHandle:close()
                app._hasLocalChanges = diffOutput and diffOutput:match("%S") ~= nil
            end
        else
            app._isDownloaded = false
            app._hasLocalChanges = false
        end
    end
    self._checkpointsTime = os.time()
end

-- Select an app
function AppConsole:select(app)
    self.selected = app
    self.showNewForm = false
    self.github:hide()
end

-- Show new app form
function AppConsole:openNewForm()
    self.showNewForm = true
    self.github:hide()
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

function AppConsole:showDetail()
    return self.selected ~= nil and not self.showNewForm
end

function AppConsole:hideDetail()
    return not self:showDetail()
end

function AppConsole:showPlaceholder()
    return self.selected == nil and not self.showNewForm and self.github:isHidden()
end

function AppConsole:hidePlaceholder()
    return not self:showPlaceholder()
end

function AppConsole:hideNewForm()
    return not self.showNewForm
end

-- GitHub form delegators (actions delegate to self.github)
function AppConsole:openGitHubForm()
    self.github:show()
end

function AppConsole:cancelGitHubForm()
    self.github:cancel()
end

function AppConsole:investigateGitHub()
    self.github:investigate()
end

function AppConsole:approveGitHub()
    self.github:approve()
end

function AppConsole:openEmbedded(name)
    -- Handle case where name might be an AppInfo object instead of string
    if type(name) == "table" and type(name.name) == "string" then
        name = name.name
    end
    if type(name) ~= "string" or name == "" then
        return
    end
    local appValue = mcp:app(name)
    if appValue then
        self.embeddedValue = appValue
        self.embeddedApp = name
    end
end

function AppConsole:closeEmbedded()
    self.embeddedApp = nil
    self.embeddedValue = nil
end

function AppConsole:hasEmbeddedApp()
    return self.embeddedApp ~= nil
end

function AppConsole:noEmbeddedApp()
    return not self:hasEmbeddedApp()
end

function AppConsole:updateRequirements(name)
    local app = self:findApp(name)
    if app and self._baseDir then
        local reqPath = self._baseDir .. "/apps/" .. name .. "/requirements.md"
        local content = readFile(reqPath)
        if content then
            app.requirementsContent = content
            app.description = parseRequirements(content)
        end
    end
end

if not session.reloading then
    appConsole = AppConsole:new()
    appConsole:scanAppsFromDisk()
end

