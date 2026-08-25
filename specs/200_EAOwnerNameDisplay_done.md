# 200 — EA Owner Name Display

> **Status:** done
> **Depends on:** 136/137 (ACL cache pattern), 139 (schema ownership)

---

## Problem Statement

The capability edit form and the import metadata prefill store the selected EA owner as the user's id (a GUID). The capability details panel and the capability drawer render the stored value verbatim, so users see a GUID instead of the EA owner's name. Spec 116's scenarios already describe the expected observable behavior ("capabilities have 'Alice Smith' set as their EA Owner"), which the current display contradicts.

The domain model contributes to the mess: `eaOwner` reuses the free-text `Owner` value object even though it now semantically holds a user reference. Older capabilities may hold free-text names (spec 036 defined the field as text), so stored values are a mix of user ids and legacy text, and nothing validates new writes.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Stakeholder** | See who owns a capability, by name, without users:read permission |
| **Architect / Admin** | See and assign EA owners by name |
| **Assistant (agent tools)** | Set an EA owner by name without knowing user ids |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: EA owner shown by name and stored as a user reference

  Scenario: Capability with a user-id EA owner shows the user's name
    Given a capability whose stored EA owner is the id of user "Alice Smith"
    When I open the capability details panel or the capability drawer
    Then the EA Owner field shows "Alice Smith"

  Scenario: Legacy free-text EA owner still displays
    Given a capability whose stored EA owner is the text "Bob Jones"
    And no user exists with that id
    When I open the capability details panel
    Then the EA Owner field shows "Bob Jones"

  Scenario: Stakeholders see the resolved name
    Given I am signed in as a stakeholder without users:read permission
    And a capability whose stored EA owner is the id of user "Alice Smith"
    When I open the capability details panel
    Then the EA Owner field shows "Alice Smith"

  Scenario: Newly created users resolve without redeploy
    Given a new user "Carol Davis" accepts an invitation
    And an architect sets her as EA owner of a capability
    When anyone views that capability
    Then the EA Owner field shows "Carol Davis"

  Scenario: Setting EA owner by name resolves to the user reference
    Given exactly one user named "Alice Smith" exists
    When the metadata update endpoint receives eaOwner "Alice Smith"
    Then the capability's stored EA owner is Alice Smith's user id

  Scenario: Unresolvable EA owner text is rejected
    Given no user is named "Zaphod"
    When the metadata update endpoint receives eaOwner "Zaphod"
    Then the request fails with a validation error naming the eaOwner field

  Scenario: Ambiguous EA owner name is rejected
    Given two users named "Alex Kim" exist
    When the metadata update endpoint receives eaOwner "Alex Kim"
    Then the request fails with a validation error indicating the name is ambiguous

  Scenario: Saving a capability with unresolvable legacy text keeps it
    Given a capability whose stored EA owner is the legacy text "Nobody Current"
    And no user matches that text
    When a metadata update is saved with eaOwner unchanged
    Then the update succeeds and the stored EA owner remains "Nobody Current"

  Scenario: Legacy capability converges on next edit
    Given a capability whose stored EA owner is the legacy text "Alice Smith"
    And exactly one user named "Alice Smith" exists
    When any metadata update is saved for that capability with eaOwner unchanged
    Then the stored EA owner becomes Alice Smith's user id
