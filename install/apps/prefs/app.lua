-- Prefs app - User preferences management
-- Design: design.md

local Prefs = session:prototype("Prefs", {
    _themes = {},
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

function Prefs:themes()
    return self._themes
end

function Prefs:setCurrentTheme(name)
    self._currentTheme = name
    -- Persist to settings file
    local settings = mcp:readSettings()
    settings.theme = name
    mcp:writeSettings(settings)
    self:applyTheme(name)
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

-- Initialize on load
if not session.reloading then
    prefs = session:create(Prefs, {
        _themes = {
            session:create(ThemeItem, {
                name = "clarity",
                description = "Clean, editorial light theme with slate blue accent",
                accentColor = "#3b6ea5",
                _prefs = nil
            }),
            session:create(ThemeItem, {
                name = "lcars",
                description = "Subtle Star Trek LCARS-inspired design",
                accentColor = "#E07A47",
                _prefs = nil
            }),
            session:create(ThemeItem, {
                name = "midnight",
                description = "Modern dark theme with teal accent",
                accentColor = "#2dd4bf",
                _prefs = nil
            }),
            session:create(ThemeItem, {
                name = "ninja",
                description = "Playful teal theme with cute cartoon ninjas",
                accentColor = "#1a8a99",
                _prefs = nil
            })
        }
    })
    -- Link theme items back to prefs
    for _, theme in ipairs(prefs._themes) do
        theme._prefs = prefs
    end
    -- Load saved theme from settings file
    prefs:loadThemeFromSettings()
end
