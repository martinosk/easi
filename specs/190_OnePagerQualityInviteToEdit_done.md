# 190 — One-Pager Quality: Invite to edit

> **Status:** done
> **Depends on:** [189 — One-Pager Quality master list](189_OnePagerQualityList_pending.md), [126 — EditGrants / AccessDelegation](126_EditGrants_AccessDelegation_done.md)

---

## Problem Statement

Spec 189 surfaces every incomplete or stale one-pager in one ranked list, but the architect
looking at it cannot act — to get a gap fixed they must leave the list, navigate to the
subject, and open its invite dialog somewhere else. The person who should fill the missing
facts is often not even an EASI user yet.

This slice closes the loop: a per-row **"Invite to edit"** action on the quality list that
delegates edit access for that subject straight from the row, reusing the existing edit-grant
mechanism (specs 126–128) — including its ability to auto-onboard a non-user by email. It adds
no new grant, notification, or email system; it wires the list to what already exists.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | From an incomplete row, invite the subject's owner — user or not — to fill the missing facts, without leaving the list |
| **Invited owner** | Receive edit access (and, if not yet a user, an onboarding invitation) so they can complete the one-pager |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Invite to edit from the One-Pager Quality list

  Scenario: Inviting an existing user to edit a subject
    Given I am viewing the One-Pager Quality list
    And I may grant edit access for the Application "Billing"
    When I use "Invite to edit" on the "Billing" row and enter a colleague's email
    Then an edit grant is created for "Billing" for that email via the existing mechanism

  Scenario: Inviting a non-user auto-raises an onboarding invitation
    Given the "Billing" row and an email that belongs to no EASI user
    When I invite that email to edit "Billing"
    Then an edit grant is created for that email
    And an onboarding invitation is raised for them through the existing edit-grant flow

  Scenario: The action is gated on permission to grant edit access
    Given a row whose subject I may not grant edit access for
    Then the row shows no "Invite to edit" action

  Scenario: The action is unavailable for Enterprise Capability rows
    Given an Enterprise Capability row
    Then it shows no "Invite to edit" action

  Scenario Outline: Each supported subject type maps to its edit-grant artifact type
    Given a "<subject type>" row I may grant edit access for
    When I invite an email to edit it
    Then the edit grant is created with artifact type "<artifact type>"

    Examples:
      | subject type    | artifact type   |
      | Capability      | capability      |
      | Application     | component       |
      | Acquired Entity | acquired_entity |
      | Vendor          | vendor          |
      | Internal Team   | internal_team   |
