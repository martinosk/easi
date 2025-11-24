# Release Notes

## User Need
As a user, I need to be informed about new features and important changes when I open the application after an update, so I can understand what's new and take advantage of improved functionality.

## Success Criteria
- Users see release notes overlay on first launch after an update
- Users can dismiss the overlay with preference options: "Hide forever" or "Hide until next release"
- Release notes are concise, relevant, and categorized for easy scanning
- Developers can efficiently generate and tag releases
- User preferences persist across sessions

## Vertical Slices
Ordered list of end-to-end slices, each delivering incremental value:

### Slice 1: Release Version Management
- [x] Add version field to backend configuration/environment
- [x] Create API endpoint GET /api/v1/version returning current version
- [x] Frontend fetches and stores current version in localStorage on startup

### Slice 2: Release Notes Content Storage
- [x] Create releases bounded context with Release aggregate. Make it simple, no need for event sourcing, tenancy or other or fancy stuff.
- [x] Store release notes as markdown with version, date, and categorized items
- [x] API endpoint GET /api/v1/releases/latest for current release notes
- [x] API endpoint GET /api/v1/releases/:version for specific version

### Slice 3: Release Notes Generation Tool
- [x] Create Claude command /generate-release-notes
- [x] Parse completed specs (done status) since last tagged release
- [x] Extract commits since last release tag
- [x] Generate categorized markdown from specs and commits
- [x] Filter out technical items (refactoring, internal improvements)
- [x] Output draft release notes for developer review

### Slice 4: Release Notes Display
- [x] Create ReleaseNotesOverlay component with dismiss options
- [x] Store user preferences in localStorage (dismissedVersion, dismissMode)
- [x] Display overlay on startup when new version detected and preferences allow
- [x] Style overlay consistently with existing dialog patterns
- [x] Include categories: Major, Bugs, API (exclude technical refactoring)

### Slice 5: Release Tagging Workflow
- [x] Create Claude command /tag-release [version]
- [x] Update version in backend configuration
- [x] Store finalized release notes in database
- [x] Create git tag with version number
- [x] Commit version bump and release notes

### Slice 6: Release Notes Browser
- [x] Add "Release Notes" menu item to UI (e.g., help menu or settings)
- [x] Create ReleaseNotesBrowser component showing release history
- [x] API endpoint GET /api/v1/releases for paginated release list
- [x] Allow user to view any historical release notes
- [x] Display current version prominently
