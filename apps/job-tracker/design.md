# Job Tracker - Design

## Intent

Track job applications through the hiring pipeline. View application list, see details with embedded job posting, manage status, and maintain activity timeline. Manage resume variants linked to applications.

## UI Convention: Header Titles

Context/title goes in the header bar after the back button, not in a separate row below.

**Pattern:** `[icon]  TITLE  [actions]`

Back navigation uses icon-only buttons:
- Detail/Resume views: house icon (returns to list)
- Form view: arrow-left icon (cancels form)

Examples:
- List view: `Job Tracker (bookmarklet)  [Reload][Resume][+]`
- Detail view: `[🏠] Acme Corp  [Edit][Delete]`
- Form view: `[←] ADD APPLICATION  [Save]`
- Resume view: `[🏠] RESUMES  [↺][+ New][Master]`

## Layout

### List View (Default)
```
+------------------------------------------+
| Job Tracker           [Reload][Resume][+]|
+------------------------------------------+
| [All] [Active] [Offers] [Archived]       |
+------------------------------------------+
| COMPANY▼       POSITION      STATUS  DATE| <- sortable headers
| ---------------------------------------- |
| > Acme Corp    Sr Engineer   Phone   1/15|
|   TechCo       Staff Eng     Applied 1/20|
|   StartupXYZ   Lead Dev      Onsite  1/10|
+------------------------------------------+
| [Paste job URL...]                    [>]|
+------------------------------------------+
```

When bookmarklet section is expanded:
```
+------------------------------------------+
| Job Tracker (bookmarklet) [Reload][R][+] |
+------------------------------------------+
| Drag this to your bookmarks bar: [Add Job]|
+------------------------------------------+
| [All] [Active] [Offers] [Archived]       |
```

Legend:
- `>` = Selected application
- `▼/▲` = Sort indicator (click column header to sort/reverse)
- Filter buttons toggle active state
- Status shown as badge with variant color

### Detail View
```
+------------------------------------------+
| [🏠] Acme Corp     [✓][↺][Edit][Delete]  |
+------------------------------------------+
| Senior Software Engineer                 |
| Status: [Phone Screen v]                 |
+------------------------------------------+
| Applied: Jan 15 | Remote | $180-220k     |
| HQ: San Francisco, CA                    |
| Resume: [AI Engineer 2026 v] [↗]         |
+------------------------------------------+
| [Notes (empty)]  <- collapsible section  |
+------------------------------------------+
| ATTACHMENTS              [+ File][+ URL] |
| +--------------------------------------+ |
| | resume.pdf                    [x]    | |
| | cover-letter.docx             [x]    | |
| +--------------------------------------+ |
| | Drop files here to attach            | |
+------------------------------------------+
| [View job posting ↗]                     |
+------------------------------------------+
| TIMELINE                         [+ Note]|
| ---------------------------------------- |
| Jan 20 - Status: Phone Screen            |
| Jan 15 - Added application               |
+------------------------------------------+
```

Legend:
- `[✓]` = Save attachments (shown when attachments changed)
- `[↺]` = Revert attachments (shown when attachments changed)
- `[x]` = Delete attachment button
- Resume dropdown shows all resumes + "(none)" option (uses `ui-view` with `lua.ViewListItem.resume-option.html`)
- `[↗]` = Go to linked resume in resume view (hidden when no resume linked)

### Add/Edit Form
```
+------------------------------------------+
| [←] ADD APPLICATION              [Save]  |
+------------------------------------------+
| Company: [_______________]               |
| Position: [______________]               |
| URL: [_______________________]           |
| Status: [Bookmarked v]                   |
| Location: [______________]               |
| HQ Address: [____________]               |
| Salary Min: [____] Max: [____]           |
| Notes:                                   |
| [                                      ] |
+------------------------------------------+
```

Title shows "ADD APPLICATION" or "EDIT APPLICATION" based on formMode.

### Resume View
```
+------------------------------------------+
| [🏠] RESUMES      [↺][+ New][Master]     |
+------------------------------------------+
| > AI Engineer        [Anthropic][Google] |
|   Full Stack 2026    [JuliaHub][Stripe]  |
|   Backend Focused                        |
|   Startup Generalist [Acme]              |
|   ...                                    |
|   (scrollable, 6 visible)                |
+------------------------------------------+
| [Anthropic x][Google x][Meta x] [+ Link] |
+------------------------------------------+
|  ┌────────────────────────────────────┐  |
|  │  Bill Burdick                      │  |
|  │  Software Architect...             │  |
|  │                                    │  |
|  │  (HTML preview of selected .md)    │  |
|  │  (rendered via iframe)             │  |
|  └────────────────────────────────────┘  |
+------------------------------------------+
```

