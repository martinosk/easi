# 186 — Mandatory Built-In One-Pager Fields

> **Status:** done
> **Depends on:** 175 (OnePagerConfiguration), 176 (OnePagerFacts capture), 177 (OnePagerView), 178 (Completeness & impact preview)
> **Supersedes:** design decision **D7** of [Configurable One-Pagers](../docs/specs/configurable-one-pagers.md)

---

## Problem Statement

Spec 178 delivered read-time completeness, but only for **custom** fields: design decision D7
ruled that built-in fields "carry no required flag and never participate in completeness."
In practice a tenant's most important stakeholder facts are often built-in — a Capability's
experts, an Application's description, a Vendor's implementation partner. Today an admin can
demand a contact-person custom field but cannot demand that every Application actually names
its experts, so completeness under-reports the real data-quality bar.

This slice **reverses D7** (explicitly confirmed with the product owner): a tenant admin may
mark an **included** built-in field required, and a required built-in participates in
completeness with **full parity** to a required custom field — on the one-pager view, on
subject-list indicators, and in the requirement-change impact preview.

The crux is mechanical. Custom fill-state lives in the `one_pager_facts` table, so spec 178
computes completeness with one set-based SQL query. Built-in values do **not** live there —
they come from supplier read models through the `BuiltInFieldSource` ports. This spec extends
completeness to built-ins **without** violating the boundary rule that `onepagers` reads
suppliers only through ports and never issues cross-schema SQL.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Tenant Administrator** | Require a built-in field (e.g. Experts) and, before confirming, see how many subjects it will mark incomplete |
| **Enterprise Architect** | See a one-pager and a list flag a missing required built-in exactly as they already flag a missing required custom field |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Mandatory built-in one-pager fields

  Scenario: Marking an included built-in field required
    Given the Application configuration includes the built-in field "Experts"
    When an admin marks "Experts" required
    Then the configuration records "Experts" as a required built-in field
    And no recorded data is validated, mutated, or blocked by the change

  Scenario: Only included built-in fields can be marked required
    Given the built-in field "Experts" is excluded from the Application configuration
    Then the configuration exposes no affordance to mark "Experts" required
    And a command to mark it required is rejected

  Scenario: A required built-in with no data flags incomplete on the one-pager
    Given the Application configuration marks the built-in field "Experts" required
    And Application "Billing" has no experts
    When I view the "Billing" one-pager
    Then "Experts" renders as "missing — required" instead of an em-dash
    And the completeness summary counts "Experts" among the required fields and as unfilled

  Scenario: A populated required built-in counts as filled
    Given the Application configuration marks the built-in field "Experts" required
    And Application "Billing" has at least one expert
    When I view the "Billing" one-pager
    Then "Experts" renders its experts and not as missing
    And it counts as a filled required field in the completeness summary

  Scenario: Experts with an empty list counts as missing when required
    Given the Capability configuration marks the built-in field "Experts" required
    And Capability "Payments" has an empty experts list
    When I view the "Payments" one-pager
    Then "Experts" renders as "missing — required"

  Scenario: Completeness combines required custom and built-in fields
    Given the Application configuration has one required custom field with a value on "Billing"
    And it marks the built-in field "Experts" required
    And "Billing" has no experts
    When I view the "Billing" one-pager
    Then the completeness summary shows "1 of 2 required fields filled"

  Scenario: Subject list indicator reflects a missing required built-in
    Given the Application configuration marks the built-in field "Experts" required
    And Application "Billing" has no experts
    When I view the application list
    Then "Billing" carries an incomplete indicator

  Scenario: Impact preview counts subjects lacking a built-in value
    Given the tenant has 120 Applications and 40 of them have no experts
    When an admin flips the built-in field "Experts" to required in the settings UI
    Then before confirming, the UI shows
      "Making Experts required will mark 40 Applications incomplete"
    And confirming appends no OnePagerFacts event and blocks no edit

  Scenario: Un-requiring a built-in restores completeness with zero facts writes
    Given "Billing" is incomplete solely because the required built-in "Experts" is empty
    When an admin marks "Experts" optional
    Then viewing the "Billing" one-pager shows it complete on its next read
    And no OnePagerFacts stream received any event

  Scenario: Excluding a required built-in restores completeness with zero facts writes
    Given "Billing" is incomplete solely because the required built-in "Experts" is empty
    When an admin excludes "Experts" from the configuration
    Then viewing the "Billing" one-pager shows it complete on its next read
    And "Experts" no longer renders on the one-pager
    And no OnePagerFacts stream received any event

  Scenario: List indicators stay constant-query-count under pagination
    Given the Application configuration has at least one active required field, built-in or custom
    And more Applications exist than one page holds
    When I page through the application list
    Then every row on every page carries the correct completeness indicator
    And the per-page query count is constant and page size, ordering, and cursor are unchanged

  Scenario: Non-admins get no built-in requirement control
    Given a signed-in architect without the configuration write permission
    Then no built-in field carries a set-requirement affordance
    And a command to change a built-in field's requirement is rejected as forbidden
