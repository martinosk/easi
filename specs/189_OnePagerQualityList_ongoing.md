# 189 — One-Pager Quality master list

> **Status:** pending
> **Depends on:** [177 — OnePagerView](177_OnePagerView_done.md), [178 — One-Pager Completeness](178_OnePagerCompleteness_done.md), [186 — Mandatory Built-In One-Pager Fields](186_MandatoryBuiltInFields_pending.md) (which depend on 175, 176)

---

## Problem Statement

Specs 177–178 and 186 let an architect open any subject's one-pager and see, per subject,
whether its required fields — custom and included built-in — are filled. What is still
missing is the estate-wide view: an
architect responsible for data quality has no way to answer *"across everything that has a
one-pager, which subjects need attention, and who owns them?"* Today that means opening
subjects one at a time — there is no read that enumerates or ranks subjects by completeness.

This slice delivers a single **global One-Pager Quality master list** — one row per subject
across all six subject types — showing each subject's name, type, completeness, creator, and
its created and last-updated dates, and letting the architect sort the list by any of those to
surface the incomplete, the stale, or a particular owner's subjects first. It is a read-only
overview: the per-row "invite the owner to fix it" action is spec 190.

There is deliberately **no per-user responsibility resolution** here — the list is the whole
estate the caller may read, ranked; "whose job is this subject" is not modelled in this slice.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | See every one-pager subject in one place and sort by completeness to find the ones needing attention, or by creator/date to triage |
| **Data-quality owner** | Rank the estate by last-updated to find stale one-pagers, and by creator to see who has gaps |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: One-Pager Quality master list

  Scenario: The list shows every one-pager subject across all six types
    Given subjects exist of every subject type — Capability, Enterprise Capability,
      Application, Acquired Entity, Vendor, and Internal Team
    When I open the One-Pager Quality list
    Then it shows one row per subject
    And each row shows the subject name, subject type, a completeness signal,
      the creator, the created date, and the last-updated date

  Scenario Outline: Sorting by a stored dimension orders the whole list
    Given more one-pager subjects exist than one page holds
    When I sort the list by "<dimension>" ascending
    Then every page of the list is ordered by "<dimension>"
    And no subject appears on two pages and none is skipped between pages

    Examples:
      | dimension     |
      | name          |
      | creator       |
      | created date  |
      | last-updated  |

  Scenario: Sorting by completeness surfaces the subjects needing attention first
    Given some subjects are incomplete and some are complete
    When I sort the list by completeness
    Then incomplete subjects appear before complete ones
    And an incomplete subject is visibly distinguishable from a complete one

  Scenario: A subject type with no required fields has no completeness signal
    Given the Vendor configuration has no active required field, custom or built-in
    When I view Vendor rows in the list
    Then their completeness signal reads as not-applicable
    And they never rank as "incomplete" when sorting by completeness

  Scenario: Completeness reflects the current configuration without any facts write
    Given Application "Billing" is complete
    When an admin makes an unfilled field required for Applications
    Then "Billing" shows as incomplete in the list on the next read
    And no OnePagerFacts stream received any event

  Scenario: The list keeps pace with the estate
    Given the list is open
    When a new Capability is created, an existing Vendor is renamed, and an
      Application is deleted
    Then the new Capability appears as a row
    And the renamed Vendor shows its new name
    And the deleted Application no longer appears

  Scenario: The list shows only subject types the caller may read
    Given a user holds the Application/Vendor read permission but not the Capability one
    When they open the list
    Then Applications and Vendors appear
    And Capabilities do not appear

  Scenario: The list is tenant-isolated
    Given subjects exist in another tenant
    When I open the list
    Then only my tenant's subjects appear