Legend:
- List shows 6 resumes, scrollable
- Each resume shows name + up to 5 company badges (linked apps)
- Selected resume shows all linked app badges above preview
- `[x]` in badge removes the link
- `[+ Link]` opens picker to add application link
- `[+ New]` creates resume from master template
- `[Master]` shows master resume in preview
- `[↺]` = Reload data from disk

## Data Model

### JobTracker (main app)

| Field | Type | Description |
|-------|------|-------------|
| _applications | Application[] | All applications |
| selected | Application | Currently selected application |
| filter | string | Current filter: "all", "active", "offers", "archived" |
| view | string | Current view: "list", "detail", "form" |
| formMode | string | "add" or "edit" |
| formData | FormData | Form fields for add/edit |
| noteInput | string | Input for adding timeline notes |
| urlInput | string | Input for pasting job URLs to scrape |
| showBookmarklet | boolean | Collapsible bookmarklet install section |
| sortColumn | string | Current sort column: "company", "position", "status", "date" |
| sortDirection | string | Sort direction: "asc" or "desc" |
| selectedStatus | string | Status value for detail view dropdown (synced with selected.status) |
| _fileUploadData | string | JS-to-Lua bridge for file uploads |
| showAttachmentWarning | boolean | Show warning dialog when leaving with unsaved attachments |
| _resumes | Resume[] | All resume instances |
| selectedResume | Resume | Currently selected resume (nil for master) |
| showMasterResume | boolean | Whether showing master resume in preview |
| showDeleteResumeDialog | boolean | Show confirm dialog for resume deletion |
| showLinkPicker | boolean | Show application link picker |

### Application

| Field | Type | Description |
|-------|------|-------------|
| id | number | Sequential numeric ID |
| company | string | Company name |
| position | string | Position title |
| url | string | Job posting URL |
| status | string | Current status |
| dateAdded | string | Date added (ISO) |
| dateApplied | string | Date applied (ISO, nil if bookmarked) |
| location | string | Job location / remote |
| hqAddress | string | Company HQ address |
| salaryMin | number | Salary range min |
| salaryMax | number | Salary range max |
| notes | string | Free text notes |
| timeline | TimelineEvent[] | Activity timeline |
| _attachmentsCache | Attachment[] | Cached list of attachments |
| attachmentsChanged | boolean | Whether attachments have been modified |
| resumeId | number | Linked resume ID (nil if none) |

### Resume

| Field | Type | Description |
|-------|------|-------------|
| id | number | Sequential ID |
| name | string | Display name (e.g., "AI Engineer 2026") |
| filename | string | Markdown filename (e.g., "ai-engineer-2026.md") |
| applicationIds | number[] | Linked application IDs |
| dateCreated | string | Date created (ISO) |
| dateModified | string | Date modified (ISO) |

### FormData

| Field | Type | Description |
|-------|------|-------------|
| company | string | Company input |
| position | string | Position input |
| url | string | URL input |
| status | string | Status select |
| location | string | Location input |
| hqAddress | string | HQ address input |
| salaryMin | string | Salary min input |
| salaryMax | string | Salary max input |
| notes | string | Notes textarea |
| dateApplied | string | Date applied (ISO format, editable in form) |
| _original | table | Original field values for change tracking |

### FormData Methods

| Method | Description |
|--------|-------------|
| hasChanges() | Returns true if any field differs from _original |
| noChanges() | Returns not hasChanges() (for disabling save button) |

### TimelineEvent

| Field | Type | Description |
|-------|------|-------------|
| date | string | Event date (ISO) |
| event | string | Event type: "added", "applied", "status_change", "note" |
| note | string | Event description |
| fromStatus | string | Previous status (for status_change) |
| toStatus | string | New status (for status_change) |

### Attachment

| Field | Type | Description |
|-------|------|-------------|
| filename | string | File name |
| path | string | Full path to file |
| applicationId | number | Parent application ID |

### Attachment Methods

| Method | Description |
|--------|-------------|
| deleteMe() | Delete this attachment |
| icon() | Returns icon name based on file extension |
| downloadUrl() | Returns file:// URL for download |

### Resume

