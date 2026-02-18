-- Example app - Todo list with bookmarklet page capture
-- Design: design.md

local json = require("mcp.json")

local Example = session:prototype("Example", {
    _todos = EMPTY,
    _draft = EMPTY,
    inputText = "",
    showBookmarklet = false
})

Example.TodoItem = session:prototype("Example.TodoItem", {
    text = "",
    done = false,
    url = "",
    _parent = EMPTY
})
local TodoItem = Example.TodoItem

local function createTodoItem(parent, text, url)
    return session:create(TodoItem, {
        text = text,
        url = url or "",
        _parent = parent
    })
end

function TodoItem:toggle()
    self.done = not self.done
    if self._parent then self._parent:save() end
end

function TodoItem:remove()
    if self._parent then self._parent:removeTodo(self) end
end

function TodoItem:label()
    if self.url ~= "" then
        local domain = self.url:match("https?://([^/]+)") or self.url
        return self.text .. " — " .. domain
    end
    return self.text
end

function Example:todos() return self._todos end
function Example:draft() return self._draft end
function Example:noDraft() return self._draft == nil end

function Example:addTodo()
    if self.inputText == "" then return end
    table.insert(self._todos, createTodoItem(self, self.inputText))
    self.inputText = ""
    self:save()
end

function Example:addDraftTodo(text, url)
    self._draft = createTodoItem(self, text, url)
end

function Example:saveDraft()
    if not self._draft then return end
    table.insert(self._todos, self._draft)
    self._draft = nil
    self:save()
end

function Example:cancelDraft()
    self._draft = nil
end

function Example:removeTodo(todo)
    for i, t in ipairs(self._todos) do
        if t == todo then
            table.remove(self._todos, i)
            self:save()
            return
        end
    end
end

function Example:toggleBookmarklet() self.showBookmarklet = not self.showBookmarklet end
function Example:isBookmarkletHidden() return not self.showBookmarklet end

-- Persistence

function Example:storagePath()
    local status = mcp:status()
    if not status or not status.base_dir then return nil end
    return status.base_dir .. "/storage/example/todos.json"
end

function Example:save()
    local path = self:storagePath()
    if not path then return end
    local dir = path:match("(.+)/[^/]+$")
    if dir then os.execute('mkdir -p "' .. dir .. '"') end
    local data = {}
    for _, item in ipairs(self._todos) do
        table.insert(data, {
            text = item.text,
            done = item.done,
            url = item.url
        })
    end
    local f = io.open(path, "w")
    if not f then return end
    f:write(json.encode(data))
    f:close()
end

function Example:load()
    local path = self:storagePath()
    if not path then return end
    local f = io.open(path, "r")
    if not f then return end
    local content = f:read("*a")
    f:close()
    local ok, data = pcall(json.decode, content)
    if not ok or type(data) ~= "table" then return end
    self._todos = {}
    for _, entry in ipairs(data) do
        local item = createTodoItem(self, entry.text or "", entry.url or "")
        item.done = entry.done or false
        table.insert(self._todos, item)
    end
end

-- Initialize
if not session.reloading then
    example = session:create(Example, {
        _todos = {},
        _draft = nil
    })
    example:load()
end