```

---

## Business Rules & Invariants

1. **Supersedes D7** — built-in fields may now carry a required flag and participate in
   completeness. Everything else 178 established is unchanged: completeness is evaluated at
   read time against the current configuration; it is never stored on nor derived from facts.
2. **Requirement flag on included built-ins only** — a configuration may mark a built-in field
   required only while that built-in is included; a command targeting an excluded (or
   unknown) built-in is rejected. The change is recorded as a new past-tense event
   `BuiltInFieldRequirementChanged`, mirroring `CustomFieldRequirementChanged`.
3. **Exclude retains the flag dormant** — parallel to custom-field retire/reactivate (175
   rule 6): excluding a built-in removes it from the display order and from completeness;
   re-including it restores its prior required flag (at the end of the display order).
4. **Extended completeness definition** — a subject is complete when every *active required*
   field — custom **and** included built-in — has a value on that subject.
5. **Built-in fill predicate** — a built-in value counts as filled when the subject's snapshot
   carries a present value for that catalog entry. List-valued built-ins (Experts today; a
   reference / reference-list kind per spec 188 later) are filled only when non-empty. An
   absent value — the em-dash case, which the adapters already encode as `nil` (spec 177
   Implementation Note 2) — counts as missing.
6. **Single-subject completeness from the snapshot** — the one-pager view evaluates required
   built-ins against the subject snapshot it already fetches; no extra query and no port
   change. A valueless required built-in renders "missing — required" instead of an em-dash.
7. **Set-based list completeness via a batched port method** — list indicators evaluate
   required-built-in fill for a page's subject IDs through a batched, set-based
   `BuiltInFieldSource` method implemented by composition-root adapters. Constant, bounded
   query count per page; per-subject snapshot fan-out (N+1) is prohibited.
8. **Impact preview via a population-wide port count** — for a built-in field, N = subjects of
   the type lacking a value for that field, obtained through a `BuiltInFieldSource` count
   method. `onepagers` never issues SQL against a supplier's tables. The preview is
   side-effect-free.
9. **Facts inviolable (D2)** — marking a built-in required, optional, or excluding it never
   mutates, archives, or blocks any subject or facts edit. Flipping a built-in required marks
   lacking subjects incomplete on their next read with **zero** writes; un-requiring or
   excluding restores completeness on the next read, also with zero writes.
10. **Indicator scope (extends 178 rule 10)** — a subject list shows a completeness indicator
    where a configuration has at least one active required field, custom **or** built-in.
11. **HATEOAS + permission** — the set-requirement affordance appears only on *included*
    built-in fields, and only to callers holding `PermMetaModelWrite`; reads use
    `PermMetaModelRead`. The UI gates the control on the link.
12. **No new persistence or cache (D6)** — no migration and no completeness cache. The
    built-in required flag lives in the event-sourced configuration document and defaults to
    optional; existing configurations need no backfill because no built-in was ever required
    before this slice.

---

## Acceptance Criteria

- [x] An admin can mark an *included* built-in field required and optional again; the change
      is recorded as `BuiltInFieldRequirementChanged` and replays correctly. Targeting an
      excluded or unknown built-in is rejected.
- [x] Excluding a required built-in retains its flag dormant; re-including restores it, at the
      end of the display order.
- [x] The one-pager view counts active required built-in fields in `requiredCount` /
      `filledCount` and lists a missing required built-in in `missingFields` (entry ID + label)
      — no second frontend call. The page renders such a field as "missing — required".
- [x] A populated required built-in counts as filled; an Experts (list-valued) built-in with
      an empty list counts as missing when required.
- [x] Subject list indicators reflect missing required built-ins alongside custom ones, with
      the extra evaluation done through a batched, set-based port method; cursor pagination,
      ordering, and page size are unchanged and the per-page query count stays constant.
- [x] The settings UI shows "Making <built-in field> required will mark <N> <subject type>s
      incomplete" before confirming, with N counted through the `BuiltInFieldSource` port; the
      preview endpoint is a side-effect-free GET gated by the configuration write permission.
- [x] Flipping a built-in required marks all subjects lacking a value incomplete on next read,
      appends no facts events, and blocks no edit; un-requiring or excluding restores
      completeness on next read, also with zero facts writes.
- [x] `onepagers` still contains no supplier application imports and no cross-schema SQL; the
      new port methods and their composition-root adapters keep the architecture boundary test
      green.
- [x] No migration, no new tables, no completeness cache exist in this slice (D6).
- [ ] Every BDD scenario above has a corresponding automated test; every modified file scores
      10.0 in CodeScene per `easi-codehealth`.

---

## Architecture

### Ownership

`onepagers` owns the change: the configuration aggregate gains a built-in requirement command
and event; the completeness query and impact preview extend to built-ins; two new methods land
on the `BuiltInFieldSource` port. Supplier contexts (`capabilitymapping`,
`enterprisearchitecture`, `architecturemodeling`) are touched only through those ports and
their composition-root adapters. No supplier or `metamodel` domain logic changes.

### Domain Model

No new aggregate. `OnePagerConfiguration` gains:

- A **per-included-built-in required flag**, stored parallel to `CustomFieldRecord.Required` —
  a small built-in selection record keyed by catalog entry ID carrying `required`. Inclusion
  remains represented by presence in the display order (175); the flag is field metadata, not
  ordering.
- Command `ChangeBuiltInFieldRequirement(entryID, required)` — guards that the entry is in the
  catalog and currently included (rule 2), emitting `BuiltInFieldRequirementChanged{entryID,
  required}`. Added to the event-type registry, the aggregate apply, and the projector.

Completeness stays a query-side concept. The single-subject completeness function extends to
iterate included required built-ins against the fetched snapshot using the rule-5 predicate;
its `MissingField` for a built-in carries the catalog entry ID and label.

### API Surface

- **Configuration command** — one endpoint to set a built-in field's requirement, gated by
  `PermMetaModelWrite`, mirroring the custom set-requirement command (optimistic concurrency
  via aggregate version, 409 on conflict). The configuration DTO's built-in field entries gain
  a `required` flag and an `x-set-requirement` link, present only when the field is included
  and the caller may write.
- **One-pager read (177/178)** — the `completeness` block's `requiredCount`, `filledCount`,
  and `missingFields` now include required built-ins. Response shape is unchanged; semantics
  broaden. Built-in field entries need no new response field — "missing" is conveyed by the
  entry's presence in `missingFields`, exactly as for custom fields.
- **Impact preview (178)** — extended to take a **field-kind discriminator** (built-in entry
  vs custom field) so it routes a built-in entry to the port count and a custom field to the
  facts count. For a built-in: N = population − subjects-with-a-value-for-that-built-in.
  Guarded by the configuration write permission; side-effect-free.

### Persistence

None new. No migration, no completeness cache (D6). The built-in required flag is part of the
event-sourced configuration document (JSONB), reprojected from events; existing configurations
project all built-ins as optional, which is the correct default, so no backfill is required
(no denormalized column sourced from another table is introduced). Completeness continues to
run over the existing `one_pager_facts` read model plus, for built-ins, the supplier read
models via ports, under the standard RLS tenancy.

### Frontend

- **Settings `/settings/one-pagers`** — the built-in field row gains a "Required" checkbox
  gated on `x-set-requirement`, wired to the new mutation; toggling it to required fetches and
  shows the impact preview (reusing the existing preview dialog) before confirming. The
  `BuiltInField` type gains `required` and the `x-set-requirement` link.
- **One-pager page** — the built-in field row renders "missing — required" when its entry ID
  is in `completeness.missingFields` (today only the custom branch does this); the completeness
  summary already reflects the broadened counts with no change.
- **Subject list surfaces** — no change: the per-row indicator is a backend-computed boolean
  that now also accounts for built-ins.

### Cross-Context Integration

No new events. Two additions to the `BuiltInFieldSource` port, both implemented by the
composition-root adapters over supplier read models (never cross-schema SQL in `onepagers`):

1. **Batched fill** — for a page's subject IDs and a set of required built-in entry IDs, which
   entries are filled per subject. The adapter fetches the page's subjects set-based and reuses
   the **same** snapshot mapping and fill predicate as the single-subject view (the single
   catalog-to-contract binding point, 177 rule 3, is preserved). Bounded, constant query count
   per page regardless of page size.
2. **Population fill count** — for one built-in entry, the count of subjects of the type that
   have a value for it — a single set-based COUNT, mirroring the port's existing
   `CountSubjects` and the facts read model's `CountSubjectsWithValue`.

`CompletenessIndicators` gains the subject-type's `BuiltInFieldSource` as a dependency (wired
where `onePagerCompletenessFor(indicators, subjectType)` already binds a subject type at the
composition root).

---

## Design Decisions

1. **Supersede D7 (product-owner confirmed)** — built-in fields may be required and participate
   in completeness with full parity to custom fields. Rationale: the most decision-relevant
   facts are often built-in, and excluding them made completeness under-report data quality.
   The design doc's decision log (D7) must be annotated during implementation to record this
   reversal — the design doc is **not** edited by this spec. Alternative: keep D7 and force
   admins to re-model built-ins as custom fields (rejected — duplicates domain data and breaks
   the single interleaved field model).
2. **Requirement flag stored parallel to the custom flag; new `BuiltInFieldRequirementChanged`
   event** — mirrors the closed-event, past-tense pattern of `CustomFieldRequirementChanged`.
   Alternative: overload the display-order `FieldRef` with a required flag (rejected —
   requiredness is field metadata, not ordering, and it would muddy reorder equality).
3. **Only included built-ins can be required; exclude retains the flag dormant** — consistent
   with custom retire/reactivate (175 rule 6). Alternative: reset to optional on exclude
   (rejected — inconsistent with the established retire semantics and surprises the admin who
   re-includes).
4. **Single-subject completeness reads the snapshot; no port change** — the view already
   fetches the full subject snapshot, so built-in fill is a pure in-memory extension of the
   existing completeness function. Alternative: route the single view through the new batched
   port too (rejected — a needless extra query when the data is already in hand).
5. **List indicators scale via a batched, set-based port method — not per-subject snapshot
   reads** — the alternative of calling `FetchSubject` per row is a per-entity fan-out (N+1)
   that violates 178 rule 6; joining supplier tables into the facts/list query is cross-schema
   coupling forbidden by D8. The batched method keeps the page cost to one facts query plus one
   bounded port call, reusing the single-subject snapshot mapping so the catalog binding stays
   in one place. This is the crux decision.
6. **Impact preview built-in count via a population-wide port COUNT** — mirrors the facts
   `CountSubjectsWithValue` and the port `CountSubjects`, and keeps the "no cross-schema SQL"
   guard (design doc Risks & Guards) intact. Alternative: raw cross-schema COUNT in `onepagers`
   (rejected — guarded by the architecture boundary test).
7. **Fill predicate defined per value-kind, forward-compatible with spec 188** — present =
   filled for Text/Date/Maturity; non-empty = filled for Experts and any future
   reference/reference-list kind. Alternative: a blanket non-nil check (kept as the adapter
   backstop, since adapters already map empty → nil, but the structural predicate is the
   contract so an accidentally non-nil empty list still counts as missing).
8. **No completeness cache (D6 upheld)** — the added cost is one bounded port call per page and
   one COUNT per preview; a cache projector remains the drop-in escape hatch if profiling ever
   demands it. Alternative: build the `one_pager_completeness` cache now (rejected —
   speculative infrastructure ahead of evidence).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Built-in fill through a batched port method | List completeness now needs one extra bounded call per page beyond the facts query | Set-based and page-scoped (≤ 100 IDs); reuses the single-subject snapshot mapping; constant query count guarded by an integration test |
| Two new `BuiltInFieldSource` methods across six adapters | More surface on the port and its adapters | Both mirror existing methods (`CountSubjects`, snapshot mapping); the catalog binding stays the single place that knows entry → supplier field |
| Requirement flag on built-ins | The completeness query shape and configuration events change (the very reason D7 deferred it) | Change is additive and event-sourced; existing configs default to optional; facts remain untouched (D2) |
| Read-time evaluation, no cache (D6) | Every list page and one-pager read recomputes built-in completeness | Bounded set-based queries; cache-projector precedent is a drop-in later |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off
