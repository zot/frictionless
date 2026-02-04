# Job Application Tracker

Track job applications through the hiring pipeline.

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
|  Job Tracker              [Reload][+ Add]|
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
| <- Back                           [Edit] |
+------------------------------------------+
| Acme Corp                                |
| Senior Software Engineer                 |
| Status: [Phone Screen v]                 |
+------------------------------------------+
| Applied: Jan 15 | Remote | $180-220k     |
| HQ: San Francisco, CA                    |
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
| <- Cancel                        [Save]  |
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
