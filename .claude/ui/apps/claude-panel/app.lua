-- Claude Panel: Universal panel for Claude Code
-- Design: design.md

-- Chat message model
ChatMessage = { type = "ChatMessage" }
ChatMessage.__index = ChatMessage
function ChatMessage:new(sender, text)
    return setmetatable({ sender = sender, text = text }, self)
end

-- Output line model (for Lua console)
OutputLine = { type = "OutputLine" }
OutputLine.__index = OutputLine
function OutputLine:new(text, panel)
    return setmetatable({ text = text, panel = panel }, self)
end

function OutputLine:copyToInput()
    -- Strip leading "> " from command lines when copying
    local text = self.text
    if text:match("^> ") then
        text = text:sub(3)
    end
    self.panel.luaInput = text
end

-- Tree item model
TreeItem = { type = "TreeItem" }
TreeItem.__index = TreeItem
function TreeItem:new(name, section)
    return setmetatable({ name = name, section = section }, self)
end

function TreeItem:invoke()
    mcp.pushState({
        app = "claude-panel",
        event = "invoke",
        type = self.section.itemType,
        name = self.name
    })
end

-- Tree section model
TreeSection = { type = "TreeSection" }
TreeSection.__index = TreeSection
function TreeSection:new(name, itemType)
    return setmetatable({
        name = name,
        itemType = itemType,
        expanded = false,
        items = {}
    }, self)
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
    table.insert(self.items, TreeItem:new(name, self))
end

-- Main app
ClaudePanel = { type = "ClaudePanel" }
ClaudePanel.__index = ClaudePanel

function ClaudePanel:new()
    local self = setmetatable({
        status = "Loading",
        branch = "...",
        changedFiles = 0,
        sections = {},
        messages = {},
        chatInput = "",
        jsCode = "",
        consoleExpanded = false,
        luaOutputLines = {},
        luaInput = ""
    }, ClaudePanel)

    -- Create sections
    local agents = TreeSection:new("Agents", "agent")
    local commands = TreeSection:new("Commands", "command")
    local skills = TreeSection:new("Skills", "skill")

    self.sections = { agents, commands, skills }

    -- Load data
    self:loadGitStatus()
    self:discoverItems()

    -- Add welcome message
    table.insert(self.messages, ChatMessage:new("Agent", "How can I help you today?"))

    self.status = "Ready"
    return self
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
    table.insert(self.messages, ChatMessage:new("You", self.chatInput))
    mcp.pushState({ app = "claude-panel", event = "chat", text = self.chatInput })
    self.chatInput = ""
end

function ClaudePanel:addAgentMessage(text)
    table.insert(self.messages, ChatMessage:new("Agent", text))
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
    table.insert(self.luaOutputLines, OutputLine:new("> " .. self.luaInput, self))

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
                table.insert(self.luaOutputLines, OutputLine:new(tostring(result), self))
            end
            -- Clear input on success
            self.luaInput = ""
        else
            -- Runtime error - keep input for correction
            table.insert(self.luaOutputLines, OutputLine:new("Error: " .. tostring(result), self))
        end
    else
        -- Syntax error - keep input for correction
        table.insert(self.luaOutputLines, OutputLine:new("Syntax error: " .. tostring(err), self))
    end
end

function ClaudePanel:appendOutput(text)
    table.insert(self.luaOutputLines, OutputLine:new(text, self))
end

function ClaudePanel:clearOutput()
    self.luaOutputLines = {}
end

-- Create or reuse app instance
claudePanel = claudePanel or ClaudePanel:new()
