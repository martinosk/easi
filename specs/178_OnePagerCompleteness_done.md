# 178 — One-Pager Completeness & Impact Preview

> **Status:** pending
> **Depends on:** 175 (OnePagerConfiguration), 176 (OnePagerFacts capture), 177 (OnePagerView)

---

> **Amended by spec 208 (2026-08-29):** the per-row `onePagerComplete` indicator on subject list responses (design decision 4) is replaced by `GET /one-pagers/{subjectType}/completeness`, served by OnePagers; Architecture Modeling, Capability Mapping and Enterprise Architecture no longer decorate their lists.

## Problem Statement

One-pager configurations evolve: tenant admins add custom fields and, over time, make them
required. Hundreds of subjects may already exist when a field is flipped optional → required,
and their recorded facts must never be invalidated or blocked by that configuration change.
Yet architects still need to see *which* one-pagers fall short of the current bar, and admins
need to understand the blast radius of a requirement change *before* committing it.

This slice delivers **Completeness** — how fully a subject's one-pager satisfies the current
configuration's required custom fields — evaluated at read time against the current
configuration (design doc D2). It surfaces on the one-pager view, on subject list views, and
as an **impact preview** in the settings UI when an admin changes a field's required flag.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Tenant Administrator** | See how many subjects a requirement change will mark incomplete before confirming it |
| **Enterprise Architect** | Spot incomplete one-pagers at a glance in lists and see exactly which required fields are missing on a one-pager |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: One-pager completeness evaluated at read time

  Scenario: Complete subject shows complete
    Given the Application configuration has 2 active required custom fields
    And Application "Billing" has a value for both
    When I view the "Billing" one-pager
    Then it shows a completeness summary of "2 of 2 required fields filled"
    And no field renders as missing

  Scenario: Missing required value flags incomplete with named fields
    Given the Application configuration has required fields "Contract link" and "Contact person"
    And Application "Billing" has a value only for "Contact person"
    When I view the "Billing" one-pager
    Then "Contract link" renders as "missing — required"
    And the completeness summary shows "1 of 2 required fields filled"

  Scenario: Optional field missing does not affect completeness
    Given the Application configuration has one required field with a value on "Billing"
    And an optional field "Notes" with no value on "Billing"
    When I view the "Billing" one-pager
    Then "Billing" is complete and "Notes" does not render as missing — required

  Scenario: Retiring a required field restores completeness
    Given Application "Billing" is incomplete solely because "Contract link" has no value
    When an admin retires the "Contract link" field
    Then viewing the "Billing" one-pager shows it as complete
    And no fact of "Billing" was modified

  Scenario: Reactivating a retired required field removes completeness again
    Given "Contract link" is retired and "Billing" shows as complete
    When an admin reactivates "Contract link"
    Then viewing the "Billing" one-pager shows it as incomplete, naming "Contract link"

  Scenario: Flipping a field to required updates completeness at read time only
    Given 37 Applications lack a value for the optional field "Contract link"
    When an admin makes "Contract link" required
    Then all 37 Applications evaluate as incomplete on their next read
    And no OnePagerFacts stream receives any new event
    And every one of the 37 Applications can still save partial one-pager facts

  Scenario: Impact preview before confirming a requirement change
    Given 37 Applications lack a value for the optional field "Contract link"
    When an admin flips "Contract link" to required in the settings UI
    Then before confirming, the UI shows
      "Making Contract link required will mark 37 Applications incomplete"
    And confirming does not block edits on any existing Application

  Scenario: Impact preview for a newly defined required field
    Given the tenant has 120 Vendors and no Vendor has one-pager facts for a new field
    When an admin defines a new required custom field on the Vendor configuration
    Then the preview shows that 120 Vendors will be marked incomplete

  Scenario: List indicator correct under pagination
    Given the Application configuration has at least one active required field
    And more Applications exist than one page holds
    When I page through the application list
    Then every row on every page carries the correct completeness indicator
    And the page size, ordering, and next-page cursor behave exactly as before

  Scenario: No indicator without required fields
    Given the Vendor configuration has no active required custom fields
    When I view the vendor list
    Then no completeness indicator is shown for vendors
