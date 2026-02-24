-- Job Tracker - Track job applications through the hiring pipeline
local json = require('mcp.json')

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
local RESUMES_DIR = STORAGE_DIR .. "/resumes"
local MASTER_RESUME_FILE = STORAGE_DIR .. "/master-resume.md"

local INACTIVE_STATUSES = {
    archived = true,
    rejected = true,
    withdrawn = true,
}

-- Helper: generate next ID (max existing + 1)
local function nextId(items)
    local maxId = 0
    for _, item in ipairs(items) do
        if type(item.id) == "number" and item.id > maxId then
            maxId = item.id
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

-- Ensure HTML serving symlink exists for markdown preview
local function ensureStorageSymlink()
    local target = "../storage/job-tracker/data"
    local link = ".ui/html/job-tracker-storage"
    -- Check if symlink already exists and points to correct target
    local handle = io.popen('readlink "' .. link .. '" 2>/dev/null')
    if handle then
        local current = handle:read("*l")
        handle:close()
        if current == target then return end  -- Already correct
    end
    -- Create or fix symlink
    os.execute('ln -sfn "' .. target .. '" "' .. link .. '"')
end

-- Ensure resumes directory exists
local function ensureResumesDir()
    os.execute('mkdir -p "' .. RESUMES_DIR .. '"')
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
    showBookmarklet = false,  -- Collapsible bookmarklet install section
    sortColumn = "date",  -- "company", "position", "status", "date"
    sortDirection = "desc",  -- "asc" or "desc"
    _fileUploadData = "",  -- JS-to-Lua bridge for file uploads
    showAttachmentWarning = false,  -- Show warning dialog when leaving with unsaved attachments
    -- Resume management
    _resumes = EMPTY,
    selectedResume = EMPTY,
    showMasterResume = false,
    showDeleteResumeDialog = false,
    showLinkPicker = false,
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
    website = "",
    salaryMin = 0,
    salaryMax = 0,
    notes = "",
    timeline = EMPTY,
    _attachmentsCache = EMPTY,  -- Internal cache for attachments list
    attachmentsChanged = false,  -- Track if attachments have been modified
    resumeId = EMPTY,  -- Linked resume ID
    selectedResumeId = "",  -- Dropdown value for resume selection
})
local Application = JobTracker.Application

JobTracker.Resume = session:prototype("JobTracker.Resume", {
    id = 0,
    name = "",
    filename = "",
    applicationIds = EMPTY,
    dateCreated = "",
    dateModified = "",
})
local Resume = JobTracker.Resume

JobTracker.ResumeBadge = session:prototype("JobTracker.ResumeBadge", {
    app = EMPTY,
    resume = EMPTY,
})
local ResumeBadge = JobTracker.ResumeBadge

JobTracker.TimelineEvent = session:prototype("JobTracker.TimelineEvent", {
    date = "",
    event = "",
    note = "",
    fromStatus = "",
    toStatus = "",
})
local TimelineEvent = JobTracker.TimelineEvent

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
    website = "",
    salaryMin = "",
    salaryMax = "",
    notes = "",
    dateApplied = "",
    -- Original values for change tracking
    _original = EMPTY,
})
local FormData = JobTracker.FormData

local FORM_FIELDS = {"company", "position", "url", "status", "location", "hqAddress", "website", "salaryMin", "salaryMax", "notes", "dateApplied"}

-- Check if form has unsaved changes
function FormData:hasChanges()
    if not self._original then return true end  -- New form always has changes
    for _, field in ipairs(FORM_FIELDS) do
        if self[field] ~= self._original[field] then return true end
    end
    return false
end

function FormData:noChanges()
    return not self:hasChanges()
end

-- Fields to serialize for applications (excluding timeline which needs special handling)
local APP_FIELDS = {
    "id", "company", "position", "url", "status",
    "dateAdded", "dateApplied", "location", "hqAddress", "website",
    "salaryMin", "salaryMax", "notes", "resumeId"
}

local TIMELINE_FIELDS = {"date", "event", "note", "fromStatus", "toStatus"}

local RESUME_FIELDS = {"id", "name", "filename", "applicationIds", "dateCreated", "dateModified"}

