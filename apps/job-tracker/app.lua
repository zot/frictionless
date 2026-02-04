-- Job Tracker - Track job applications through the hiring pipeline
local json = require('job-tracker.json')

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

local STORAGE_DIR = ".ui/storage/job-tracker/data"
local DATA_FILE = STORAGE_DIR .. "/data.json"

local INACTIVE_STATUSES = {
    archived = true,
    rejected = true,
    withdrawn = true,
}

-- Helper: generate next ID (max existing + 1)
local function nextId(applications)
    local maxId = 0
    for _, app in ipairs(applications) do
        if type(app.id) == "number" and app.id > maxId then
            maxId = app.id
        end
    end
    return maxId + 1
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

-- Commit data.json changes to fossil using shell script
local function commitData(message)
    message = message or "Update data"
    local cmd = string.format('.ui/apps/job-tracker/data-commit.sh "%s" 2>/dev/null', message:gsub('"', '\\"'))
    os.execute(cmd)
end

-- Base64 decoding table (keyed by byte value for efficiency)
local b64chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/'
local b64decode = {}
for i = 1, #b64chars do
    b64decode[string.byte(b64chars, i)] = i - 1
end
b64decode[string.byte('=')] = 0

-- Decode base64 and write directly to file handle (avoids registry overflow)
local function decodeBase64ToFile(data, handle)
    local eq = string.byte('=')
    local i = 1
    handle:setvbuf('full', 4096)
    while i <= #data do
        local b1, b2, b3, b4 = string.byte(data, i, i+3)
        local c1 = b64decode[b1] or 0
        local c2 = b64decode[b2] or 0
        local c3 = b64decode[b3] or 0
        local c4 = b64decode[b4] or 0

        local n = c1 * 262144 + c2 * 4096 + c3 * 64 + c4

        handle:write(string.char(math.floor(n / 65536) % 256))
        if b3 ~= eq then
            handle:write(string.char(math.floor(n / 256) % 256))
        end
        if b4 ~= eq then
            handle:write(string.char(n % 256))
        end
        i = i + 4
    end
end

-- Simple JSON encoder
local encodeJSON = json.encode
local decodeJSON = json.decode

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
    _fileUploadData = "",  -- JS-to-Lua bridge for file uploads
    showAttachmentWarning = false,  -- Show warning dialog when leaving with unsaved attachments
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
    _attachmentsCache = EMPTY,  -- Internal cache for attachments list
    attachmentsChanged = false,  -- Track if attachments have been modified
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

JobTracker.Attachment = session:prototype("JobTracker.Attachment", {
    filename = "",
    path = "",
    applicationId = 0,
})
local Attachment = JobTracker.Attachment

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

-- Reload data from disk (for when user edits data.json externally)
function JobTracker:reload()
    self:loadData()
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
    commitData("Update applications")
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
function JobTracker:sortIcon(column)
    if self.sortColumn ~= column then return "" end
    return self.sortDirection == "asc" and "▲" or "▼"
end

-- Generated sort methods for viewdef binding
for _, col in ipairs({"company", "position", "status", "date"}) do
    local sortName = "sort" .. col:sub(1,1):upper() .. col:sub(2)
    local iconName = col .. "Icon"
    JobTracker[sortName] = function(self) self:toggleSort(col) end
    JobTracker[iconName] = function(self) return self:sortIcon(col) end
end

function JobTracker:allApplications()
    return self._applications
end

function JobTracker:setFilter(f)
    self.filter = f
    self.selected = nil
end

function JobTracker:selectApp(app)
    self.selected = app
    self.selectedStatus = app.status
    self.view = "detail"
end

function JobTracker:showList()
    -- Check if selected app has unsaved attachment changes
    if self.selected and self.selected.attachmentsChanged then
        self.showAttachmentWarning = true
        return
    end
    self.view = "list"
    self.selected = nil
end

-- Hide the attachment warning dialog
function JobTracker:hideAttachmentWarning()
    self.showAttachmentWarning = false
end

function JobTracker:isAttachmentWarningVisible() return self.showAttachmentWarning end
function JobTracker:isAttachmentWarningHidden() return not self.showAttachmentWarning end

