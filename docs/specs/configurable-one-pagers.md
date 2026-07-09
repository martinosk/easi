# Design: Configurable One-Pagers

> Status: DRAFT — awaiting Phase 1 human approval
> Author: agent + maosk
> Date: 2026-07-09
> Reviewed: strategic DDD review 2026-07-09 — shape approved; 8 required changes incorporated (D7–D10, composition mechanism, catalog ownership, deletion policy, field ordering)

## Problem Statement

Architects need to communicate the current state of the architecture to stakeholders who do not live in EASI. Today the information for a capability, application, acquired entity, vendor, or internal team is spread across detail panels, and there is no way to capture stakeholder-facing facts that are not part of the base domain model (a contact person, a link to a contract).

A **One-Pager** is a tenant-configurable fact sheet for a single subject entity. A tenant admin decides which fields appear on the one-pager for each subject type. Fields are either **built-in** (data already in the domain model: name, description, maturity, experts, origin links, …) or **custom** (tenant-defined, strongly typed: contact person, link, date, selection, …).

Three qualities drive the design:

1. **Data quality** — custom fields are typed domain values, not free-text strings. A contact person is a (name, email, company) triple with a validated email; a contract link is a validated URL with a label.
2. **Performance** — rendering a one-pager must be a constant, small number of indexed queries; completeness overviews must not fan out per entity.
3. **Config evolution** — the field configuration will change over time (fields added, retired, made required). Existing data must never be silently invalidated or lost.

## Research Summary

- Backend is DDD + CQRS + event sourcing: one Go package per bounded context, events in `infrastructure.events`, per-context read-model tables, synchronous in-memory event bus, RLS tenancy (`tenant_id` + `app.current_tenant`).
- **Tenant-configuration precedent**: the `metamodel` context's `MetaModelConfiguration` — one aggregate per tenant, lazily created with defaults, edited from the admin Settings page (`/settings/*`), consumed by the frontend via TanStack Query with `staleTime: Infinity` (specs 095, 098, 152).
- **Subject entities all exist**: `Capability` (capabilitymapping), `EnterpriseCapability` (enterprisearchitecture), `ApplicationComponent`, `AcquiredEntity`, `Vendor`, `InternalTeam` (architecturemodeling, specs 117/121).
- **No custom-field mechanism exists.** Closest analogs: typed per-aggregate metadata (spec 024), `Expert` entities on components/capabilities (specs 114/115), tags on capabilities.
- **Performance precedent**: denormalized cache tables maintained by dedicated cache projectors (`capability_component_cache`, `ea_*_cache`, …); cursor pagination; denormalized name columns.
- **Value objects**: constructor-validated immutable VOs are the norm. Email VOs exist (auth, metamodel); no URL VO exists yet — one must be introduced.
- Frontend: React Hook Form + Zod (`frontend/src/lib/schemas/`), detail *panels* (not routed pages) per entity, settings sections under `features/settings/`, deep-link share URLs exist (spec 113). No print/PDF export exists.

## Ubiquitous Language

| Term | Meaning |
|---|---|
| **One-Pager** | The rendered, stakeholder-facing fact sheet for one subject entity |
| **Subject** | The entity a one-pager describes; one of the supported subject types |
| **Subject Type** | Capability, Enterprise Capability, Application, Acquired Entity, Vendor, Internal Team |
| **One-Pager Configuration** | Per-tenant, per-subject-type definition of which fields the one-pager shows |
| **Built-in Field** | A field sourced from the existing domain model, selected for display |
| **Custom Field Definition** | A tenant-defined field with a name, field type, and required flag |
| **Field Type** | The value shape of a custom field: Text, Number, Date, Link, Selection, Contact Person |
| **Field Value** | A typed value for one custom field on one subject |
| **Completeness** | How fully a subject's one-pager satisfies the current configuration's required fields |

## Proposed Approach

### New bounded context: `onepagers`