| Method | Description |
|--------|-------------|
| idStr() | Returns string representation of id (for dropdown value binding) |
| selectMe() | Select this resume in jobTracker |
| isSelected() | Returns self == jobTracker.selectedResume |
| linkedApps() | Returns array of linked Application objects |
| linkedAppsBadges(max) | Returns up to max linked apps for display |
| hasMoreApps(max) | Returns true if more than max linked apps |
| moreAppsCount(max) | Returns count of apps beyond max |
| previewUrl() | Returns URL for iframe preview (with cache-busting timestamp) |
| filePath() | Returns full path to markdown file |
| unlinkApp(app) | Remove app from applicationIds |
| linkApp(app) | Add app to applicationIds |
| deleteMe() | Delete resume (with confirmation) |

### ResumeBadge

| Field | Type | Description |
|-------|------|-------------|
| app | Application | The linked application |
| resume | Resume | Parent resume |

| Method | Description |
|--------|-------------|
| company() | Returns app.company |
| goToApp() | Navigate to application detail view |
| unlinkMe() | Remove this app from resume's links |

## Status Values

| Status | Display | Badge Variant |
|--------|---------|---------------|
| bookmarked | Bookmarked | neutral |
| applied | Applied | primary |
| phone | Phone Screen | primary |
| technical | Technical | primary |
| onsite | Onsite | primary |
| offer | Offer | success |
| rejected | Rejected | danger |
| withdrawn | Withdrawn | neutral |
| archived | Archived | neutral |

## Methods

### JobTracker

| Method | Description |
|--------|-------------|
| applications() | Returns filtered and sorted applications |
| allApplications() | Returns _applications for binding |
| loadData() | Load from data.json |
| saveData() | Save to data.json |
| reload() | Reload data from disk (for external edits) |
| setFilter(f) | Set filter and clear selection |
| selectApp(app) | Select application, set selectedStatus, show detail view |
| showList() | Return to list view (warns if unsaved attachments) |
| hideAttachmentWarning() | Hide the attachment warning dialog |
| isAttachmentWarningVisible() | Returns showAttachmentWarning |
| isAttachmentWarningHidden() | Returns not showAttachmentWarning |
| showAddForm() | Show add form with empty formData |
| showEditForm() | Show edit form with selected app data |
| saveForm() | Save form data as new or updated application |
| cancelForm() | Cancel form, return to previous view |
| addNote() | Add note to selected application timeline |
| deleteApp() | Delete selected application |
| changeStatus() | Change selected application status from selectedStatus |
| uploadFile(filename, content) | Upload file attachment to selected app |
| deleteAttachment(attachment) | Delete attachment file |
| saveAttachments() | Commit attachment changes to fossil |
| revertAttachments() | Revert attachment changes from fossil |
| saveAttachmentsAndBack() | Save attachments and return to list |
| revertAttachmentsAndBack() | Revert attachments and return to list |
| processFileUpload() | Process file upload from JS-to-Lua bridge |
| promptAttachUrl() | Placeholder for URL attachment (not implemented) |
| submitUrl() | Send urlInput to Claude for scraping via pushState |
| filterAll() | Set filter to "all" |
| filterActive() | Set filter to "active" |
| filterOffers() | Set filter to "offers" |
| filterArchived() | Set filter to "archived" |
| isFilterAll() | Returns filter == "all" |
| isFilterActive() | Returns filter == "active" |
| isFilterOffers() | Returns filter == "offers" |
| isFilterArchived() | Returns filter == "archived" |
| isListView() | Returns view == "list" |
| isDetailView() | Returns view == "detail" |
| isFormView() | Returns view == "form" |
| notListView() | Returns view ~= "list" |
| notDetailView() | Returns view ~= "detail" |
| notFormView() | Returns view ~= "form" |
| toggleBookmarklet() | Toggle showBookmarklet state |
| isBookmarkletHidden() | Returns not showBookmarklet |
| findResumeById(id) | Find resume by ID in _resumes |
| repairResumeLinks() | Repair bidirectional links between apps and resumes on load |
| mutate() | Hot-reload mutation to initialize new fields on existing instances |
| toggleSort(column) | Toggle sort on column; if same column, reverse direction |
| sortCompany() | Call toggleSort("company") |
| sortPosition() | Call toggleSort("position") |
| sortStatus() | Call toggleSort("status") |
| sortDate() | Call toggleSort("date") |
| sortIcon(column) | Returns "▲" or "▼" if sorted by column, else "" |
| companyIcon() | Returns sortIcon("company") |
| positionIcon() | Returns sortIcon("position") |
| statusIcon() | Returns sortIcon("status") |
| dateIcon() | Returns sortIcon("date") |
| showResumeView() | Show resume view |
| isResumeView() | Returns view == "resume" |
| notResumeView() | Returns view ~= "resume" |
| resumes() | Returns _resumes array |
| selectResume(resume) | Select a resume for preview |
| showMaster() | Show master resume in preview |
| isShowingMaster() | Returns showMasterResume |
| notShowingMaster() | Returns not showMasterResume |
| createResume() | Create new resume from master template |
| deleteSelectedResume() | Delete selected resume (shows confirm) |
| confirmDeleteResume() | Confirm and delete resume |
| cancelDeleteResume() | Cancel delete dialog |
| isDeleteResumeDialogVisible() | Returns showDeleteResumeDialog |
| isDeleteResumeDialogHidden() | Returns not showDeleteResumeDialog |
| toggleLinkPicker() | Toggle application link picker |
| isLinkPickerVisible() | Returns showLinkPicker |
| isLinkPickerHidden() | Returns not showLinkPicker |
| unlinkableApps() | Returns apps not linked to selected resume |
| linkAppToResume(app) | Link app to selected resume |
| currentResumePreviewUrl() | Returns URL for current preview (selected or master) with cache-busting timestamp |
| hasSelectedResume() | Returns selectedResume ~= nil |
| noSelectedResume() | Returns selectedResume == nil |
| loadResumes() | Load resumes from data.json |
| saveResumes() | Save resumes to data.json |

