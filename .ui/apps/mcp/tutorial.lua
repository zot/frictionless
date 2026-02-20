-- Tutorial walkthrough for the MCP shell
-- Loaded via dofile() from app.lua after all type definitions

MCP.Tutorial = session:prototype("MCP.Tutorial", {
    active = false,
    step = 0,
    repositionCode = "",
    _shell = EMPTY,
})
Tutorial = MCP.Tutorial

function Tutorial:new(shell)
    local instance = session:create(Tutorial, {})
    instance._shell = shell
    return instance
end

--------------------------------------------------------------------
-- Helpers (must precede STEPS so closures can capture them)
--------------------------------------------------------------------

local function exampleAppInstalled()
    if not appConsole then return false end
    for _, app in ipairs(appConsole._apps or {}) do
        if app.name == "example" then return true end
    end
    return false
end

-- Bottom Controls highlight items: {text, buttonIndex}
local CONTROLS_ITEMS = {
    {text = "{} variables inspector", idx = 0},
    {text = "? help docs", idx = 1},
    {text = "wrench app tools", idx = 2},
    {text = "rocket/gem build mode (fast vs thorough)", idx = 3},
    {text = "hourglass foreground/background", idx = 4},
    {text = "speech bubble for the chat panel", idx = 5},
}

--------------------------------------------------------------------
-- Step definitions — each step is self-contained
--------------------------------------------------------------------

