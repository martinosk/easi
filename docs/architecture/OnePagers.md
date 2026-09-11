# Bounded Context Canvas: OnePagers

## Name
**OnePagers**

## Purpose
Deliver the One-Pager — a stakeholder-facing fact sheet for a single subject entity. Tenant administrators shape it per subject type by choosing built-in fields and the custom attributes MetaModel defines, ordering them and marking them required; architects record Field Values on subjects; anyone with the subject's read permission opens the composed, presentable sheet. The context owns the One-Pager Configuration and One-Pager Facts aggregates, the code-owned built-in field catalog, and the composed read; it does not own subject entities or the custom-field schema and reads both only through its own event-fed caches.

**Key Stakeholders:**
- Tenant Administrators (configure the field set per subject type)
- Enterprise Architects (record facts, share one-pagers)
- Stakeholders outside EASI (read a shared one-pager)

## Strategic Classification
**Supporting Domain** — makes the content of the core modeling and analysis contexts presentable to stakeholders outside EASI. The competitive differentiation lies in those core domains; this context increases their reach, which justifies reusing established patterns (CQRS/ES, metamodel settings precedent) over novel investment.

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **One-Pager** | The rendered, stakeholder-facing fact sheet for one subject entity |
| **Subject Type** | Capability, Application, Acquired Entity, Vendor, Internal Team |
| **One-Pager Configuration** | Per-(tenant, subject type) definition of which fields the one-pager shows |
| **One-Pager Facts** | Per-subject aggregate holding that subject's custom Field Values |
| **Built-in Field** | A catalog-defined field sourced from the owning context's read contract at query time |
| **Relation Built-in Field** | A read-only built-in whose value is a list of references to related entities (each rendered by name, deep-linked to its own one-pager when it is a subject type); excluded by default, opt-in via the spec-175 include/exclude/reorder machinery |
| **Built-in Field Catalog** | Code-owned, per-subject-type list of built-in field entries (stable ID, label) |
| **Built-in Field Source** | Consumer-defined port through which the context reads a subject's built-in field data |
| **Custom Field Definition** | MetaModel's subject attribute (FieldID, name, type, help text, options, bounds, active) as cached by this context; presentation policy (inclusion, order, required flag) lives on the configuration |
| **Field Type** | Value shape of a custom field: Text, Number, Date, Link, Selection, Contact Person |
| **Field Value** | A typed, constructor-validated value for one custom field on one subject |
| **Value Envelope** | Persistence shape of every Field Value: discriminated `{type, version, value}` |
| **Display Order** | Single interleaved ordering over included built-in and included custom fields |

## Inbound Communication

**Commands** (from Frontend/API, REST under `/api/v1/one-pagers`):
- Configuration writes, gated by `PermMetaModelWrite` under `/one-pagers/configurations/{subjectType}`: `ChangeCustomFieldRequirement`, `IncludeBuiltInField`, `ExcludeBuiltInField`, `ChangeBuiltInFieldRequirement`, `ReorderOnePagerFields`. `IncludeCustomField` / `ExcludeCustomField` are dispatched only by the in-context reactor on MetaModel's attribute events; schema changes go to MetaModel (`x-attribute-schema` on the configuration response)
- Facts writes, gated by the subject's write permission under `/one-pagers/{subjectType}/{subjectID}/facts`: `RecordFieldValue`, `ClearFieldValue`

