# Job Tracker - Testing

## Gaps

### Dead Methods (by design)

The following methods are defined but not directly called from viewdefs. They are either:
- Used internally by other methods
- Called by Claude during event handling
- Part of a complete API that may be used in future features

**Internal/API methods:**
- `Resume:filePath` - Returns path to resume file (for Claude to edit)
- `Resume:previewUrl` - Alternative to currentResumePreviewUrl
- `Resume:linkApp` - Used by linkAppToResume
- `JobTracker:hasSelectedResume` - Paired with noSelectedResume
- `JobTracker:noResumes` - Paired with hasResumes
- `JobTracker:isShowingMaster` - Paired with notShowingMaster
- `JobTracker:notShowingMaster` - Visibility helper
- `JobTracker:isResumeView` - Generated view method
- `JobTracker:isLinkPickerVisible` - Paired with isLinkPickerHidden
- `JobTracker:noUnlinkableApps` - Paired with hasUnlinkableApps
- `JobTracker:isDeleteResumeDialogHidden` - Paired with isDeleteResumeDialogVisible
- `JobTracker:isAttachmentWarningHidden` - Paired with isAttachmentWarningVisible
- `JobTracker:noResumeChatMessages` - Paired with hasResumeChatMessages
- `JobTracker:noChatMessages` - Paired with hasChatMessages
- `JobTracker:allApplications` - Returns raw _applications array
- `JobTracker:uploadFile` - Used by file upload JS bridge
- `JobTracker:prefillFromScrape` - Called by Claude when scraping job URLs
- `Application:linkedResume` - Returns the linked Resume object
- `Application:hasLinkedResume` - Paired with noLinkedResume
- `Application:noLinkedResume` - Visibility helper
- `Application:resumeOptions` - Intended for sl-select dynamic options
- `Application:noAttachments` - Paired with hasAttachments
- `Application:hasAttachmentsChanged` - Visibility helper (uses explicit check)
- `Application:hasSalary` - Paired with noSalary

### Dynamically Generated Methods

The following methods are generated using `for` loops and the audit tool doesn't detect them:
- `filterAll()`, `filterActive()`, `filterOffers()`, `filterArchived()`
- `allVariant()`, `activeVariant()`, `offersVariant()`, `archivedVariant()`
- `isFilterAll()`, `isFilterActive()`, `isFilterOffers()`, `isFilterArchived()`
- `sortCompany()`, `sortPosition()`, `sortStatus()`, `sortDate()`
- `companyIcon()`, `positionIcon()`, `statusIcon()`, `dateIcon()`
- `isListView()`, `isDetailView()`, `isFormView()`, `isResumeView()`
- `notListView()`, `notDetailView()`, `notFormView()`, `notResumeView()`
- `hasUrl()`, `noUrl()`, `hasLocation()`, `noLocation()`, `hasHq()`, `noHq()`, `hasNotes()`, `noNotes()`