function JobTracker:showAddForm()
    self.formMode = "add"
    self.formData = session:create(FormData, { status = "bookmarked" })
    self.view = "form"
end

function JobTracker:showEditForm()
    if not self.selected then return end
    self.formMode = "edit"
    local app = self.selected
    local data = {
        company = app.company,
        position = app.position,
        url = app.url,
        status = app.status,
        location = app.location,
        hqAddress = app.hqAddress,
        salaryMin = tostring(app.salaryMin or ""),
        salaryMax = tostring(app.salaryMax or ""),
        notes = app.notes,
        dateApplied = app.dateApplied or "",
    }
    -- Copy data for change tracking (shallow copy of string/number values)
    local original = {}
    for k, v in pairs(data) do original[k] = v end
    data._original = original
    self.formData = session:create(FormData, data)
    self.view = "form"
end

function JobTracker:saveForm()
    local fd = self.formData
    if self.formMode == "add" then
        local app = session:create(Application, {
            id = nextId(self._applications),
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

function JobTracker:changeStatus()
   if not self.selected then return end
   if self.selectedStatus == self.selected.status then return end
   self:_addStatusChange(self.selected, self.selectedStatus)
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

function JobTracker:chatPanelHidden() return not self.chatPanelOpen end

-- Attachment handling
function JobTracker:uploadFile(filename, content)
    if not self.selected then return end
    local dir = self.selected:attachmentsDir()
    os.execute('mkdir -p "' .. dir .. '"')
    local path = dir .. "/" .. filename
    local handle = io.open(path, "wb")
    if handle then
        handle:write(content)
        handle:close()
        self.selected:clearAttachmentsCache()
        self.selected.attachmentsChanged = true  -- Mark as changed, don't commit yet
    end
end

function JobTracker:deleteAttachment(attachment)
    os.execute('rm -f "' .. attachment.path .. '"')
    if self.selected then
        self.selected:clearAttachmentsCache()
        self.selected.attachmentsChanged = true  -- Mark as changed, don't commit yet
    end
end

-- Save attachment changes (commit to fossil)
function JobTracker:saveAttachments()
    if self.selected and self.selected.attachmentsChanged then
        commitData("Update attachments")
        self.selected.attachmentsChanged = false
    end
    self.showAttachmentWarning = false
end

-- Revert attachment changes (revert from fossil)
function JobTracker:revertAttachments()
   local app = self.selected
    if app then
       app.attachmentsChanged = false
        -- Revert directory using fossil
       os.execute('cd "' .. STORAGE_DIR .. '" && "$HOME/.claude/bin/fossil" revert "jobs/' .. app:idDir() .. '" 2>/dev/null')
        app:clearAttachmentsCache()
    end
    self.showAttachmentWarning = false
end

-- Helper for going back to list after attachment action
local function goBackToList(tracker)
    tracker.view = "list"
    tracker.selected = nil
end

function JobTracker:saveAttachmentsAndBack()
    self:saveAttachments()
    goBackToList(self)
end

function JobTracker:revertAttachmentsAndBack()
    self:revertAttachments()
    goBackToList(self)
end

-- Placeholder for URL attachment (not implemented)
function JobTracker:promptAttachUrl()
    -- TODO: Implement URL attachment dialog
end

-- Process file upload from JS-to-Lua bridge
function JobTracker:processFileUpload()
    if self._fileUploadData == "" or not self.selected then return end

    local upload = self._fileUploadData
    self._fileUploadData = ""  -- Clear immediately

    -- Parse filename:base64content
    local colonPos = upload:find(":")
    if not colonPos then return end

    local filename = upload:sub(1, colonPos - 1)
    local base64 = upload:sub(colonPos + 1)

    -- Decode and write file
    local dir = self.selected:attachmentsDir()
    os.execute('mkdir -p "' .. dir .. '"')
    local path = dir .. "/" .. filename

    local handle = io.open(path, "wb")
    if handle then
        decodeBase64ToFile(base64, handle)
        handle:close()
        self.selected:clearAttachmentsCache()
        self.selected.attachmentsChanged = true  -- Mark as changed, don't commit yet
    end
end

function JobTracker:hasChatMessages() return #self.chatMessages > 0 end
function JobTracker:noChatMessages() return #self.chatMessages == 0 end

-- Generated filter methods for viewdef binding
for _, name in ipairs({"all", "active", "offers", "archived"}) do
    local capName = name:sub(1,1):upper() .. name:sub(2)
    JobTracker["filter" .. capName] = function(self) self:setFilter(name) end
    JobTracker["isFilter" .. capName] = function(self) return self.filter == name end
    JobTracker[name .. "Variant"] = function(self)
        return self.filter == name and "primary" or "default"
    end
end

-- Generated view methods for viewdef binding
for _, name in ipairs({"list", "detail", "form"}) do
    local capName = name:sub(1,1):upper() .. name:sub(2)
    JobTracker["is" .. capName .. "View"] = function(self) return self.view == name end
    JobTracker["not" .. capName .. "View"] = function(self) return self.view ~= name end
end

-- Application methods
function Application:selectMe()
   self._attachmentsCache = nil
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

-- Generated has/no methods for optional fields
for _, field in ipairs({"url", "location", "hqAddress", "notes"}) do
    local capField = field:sub(1,1):upper() .. field:sub(2)
    -- Use short names for hqAddress
    local shortName = field == "hqAddress" and "Hq" or capField
    Application["has" .. shortName] = function(self) return isPresent(self[field]) end
    Application["no" .. shortName] = function(self) return not isPresent(self[field]) end
end

function Application:hasSalary() return self:salaryDisplay() ~= "" end
function Application:noSalary() return not self:hasSalary() end

-- Helper: format ID as 4-digit directory name
function Application:idDir()
    return string.format("%04d", self.id)
end

-- Attachments: directory for this application's files
function Application:attachmentsDir()
   return STORAGE_DIR .. "/jobs/" .. self:idDir()
end

-- List attachment files for this application (cached to avoid repeated ls calls)
function Application:attachments()
    -- Use cached result if available (cache cleared on file operations)
    if self._attachmentsCache then
        return self._attachmentsCache
    end

    local dir = self:attachmentsDir()
    local result = {}
    local handle = io.popen('ls -1 "' .. dir .. '" 2>/dev/null')
    if handle then
        for filename in handle:lines() do
            if filename ~= "" then
                table.insert(result, session:create(Attachment, {
                    filename = filename,
                    path = dir .. "/" .. filename,
                    applicationId = self.id,
                }))
            end
        end
        handle:close()
    end
    self._attachmentsCache = result
    return result
end

-- Clear attachment cache (call after adding/removing files)
function Application:clearAttachmentsCache()
    self._attachmentsCache = nil
end

function Application:hasAttachments() return #self:attachments() > 0 end
function Application:noAttachments() return #self:attachments() == 0 end
function Application:hasAttachmentsChanged() return self.attachmentsChanged == true end
function Application:noAttachmentsChanged() return not self.attachmentsChanged end

function Application:appliedDateDisplay()
    return formatDate(self.dateApplied)
end

-- Attachment methods
function Attachment:deleteMe()
    jobTracker:deleteAttachment(self)
end

local EXTENSION_ICONS = {
    pdf = "file-earmark-pdf",
    doc = "file-earmark-word",
    docx = "file-earmark-word",
    md = "file-earmark-text",
    txt = "file-earmark-text",
    png = "file-earmark-image",
    jpg = "file-earmark-image",
    jpeg = "file-earmark-image",
    gif = "file-earmark-image",
}

function Attachment:icon()
    local ext = self.filename:match("%.([^%.]+)$")
    if ext then ext = ext:lower() end
    return EXTENSION_ICONS[ext] or "file-earmark"
end

function Attachment:downloadUrl()
    -- Return a file:// URL for download
    return "file://" .. self.path
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
function ChatMessage:isUser() return self.role == "user" end
function ChatMessage:isAssistant() return self.role == "assistant" end

function ChatMessage:copyToInput()
    jobTracker.chatInput = self.content
end

-- Instance creation
if not session.reloading then
    jobTracker = JobTracker:new()
end
