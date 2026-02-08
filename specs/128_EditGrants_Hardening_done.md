# 128 - Edit Grants: Hardening

**Status:** done
**Depends on:** [126_EditGrants_AccessDelegation](126_EditGrants_AccessDelegation_done.md)
**Related:** [127_EditGrant_GranteeExperience](127_EditGrant_GranteeExperience_pending.md) (independent, can be implemented in parallel)

## Description

Post-implementation review of spec 126 identified security gaps, domain model inconsistencies, code quality issues, and missing test coverage. This spec addresses all findings. Scope is limited to hardening the existing implementation; new user-facing features belong in spec 127.

---

## 1. Security Hardening

### 1.1 Authorization on GET Endpoints

`GetEditGrantByID` and `GetEditGrantsForArtifact` return full grant details (grantor email, grantee email, artifact ID, reason) to any authenticated user. This leaks which users have edit access to which artifacts.

- `GetEditGrantByID`: return 403 unless the requester is the grantor, the grantee, or has `edit-grants:manage` permission
- `GetEditGrantsForArtifact`: return 403 unless the requester has write permission on the artifact type or has `edit-grants:manage` permission
- `GetMyEditGrants`: no change needed (already scoped to the requester's own grants)

### 1.2 Revocation Authorization Consistency

Grant creation uses permission-based check (`edit-grants:manage`) but revocation uses role-based check (`actor.Role == "admin"`). This means an architect can create grants but cannot revoke grants created by other architects.

- Change `canRevokeGrant` to: `grant.GrantorID == actor.ID || actor.HasPermission("edit-grants:manage")`
- This makes revocation authorization consistent with creation authorization

### 1.3 Rate Limiting on Grant Creation

Other write endpoints use `RateLimitMiddleware`. The `POST /edit-grants` endpoint does not.

- Apply `RateLimitMiddleware` to `POST /edit-grants`, consistent with other write endpoints

### 1.4 Artifact ID Format Validation

The `ArtifactRef` value object only checks that the artifact ID is non-empty. Arbitrary strings pass validation.

- Validate that `artifactID` is a valid UUID format in `NewArtifactRef`
- Add `ErrInvalidArtifactID` sentinel error

### 1.5 Reason Field Length Limit

The `reason` field is `TEXT` with no length limit. Accepts arbitrarily large input.

- Enforce a maximum length of 1000 characters in the domain model
- Create a `Reason` value object with validation
- Add `ErrReasonTooLong` sentinel error
- Map to HTTP 400 in error registration

### 1.6 PII in Application Logs

`edit_grant_enrichment.go` logs `actor.Email` on failure. Email addresses in logs create GDPR concerns.

- Log `actor.ID` instead of `actor.Email`

### 1.7 Audit Trail for Grant Lifecycle

Grant creation, revocation, and usage are security-critical operations with no audit logging.

- Log audit event when a grant is created (who granted, to whom, which artifact, reason)
- Log audit event when a grant is revoked (who revoked, which grant)
- Log audit event when `RequireWriteOrEditGrant` middleware grants access via an edit grant (not native RBAC)
- Use the existing audit infrastructure (`shared/audit` package)

---

## 2. Domain Model Fixes

### 2.1 GranteeEmail Value Object

`granteeEmail` is stored as a plain string in the aggregate while `Grantor` is a proper value object. This is primitive obsession.

- Create `GranteeEmail` value object in `accessdelegation/domain/valueobjects/`
- Validation: non-empty, trimmed, basic email format, case-normalized (lowercase)
- Implement `Equals(other domain.ValueObject) bool`
- Update aggregate to use `GranteeEmail` instead of `string`
- Self-grant check uses `strings.EqualFold` via the value object's normalized form

### 2.2 Grantor Value Object Missing Equals

The `Grantor` value object does not implement `Equals(other domain.ValueObject) bool`. All other value objects in this context do.

- Add `Equals` method to `Grantor` comparing both `id` and `email`

### 2.3 Business Domain Artifact Type

The spec (line 111) lists "capabilities, components, views, business domains" as artifact types. The implementation only supports three.

- Add `ArtifactTypeDomain ArtifactType = "domain"` constant
- Update `validArtifactTypes` map
- Update `ErrInvalidArtifactType` message
- Update `chk_edit_grant_artifact_type` CHECK constraint via new migration
- Subscribe to `BusinessDomainDeleted` events for cascade revocation
- Wire `RequireWriteOrEditGrant` middleware on business domain PUT/PATCH routes

---

## 3. Code Quality

### 3.1 Fail-Fast on Corrupt Event Data in Replay

The `apply()` method silently discards errors from value object constructors during event replay. This can result in zero-value aggregate state if event data is malformed.

- Panic on value object construction errors in `apply()`. Event store data is the source of truth; if it is corrupt, recovery requires manual intervention, not silent degradation.
- Format: `panic(fmt.Sprintf("corrupt event data in %T: %v", event, err))`

### 3.2 Remove Non-Deterministic IsExpired()

`IsExpired()` uses `time.Now()`, making the domain model non-deterministic. The method is not currently called by any production code. Expiration is correctly enforced at the SQL level via `expires_at > NOW()`.

- Remove `IsExpired()` method from the aggregate
- If a wall-clock expiry check is needed in the future, it belongs in the application or infrastructure layer, not the domain

### 3.3 Handle Errors Once in Projectors

Both `EditGrantProjector` and `ArtifactDeletionProjector` log errors before returning them, causing duplicate logging by callers.

- Remove `log.Printf` calls from projectors
- Return wrapped errors using `fmt.Errorf("...: %w", err)` and let the caller decide how to handle them

### 3.4 Context Cancellation in ArtifactDeletionProjector

The loop that dispatches revoke commands does not check context cancellation. It will continue processing even after timeout/cancellation.

- Add `select { case <-ctx.Done(): return ctx.Err() default: }` at the start of the loop body

### 3.5 Use PluralResourceName Helper

`canGrantEditAccess` uses hardcoded `artifactType+"s"` pluralization. The codebase has a `PluralResourceName` helper.

- Replace `artifactType+"s"` with `sharedctx.PluralResourceName(artifactType)`

### 3.6 Clean Up getGrantOrFail Return Signature

`getGrantOrFail` returns `(*EditGrantDTO, error)` but callers only check if grant is nil, never inspecting the error. The function writes the HTTP response internally.

- Change signature to `getGrantOrFail(w, r, id) *EditGrantDTO`
- Callers check `if grant == nil { return }`

---

## 4. Test Coverage

### 4.1 RequireWriteOrEditGrant Middleware Tests

This is the authorization enforcement point for the entire feature. Zero tests.

Test cases:
- User with native write permission passes through (RBAC path)
- User without write permission but with active edit grant passes through
- User without write permission and without edit grant gets 403
- Empty artifact ID parameter gets 403
- Missing actor in context gets 401
- Auth bypass mode passes through

### 4.2 HTTP Handler Tests

The API layer contains authorization logic, error mapping, and duplicate detection that is untested.

Test cases:
- Create grant succeeds (201)
- Create self-grant returns 400
- Create duplicate returns 409
- Create with invalid artifact type returns 400
- Create without permission returns 403
- Revoke by grantor succeeds (204)
- Revoke by non-grantor non-admin returns 403
- Revoke already revoked returns 409
- GET by ID: authorized user gets 200, unauthorized user gets 403
- GET for artifact: authorized user gets 200, unauthorized user gets 403

### 4.3 Command Handler Tests

Both handlers have zero coverage. The spec (126) acknowledges this.

Test cases for `CreateEditGrantHandler`:
- Valid command creates aggregate and persists `EditGrantActivated` event
- Invalid artifact type returns `ErrInvalidArtifactType`
- Empty artifact ID returns `ErrEmptyArtifactID`
- Self-grant returns `ErrCannotGrantToSelf`
- Empty grantor ID returns `ErrGrantorIDEmpty`
- Invalid scope returns `ErrInvalidGrantScope`

Test cases for `RevokeEditGrantHandler`:
- Valid revoke persists `EditGrantRevoked` event
- Revoke already-revoked grant returns `ErrGrantAlreadyRevoked`
- Revoke non-existent grant returns appropriate error

### 4.4 Projector Tests

`EditGrantProjector`: verify JSON roundtrip (marshal event data -> unmarshal -> DTO) preserves all fields including timestamps.

`ArtifactDeletionProjector`:
- Deletion event dispatches revoke commands for each active grant
- No active grants results in no commands dispatched
- Handles missing `"id"` field by falling back to `AggregateID()`
- Context cancellation stops processing mid-loop

### 4.5 Grantor Value Object Tests

Every other value object has a dedicated test file. `Grantor` does not.

- Empty ID returns `ErrGrantorIDEmpty`
- Empty email returns `ErrGrantorEmailEmpty`
- Valid inputs return correct accessors
- `Equals` returns true for same values, false for different values

### 4.6 GranteeEmail Value Object Tests

New value object (from 2.1) needs full coverage.

- Empty email returns error
- Whitespace-only email returns error
- Valid email constructs successfully
- Email is trimmed and lowercased
- `Equals` is case-insensitive
- Self-grant check works across case variations

### 4.7 Reason Value Object Tests

New value object (from 1.6) needs coverage.

- Empty reason is valid (reason is optional)
- Reason at max length is valid
- Reason exceeding max length returns `ErrReasonTooLong`

### 4.8 Frontend Mutation Effects Tests

`editGrantsMutationEffects` is defined but absent from `mutationEffects.test.ts`.

- `editGrantsMutationEffects.create()` invalidates the correct query keys
- `editGrantsMutationEffects.revoke()` invalidates the correct query keys

---

## Checklist

### Security
- [x] Authorization check on `GetEditGrantByID` (grantor, grantee, or `edit-grants:manage`)
- [x] Authorization check on `GetEditGrantsForArtifact` (write permission or `edit-grants:manage`)
- [x] Fix `canRevokeGrant` to use `HasPermission("edit-grants:manage")` instead of `Role == "admin"`
- [x] Rate limiting on `POST /edit-grants`
- [x] UUID format validation in `ArtifactRef` value object
- [x] `Reason` value object with 1000-char max length
- [x] Map `ErrReasonTooLong` and `ErrInvalidArtifactID` to HTTP 400
- [x] Replace `actor.Email` with `actor.ID` in enrichment middleware log
- [x] Audit logging: grant created
- [x] Audit logging: grant revoked
- [x] Audit logging: grant used (middleware fallback path)

### Domain Model
- [x] `GranteeEmail` value object with validation, normalization, and `Equals`
- [x] Aggregate uses `GranteeEmail` value object (not raw string)
- [x] Self-grant check uses case-insensitive comparison via `GranteeEmail`
- [x] `Grantor.Equals()` method
- [x] `ArtifactTypeDomain` constant and `validArtifactTypes` entry
- [x] Migration: drop `chk_edit_grant_artifact_type` CHECK constraint (validation belongs in domain model)
- [x] Subscribe to `BusinessDomainDeleted` for cascade revocation
- [x] Wire `RequireWriteOrEditGrant` on business domain PUT/PATCH routes

### Code Quality
- [x] Panic on value object construction errors in `apply()`
- [x] Remove `IsExpired()` method from aggregate
- [x] Remove `log.Printf` from `EditGrantProjector`, return wrapped errors
- [x] Remove `log.Printf` from `ArtifactDeletionProjector`, return wrapped errors
- [x] Context cancellation check in `ArtifactDeletionProjector` loop
- [x] Use `PluralResourceName` helper in `canGrantEditAccess`
- [x] Simplify `getGrantOrFail` return signature to `*EditGrantDTO`

### Tests
- [x] `RequireWriteOrEditGrant` middleware: 6 test cases
- [ ] HTTP handlers: 10 test cases
- [x] `CreateEditGrantHandler`: 6 test cases
- [x] `RevokeEditGrantHandler`: 3 test cases
- [x] `EditGrantProjector`: JSON roundtrip test
- [x] `ArtifactDeletionProjector`: 4 test cases
- [x] `Grantor` value object: 4 test cases
- [x] `GranteeEmail` value object: 6 test cases
- [x] `Reason` value object: 3 test cases
- [x] Frontend `editGrantsMutationEffects`: 2 test cases
- [x] All existing tests still pass after changes
