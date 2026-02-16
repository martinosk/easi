# Error Context Wrapping: High-Value Operational Paths

## Description
Improve debuggability by wrapping returned errors with operation and entity context in high-impact runtime paths.

## Why
Bare `return err` from projectors/read models/handlers hides where failures occurred and slows incident response.

## Scope (This Spec)
- Priority 1: projectors in capabilitymapping, architecturemodeling, enterprisearchitecture
- Priority 2: importing handlers and repositories used by asynchronous import flow
- Priority 3: read model mutation paths touched by specs 140/143

## Out of Scope
- Mechanical whole-repo rewrite in one PR
- Rewording already-clear wrapped errors

## Required Changes

### 1) Wrap returned errors on boundary crossings
When returning from DB calls, event deserialization, command dispatch, or repository calls, wrap with `%w` and concise operation context.

### 2) Include identifying context
Add IDs/event type when available (command ID, aggregate ID, tenant ID, event type).

### 3) Keep chains inspectable
Do not replace errors with strings; preserve original cause for `errors.Is` and `errors.As`.

## Error Message Rules
- lower-case messages
- operation-first phrasing (`load import session %s: %w`)
- no duplicate stack/context spam

## Acceptance Criteria
- Newly touched files in specs 140/141/143 contain no bare `return err` for boundary-crossing operations.
- Projector failures include event or aggregate context.
- Import handler path failures include session/command context.
- Tests continue to pass.

## Verification
- `go test ./internal/...`
- Review changed files for `%w` wrapping and context fields.
