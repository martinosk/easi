# 188 — Relations As Built-In Fields

> **Status:** pending
> **Depends on:** [175 — OnePagerConfiguration](175_OnePagerConfiguration_done.md), [176 — OnePagerFacts capture](176_OnePagerFacts_done.md), [177 — OnePagerView](177_OnePagerView_done.md)
> **Relates to:** 186 — Mandatory built-in fields (this spec defines the "filled" predicate 186's built-in completeness consumes)
> **Design:** extends slice 3 of [Configurable One-Pagers](../docs/specs/configurable-one-pagers.md); conforms to decisions D6, D8, D10

---

## Problem Statement

The one-pager (spec 177) renders a subject's own attributes — name, description, maturity,
experts — but not the subject's **place in the architecture**. A capability's most
stakeholder-relevant facts are often relational: which applications realize it, which
business domains own it, what it depends on, its parent and children. An application's are
who built it, which vendor it was bought from, and which capabilities it serves. Today none
of this reaches the fact sheet, so a one-pager describes an entity in isolation rather than
in context.

The relations already exist in the model and are exposed by each owning context's read
models. What is missing is a way to put them on the one-pager. Built-in fields are already
include/exclude-selectable per subject type (spec 175), so the configuration side needs
nothing new — an admin should simply be able to opt a **relation** onto a subject type's
one-pager the same way they opt in "Maturity" today, and the one-pager should render the
related entities, linked to their own one-pagers.

The single obstacle is representational: the built-in field value union (spec 177) is closed
over four scalar/collection-of-scalar kinds — Text, Date, Maturity, Experts — with **no kind
that represents a reference to another entity**. This slice introduces that kind and wires
each subject type's available relations into the catalog, the composed read, and the page.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Tenant Administrator** | Opt a relation (e.g. "Realizing Applications", "Built By") onto a subject type's one-pager, ordered among the other fields, with no new configuration concepts to learn |
| **Enterprise Architect** | Open a subject's one-pager and see its architectural context — related entities by name, each a link to that entity's own one-pager |
| **Stakeholder** | Read who and what a subject connects to without navigating EASI's detail panels |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Relations as built-in one-pager fields

  Scenario Outline: The catalog exposes each subject type's relations as built-in fields
    Given the Built-in Field catalog for subject type "<subject type>"
    Then it lists the relation entry "<relation label>" as an includable built-in field

    Examples:
      | subject type          | relation label          |
      | Capability            | Realizing Applications  |
      | Capability            | Business Domains        |
      | Capability            | Parent Capability       |
      | Capability            | Child Capabilities      |
      | Capability            | Depends On              |
      | Enterprise Capability | Included Capabilities   |
      | Application           | Realized Capabilities   |
      | Application           | Built By                |
      | Application           | Purchased From          |
      | Application           | Acquired Via            |
      | Application           | Triggers / Serves       |
      | Acquired Entity       | Applications            |
      | Vendor                | Applications            |
      | Internal Team         | Applications            |

  Scenario: An administrator includes a relation built-in field
    Given the Capability One-Pager Configuration with "Realizing Applications" excluded
    When the administrator includes "Realizing Applications"
    Then it appears at the end of the display order as an included built-in field
    And no new configuration command or field-type concept is introduced

  Scenario: The one-pager renders related entities in the configured order
    Given a Capability realized by "Billing Service" and "Order Service"
    And a configuration ordering "Description", "Realizing Applications", "Maturity"
    When I open the Capability's one-pager
    Then "Realizing Applications" renders between "Description" and "Maturity"
    And it lists "Billing Service" and "Order Service"
    And each links to that application's one-pager

  Scenario: A relation with no related entities renders an em-dash
    Given a Capability with no realizing applications
    And a configuration that includes "Realizing Applications"
    When I open the one-pager
    Then "Realizing Applications" renders an em-dash in place of a value

  Scenario: A relation with non-denormalized counterpart names still renders labels
    Given a Capability whose parent and dependency counterparts are stored by ID only
    And a configuration that includes "Parent Capability" and "Depends On"
    When I open the one-pager
    Then each related capability renders by name, not by identifier

  Scenario: A reference to a non-subject entity renders as plain text
    Given a Capability assigned to business domain "Payments"
    And a configuration that includes "Business Domains"
    When I open the one-pager
    Then "Payments" renders as a label with no link

  Scenario: Constant query count holds for a many-relation configuration
    Given a configuration that includes every relation built-in field for Capability
    And a Capability with many children, many realizing applications, and many dependencies
    When the one-pager is rendered
    Then the number of database queries equals that of a Capability with few related
      entities under the same configuration
    And it stays within the documented ceiling

  Scenario: Catalog-contract test covers every relation entry
    Given the Built-in Field catalog of each subject type
    Then a per-subject-type integration test asserts every relation entry resolves against
      the supplier read model, failing the build on supplier drift

  Scenario: A required relation with zero relations is missing (interplay with spec 186)
    Given a relation built-in field marked required under spec 186
    And a subject with no related entities for that relation
    When completeness is evaluated
    Then the field counts as missing
```

---

## Business Rules & Invariants

1. **Reference-list value kind** — a new `BuiltInFieldValue` kind, `ReferenceListValue`, holds
   an ordered list of `Reference{ID, Label, SubjectType}`. It is the only reference-carrying
   kind; it represents both 0..N relations and 0..1 origin relations (a list of length ≤ 1).
   `Label` is always a human-readable name; `SubjectType` is the referenced entity's one-pager
   subject type, or empty when the referent is not a one-pager subject.
2. **Relations are read-only built-ins** — relation fields are sourced from the owning contexts
   at query time and are never edited on the one-pager; they are managed on the entity. They
   carry no custom-field concepts (no required flag from spec 175, no options, no value
   capture). They are ordinary catalog entries, so spec 175's include/exclude/reorder
   machinery governs them with no new configuration command.
3. **Ports only (D8)** — relation data flows exclusively through the `BuiltInFieldSource` port.
   Composition-root adapters wrap the supplier relation read models (and the enterprise
   `CompositionService`); `onepagers` imports no supplier package and issues no cross-schema
   SQL. The architecture boundary test from spec 175 stays green.
4. **Bounded, name-resolving adapters** — each included relation entry costs at most one
   bounded (indexed) edge read plus at most one batched counterpart-name lookup
   (`id IN (...)`, never per-edge). Adapters resolve ID-only counterpart names via the
   supplier read model so every reference renders a label, never a raw identifier.
5. **Constant query count (extends 177 rule 1)** — rendering one one-pager stays independent
   of the number of configured custom fields **and** of the number of related entities per
   relation; it grows only by a bounded constant per *included* relation field, itself bounded
   by the fixed per-subject-type catalog. No per-field or per-entity fan-out. Guarded by the
   extended constant-query-count integration test with a fixed documented ceiling.
6. **Interleaved order (D10)** — relation fields participate in the configuration's single
   interleaved display order alongside scalar built-ins and custom fields.
7. **Empty renders em-dash (177 rule 11)** — a relation with zero related entities (empty
   reference list) renders an em-dash. A relation is *filled* iff its reference list is
   non-empty.
8. **Filled semantics compose with spec 186** — a relation built-in marked required (under
   spec 186's mandatory-built-in mechanism) is satisfied iff it is filled per rule 7, i.e. it
   has at least one related entity. This spec owns the *filled* predicate; spec 186's built-in
   completeness consumes it.
9. **Catalog-contract guard (extends 177 rule 4)** — one catalog-contract integration test per
   new relation entry per subject type asserts the entry resolves against the supplier read
   model; supplier drift breaks the build, not a tenant's one-pager.
10. **Non-subject references render unlinked** — a reference whose `SubjectType` is empty (e.g.
    a business domain, which is not a one-pager subject type) renders as a plain label.

---

## Acceptance Criteria

- [x] `ports.BuiltInFieldValue` gains a `ReferenceListValue` kind carrying
      `[]Reference{ID, Label, SubjectType}`; the sealed union otherwise unchanged.
- [x] The wire DTO (`one_pager_view_dto.go`) and frontend (`types.ts` `BuiltInValue` union +
      `BuiltInValueDisplay.tsx`) gain a `references` case rendering the labels, each linked to
      the referent's one-pager when `SubjectType` is a one-pager subject type and unlinked
      otherwise.
- [x] The catalog (`catalog.go`) gains the relation entries enumerated in the Domain Model per
      subject type; each is included/excluded/reordered through the existing spec-175 endpoints
      with no new command.
- [x] Composition-root adapters populate every included relation entry via bounded read-model
      calls, resolving ID-only counterpart names (parent capability, capability dependencies,
      realized capabilities, component relations) through the supplier read model.
- [x] Rendering a one-pager whose configuration includes relation fields interleaves them in
      the configured display order; empty relations render em-dash; filled relations list
      related entities by name.
- [x] The constant-query-count integration test is extended to a many-relation configuration
      and asserts query count is independent of related-entity count and within the new
      documented ceiling.
- [x] One catalog-contract integration test exists per new relation entry per subject type and
      fails when an entry no longer resolves against the supplier read model.
- [x] The spec-175 architecture boundary test stays green: no `internal/onepagers` import of a
      supplier package; all relation data flows through ports with adapters at the composition
      root.
- [x] A relation's *filled* predicate (non-empty reference list) is exposed for spec 186's
      built-in completeness; a required relation with zero related entities counts as missing.
- [x] The bounded-context canvas `docs/architecture/OnePagers.md` is updated with the newly
      consumed supplier relation read contracts and the enterprise `CompositionService`.
- [x] Every BDD scenario above has a corresponding automated test; every modified file scores
      10.0 in CodeScene per `easi-codehealth`.

---

## Architecture

### Ownership

`onepagers` owns the new value kind, the extended catalog, and the composed read's handling
of reference values. The catalog-to-supplier binding for relations lives, like all supplier
coupling, in the composition-root adapters (`internal/infrastructure/api`). Supplier contexts
(`capabilitymapping`, `architecturemodeling`, `enterprisearchitecture`) are unchanged except
where a **batched name lookup** (`GetByIDs`) is needed to resolve ID-only counterparts — an
additive method on the supplier's own read model. No supplier detail endpoints change.

### Domain Model

**New value kind** (`onepagers/application/ports/builtin_fields.go`):

```
Reference          { ID, Label, SubjectType string }
ReferenceListValue { References []Reference }   // implements BuiltInFieldValue
```

`SubjectType` is one of the six one-pager subject types (for deep-linking) or empty. One list
kind covers every relation: 0..1 origin relations are a list of length ≤ 1.

**New catalog entries** (`onepagers/domain/catalog/catalog.go`), each a relation whose value is
a `ReferenceListValue`. "Names denormalized" indicates whether the supplier read model already
carries counterpart names (No ⇒ the adapter resolves them via a batched lookup):

| Subject Type | Entry ID | Label | Supplier read source | Names denormalized |
|---|---|---|---|---|
| Capability | `realizing-applications` | Realizing Applications | `RealizationReadModel.GetByCapabilityID` | Yes (`ComponentName`) |
| Capability | `business-domains` | Business Domains | `DomainCapabilityAssignmentReadModel.GetByCapabilityID` | Yes (`BusinessDomainName`) — **unlinked** (not a subject type) |
| Capability | `parent-capability` | Parent Capability | `CapabilityDTO.ParentID` + batched `CapabilityReadModel` | No — resolve |
| Capability | `child-capabilities` | Child Capabilities | `CapabilityReadModel.GetChildren` | Yes (`Name`) |
| Capability | `depends-on` | Depends On | `DependencyReadModel.GetOutgoing` + batched `CapabilityReadModel` | No — resolve |
| Enterprise Capability | `included-capabilities` | Included Capabilities | `CompositionService.CompositionForEC` | Yes (`Name`) |
| Application | `realized-capabilities` | Realized Capabilities | `RealizationReadModel.GetByComponentID` + batched `CapabilityReadModel` | No — resolve (`CapabilityID` only) |
| Application | `built-by` | Built By | `BuiltByRelationshipReadModel.GetByComponentID` | Yes (`InternalTeamName`) |
| Application | `purchased-from` | Purchased From | `PurchasedFromRelationshipReadModel.GetByComponentID` | Yes (`VendorName`) |
| Application | `acquired-via` | Acquired Via | `AcquiredViaRelationshipReadModel.GetByComponentID` | Yes (`AcquiredEntityName`) |
| Application | `component-relations` | Triggers / Serves | `ComponentRelationReadModel.GetBySourceID` + batched `ApplicationComponentReadModel` | No — resolve (`TargetComponentID` only) |
| Acquired Entity | `acquired-applications` | Applications | `AcquiredViaRelationshipReadModel.GetByEntityID` | Yes (`ComponentName`) |
| Vendor | `purchased-applications` | Applications | `PurchasedFromRelationshipReadModel.GetByVendorID` | Yes (`ComponentName`) |
| Internal Team | `built-applications` | Applications | `BuiltByRelationshipReadModel.GetByTeamID` | Yes (`ComponentName`) |

`Reference.SubjectType` per entry: application-valued relations → `application`; capability
relations → `capability`; included-capabilities → `capability`; the three reverse-origin
"Applications" entries → `application`; business-domains → empty (unlinked). Directional
relations (`depends-on`, `component-relations`) expose the outgoing direction; the incoming
methods are intentionally not bound.

No new aggregates or events — this remains a read-only extension of the composed one-pager
query (spec 177). Completeness of relation fields is left to spec 186; this spec supplies only
the *filled* predicate (non-empty reference list).

### API Surface

No new endpoints and no contract-shape change to `GET /api/v1/one-pagers/{subjectType}/{subjectID}`.
`BuiltInValueDTO` gains a discriminated `"references"` variant (`references: [{id, label,
subjectType}]`), additive alongside the existing `text`/`date`/`maturity`/`experts` variants.
Configuration endpoints (spec 175) are unchanged — relation entries flow through the existing
include/exclude/reorder commands.

### Persistence

None. Query-time composition, no cache (D6). Relations are read through the supplier read
models via ports; RLS tenancy applies unchanged. A supplier read model may gain an additive
batched `GetByIDs` method for name resolution; no schema change.

### Frontend

- `types.ts`: extend the `BuiltInValue` union with
  `BuiltInReferencesValue { type: 'references'; references: { id: string; label: string;
  subjectType?: OnePagerSubjectType }[] }`.
- `BuiltInValueDisplay.tsx`: add a `case 'references'` rendering the labels as a list; each
  label links to the referent's one-pager (via the existing one-pager route) when `subjectType`
  is set, and renders as plain text otherwise. Empty list is handled by the existing null →
  em-dash branch. Mantine v8 primitives only.

### Cross-Context Integration

Customer/Supplier, query-time, no new events. The composition-root adapters gain dependencies
on the supplier relation read models above and on the enterprise `CompositionService` (an
application service, imported at the composition root — outside `onepagers` — exactly like the
existing read-model wiring). `onepagers` sees only resolved `Reference` values.

---

## Design Decisions

1. **One `ReferenceListValue` kind, not per-relation kinds** — a single list of
   `{ID, Label, SubjectType}` represents every relation, including 0..1 origin relations (a
   list of length ≤ 1). Keeps the sealed union minimal and the wire/frontend switch to one new
   case. Alternative: distinct single-reference and multi-reference kinds (rejected —
   two cases for one concept; 0..1 is just a bounded list).
2. **Relations are built-in, not a new editable field type** — they are model-owned and
   read-only on the one-pager, so they reuse spec 175's include/exclude/reorder machinery with
   zero configuration-side change. Alternative: a "relation" custom-field type with capture
   (rejected — relations are not captured on the one-pager; they are the entity's own edges).
3. **Config-gated, bounded, name-resolving adapters (the query-count crux)** — the composed
   read resolves the configuration first, then fetches the subject snapshot for the *included*
   built-in entries only. Scalar entries stay free (one base-DTO read); each **included**
   relation entry costs at most one indexed edge read plus at most one **batched** counterpart-
   name lookup (`id IN (...)`), never a per-edge lookup. Total relation reads are therefore
   bounded by the number of included relation entries (bounded by the fixed catalog) and are
   independent of the number of related entities. Alternatives rejected: (a) eager resolution
   of every catalog relation regardless of inclusion — provably constant but makes a
   Name-only one-pager pay for every relation it does not display; (b) per-edge name lookups —
   an N+1 fan-out that violates the design's performance quality attribute.
4. **Resolve missing counterpart names rather than restrict the catalog** — parent capability,
   capability dependencies, realized capabilities, and component relations carry counterpart
   IDs only. The adapter resolves their names through the supplier read model (one batched
   lookup per relation) so every reference is human-readable. Alternative: expose only relations
   whose names are already denormalized (rejected — it would drop the four most contextually
   valuable capability/application relations to save one batched, bounded query each).
5. **Enterprise composition via its application service** — `Included Capabilities` binds to
   `CompositionService.CompositionForEC`, whose several internal statements execute inside one
   port call, constant per subject — the same shape as the capability subject read issuing
   three statements in one call (177 Implementation Note 6). Alternative: re-derive composition
   from raw read models inside the adapter (rejected — duplicates domain logic the service owns).
6. **Filled ⇔ non-empty; completeness deferred to 186** — this spec defines only the *filled*
   predicate for reference values and renders em-dash when empty; whether a relation is
   *required* and counts toward completeness is spec 186's mandatory-built-in mechanism.
   Alternative: add built-in completeness here (rejected — out of this slice; D7 keeps
   required-ness a separate concern, and 186 owns it).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Config-gated relation resolution | The composed read must resolve the configuration before the subject fetch and pass the included-entry set to the port | Small reorder in the query service; the configuration read already happens; each included relation stays one bounded call |
| Resolve ID-only counterpart names | Four relations pay a second batched read to render labels | One `id IN (...)` query per such relation, never per-edge; folded into the documented ceiling |
| New reference value kind touches every layer in lockstep | Port union, wire DTO, frontend union, and display must move together | The sealed union and exhaustive switches fail the build if a layer is missed; one added case each |
| Relation reads couple the composed read to more supplier read models | More supplier contracts the catalog binds to | One catalog-contract integration test per relation entry fails the build on drift; all binding is at the composition root |
| Batched `GetByIDs` added to some supplier read models | A small additive change in supplier packages | Additive read-only method, no schema change, covered by the catalog-contract test |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [ ] User sign-off
