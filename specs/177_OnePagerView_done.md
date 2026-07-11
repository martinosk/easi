# 177 — OnePagerView

> **Status:** done
> **Depends on:** [175 — OnePagerConfiguration](175_OnePagerConfiguration_done.md), [176 — OnePagerFacts capture](176_OnePagerFacts_done.md)

---

## Problem Statement

Specs 175 and 176 let a tenant admin configure a One-Pager per Subject Type and let
architects record typed Field Values on Subjects. What is still missing is the payoff:
the **One-Pager itself** — the rendered, stakeholder-facing fact sheet for one Subject.
Today the same information is scattered across detail panels, and the recorded custom
facts are visible only inside the edit section.

This slice delivers the composed read: a single endpoint that assembles the tenant's
One-Pager Configuration, the Subject's Field Values, and Built-in Field data sourced from
the owning contexts, plus a routed, deep-linkable page that renders it well enough to
present or screen-share with stakeholders who do not live in EASI. Slice 3 of the
approved design `docs/specs/configurable-one-pagers.md`.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | Open a presentable one-pager for any Subject from its detail panel and share its URL with stakeholders |
| **Stakeholder** | Follow a shared link and read a clean, current fact sheet without navigating EASI |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: One-Pager view

  Scenario Outline: Fully populated one-pager renders for every subject type
    Given a "<subject type>" Subject whose One-Pager Configuration selects built-in
      fields and defines custom fields
    And Field Values are recorded for every custom field
    When I open the Subject's one-pager
    Then I see a subject header with the Subject's name and subject type
    And every configured field renders with its value

    Examples:
      | subject type          |
      | Capability            |
      | Enterprise Capability |
      | Application           |
      | Acquired Entity       |
      | Vendor                |
      | Internal Team         |

  Scenario: Interleaved display order is respected
    Given a configuration ordering "Description" (built-in), "Contract link" (custom),
      "Maturity" (built-in)
    When I open the one-pager
    Then the fields render in exactly that order

  Scenario: Custom field values render per field type
    Given recorded values for a Link, a Contact Person, a Date, and a Selection field
    When I open the one-pager
    Then the Link renders as an anchor with its label
    And the Contact Person renders as a name, email, and company block
    And the Date renders formatted for display
    And the Selection renders its option label

  Scenario: Retired custom field does not render
    Given a Subject with a recorded value for a retired custom field
    When I open the one-pager
    Then the retired field does not appear

  Scenario: Value referencing a retired Selection option is flagged
    Given a recorded Selection value whose option has since been retired
    When I open the one-pager
    Then the option label renders visibly flagged as retired

  Scenario: Maturity built-in field renders with tenant scale semantics
    Given the tenant's maturity scale names the section covering value 85 "Optimizing"
    And a Capability with maturity 85 whose configuration includes the maturity built-in field
    When I open its one-pager
    Then the maturity field renders the value together with the section name "Optimizing"

  Scenario: Empty optional custom field renders as an em-dash
    Given a configured optional custom field with no recorded value
    When I open the one-pager
    Then that field renders an em-dash in place of a value

  Scenario: Share URL round-trip
    Given I am viewing a Subject's one-pager
    When I use "Share (copy URL)" and open the copied URL in a fresh session
    Then the same one-pager renders

  Scenario: Entry point is gated on the HATEOAS link
    Given a Subject detail response containing a one-pager link
    Then the detail panel shows a "One-Pager" action
    And when the detail response carries no one-pager link the action is absent

  Scenario: Subject not found
    When I request the one-pager for a Subject that does not exist
    Then the API responds 404 and the page shows a not-found state

  Scenario: Read permission matches the Subject
    Given a user lacking the read permission of the Subject's own detail endpoint
    When they request the one-pager
    Then the API responds 403

  Scenario: Catalog entries resolve against supplier read contracts
    Given the Built-in Field catalog of each subject type
    Then a per-subject-type integration test asserts every catalog entry resolves
      against the supplier's read model, failing the build on supplier drift