```

---

## Business Rules & Invariants

1. **Reuse, no new mechanism** — the action creates an EditGrant through the existing
   `POST /api/v1/edit-grants` (126); no new grant, notification, or email path is introduced.
2. **Grantee by email; non-user onboarding is automatic** — the invitee is identified by
   email; a non-user grantee raises `EditGrantForNonUserCreated`, which the existing projector
   turns into an auth onboarding invitation. This slice adds none of that plumbing.
3. **Subject-type → artifact-type mapping** — Capability → `capability`, Application →
   `component`, Acquired Entity → `acquired_entity`, Vendor → `vendor`, Internal Team →
   `internal_team`.
4. **Enterprise Capability is unsupported here** — the edit-grant `ArtifactType` enum has no
   `enterprise_capability` value; Enterprise Capability rows carry no invite affordance.
   Extending edit grants to Enterprise Capability is a separate numbered spec that would add
   the artifact type across `accessdelegation`.
5. **HATEOAS gating** — a row carries the invite link only when its subject type is supported
   (rule 3) *and* the caller may grant edit access for it — the existing grantor gate (write
   permission on the mapped artifact type, or `edit-grants:manage`). The frontend renders the
   action only when the link is present.
6. **Tenant isolation** — inherited from the 189 list and the existing edit-grant handler's
   tenancy; a grant is only ever created within the caller's tenant.

---

## Acceptance Criteria

- [x] Each 189 list row carries an `x-edit-grants` link when, and only when, the subject type
      is supported (rule 3) and the caller passes the existing edit-grant grantor gate.
- [x] The row action opens the existing invite-to-edit dialog prefilled with the subject's
      mapped artifact type and ID, and creating the grant calls the existing
      `POST /api/v1/edit-grants` — no new endpoint.
- [x] Inviting an email that belongs to no user results in the existing onboarding invitation
      via `EditGrantForNonUserCreated`; no new notification or email code is added.
- [x] Each of the five supported subject types maps to its correct artifact type; Enterprise
      Capability rows expose no invite affordance.
- [x] The action never appears on a row the caller may not grant edit access for.
- [x] Every BDD scenario above has a corresponding automated test.

---

## Architecture

### Ownership

The affordance is a thin composition over two existing pieces: the 189 list (row surface) and
`accessdelegation`'s edit-grant creation (126). The subject-type → artifact-type mapping and
the per-row link gating live where the 189 rows are serialized (composition root); the UI
reuses the existing edit-grant components. No bounded context gains new domain behavior.

### Domain Model

None new. The action produces the existing `EditGrant` aggregate's events
(`EditGrantActivated`, and `EditGrantForNonUserCreated` for non-users) unchanged.

### API Surface

No new endpoint. The 189 list response rows gain an `x-edit-grants` link (the rel the existing
`InviteToEditButton` consumes), gated per rule 5. Grant creation reuses
`POST /api/v1/edit-grants` with `{granteeEmail, artifactType, artifactId, reason, scope:
'write'}`, where `artifactType` is the mapped value (rule 3).

### Persistence

None. No new tables or migrations; edit grants persist through the existing
`accessdelegation` store.

### Frontend

Each 189 row renders the existing `InviteToEditButton` gated on the row's `x-edit-grants`
link; clicking opens the existing `InviteToEditDialog` prefilled with the subject's artifact
type and ID; submission uses the existing `useCreateEditGrant` hook. Enterprise Capability
rows render no button (no link).

### Cross-Context Integration

`onepagers` list row → `accessdelegation` edit-grant creation via the existing endpoint. No
new events; the non-user onboarding path (`EditGrantForNonUserCreated` → auth invitation) is
unchanged.

---

## Design Decisions

1. **Reuse the edit-grant mechanism; add no notification system** — the invite *is* an edit
   grant, and non-user onboarding already flows through `EditGrantForNonUserCreated` → auth.
   Alternative: a bespoke notification/email for the list (rejected — none exists; it would
   duplicate the onboarding path).
2. **Map subject type → artifact type and scope Enterprise Capability out as a business rule**
   — the `ArtifactType` enum has no `enterprise_capability`; adding it is a cross-context
   change to `accessdelegation`'s domain (126–128) that would balloon this slice, and HATEOAS
   gating makes the absence clean and forward-compatible (no link → no button). Alternative:
   add the `enterprise_capability` artifact type now (rejected — expands scope into another
   bounded context; a separate numbered spec if Enterprise Capability edit grants are wanted).
3. **Per-row link generated on the 189 response, gated by the existing grantor permission** —
   the affordance is HATEOAS-driven, reusing `InviteToEditButton` / `InviteToEditDialog`.
   Alternative: a client-side permission/validity table (rejected — the UI must be driven by
   backend links, not a duplicated client-side rules table).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Enterprise Capability excluded by business rule | EC subjects cannot be delegated from the list | HATEOAS gating hides the action cleanly; a later numbered spec can add the artifact type end-to-end |
| Row link reuses the grantor permission gate | The list response must resolve the grant permission per row | One check per row against the caller's already-loaded permissions; no extra query |
| Reusing the edit-grant dialog verbatim | The action inherits the edit-grant flow's fields (reason, scope) | Consistent with every other invite-to-edit entry point; nothing bespoke to maintain |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off
