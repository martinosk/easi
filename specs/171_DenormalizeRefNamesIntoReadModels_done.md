# 171 — Denormalize Cross-Context Reference Names into Read Models

> **Status:** pending
> **Depends on:** [167 — Direction on an Enterprise Capability](167_Direction_Aggregate_Capture_done.md), [170 — Standard Application on an Enterprise Capability](170_StandardAppDesignation_ongoing.md)

---

## Problem Statement

`Direction` source-capability rows and `StandardApplication` rows store only the ID of the referenced foreign-context entity (physical capability, application component). The name lives in the producing bounded context's read model. Today every UI consumer joins those names in JavaScript by issuing extra queries (`useCapabilities`, `useComponents`, `useEnterpriseCapabilityLinks`) and falling back to the raw GUID when a lookup misses. The result: GUIDs flash on load, unlinked references render as GUIDs permanently, and rendering a single Direction requires three to four cross-context HTTP calls.

This slice moves the join to the read side, where it belongs. The reading projector subscribes to the producer's create / rename / delete events and maintains the name on the row alongside the existing `stale` flag.

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Cross-context reference names render directly from the read model

  Scenario: A Direction's source capabilities render with names, never GUIDs
    Given a Direction with one or more source physical capabilities
    When I view the Direction
    Then each source renders with the capability's current name
    And no source ever renders as a raw UUID, regardless of whether it is linked to the EC

  Scenario: A renamed source capability is reflected on the Direction
    Given a Direction references a physical capability named "Payroll"
    When the physical capability is renamed to "Payroll (Group)"
    Then the Direction's source list shows "Payroll (Group)" within the projection lag window

  Scenario: A source capability's business domain renders by name
    Given a Direction's source capability belongs to the business domain "Passenger"
    When I view the Direction
    Then the source line renders "Passenger" alongside the capability name, never the domain's UUID

  Scenario: A renamed business domain is reflected on every Direction whose sources reference it
    Given the business domain "Passenger" is referenced by source capabilities on one or more Directions
    When the business domain is renamed to "Passenger Operations"
    Then every Direction's affected source rows show "Passenger Operations" within the projection lag window

  Scenario: A Standard Application renders the application name directly
    Given an Enterprise Capability has a Standard Application set
    When I view the Enterprise Capability
    Then the panel shows the application's current name without a separate client-side lookup
    And the history dialog shows each entry's application name (current and previous)

  Scenario: A renamed application is reflected on every Standard Application that references it
    Given the application "Acme ERP" is the current standard on two Enterprise Capabilities
    When the application is renamed to "Acme ERP (Cloud)"
    Then both panels show "Acme ERP (Cloud)" within the projection lag window
