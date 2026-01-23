-- Claude Panel: Universal panel for Claude Code
-- Design: design.md
-- Hot-loadable: uses session:prototype() for live code updates

-- App prototype (serves as namespace)
ClaudePanel = session:prototype("ClaudePanel", {
    status = "Loading",
    branch = "...",
    changedFiles = 0,
    sections = EMPTY,
    messages = EMPTY,
    chatInput = "",
    jsCode = ""
})

-- Nested prototype: Chat message model
ClaudePanel.ChatMessage = session:prototype("ClaudePanel.ChatMessage", {
    sender = "",
    text = ""
})
local ChatMessage = ClaudePanel.ChatMessage

-- Nested prototype: Tree item model
ClaudePanel.TreeItem = session:prototype("ClaudePanel.TreeItem", {
    name = "",
    section = EMPTY
})
local TreeItem = ClaudePanel.TreeItem

function TreeItem:invoke()
    mcp.pushState({
        app = "claude-panel",
        event = "invoke",
        type = self.section.itemType,
        name = self.name
    })
end

-- Nested prototype: Tree section model
ClaudePanel.TreeSection = session:prototype("ClaudePanel.TreeSection", {
    name = "",
    itemType = "",
    expanded = false,
    items = EMPTY
})
local TreeSection = ClaudePanel.TreeSection

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

-- Main app methods
function ClaudePanel:new(instance)
    instance = session:create(ClaudePanel, instance)
    instance.sections = instance.sections or {}
    instance.messages = instance.messages or {}

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

-- Hot-load mutation: clear removed fields from existing instances
function ClaudePanel:mutate()
    self.consoleExpanded = nil
    self.luaOutputLines = nil
    self.luaInput = nil
end

-- Idempotent instance creation
if not session.reloading then
    claudePanel = ClaudePanel:new()
    claudePanel:initialize()
end