**Queries served**:
- Configuration read (`PermMetaModelRead`; lazily creates the catalog default on first read)
- Facts read (subject's read permission)
- Composed one-pager read `GET /api/v1/one-pagers/{subjectType}/{subjectID}` (subject's read permission): subject header plus built-in and custom fields interleaved in the configured display order

**Events consumed** (supplier published language, projected into OnePagers' own caches — spec 209):
- Subject lifecycle and attributes → `one_pager_subject_index` (name, existence, completeness counters, and `built_in_fields`: the subject's **complete** published attribute set, keyed by the supplier's attribute names): `Capability*` incl. `CapabilityMetadataUpdated`, `CapabilityLevelChanged` and expert events (capabilitymapping); `ApplicationComponent*` incl. expert events, `AcquiredEntity*`, `Vendor*`, `InternalTeam*` (architecturemodeling). Deletion events also archive the subject's facts.
- Relations → `subject_relation_cache` (+ `business_domain_name_cache` for domain labels): `SystemLinkedToCapability`, `SystemRealizationDeleted`, `CapabilityRealizationsInherited/Uninherited`, `CapabilityDependencyCreated/Deleted`, `CapabilityAssignedToDomain/UnassignedFromDomain`, `CapabilityParentChanged`, `BusinessDomain*` (capabilitymapping); `ComponentRelation*`, `OriginLink*` (architecturemodeling). Expert names travel on the expert events, so no user cache exists.
- Rendering semantics → `maturity_scale_cache`: `MaturityScaleConfigUpdated/Reset`, `MetaModelConfigurationCreated` (metamodel)
- Custom-field schema → `custom_field_definition_cache`: `SubjectAttributeDefined/Renamed/Retired/Reactivated/OptionAdded/OptionRetired/BoundsChanged` (metamodel). The same Defined / Retired / Reactivated events drive a reactor that includes or excludes the field on the subject type's configuration, creating the configuration when it does not exist yet.

Every cache is backfilled by migration 148 (subjects) and 161 (definitions) from the suppliers' tables; definitions still pending transfer are handed to MetaModel at startup through its published `ImportSubjectAttribute` command, which preserves every FieldID. Adding a built-in field over an attribute a supplier already publishes is a catalog change only; a genuinely new supplier attribute ships with a supplier event change and an OnePagers backfill.

## Outbound Communication

**Events published**: none — the context has no published language; its event types are internal aggregate mechanics. Machine-enforced by the boundary test, which asserts the `publishedlanguage` package does not exist.

**Commands issued**: MetaModel's `ImportSubjectAttribute`, once per legacy definition, by the startup transfer.

**Queries made**: none. The ports in `/backend/internal/onepagers/application/ports` (`BuiltInFieldSource` per subject type, `MaturityScaleSource`, `SubjectExistenceChecker`) are implemented inside the context over its own caches; the catalog-entry → published-attribute binding lives only in those adapters.

## Business Rules

1. One configuration per (tenant, subject type): handler-level uniqueness check, DB unique constraint as backstop; the aggregate ID is an intrinsic UUID
2. First configuration read lazily creates the default: every catalog built-in field, catalog order, no custom fields; creation is idempotent
3. A configuration may only include built-in fields from the per-subject-type catalog
4. A custom field enters the display order when MetaModel defines or reactivates the attribute and leaves it when MetaModel retires it; only included fields can be marked required
5. One interleaved display order over included built-in and included custom fields
6. The required flag is presentation policy, recorded uniformly for built-in and custom fields on the configuration; it never validates or blocks recorded data, and a dormant flag survives exclude-and-reinclude
7. Facts writes validate against the cached definition: unknown or retired fields, undefined or retired options and out-of-bounds numbers are rejected
8. Legacy schema events in configuration streams (`CustomFieldDefined` and friends) still replay their inclusion and requirement effects; no new ones are appended
9. One facts aggregate per subject, created on first recorded value after a subject-existence check through the subject port
10. Every Field Value is a typed, constructor-validated VO persisted as a Value Envelope; validation against the current configuration happens in the command handler
11. Subject deletion archives the facts aggregate in its own stream and removes its read-model rows; archived facts reject further writes
12. The composed read assembles at query time from OnePagers' own tables with a constant query count: one configuration read, one facts read, one subject-index read, at most one maturity-scale read, plus — per *included* relation built-in — one bounded relation-cache read joined to the subject index for names, so the total stays independent of the number of related entities and of the number of configured fields
13. Fields whose definition is retired or unknown never render or count toward completeness; values referencing retired selection options render flagged, never invalid
14. The composed read is authorized with the subject's own read permission; a missing configuration falls back to the catalog default without persisting

## Design Constraints

1. Catalog as code: each entry binds to exactly one attribute of the subject's cached published attribute set, and the in-context adapter is the only place that binding exists; per-subject-type adapter tests fail the build on drift
2. One aggregate per (tenant, subject type) keeps the configuration consistency boundary small
3. Reuses `PermMetaModelRead`/`PermMetaModelWrite` for configuration because the required grant matrix is identical to the metamodel settings gate; the composed read inherits the subject's read permission so it can never reveal more than the subject's own detail endpoint
4. Configuration reads are frequent, writes rare: the frontend caches with `staleTime: Infinity` and invalidates on mutation
5. Event-fed, backfilled caches of supplier data (spec 209): the composed read never touches another context; freshness follows the synchronous in-process event bus

## Open Questions

1. Should one-pager settings get a dedicated permission pair if role matrices ever diverge from the metamodel gate?
2. Does completeness at scale (spec 178) need a cache projector, or do set-based queries suffice?

## Boundary Health

- **Zero cross-context imports**: machine-enforced by `/backend/internal/onepagers/architecture_boundary_test.go` — only `internal/shared`, other contexts' `publishedlanguage` packages, and shared eventstore/database infrastructure are importable
- **No published language**: the context publishes nothing; consumers of its data go through its REST API only
- **Catalog binding integrity**: every catalog entry resolves against the cached published attribute set, enforced by per-subject-type adapter tests inside the context; the composition root wires nothing for OnePagers (`TestCompositionRootOnlyRegistersRoutes`)

## Architecture Notes

### Implementation Location
`/backend/internal/onepagers/`

### Key Packages
- `domain/aggregates/` - OnePagerConfiguration and OnePagerFacts aggregates
- `domain/catalog/` - code-owned built-in field catalog per subject type
- `domain/valueobjects/` - SubjectType, SubjectRef, FieldID, FieldRef, FieldValue, ValueEnvelope
- `domain/events/` - configuration and facts events
- `application/ports/` - BuiltInFieldSource, MaturityScaleSource, SubjectExistenceChecker
- `application/queries/` - composed one-pager read assembly
- `application/` - commands, handlers, projectors, read models
- `infrastructure/` - API routes/handlers/DTOs, event-sourced repositories

### Technical Patterns
- **CQRS with Event Sourcing** in the shared event store; own PostgreSQL schema `onepagers` with RLS
- **Read models**: `onepagers.one_pager_configurations` (one row per tenant + subject type: display order and required flags), `onepagers.custom_field_definition_cache` (one row per field, projected from MetaModel) and `onepagers.one_pager_facts` (one row per subject + field, typed-value JSONB)
- **Lazy default creation** on first configuration read, mirroring the MetaModel configuration precedent
- **Optimistic concurrency** via aggregate version, conflicts surfaced as 409

### API Style
- REST Level 3 with HATEOAS; write affordances advertised only to permitted actors; subject detail responses link to the one-pager via `x-one-pager`, and the one-pager links back via `x-subject`
- Configuration commands under `/api/v1/one-pagers/configurations/{subjectType}`; facts and the composed read under `/api/v1/one-pagers/{subjectType}/{subjectID}`
