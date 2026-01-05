-- Prompt support for permission dialogs
-- CRC: crc-PromptViewdef.md
-- Spec: prompt-ui.md

-- respondToPrompt is called when user clicks a button in the Prompt viewdef
-- It sends the response back to Go via mcp.promptResponse()
-- @param option table - the selected option with label and value
function respondToPrompt(option)
    local app = session:getApp()
    if app.pendingPrompt and option then
        -- Call Go callback to unblock the prompt
        mcp.promptResponse(app.pendingPrompt.id, option.value, option.label)

        -- Clear the prompt and restore previous presenter
        app.pendingPrompt = nil
        if app._previousPresenter then
            app._presenter = app._previousPresenter
            app._previousPresenter = nil
        end
    end
end