```

---

## Business Rules & Invariants

1. **Read-model rows that reference a foreign-context entity carry the entity's current name alongside the ID.** Applies to `architecturedirection.direction_source_capabilities` (`capability_name`, `business_domain_id`, `business_domain_name`) and `architecturedirection.standard_applications` (`application_name`) and `architecturedirection.standard_application_history` (`application_name`, `previous_application_name`).
2. **The name is maintained by the reading projector** subscribing to the producer's create / update events. Same projector pattern as the existing stale-reference projector for delete events.
3. **A row may briefly carry a NULL name** if the producer's event has not yet been observed. The read model is eventually consistent. The UI must render a NULL name as a clear loading placeholder (e.g. `—`), never as the raw ID.
4. **The frontend renders the name directly from the DTO.** Client-side cross-context joins for the sole purpose of name resolution are removed.

---

## Acceptance Criteria

- [ ] `architecturedirection.direction_source_capabilities` has `capability_name`, `business_domain_id`, `business_domain_name` columns; populated on row insert and maintained by subscriptions to the relevant `capabilitymapping` create / rename events for both physical capabilities and business domains
- [ ] `architecturedirection.standard_applications` has an `application_name` column; populated and maintained by subscriptions to `architecturemodeling` `ApplicationComponentCreated` / `ApplicationComponentUpdated`
- [ ] `architecturedirection.standard_application_history` has `application_name` and `previous_application_name` columns, populated on append
- [ ] `DirectionSourceCapabilityDTO` adds `name`, `businessDomainId`, `businessDomainName`; the Direction GET response carries them directly
- [ ] `StandardApplicationDTO` adds `applicationName string`; `StandardApplicationHistoryEntryDTO` adds `applicationName` and `previousApplicationName`
- [ ] On the frontend, `DirectionPanel` no longer calls `useCapabilities`, `useEnterpriseCapabilityLinks`, or any other cross-context query for source-row rendering; `useBusinessDomainsQuery` remains only for placement-domain resolution
- [ ] On the frontend, `StandardApplicationPanel` and `StandardApplicationHistoryDialog` no longer call `useComponents`
- [ ] The name fallback rendering "raw UUID in `<Text component='code'>`" is removed from `DirectionPanel.SourceList`
- [ ] A row may carry NULL name fields during the projection lag window; the UI renders a placeholder (e.g. `—`), never the raw UUID
- [ ] All BDD scenarios above have at least one corresponding test, including backend integration tests verifying rename propagation for both physical capability and business domain
- [ ] Every modified file scores 10.0 in CodeScene per `easi-codehealth`

---

## Architecture

### Ownership
`architecturedirection` projector subscriptions expand; no aggregate change. Producing contexts (`capabilitymapping`, `architecturemodeling`) are unaffected.

### Domain Model
No aggregate change. Reference VOs unchanged. The denormalized name lives only in the read model; it is *not* part of the aggregate.

### Persistence
New columns added via simple `ALTER TABLE ADD COLUMN` — no backfill needed (neither aggregate has shipped, both read-model tables are empty in every environment).

### Projectors
New projector subscriptions in `architecturedirection`:
- `capabilitymapping` capability create / rename → update `direction_source_capabilities.capability_name` for every row matching `capability_id`
- `capabilitymapping` business-domain create / rename → update `direction_source_capabilities.business_domain_name` for every row matching `business_domain_id`
- `architecturemodeling` `ApplicationComponentCreated` / `ApplicationComponentUpdated` → update `standard_applications.application_name` for every row matching `application_id`

History rows are immutable per-entry; `application_name` and `previous_application_name` are resolved at append time and never re-written.

### Frontend
Drop client-side name resolution for these surfaces. The DTO contains every name needed. `DirectionPanel`'s source row no longer joins against any external query. `useEnterpriseCapabilityLinks` is removed from `DirectionPanel` entirely. `useBusinessDomainsQuery` remains only for placement-domain rendering (the projection there is independent and small).

### Cross-Context Integration
`architecturedirection` already subscribes to delete events from both producers. This slice adds the create/update subscriptions using the same mechanism. Payload structures are duplicated locally per the existing pattern in `stale_reference_projector.go` and `stale_application_projector.go` (see `easi-domain-driven-design` — the rule against cross-BC contract imports is unchanged here).

---

## Cleanup (delete in this slice)

Backend:
- No projector files removed; new subscriptions extend existing files

Frontend (`DirectionPanel.tsx`):
- Remove `useCapabilities`, `useEnterpriseCapabilityLinks`, and the `useNameResolvers` helper's source-capability + source-domain branches entirely
- Remove the `<Text component="code">{source.id}</Text>` fallback in `SourceList`
- Render `source.name` and `source.businessDomainName` directly from the DTO

Frontend (`DirectionPanel.test.tsx`):
- Remove the `useCapabilities` and `useEnterpriseCapabilityLinks` mocks for source-row resolution
- Remove `CapabilityFixture`, `LinkFixture`, and the "resolves source capability names via the global capabilities query" test (its scenario is structurally impossible after this slice)
- Source-name assertions become direct DTO-field assertions

Frontend (`StandardApplicationPanel.tsx` + test):
- Remove `useComponents` import and the `applicationName` `useMemo` in `StandardDetail`
- Render `standard.applicationName` directly
- Remove the `useComponents` mock from the test; assert directly on the DTO field

Frontend (`StandardApplicationHistoryDialog.tsx` + test):
- Remove `useComponents` import and `nameByApplicationId` `useMemo`
- Render `entry.applicationName` and `entry.previousApplicationName` directly
- Remove the `useComponents` mock from the test; assert directly on the DTO fields

---

## Design Decisions

1. **Denormalize at the read side, not the write side.** The aggregate has no business owning the foreign entity's name. Writing into the read model via a subscription keeps the aggregate clean and the join cheap at query time.
2. **History rows are snapshots.** `application_name` and `previous_application_name` are resolved at append time and never re-written. Renaming an application later does not retroactively change what past history rows say — that is the correct audit behaviour (a past entry recorded what was true *then*).
3. **Current-state rows track renames.** The current row on `standard_applications` and the `direction_source_capabilities` row track the producer's current name via event subscription. Users see the up-to-date name on the active panel.
4. **NULL name placeholder, not raw ID.** During the projection lag window the UI renders `—` or similar. Never the raw UUID. The raw UUID has no business surfacing to a non-developer user.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|---|---|---|
| Denormalize names into the read model | Cross-context event subscriptions multiply; the architecturedirection projector now listens to more producer events | Same subscription mechanism already exists for stale-flag handling; one additional event per producer per row type |
| History rows snapshot the name at append | Renaming an application does not update past history entries | Correct audit behaviour; the history records what was true at the time |
| Simple migration, no backfill | Existing rows (none, since neither aggregate has shipped) would carry NULL names if any existed | Not applicable — both read-model tables are empty in every environment |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off