local STEPS = {
    -- Step 1: App Menu
    {
        title = "App Menu",
        description = "Tap the grid icon to switch between apps. Your installed apps appear here as an icon grid.",
        selector = ".mcp-menu-button",
        position = "left",
        cycling = true,
        -- Menu open is delayed; handled by JS in triggerReposition
        cleanup = function(tut, shell) shell.menuOpen = false end
    },
    -- Step 2: Connection Status
    {
        title = "Connection Status",
        description = "When Claude disconnects, this button pulses and counts seconds waiting. Use /ui events in Claude Code to reconnect.",
        selector = ".mcp-menu-button",
        position = "left",
        run = function(tut, shell) tut:startFakeWait() end,
        cleanup = function(tut, shell) tut:stopFakeWait() end
    },
    -- Step 3: Status Bar
    {
        title = "Status Bar",
        description = "Shows what Claude is thinking. Watch here for progress updates while Claude works on your apps.",
        selector = ".mcp-status-bar",
        position = "top",
        run = function(tut, shell)
            shell.statusLine = "Reading the design..."
            shell.statusClass = "thinking"
        end,
        cleanup = function(tut, shell)
            shell.statusLine = ""
            shell.statusClass = ""
        end
    },
    -- Step 4: Bottom Controls
    {
        title = "Bottom Controls",
        selector = ".mcp-status-toggles",
        position = "top",
        cycling = true,
        description = function()
            local parts = {"From left to right: "}
            for i, item in ipairs(CONTROLS_ITEMS) do
                if i > 1 then table.insert(parts, ", ") end
                table.insert(parts, '<span data-ctrl-idx="' .. item.idx .. '">' .. item.text .. '</span>')
            end
            table.insert(parts, ".")
            return table.concat(parts)
        end
    },
    -- Step 5: Chat Panel
    {
        title = "Chat Panel",
        selector = "#mcp-chat-panel",
        position = "center-top",
        cycling = true,
        description = '<span data-panel-idx="0">Chat with Claude about the current app.</span> '
            .. '<span data-panel-idx="1">Switch to the Lua tab for a live REPL.</span> '
            .. '<span data-panel-idx="2">The todo column tracks task progress.</span> '
            .. '<span data-panel-idx="3">Drag the top edge to resize.</span>',
        run = function(tut, shell)
            shell.panelOpen = true
            shell.panelMode = "chat"
            tut._savedMessages = shell.messages
            tut._savedLuaOutput = shell.luaOutputLines
            tut._savedTodos = shell.todos
            tut._savedTodosCollapsed = shell.todosCollapsed
            shell.messages = {
                MCP.ChatMessage:new("You", "How does the contacts app work?"),
                MCP.ChatMessage:new("Agent", "It stores names and emails in a searchable list. Select a contact to edit, or click + to add one."),
            }
            shell.luaOutputLines = {
                session:create(MCP.OutputLine, { text = "> #appConsole._apps" }),
                session:create(MCP.OutputLine, { text = "5" }),
                session:create(MCP.OutputLine, { text = "> mcp:currentAppName()" }),
                session:create(MCP.OutputLine, { text = "app-console" }),
            }
            shell.todos = {
                session:create(MCP.TodoItem, { content = "Read requirements", status = "completed" }),
                session:create(MCP.TodoItem, { content = "Write code", status = "completed" }),
                session:create(MCP.TodoItem, { content = "Write viewdefs", status = "in_progress", activeForm = "Writing viewdefs..." }),
                session:create(MCP.TodoItem, { content = "Audit", status = "pending" }),
            }
            shell.todosCollapsed = false
        end,
        cleanup = function(tut, shell)
            if tut._savedMessages then
                shell.messages = tut._savedMessages
                tut._savedMessages = nil
            end
            if tut._savedLuaOutput then
                shell.luaOutputLines = tut._savedLuaOutput
                tut._savedLuaOutput = nil
            end
            if tut._savedTodos then
                shell.todos = tut._savedTodos
                tut._savedTodos = nil
            end
            if tut._savedTodosCollapsed ~= nil then
                shell.todosCollapsed = tut._savedTodosCollapsed
                tut._savedTodosCollapsed = nil
            end
            shell.panelOpen = false
        end
    },
    -- Step 6: App Console
    {
        title = "App Console",
        selector = ".app-list-panel",
        position = "right",
        cycling = true,
        description = '<span data-console-idx="0">The left panel lists all apps</span> with build status and test results. '
            .. '<span data-console-idx="1">Use + to create a new app</span> or '
            .. '<span data-console-idx="2">the GitHub icon to download one</span>.',
        run = function(tut, shell) shell:display("app-console") end
    },
    -- Step 7: Download from GitHub
    {
        title = "Download from GitHub",
        position = "left",
        selector = function()
            if exampleAppInstalled() then return ".app-list-header" end
            return ".github-form"
        end,
        description = function()
            if exampleAppInstalled() then
                return "The example app is already installed. To see the live download demo, delete it first and re-run the tutorial."
            end
            return "We've pre-filled the URL for an example app and fetched its files. The tabs show each file for you to review before installing."
        end,
        run = function(tut, shell)
            if not exampleAppInstalled() and appConsole then
                appConsole:openGitHubForm()
                appConsole.github.url = "https://github.com/zot/frictionless/tree/main/apps/example"
                appConsole.github:investigate()
            end
        end
    },
    -- Step 8: Security Review
    {
        title = "Security Review",
        position = "left",
        selector = function()
            if exampleAppInstalled() then return ".detail-panel" end
            return ".github-content-wrapper"
        end,
        description = function()
            if exampleAppInstalled() then
                return "When downloading apps, each file tab must be reviewed before you can approve. Lua files are scanned: orange highlights show events sent to Claude, red highlights show dangerous system calls."
            end
            return "Each file tab must be reviewed before you can approve. Orange highlights show events sent to Claude (pushState calls). Red highlights show dangerous system calls like os.execute and io.popen. Scrollbar markers help find warnings in long files."
        end,
        run = function(tut, shell)
            if appConsole and appConsole.github and appConsole.github.tabs then
                for _, tab in ipairs(appConsole.github.tabs) do
                    if tab:isLuaFile() and tab.dangerCount > 0 then
                        appConsole.github:selectTab(tab.filename)
                        appConsole.github:scrollToDanger()
                        break
                    end
                end
            end
        end
    },
    -- Step 9: App Info
    {
        title = "App Info",
        selector = ".detail-panel",
        position = "left",
        cycling = true,
        description = '<span data-info-idx="0">Build apps from requirements, test them, fix issues, open them live, or delete them.</span> '
            .. 'Collapsible sections show <span data-info-idx="1">requirements</span>, '
            .. 'test results, and <span data-info-idx="2">known issues</span>.',
        run = function(tut, shell)
            -- Close the GitHub form if it was open from the live demo
            if appConsole and appConsole.github and appConsole.github.visible then
                appConsole.github:cancel()
            end
            if appConsole then
                -- Prefer the example app, then any downloaded app, then any non-protected app
                local example, downloaded, fallback = nil, nil, nil
                for _, app in ipairs(appConsole._apps or {}) do
                    if app.name == "example" then example = app end
                    if app._isDownloaded and not downloaded then downloaded = app end
                    if not fallback and not app:isProtected() then fallback = app end
                end
                local target = example or downloaded or fallback
                if target then appConsole:select(target) end
            end
        end
    },
    -- Step 10: Preferences
    {
        title = "Preferences",
        description = "Find the Prefs app in the app menu to change themes and update settings. You can re-run this tutorial anytime from there.",
        selector = ".mcp-menu-button",
        position = "left"
    },
}

--------------------------------------------------------------------
-- Local helpers
--------------------------------------------------------------------

local function currentStep(self)
    return STEPS[self.step]
end

local function resolveField(s, field)
    local val = s[field]
    if type(val) == "function" then return val() end
    return val or ""
end