```

---

## Business Rules & Invariants

1. **List scope** — one row per existing subject across the six subject types; every subject
   has a one-pager (177), so the list is the full readable subject population, not a subset.
2. **Row attributes** — name, subject type, completeness signal, creator, created date,
   last-updated date. No attribute requires a second per-row call.
3. **Sortable dimensions** — completeness, creator, name, created date, last-updated date.
   Every sort yields a deterministic total order via a stable secondary key (subject type +
   subject ID), so pagination never duplicates or skips a row.
4. **Completeness definition (178 + 186)** — a subject is complete when every active required
   field — custom **and** included built-in — has a value; built-in fill follows 186's fill
   predicate (a present snapshot value; list-valued built-ins such as Experts count as filled
   only when non-empty). A subject type with no active required field of either kind has no
   completeness signal (not-applicable), matching 178 rule 10 as extended by 186 rule 10.
5. **Completeness ordering** — incomplete first (needs attention), then complete, then
   not-applicable; ties broken by missing-required-field count — custom and built-in —
   descending, then the stable secondary key.
6. **Read-time semantics preserved (D2)** — the completeness surfaced here always equals a
   read-time evaluation against the *current* configuration; it is recomputed on every fact
   and configuration change; facts are never mutated and no edit is ever blocked.
7. **Materialized index (reverses D6)** — the list is served from one denormalized read model
   kept current by a projector; a page is returned by a single query with no per-row or
   per-page fan-out.
8. **Freshness upkeep** — creating a subject inserts its row; deleting a subject removes it
   (mirroring the 176 facts-archival deletion policy); renaming refreshes the stored name.
   Completeness is recomputed on: recording/clearing facts; defining, retiring, reactivating,
   or changing the requirement of a custom field; changing a built-in field's requirement
   (186's `BuiltInFieldRequirementChanged`); and any supplier update event that flips a
   required built-in's fill-state (e.g. an Application's experts or description, a Capability's
   maturity or experts) — because a required built-in going empty↔filled changes completeness
   with no facts event. Built-in fill is evaluated through 186's built-in fill mechanism (the
   `BuiltInFieldSource` ports), never cross-schema SQL.
9. **Creator and dates source** — creator (actor) and created date come from the subject's
   creation event (version = 1); last-updated from the latest event's `occurred_at`; sourced
   from the shared event store, covering all six types including Enterprise Capability.
10. **Read permission (177 rule 9)** — the list returns only subjects whose subject-type read
    permission the caller holds (`capabilities:read`, `enterprise-arch:read`,
    `components:read`); a caller holding none of the three receives 403.
11. **Tenant isolation** — RLS scopes the index to the caller's tenant.
12. **Boundary (D8)** — the index is fed by supplier published-language events and
    `onepagers`' own events, and built-in fill is evaluated through the `BuiltInFieldSource`
    ports; `onepagers` issues no cross-schema SQL and imports no supplier application
    packages; the projector is wired at the composition root.

---

## Acceptance Criteria

- [x] `GET /api/v1/one-pager-quality` returns a cursor-paginated list of subjects across all
      six subject types, each row carrying name, subject type, completeness signal, creator,
      created date, and last-updated date, plus `_links` (self, next).
- [x] The list is sortable by completeness, creator, name, created date, and last-updated
      date; each sort is a deterministic total order, and cursor pagination over it neither
      duplicates nor skips a row.
- [x] Sorting by completeness places incomplete subjects before complete ones; not-applicable
      subjects (no active required field of either kind) never rank as incomplete.
- [x] A completeness-changing event — a custom or built-in requirement change (186's
      `BuiltInFieldRequirementChanged`), a field retire / reactivate, or a supplier update that
      flips a required built-in's fill-state — is reflected in the list on the next read with
      zero writes to any OnePagerFacts stream.
- [x] Creating, deleting, and renaming a subject, recording/clearing facts, and supplier
      updates to a required built-in's attributes are reflected in the list via the projector;
      a backfill migration seeds the index from existing subjects, the event store, and
      existing facts/configuration, with built-in fill evaluated through the ports.
- [x] The list returns only subject types the caller may read (rule 10) and only the caller's
      tenant's subjects; a caller with none of the three read permissions receives 403.
- [x] A page is returned by a single query — no per-row or per-page completeness/creator
      fan-out; `onepagers` contains no supplier imports and no cross-schema SQL (the spec-175
      architecture boundary test stays green).
- [x] A routed page renders the list as a sortable table (Mantine v8), with a visible
      complete/incomplete indicator, sort controls, and cursor pagination.
- [x] Every BDD scenario above has a corresponding automated test.

---

## Architecture

### Ownership

`onepagers` owns the master list: a new denormalized read model, its projector, the query
service, and the endpoint. Completeness is already an `onepagers` concept (178). Supplier
contexts (`capabilitymapping`, `enterprisearchitecture`, `architecturemodeling`) are touched
only as event sources — the projector reacts to their published-language creation, deletion,
and rename events, exactly as the 176 deletion policy already reacts to their deletions.

### Domain Model

No new aggregates or events. The list is a query-side projection over existing facts: the six
subjects' lifecycle events (for name, creator, created, last-updated), the shared event store
(for actor and `occurred_at`), the OnePagerFacts read model (176) for custom fill, the subject
snapshots via the `BuiltInFieldSource` ports for built-in fill (186), and the active required
custom **and** built-in fields of the OnePagerConfiguration read model (175 + 186).

### API Surface

- `GET /api/v1/one-pager-quality?sort=&order=&limit=&after=` — `sort` ∈ {completeness,
  creator, name, created, updated} (default: completeness), `order` ∈ {asc, desc}, `limit`
  default 50 / max 100, `after` an opaque keyset cursor (the `GET /api/v1/components`
  precedent). Response: rows with the six attributes and `_links` (self, next). Filtered to
  the caller's readable subject types (rule 10); 403 when the caller may read none.

### Persistence

New denormalized read model `onepagers.one_pager_subject_index`, PK `(tenant_id, subject_type,
subject_id)`, carrying the stored name, `creator_actor_id`, `creator_email`, `created_at`,
`last_updated_at`, and the completeness inputs (`required_count`, `filled_count`) — now
spanning **both** required custom fields and required included built-in fields (186) — from
which the complete / not-applicable signal and missing count derive. Custom fill-state comes
from the `one_pager_facts` read model; built-in fill-state derives from the subject's own
attributes via 186's built-in fill predicate (evaluated through the `BuiltInFieldSource`
ports), not the facts table. Indexes cover each sort key; RLS tenancy as on every table. A
projector maintains it; a **backfill migration** seeds it from the existing subject read
models (names), the shared event store (creator / dates), and the facts + configuration read
models plus the built-in fill ports (completeness) — the mandatory backfill for any new read
model.

### Frontend

- New routed page rendering the list as a sortable, cursor-paginated Mantine table with a
  visible complete/incomplete indicator per row and column sort controls. Data via TanStack
  Query keyed on sort + order + cursor. Entry point (nav/menu) gated on the HATEOAS link.

### Cross-Context Integration

Customer/Supplier, query-time, no new outbound events. The index projector consumes supplier
published-language creation, deletion, and rename events, the supplier **update** events that
flip a required built-in's fill-state (e.g. an Application's experts/description, a
Capability's maturity/experts), and `onepagers`' own OnePagerFacts and OnePagerConfiguration
events — including 186's `BuiltInFieldRequirementChanged`. When a built-in requirement or
fill-state changes, the projector recomputes the affected subject's completeness through 186's
built-in fill mechanism (the `BuiltInFieldSource` ports). It is wired at the composition root,
never through cross-schema SQL or supplier imports.

---

## Design Decisions

1. **Denormalized subject index + projector, not a query-time union** — six heterogeneous
   read models in separate schemas plus event-store creator/dates cannot be unioned at query
   time within the D8 boundary, and an `ORDER BY completeness` cursor cannot be served by the
   page-scoped set-based query (178). The `ea_*_cache` / `capability_component_cache`
   projector precedent is the idiomatic drop-in. Alternatives: query-time union/merge across
   contexts (rejected — boundary violation, unbounded cost, no cursor-stable completeness
   sort); page-then-decorate completeness per 178 (rejected — serves stored-column sorts but
   cannot order the *global* list by completeness, which is a first-class requirement here).
2. **Reverse D6: materialize completeness in the index** — D6 named the completeness cache as
   the drop-in for "when set-based proves insufficient." A globally sorted, cursor-paginated
   list ordered by completeness is exactly that case. This does **not** reverse D2: the
   materialized value is a projector-maintained cache of the read-time computation, recomputed
   on every fact, configuration, *and* supplier built-in-attribute event, so it always equals
   a current-configuration read-time evaluation — facts are never mutated and edits are never
   blocked. Completeness spans required custom and included built-in fields (186); because
   built-in fill-state lives in the subject's own attributes rather than the facts table, the
   projector also recomputes on 186's `BuiltInFieldRequirementChanged` and on the supplier
   update events that flip a required built-in empty↔filled, evaluating built-in fill through
   the `BuiltInFieldSource` ports — keeping the D8 boundary intact. Alternative: compute
   completeness for the whole population per request and sort in SQL (rejected — the hot,
   cursor-unstable case D6 warned against).
3. **Materialize name, creator, and dates too** — name is kept fresh by rename events (the
   `ea_realization_cache` `UpdateComponentName` precedent); creator and dates come from the
   event store (version = 1 and latest `occurred_at`), covering Enterprise Capability, which
   the existing `GetArtifactCreators` read model omits and which carries no dates. Alternative:
   store IDs only and decorate name/creator/dates per page via ports/audit (rejected — you
   cannot sort a global list by a value you have not stored, and no set-based per-subject
   creator/date projection exists).
4. **Cursor (keyset) pagination with a total-order tiebreaker** — matches
   `GET /api/v1/components`; stable under concurrent inserts. Alternative: offset (rejected —
   drifts and duplicates rows as subjects are created).
5. **Row-level filtering by the subject-type read permission (177 rule 9)** — a global list
   cannot resolve a single subject permission, so rows are filtered to the types the caller
   may read; no new permission is introduced. Alternative: a dedicated quality-overview
   permission (rejected — 177 already rejected a bespoke one-pager permission; it would reveal
   or hide data inconsistently with the subject's own read gate).
6. **Index owned by `onepagers`, projector wired at the composition root** — completeness is
   `onepagers`' concept, and supplier data arrives via published-language events (like the 176
   deletion policy), preserving D8.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Materialized completeness cache (reverses D6) | A read-time value is now stored, and configuration changes trigger a recompute | Recompute is set-based — one bulk UPDATE per configuration change, one row per fact or supplier built-in-attribute change; the value always reflects current config, so D2 semantics hold |
| Completeness spans built-ins (186), whose fill lives in supplier attributes not the facts table | The projector must also subscribe to the supplier built-in-attribute update events and to `BuiltInFieldRequirementChanged`, and re-evaluate fill through ports — a wider subscription surface | Reuses 186's built-in fill mechanism through the `BuiltInFieldSource` ports; recompute stays set-based and boundary-clean (no cross-schema SQL) |
| Denormalized index across six contexts | Many event subscriptions plus a backfill migration to seed | Mirrors the `ea_*_cache` precedent; backfill `INSERT…SELECT` seeds from existing subjects, the event store, and facts/config; one query per page is the payoff |
| Denormalized name | A rename must re-project | One rename-event subscription per subject type (`UpdateComponentName` precedent) |
| Row-level permission filtering | The paged set varies with the caller's permissions | Filter is a `WHERE subject_type IN (…)`; the keyset cursor still yields a total order per caller |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [ ] User sign-off