A supporting context. It owns configuration and custom field values; it does **not** own subject entities — it references them by typed ID (`SubjectRef` = subject type + subject ID, scoped to this context) and sources built-in field data from the owning contexts at query time. Context relationships:

| Relationship | Pattern |
|---|---|
| onepagers → capabilitymapping / enterprisearchitecture / architecturemodeling | **Customer/Supplier** — consumes published-language events (e.g. entity deleted) and built-in field data via ports |
| onepagers → metamodel | **Customer/Supplier** — rendering built-in maturity/pillar fields requires the tenant's maturity-scale and strategy-pillar semantics |

**Composition mechanism (mandatory):** built-in field data is read through **onepagers-defined ports** (`BuiltInFieldSource` interfaces in `onepagers/application/ports`) implemented by **adapters at the composition root** wrapping the supplier contexts' read models — the `direction_composition_adapters.go` precedent. `onepagers` never imports another context's application packages and never issues SQL against another context's tables. Query-time composition through ports keeps names always fresh with no denormalization staleness.

**Built-in field catalog:** owned by `onepagers`, as **code, not data** — each catalog entry binds to a field of a supplier's published read contract, and the composition adapter is the single binding point. Coupling cost is mitigated by a mandatory **integration test per subject type** asserting every catalog entry resolves against the supplier read model, so supplier drift breaks the build instead of a tenant's one-pager.

### Aggregate 1: `OnePagerConfiguration`

One aggregate per **(tenant, subject type)** — smaller consistency boundary than metamodel's single per-tenant aggregate, so configuring the Application one-pager never contends with the Vendor one-pager. Uniqueness enforced at the command handler via read-model lookup, DB unique `(tenant_id, subject_type)` as backstop. Created lazily with a default configuration (mirrors `MetaModelConfiguration`).

Contents:
- **Built-in field selections** (from the per-subject-type catalog owned by this context).
- **Custom field definitions** — entities within the aggregate: stable `FieldID` (UUID, intrinsic), display name, field type, required flag, help text, options (Selection only), and an **active/retired** status.
- A **single interleaved display order** over the mixed set of built-in and custom fields — the aggregate models one ordering, not two sectioned lists, so an admin can place "Contract link" between "Description" and "Maturity".

Rules:
- A field's **type is immutable**. Changing type = retire the old field, add a new one.
- Fields are **retired, never deleted** — values remain in history and reappear if the field is reactivated.
- Renaming is free (identity is `FieldID`, not the name).
- Retiring a Selection option keeps existing values valid; they render flagged as using a retired option.
- The **required flag exists only on custom field definitions**; built-in fields carry no required flag and do not participate in completeness (D7).

Events (past-tense business facts): `OnePagerConfigurationCreated`, `CustomFieldDefined`, `CustomFieldRenamed`, `CustomFieldRequirementChanged`, `CustomFieldRetired`, `CustomFieldReactivated`, `BuiltInFieldIncluded`, `BuiltInFieldExcluded`, `OnePagerFieldsReordered`, `SelectionOptionAdded`, `SelectionOptionRetired`.

### Aggregate 2: `OnePagerFacts`

One aggregate per subject, holding that subject's custom field values. Own intrinsic UUID; references the subject via `SubjectRef` VO (never derives its ID from the subject's). One-per-subject uniqueness at the handler + DB unique `(tenant_id, subject_type, subject_id)`; the creation handler verifies **subject existence** through the onepagers subject port (introduced in minimal existence-check form in slice 2, extended to full built-in field reading in slice 3) before creating the aggregate.

Events: `FieldValueRecorded(fieldID, value envelope)`, `FieldValueCleared(fieldID)`, `OnePagerFactsArchived(reason: subject deleted)`.

**Value envelope:** every recorded value is persisted — in the event payload and in the read-model column — as a discriminated envelope `{type, version, value}`, so introducing new field types or evolving a value shape is an upcaster exercise, never JSONB archaeology.

**Typed values (data quality).** `FieldValue` is a closed set of constructor-validated VOs:

