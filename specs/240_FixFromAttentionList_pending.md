# 240 — Fix from the Attention List

> **Status:** pending
> **Depends on:** 230 (attention list), 236 (model hygiene checks), 241 (merge, for the duplicate-name shortcut)
> **Roadmap alignment:** `SD8 / H2-6` — slice F of [`docs/specs/data-quality.md`](../docs/specs/data-quality.md) (decision Q6)

---

## Problem Statement

The attention list (spec 230) tells a user what is wrong and lets them route it to someone else. To fix a finding themselves they must open the subject, find the control, change one field and come back, once per finding. A tenant with 185 applications of unknown hosting will not be cleaned up that way.

For findings whose fix is a single field, this slice puts the fix on the row, and lets the user select many rows of the same check and apply one value to all of them.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architect / Admin** | Work through a long list of simple gaps without leaving it |
| **Owner / Creator** | Fix their own application from the finding |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Fix from the attention list

  Scenario: Classifying hosting on the row
    Given a "hosting unclassified" item for "CRM" and the caller may classify its hosting
    When the caller selects "SaaS" on the row
    Then the hosting of "CRM" is "saas"
    And the row leaves the list

  Scenario: Assigning an owner on the row
    Given an "application without an owner" item for "CRM" and the caller may assign owners
    When the caller picks "Ole Berg" on the row
    Then "Ole Berg" owns "CRM" and the row leaves the list

  Scenario: Nominating when the caller may only nominate
    Given the caller may nominate but not assign
    When they pick "Ole Berg" on the row
    Then "Ole Berg" is nominated and the row is replaced by "ownership not confirmed"

  Scenario: Assigning a domain on the row
    Given a "capability outside every domain" item for the L1 "Tracking"
    When the caller picks the domain "Logistics"
    Then "Tracking" is assigned to "Logistics" and the row leaves the list

  Scenario: Setting an EA owner on the row
    Given a "capability without an EA owner" item for "Billing"
    When the caller picks "Alice Smith"
    Then "Alice Smith" is EA owner of "Billing" and the row leaves the list

  Scenario: Linking a capability on the row
    Given an "application realises nothing" item for "Portal"
    When the caller picks the capability "Customer Self Service"
    Then "Portal" realises it and the row leaves the list

  Scenario: Rows the caller may not fix offer no fix
    Given the caller may not classify the hosting of "CRM"
    Then the row for "CRM" offers no fix, and still offers "Invite to edit" where spec 230 allows

  Scenario: Fixing many at once
    Given the list is filtered to "hosting unclassified" with 50 rows shown
    When the caller selects 30 rows and applies "On-premises"
    Then each of the 30 applications is classified "on-premises"
    And a summary reports 30 fixed

  Scenario: Bulk fix reports partial failure
    Given 30 rows are selected and 2 of the applications were deleted meanwhile
    When the caller applies "Cloud"
    Then 28 are classified and the summary names the 2 that failed and why
    And the 2 rows are refreshed from the server

  Scenario: Selection needs one check
    Given the list is not filtered to a single check
    Then rows cannot be selected for a bulk fix
    And the list says to filter by a check to fix several at once

  Scenario: Only fixable rows can be selected
    Given the list is filtered to "hosting unclassified"
    Then only rows the caller may fix have a checkbox

  Scenario: Checks without a single-field fix
    Given a "duplicate name" item
    Then the row offers "Rename" and, for applications, "Merge", which open the subject's own dialogs
    Given a "realisation not assessed" item
    Then the row offers no inline fix
