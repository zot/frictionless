# Job Tracker - Design

## Intent

Track job applications through the hiring pipeline. View application list, see details with embedded job posting, manage status, and maintain activity timeline. Manage resume variants linked to applications.

## UI Convention: Header Titles

Context/title goes in the header bar after the back button, not in a separate row below.

**Pattern:** `<- Back  TITLE  [actions]`

Examples:
- List view: `Job Tracker  [Reload][Resume][+]`
- Detail view: `<- Back  Acme Corp  [Edit][Delete]`
- Form view: `<- Back  ADD APPLICATION  [Save]`
- Resume view: `<- Back  RESUMES  [+ New][Master]`

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
                                        (o) <- FAB
```

When chat panel is open:
```
+------------------------------------------+
| [Paste job URL...]                    [>]|
+==========================================+
|  Assistant: I found the job details...   |
|  (auto-scrolls on new messages)          |
+------------------------------------------+
| [Chat with Claude...]         [🗑] [X]   |
+------------------------------------------+
```

Legend:
- `>` = Selected application
- `▼/▲` = Sort indicator (click column header to sort/reverse)
- `[🗑]` = Clear chat button
- `[X]` = Close chat panel
- Filter buttons toggle active state
- Status shown as badge with variant color

### Chat FAB

A floating action button (FAB) appears in the bottom-right corner above the status bar on all screens when the chat panel is closed. Clicking it opens the chat panel. The FAB uses fixed positioning (`bottom: 52px`, `right: 20px`) with `z-index: 100`. Content areas that could be obscured by the FAB include a spacer element (`.jt-fab-spacer`, 56px wide) to prevent overlap.

### Detail View
```
+------------------------------------------+
| <- Back  Acme Corp  [✓][↺][Edit][Delete] |
+------------------------------------------+
| Senior Software Engineer                 |
| Status: [Phone Screen v]                 |
+------------------------------------------+
| Applied: Jan 15 | Remote | $180-220k     |
| HQ: San Francisco, CA                    |
| Resume: [AI Engineer 2026 v]             |
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
- Resume dropdown shows all resumes + "(none)" option

### Add/Edit Form
```
+------------------------------------------+
| <- Back  ADD APPLICATION         [Save]  |
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
| <- Back  RESUMES          [+ New][Master]|
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
| [Chat with Claude about this resume...]  |
+==========================================+
|  Claude: I can help tailor this...       |
+------------------------------------------+
| [Type here...]                           |
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
- Chat panel for Claude-assisted editing

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
| chatInput | string | Claude chat input field |
| chatMessages | ChatMessage[] | Chat message history |
| chatPanelOpen | boolean | Whether output panel is open |
| sortColumn | string | Current sort column: "company", "position", "status", "date" |
| sortDirection | string | Sort direction: "asc" or "desc" |
| selectedStatus | string | Status value for detail view dropdown (synced with selected.status) |
| _fileUploadData | string | JS-to-Lua bridge for file uploads |
| showAttachmentWarning | boolean | Show warning dialog when leaving with unsaved attachments |
| _resumes | Resume[] | All resume instances |
| selectedResume | Resume | Currently selected resume (nil for master) |
| showMasterResume | boolean | Whether showing master resume in preview |
| resumeChatInput | string | Chat input for resume view |
| resumeChatMessages | ChatMessage[] | Chat history for resume view |
| resumeChatPanelOpen | boolean | Whether resume chat panel is open |
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

### ChatMessage

| Field | Type | Description |
|-------|------|-------------|
| role | string | "user" or "assistant" |
| content | string | Message content |

### ChatMessage Methods

| Method | Description |
|--------|-------------|
| isUser() | Returns role == "user" |
| isAssistant() | Returns role == "assistant" |
| copyToInput() | Copy message content to jobTracker.chatInput |

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
| selectMe() | Select this resume in jobTracker |
| isSelected() | Returns self == jobTracker.selectedResume |
| linkedApps() | Returns array of linked Application objects |
| linkedAppsBadges(max) | Returns up to max linked apps for display |
| hasMoreApps(max) | Returns true if more than max linked apps |
| moreAppsCount(max) | Returns count of apps beyond max |
| previewUrl() | Returns URL for iframe preview |
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
| submitChat() | Send chatInput to Claude via pushState, clear input |
| toggleChatPanel() | Toggle chatPanelOpen state |
| clearChat() | Clear all chat messages |
| addChatMessage(role, content) | Add message to chatMessages |
| chatPanelHidden() | Returns not chatPanelOpen |
| hasChatMessages() | Returns chatMessages has items |
| noChatMessages() | Returns chatMessages is empty |
| mutate() | Hot-reload mutation to initialize chatMessages on existing instances |
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
| submitResumeChat() | Send resumeChatInput to Claude via pushState |
| toggleResumeChatPanel() | Toggle resumeChatPanelOpen |
| clearResumeChat() | Clear resume chat messages |
| addResumeChatMessage(role, content) | Add message to resumeChatMessages |
| resumeChatPanelHidden() | Returns not resumeChatPanelOpen |
| hasResumeChatMessages() | Returns resumeChatMessages has items |
| noResumeChatMessages() | Returns resumeChatMessages is empty |
| currentResumePreviewUrl() | Returns URL for current preview (selected or master) |
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
| changeResume() | Update resumeId from selectedResumeId dropdown |
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
| JobTracker.ChatMessage.list-item.html | ChatMessage | Message in chat output panel |
| JobTracker.Attachment.list-item.html | Attachment | Row in attachments list |
| JobTracker.Resume.list-item.html | Resume | Row in resume list (with badges) |
| JobTracker.ResumeBadge.list-item.html | ResumeBadge | Badge showing linked app |

## Events

### From UI to Claude

```json
{"app": "job-tracker", "event": "chat", "text": "https://...", "handler": "/ui-fast"}
```

### Claude Event Handling

| Event | Action |
|-------|--------|
| `chat` | Handle as URL or general chat (see below). **Note:** Lua already adds the user message to chat history, so Claude only adds the assistant response. |
| `resume_chat` | Chat about a resume (see below). **Note:** Lua already adds the user message to resumeChatMessages, so Claude only adds the assistant response. |

#### Resume Chat Handling

When receiving a `resume_chat` event:

1. **Read context:**
   - `resumeId` - The resume being discussed (nil for master)
   - `text` - The user's message
   - Read the resume markdown file (master or specific)
   - If resume has linked apps, read their details for context

2. **Handle request:**
   - If user asks to edit, modify the markdown file directly
   - If user asks to tailor for a job, read the linked application details
   - Provide suggestions, make edits, answer questions

3. **Respond:**
   ```lua
   jobTracker:addResumeChatMessage("assistant", "response text")
   ```

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
