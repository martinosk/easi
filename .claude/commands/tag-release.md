Tag a new release with the specified version number.

**Usage**: `/tag-release [version]`

**Arguments**:
- `version`: Semantic version number (e.g., 1.2.0). If not provided, prompt the user.

## Steps

1. **Validate version format**
   Ensure the version follows semver format: `X.Y.Z` where X, Y, Z are numbers.

2. **Check for uncommitted changes**
   ```bash
   git status --porcelain
   ```
   If there are uncommitted changes, warn the user and ask if they want to proceed.

3. **Ask for release notes**
   If the user hasn't run `/generate-release-notes` recently, ask them to provide release notes markdown.

   Example prompt: "Please provide the release notes content (you can use markdown):"

4. **Create migration to add release notes**
   Find the next migration number and create a new migration file:
   ```bash
   ls backend/deploy-scripts/migrations/*.sql | sort | tail -1
   ```

   Create `backend/deploy-scripts/migrations/XXX_add_release_<version>.sql`:
   ```sql
   -- Migration: Add Release <version>
   -- Description: Adds release notes for version <version>

   INSERT INTO releases (version, release_date, notes, created_at) VALUES
   ('<version>', '<today YYYY-MM-DD>', '<markdown content with escaped quotes>', CURRENT_TIMESTAMP)
   ON CONFLICT (version) DO UPDATE SET
     release_date = EXCLUDED.release_date,
     notes = EXCLUDED.notes;
   ```

5. **Commit migration**
   ```bash
   git add backend/deploy-scripts/migrations/
   git commit -m "chore: add release notes for v<version>

   🤖 Generated with [Claude Code](https://claude.com/claude-code)

   Co-Authored-By: Claude <noreply@anthropic.com>"
   ```

6. **Create git tag**
   ```bash
   git tag -a v<version> -m "Release v<version>"
   ```

7. **Summary**
   Display a summary of what was done:
   ```
   ✅ Release v<version> created successfully!

   - Migration created: backend/deploy-scripts/migrations/XXX_add_release_<version>.sql
   - Git tag v<version> created
   - APP_VERSION will be set automatically from git tag at build time

   Next steps:
   - Review the changes: git log -1 && git diff HEAD~1
   - Push commits: git push origin main
   - Push the tag: git push origin v<version>
   - Pipeline will inject version from tag and run migrations
   ```

## Error Handling

- If the version already exists as a tag, warn and ask for confirmation to overwrite
- If git operations fail, provide rollback instructions
- Remember to escape single quotes in SQL by doubling them ('')

## Example

```
User: /tag-release 0.8.0
Claude: Creating release v0.8.0...

Please provide the release notes content (you can use markdown), or paste the output from /generate-release-notes:

User: ## What's New in 0.8.0
### Major
- Added new dashboard feature

Claude: ✅ Release v0.8.0 created successfully!
- Migration created: backend/deploy-scripts/migrations/017_add_release_0.8.0.sql
- Git tag v0.8.0 created
- APP_VERSION updated to 0.8.0
...
```