### Application

| Method | Description |
|--------|-------------|
| selectMe() | Call jobTracker:selectApp(self) |
| isSelected() | Returns self == jobTracker.selected |
| statusDisplay() | Returns human-readable status |
| statusVariant() | Returns badge variant for status |
| dateDisplay() | Returns formatted date applied or added |
| salaryDisplay() | Returns formatted salary range or empty |
| hasUrl() | Returns url is not empty |
| noUrl() | Returns url is empty |
| hasLocation() | Returns location is not empty |
| noLocation() | Returns location is empty |
| hasSalary() | Returns salary display is not empty |
| noSalary() | Returns salary display is empty |
| hasHq() | Returns hqAddress is not empty |
| noHq() | Returns hqAddress is empty |
| hasNotes() | Returns notes is not empty |
| noNotes() | Returns notes is empty |
| idDir() | Returns zero-padded 4-digit ID string |
| attachmentsDir() | Returns path to attachments directory |
| attachments() | Returns cached list of Attachment objects |
| clearAttachmentsCache() | Clear attachment cache (after file operations) |
| hasAttachments() | Returns attachments count > 0 |
| noAttachments() | Returns attachments count == 0 |
| hasAttachmentsChanged() | Returns attachmentsChanged |
| noAttachmentsChanged() | Returns not attachmentsChanged |
| appliedDateDisplay() | Returns formatted date applied |
| linkedResume() | Returns Resume linked to this app (or nil) |
| hasLinkedResume() | Returns resumeId ~= nil |
| noLinkedResume() | Returns resumeId == nil |
| resumeOptions() | Returns all resumes for dropdown |
| changeResume() | Update resumeId from selectedResumeId dropdown (unlinks old, links new bidirectionally) |
| goToResume() | Navigate to resume view and select the linked resume |
| selectedResumeId | string | Dropdown value for resume selection |

### TimelineEvent

| Method | Description |
|--------|-------------|
| dateDisplay() | Returns formatted date |
| description() | Returns event description |

## ViewDefs

| File | Type | Purpose |
|------|------|---------|
| JobTracker.DEFAULT.html | JobTracker | Main layout with list/detail/form/resume views |
| JobTracker.Application.list-item.html | Application | Row in application list |
| JobTracker.TimelineEvent.list-item.html | TimelineEvent | Row in timeline |
| JobTracker.Attachment.list-item.html | Attachment | Row in attachments list |
| JobTracker.Resume.list-item.html | Resume | Row in resume list (with badges via ui-view) |
| JobTracker.ResumeBadge.list-item.html | ResumeBadge | Badge showing linked app |
| lua.ViewListItem.resume-option.html | ViewListItem | sl-option for resume dropdown in detail view |

## Events

### From UI to Claude

```json
{"app": "job-tracker", "event": "chat", "text": "https://...", "handler": "/ui-fast"}
{"app": "job-tracker", "event": "page_received", "url": "https://...", "title": "Sr Engineer at Acme", "text": "...rendered page text..."}
```

### Claude Event Handling

| Event | Action |
|-------|--------|
| `chat` | Handle as URL submission (see below). Chat is handled by MCP shell. |
| `page_received` | Page content from bookmarklet (see below). Extract structured data from text, prefill add form. No fetching needed — text is pre-rendered `innerText` from the user's browser. |

