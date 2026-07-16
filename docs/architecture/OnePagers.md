# Bounded Context Canvas: OnePagers

## Name
**OnePagers**

## Purpose
Deliver the One-Pager — a stakeholder-facing fact sheet for a single subject entity. Tenant administrators shape it per subject type by choosing built-in fields and defining typed custom fields; architects record Field Values on subjects; anyone with the subject's read permission opens the composed, presentable sheet. The context owns the One-Pager Configuration and One-Pager Facts aggregates, the code-owned built-in field catalog, and the composed read; it does not own subject entities and reads their data only through its own ports.

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
| **Subject Type** | Capability, Enterprise Capability, Application, Acquired Entity, Vendor, Internal Team |
| **One-Pager Configuration** | Per-(tenant, subject type) definition of which fields the one-pager shows |
| **One-Pager Facts** | Per-subject aggregate holding that subject's custom Field Values |
| **Built-in Field** | A catalog-defined field sourced from the owning context's read contract at query time |
| **Relation Built-in Field** | A read-only built-in whose value is a list of references to related entities (each rendered by name, deep-linked to its own one-pager when it is a subject type); excluded by default, opt-in via the spec-175 include/exclude/reorder machinery |
| **Built-in Field Catalog** | Code-owned, per-subject-type list of built-in field entries (stable ID, label) |
| **Built-in Field Source** | Consumer-defined port through which the context reads a subject's built-in field data |
| **Custom Field Definition** | Tenant-defined field: FieldID, display name, field type, required flag, help text, options |
| **Field Type** | Value shape of a custom field: Text, Number, Date, Link, Selection, Contact Person |
| **Field Value** | A typed, constructor-validated value for one custom field on one subject |
| **Value Envelope** | Persistence shape of every Field Value: discriminated `{type, version, value}` |
| **Display Order** | Single interleaved ordering over included built-in and active custom fields |

## Inbound Communication

**Commands** (from Frontend/API, REST under `/api/v1/one-pagers`):
- Configuration writes, gated by `PermMetaModelWrite` under `/one-pagers/configurations/{subjectType}`: `DefineCustomField`, `RenameCustomField`, `ChangeCustomFieldRequirement`, `RetireCustomField`, `ReactivateCustomField`, `IncludeBuiltInField`, `ExcludeBuiltInField`, `ReorderOnePagerFields`, `AddSelectionOption`, `RetireSelectionOption`
- Facts writes, gated by the subject's write permission under `/one-pagers/{subjectType}/{subjectID}/facts`: `RecordFieldValue`, `ClearFieldValue`

