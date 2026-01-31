-- Job Tracker - Track job applications through the hiring pipeline

local STATUS_CONFIG = {
    bookmarked = { display = "Bookmarked", variant = "neutral" },
    applied = { display = "Applied", variant = "primary" },
    phone = { display = "Phone Screen", variant = "primary" },
    technical = { display = "Technical", variant = "primary" },
    onsite = { display = "Onsite", variant = "primary" },
    offer = { display = "Offer", variant = "success" },
    rejected = { display = "Rejected", variant = "danger" },
    withdrawn = { display = "Withdrawn", variant = "neutral" },
    archived = { display = "Archived", variant = "neutral" },
}

local DATA_FILE = ".ui/apps/job-tracker/data.json"

local INACTIVE_STATUSES = {
    archived = true,
    rejected = true,
    withdrawn = true,
}

-- Helper: generate UUID
local function uuid()
    local template = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'
    return string.gsub(template, '[xy]', function(c)
        local v = (c == 'x') and math.random(0, 0xf) or math.random(8, 0xb)
        return string.format('%x', v)
    end)
end

-- Helper: today's date
local function today()
    return os.date("%Y-%m-%d")
end

-- Helper: format date for display
local function formatDate(isoDate)
    if not isoDate or isoDate == "" then return "" end
    local y, m, d = isoDate:match("(%d+)-(%d+)-(%d+)")
    if not y then return isoDate end
    local months = {"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
    return months[tonumber(m)] .. " " .. tonumber(d)
end

-- Helper: check if string is present (non-nil and non-empty)
local function isPresent(str)
    return str and str ~= ""
end

-- Helper: format salary value to display string (e.g., 150000 -> "$150k")
local function formatSalary(value)
    if not value or value == 0 then return nil end
    return "$" .. math.floor(value / 1000) .. "k"
end

-- Helper: read file contents
local function readFile(path)
    local handle = io.open(path, "r")
    if not handle then return nil end
    local content = handle:read("*a")
    handle:close()
    return content
end

-- Helper: write file contents
local function writeFile(path, content)
    local handle = io.open(path, "w")
    if not handle then return false end
    handle:write(content)
    handle:close()
    return true
end

-- Simple JSON encoder
local function encodeJSON(val, indent, level)
    indent = indent or "  "
    level = level or 0
    local prefix = string.rep(indent, level)
    local nextPrefix = string.rep(indent, level + 1)

    if val == nil then
        return "null"
    elseif type(val) == "boolean" then
        return val and "true" or "false"
    elseif type(val) == "number" then
        return tostring(val)
    elseif type(val) == "string" then
        return string.format("%q", val)
    elseif type(val) == "table" then
        -- Check if array
        local isArray = #val > 0 or next(val) == nil
        if isArray then
            local items = {}
            for i, v in ipairs(val) do
                items[i] = nextPrefix .. encodeJSON(v, indent, level + 1)
            end
            if #items == 0 then return "[]" end
            return "[\n" .. table.concat(items, ",\n") .. "\n" .. prefix .. "]"
        else
            local items = {}
            for k, v in pairs(val) do
                if type(k) == "string" and not k:match("^_") then
                    table.insert(items, nextPrefix .. string.format("%q", k) .. ": " .. encodeJSON(v, indent, level + 1))
                end
            end
            if #items == 0 then return "{}" end
            return "{\n" .. table.concat(items, ",\n") .. "\n" .. prefix .. "}"
        end
    end
    return "null"
end

-- Simple JSON decoder
local function decodeJSON(str)
    if not str then return nil end
    -- Transform JSON to Lua table syntax
    local luaStr = str
        :gsub("%[", "{")
        :gsub("%]", "}")
        :gsub(":null", "=nil")
        :gsub(":(%s*)", "=%1")
        :gsub('"([^"]-)"=', "[%1]=")
    -- Fix the key quoting
    luaStr = luaStr:gsub('%[([^%]]+)%]=', function(key)
        return '["' .. key .. '"]='
    end)
    local chunk = "return " .. luaStr
    local f = loadstring(chunk)
    if not f then return nil end
    local ok, result = pcall(f)
    return ok and result or nil
end

-- Prototypes
JobTracker = session:prototype("JobTracker", {
    _applications = EMPTY,
    selected = EMPTY,
    filter = "all",
    view = "list",
    formMode = "add",
    formData = EMPTY,
    noteInput = "",
    urlInput = "",
    chatInput = "",
    chatMessages = EMPTY,
    chatPanelOpen = false,
    sortColumn = "date",  -- "company", "position", "status", "date"
    sortDirection = "desc",  -- "asc" or "desc"
})

JobTracker.Application = session:prototype("JobTracker.Application", {
    id = "",
    company = "",
    position = "",
    url = "",
    status = "bookmarked",
    dateAdded = "",
    dateApplied = "",
    location = "",
    hqAddress = "",
    salaryMin = 0,
    salaryMax = 0,
    notes = "",
    timeline = EMPTY,
})
local Application = JobTracker.Application

JobTracker.TimelineEvent = session:prototype("JobTracker.TimelineEvent", {
    date = "",
    event = "",
    note = "",
    fromStatus = "",
    toStatus = "",
})
local TimelineEvent = JobTracker.TimelineEvent

JobTracker.ChatMessage = session:prototype("JobTracker.ChatMessage", {
    role = "",
    content = "",
})
local ChatMessage = JobTracker.ChatMessage

JobTracker.FormData = session:prototype("JobTracker.FormData", {
    company = "",
    position = "",
    url = "",
    status = "bookmarked",
    location = "",
    hqAddress = "",
    salaryMin = "",
    salaryMax = "",
    notes = "",
    dateApplied = "",
    -- Original values for change tracking
    _original = EMPTY,
})
local FormData = JobTracker.FormData

-- Check if form has unsaved changes
function FormData:hasChanges()
    if not self._original then return true end  -- New form always has changes
    return self.company ~= self._original.company
        or self.position ~= self._original.position
        or self.url ~= self._original.url
        or self.status ~= self._original.status
        or self.location ~= self._original.location
        or self.hqAddress ~= self._original.hqAddress
        or self.salaryMin ~= self._original.salaryMin
        or self.salaryMax ~= self._original.salaryMax
        or self.notes ~= self._original.notes
        or self.dateApplied ~= self._original.dateApplied
end

function FormData:noChanges()
    return not self:hasChanges()
end

-- Fields to serialize for applications (excluding timeline which needs special handling)
local APP_FIELDS = {
    "id", "company", "position", "url", "status",
    "dateAdded", "dateApplied", "location", "hqAddress",
    "salaryMin", "salaryMax", "notes"
}

local TIMELINE_FIELDS = {"date", "event", "note", "fromStatus", "toStatus"}

-- JobTracker methods
function JobTracker:new(instance)
    instance = session:create(JobTracker, instance)
    instance._applications = instance._applications or {}
    instance.chatMessages = instance.chatMessages or {}
    instance.formData = session:create(FormData, {})
    instance:loadData()
    return instance
end

-- Hot-reload mutation: initialize new fields on existing instances
function JobTracker:mutate()
    if self.chatMessages == nil then
        self.chatMessages = {}
    end
end

function JobTracker:loadData()
    local content = readFile(DATA_FILE)
    if not content then return end

    local data = decodeJSON(content)
    if not data or not data.applications then return end

    self._applications = {}
    for _, appData in ipairs(data.applications) do
        local app = session:create(Application, appData)
        app.timeline = app.timeline or {}
        for i, evt in ipairs(app.timeline) do
            app.timeline[i] = session:create(TimelineEvent, evt)
        end
        table.insert(self._applications, app)
    end
end

function JobTracker:saveData()
    local data = { applications = {} }
    for _, app in ipairs(self._applications) do
        local appData = {}
        for _, field in ipairs(APP_FIELDS) do
            appData[field] = app[field]
        end
        appData.timeline = {}
        for _, evt in ipairs(app.timeline or {}) do
            local evtData = {}
            for _, field in ipairs(TIMELINE_FIELDS) do
                evtData[field] = evt[field]
            end
            table.insert(appData.timeline, evtData)
        end
        table.insert(data.applications, appData)
    end
    writeFile(DATA_FILE, encodeJSON(data))
end

function JobTracker:applications()
    local result = {}
    for _, app in ipairs(self._applications) do
        local include = self.filter == "all"
            or (self.filter == "active" and not INACTIVE_STATUSES[app.status])
            or (self.filter == "offers" and app.status == "offer")
            or (self.filter == "archived" and app.status == "archived")
        if include then
            table.insert(result, app)
        end
    end

    -- Sort the result
    local col = self.sortColumn
    local asc = self.sortDirection == "asc"
    table.sort(result, function(a, b)
        local va, vb
        if col == "company" then
            va, vb = (a.company or ""):lower(), (b.company or ""):lower()
        elseif col == "position" then
            va, vb = (a.position or ""):lower(), (b.position or ""):lower()
        elseif col == "status" then
            va, vb = a.status or "", b.status or ""
        else  -- date
            va, vb = a.dateApplied or a.dateAdded or "", b.dateApplied or b.dateAdded or ""
        end
        if asc then
            return va < vb
        else
            return va > vb
        end
    end)

    return result
end

-- Toggle sort column/direction
function JobTracker:toggleSort(column)
    if self.sortColumn == column then
        self.sortDirection = self.sortDirection == "asc" and "desc" or "asc"
    else
        self.sortColumn = column
        self.sortDirection = "asc"
    end
end

-- Sort indicator methods for viewdef binding
function JobTracker:sortCompany() self:toggleSort("company") end
function JobTracker:sortPosition() self:toggleSort("position") end
function JobTracker:sortStatus() self:toggleSort("status") end
function JobTracker:sortDate() self:toggleSort("date") end

function JobTracker:sortIcon(column)
    if self.sortColumn ~= column then return "" end
    return self.sortDirection == "asc" and "▲" or "▼"
end

function JobTracker:companyIcon() return self:sortIcon("company") end
function JobTracker:positionIcon() return self:sortIcon("position") end
function JobTracker:statusIcon() return self:sortIcon("status") end
function JobTracker:dateIcon() return self:sortIcon("date") end

function JobTracker:allApplications()
    return self._applications
end

function JobTracker:setFilter(f)
    self.filter = f
    self.selected = nil
end

function JobTracker:selectApp(app)
    self.selected = app
    self.view = "detail"
end

function JobTracker:showList()
    self.view = "list"
    self.selected = nil
end

function JobTracker:showAddForm()
    self.formMode = "add"
    self.formData = session:create(FormData, { status = "bookmarked" })
    self.view = "form"
end

function JobTracker:showEditForm()
    if not self.selected then return end
    self.formMode = "edit"
    local data = {
        company = self.selected.company,
        position = self.selected.position,
        url = self.selected.url,
        status = self.selected.status,
        location = self.selected.location,
        hqAddress = self.selected.hqAddress,
        salaryMin = tostring(self.selected.salaryMin or ""),
        salaryMax = tostring(self.selected.salaryMax or ""),
        notes = self.selected.notes,
        dateApplied = self.selected.dateApplied or "",
    }
    -- Store original values for change tracking
    data._original = {
        company = data.company,
        position = data.position,
        url = data.url,
        status = data.status,
        location = data.location,
        hqAddress = data.hqAddress,
        salaryMin = data.salaryMin,
        salaryMax = data.salaryMax,
        notes = data.notes,
        dateApplied = data.dateApplied,
    }
    self.formData = session:create(FormData, data)
    self.view = "form"
end

function JobTracker:saveForm()
    local fd = self.formData
    if self.formMode == "add" then
        local app = session:create(Application, {
            id = uuid(),
            company = fd.company,
            position = fd.position,
            url = fd.url,
            status = fd.status,
            dateAdded = today(),
            dateApplied = fd.status ~= "bookmarked" and today() or "",
            location = fd.location,
            hqAddress = fd.hqAddress,
            salaryMin = tonumber(fd.salaryMin) or 0,
            salaryMax = tonumber(fd.salaryMax) or 0,
            notes = fd.notes,
            timeline = {},
        })
        local evt = session:create(TimelineEvent, {
            date = today(),
            event = "added",
            note = "Added application",
        })
        app.timeline = { evt }
        table.insert(self._applications, 1, app)
        self.selected = app
    else
        local app = self.selected
        app.company = fd.company
        app.position = fd.position
        app.url = fd.url
        app.location = fd.location
        app.hqAddress = fd.hqAddress
        app.salaryMin = tonumber(fd.salaryMin) or 0
        app.salaryMax = tonumber(fd.salaryMax) or 0
        app.notes = fd.notes
        app.dateApplied = fd.dateApplied
        if app.status ~= fd.status then
            self:_addStatusChange(app, fd.status)
        end
    end
    self:saveData()
    self.view = "detail"
end

function JobTracker:cancelForm()
    self.view = self.formMode == "add" and "list" or "detail"
end

function JobTracker:_addStatusChange(app, newStatus)
    local oldStatus = app.status
    app.status = newStatus
    if newStatus ~= "bookmarked" and not isPresent(app.dateApplied) then
        app.dateApplied = today()
    end
    local evt = session:create(TimelineEvent, {
        date = today(),
        event = "status_change",
        note = "Status changed",
        fromStatus = oldStatus,
        toStatus = newStatus,
    })
    table.insert(app.timeline, 1, evt)
end

function JobTracker:changeStatus(status)
    if not self.selected then return end
    self:_addStatusChange(self.selected, status)
    self:saveData()
end

function JobTracker:addNote()
    if not self.selected or self.noteInput == "" then return end
    local evt = session:create(TimelineEvent, {
        date = today(),
        event = "note",
        note = self.noteInput,
    })
    table.insert(self.selected.timeline, 1, evt)
    self.noteInput = ""
    self:saveData()
end

function JobTracker:deleteApp()
    if not self.selected then return end
    for i, app in ipairs(self._applications) do
        if app.id == self.selected.id then
            table.remove(self._applications, i)
            break
        end
    end
    self.selected = nil
    self.view = "list"
    self:saveData()
end

function JobTracker:prefillFromScrape(data)
    self.formMode = "add"
    self.formData = session:create(FormData, {
        company = data.company or "",
        position = data.position or "",
        url = data.url or "",
        status = "bookmarked",
        location = data.location or "",
        hqAddress = data.hqAddress or "",
        salaryMin = data.salaryMin and tostring(data.salaryMin) or "",
        salaryMax = data.salaryMax and tostring(data.salaryMax) or "",
        notes = "",
    })
    self.view = "form"
end

-- Send URL to Claude for scraping
function JobTracker:submitUrl()
    if self.urlInput == "" then return end
    local url = self.urlInput
    self.urlInput = ""
    mcp.pushState({
        app = "job-tracker",
        event = "chat",
        text = url,
    })
end

-- Chat methods
function JobTracker:submitChat()
    if self.chatInput == "" then return end
    local text = self.chatInput
    self.chatInput = ""
    self:addChatMessage("user", text)
    mcp.pushState({
        app = "job-tracker",
        event = "chat",
        text = text,
    })
end

function JobTracker:toggleChatPanel()
    self.chatPanelOpen = not self.chatPanelOpen
end

function JobTracker:addChatMessage(role, content)
    local msg = session:create(ChatMessage, {
        role = role,
        content = content,
    })
    table.insert(self.chatMessages, msg)
end

function JobTracker:clearChat()
    self.chatMessages = {}
end

function JobTracker:chatPanelHidden()
    return not self.chatPanelOpen
end

function JobTracker:hasChatMessages()
    return #self.chatMessages > 0
end

function JobTracker:noChatMessages()
    return not self:hasChatMessages()
end

-- Filter methods
function JobTracker:filterAll() self:setFilter("all") end
function JobTracker:filterActive() self:setFilter("active") end
function JobTracker:filterOffers() self:setFilter("offers") end
function JobTracker:filterArchived() self:setFilter("archived") end

function JobTracker:isFilterAll() return self.filter == "all" end
function JobTracker:isFilterActive() return self.filter == "active" end
function JobTracker:isFilterOffers() return self.filter == "offers" end
function JobTracker:isFilterArchived() return self.filter == "archived" end

-- View methods
function JobTracker:isListView() return self.view == "list" end
function JobTracker:isDetailView() return self.view == "detail" end
function JobTracker:isFormView() return self.view == "form" end
function JobTracker:notListView() return not self:isListView() end
function JobTracker:notDetailView() return not self:isDetailView() end
function JobTracker:notFormView() return not self:isFormView() end

-- Filter button variants (helper to reduce repetition)
local function filterVariant(tracker, filterName)
    return tracker.filter == filterName and "primary" or "default"
end

function JobTracker:allVariant() return filterVariant(self, "all") end
function JobTracker:activeVariant() return filterVariant(self, "active") end
function JobTracker:offersVariant() return filterVariant(self, "offers") end
function JobTracker:archivedVariant() return filterVariant(self, "archived") end

-- Application methods
function Application:selectMe()
    jobTracker:selectApp(self)
end

function Application:isSelected()
    return self == jobTracker.selected
end

function Application:statusConfig()
    return STATUS_CONFIG[self.status] or {}
end

function Application:statusDisplay()
    return self:statusConfig().display or self.status
end

function Application:statusVariant()
    return self:statusConfig().variant or "neutral"
end

function Application:dateDisplay()
    local d = isPresent(self.dateApplied) and self.dateApplied or self.dateAdded
    return formatDate(d)
end

function Application:salaryDisplay()
    local min = formatSalary(self.salaryMin)
    local max = formatSalary(self.salaryMax)
    if min and max then
        return min .. "-" .. max
    end
    return min or max or ""
end

function Application:hasUrl() return isPresent(self.url) end
function Application:noUrl() return not self:hasUrl() end
function Application:hasLocation() return isPresent(self.location) end
function Application:noLocation() return not self:hasLocation() end
function Application:hasSalary() return self:salaryDisplay() ~= "" end
function Application:noSalary() return not self:hasSalary() end
function Application:hasHq() return isPresent(self.hqAddress) end
function Application:noHq() return not self:hasHq() end

function Application:appliedDateDisplay()
    return formatDate(self.dateApplied)
end

-- TimelineEvent methods
function TimelineEvent:dateDisplay()
    return formatDate(self.date)
end

local function getStatusDisplay(status)
    local cfg = STATUS_CONFIG[status]
    return cfg and cfg.display or status
end

function TimelineEvent:description()
    if self.event == "status_change" then
        return "Status: " .. getStatusDisplay(self.fromStatus) .. " -> " .. getStatusDisplay(self.toStatus)
    elseif self.event == "added" then
        return "Added application"
    end
    return self.note or ""
end

-- ChatMessage methods
function ChatMessage:isUser()
    return self.role == "user"
end

function ChatMessage:isAssistant()
    return self.role == "assistant"
end

function ChatMessage:copyToInput()
    jobTracker.chatInput = self.content
end

-- Instance creation
if not session.reloading then
    jobTracker = JobTracker:new()
end
