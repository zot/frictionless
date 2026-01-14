-- Claude Panel: Universal panel for Claude Code
-- Design: design.md
-- Hot-loadable: uses session:prototype() for live code updates

-- Chat message model
ChatMessage = session:prototype("ChatMessage", {
    sender = "",
    text = ""
})

-- Output line model (for Lua console)
OutputLine = session:prototype("OutputLine", {
    text = "",
    panel = EMPTY
})

function OutputLine:copyToInput()
    -- Strip leading "> " from command lines when copying
    local text = self.text
    if text:match("^> ") then
        text = text:sub(3)
    end
    self.panel.luaInput = text
end

-- Tree item model
TreeItem = session:prototype("TreeItem", {
    name = "",
    section = EMPTY
})

function TreeItem:invoke()
    mcp.pushState({
        app = "claude-panel",
        event = "invoke",
        type = self.section.itemType,
        name = self.name
    })
end

-- Tree section model
TreeSection = session:prototype("TreeSection", {
    name = "",
    itemType = "",
    expanded = false,
    items = EMPTY
})

function TreeSection:new(instance)
    instance = session:create(TreeSection, instance)
    instance.items = instance.items or {}
    return instance
end

function TreeSection:toggle()
    self.expanded = not self.expanded
end

function TreeSection:isCollapsed()
    return not self.expanded
end

function TreeSection:itemCount()
    return "(" .. #self.items .. ")"
end

function TreeSection:addItem(name)
    local item = session:create(TreeItem, { name = name, section = self })
    table.insert(self.items, item)
end

-- Main app
ClaudePanel = session:prototype("ClaudePanel", {
    status = "Loading",
    branch = "...",
    changedFiles = 0,
    sections = EMPTY,
    messages = EMPTY,
    chatInput = "",
    jsCode = "",
    consoleExpanded = false,
    luaOutputLines = EMPTY,
    luaInput = ""
})

function ClaudePanel:new(instance)
    instance = session:create(ClaudePanel, instance)
    instance.sections = instance.sections or {}
    instance.messages = instance.messages or {}
    instance.luaOutputLines = instance.luaOutputLines or {}

    -- Create sections if new instance
    if #instance.sections == 0 then
        local agents = TreeSection:new({ name = "Agents", itemType = "agent" })
        local commands = TreeSection:new({ name = "Commands", itemType = "command" })
        local skills = TreeSection:new({ name = "Skills", itemType = "skill" })
        instance.sections = { agents, commands, skills }
    end

    return instance
end

-- Initialize app data (called once on creation, or on refresh)
function ClaudePanel:initialize()
    self:loadGitStatus()
    self:discoverItems()

    -- Add welcome message if none
    if #self.messages == 0 then
        local msg = session:create(ChatMessage, { sender = "Agent", text = "How can I help you today?" })
        table.insert(self.messages, msg)
    end

    self.status = "Ready"
end

-- Quick actions
function ClaudePanel:commitAction()
    mcp.pushState({ app = "claude-panel", event = "action", action = "commit" })
end

function ClaudePanel:testAction()
    mcp.pushState({ app = "claude-panel", event = "action", action = "test" })
end

function ClaudePanel:buildAction()
    mcp.pushState({ app = "claude-panel", event = "action", action = "build" })
end

-- Chat
function ClaudePanel:sendChat()
    if self.chatInput == "" then return end
    local msg = session:create(ChatMessage, { sender = "You", text = self.chatInput })
    table.insert(self.messages, msg)
    mcp.pushState({ app = "claude-panel", event = "chat", text = self.chatInput })
    self.chatInput = ""
end

function ClaudePanel:addAgentMessage(text)
    local msg = session:create(ChatMessage, { sender = "Agent", text = text })
    table.insert(self.messages, msg)
end

-- Git status
function ClaudePanel:loadGitStatus()
    -- Get branch
    local handle = io.popen("git branch --show-current 2>/dev/null")
    if handle then
        local result = handle:read("*a")
        handle:close()
        self.branch = result:gsub("%s+$", "")
        if self.branch == "" then self.branch = "unknown" end
    end

    -- Get changed file count
    handle = io.popen("git status --porcelain 2>/dev/null | wc -l")
    if handle then
        local result = handle:read("*a")
        handle:close()
        self.changedFiles = tonumber(result) or 0
    end
