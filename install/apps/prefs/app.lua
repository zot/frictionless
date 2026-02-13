-- Prefs app - User preferences management
-- Design: design.md

local Prefs = session:prototype("Prefs", {
    _themes = {},
    _currentTheme = "lcars",
    _browserTheme = ""  -- Set by JS on load via hidden input
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

function Prefs:syncFromBrowser()
    if self._browserTheme ~= "" then
        self._currentTheme = self._browserTheme
    end
end

-- Add missing themes during hot-reload
function Prefs:mutate()
    local themeNames = {}
    for _, theme in ipairs(self._themes) do
        themeNames[theme.name] = true
    end
    -- Add ninja theme if missing
    if not themeNames["ninja"] then
        local ninjaTheme = session:create(ThemeItem, {
            name = "ninja",
            description = "Playful teal theme with cute cartoon ninjas",
            accentColor = "#1a8a99",
            _prefs = self
        })
        table.insert(self._themes, ninjaTheme)
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
end
