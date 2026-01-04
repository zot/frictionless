-- Contacts app

-- Chat message model
ChatMessage = { type = "ChatMessage" }
ChatMessage.__index = ChatMessage
function ChatMessage:new(sender, text)
    return setmetatable({ sender = sender, text = text }, self)
end

-- Contact model
Contact = { type = "Contact" }
Contact.__index = Contact
function Contact:new(name)
    return setmetatable({
        name = name or "",
        email = "",
        status = "active",
        vip = false
    }, self)
end

function Contact:clone()
    local c = Contact:new(self.name)
    c.email = self.email
    c.status = self.status
    c.vip = self.vip
    return c
end

function Contact:copyFrom(other)
    self.name = other.name
    self.email = other.email
    self.status = other.status
    self.vip = other.vip
end

function Contact:selectMe()
    app:select(self)
end

function Contact:isSelected()
    return app:isEditing(self)
end

-- Main app
ContactApp = { type = "ContactApp" }
ContactApp.__index = ContactApp
function ContactApp:new()
    return setmetatable({
        _allContacts = {},
        searchQuery = "",
        current = nil,        -- Temp contact being edited
        _editing = nil,       -- Original contact (nil = adding new)
        hideDetail = true,
        darkMode = false,
        messages = {},
        chatInput = ""
    }, self)
end

-- Computed: filtered contacts based on searchQuery
function ContactApp:contacts()
    local query = (self.searchQuery or ""):lower()
    local result = {}
    for _, contact in ipairs(self._allContacts) do
        if query == "" then
            table.insert(result, contact)
        else
            local name = (contact.name or ""):lower()
            local email = (contact.email or ""):lower()
            if name:find(query, 1, true) or email:find(query, 1, true) then
                table.insert(result, contact)
            end
        end
    end
    return result
end

function ContactApp:contactCount()
    return #self:contacts()
end

-- Add new contact (creates temp, doesn't insert until save)
function ContactApp:add()
    self.current = Contact:new("New Contact")
    self._editing = nil
    self.hideDetail = false
end

-- Edit existing contact (clones into temp)
function ContactApp:select(contact)
    self.current = contact:clone()
    self._editing = contact
    self.hideDetail = false
end

function ContactApp:isEditing(contact)
    return self._editing == contact
end

-- Save: insert new or update existing
function ContactApp:save()
    if not self.current then return end

    if self._editing then
        self._editing:copyFrom(self.current)
    else
        table.insert(self._allContacts, self.current)
        self._editing = self.current
    end

    mcp.pushState({
        app = "contacts",
        event = "contact_saved",
        name = self.current.name,
        email = self.current.email
    })
end

-- Cancel editing (discard changes)
function ContactApp:cancel()
    self.current = nil
    self._editing = nil
    self.hideDetail = true
end

-- Delete the contact being edited
function ContactApp:deleteCurrent()
    if self._editing then
        for i, c in ipairs(self._allContacts) do
            if c == self._editing then
                table.remove(self._allContacts, i)
                break
            end
        end
    end
    self.current = nil
    self._editing = nil
    self.hideDetail = true
end

function ContactApp:sendChat()
    if self.chatInput == "" then return end
    table.insert(self.messages, ChatMessage:new("You", self.chatInput))
    mcp.pushState({ app = "contacts", event = "chat", text = self.chatInput })
    self.chatInput = ""
end

function ContactApp:addAgentMessage(text)
    table.insert(self.messages, ChatMessage:new("Agent", text))
end

contacts = contacts or ContactApp:new()
