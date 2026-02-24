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
- Company HQ street address (specific street address, not just city/state)
- Company website URL
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
| [🏠] Acme Corp            [Edit][Delete] |
+------------------------------------------+
| Senior Software Engineer                 |
| Status: [Phone Screen v]                 |
+------------------------------------------+
| Applied: Jan 15 | Remote | $180-220k     |
| HQ: 123 Market St, San Francisco, CA     |
| Web: acmecorp.com                        |
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
| [←] ADD APPLICATION              [Save]  |
+------------------------------------------+
| Company: [_______________]               |
| Position: [______________]               |
| URL: [_______________________]           |
| Status: [Bookmarked v]                   |
| Location: [______________]               |
| HQ: [____________________]               |
| Website: [_________________]             |
| Salary Min: [____] Max: [____]           |
| Notes:                                   |
| [                                      ] |
+------------------------------------------+
```

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

## Data Persistence

Store data in `.ui/storage/job-tracker/data/data.json`. Load on app start, save after each modification. Attachments are stored in `.ui/storage/job-tracker/data/jobs/<id>/` where `<id>` is a zero-padded 4-digit application ID. Changes are tracked with fossil SCM for version control.

### Bookmarklet

A collapsible section in the list view header provides a "bookmarklet" link:
- Toggle via "(bookmarklet)" text link next to header title
- When expanded, shows a draggable "Add Job" button to bookmark bar
- Bookmarklet sends the current page's URL, title, and text to the publisher endpoint
- Claude receives a `page_received` event and prefills the add form

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
| `chat` | `{text}` | URL submission to Claude for scraping |
| `page_received` | `{url, title, text}` | Page content from bookmarklet; Claude extracts data and prefills form |

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
- In application detail: dropdown to select/change/unlink resume, with "go to resume" button
- Changing resume in dropdown properly unlinks from old resume and links to new
- Links are repaired on data load to ensure consistency
- Dropdown uses `ui-view` pattern with themed options

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