**Queries served**:
- Configuration read (`PermMetaModelRead`; lazily creates the catalog default on first read)
- Facts read (subject's read permission)
- Composed one-pager read `GET /api/v1/one-pagers/{subjectType}/{subjectID}` (subject's read permission): subject header plus built-in and custom fields interleaved in the configured display order

**Events consumed** (supplier published language; each archives the subject's facts):
`CapabilityDeleted` (capabilitymapping), `EnterpriseCapabilityDeleted` (enterprisearchitecture), `ApplicationComponentDeleted`, `AcquiredEntityDeleted`, `VendorDeleted`, `InternalTeamDeleted` (architecturemodeling)

## Outbound Communication

**Events published**: none — the context has no published language; its event types are internal aggregate mechanics. Machine-enforced by the boundary test, which asserts the `publishedlanguage` package does not exist.

**Queries made** (Customer/Supplier, query-time, through consumer-defined ports in `/backend/internal/onepagers/application/ports` with adapters at the composition root `/backend/internal/infrastructure/api`):
- `BuiltInFieldSource` (one per subject type) wrapping the supplier read contracts: `CapabilityReadModel`, `EnterpriseCapabilityReadModel`, `ApplicationComponentReadModel`, `AcquiredEntityReadModel`, `VendorReadModel`, `InternalTeamReadModel`
- Relation built-in fields (spec 188) resolve, in the same `BuiltInFieldSource` adapters and only for *included* relation entries, through the supplier relation read contracts: `RealizationReadModel` (`GetByCapabilityID`/`GetByComponentID`), `DependencyReadModel` (`GetOutgoing`), `DomainCapabilityAssignmentReadModel` (`GetByCapabilityID`), `CapabilityReadModel` (`GetChildren` + batched `GetByIDs` name resolution), `ApplicationComponentReadModel` (batched `GetByIDs` name resolution), `BuiltByRelationshipReadModel`/`PurchasedFromRelationshipReadModel`/`AcquiredViaRelationshipReadModel` (forward `GetByComponentID` and reverse `GetByTeamID`/`GetByVendorID`/`GetByEntityID`), `ComponentRelationReadModel` (`GetBySourceID`), and the enterprise `CompositionService` (`CompositionForEC`) for `Included Capabilities`
- `MaturityScaleSource` wrapping the metamodel configuration read model — **metamodel is the upstream supplier of rendering semantics** (tenant maturity-scale sections)
- `SubjectExistenceChecker` wrapping the six supplier read models (facts creation guard)

## Business Rules

1. One configuration per (tenant, subject type): handler-level uniqueness check, DB unique constraint as backstop; the aggregate ID is an intrinsic UUID
2. First configuration read lazily creates the default: every catalog built-in field, catalog order, no custom fields; creation is idempotent
3. A configuration may only include built-in fields from the per-subject-type catalog
4. Field identity is the FieldID; renaming never changes identity; field type is immutable — the path is retire-and-redefine
5. Fields and selection options are retired, never deleted; reactivation restores identity, type, required flag, and options
6. One interleaved display order over included built-in and active custom fields
7. The required flag exists only on custom field definitions and never validates or blocks recorded data
8. Active display names are unique per configuration, case-insensitive, across custom fields and included built-in labels
9. One facts aggregate per subject, created on first recorded value after a subject-existence check through the subject port
10. Every Field Value is a typed, constructor-validated VO persisted as a Value Envelope; validation against the current configuration happens in the command handler
11. Subject deletion archives the facts aggregate in its own stream and removes its read-model rows; archived facts reject further writes
12. The composed read assembles at query time with a constant query count: one configuration read, one facts read, one subject read through the port, at most one metamodel semantics read, plus — per *included* relation built-in — at most one bounded edge read and one batched counterpart-name lookup (`id IN (…)`, never per-edge), so the total stays independent of the number of related entities and of the number of configured fields
13. Retired fields never render on the one-pager; values referencing retired selection options render flagged, never invalid
14. The composed read is authorized with the subject's own read permission; a missing configuration falls back to the catalog default without persisting

## Design Constraints

1. Catalog as code: each entry binds to exactly one field of a supplier's published read contract, and the composition adapter is the only place that binding exists; per-subject-type catalog-contract integration tests fail the build on supplier drift
2. One aggregate per (tenant, subject type) keeps the configuration consistency boundary small
3. Reuses `PermMetaModelRead`/`PermMetaModelWrite` for configuration because the required grant matrix is identical to the metamodel settings gate; the composed read inherits the subject's read permission so it can never reveal more than the subject's own detail endpoint
4. Configuration reads are frequent, writes rare: the frontend caches with `staleTime: Infinity` and invalidates on mutation
5. Query-time composition with no denormalized one-pager cache: supplier names and values are always fresh; the cache-projector precedent is the escape hatch if profiling demands it

## Open Questions

1. Should one-pager settings get a dedicated permission pair if role matrices ever diverge from the metamodel gate?
2. Does completeness at scale (spec 178) need a cache projector, or do set-based queries suffice?

## Boundary Health

- **Zero cross-context imports**: machine-enforced by `/backend/internal/onepagers/architecture_boundary_test.go` — only `internal/shared`, other contexts' `publishedlanguage` packages, and shared eventstore/database infrastructure are importable
- **No published language**: the context publishes nothing; consumers of its data go through its REST API only
- **Catalog binding integrity**: every catalog entry resolves against its supplier read contract, enforced by per-subject-type catalog-contract integration tests at `/backend/internal/infrastructure/api`

## Architecture Notes

### Implementation Location
`/backend/internal/onepagers/`

### Key Packages
- `domain/aggregates/` - OnePagerConfiguration and OnePagerFacts aggregates
- `domain/catalog/` - code-owned built-in field catalog per subject type
- `domain/valueobjects/` - SubjectType, SubjectRef, FieldID, FieldType, CustomField, FieldValue, ValueEnvelope
- `domain/events/` - configuration and facts events
- `application/ports/` - BuiltInFieldSource, MaturityScaleSource, SubjectExistenceChecker
- `application/queries/` - composed one-pager read assembly
- `application/` - commands, handlers, projectors, read models
- `infrastructure/` - API routes/handlers/DTOs, event-sourced repositories

### Technical Patterns
- **CQRS with Event Sourcing** in the shared event store; own PostgreSQL schema `onepagers` with RLS
- **Read models**: `onepagers.one_pager_configurations` (one row per tenant + subject type) and `onepagers.one_pager_facts` (one row per subject + field, typed-value JSONB)
- **Lazy default creation** on first configuration read, mirroring the MetaModel configuration precedent
- **Optimistic concurrency** via aggregate version, conflicts surfaced as 409

### API Style
- REST Level 3 with HATEOAS; write affordances advertised only to permitted actors; subject detail responses link to the one-pager via `x-one-pager`, and the one-pager links back via `x-subject`
- Configuration commands under `/api/v1/one-pagers/configurations/{subjectType}`; facts and the composed read under `/api/v1/one-pagers/{subjectType}/{subjectID}`