| Field Type | Value Object | Validation |
|---|---|---|
| Text | `TextValue` | trimmed, non-empty, length cap |
| Number | `NumberValue` | finite decimal |
| Date | `DateValue` | ISO date |
| Link | `LinkValue{Label, URL}` | new shared `URL` VO — absolute http(s) URL |
| Selection | `SelectionValue{OptionID}` | option exists on the field definition |
| Contact Person | `ContactPerson{Name, Email, Company}` | non-empty name; validated email; company optional |

Value-vs-definition validation (type match, active field, valid option) is a cross-aggregate concern: the command handler loads the current configuration read model, validates, and passes the constructed VO to the aggregate. The aggregate still guards its own invariants (no value without a field ID, one value per field).

**Deletion of subjects**: a policy in `onepagers` reacts to published-language deletion events of subject contexts by appending `OnePagerFactsArchived` to the facts aggregate's own stream — closing it in the domain, not just the read model — and the projector removes the read-model rows. An archived facts aggregate rejects further `FieldValueRecorded` commands, so facts for a deleted subject can never be resurrected. The retire-field-vs-record-value race between the two aggregates is accepted eventual consistency: retired fields keep their values by design, so no invariant is violated.

### Configuration change handling (the optional → required case)

**Principle: configuration changes never mutate or invalidate recorded facts. Completeness is evaluated at read time against the *current* configuration.**

- Making a field required does not block anything retroactively. Subjects missing the value become **incomplete**; the one-pager renders the field as "missing — required", and list views can surface a completeness indicator.
- At write time, "required" is enforced softly: the edit form highlights missing required fields, but saving partial facts is allowed (an EA tool must accept incremental data entry — a hard gate would freeze edits on hundreds of pre-existing entities the moment an admin flips a flag).
- The settings UI shows an **impact preview** before the admin confirms: "Making *Contract link* required will mark 37 Applications incomplete."
- Retired fields disappear from the one-pager and the edit form but their values are preserved (event history + read model row flagged retired).

### Performance

- **Rendering one one-pager**: `GET /api/v1/one-pagers/{subjectType}/{subjectID}` composes: configuration (1 indexed row, effectively cached — frontend uses `staleTime: Infinity` and invalidates on settings mutations), custom facts (1 query on `(tenant_id, subject_type, subject_id)`), the subject's existing detail read model (1 query, already indexed), and — when maturity/pillar built-in fields are configured and their semantics are not already carried by the subject read — at most 1 metamodel semantics read. Constant query count, no N+1.
- **Facts read model**: `onepagers.one_pager_facts` — one row per (subject, field) with a typed-value JSONB column plus extracted columns for the common render path; PK `(tenant_id, subject_type, subject_id, field_id)`, RLS policy as per every other table.
- **Completeness at scale** (list views, impact preview): computed with one set-based SQL query per subject type (facts table ⨝ required-field list), not per-entity fan-out. If profiling shows this is hot, the existing cache-projector pattern (`ea_*_cache`) is the escape hatch — a `one_pager_completeness` cache maintained by a projector reacting to fact and configuration events. Start without it; the precedent makes it a drop-in later.

### Frontend

- **Settings**: new `/settings/one-pagers` section (per settings-page precedent), one tab per subject type: pick built-in fields from the catalog, define/reorder/retire custom fields, with the impact preview on requirement changes.
- **Capture**: detail panels gain a configuration-driven "One-Pager" section; the Zod schema for the section is generated from the field definitions (RHF + `zodResolver`, matching existing form conventions).
- **View**: a routed, deep-linkable one-pager page (spec 113 deep-link precedent) rendering built-in + custom fields in the configured order — clean enough to present or screen-share. (Print/PDF export is deliberately a separate later spec.)

## Alternatives Considered

