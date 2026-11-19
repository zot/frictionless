-- Claude Panel: Universal panel for Claude Code
-- Design: design.md
-- Hot-loadable: uses session:prototype() for live code updates

-- App prototype (serves as namespace)
ClaudePanel = session:prototype("ClaudePanel", {
    status = "Loading",
    branch = "...",
    changedFiles = 0,
    sections = EMPTY,
    jsCode = ""
})

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

-- Git status
-- Both git calls run off the session; each field fills in when git answers.
-- Outside a git repository git fails: branch is "unknown", changedFiles 0.
function ClaudePanel:loadGitStatus()
    ui.execAsync(self, "_gitBranch", {"git", "branch", "--show-current"}, function(out)
        local branch = (out or ""):gsub("%s+$", "")
        self.branch = branch ~= "" and branch or "unknown"
    end)
    ui.execAsync(self, "_gitStatus", {"git", "status", "--porcelain"}, function(out)
        local count = 0
        for _ in (out or ""):gmatch("[^\n]+") do count = count + 1 end
        self.changedFiles = count
    end)
end

-- Discovery helpers. Dot-entries are skipped, as the shell globs did.

-- Names of the *.md files in dir, without the extension
local function scanDir(dir)
    local items = {}
    for _, e in ipairs(ui.listDir(dir) or {}) do
        if not e.isDir then
            local name = e.name:match("^([^.].*)%.md$")
            if name then items[name] = true end
        end
    end
    return items
end

local function scanSkillDirs(path)
    local items = {}
    for _, e in ipairs(ui.listDir(path) or {}) do
        if e.isDir and e.name:sub(1, 1) ~= "." then items[e.name] = true end
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
    for name in pairs(scanDir(".claude/agents")) do agentSet[name] = true end

    -- User agents
    for name in pairs(scanDir(home .. "/.claude/agents")) do agentSet[name] = true end

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
    for name in pairs(scanDir(".claude/commands")) do cmdSet["/" .. name] = true end

    -- User commands
    for name in pairs(scanDir(home .. "/.claude/commands")) do cmdSet["/" .. name] = true end

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
    self.messages = nil
    self.chatInput = nil
end

-- Idempotent instance creation
if not session.reloading then
    claudePanel = ClaudePanel:new()
    claudePanel:initialize()
end