#### URL Chat Handling

When the chat text is a URL (job posting link):

1. **Fetch and scrape** the job posting for: company, position, location, salary (if available)
2. **Show add form** with prefilled data:
   ```lua
   jobTracker:showAddForm()
   jobTracker.formData.company = "..."
   jobTracker.formData.position = "..."
   jobTracker.formData.url = "<the URL>"
   jobTracker.formData.location = "..."
   ```
3. **If salary is empty**, web search for typical salary range:
   - Search for "{company} {position} salary range"
   - Use data from sources like Glassdoor, Levels.fyi, LinkedIn, or similar
   - Fill in salaryMin and salaryMax with the found range
   - Note in chat that salary was estimated from market data
4. **If hqAddress is empty**, search for company HQ:
   - First try: US headquarters
   - Fallback: international HQ if no US location
   - Update the formData
5. **Reply in chat** with confirmation of what was found

#### Page Received Handling (Bookmarklet)

When a `page_received` event arrives (from the publisher bookmarklet via `init.lua`):

The event contains `url`, `title`, and `text` (the page's `innerText`, already clean — no HTML, no fetching needed).

1. **Extract structured data** from the `text` and `title` fields: company, position, location, salary (if available)
2. **Show add form** with prefilled data:
   ```lua
   jobTracker:showAddForm()
   jobTracker.formData.company = "..."
   jobTracker.formData.position = "..."
   jobTracker.formData.url = "<the url from the event>"
   jobTracker.formData.location = "..."
   ```
3. **If salary is empty**, web search for typical salary range (same as URL flow)
4. **If hqAddress is empty**, search for company HQ (same as URL flow)
5. **Reply in chat** with confirmation: "Added application for {company} - {position}"

This flow is identical to URL Chat Handling except:
- The page content is already available (no WebFetch/Playwright needed)
- The text is `innerText` (clean, no HTML parsing)
- The `title` field often contains company/position info
- No "user: {url}" chat message is added (the event is distinct from `chat`)

## File I/O

### Storage Structure
```
.ui/storage/job-tracker/
├── data/
│   ├── data.json           # Application data + resume metadata
│   ├── master-resume.md    # Master resume template
│   ├── resumes/            # Resume variants
│   │   ├── ai-engineer-2026.md
│   │   └── full-stack-2026.md
│   └── jobs/
│       ├── 0001/           # Attachments for app ID 1
│       │   ├── resume.pdf
│       │   └── cover.docx
│       └── 0002/           # Attachments for app ID 2
└── data.fossil             # Fossil repo for version control
```

### HTML Serving Symlinks

To render markdown as HTML in iframes, a symlink is created on app initialization:
```
.ui/html/job-tracker-storage -> .ui/storage/job-tracker/data
```

Preview URLs:
- Master resume: `/job-tracker-storage/master-resume.md`
- Resume variant: `/job-tracker-storage/resumes/ai-engineer-2026.md`

The ui-engine auto-renders `.md` files as HTML when served.

**Implementation:** The symlink is created in the app's init code (not manually) so it works for all users who install the app. The `JobTracker:new()` method calls `ensureStorageSymlink()` which creates the symlink if it doesn't exist.

### Loading
On app init, call `loadData()` which reads `.ui/storage/job-tracker/data/data.json` using Lua `io.open`.

### Saving
After any modification (add, edit, delete, status change, note), call `saveData()` which writes to `data.json` and commits to fossil.

### Attachments
Files are stored in `.ui/storage/job-tracker/data/jobs/<id>/` where `<id>` is the zero-padded 4-digit application ID. Attachment changes are tracked separately and must be explicitly saved or reverted.

### JSON Format
```json
{
  "applications": [
    {
      "id": 1,
      "company": "Acme Corp",
      "position": "Senior Engineer",
      "url": "https://...",
      "status": "phone",
      "dateAdded": "2025-01-15",
      "dateApplied": "2025-01-15",
      "location": "Remote",
      "hqAddress": "San Francisco, CA",
      "salaryMin": 180000,
      "salaryMax": 220000,
      "notes": "",
      "resumeId": 1,
      "timeline": [
        {"date": "2025-01-15", "event": "added", "note": "Added application"}
      ]
    }
  ],
  "resumes": [
    {
      "id": 1,
      "name": "AI Engineer 2026",
      "filename": "ai-engineer-2026.md",
      "applicationIds": [1, 3, 5],
      "dateCreated": "2026-01-15",
      "dateModified": "2026-01-20"
    }
  ]
}
```
