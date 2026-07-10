# 175 — One-Pager Configuration

> **Status:** pending
> **Depends on:** _none_
> **Depended on by:** 176 — OnePagerFacts capture, 177 — OnePagerView, 178 — Completeness & impact preview
> **Design:** slice 1 of [Configurable One-Pagers](../docs/specs/configurable-one-pagers.md); conforms to decisions D1–D10

---

## Problem Statement

Architects need to communicate the current state of the architecture to stakeholders who do
not live in EASI. A **One-Pager** is a tenant-configurable fact sheet for a single subject
entity, and today nothing lets a tenant decide what such a fact sheet should contain: which
facts from the existing domain model matter to their stakeholders, and which
stakeholder-facing facts (a contact person, a contract link) are missing from the model
entirely.

This slice delivers the configuration side: a new `onepagers` bounded context whose
**One-Pager Configuration** aggregate lets a tenant administrator shape, per **Subject
Type** (Capability, Enterprise Capability, Application, Acquired Entity, Vendor, Internal
Team), which **Built-in Fields** the one-pager shows and which **Custom Field Definitions**
extend it — before a single fact is captured. It is independently deployable: an admin can
fully configure one-pagers for their organization with this slice alone.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Tenant Administrator** | Decide per subject type which fields a one-pager shows, add typed custom fields matching the organization's vocabulary, and evolve that configuration safely over time. |
| **Enterprise Architect** | Trust that the configured field set is stable and well-typed, so the facts captured against it stay valid as the configuration evolves. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: One-Pager Configuration

  Scenario: First read lazily creates the default configuration
    Given no One-Pager Configuration has ever been saved for subject type "Vendor"
    When an administrator opens the Vendor tab of the One-Pagers settings section
    Then a configuration is shown containing every catalog Built-in Field for Vendor
      in catalog order
    And it contains no Custom Field Definitions
    And subsequent reads return that same configuration

  Scenario Outline: An administrator defines a custom field of each field type
    Given the One-Pager Configuration for subject type "Application"
    When the administrator defines a custom field "<name>" with field type "<type>"
    Then "<name>" is listed as an active Custom Field Definition of type "<type>"
      with a stable FieldID
    And "<name>" appears at the end of the display order

    Examples:
      | name              | type           |
      | Business summary  | Text           |
      | Annual cost       | Number         |
      | Contract renewal  | Date           |
      | Contract          | Link           |
      | Hosting model     | Selection      |
      | Product owner     | Contact Person |

  Scenario: Renaming a custom field preserves its identity
    Given an active custom field "Contract" on the Application configuration
    When the administrator renames it to "Contract link" and updates its help text
    Then the field keeps its FieldID, type, required flag, and display position
    And it is listed as "Contract link"

  Scenario: Retiring and reactivating a custom field
    Given an active custom field "Contract link" on the Application configuration
    When the administrator retires it
    Then it disappears from the display order and is listed under retired fields
    When the administrator reactivates it
    Then it is active again with its original FieldID, type, and options
    And it appears at the end of the display order

  Scenario: Reordering interleaves built-in and custom fields
    Given the Application display order is [Description, Experts, Contract link]
    When the administrator moves "Contract link" between "Description" and "Experts"
    Then the display order is [Description, Contract link, Experts]

  Scenario: Excluding and including a built-in field
    Given the built-in field "Experts" is included on the Application configuration
    When the administrator excludes it
    Then "Experts" leaves the display order but remains available in the catalog list
    When the administrator includes it again
    Then it reappears at the end of the display order

  Scenario: Changing the requirement flag on a custom field
    Given an active optional custom field "Product owner"
    When the administrator marks it required
    Then the configuration records "Product owner" as required
    And no recorded data is validated or blocked by this change

  Scenario: Selection option lifecycle
    Given a Selection field "Hosting model" defined with options "On-prem" and "Cloud"
    When the administrator adds option "Hybrid" and retires option "On-prem"
    Then the field definition lists "Cloud" and "Hybrid" as active options
    And "On-prem" remains on the definition marked retired

  Scenario: Field type changes are rejected
    Given an active custom field "Contract link" of type Link
    When a request attempts to alter its field type to Text
    Then the request is rejected with an error explaining that field types are
      immutable and the path is retire-and-redefine
    And the field remains of type Link

  Scenario: Only administrators can change a configuration
    Given a signed-in architect without the tenant-configuration write permission
    Then the configuration response carries no HATEOAS affordances for changing it
    And any write command they submit is rejected as forbidden