```

---

## Business Rules & Invariants

1. **Server-side resolution** — EA owner id→name resolution happens in CapabilityMapping read models; clients never need users:read to display an EA owner.
2. **BC isolation preserved** — CapabilityMapping runtime SQL touches only `capabilitymapping` schema tables; user names come from a CM-owned cache, never from `auth.users` at runtime.
3. **Graceful display fallback** — when the stored EA owner does not match a cached user id, the raw stored value is returned unchanged.
4. **EA owner is a reference** — after this change, every newly stored EA owner is either empty or a user id. The command boundary accepts a value that matches exactly one cached user by id, name, or email; anything else — including a well-formed id of a user that is not in the cache — is rejected with a validation error.
5. **History is append-only** — existing events are never rewritten; rehydration accepts historical free-text values. Validation applies only at the command boundary.
6. **Unchanged values are never rejected** — resubmitting a capability's current EA owner value succeeds even when it is unresolvable or ambiguous, so saving an untouched edit form cannot fail. Infrastructure failures during resolution are still surfaced as errors.

---

## Acceptance Criteria

- [x] Capability GET/list responses include a resolved EA owner display name whenever `eaOwner` is set.
- [x] Capability details panel and capability drawer render the resolved name, never a GUID, for owners selected via the dropdown.
- [x] A capability with a legacy free-text EA owner renders that text unchanged.
- [x] A stakeholder session renders the resolved name (no dependency on `/api/v1/users`).
- [x] A user created after deployment resolves via the event-driven cache without manual intervention.
- [x] A migration backfills the cache for all existing users, so pre-existing capabilities resolve immediately.
- [x] Metadata updates with an eaOwner user id are stored as-is; with a uniquely matching name/email are stored as the resolved id; with unresolvable or ambiguous text are rejected with 400.
- [x] The assistant's update_capability_metadata tool documents that eaOwner accepts a user name or id.
- [x] Architecture guard tests pass (no cross-schema SQL from CM read models).

---

## Architecture

### Ownership

CapabilityMapping owns the change. Auth is affected only as an upstream event publisher (already publishes `UserCreated` with `id` and `name`).

### Domain Model

New `EAOwner` value object replaces the reuse of `Owner` for `eaOwner` in `CapabilityMetadata`: empty or a canonical user-id string. Construction from historical event payloads remains permissive; strict resolution happens in the command handler. No event schema changes.

### API Surface

Capability read DTOs gain an optional `eaOwnerName` field (omitted when `eaOwner` is empty). The metadata update contract narrows: `eaOwner` must be a user id or a resolvable unique name/email (400 otherwise). No new endpoints, no permission changes.

### Persistence

New CM-owned cache table `capabilitymapping.user_names` (tenant_id, user_id, name, email), RLS-enabled like other CM tables, following the `reference_name_cache` precedent from ArchitectureDirection. Backfilled from `auth.users` in the migration (cross-schema access is allowed in migrations, precedent: migration 126). Eventually consistent via events thereafter.

### Frontend

Capability details panel and capability drawer display `eaOwnerName`, falling back to `eaOwner`. The edit form keeps using the user id as the select value.

### Cross-Context Integration

CM subscribes to Auth's published-language `UserCreated` event and upserts (user id → name, email) into its cache, mirroring ArchitectureDirection's reference projector.

---

## Design Decisions

1. **Resolve names server-side in CM** — stakeholders lack `users:read`, so a frontend lookup via `/api/v1/users` would 403 for them. Alternatives considered: frontend resolution (rejected — permission gap); cross-schema `JOIN auth.users` at runtime (rejected — violates spec 139 BC isolation and the architecture guard tests).
2. **Keep storing the user id** — event history already holds ids; changing the select to store names would break edit-form preselection for existing data and revert on projection replay. Alternative: store the display name (rejected for those reasons).
3. **COALESCE fallback to the raw value** — legacy free-text owners and deleted/unknown ids degrade to the stored string instead of blank, matching rule 3.
4. **Accept name-or-id at the command boundary** — the assistant agent tool and any API client can pass a human name; the handler resolves it against the CM cache to keep the aggregate clean without breaking callers. Alternative: strict UUID-only input (rejected — breaks the assistant's update_capability_metadata tool, which documents eaOwner as a name).
5. **Organic convergence for legacy data, no corrective job** — legacy free-text rows display correctly via the fallback and are normalized to ids by rule 4 the next time they are edited. Alternative: a one-off corrective-event job at startup (rejected — requires new cross-tenant system-actor machinery with no user-visible benefit; history cannot be rewritten either way).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Event-driven name cache | Name changes in auth after user creation are not propagated (auth publishes no name-change event) | Names originate from OIDC at invitation acceptance and are effectively immutable today; a future auth event can update the cache |
| Extra cache table | One more projection to maintain | Follows the established ACL cache pattern (specs 136/137) |
| Organic legacy convergence | Dormant capabilities keep free-text values in the aggregate indefinitely | Display is unaffected (fallback); values converge on any subsequent edit |
| Single id/name/email lookup | A name or email equal to another user's id would be ambiguous | User ids are UUIDs; names and emails are never UUID-shaped in practice |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off
