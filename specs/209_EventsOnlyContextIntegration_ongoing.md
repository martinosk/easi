# 209 — Events-Only Context Integration

> **Status:** ongoing
> **Depends on:** 206 (Composition-Root Dependency Guard), 207 (Direction-Derived Reads Owned by Direction), 208 (One-Pager Completeness Served by OnePagers)
> **Amends:** 206 (rule 9 and design decision 8 — declared bridges are no longer permitted), 207 (Architecture Direction bridges), 178/189 (OnePagers built-in and relation fields), 134 (Architecture Guard Tests), 039 (tenant provisioning), 127 (edit-grant auto-invitation)

---

## Problem Statement

After 206–208 the bounded-context graph is acyclic, but eight declared composition-root bridges still let one context read another context's read models or services at request time, and `architecture_sql_test.go` allowlists two cross-schema SQL reads — one of which (`auth → platform`) closes a cycle with Platform's dispatch of Auth's `CreateInvitation`. Every such read is a hidden coupling that will break the moment a context is merged, split or renamed.

The owner intends a major domain-model clean-up (merging and creating bounded contexts). For that to be safe, every context must depend on every other context **only** through the supplier's published language: its published events and its published commands. No context may read another context's read models, services or tables — in code or in the database.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architect (owner)** | Can merge or split a bounded context by moving its directory and its published language, with nothing else to untangle |
| **Backend Developer** | A failing test — not a review — names any import, SQL statement or wiring that reaches into another context |
| **End users** | Notice no change: grants still invite non-users, login still resolves the tenant by e-mail domain, one-pagers still render their fields, imports still work |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Bounded contexts integrate only through published events and commands

  Scenario: A read of another context's read model fails the build
    Given a composition-root file imports a package of a context other than its infrastructure/api package
    When the architecture tests run
    Then they fail naming the file and the import

  Scenario: A cross-schema SQL statement fails the build
    Given runtime code in context A references a table in schema B
    When the architecture tests run
    Then they fail naming the file and the table, with no allowlist

  Scenario: Tenant login and invitations work from Auth's own data
    Given a tenant provisioned before this change
    When a user of that tenant signs in with an e-mail of one of its domains
    Then the tenant and its OIDC configuration are resolved from Auth's local cache

  Scenario: Provisioning a tenant invites its first admin
    When a platform operator creates a tenant with a first-admin e-mail
    Then Auth creates an admin invitation for that e-mail by reacting to TenantCreated

  Scenario: Edit grant for a non-user still invites them
    Given an admin grants edit access to an e-mail without an account or pending invitation
    When the grant is created
    Then Auth creates a stakeholder invitation and the response reports invitationCreated = true
    And granting to an existing user or an already-invited e-mail reports invitationCreated = false

  Scenario: One-pagers render from OnePagers' own caches
    Given an application with a description and experts, realizations and origin links
    When its one-pager is rendered
    Then name, description, experts and relation fields are identical to today's response
    And renaming the application, changing an expert or a realization is reflected on the next read

  Scenario: Composition is no longer shown on the enterprise-capability one-pager
    When an enterprise-capability one-pager is rendered
    Then it has no "included-capabilities" field

  Scenario: Existing tenants see complete data after deployment
    When the migrations have run and the backend starts
    Then every new cache is populated from its source tables without waiting for any event