```

---

## Business Rules & Invariants

1. **One configuration per (tenant, subject type)** — enforced in the command handler via
   read-model lookup, with a database unique constraint on `(tenant_id, subject_type)` as
   backstop. The aggregate's identity is its own intrinsic UUID.
2. **Lazy default creation** — the first read for a (tenant, subject type) creates the
   configuration with a default: every catalog built-in field included in catalog order,
   no custom fields. Creation is idempotent.
3. **Catalog-bounded built-ins** — a configuration may only include built-in fields from
   the per-subject-type catalog owned by `onepagers`. The catalog is code, not data.
4. **Field identity is the FieldID** — a stable UUID assigned at definition. Renaming a
   field (display name, help text) is free and never changes identity.
5. **Field type is immutable** — no operation changes a defined field's type; the
   documented path is retire the old field and define a new one (design doc D4).
6. **Fields are retired, never deleted** — a retired field can be reactivated with its
   FieldID, type, required flag, and options intact. Reactivated and newly included
   fields enter the display order at the end.
7. **One interleaved display order** — the configuration holds a single ordering over the
   mixed set of included built-in and active custom fields (design doc D10). Every such
   field appears exactly once; retired and excluded fields are not in the order.
8. **Required flag exists only on Custom Field Definitions** — built-in fields carry no
   required flag (design doc D7). Requirement changes never validate, mutate, or block
   any recorded data.
9. **Selection options have stable identity and are retire-only** — a Selection field is
   defined with at least one option; options are identified by OptionID, can be added and
   retired, and retired options remain on the definition.
10. **Active display names are unique** — case-insensitive, across the active custom
    fields and included built-in field labels of one configuration.
11. **Every change is a past-tense domain event** — `OnePagerConfigurationCreated`,
    `CustomFieldDefined`, `CustomFieldRenamed`, `CustomFieldRequirementChanged`,
    `CustomFieldRetired`, `CustomFieldReactivated`, `BuiltInFieldIncluded`,
    `BuiltInFieldExcluded`, `OnePagerFieldsReordered`, `SelectionOptionAdded`,
    `SelectionOptionRetired`. Replay reconstructs the full configuration.
12. **Writes are admin-only** — gated by `PermMetaModelWrite` (the existing
    tenant-configuration settings gate); reads by `PermMetaModelRead`. HATEOAS advertises
    write operations only to authorized callers.

---

## Acceptance Criteria

- [ ] Opening a subject type's settings tab with no prior configuration shows the default
      (all catalog built-ins, catalog order, no custom fields) and persists it lazily
- [ ] An admin can define a custom field of each of the six field types: Text, Number,
      Date, Link, Selection, Contact Person
- [ ] An admin can rename a custom field and edit its help text without changing its
      FieldID, type, required flag, or position
- [ ] An admin can retire a custom field and later reactivate it with definition intact;
      retired fields never disappear from the aggregate's history or read model
- [ ] An admin can reorder the single interleaved display order, placing custom fields
      between built-in fields
- [ ] An admin can exclude and re-include catalog built-in fields
- [ ] An admin can toggle the required flag on custom fields only; the change is recorded
      and has no effect beyond the configuration
- [ ] An admin can add and retire Selection options; retired options remain on the
      definition
- [ ] Any attempt to change a field's type is rejected with a clear error
- [ ] A second configuration for the same (tenant, subject type) cannot be created:
      handler-level rejection with the DB unique constraint as backstop
- [ ] Non-admin users receive no write affordances (HATEOAS) and their write commands
      return 403; the UI gates its controls on the links
- [ ] Every configuration change is persisted as its named past-tense event; replay
      reconstructs the configuration
- [ ] A bounded-context canvas exists at `docs/architecture/OnePagers.md` following the
      established canvas format (purpose, strategic classification, domain roles, inbound/
      outbound communication, relationship types, ubiquitous language), and the context map
      in `docs/architecture/README.md` includes `onepagers`
- [ ] An architecture test fails the build when `internal/onepagers` imports any other
      bounded context's packages — allowed imports are `internal/shared/**`, other contexts'
      `publishedlanguage` packages, and third-party/stdlib
- [ ] Every BDD scenario above has at least one corresponding test
- [ ] Every modified file scores 10.0 in CodeScene per `easi-codehealth`

---

## Architecture

### Ownership

A new supporting bounded context `onepagers` (backend: `backend/internal/onepagers`,
frontend: settings components under `features/settings/`). It owns the One-Pager
Configuration aggregate and the built-in field catalog. It does not own subject entities;
it names them only through a `SubjectType` enumeration scoped to this context. No other
context is modified.

### Domain Model

`OnePagerConfiguration` — one aggregate per (tenant, subject type), event-sourced like
`MetaModelConfiguration`, holding:

- **Built-in field selections** — references into the per-subject-type catalog.
- **Custom Field Definitions** — entities within the aggregate: FieldID (UUID), display
  name, field type (Text, Number, Date, Link, Selection, Contact Person), required flag,
  help text, options (Selection only, each with OptionID, label, active/retired status),
  and active/retired status.
- **One interleaved display order** over the mixed built-in + custom field set.

Field types are definition metadata in this aggregate; the aggregate produces the eleven
events in rule 11 and consumes none.

**Built-in field catalog** — code-owned by `onepagers`: per subject type, a list of
entries (stable entry ID, display label). In this slice entries are definition-only
metadata; nothing reads supplier data. Initial catalog:

| Subject Type | Initial entries |
|---|---|
| Capability | Name, Description, Maturity, Experts |
| Enterprise Capability | Name, Description, Category |
| Application | Name, Description, Experts |
| Acquired Entity | Name, Acquisition Date, Integration Status |
| Vendor | Name, Implementation Partner, Notes |
| Internal Team | Name, Department, Contact Person |

Adding a subject type means: add it to the `SubjectType` enumeration, add its catalog
entries, and add its settings tab — the enumeration is the single compile-visible
extension point.

### API Surface

REST under `/api/v1/one-pagers/`, per `easi-api-standards` (HATEOAS links, swagger
annotations, error registry). Contract level:

- **Get configuration** for a subject type — lazily materializes the default on first
  read; gated by `PermMetaModelRead`.
- **Commands**, one endpoint per configuration change, gated by `PermMetaModelWrite`:
  define custom field, rename custom field, change requirement flag, retire custom field,
  reactivate custom field, include built-in field, exclude built-in field, reorder
  fields, add Selection option, retire Selection option.
- Optimistic concurrency via the aggregate version, surfacing conflicts as 409
  (metamodel settings precedent).

### Persistence

Event-sourced in the shared event store. New PostgreSQL schema `onepagers` (schema
creation and grants shaped per migrations 103/115). One read-model table
`onepagers.one_pager_configurations`: one row per (tenant, subject type) with `tenant_id`,
`subject_type`, the configuration document, and version; unique constraint on
`(tenant_id, subject_type)`; RLS policy identical in shape to migration 076
(`tenant_id = current_setting('app.current_tenant', true)` for `easi_app`). Migration
number: 122 (next after 121).

### Frontend

New "One-Pagers" settings section at `/settings/one-pagers` (nav link + route in
`SettingsPage`, per the maturity-scale/strategy-pillars precedent), with one tab per
subject type. Each tab shows the interleaved field list in display order with a
built-in/custom distinction, reorder controls (drag or up/down), inline rename, required
toggle (custom fields only), retire action, a retired-fields list with reactivate, a
Selection options editor, and an add-field form built with React Hook Form +
`zodResolver`. Data via TanStack Query with `staleTime: Infinity` and invalidation on
every mutation; query keys registered in the settings `queryKeys`; all write controls
gated on HATEOAS links per `easi-frontend-data`.

### Cross-Context Integration

None. In this slice `onepagers` reads nothing from other contexts and publishes its
events only to its own projector; the catalog is definition-only metadata. Supplier data
flows only through consumer-defined ports and composition-root adapters (design doc D8),
which land with the composed one-pager read, not with configuration.

### Boundary Enforcement & Documentation

- **Architecture test** — a Go test parses the import graph of `internal/onepagers/...`
  and fails when any package imports another bounded context outside the allowed set
  (`internal/shared/**`, other contexts' `publishedlanguage` packages). The new context
  starts with its boundary machine-enforced; later slices inherit the guard unchanged
  (ports live inside `onepagers`, adapters at the composition root outside it).
- **Bounded-context canvas** — `docs/architecture/OnePagers.md` per the existing canvas
  format, documenting purpose, strategic classification, inbound/outbound communication
  (published language: the eleven configuration events; consumed: none in this slice),
  and relationship types; `docs/architecture/README.md` context map updated.

---

## Design Decisions

1. **One aggregate per (tenant, subject type)** — configuring the Application one-pager
   never contends with the Vendor one-pager. Alternative: a single per-tenant aggregate
   (exact metamodel clone) — rejected, unnecessarily large consistency boundary.
2. **Intrinsic aggregate ID; uniqueness at the handler with a DB backstop** — per the
   stream-collision lesson recorded in spec 170: aggregate IDs identify the aggregate,
   never a (tenant, subject type) pair. Alternative: derive the ID from the subject type
   — rejected, violates tactical DDD and risks stream collisions.
3. **Lazy creation on first read** — mirrors `MetaModelConfiguration` defaults; the
   default is the full catalog in catalog order. Alternative: eager creation on
   `TenantCreated` — rejected, six aggregates per tenant before anyone opens the page.
4. **Field type immutable; retire-and-redefine** (design doc D4) — protects captured
   values from silent reinterpretation. Alternative: in-place type migration — rejected.
5. **Required flag on custom fields only** (design doc D7) — built-in fields carry no
   required flag. Alternative: required built-ins — rejected by D7 as a separate concern.
6. **Single interleaved display order** (design doc D10) — an admin can place "Contract
   link" between "Description" and "Maturity". Alternative: two sectioned lists —
   rejected by the strategic DDD review.
7. **Catalog as code** — each entry is a code-owned definition; adding one is a reviewed
   change, not tenant data. Alternative: catalog rows in the database — rejected by the
   design doc (the catalog binds to published read contracts, a code concern).
8. **Reuse `PermMetaModelRead` / `PermMetaModelWrite`** — the existing tenant-configuration
   settings gate: reads on all roles, writes admin-only, exactly the required matrix.
   Alternative: a new `one-pagers:*` permission pair — rejected, identical grant matrix
   today; split later if roles ever diverge.
9. **`CustomFieldRenamed` covers display metadata (name and help text)** — the design
   doc's closed event list has no separate help-text event, and help text must stay
   editable; scoping "rename" to display metadata keeps the event list closed.
10. **Reactivated and newly included fields append to the end of the display order** —
    retired and excluded fields leave the order entirely, so no ghost positions
    accumulate. Alternative: remember the previous position — rejected, the surrounding
    order may have changed and the admin reorders explicitly anyway.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Catalog as code | Adding a built-in field requires a deployment | Single code-owned binding point; entries are small declarative definitions |
| Reuse metamodel permissions | One-pager settings writes coupled to the `metamodel:write` grant | Grant matrix is identical today; introducing a dedicated permission later is a small role-matrix change |
| Lazy creation on GET | The first read has a write side-effect | Idempotent under rules 1–2; mirrors the established metamodel precedent |
| Retire, never delete | Retired definitions accumulate in the aggregate and UI | Retired fields listed separately in settings; reactivation stays cheap |
| Six per-type aggregates per tenant | More streams and read-model rows than one per-tenant aggregate | Small consistency boundaries; rows are one per subject type, bounded at six |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off