```

---

## Business Rules & Invariants

1. **Constant query count** — rendering one one-pager performs one configuration read,
   one Field Values read, and one Subject read through the Built-in Field port; where the
   configuration includes maturity or strategy-pillar built-in fields not already carried
   by the Subject read, at most one additional metamodel semantics read. Never a
   per-field or per-entity fan-out.
2. **Ports only (D8)** — Built-in Field data is read exclusively through
   `BuiltInFieldSource` ports defined in `onepagers/application/ports` and implemented by
   composition-root adapters wrapping supplier read models. `onepagers` never imports a
   supplier context's application packages and never queries supplier tables directly.
3. **Single binding point** — each Built-in Field catalog entry (metadata per spec 175)
   binds to exactly one field of a supplier's published read contract, and the
   composition adapter is the only place that binding exists in code.
4. **Catalog-contract guard** — a mandatory integration test per subject type asserts
   every catalog entry resolves against the supplier read model.
5. **Interleaved order (D10)** — the response and the page present built-in and custom
   fields in the configuration's single interleaved display order.
6. **Retired fields hidden** — retired custom field definitions do not render; their
   recorded values remain untouched.
7. **Retired options flagged** — a value referencing a retired Selection option renders
   its label flagged as retired; it is never treated as invalid.
8. **Tenant semantics** — maturity and strategy-pillar built-in fields render using the
   tenant's configured maturity-scale sections and pillar definitions from `metamodel`,
   sourced through the same port mechanism.
9. **Subject's read permission** — the endpoint is authorized with the same
   published-language read permission that gates the Subject's own detail endpoint:
   `capabilities:read` (Capability), `enterprise-arch:read` (Enterprise Capability),
   `components:read` (Application, Acquired Entity, Vendor, Internal Team), enforced via
   the existing `RequirePermission` middleware resolved from the subject type.
10. **HATEOAS both ways** — Subject detail responses carry a link to their one-pager;
    the one-pager response links back to the Subject. The frontend entry point renders
    only when the link is present.
11. **Empty renders as em-dash** — a configured field with no value renders an em-dash;
    configured fields are never silently omitted, so the sheet layout is stable across
    Subjects.
12. **Query-time freshness (D6)** — the one-pager composes supplier data at read time;
    no denormalized one-pager cache table is introduced.

---

## Acceptance Criteria

- [x] `GET /api/v1/one-pagers/{subjectType}/{subjectID}` returns the subject header,
      and the configured fields — built-in values and custom Field Values interleaved in
      the configured display order — for all six subject types.
- [x] An integration test asserts the endpoint's constant query count (rule 1).
- [x] `onepagers` contains no imports of supplier application packages and no SQL against
      supplier tables; all Built-in Field data flows through `BuiltInFieldSource` ports
      with adapters at the composition root — enforced by the architecture boundary test
      introduced in spec 175, which this slice's port/adapter split must keep green.
- [x] The bounded-context canvas `docs/architecture/OnePagers.md` is updated with the
      consumed supplier read contracts and the metamodel upstream relationship.
- [x] One catalog-contract integration test exists per subject type and fails when a
      catalog entry no longer resolves against the supplier read model.
- [x] Retired custom fields are absent from the response and the page; retired Selection
      options render flagged.
- [x] Maturity and strategy-pillar built-in fields render with the tenant's configured
      semantics (strategy pillars vacuously — no catalog entry exists; see
      Implementation Notes 1).
- [x] All six Subject detail responses include the one-pager HATEOAS link; the one-pager
      response links back to the Subject; the endpoint enforces the Subject's read
      permission and returns 404 for unknown Subjects.
- [x] A routed one-pager page renders the sheet with per-field-type presentation (anchor,
      contact block, formatted date, selection label), em-dash for empty fields, Mantine
      components only, suitable for screen-sharing.
- [x] The page is deep-linkable: its share-URL generator is registered in
      `frontend/src/lib/deepLinks`, and "Share (copy URL)" copies a working absolute URL
      with the spec-113 toast behavior.
- [x] Each of the six detail panels shows a "One-Pager" action gated on the HATEOAS link.
- [x] Every BDD scenario above has a corresponding automated test.

---

## Architecture

### Ownership

`onepagers` owns the composed read endpoint, its read-side application service, and the
catalog-to-contract binding adapters (which live at the composition root,
`internal/infrastructure/api`, per the `direction_composition_adapters.go` precedent).
Supplier contexts (`capabilitymapping`, `enterprisearchitecture`, `architecturemodeling`)
are affected only by adding the one-pager link to their detail responses. `metamodel` is
an upstream supplier of rendering semantics.

### Domain Model

No new aggregates or events — this is a read slice over the `OnePagerConfiguration`
(spec 175) and `OnePagerFacts` (spec 176) read models. New in `onepagers/application`:
the composed one-pager query and the `BuiltInFieldSource` port contracts — one port per
subject type returning the subject header and every catalog entry's value in a single
read, plus the metamodel-facing semantics port for maturity-scale sections and strategy
pillars.

### API Surface

- `GET /api/v1/one-pagers/{subjectType}/{subjectID}` — `subjectType` is one of six
  identifiers covering Capability, Enterprise Capability, Application, Acquired Entity,
  Vendor, and Internal Team. Response: subject header (name, subject type), an ordered
  field list where each entry is discriminated built-in (catalog key, rendered value) or
  custom (definition metadata, typed value envelope, retired-option flag), and `_links`
  including the Subject. 404 for unknown subject type or Subject; 403 per rule 9.
- Existing Subject detail endpoints: response gains a one-pager link (additive).

### Persistence

None. Reads existing `onepagers` read-model tables and supplier read models via ports.
RLS tenancy applies unchanged through the existing read-model access paths.

### Frontend

- New routed page rendering the one-pager: subject header, fields in configured order,
  per-field-type presentation, retired-option flag, em-dash for empty values. Mantine v8
  primitives only; presentation quality suitable for stakeholders.
- Share-URL generator and route registration in `frontend/src/lib/deepLinks` (spec 113
  precedent); "Share (copy URL)" affordance with clipboard + toast behavior.
- "One-Pager" action on the six detail panels, rendered only when the detail response
  carries the one-pager link (HATEOAS gating per `easi-frontend-data`).
- Data fetching via TanStack Query keyed on subject type + ID.

### Cross-Context Integration

Customer/Supplier, all query-time, no new events: `onepagers` consumes supplier read
contracts exclusively through its own ports; adapters at the composition root wrap
`CapabilityReadModel`, `EnterpriseCapabilityReadModel`, `ApplicationComponentReadModel`,
`AcquiredEntityReadModel`, `VendorReadModel`, `InternalTeamReadModel`, and the
`metamodel` configuration read model.

---

## Design Decisions

1. **Query-time composition through ports, no cache (D6, D8)** — supplier names and
   values are always fresh with zero denormalization staleness; the adapter layer is the
   single supplier coupling point. Alternative: a denormalized one-pager cache projector
   (rejected now per D6 — drop-in later via the existing cache-projector precedent if
   profiling demands it).
2. **One endpoint with a subjectType segment, six ports behind it** — a single contract
   and page for all subject types keeps the frontend uniform; per-subject variation lives
   in the port adapters. Alternative: six per-context endpoints (rejected — six contracts
   to keep aligned, and the interleaved composition logic would be duplicated).
3. **Constant query count as a stated contract** — one configuration read, one facts
   read, one subject read via port, at most one metamodel semantics read; guarded by an
   integration test. Alternative: unbounded per-field lookups (rejected — violates the
   design's performance quality attribute).
4. **Interleaved order is server-composed (D10)** — the endpoint returns fields already
   merged in the configuration's single ordering so every consumer renders identically.
   Alternative: return built-in and custom lists separately and merge client-side
   (rejected — re-implements the ordering per consumer and invites drift).
5. **Em-dash for empty optional fields** — keeps the configured layout stable across
   Subjects and makes gaps visible when presenting. Alternative: omit empty fields
   (rejected — two Subjects of the same type would render structurally different sheets).
6. **Subject's own read permission, resolved from subject type** — a one-pager exposes
   exactly the Subject's data plus its Field Values, so it inherits the Subject's read
   gate; no new permission is introduced. Alternative: a dedicated one-pager permission
   (rejected — would let the one-pager reveal data the subject permission denies, or
   block data the user can already see).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Query-time composition (D6) | Every render pays supplier reads instead of one cache-row read | Constant indexed query count, guarded by test; cache projector precedent is a drop-in later |
| Catalog bound to supplier contracts in code | Supplier read-model changes can break the catalog | Mandatory per-subject-type catalog-contract integration tests fail the build, not the tenant's one-pager |
| Single dynamic-subject-type endpoint | Handler must dispatch permission and port per subject type | Dispatch table in one place at the composition root; unknown types 404 |
| Em-dash for empty fields | Sparse sheets show visible gaps | Gaps are the point for data quality; completeness indicators arrive in spec 178 |

---

## Implementation Notes

Decisions made at implementation start (2026-07-11), within the approved architecture:

1. **Strategy pillars are inapplicable in this slice** — the spec-175 catalog defines no
   strategy-pillar built-in entry on any subject type, so there is nothing to bind or
   render. The metamodel semantics port (`ports.MaturityScaleSource`) covers maturity
   sections only; a pillar catalog entry is a future spec that adds a sibling port
   method, adapter binding, and catalog-contract coverage. The pillar half of rule 8 and
   its acceptance criterion are therefore vacuous here — flagged for user sign-off.
2. **Port contract** (`onepagers/application/ports/builtin_fields.go`): one
   `BuiltInFieldSource` instance per subject type, selected via a map keyed by subject
   type. `FetchSubject(ctx, subjectID)` returns `nil, nil` for unknown subjects
   (mirrors supplier `GetByID` semantics). `SubjectSnapshot.Fields` is keyed by catalog
   entry ID; adapters populate **every** catalog entry key, storing `nil` for empty
   values — key presence is what the catalog-contract tests assert, and `nil` renders
   as the em-dash. Values are a sealed union: Text, Date, Maturity, Experts.
3. **Missing configuration falls back to the synthesized default** (every catalog
   built-in field in catalog order, no custom fields) without persisting anything — the
   composed read never issues commands; lazy creation remains the configuration GET's
   concern.
4. **Maturity semantics resolve in the query service**, not the adapter: the service
   requests tenant sections through the port only when a configured field carries a
   maturity value, then matches `MinValue <= value <= MaxValue`. The hardcoded section
   names on the capabilitymapping DTO are deliberately unused (rule 8). A tenant without
   metamodel configuration renders the bare maturity value.
5. **HATEOAS rels**: subject detail responses gain `x-one-pager` (GET); the one-pager
   response carries `self` and `x-subject` pointing at the subject's detail endpoint.
6. **Constant-query-count test** counts data statements through a wrapping SQL driver
   (new test-only helper; no precedent existed). The binding contract is asserted as:
   query count is identical for a minimal and a many-field configuration, with a fixed
   documented ceiling — note the capability subject read issues three statements
   (row, experts, tags) inside one port call, which is constant and allowed.
7. **Retired-option detection is consolidated** in
   `readmodels/custom_field_rules.go` (`CustomFieldRecord.RetiredOptionReferenced`);
   the facts DTO mapper and the composed query both delegate to it.
8. **Selection labels resolve server-side** (review finding): the facts read model's
   `display_text` column stores the option ID, so the composed query resolves the
   option label through `CustomFieldRecord.SelectionOptionLabel` at read time (D4:
   server-composed rendering). The write path and projector are unchanged.
9. **Date-only values render as local calendar dates** (review finding): the frontend
   parses `YYYY-MM-DD` with local-date components (`frontend/src/utils/date.ts`);
   parsing via `new Date(iso)` would shift the date a day back in negative-UTC-offset
   timezones.

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off
