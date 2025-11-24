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

4. **Store release notes in database**
   Use curl to POST the release notes to the API:
   ```bash
   curl -X POST http://localhost:8080/api/v1/releases \
     -H "Content-Type: application/json" \
     -d '{
       "version": "<version>",
       "releaseDate": "<today ISO date>",
       "notes": "<markdown content>"
     }'
   ```

5. **Update APP_VERSION in configuration**
   Look for where APP_VERSION is configured and update it. Check:
   - `.env` file
   - `docker-compose.yml` or similar
   - Any deploy configuration files

   If no explicit version file exists, inform the user they need to set `APP_VERSION=<version>` in their deployment.

6. **Create git tag**
   ```bash
   git tag -a v<version> -m "Release v<version>"
   ```

7. **Commit any version changes**
   If files were modified in step 5:
   ```bash
   git add <modified files>
   git commit -m "chore: bump version to <version>

   🤖 Generated with [Claude Code](https://claude.com/claude-code)

   Co-Authored-By: Claude <noreply@anthropic.com>"
   ```

8. **Summary**
   Display a summary of what was done:
   ```
   ✅ Release v<version> created successfully!

   - Release notes stored in database
   - Git tag v<version> created
   - Version configuration updated (if applicable)

   Next steps:
   - Review the changes: git log -1 && git tag -l 'v*'
   - Push the tag: git push origin v<version>
   - Push commits: git push origin main
   ```

## Error Handling

- If the version already exists as a tag, warn and ask for confirmation to overwrite
- If the API call fails, provide instructions for manual storage
- If git operations fail, provide rollback instructions

## Example

```
User: /tag-release 1.2.0
Claude: Creating release v1.2.0...

Please provide the release notes content (you can use markdown), or paste the output from /generate-release-notes:

User: ## What's New in 1.2.0
### Major
- Added release notes system

Claude: ✅ Release v1.2.0 created successfully!
...
```
