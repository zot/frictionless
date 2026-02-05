# Job Application Tracker

Track job applications through the hiring pipeline.

## UI Conventions

### Header Title Pattern
All views use a consistent header pattern: `<- Back  TITLE  [actions]`

The title appears in the header bar after the back button, not in a separate row below. Examples:
- List view: `Job Tracker  [Reload][Resume][+ Add]`
- Detail view: `<- Back  Acme Corp  [Edit][Delete]`
- Form view: `<- Back  ADD APPLICATION  [Save]`
- Resume view: `<- Back  RESUMES  [+ New][Master]`

## Core Features

### Application List
- List all job applications with key info at a glance
- Columns: Company, Position, Status, Date Applied
- Filter by status: All, Active (not archived/rejected/withdrawn), Offers, Archived
- Click an application to view details

### Application Statuses
- **Bookmarked** - Interested, haven't applied yet
- **Applied** - Application submitted
- **Phone Screen** - Initial recruiter/HR call
- **Technical** - Technical interview stage
- **Onsite** - Onsite/final round interviews
- **Offer** - Received offer
- **Rejected** - Application rejected
- **Withdrawn** - Withdrew application
- **Archived** - No longer active

### Application Details
Each application tracks:
- Company name
- Position title
- Job posting URL (shown in iframe when viewing details)
- Date added / Date applied
- Current status (dropdown to change)
- Salary range (min/max)
- Location / Remote status
- Company HQ address
- Notes (free text, shown in collapsible section)
- File attachments (resumes, cover letters, etc.)
- Linked resume (dropdown to select from available resumes)

### Detail View
When an application is selected:
- Show all application fields
- Notes shown in collapsible section (disabled when empty)
- File attachments section with:
  - Drag-and-drop zone for uploading files
  - File picker button
  - URL attachment button (placeholder)
  - List of attached files with delete button
  - Save/Revert buttons when attachments are modified
  - Warning dialog when leaving with unsaved attachment changes
- "View job posting" link to open URL in new tab
- Back button to return to list
- Edit button to modify fields
- Delete button to remove application

### Add Application
- Manual add form with all fields
- URL field auto-populates when pasting a job posting URL (via chat event to Claude)

### Activity Timeline
Each application has a timeline:
- Status changes (automatic)
- Notes added
- Custom events

## UI Layout

### List View (Default)
```
+------------------------------------------+
| Job Tracker           [Reload][Resume][+]|
+------------------------------------------+
| [All] [Active] [Offers] [Archived]       |
+------------------------------------------+
| COMPANY        POSITION      STATUS  DATE|
| ---------------------------------------- |
| Acme Corp      Sr Engineer   Phone   1/15|
| TechCo         Staff Eng     Applied 1/20|
| StartupXYZ     Lead Dev      Onsite  1/10|
+------------------------------------------+
```

### Detail View
```
+------------------------------------------+
| <- Back  Acme Corp         [Edit][Delete]|
+------------------------------------------+
| Senior Software Engineer                 |
| Status: [Phone Screen v]                 |
+------------------------------------------+
| Applied: Jan 15 | Remote | $180-220k     |
| HQ: San Francisco, CA                    |
| Resume: [AI Engineer 2026 v]             |
+------------------------------------------+
| [Open in new tab]                        |
+==========================================+
|                                          |
|        (iframe showing job URL)          |
|                                          |
+==========================================+
| TIMELINE                         [+ Note]|
| Jan 20 - Phone screen scheduled          |
| Jan 15 - Applied via website             |
+------------------------------------------+
```

### Add/Edit Form (replaces detail)
```
+------------------------------------------+
| <- Back  ADD APPLICATION         [Save]  |
+------------------------------------------+
| Company: [_______________]               |
| Position: [______________]               |
| URL: [_______________________]           |
| Status: [Bookmarked v]                   |
| Location: [______________]               |
| HQ: [____________________]               |
| Salary Min: [____] Max: [____]           |
| Notes:                                   |
| [                                      ] |
+------------------------------------------+
```

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

## Data Persistence

Store data in `.ui/storage/job-tracker/data/data.json`. Load on app start, save after each modification. Attachments are stored in `.ui/storage/job-tracker/data/jobs/<id>/` where `<id>` is a zero-padded 4-digit application ID. Changes are tracked with fossil SCM for version control.

### Claude Chat Panel

A chat field at the very bottom of the app for communicating with Claude:
- **Input field**: Multi-line textarea that auto-sizes from 1 line up to 4 lines, then scrolls
- **No send button**: Press Enter to send, Ctrl-Enter or Shift-Enter for newline
- **Output panel**: Flip-up panel above the input that shows chat history
  - Toggle via handle/button to the right of input
  - Scrollable when content overflows
  - Shows both user and assistant messages
  - **Click any message** to copy its content to the input field

## Events

| Event | Payload | Action |
|-------|---------|--------|
| `select` | `{id}` | Show application details |
| `back` | - | Return to list view |
| `add` | - | Show add form |
| `edit` | `{id}` | Show edit form |
| `save` | - | Save form changes |
| `cancel` | - | Cancel form, return to previous view |
| `status` | `{id, status}` | Change status |
| `note` | `{id, text}` | Add note to timeline |
| `delete` | `{id}` | Delete application |
| `filter` | `{filter}` | Filter list |
| `chat` | `{text}` | Send to Claude; Lua already adds user message to chat, Claude only adds assistant response |
| `resume_chat` | `{text, resumeId}` | Chat about a specific resume; Claude can read/edit the resume markdown |

### Resume Management

Manage markdown resume variants linked to applications.

#### Resume List View
- Scrollable list showing 6 resumes at a time
- Each resume shows name and up to 5 company badges (linked applications)
- Click resume to select and preview
- `[+ New]` button creates new resume from master template
- `[Master]` button previews the master resume

#### Resume Preview
- Selected resume shown as rendered HTML in iframe
- Symlink to `.ui/html/` enables automatic markdown rendering
- Above preview: all linked application badges with [x] to unlink
- `[+ Link]` picker to add application links

#### Resume Linking
Bidirectional linking between resumes and applications:
- In resume view: badges show linked apps, picker adds links, [x] removes links
- In application detail: dropdown to select/change/unlink resume

#### Resume Chat
Chat panel for Claude-assisted resume editing:
- Claude can read the resume markdown file
- Claude can read linked application details
- Claude can edit the resume and respond with changes made

#### Master Resume
- Always exists at `.ui/storage/job-tracker/data/master-resume.md`
- Used as template when creating new resumes
- Viewable via [Master] button
- Editable via chat

#### Storage
All resume data under `.ui/storage/job-tracker/data/` for fossil checkpointing:
- `master-resume.md` - Master resume
- `resumes/` - Individual resume variants
- Resume metadata stored in `data.json` alongside applications

#### HTML Serving
The app creates a symlink on initialization to enable markdown preview:
- `.ui/html/job-tracker-storage` -> `.ui/storage/job-tracker/data`
- This allows iframes to render markdown as HTML via `/job-tracker-storage/...`
- Symlink is created automatically in app init code (not manually)
