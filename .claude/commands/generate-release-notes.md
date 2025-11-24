Generate release notes for the upcoming release by analyzing changes since the last release.

## Steps

1. **Find the last release tag**
   ```bash
   git tag --sort=-version:refname | head -1
   ```
   If no tags exist, use the first commit as the baseline.

2. **Get commits since last release**
   ```bash
   git log <last-tag>..HEAD --oneline --no-merges
   ```

3. **Find completed specs since last release**
   Look for spec files in `specs/` with `_done.md` suffix that were modified since the last release:
   ```bash
   git diff --name-only <last-tag>..HEAD -- specs/*_done.md
   ```

4. **Analyze and categorize changes**

   Read each completed spec and categorize the changes:
   - **Major**: New features, significant enhancements
   - **Bugs**: Bug fixes, corrections
   - **API**: New or changed API endpoints

   Filter out technical items that users don't care about:
   - Refactoring
   - Internal code improvements
   - Test additions
   - Documentation updates
   - Build/CI changes

5. **Generate the release notes**

   Create markdown content structured as:
   ```markdown
   ## What's New in X.Y.Z

   ### Major
   - Feature description (from spec)

   ### Bugs
   - Bug fix description

   ### API
   - New endpoint: `GET /api/v1/resource`
   ```

6. **Output the draft**

   Display the generated release notes for developer review. The notes should be:
   - User-focused (what they can do, not how it was implemented)
   - Concise but descriptive
   - Written in past tense ("Added", "Fixed", "Improved")

## Example Output

```markdown
## What's New in 1.2.0

### Major
- Added capability hierarchy visualization for better understanding of business capabilities
- Introduced release notes system to keep you informed about updates

### Bugs
- Fixed component position not saving correctly when dragging

### API
- New endpoint: `GET /api/v1/releases/latest` for retrieving release notes
```

The developer should review and edit these notes before using `/tag-release` to finalize.