```

---

## Business Rules & Invariants

1. **Two integration channels, both owned by the supplier.** Context Y may depend on context X only by (a) subscribing to events named by constants in `X/publishedlanguage` or (b) dispatching a command whose struct is declared in `X/publishedlanguage` (with `CommandName()`; X aliases it internally, as Auth does for `CreateInvitation`). A command result carries at most the created ID.
2. **No composition-root bridges.** Every production file under `internal/infrastructure/api/` (including subdirectories) imports from a context only its `infrastructure/api` package. The router passes shared infrastructure (database, buses, HATEOAS, auth middleware) and nothing else. The bridge registry of 206 is deleted.
3. **No cross-schema SQL at runtime.** `allowedSchemaAccess` is removed; any schema-qualified reference to another context's schema in runtime code fails. Migrations numbered 139 and above may reference more than one schema only when their filename contains `backfill` (one-time seeding of a cache).
4. **`shared/` and `infrastructure/` import no context package.** `shared/audit` becomes the `audit` context (generic subdomain).
5. **Actor identity and role come from the request context**, never from Auth's read models.
6. **Every cache is tenant-scoped, RLS-protected, maintained by a projector on the supplier's published events, and backfilled by a migration from the supplier's tables.**
7. **Event subscriptions and cache projectors are wired inside the consuming context.**
8. **Dependency graph = published-language import edges only**, direction importer → imported; it must be acyclic.
9. **Auth is downstream of Platform.** Auth caches tenants, domains and OIDC configuration from `TenantCreated` and reacts to it to invite the first admin. Platform dispatches no Auth command.
10. **Enterprise-capability composition is not published.** Enterprise capabilities are expected to be removed; OnePagers drops the `included-capabilities` relation field instead of caching composition.
11. **OnePagers caches the complete published attribute set of every subject**, keyed by the supplier's published attribute names, never by catalogue entry — the catalogue-entry → attribute mapping lives only in the render-time adapter. Adding a built-in field over an attribute the supplier already publishes is a catalogue change with no projector change and no migration. A genuinely new supplier attribute is published in the supplier's events and shipped with an OnePagers `*backfill*` migration for existing subjects, exactly like any other cache column.

---

## Acceptance Criteria

- [x] `backend/internal/architecture_bridges_test.go` is replaced by `architecture_integration_test.go` with `TestCompositionRootOnlyRegistersRoutes`, `TestSharedAndInfrastructureImportNoContext`, `TestProductionCodeDoesNotImportTestSupport`, `TestContextDependencyGraphIsAcyclic` (import edges only) and `TestNewMigrationsCrossSchemasOnlyInBackfills`; `allowedSchemaAccess` is removed from `architecture_sql_test.go`; all pass together with `TestNoCrossBoundedContextImports` and `TestPublishedLanguageContractsPurity`.
- [x] `infrastructure/api/` contains `router.go`, `middleware/` and tests only; all `*_bridges.go` and `onepager_*_adapters.go` files are deleted.
- [x] **Auth / Platform:** `TenantCreated` carries `discoveryUrl`, `issuerUrl`, `clientId`, `authMethod`, `scopes`; Auth keeps `auth.tenant_cache`, `auth.tenant_domain_cache`, `auth.tenant_oidc_cache` (migration 139, backfill 140) fed by a `TenantCreated` projector; `TenantOIDCRepository` and `TenantDomainChecker` read them; an Auth reactor on `TenantCreated` creates the first-admin invitation; Platform no longer imports `auth/publishedlanguage`; `POST /tenants/{id}/invitations` is replaced by `POST /auth/invitations` guarded by the platform-admin key middleware moved to `shared/api`; Auth publishes `EnsureInvitation` (creates an invitation unless the e-mail has a user or a pending invitation; rejects disallowed domains; result `CreatedID` empty when nothing was created).
- [x] **Access Delegation:** ports `UserEmailLookup`, `InvitationChecker`, `DomainAllowlistChecker`, `InvitationRequester` and `ArtifactNameResolverDeps` are deleted; auto-invitation dispatches `authPL.EnsureInvitation`; artifact display names come from `accessdelegation.artifact_name_cache` (migration 141, backfill 142) fed by the Created/Updated/Deleted events of capabilities, business domains, components, vendors, acquired entities, internal teams and views.
- [x] **Architecture Views:** `UserRoleChecker` is deleted; `ChangeViewVisibility` uses the actor's role from context; `RegisterCommands` is called from the context's own route setup.
- [x] **Enterprise Architecture:** `BusinessDomainNameLookup` is deleted; domain names come from `enterprisearchitecture.business_domain_name_cache` (migration 143, backfill 144) fed by `BusinessDomainCreated/Updated/Deleted`; `links.go` no longer emits `x-direction` / `x-composition`; the empty `application/services` directory is removed; `ea_realization_cache_projector.go` uses published-language constants.
- [x] **Capability Mapping:** `DirectStrategyPillarsGateway` is deleted; test fixtures use the local pillar cache.
- [x] **Architecture Direction:** capability existence and effective-domain checks read `capability_node_cache`; enterprise-capability checks read `enterprise_capability_cache`; component and business-domain existence read the existing `architecturedirection.reference_name_cache`, whose rows are now removed on `ApplicationComponentDeleted` / `BusinessDomainDeleted`; direct-realization lookup reads `architecturedirection.realization_cache` (direct realizations only — inherited rows have no published identity; migration 145 adds the table and makes `capability_node_cache.maturity_value NOT NULL DEFAULT 0`; backfill 146 seeds it and reconciles the reference-name cache); `capability_node_cache.Insert` writes `maturity_value`; `composition_wiring.go` exports nothing; `GET /enterprise-capability-compositions` items carry `_links.x-composition` and `_links.x-direction`; `RoutesDeps` has no lookup fields.
- [x] **OnePagers:** `SubjectExistenceChecker`, `BuiltInFieldSource` and `MaturityScaleSource` are implemented inside OnePagers over `onepagers.one_pager_subject_index` (extended with a `built_in_fields` JSONB column holding the complete published attribute set), `onepagers.subject_relation_cache`, `onepagers.business_domain_name_cache` and `onepagers.maturity_scale_cache` (migration 147, backfill 148), fed by AM, CM, EA and MetaModel events (expert names travel on the expert events, so no user cache is needed); the enterprise-capability one-pager has no `included-capabilities` entry; `OnePagersRoutesDeps` has no `Subjects`, `BuiltInFields`, `MaturityScale` fields; rendered one-pagers, completeness and quality list are unchanged for the remaining fields.
- [x] **Importing:** Architecture Modeling, Capability Mapping and Value Streams publish their import commands in their published language (`CreateApplicationComponent`, `CreateComponentRelation`, `CreateCapability`, `UpdateCapabilityMetadata`, `LinkSystemToCapability`, `AssignCapabilityToDomain`, `CreateValueStream`, `AddStage`, `AddStageCapability`) and alias them internally; the gateway implementations move into `importing/infrastructure/adapters`; `import_gateway_contracts.go` files and the supplier-side gateway adapters are deleted.
- [x] **Audit:** `shared/audit` moves to `internal/audit` with its own `infrastructure/api` route setup.
- [x] **Arch Assistant / Auth session:** `GET /auth/sessions/current` emits `x-assistant-status` (never `x-assistant` / `x-assistant-write`); `GET /assistant/status` emits `x-conversations` and `x-conversations-write` when configured; the frontend chat entry point follows those links.
- [x] **Frontend:** enterprise-capability table and detail read `x-composition` / `x-direction` from the composition summaries; chat availability follows the new links; MSW stubs updated; `npm run build`, `npm test -- --run`, `npm run lint` green.
- [x] Swagger regenerated; `go test ./...` green; integration tests for every backfill.
- [x] `docs/architecture/README.md`, `docs/backend/cross-context-events.md` and the canvases describe the events-only rule and every consumed event per context; the bridge registry section is removed.

---

## Architecture

### Ownership

Each consuming context owns its caches, projectors and backfills. Suppliers own their published events and commands. The composition root owns nothing but route registration.

### Domain Model

No aggregate changes. New published command `authPL.EnsureInvitation`. `TenantCreated` gains OIDC fields. Auth gains a `TenantCreated` reactor (first-admin invitation) and tenant caches.

### API Surface

- Removed: `POST /tenants/{id}/invitations` (Platform). Added: `POST /auth/invitations` (platform-admin key, body `{tenantId, email, role}`).
- Changed: enterprise capability DTO `_links` lose `x-direction`, `x-composition`; composition summary items gain them.
- Changed: session links `x-assistant`/`x-assistant-write` → `x-assistant-status`; assistant status links gain `x-conversations-write`.
- Changed: enterprise-capability one-pager loses `included-capabilities`.

### Persistence

Migrations 139–148 as listed in the acceptance criteria; every table tenant-scoped with the standard RLS policy; every backfill idempotent (`INSERT … SELECT … ON CONFLICT DO UPDATE`).

### Cross-Context Integration (target graph, importer → imported)

```
accessdelegation      → architecturemodeling, architectureviews, capabilitymapping, auth
archassistant         → architecturedirection, architecturemodeling, architectureviews, auth, capabilitymapping, enterprisearchitecture, metamodel, platform, valuestreams
architecturedirection → architecturemodeling, auth, capabilitymapping, enterprisearchitecture
architecturemodeling  → auth
architectureviews     → architecturemodeling, auth
audit                 → auth
auth                  → platform
capabilitymapping     → architecturemodeling, auth, metamodel
enterprisearchitecture→ architecturemodeling, auth, capabilitymapping, metamodel
importing             → architecturemodeling, capabilitymapping, valuestreams
metamodel             → auth
onepagers             → architecturemodeling, auth, capabilitymapping, enterprisearchitecture, metamodel
platform              → (none)
valuestreams          → auth, capabilitymapping
```

---

## Design Decisions

1. **Published commands are the second channel** — user decision 2026-08-30. Importing's saga and the invitation flows need synchronous writes with created IDs; choreographing them would push consumer correlation into supplier events.
2. **Auth is downstream of Platform** — user decision 2026-08-30. Provisioning precedes identity; `TenantCreated` already carries what Auth needs.
3. **Composition is not published** — user decision 2026-08-30: enterprise capabilities are likely to be removed, so OnePagers drops the field rather than AD materialising composition.
4. **Migrations may read other schemas only to backfill** — a cache must start complete; the rule is enforced by filename so the intent is visible in the migration list.
5. **Actor role from context, not from Auth** — the session already carries it; a read model lookup was never needed.
6. **`EnsureInvitation` encapsulates Auth's invitation policy** — user-exists, pending-invitation and domain-allowlist checks belong to the context that owns users and invitations; Access Delegation only decides the role.
7. **Auth's tenant caches carry no RLS policy** — they are read by e-mail domain before any tenant context exists (login), exactly like the `platform.tenants` / `tenant_domains` / `tenant_oidc_configs` tables they mirror; every query filters by tenant or domain explicitly. This is the one deliberate exception to rule 6.
8. **Platform now publishes `TenantCreated`** — before this spec the create-tenant handler discarded the aggregate's events and `PlatformRoutesDeps` had no event bus, so MetaModel's and Arch Assistant's `TenantCreated` subscribers had never run. Provisioning of default configurations and of the first-admin invitation now happens in the same request; a failing subscriber fails the request, as the old synchronous `inviteFirstAdmin` did.
9. **Architecture Direction reuses `reference_name_cache`** for component and domain existence (rows now removed on deletion) rather than adding a second reference table, and its realization cache holds direct realizations only — inherited realizations get their identity inside Capability Mapping's read model and are not published.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Many small caches | More projector code and tables | Uniform pattern, backfilled, integration-tested |
| Auth caches tenant data | Eventual consistency for tenant changes; no RLS on the caches | Synchronous in-process bus; tenants are immutable after creation today; explicit tenant/domain filters on every query |
| Tenant creation now runs every `TenantCreated` subscriber inline | A failing subscriber fails `POST /tenants` | Same strictness the synchronous first-admin invitation had; subscribers are idempotent projectors |
| OnePagers replicates subject fields | Duplicated data | Read-side only, fed by the same events the subject index already consumes |
| `/tenants/{id}/invitations` moves | Ops scripts referencing it must change | No frontend caller; documented in release notes |

---

## Checklist

- [x] Specification ready — approved in session 2026-08-30 ("fix all that's necessary")
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [ ] User sign-off
