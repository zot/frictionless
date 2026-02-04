# Job Tracker for Frictionless UI

Track job applications through the hiring pipeline with an AI-powered assistant.

## Features

### Application List
- Sortable columns: Company, Position, Status, Date
- Click headers to sort ascending/descending
- Filter tabs: All, Active, Offers, Archived
- Click any row to view details
- Reload button to refresh from disk

### Application Details
Each application tracks:
- Company name and position title
- Job posting URL (viewable in embedded iframe)
- Date added and date applied
- Current status (quick-change dropdown)
- Salary range (min/max)
- Location / Remote status
- Company HQ address
- Notes (collapsible section)

### Status Pipeline
- **Bookmarked** - Interested, haven't applied yet
- **Applied** - Application submitted
- **Phone Screen** - Initial recruiter/HR call
- **Technical** - Technical interview stage
- **Onsite** - Onsite/final round interviews
- **Offer** - Received offer
- **Rejected** / **Withdrawn** / **Archived**

### File Attachments
- Drag-and-drop files onto any application
- File picker button for selecting files
- Attach resumes, cover letters, offer documents
- Save/Revert buttons appear when files are modified
- Warning dialog prevents losing unsaved changes
- Files stored locally with fossil version control

### Activity Timeline
- Automatic status change logging
- Add custom notes to any application
- Full history of each application's journey

### Claude Chat Integration
- Floating action button (FAB) opens chat panel from any screen
- **Paste a job URL** - Claude scrapes company, position, location, salary
- Claude searches for salary data if not listed in posting
- Claude finds company HQ address
- Click any chat message to copy it to input
- Clear chat and close panel buttons

## Data Storage

All data stored locally in `.ui/storage/job-tracker/`:
```
data/
├── data.json           # Application data
└── jobs/
    ├── 0001/           # Attachments for app ID 1
    │   ├── resume.pdf
    │   └── cover.docx
    └── 0002/           # Attachments for app ID 2
```

Version controlled with fossil for history and rollback.

## UI Layout

### List View
```
+------------------------------------------+
|  Job Tracker              [Reload][+ Add]|
+------------------------------------------+
| [All] [Active] [Offers] [Archived]       |
+------------------------------------------+
| COMPANY▼       POSITION      STATUS  DATE|
| ---------------------------------------- |
| > SoFi         Staff Eng     Applied 2/1 |
|   Deel         Sr Full Stack Applied 1/28|
|   Stripe       Backend Eng   Phone   1/25|
+------------------------------------------+
| [Paste job URL...]                    [>]|
+------------------------------------------+
                                        (o) <- Chat FAB
```

### Detail View
```
+------------------------------------------+
| <- Back           [✓][↺][Edit][Delete]   |
+------------------------------------------+
| Acme Corp                                |
| Senior Software Engineer                 |
| Status: [Phone Screen v]                 |
+------------------------------------------+
| Applied: Jan 15 | Remote | $180-220k     |
| HQ: San Francisco, CA                    |
+------------------------------------------+
| [Notes]  <- collapsible                  |
+------------------------------------------+
| ATTACHMENTS                    [+ File]  |
| resume.pdf                          [x]  |
| cover-letter.docx                   [x]  |
| Drop files here to attach                |
+------------------------------------------+
| [View job posting ↗]                     |
+------------------------------------------+
| TIMELINE                         [+ Note]|
| Jan 20 - Status: Phone Screen            |
| Jan 15 - Added application               |
+------------------------------------------+
```

## Getting Started

1. Install [Frictionless UI](https://github.com/zot/frictionless)
2. Run `frictionless mcp --port 8000 --dir .ui`
3. Open browser to the UI
4. Select "Job Tracker" from the app console

## License

Part of [Frictionless UI](https://github.com/zot/frictionless) - MIT License