1. **Store custom values on each subject aggregate** (extend Capability, ApplicationComponent, … with a values map). Rejected: pollutes six aggregates across three contexts with a concern none of them owns, couples every context to the field-type system, and makes the configuration ↔ value validation a shared-kernel problem.
2. **Single per-tenant configuration aggregate (exact metamodel clone)**. Rejected: one aggregate serializing edits across six subject types is an unnecessarily large consistency boundary; per-(tenant, subject type) keeps transactions small. The lazy-default creation pattern is kept.
3. **Free-form JSON custom fields ("just strings")**. Rejected by the feature's own premise: data quality requires constructor-validated typed values.
4. **Hard-enforce required fields at write time.** Rejected: flipping optional → required would instantly block edits on every incomplete pre-existing entity. Read-time completeness + impact preview fosters quality without freezing the tenant.
5. **Generic EAV / dynamic schema tables (one column per custom field, ALTER TABLE per config change).** Rejected: schema churn per tenant config edit is operationally hostile; a typed-JSONB fact row per (subject, field) with RLS matches the existing read-model style and stays index-friendly.

## Key Decisions

| # | Decision | Resolution |
|---|---|---|
| D1 | Which "capability" gets one-pagers — `Capability`, `EnterpriseCapability`, or both? | **Both** as distinct subject types (confirmed by maosk, 2026-07-09) |
| D2 | Optional → required handling | **Read-time completeness + impact preview**; no retroactive blocking (confirmed by maosk, 2026-07-09) |
| D3 | Contact Person field type vs existing `Expert` concept | **Keep both, distinct**: Experts are a *built-in* catalog field; Contact Person is a *custom* field type for stakeholder-facing contacts incl. external companies (confirmed by maosk, 2026-07-09) |
| D4 | Field type changes | Immutable type; retire + redefine |
| D5 | New shared `URL` value object | Add to shared kernel (`internal/shared/eventsourcing/valueobjects`) since Link fields need it and nothing exists today |
| D6 | Completeness cache table now or later | Later, only if set-based queries prove hot — precedent makes it drop-in |
| D7 | Can built-in fields be "required" / count toward completeness? | **No — custom fields only.** Built-in completeness changes the completeness query shape and the configuration events; if wanted, it is a separate spec (confirmed by maosk, 2026-07-09) |
| D8 | How `onepagers` reads supplier data | Consumer-defined ports + composition-root adapters (per `direction_composition_adapters.go`); direct imports or cross-schema SQL prohibited (strategic DDD review, 2026-07-09) |
| D9 | Field-value persistence shape | Discriminated `{type, version, value}` envelope in events and read model from day one (strategic DDD review, 2026-07-09) |
| D10 | Display ordering of fields | One interleaved order over built-in + custom fields, modeled as a single ordering in the configuration aggregate (strategic DDD review, 2026-07-09) |

## Risks & Guards (from strategic DDD review)

| Risk | Guard |
|---|---|
| `OnePagerFacts` drifts into a generic custom-fields/EAV platform | Keep every term one-pager-scoped; if generic custom fields become a real requirement, extract a new subdomain then — do not stretch this one |
| Supplier read-model drift silently breaks catalogs | Mandatory per-subject-type integration tests binding catalog entries to supplier read contracts |
| New subject types miss wiring steps | An add-a-subject-type checklist maintained across the slice specs: subject-type enumeration, catalog entries + settings tab (175), deletion policy subscription + subject port (176), field-source port + adapter + catalog-contract test (177) |
| Impact preview implemented as raw cross-schema SQL under deadline | Subject population counts are a port method on `BuiltInFieldSource`, like all supplier reads |

## Vertical Slice Plan (each becomes its own numbered spec)

| Slice | Spec | Scope |
|---|---|---|
| 1 | 175 — OnePagerConfiguration | `onepagers` context skeleton, configuration aggregate, built-in field catalog, settings UI |
| 2 | 176 — OnePagerFacts capture | Facts aggregate, typed field values, config-driven edit section on detail panels |
| 3 | 177 — OnePagerView | Composed read endpoint + routed, deep-linkable one-pager page |
| 4 | 178 — Completeness & impact preview | Read-time completeness, incomplete indicators, requirement-change impact preview |

Slice 1 is independently deployable (an admin can configure fields before any are filled); each later slice is demonstrable on its own.
