# 238 — Guided Application Creation

> **Status:** pending
> **Depends on:** 237 (similar-entity suggestions), 214 (ownership), 215 (hosting), 232 (realising applications add/remove)
> **Roadmap alignment:** `SD6` (application record) — slice D of [`docs/specs/data-quality.md`](../docs/specs/data-quality.md) (decision Q3). No bounded-context change.

---

## Problem Statement

Creating an application asks for a name and a description. Hosting, owner and the capability it realises are each set later from a different place, and in practice often never: in a reviewed production tenant 185 of 305 applications had hosting unknown, 293 had no owner and 49 realised no capability. The person creating the application usually knows all three at that moment.

This slice continues creation into a short **complete the record** step that asks for hosting, owner and the first realised capability, using the existing commands. Each can be skipped; what is skipped is shown, so an incomplete record is a visible choice rather than an accident.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architect** | Enter a complete application in one sitting |
| **Architect sketching on the canvas** | Create quickly, and be offered completion without being forced through it |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Guided application creation

  Scenario: Creation continues into completing the record
    When an architect creates the application "Dwell Time Prediction" from the create dialog
    Then the application exists
    And the dialog shows "Complete the record" with hosting, owner and realised capability, none of them set

  Scenario: Classifying hosting
    Given the completion step is shown for "Dwell Time Prediction"
    When the architect selects hosting "Cloud"
    Then the application's hosting is "cloud"
    And the step shows hosting as done

  Scenario: Nominating an owner
    When the architect picks the user "Ole Berg" as owner
    Then "Ole Berg" is nominated as owner of the application
    And the step shows owner as done

  Scenario: Assigning a team
    When the architect picks the internal team "Terminal Ops" as owner
    Then the application is managed by "Terminal Ops"

  Scenario: Linking the first capability
    When the architect picks the capability "Unit Status Monitoring"
    Then the application realises "Unit Status Monitoring"
    And the step shows the domains the application now belongs to

  Scenario: Finishing with everything set
    Given hosting, owner and a capability are set
    When the architect selects "Done"
    Then the dialog closes

  Scenario: Finishing with gaps says so
    Given hosting is set and owner and capability are not
    Then the finish button reads "Finish later"
    And the step states "Owner and realised capability are not set"
    When the architect selects "Finish later"
    Then the dialog closes and the application keeps what was set

  Scenario: Closing the dialog keeps the application
    Given the completion step is shown
    When the architect closes the dialog
    Then the application still exists with whatever was set

  Scenario: Canvas creation stays quick
    When an architect creates an application from a canvas handle
    Then the application appears on the canvas immediately
    And a notification offers "Complete the record"
    When the architect selects it
    Then the completion step opens for that application

  Scenario: Completing later from the details panel
    Given the application "Portal" has hosting unknown and no owner
    When an architect opens its details
    Then a "Complete the record" prompt names the missing items
    And selecting it opens the completion step
    Given every item is set
    Then no prompt is shown

  Scenario: Only what the user may do is offered
    Given the caller may edit the application but may not link capabilities
    Then the completion step offers hosting and owner and not the capability

  Scenario: A failing step does not lose the others
    Given hosting was set
    When nominating the owner fails
    Then the step shows the error on owner only
    And hosting remains set

  Scenario: Required one-pager fields are part of completing
    Given the tenant requires the custom field "Criticality" on application one-pagers
    Then the completion step also asks for "Criticality"
    When the architect selects "High"
    Then the application's one-pager records "High"