-- JobTracker methods
function JobTracker:new(instance)
    instance = session:create(JobTracker, instance)
    instance._applications = instance._applications or {}
    instance._resumes = instance._resumes or {}
    instance.formData = session:create(FormData, {})
    -- Ensure storage setup
    ensureStorageSymlink()
    ensureResumesDir()
    instance:loadData()
    return instance
end

-- Hot-reload mutation: initialize new fields on existing instances
function JobTracker:mutate()
    if self._resumes == nil then
        self._resumes = {}
    end
end

function JobTracker:loadData()
    local content = readFile(DATA_FILE)
    if not content then return end

    local data = decodeJSON(content)
    if not data then return end

    -- Load applications
    self._applications = {}
    if data.applications then
        for _, appData in ipairs(data.applications) do
            local app = session:create(Application, appData)
            app.timeline = app.timeline or {}
            for i, evt in ipairs(app.timeline) do
                app.timeline[i] = session:create(TimelineEvent, evt)
            end
            -- Initialize selectedResumeId for dropdown binding
            app.selectedResumeId = app.resumeId and tostring(app.resumeId) or ""
            table.insert(self._applications, app)
        end
    end

    -- Load resumes
    self._resumes = {}
    if data.resumes then
        for _, resumeData in ipairs(data.resumes) do
            local resume = session:create(Resume, resumeData)
            resume.applicationIds = resume.applicationIds or {}
            table.insert(self._resumes, resume)
        end
    end

    -- Repair bidirectional links: ensure resume.applicationIds matches app.resumeId
    self:repairResumeLinks()
end

-- Repair bidirectional links between apps and resumes
function JobTracker:repairResumeLinks()
    local dirty = false
    for _, app in ipairs(self._applications) do
        if app.resumeId then
            local resume = self:findResumeById(app.resumeId)
            if resume then
                -- Check if app is already in resume's list
                local found = false
                for _, id in ipairs(resume.applicationIds or {}) do
                    if id == app.id then found = true; break end
                end
                if not found then
                    resume.applicationIds = resume.applicationIds or {}
                    table.insert(resume.applicationIds, app.id)
                    dirty = true
                end
            end
        end
    end
    if dirty then self:saveData() end
end

-- Reload data from disk (for when user edits data.json externally)
function JobTracker:reload()
    self:loadData()
    self:syncResumesWithDisk()
end

-- Sync _resumes with actual files in RESUMES_DIR
function JobTracker:syncResumesWithDisk()
    -- Get list of .md files in resumes directory
    local filesOnDisk = {}
    local handle = io.popen('ls -1 "' .. RESUMES_DIR .. '"/*.md 2>/dev/null')
    if handle then
        for line in handle:lines() do
            local filename = line:match("([^/]+)$")
            if filename then
                filesOnDisk[filename] = true
            end
        end
        handle:close()
    end

    -- Build map of existing resume filenames
    local existingFiles = {}
    for _, resume in ipairs(self._resumes) do
        existingFiles[resume.filename] = resume
    end

    -- Add new files not in _resumes
    local changed = false
    for filename, _ in pairs(filesOnDisk) do
        if not existingFiles[filename] then
            -- Create new resume entry
            local name = filename:gsub("%.md$", ""):gsub("-", " "):gsub("(%a)([%w_']*)", function(a, b) return a:upper() .. b end)
            local resume = session:create(Resume, {
                id = nextId(self._resumes),
                name = name,
                filename = filename,
                applicationIds = {},
                dateCreated = os.date("%Y-%m-%d"),
                dateModified = os.date("%Y-%m-%d"),
            })
            table.insert(self._resumes, resume)
            changed = true
        end
    end

    -- Remove entries for deleted files
    local i = 1
    while i <= #self._resumes do
        local resume = self._resumes[i]
        if not filesOnDisk[resume.filename] then
            -- Unlink from all applications
            for _, app in ipairs(self._applications) do
                if app.resumeId == resume.id then
                    app.resumeId = nil
                    app.selectedResumeId = ""
                end
            end
            table.remove(self._resumes, i)
            changed = true
        else
            i = i + 1
        end
    end

    if changed then
        self:saveData()
    end
end


function JobTracker:saveData()
    local data = { applications = {}, resumes = {} }

    -- Save applications
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

    -- Save resumes
    for _, resume in ipairs(self._resumes) do
        local resumeData = {}
        for _, field in ipairs(RESUME_FIELDS) do
            resumeData[field] = resume[field]
        end
        table.insert(data.resumes, resumeData)
    end

    writeFile(DATA_FILE, encodeJSON(data))
    commitData("Update data")