end

-- Discovery helpers
local function scanDir(pattern)
    local items = {}
    local handle = io.popen("ls -1 " .. pattern .. " 2>/dev/null")
    if handle then
        for line in handle:lines() do
            local name = line:match("([^/]+)%.md$") or line:match("([^/]+)/?$")
            if name and name ~= "" then
                items[name] = true
            end
        end
        handle:close()
    end
    return items
end

local function scanSkillDirs(path)
    local items = {}
    local handle = io.popen("ls -1d " .. path .. "*/ 2>/dev/null")
    if handle then
        for line in handle:lines() do
            local name = line:match("([^/]+)/$")
            if name and name ~= "" then
                items[name] = true
            end
        end
        handle:close()
    end
    return items
end

function ClaudePanel:discoverItems()
    local home = os.getenv("HOME") or ""

    -- Agents
    local agents = self.sections[1]
    agents.items = {}  -- Clear existing
    local agentSet = {}

    -- Built-in agents
    local builtIn = {"general-purpose", "Explore", "Plan", "commit"}
    for _, name in ipairs(builtIn) do agentSet[name] = true end

    -- Project agents
    for name in pairs(scanDir(".claude/agents/*.md")) do agentSet[name] = true end

    -- User agents
    for name in pairs(scanDir(home .. "/.claude/agents/*.md")) do agentSet[name] = true end

    -- Sort and add
    local agentList = {}
    for name in pairs(agentSet) do table.insert(agentList, name) end
    table.sort(agentList)
    for _, name in ipairs(agentList) do agents:addItem(name) end

    -- Commands
    local commands = self.sections[2]
    commands.items = {}  -- Clear existing
    local cmdSet = {}

    -- Built-in commands
    local builtInCmds = {
        "/help", "/clear", "/compact", "/config", "/cost", "/doctor",
        "/init", "/login", "/logout", "/memory", "/model", "/permissions",
        "/review", "/status", "/terminal-setup", "/vim"
    }
    for _, name in ipairs(builtInCmds) do cmdSet[name] = true end

    -- Project commands
    for name in pairs(scanDir(".claude/commands/*.md")) do cmdSet["/" .. name] = true end

    -- User commands
    for name in pairs(scanDir(home .. "/.claude/commands/*.md")) do cmdSet["/" .. name] = true end

    -- Sort and add
    local cmdList = {}
    for name in pairs(cmdSet) do table.insert(cmdList, name) end
    table.sort(cmdList)
    for _, name in ipairs(cmdList) do commands:addItem(name) end

    -- Skills
    local skills = self.sections[3]
    skills.items = {}  -- Clear existing
    local skillSet = {}

    -- Project skills
    for name in pairs(scanSkillDirs(".claude/skills/")) do skillSet[name] = true end

    -- User skills
    for name in pairs(scanSkillDirs(home .. "/.claude/skills/")) do skillSet[name] = true end

    -- Sort and add
    local skillList = {}
    for name in pairs(skillSet) do table.insert(skillList, name) end
    table.sort(skillList)
    for _, name in ipairs(skillList) do skills:addItem(name) end
end

-- Changed files display
function ClaudePanel:changedDisplay()
    return self.changedFiles .. " files"
end

-- Polling status indicator (empty when connected, asterisk when not)
function ClaudePanel:pollingIndicator()
    return mcp.pollingEvents() and "" or "*"
end

-- Lua console
function ClaudePanel:toggleConsole()
    self.consoleExpanded = not self.consoleExpanded
end

function ClaudePanel:isConsoleCollapsed()
    return not self.consoleExpanded
end

function ClaudePanel:runLua()
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

function ClaudePanel:appendOutput(text)
    local line = session:create(OutputLine, { text = text, panel = self })
    table.insert(self.luaOutputLines, line)
end

function ClaudePanel:clearOutput()
    self.luaOutputLines = {}
end

-- Guard instance creation (idempotent)
if not session.reloading then
    claudePanel = ClaudePanel:new()
    claudePanel:initialize()
end