```

---

## Business Rules & Invariants

1. **Fixable checks and their fix** — exactly:

   | Check | Fix | Subject affordance |
   |-------|-----|--------------------|
   | `hosting-unclassified` | choose a hosting classification | `x-classify-hosting` |
   | `application-unowned` | choose a user or team | `x-assign-owner`, else `x-nominate-owner` |
   | `application-unrealised` | choose a capability | the realisation affordance of spec 232 |
   | `capability-outside-domain` | choose one or more domains | the domain assignment affordance |
   | `capability-without-ea-owner` | choose a user | the capability's edit affordance |

   `application-duplicate-name` and `capability-duplicate-name` offer shortcuts to rename and, for applications, to merge (spec 241); they are not inline or bulk fixes. Every other check has no fix on the row.
2. **The subject decides** — a fix is offered only when the subject's own resource carries the affordance in the table. The attention item never asserts that the caller may change the subject.
3. **A fix is the existing command** — one request per subject through that affordance. No new write endpoint, no bulk endpoint.
4. **Bulk is one check, one value, the loaded page** — bulk selection is available only when the list is filtered to one fixable check; it covers rows currently loaded; the same value is applied to each selected subject.
5. **Bulk applies independently** — subjects are processed with bounded concurrency; one failure neither stops nor rolls back the others. The outcome names every failure with its reason.
6. **The list is the truth after a fix** — after a fix the affected rows are refetched; a row disappears because the server no longer reports the finding, never because the client removed it.
7. **Bulk never applies to more than the user selected** — there is no "all matching" mode beyond the loaded rows.

---

## Acceptance Criteria

- [ ] Each fixable check renders its inline control only when the subject's affordance is present; affordances are fetched for the visible page in one request per subject collection, not one per row
- [ ] Each fix issues the existing command through the affordance's `href` and `method`; the frontend builds no URL
- [ ] `application-unowned` assigns when `x-assign-owner` is present and nominates otherwise
- [ ] Bulk selection is enabled only under rule 4; checkboxes appear only on fixable rows; select-all selects the fixable rows of the loaded page
- [ ] Bulk apply runs with a concurrency of at most 5, shows progress, and ends with a summary of fixed and failed subjects with reasons
- [ ] A 404 or 409 on one subject is reported for that subject and does not stop the rest
- [ ] After any fix the attention list, Home summary and the subject's own queries are invalidated; no row is removed optimistically
- [ ] Duplicate-name rows open the subject's rename dialog, and for applications the merge dialog when the subject carries `x-merge` (spec 241)
- [ ] `GET /api/v1/components` and `GET /api/v1/capabilities` accept an `ids` filter of up to 50 ids so the page's affordances are fetched in one request each
- [ ] Every BDD scenario has at least one corresponding test
- [ ] E2E: filter to hosting unclassified, select three applications, apply a classification and see them leave the list
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

Frontend: the attention list page of spec 230. Architecture Modeling and Capability Mapping gain an `ids` filter on their collections; their commands are unchanged. Stewardship is unchanged.

### Domain Model

None.

### API Surface

`ids` (comma-separated, at most 50) on the application and capability collections, returning the usual item shape with `_links`. More than 50 answers 400.

### Persistence

None.

### Frontend

- A fix registry mapping check key to an inline control and to the affordance it needs.
- Row controls built from Mantine primitives; a selection bar with the value control and "Apply to n".
- A bulk runner with bounded concurrency and a result summary.
- Mutation effects for each underlying command already exist; the attention and home query keys are added to them.

### Cross-Context Integration

None at runtime between contexts. The page reads attention items from Stewardship and affordances from the subject's owning context.

---

## Design Decisions

1. **Affordances come from the subject, not from the attention item** — the owning context alone knows who may change a subject, including edit-grant holders and ownership-state rules. Alternative: Stewardship adds fix links per item (rejected: it would re-implement other contexts' permission rules and drift).
2. **No bulk endpoint** — a bulk command per field per context is a large API surface for what is, per subject, an ordinary command; independent requests also give per-subject outcomes for free. Alternative: `POST /components/hosting-classifications` taking many ids (rejected for now by the cost-benefit; 50 requests at concurrency 5 complete in seconds).
3. **Bulk limited to the loaded page** — it keeps the selection something the user has seen. Applying a value to hundreds of unseen records invites exactly the invented data this family is meant to prevent.
4. **No optimistic removal** — the list is computed; removing rows client-side would show a state the server has not confirmed.
5. **Only single-field fixes** — TIME assessment, journeys and one-pager completion need context and judgement; they keep their own surfaces.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| One request per subject | More requests than a bulk endpoint | Bounded concurrency; page size 50 |
| Extra read for affordances | Two requests per page of the list | One per subject collection, cached by the existing query keys |
| Bulk owner assignment | Easy to assign one person to many applications wrongly | The summary lists what changed; ownership has its own clear and reassign commands |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
