# /release-notes — Draft and Commit Release Notes

Collect commits since the last git tag, categorise them, write a SQL migration, commit, and tag.

## Steps

1. **Determine the new version**
   - Run `git tag --sort=-version:refname | head -1` to find the latest tag (e.g. `v0.27.2`).
   - Decide the next version by applying semver rules to the commits since that tag:
     - Breaking change or major new capability → bump **minor** (project is pre-1.0; no major bumps).
     - New user-facing feature → bump **minor**.
     - Bug fix or small improvement only → bump **patch**.

2. **Collect commits since the last tag**
   - Run `git log <last-tag>..HEAD --oneline` to list every commit in scope.
   - Categorise each commit into one of three buckets using its prefix and content:
     - **Major** — `feat:` commits that introduce a significant user-facing capability.
     - **Minor** — `feat:` commits with smaller scope, `refactor:`, `chore:`, `docs:`, UI/UX improvements.
     - **Bugs** — `fix:` commits.
   - Ignore pure tooling or CI commits that users never see (e.g. `chore: update deps`).
   - Write each item as a single plain-English sentence describing the user-visible effect, not the code change.

3. **Find the next migration number**
   - Run `ls backend/deploy-scripts/migrations/ | sort | tail -1` from the repo root to find the highest-numbered migration file.
   - Increment that number by 1 and zero-pad to 3 digits (e.g. `110` → `111`).

4. **Write the SQL migration file**
   - Create `backend/deploy-scripts/migrations/NNN_add_release_X.Y.Z.sql` (where `NNN` is the number from step 3 and `X.Y.Z` is the version without the `v` prefix).
   - Use this exact structure (note SQL single-quote escaping with `''`):

   ```sql
   -- Migration: Add Release X.Y.Z
   -- Description: Adds release notes for version X.Y.Z

   INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
   ('X.Y.Z', 'YYYY-MM-DD', '## What''s New in vX.Y.Z

   ### Major
   - ...

   ### Minor
   - ...

   ### Bugs
   - ...', CURRENT_TIMESTAMP)
   ON CONFLICT (version) DO UPDATE SET
     release_date = EXCLUDED.release_date,
     notes = EXCLUDED.notes;
   ```

   - `release_date` is today's date in `YYYY-MM-DD` format.
   - Omit any section (`### Major`, `### Minor`, `### Bugs`) that has no entries.
   - Any single quote inside the notes text must be escaped as `''` (two single quotes).

5. **Review with the human**
   - Show the full migration file content and the categorised commit list.
   - Wait for explicit approval before proceeding.

6. **Commit and tag**
   - Stage the migration: `git add backend/deploy-scripts/migrations/NNN_add_release_X.Y.Z.sql`
   - Commit: `git commit -m "chore: add release notes for vX.Y.Z"`
   - Tag: `git tag vX.Y.Z`

## Notes

- Never bump the major version — the project uses `v0.x.y` and major version stays at `0` until explicitly decided otherwise.
- The `version` field in the SQL does **not** include the `v` prefix (e.g. `'0.27.2'`, not `'v0.27.2'`).
- The git tag **does** include the `v` prefix (e.g. `v0.27.2`).
- Do not push the tag automatically — leave that for the `/pr` workflow.
- If a section would only contain internal/tooling changes invisible to users, omit it entirely rather than padding with low-value entries.