```

---

## Business Rules & Invariants

1. **Read-time evaluation (D2)** — completeness is computed at read time against the current
   One-Pager Configuration; it is never stored on, nor derived from, the facts themselves.
2. **Completeness definition** — a subject is complete when every *active, required* custom
   field of its subject type's configuration has a recorded Field Value on that subject.
3. **Built-in fields excluded (D7)** — built-in fields never participate in completeness.
4. **Facts are inviolable** — configuration changes (require, retire, reactivate) never mutate,
   archive, or invalidate recorded facts, and never retroactively block any subject edit.
5. **Soft enforcement** — saving partial facts remains allowed regardless of completeness;
   incompleteness is a rendered signal, not a write gate.
6. **Set-based computation** — completeness for many subjects is computed with one set-based
   query per subject type (facts joined against the active required-field list); per-entity
   fan-out is prohibited.
7. **Impact preview is a pure query** — previewing a requirement change is side-effect-free:
   no event is appended, no configuration or fact changes. N = subjects of the type lacking a
   value for the field; for a field not yet defined, N = the full subject population.
8. **Population counts come through the subject port** — the count of subjects of a type is a
   port method on the `BuiltInFieldSource` subject port, implemented by composition-root
   adapters; `onepagers` never issues SQL against another context's tables.
9. **Preview permission** — the impact preview requires the same admin permission as writing
   the One-Pager Configuration (spec 175).
10. **Indicator scope** — subject list views show a completeness indicator only where a
    One-Pager Configuration with at least one active required custom field exists for that
    subject type.

---

## Acceptance Criteria

- [x] The one-pager response includes per-subject completeness: filled/required counts plus
      the IDs and display names of missing required fields — no second frontend call needed.
- [x] The one-pager page renders each valueless required field as "missing — required" and
      shows a completeness summary (e.g. "3 of 5 required fields filled").
- [x] Optional fields without values never affect completeness or render as missing.
- [x] Retiring a required field restores completeness on next read; reactivating it removes
      completeness again — with zero writes to any OnePagerFacts stream in both cases.
- [x] Flipping a field required marks all subjects lacking a value incomplete on next read,
      appends no facts events, and blocks no subject or facts edit.
- [x] Subject list endpoints return a per-row completeness indicator where rule 10 applies,
      computed set-based for the returned page; cursor pagination, ordering, and page size
      are unchanged, and the query count per page stays constant (no N+1).
- [x] The settings UI shows "Making <field> required will mark <N> <subject type>s
      incomplete" before the admin confirms a requirement change or a new required field,
      with N correct per rule 7.
- [x] The impact preview endpoint is a side-effect-free GET guarded by the configuration
      write permission; population counts flow through the subject port method (rule 8).
- [x] No completeness cache table or projector exists in this slice (D6).
- [x] Every BDD scenario above has a corresponding automated test.

---

## Architecture

### Ownership

`onepagers` bounded context owns completeness evaluation and the impact preview. Supplier
contexts (capabilitymapping, enterprisearchitecture, architecturemodeling) are touched only
through the existing consumer-defined ports; their list endpoints gain the indicator by
composing the `onepagers` completeness query into the list read at the composition root.

### Domain Model

No new aggregates, entities, or events. Completeness is a query-side concept derived from two
existing read models: the configuration's active required Custom Field Definitions (175) and
the OnePagerFacts read model (176). The subject port gains a population-count method per the
Risks & Guards of the design doc.

### API Surface

- **One-pager read (177)** — the response is extended with a completeness block: required
  count, filled count, and the missing required fields (FieldID + display name).
- **Subject list reads** — the list response DTOs gain a completeness indicator per row.
  Reference shape (component listing, `GET /api/v1/components`, cursor-paginated): after the
  page of rows is selected, one set-based query over exactly that page's subject IDs joins
  facts against the required-field list and decorates the rows — a constant one extra query
  per page, cursor generation untouched. When rule 10 does not apply, the extra query is
  skipped and the field is absent.
- **Impact preview** — a side-effect-free GET query on the One-Pager Configuration resource
  taking the subject type and field (existing FieldID, or "new field" for a definition being
  created), returning the affected-subject count. Requires the configuration write permission.

### Persistence

None. No migration, no new tables, no completeness cache (D6). The set-based query runs over
the existing `one_pager_facts` read model (PK `(tenant_id, subject_type, subject_id,
field_id)`) and the configuration read model, under the standard RLS tenancy.

### Frontend

- **One-pager page (177)**: "missing — required" rendering per field and the completeness
  summary, driven entirely by the extended one-pager response.
- **Subject list/table surfaces**: a per-row completeness indicator where rule 10 applies.
- **Settings `/settings/one-pagers` (175, per the spec 095 settings precedent)**: on toggling
  a field's required flag or defining a new required field, fetch the impact preview and show
  the count in the confirmation step before the mutation is sent.

### Cross-Context Integration

No new events. The only new cross-context touchpoint is the population-count port method on
the subject port, implemented by the composition-root adapters — never raw cross-schema SQL.

---

## Design Decisions

1. **Read-time completeness, no retroactive enforcement (D2)** — configuration changes never
   mutate facts and never block edits; incompleteness is a signal evaluated against the
   current configuration on every read. Alternative: hard-enforce required at write time
   (rejected in the design doc — flipping a flag would freeze edits on every incomplete
   pre-existing entity).
2. **No completeness cache in this slice (D6)** — set-based queries per subject type are
   expected to suffice; the existing `ea_*_cache` projector precedent makes a
   `one_pager_completeness` cache a drop-in later if profiling demands it. Alternative:
   build the cache projector now (rejected — speculative infrastructure ahead of evidence).
3. **Custom fields only (D7)** — built-in fields carry no required flag and never enter the
   completeness computation. Alternative: let built-ins count (rejected — changes the query
   shape and configuration events; a separate spec if ever wanted).
4. **Page-scoped set query for list indicators** — decorate the already-selected page with
   one query over its subject IDs rather than joining completeness into the paginated query
   itself. Keeps cursors, ordering, and supplier read models untouched at a constant query
   cost. Alternative: per-row lookups (rejected — N+1); joining across contexts in the list
   SQL (rejected — cross-schema coupling prohibited by D8).
5. **Impact preview as a port-backed query** — subject population counts come from a method
   on the `BuiltInFieldSource` subject port, matching every other supplier read. Alternative:
   raw cross-schema count SQL (rejected — explicitly guarded against in the design doc).
6. **Preview shares the configuration write permission** — only an admin who can flip the
   flag can preview its impact; the preview exists solely inside that admin workflow.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Read-time evaluation, no cache (D6) | Every list page and one-pager read recomputes completeness | Set-based indexed queries at constant per-request cost; cache projector precedent is a drop-in escape hatch |
| Soft enforcement of required fields | Data can remain incomplete indefinitely | Visible indicators on lists and one-pagers plus the impact preview make gaps actionable |
| Page-scoped decoration query | One extra query per list page | Bounded by page size (≤ 100), keyed on the facts PK |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented and passing
- [x] API documentation updated
- [x] User sign-off
