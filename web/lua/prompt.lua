-- Prompt support for permission dialogs
-- CRC: crc-PromptViewdef.md
-- Spec: prompt-ui.md

-- PromptOption model - represents a single option in a permission prompt
-- Each option knows its value/label and has a respond() method for viewdef buttons
PromptOption = {type = "PromptOption"}
PromptOption.__index = PromptOption

-- Create a new PromptOption instance
-- @param tbl table - initial values (value, label required)
-- @param prompt Prompt - parent prompt reference
-- @return PromptOption instance
function PromptOption:new(tbl, prompt)
    tbl = tbl or {}
    setmetatable(tbl, self)
    tbl.type = "PromptOption"
    tbl._prompt = prompt  -- reference to parent Prompt
    return tbl
end

-- Respond with this option (zero-argument method for viewdef buttons)
function PromptOption:respond()
    if self._prompt then
        self._prompt:respondWith(self.value, self.label)
    end
end

-- Prompt model using prototype pattern
-- Encapsulates permission prompt state and response handling
Prompt = {type = "Prompt"}
Prompt.__index = Prompt

-- Create a new Prompt instance
-- @param tbl table - initial values (id, message, items required)
-- @return Prompt instance
function Prompt:new(tbl)
    tbl = tbl or {}
    setmetatable(tbl, self)
    tbl.type = "Prompt"

    -- Wrap raw options in PromptOption instances
    if tbl.items then
        local wrappedItems = {}
        for i, opt in ipairs(tbl.items) do
            wrappedItems[i] = PromptOption:new(opt, tbl)
        end
        tbl.items = wrappedItems
    end

    return tbl
end

-- Respond to the prompt with the given value and label
-- Called by PromptOption:respond()
-- @param value string - the selected option value
-- @param label string - the selected option label
function Prompt:respondWith(value, label)
    -- Call Go callback to unblock the prompt
    mcp.promptResponse(self.id, value, label or value)

    -- Clear the prompt and restore previous presenter
    local app = session:getApp()
    if app then
        app.pendingPrompt = nil
        if app._previousPresenter then
            app._presenter = app._previousPresenter
            app._previousPresenter = nil
        end
    end
end
