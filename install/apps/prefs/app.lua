-- Prefs app - User preferences management
-- Design: design.md

local Prefs = session:prototype("Prefs", {
    _themes = {},
    _themeScanTime = 0,
    _currentTheme = "lcars"
})

Prefs.ThemeItem = session:prototype("Prefs.ThemeItem", {
    name = "",
    description = "",
    accentColor = "",
    _prefs = nil
})
local ThemeItem = Prefs.ThemeItem

function ThemeItem:isSelected()
    return self.name == self._prefs._currentTheme
end

function ThemeItem:select()
    self._prefs:setCurrentTheme(self.name)
end

function ThemeItem:swatchStyle()
    if self.accentColor == "" then return "" end
    return "background-color: " .. self.accentColor .. ";"
end

function Prefs:scanThemes()
    local themeDir = mcp:status().base_dir .. "/html/themes"
    local names = {}
    local handle = io.popen('ls "' .. themeDir .. '"/*.css 2>/dev/null')
    if not handle then return end
    for line in handle:lines() do
        local file = io.open(line, "r")
        if file then
            local head = file:read(2048) or ""
            file:close()
            local name = head:match("@theme%s+(%S+)")
            if name and name ~= "base" then
                local desc = head:match("@description%s+([^\n]+)") or ""
                local accent = head:match("%-%-term%-accent:%s*(#%x+)") or ""
                names[name] = {description = desc, accentColor = accent}
            end
        end
    end
    handle:close()
    -- Build sorted list, reusing existing ThemeItems where possible
    local byName = {}
    for _, t in ipairs(self._themes) do
        byName[t.name] = t
    end
    local result = {}
    local sorted = {}
    for n in pairs(names) do sorted[#sorted + 1] = n end
    table.sort(sorted)
    for _, n in ipairs(sorted) do
        local existing = byName[n]
        if existing then
            existing.description = names[n].description
            existing.accentColor = names[n].accentColor
            result[#result + 1] = existing
        else
            local item = session:create(ThemeItem, {
                name = n,
                description = names[n].description,
                accentColor = names[n].accentColor,
                _prefs = self
            })
            result[#result + 1] = item
        end
    end
    self._themes = result
    self._themeScanTime = os.time()
end

function Prefs:themes()
    local now = os.time()
    if now - (self._themeScanTime or 0) >= 1 then
        self:scanThemes()
    end
    return self._themes
end

function Prefs:themesHidden()
    return #self._themes == 0
end

function Prefs:setCurrentTheme(name)
    self._currentTheme = name
    -- Persist to settings file
    local settings = mcp:readSettings()
    settings.theme = name
    mcp:writeSettings(settings)
    self:applyTheme(name)
    -- Re-inject theme block so future page loads include all theme CSS links
    mcp:reinjectThemes()
end

function Prefs:applyTheme(name)
    mcp.codeCounter = (mcp.codeCounter or 0) + 1
    mcp.code = string.format([[
        const panel = document.querySelector('.prefs-inner');
        if (panel && panel.applyTheme) panel.applyTheme('%s');
        // %d
    ]], name, mcp.codeCounter)
end

function Prefs:loadThemeFromSettings()
    local settings = mcp:readSettings()
    if settings.theme and settings.theme ~= "" then
        self._currentTheme = settings.theme
    end
    self:applyTheme(self._currentTheme)
end

-- Update check preference
function Prefs:checkUpdates()
    return mcp:getUpdatePreference()
end

function Prefs:toggleCheckUpdates()
    local current = mcp:getUpdatePreference()
    mcp:setUpdatePreference(not current)
end

function Prefs:checkNow()
    mcp:checkForUpdates()
    if mcp._needsUpdate then
        mcp:notify("Update available: v" .. mcp.latestVersion .. " — use the star menu to update", "primary")
    else
        mcp:notify("You're up to date (v" .. (mcp:currentVersion() or "?") .. ")", "success")
    end
end

-- Tutorial re-run
function Prefs:startTutorial()
    mcp:startTutorial()
end

-- Initialize on load
if not session.reloading then
    prefs = session:create(Prefs)
    prefs:scanThemes()
    -- Load saved theme from settings file
    prefs:loadThemeFromSettings()
end