end

-- Helper: check if app matches filter
local function matchesFilter(app, filter)
    if filter == "all" then return true end
    if filter == "active" then return not INACTIVE_STATUSES[app.status] end
    if filter == "offers" then return app.status == "offer" end
    if filter == "archived" then return app.status == "archived" end
    return false
end

-- Helper: get sort value for an application
local function getSortValue(app, col)
    if col == "company" then return (app.company or ""):lower() end
    if col == "position" then return (app.position or ""):lower() end
    if col == "status" then return app.status or "" end
    return app.dateApplied or app.dateAdded or ""  -- date (default)
end

function JobTracker:applications()
    local result = {}
    for _, app in ipairs(self._applications) do
        if matchesFilter(app, self.filter) then
            table.insert(result, app)
        end
    end

    local col, asc = self.sortColumn, self.sortDirection == "asc"
    table.sort(result, function(a, b)
        local va, vb = getSortValue(a, col), getSortValue(b, col)
        return asc and va < vb or va > vb
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
    -- Sync dropdown value with actual resumeId BEFORE setting selected
    -- This ensures the binding sees the correct value when selected changes
    app.selectedResumeId = app.resumeId and tostring(app.resumeId) or ""

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
function JobTracker:isDeleteResumeDialogVisible() return self.showDeleteResumeDialog end
function JobTracker:isDeleteResumeDialogHidden() return not self.showDeleteResumeDialog end
function JobTracker:isLinkPickerVisible() return self.showLinkPicker end
function JobTracker:isLinkPickerHidden() return not self.showLinkPicker end

function JobTracker:toggleBookmarklet() self.showBookmarklet = not self.showBookmarklet end
function JobTracker:isBookmarkletHidden() return not self.showBookmarklet end

function JobTracker:showAddForm()
    self.formMode = "add"
    self.formData = session:create(FormData, { status = "bookmarked" })
    self.view = "form"
end

function JobTracker:showEditForm()
    if not self.selected then return end
    self.formMode = "edit"
    local app = self.selected
    -- Build form data from application, converting numbers to strings for salary fields
    local data = {}
    for _, field in ipairs(FORM_FIELDS) do
        local value = app[field]
        if field == "salaryMin" or field == "salaryMax" then
            data[field] = tostring(value or "")
        else
            data[field] = value or ""
        end
    end
    -- Copy data for change tracking
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
            website = fd.website,
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
        -- Save pending page content as attachment if present
        if self._pendingPageFile and self._pendingPageFile ~= "" then
            local dir = app:attachmentsDir()
            os.execute('mkdir -p "' .. dir .. '"')
            local fname = self._pendingPageFilename or "job-listing.md"
            os.execute('mv "' .. self._pendingPageFile .. '" "' .. dir .. '/' .. fname .. '"')
            self._pendingPageFile = nil
            self._pendingPageFilename = nil
            app:clearAttachmentsCache()
            commitData("Attach job listing page")
        end
    else
        local app = self.selected
        app.company = fd.company
        app.position = fd.position
        app.url = fd.url
        app.location = fd.location
        app.hqAddress = fd.hqAddress
        app.website = fd.website
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
        website = data.website or "",
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

-- Resume view methods
function JobTracker:showResumeView()
    self.view = "resume"
    self.showMasterResume = false
    self.selectedResume = nil
end

function JobTracker:isResumeView() return self.view == "resume" end
function JobTracker:notResumeView() return self.view ~= "resume" end

function JobTracker:resumes() return self._resumes end

function JobTracker:findResumeById(id)
    for _, resume in ipairs(self._resumes) do
        if resume.id == id then return resume end
    end
    return nil
end

function JobTracker:selectResume(resume)
    self.selectedResume = resume
    self.showMasterResume = false
    self.showLinkPicker = false
end

function JobTracker:showMaster()
    self.selectedResume = nil
    self.showMasterResume = true
    self.showLinkPicker = false
end

function JobTracker:isShowingMaster() return self.showMasterResume end
function JobTracker:notShowingMaster() return not self.showMasterResume end
function JobTracker:hasSelectedResume() return self.selectedResume ~= nil end
function JobTracker:noSelectedResume() return self.selectedResume == nil end


-- Helper: slugify name for filename
local function slugify(name)
    return name:lower():gsub("[^%w]+", "-"):gsub("^-+", ""):gsub("-+$", "")
end

function JobTracker:createResume()
    -- Create new resume from master template
    local name = "New Resume " .. os.date("%Y-%m-%d")
    local filename = slugify(name) .. ".md"
    local id = nextId(self._resumes)

    -- Copy master resume content
    local masterContent = readFile(MASTER_RESUME_FILE) or "# Resume\n\nEdit this resume."
    os.execute('mkdir -p "' .. RESUMES_DIR .. '"')
    writeFile(RESUMES_DIR .. "/" .. filename, masterContent)

    local resume = session:create(Resume, {
        id = id,
        name = name,
        filename = filename,
        applicationIds = {},
        dateCreated = today(),
        dateModified = today(),
    })
    table.insert(self._resumes, 1, resume)
    self:saveData()
    self:selectResume(resume)
end

function JobTracker:deleteSelectedResume()
    if not self.selectedResume then return end
    self.showDeleteResumeDialog = true
end

function JobTracker:confirmDeleteResume()
    if not self.selectedResume then return end
    local resume = self.selectedResume

    -- Remove file
    os.execute('rm -f "' .. RESUMES_DIR .. '/' .. resume.filename .. '"')

    -- Remove from list
    for i, r in ipairs(self._resumes) do
        if r.id == resume.id then
            table.remove(self._resumes, i)
            break
        end
    end

    -- Clear links from applications
    for _, app in ipairs(self._applications) do
        if app.resumeId == resume.id then
            app.resumeId = nil
            app.selectedResumeId = ""
        end
    end

    self.selectedResume = nil
    self.showDeleteResumeDialog = false
    self:saveData()
end

function JobTracker:cancelDeleteResume()
    self.showDeleteResumeDialog = false
end

-- Link picker
function JobTracker:toggleLinkPicker()
    self.showLinkPicker = not self.showLinkPicker
end

function JobTracker:unlinkableApps()
    if not self.selectedResume then return {} end
    local linked = {}
    for _, id in ipairs(self.selectedResume.applicationIds or {}) do
        linked[id] = true
    end
    local result = {}
    for _, app in ipairs(self._applications) do
        if not linked[app.id] then
            table.insert(result, app)
        end
    end
    return result
end

function JobTracker:linkAppToResume(app)
    if not self.selectedResume or not app then return end
    local resume = self.selectedResume
    resume.applicationIds = resume.applicationIds or {}
    table.insert(resume.applicationIds, app.id)
    resume.dateModified = today()
    self.showLinkPicker = false
    self:saveData()
end

-- Helper methods for viewdef bindings (avoid operators in paths)
function JobTracker:isFormModeAdd() return self.formMode == "add" end
function JobTracker:isFormModeEdit() return self.formMode == "edit" end
function JobTracker:hasResumes() return #self._resumes > 0 end
function JobTracker:noResumes() return #self._resumes == 0 end
function JobTracker:hasUnlinkableApps() return #self:unlinkableApps() > 0 end
function JobTracker:noUnlinkableApps() return #self:unlinkableApps() == 0 end
function JobTracker:showResumePreview() return self.selectedResume ~= nil or self.showMasterResume end
function JobTracker:hideResumePreview() return self.selectedResume == nil and not self.showMasterResume end
function JobTracker:currentResumePreviewUrl()
    local status = mcp:status()
    local base = "http://localhost:" .. status.mcp_port
    local ts = "?" .. os.time()
    if self.showMasterResume then
        return base .. "/job-tracker-storage/master-resume.md" .. ts
    elseif self.selectedResume then
        return base .. "/job-tracker-storage/resumes/" .. self.selectedResume.filename .. ts
    end
    return ""
end

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
for _, name in ipairs({"list", "detail", "form", "resume"}) do
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
for _, field in ipairs({"url", "location", "hqAddress", "website", "notes"}) do
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

-- Resume linking methods
function Application:linkedResume()
    if not self.resumeId then return nil end
    for _, resume in ipairs(jobTracker._resumes) do
        if resume.id == self.resumeId then
            return resume
        end
    end
    return nil
end

function Application:hasLinkedResume() return self.resumeId ~= nil end
function Application:noLinkedResume() return self.resumeId == nil end

function Application:resumeOptions()
    local result = {}
    for _, resume in ipairs(jobTracker._resumes) do
        table.insert(result, {
            value = tostring(resume.id),
            label = resume.name,
        })
    end
    return result
end

function Application:changeResume()
    local newId = self.selectedResumeId
    local oldResumeId = self.resumeId

    -- Unlink from old resume if any
    if oldResumeId then
        local oldResume = jobTracker:findResumeById(oldResumeId)
        if oldResume then
            oldResume:unlinkApp(self)
        end
    end

    -- Set new resumeId
    if newId == "" then
        self.resumeId = nil
    else
        self.resumeId = tonumber(newId)
        -- Link to new resume
        local newResume = jobTracker:findResumeById(self.resumeId)
        if newResume then
            newResume:linkApp(self)
        end
    end
    jobTracker:saveData()
end

function Application:goToResume()
    if self.resumeId then
        local resume = jobTracker:findResumeById(self.resumeId)
        if resume then
            jobTracker:showResumeView()
            jobTracker:selectResume(resume)
        end
    end
end

-- Link this application to the currently selected resume
function Application:linkToSelectedResume()
    if jobTracker.selectedResume then
        jobTracker:linkAppToResume(self)
    end
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

function Attachment:viewUrl()
    local idDir = string.format("%04d", self.applicationId)
    local status = mcp:status()
    return "http://localhost:" .. status.mcp_port .. "/job-tracker-storage/jobs/" .. idDir .. "/" .. self.filename
end

-- Resume methods
function Resume:idStr()
    return tostring(self.id)
end

function Resume:selectMe()
    jobTracker:selectResume(self)
end

function Resume:isSelected()
    return self == jobTracker.selectedResume
end

function Resume:linkedApps()
    local result = {}
    local appIds = self.applicationIds or {}
    for _, appId in ipairs(appIds) do
        for _, app in ipairs(jobTracker._applications) do
            if app.id == appId then
                table.insert(result, app)
                break
            end
        end
    end
    return result
end

-- Helper: create badge list with max count
local function createBadgeList(resume, maxCount)
    local apps = resume:linkedApps()
    local result = {}
    for i = 1, math.min(#apps, maxCount) do
        table.insert(result, session:create(ResumeBadge, {
            app = apps[i],
            resume = resume,
        }))
    end
    return result
end

-- For list-item display (5 max badges)
function Resume:linkedAppsBadges5()
    return createBadgeList(self, 5)
end

-- For detail display (10 max badges)
function Resume:linkedAppsBadges()
    return createBadgeList(self, 10)
end

function Resume:hasMoreApps()
    return #(self.applicationIds or {}) > 5
end

function Resume:noMoreApps()
    return not self:hasMoreApps()
end

function Resume:moreAppsCount()
    local total = #(self.applicationIds or {})
    return total - 5
end

function Resume:previewUrl()
    local status = mcp:status()
    return "http://localhost:" .. status.mcp_port .. "/job-tracker-storage/resumes/" .. self.filename .. "?" .. os.time()
end

function Resume:filePath()
    return RESUMES_DIR .. "/" .. self.filename
end

function Resume:unlinkApp(app)
    if not app then return end
    local newIds = {}
    for _, id in ipairs(self.applicationIds or {}) do
        if id ~= app.id then
            table.insert(newIds, id)
        end
    end
    self.applicationIds = newIds
    self.dateModified = today()
    jobTracker:saveData()
end

function Resume:linkApp(app)
    if not app then return end
    self.applicationIds = self.applicationIds or {}
    -- Check for duplicates
    for _, id in ipairs(self.applicationIds) do
        if id == app.id then return end
    end
    table.insert(self.applicationIds, app.id)
    self.dateModified = today()
    jobTracker:saveData()
end

function Resume:deleteMe()
    jobTracker:deleteSelectedResume()
end

-- ResumeBadge methods
function ResumeBadge:company()
    return self.app and self.app.company or ""
end

function ResumeBadge:goToApp()
    if self.app then
        jobTracker:selectApp(self.app)
    end
end

function ResumeBadge:unlinkMe()
    if self.resume and self.app then
        self.resume:unlinkApp(self.app)
    end
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

-- Instance creation
if not session.reloading then
    jobTracker = JobTracker:new()
end
