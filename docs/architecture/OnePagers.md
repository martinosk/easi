# Bounded Context Canvas: OnePagers

## Name
**OnePagers**

## Purpose
Let tenant administrators shape, per subject type, the One-Pager — a stakeholder-facing fact sheet for a single subject entity — by choosing which built-in fields it shows and defining typed custom fields that extend it. The context owns the One-Pager Configuration aggregate and the code-owned built-in field catalog; it does not own subject entities and names them only through a context-scoped `SubjectType` enumeration.

**Key Stakeholders:**
- Tenant Administrators (configure the field set per subject type)
- Enterprise Architects (rely on a stable, well-typed field set for captured facts)

## Strategic Classification
**Supporting Domain** — makes the content of the core modeling and analysis contexts presentable to stakeholders outside EASI. The competitive differentiation lies in those core domains; this context increases their reach, which justifies reusing established patterns (CQRS/ES, metamodel settings precedent) over novel investment.

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **One-Pager** | The rendered, stakeholder-facing fact sheet for one subject entity |
| **Subject Type** | Capability, Enterprise Capability, Application, Acquired Entity, Vendor, Internal Team |
| **One-Pager Configuration** | Per-(tenant, subject type) definition of which fields the one-pager shows |
| **Built-in Field** | A catalog-defined field sourced from the existing domain model, selectable for display |
| **Built-in Field Catalog** | Code-owned, per-subject-type list of built-in field entries (stable ID, label) |
| **Custom Field Definition** | Tenant-defined field: FieldID, display name, field type, required flag, help text, options |
| **Field Type** | Value shape of a custom field: Text, Number, Date, Link, Selection, Contact Person |
| **FieldID** | Stable UUID identity of a custom field; survives renames, retirement, reactivation |
| **Selection Option** | Option on a Selection field with stable OptionID; addable and retirable, never deleted |
| **Display Order** | Single interleaved ordering over included built-in and active custom fields |

## Inbound Communication

**Commands** (from Frontend/API, REST under `/api/v1/one-pagers/configurations/{subjectType}`):
- Reads gated by `PermMetaModelRead`: get configuration (lazily creates the catalog default on first read)
- Writes gated by `PermMetaModelWrite`: `DefineCustomField`, `RenameCustomField`, `ChangeCustomFieldRequirement`, `RetireCustomField`, `ReactivateCustomField`, `IncludeBuiltInField`, `ExcludeBuiltInField`, `ReorderOnePagerFields`, `AddSelectionOption`, `RetireSelectionOption`

**Events** (from other contexts): none — the context consumes no external events (subject-deletion subscriptions arrive with the facts aggregate, see Planned Slices).

## Outbound Communication

**Events** (published language, `/backend/internal/onepagers/publishedlanguage/events.go`):
- `OnePagerConfigurationCreated`, `CustomFieldDefined`, `CustomFieldRenamed`, `CustomFieldRequirementChanged`, `CustomFieldRetired`, `CustomFieldReactivated`, `BuiltInFieldIncluded`, `BuiltInFieldExcluded`, `OnePagerFieldsReordered`, `SelectionOptionAdded`, `SelectionOptionRetired`

Today the only subscriber is the context's own read-model projector; no other context consumes these events, and no queries or commands leave the context.

## Business Rules

1. One configuration per (tenant, subject type): handler-level uniqueness check, DB unique constraint as backstop; the aggregate ID is an intrinsic UUID
2. First read lazily creates the default: every catalog built-in field, catalog order, no custom fields; creation is idempotent
3. A configuration may only include built-in fields from the per-subject-type catalog
4. Field identity is the FieldID; renaming (display name, help text) never changes identity
5. Field type is immutable — the path is retire-and-redefine
6. Fields are retired, never deleted; reactivation restores FieldID, type, required flag, and options
7. One interleaved display order over included built-in and active custom fields; retired and excluded fields leave it, reactivated and re-included fields append at the end
8. The required flag exists only on custom field definitions and never validates or blocks recorded data
9. Selection options have stable OptionIDs and are retire-only; a Selection field is defined with at least one option
10. Active display names are unique per configuration, case-insensitive, across custom fields and included built-in labels
11. Every change is one of the eleven past-tense events; replay reconstructs the configuration

## Design Constraints

1. Catalog as code: each entry binds to a published read contract of a supplier context, so adding one is a reviewed deployment, not tenant data
2. One aggregate per (tenant, subject type) keeps the consistency boundary small — at most six aggregates and read-model rows per tenant
3. Reuses `PermMetaModelRead`/`PermMetaModelWrite` because the required grant matrix is identical to the metamodel settings gate
4. Configuration reads are frequent, writes rare: the frontend caches with `staleTime: Infinity` and invalidates on mutation

## Planned Slices

Designed in [/docs/specs/configurable-one-pagers.md](/docs/specs/configurable-one-pagers.md); not yet built:

| Slice | Spec | Adds to this context |
|-------|------|----------------------|
| Facts capture | 176 | `OnePagerFacts` aggregate (typed field values), subscriptions to subject contexts' published-language deletion events, subject-existence port |
| Composed view | 177 | `BuiltInFieldSource` ports implemented by composition-root adapters over supplier read models; MetaModel becomes upstream for maturity/pillar semantics |
| Completeness | 178 | Read-time completeness evaluation and requirement-change impact preview |

## Open Questions

1. Should one-pager settings get a dedicated permission pair if role matrices ever diverge from the metamodel gate?
2. Does completeness at scale (spec 178) need a cache projector, or do set-based queries suffice?

## Boundary Health

- **Zero cross-context imports**: machine-enforced by `/backend/internal/onepagers/architecture_boundary_test.go` — only `internal/shared`, other contexts' `publishedlanguage` packages, and shared eventstore/database infrastructure are importable
- **Published-language consumer count**: currently zero external subscribers; growth is deliberate, per-slice
- **Catalog binding integrity**: catalog entries are definition-only metadata today; once ports land (spec 177), per-subject-type integration tests must bind every entry to a supplier read contract

## Architecture Notes

### Implementation Location
`/backend/internal/onepagers/`

### Key Packages
- `domain/aggregates/` - OnePagerConfiguration aggregate
- `domain/catalog/` - code-owned built-in field catalog per subject type
- `domain/valueobjects/` - SubjectType, FieldID, FieldType, CustomField, SelectionOption
- `domain/events/` - the eleven configuration events
- `application/` - commands, handlers, read-model projector, read models
- `infrastructure/` - API routes/handlers/DTOs, event-sourced repository
- `publishedlanguage/` - event type constants

### Technical Patterns
- **CQRS with Event Sourcing** in the shared event store; own PostgreSQL schema `onepagers` with RLS (migration 122)
- **Read model**: `onepagers.one_pager_configurations`, one row per (tenant, subject type), unique constraint on `(tenant_id, subject_type)`
- **Lazy default creation** on first read, mirroring the MetaModel configuration precedent
- **Optimistic concurrency** via aggregate version, conflicts surfaced as 409

### API Style
- REST Level 3 with HATEOAS; write affordances advertised only to `PermMetaModelWrite` holders
- One endpoint per configuration command under `/api/v1/one-pagers/configurations/{subjectType}`