-- Undo the side effects of the current step before leaving it
local function cleanupStep(self)
    local s = STEPS[self.step]
    if s and s.cleanup then s.cleanup(self, self._shell) end
end

-- Navigate to a step by number, running its action and repositioning
local function goToStep(self, stepNum)
    local prevStep = self.step
    cleanupStep(self)
    -- Close the GitHub form when leaving the download demo zone (steps 7-8)
    if (prevStep == 7 or prevStep == 8) and (stepNum < 7 or stepNum > 8) then
        if appConsole and appConsole.github and appConsole.github.visible then
            appConsole.github:cancel()
        end
    end
    self.step = stepNum
    self:runAction(stepNum)
    self:triggerReposition()
end

--------------------------------------------------------------------
-- Tutorial lifecycle
--------------------------------------------------------------------

function Tutorial:start()
    self._shell:display("app-console")
    self.active = true
    goToStep(self, 1)
end

function Tutorial:finish()
    cleanupStep(self)
    -- Unconditionally clean up tutorial side-effects
    self._shell.menuOpen = false
    if appConsole and appConsole.github and appConsole.github.visible then
        appConsole.github:cancel()
    end
    self.active = false
    self.step = 0
    -- Clear repositionCode so stale ui-code doesn't re-fire on page reload.
    -- JS-side cleanup happens automatically via the cycling update() auto-cleanup.
    self.repositionCode = ""
    local userSettings = self._shell:readUserSettings()
    userSettings.tutorialCompleted = true
    self._shell:writeUserSettings(userSettings)
end

function Tutorial:next()
    if self.step >= #STEPS then
        self:finish()
        return
    end
    goToStep(self, self.step + 1)
end

function Tutorial:prev()
    if self.step <= 1 then return end
    goToStep(self, self.step - 1)
end

--------------------------------------------------------------------
-- Step actions
--------------------------------------------------------------------

function Tutorial:runAction(stepNum)
    local s = STEPS[stepNum]
    if s and s.run then s.run(self, self._shell) end
end

--------------------------------------------------------------------
-- Fake wait (simulates disconnected state)
--------------------------------------------------------------------

function Tutorial:startFakeWait()
    local shell = self._shell
    if not self._realWaitTime then
        self._realWaitTime = shell.waitTime
    end
    self._tutorialWaitStart = os.time()
    shell.waitTime = function()
        if not self.active then
            self:stopFakeWait()
            return shell:waitTime()
        end
        return os.time() - self._tutorialWaitStart + 10
    end
end

function Tutorial:stopFakeWait()
    local shell = self._shell
    if self._realWaitTime then
        shell.waitTime = self._realWaitTime
        self._realWaitTime = nil
        self._tutorialWaitStart = nil
    end
end

--------------------------------------------------------------------
-- Viewdef binding methods
--------------------------------------------------------------------

function Tutorial:overlayHidden()
    return not self.active
end

function Tutorial:overlayShowing()
    return self.active
end

function Tutorial:title()
    local s = currentStep(self)
    return s and s.title or ""
end

function Tutorial:description()
    local s = currentStep(self)
    if not s then return "" end
    return resolveField(s, "description")
end

function Tutorial:selector()
    local s = currentStep(self)
    if not s then return "" end
    return resolveField(s, "selector")
end

function Tutorial:position()
    local s = currentStep(self)
    return s and s.position or "left"
end

function Tutorial:stepLabel()
    return self.step .. " of " .. #STEPS
end

function Tutorial:isFirstStep()
    return self.step == 1
end

function Tutorial:nextLabel()
    return self.step >= #STEPS and "Finish" or "Next"
end

-- Bridge method: JS polls this to know which step needs highlight cycling
-- Returns step number as string or "0" when inactive/no cycling
function Tutorial:highlightActive()
    if not self.active then return "0" end
    local s = currentStep(self)
    if s and s.cycling then return tostring(self.step) end
    return "0"
end

function Tutorial:deleteExampleHidden()
    return not (self.step == 7 and exampleAppInstalled())
end

function Tutorial:deleteExampleApp()
    if not appConsole then return end
    for _, app in ipairs(appConsole._apps or {}) do
        if app.name == "example" then
            app:confirmDeleteApp()
            -- Re-run the current step to switch from Path B to Path A
            self:runAction(self.step)
            self:triggerReposition()
            return
        end
    end
end

function Tutorial:triggerReposition()
    local sel = self:selector()
    local pos = self:position()
    if sel == "" then return end
    self._repoCounter = (self._repoCounter or 0) + 1
    self.repositionCode = string.format(
        "window._tutReposition(%q, %q, %d) // %d",
        sel, pos, self.step, self._repoCounter
    )
end