```

---

## Business Rules & Invariants

1. **Create first, then complete** — the application is created by the existing command as soon as name and description are submitted. Every completion item is a separate, existing command on the created application. There is no composite command and no draft application.
2. **Completion items** — hosting classified (not `unknown`); owner present (ownership state not `unknown`); at least one direct realisation; every required custom field of the application one-pager that the meta-model can render as a single input (selection, text, number, date). In that order.
3. **Each item is offered only by its affordance** — hosting by `x-classify-hosting`; owner by `x-nominate-owner` or `x-assign-owner`; capability by the realisation affordance of spec 232; facts by the one-pager's `x-record`. An item without its link is not shown and not counted as missing.
4. **Owner choice follows spec 214** — a user is nominated when the caller may only nominate and assigned when they may assign; a team is assigned as `managed`.
5. **Items save individually and immediately** — setting one item never waits for another and a failure affects only that item.
6. **Nothing is mandatory** — the flow can be left at any point; the application keeps whatever was set.
7. **Completeness is derived** — "complete" is computed from the application's current state by the same definition everywhere (dialog, details prompt). Nothing records that the guide was shown, finished or skipped.
8. **Quick creation is not slowed** — canvas handle-create and any other one-step creation create immediately and only offer the completion step.

---

## Acceptance Criteria

- [ ] After a successful create from `CreateComponentDialog`, the dialog shows the completion step for the created application without a second navigation
- [ ] Each completion item calls the existing command through the link named in rule 3 and reflects the server's response; no new backend endpoint is added by this spec
- [ ] The step shows the per-item state (not set, saving, done, failed) and a summary of what is missing; the finish label is "Done" when nothing offered is missing and "Finish later" otherwise
- [ ] An item whose affordance is absent is neither rendered nor counted
- [ ] The capability picker searches capabilities by name and shows level and parent path; after linking, the application's domains are shown
- [ ] The owner picker offers users and internal teams and applies rule 4
- [ ] Required single-input custom fields appear as items and record through the one-pager facts API; fields of other types are not shown
- [ ] Canvas handle-create shows a notification with "Complete the record" that opens the step for the created application
- [ ] The application details panel shows the prompt while any offered item is missing and hides it otherwise, using the same completeness function as the dialog
- [ ] Closing or leaving the step never deletes or alters the application
- [ ] Mutation effects invalidate the component, ownership statistics, realisations, one-pager and home queries so every open surface reflects completed items
- [ ] Every BDD scenario has at least one corresponding test
- [ ] E2E: create an application, set hosting, owner and capability in the guide, and find all three on its details
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

Frontend only, in the `components` feature. Architecture Modeling, Capability Mapping and OnePagers are used through their existing commands and affordances.

### Domain Model

None. The completeness function is a pure frontend calculation over the component resource, its realisations and its one-pager.

### API Surface

No change.

### Persistence

None.

### Frontend

- `CompleteRecordStep`: a list of items, each a small form bound to one mutation, built from Mantine primitives.
- `CreateComponentDialog` gains the step after create; the canvas handle-create controller gains the notification; the application details panel gains the prompt.
- A `recordCompleteness` utility shared by all three.
- Affordances are read from the created resource's `_links` (the create response, then the refetched resource).

### Cross-Context Integration

None beyond the existing client calls.

---

## Design Decisions

1. **Existing commands, no composite create** — model state is immediate; a second write path would drift from the single-field commands, and linking a capability is another context's command that cannot share a transaction with creation. Alternative: `POST /components` accepting hosting, owner and capability (rejected for those reasons).
2. **Nothing is mandatory** — an architect sketching a landscape must be able to name a box and move on; forcing fields produces invented values, which are worse than a visible gap. The gap is then reported (spec 236) to the creator (spec 235).
3. **Derived completeness, no "guide dismissed" state** — the prompt disappears when the data is there and for no other reason, the same principle as attention items.
4. **Required custom fields come from the tenant's configuration** — the tenant already declares what a complete application one-pager is; the guide asks for exactly that and no hard-coded extra fields. Criticality is such a field.
5. **Description is not an item** — it is already on the create form.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Several requests instead of one | A partially completed record if the user leaves midway | Each saved item is valid on its own; the details prompt offers the rest |
| Frontend-only completeness | The definition lives in the client | It mirrors the checks of specs 230 and 236 item for item, with tests naming each check |
| Only single-input custom fields | Rich fields are not asked for in the guide | They remain on the one-pager and count toward its completeness |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
