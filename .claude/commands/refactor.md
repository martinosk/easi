Refactor a file using CodeScene MCP to achieve a code health score of 10.0.

**Usage**: `/refactor <file-path>`

**Arguments**:
- `file-path`: Path to the file to analyze and refactor. If not provided, prompt the user.

## Steps

1. **Check code health score**
   Use the `mcp__codescene__code_health_score` tool to get the current code health score for the file.
   Report the score to the user.

2. **Evaluate score**
   - If the score is **10.0**: report that the file is already at maximum code health. Stop here.
   - If the score is **below 10.0**: continue to step 3.

3. **Get review recommendations**
   Use the `mcp__codescene__code_health_review` tool to get detailed review recommendations for the file.
   Summarize the findings to the user, listing each issue and its impact on the score.

4. **Refactor the file**
   Apply the recommended refactorings one at a time:
   - Read the file first to understand the full context.
   - Address each recommendation from the review, prioritizing by impact.
   - Use the Edit tool to make targeted changes. Do not rewrite the file wholesale.
   - Follow the existing code style and conventions.
   - Do not add comments unless the review specifically recommends improving clarity.

5. **Re-check code health score**
   Use `mcp__codescene__code_health_score` again on the refactored file.
   - If the score is **10.0**: report success and stop.
   - If the score is **still below 10.0**: get a new review with `mcp__codescene__code_health_review`, apply further refactorings, and re-check. Repeat up to 3 iterations total.

6. **Verify build and tests**
   If backend files has been modified: Verify that backend builds and all tests are passing.
   If fronted files has been modified: Verify that frontend builds with `npm run build` and that all tests are passing.

7. **Final report**
   Display a summary:
   ```
   Refactoring complete: <file-path>

   Before: <original-score>
   After:  <final-score>

   Changes applied:
   - <brief description of each change made>
   ```

   If the score is still below 10.0 after 3 iterations, note the remaining issues and suggest what the user can address manually.

## Rules

- Preserve all existing behavior — refactoring only, no functional changes.
- Respect the project's code style (no added comments, no over-engineering).
- If a recommendation conflicts with project conventions (e.g., suggests adding comments), skip it and note why.

# Refactoring hints

## Fixing "Excess Number of Function Arguments"

Do NOT fix it by simply bundling arguments into a struct to reduce the count. Instead, investigate the root cause:

1. **Low cohesion / too many responsibilities**: The class or function is doing too much. Split responsibilities into separate, cohesive types.
2. **Missing domain abstraction**: There is a coherent concept hiding behind the arguments that deserves its own type. Only introduce such a type if it genuinely encapsulates something meaningful — e.g., a "result" that bundles an ID + status code + location, or a domain value object. Do NOT create a struct that is just a bag of unrelated arguments passed together for convenience.
