-- Contact Manager App
-- Example demonstrating lists, forms, and chat with filtering

-- App prototype (serves as namespace for all related types)
Contacts = session:prototype("Contacts", {
    _allContacts = EMPTY,
    searchQuery = "",
    current = EMPTY,
    _editing = EMPTY,
    hideDetail = true,
    darkMode = false,
    messages = EMPTY,
    chatInput = ""
})

-- Nested prototype: Chat message model
Contacts.ChatMessage = session:prototype("Contacts.ChatMessage", {
    sender = "",
    text = ""
})
local ChatMessage = Contacts.ChatMessage

function ChatMessage:new(sender, text)
    return session:create(ChatMessage, { sender = sender, text = text })
end

-- Nested prototype: Contact model
Contacts.Contact = session:prototype("Contacts.Contact", {
    name = "",
    email = "",
    status = "active",
    vip = false
})
local Contact = Contacts.Contact

function Contact:new(name)
    return session:create(Contact, { name = name or "" })
end

function Contact:clone()
    return Contact:new(self.name, {
        email = self.email,
        status = self.status,
        vip = self.vip
    })
end

function Contact:copyFrom(other)
    self.name = other.name
    self.email = other.email
    self.status = other.status
    self.vip = other.vip
end

-- Select this contact (called from viewdef click)
function Contact:selectMe()
    contacts:select(self)
end

-- Check if this contact is the one being edited
function Contact:isSelected()
    return contacts:isEditing(self)
end

-- Main app methods
function Contacts:new(instance)
    instance = session:create(Contacts, instance)
    instance._allContacts = instance._allContacts or {}
    instance.messages = instance.messages or {}
    return instance
end

-- Computed: filtered contacts based on searchQuery
function Contacts:contacts()
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

function Contacts:contactCount()
    return #self:contacts()
end

-- Add new contact (creates temp, doesn't insert until save)
function Contacts:add()
    self.current = Contact:new("New Contact")
    self._editing = nil  -- nil means adding new
    self.hideDetail = false
end

-- Edit existing contact (clones into temp)
function Contacts:select(contact)
    self.current = contact:clone()
    self._editing = contact  -- remember original
    self.hideDetail = false
end

-- Check if contact is the one being edited
function Contacts:isEditing(contact)
    return self._editing == contact
end

-- Save: insert new or update existing
function Contacts:save()
    if not self.current then return end

    if self._editing then
        -- Editing existing: copy temp back to original
        self._editing:copyFrom(self.current)
    else
        -- Adding new: insert into list
        table.insert(self._allContacts, self.current)
        self._editing = self.current  -- now it's a real contact
    end

    mcp.pushState({
        app = "contacts",
        event = "contact_saved",
        name = self.current.name,
        email = self.current.email
    })
end

-- Cancel editing (discard changes)
function Contacts:cancel()
    self.current = nil
    self._editing = nil
    self.hideDetail = true
end

-- Delete the contact being edited
function Contacts:deleteCurrent()
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

function Contacts:sendChat()
    if self.chatInput == "" then return end
    table.insert(self.messages, ChatMessage:new("You", self.chatInput))
    mcp.pushState({ app = "contacts", event = "chat", text = self.chatInput })
    self.chatInput = ""
end

function Contacts:addAgentMessage(text)
    table.insert(self.messages, ChatMessage:new("Agent", text))
end

-- Initialize (idempotent - only runs on first load)
if not session.reloading then
    contacts = Contacts:new()
end
