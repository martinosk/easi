# Bounded Context Canvas: Stewardship

## Name
**Stewardship**

## Purpose
Record who answers for which concern in which business domain, and what currently needs their attention. The write side is one fact — a stewardship assigns one user to one concern in one domain — and everything else the context serves is a projection over other contexts' published events (spec 226 ships the write side; the attention read side follows in the later slices of the [personal-home design doc](../specs/personal-home.md)). The context speaks the language of accountability, not of the model: it never learns what a capability or an application is beyond the id and name it caches.

**Key Stakeholders:**
- Admins and Architects (assign and release stewards, see which concerns nobody answers for)
- Stewards of any role (see the domains and concerns they answer for)
- Domain Architects (steward of last resort for every unassigned concern in their domain)

## Strategic Classification

### Domain Importance
**Supporting Domain** — accountability is what makes the platform's findings actionable, but the fact itself is small and its rules are few. The investment goes into the read side that ranks findings per steward, which is pure projection; the write side has the Access Delegation shape (one aggregate, event-fed caches, affordances on the artifact).

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Concern** | One kind of thing a domain's landscape can be wrong about: `ownership`, `assessment`, `documentation`, `planning`, `structure`. Fixed in code; a concern exists only because a check produces its attention items |
| **Stewardship** | One user accountable for one concern in one domain; at most one per (domain, concern), any number per user |
| **Steward** | The user holding a stewardship; distinct from the domain architect (curates the model) and from an application owner (fixes their own application) |
| **Assign / Release** | The two commands: assign creates a stewardship or changes its steward; release ends it |
| **Unassigned** | A concern with no stewardship in a domain |
| **Fallback** | The domain architect, derived on read as steward of last resort for an unassigned concern; never recorded |
| **Domain Cache** | Local copy of business-domain existence, name and domain architect id fed by Capability Mapping events |
| **User Cache** | Local copy of user id, name and active status fed by Auth events |

## Inbound Communication

**Commands** (from Frontend/API):
- `AssignSteward`, `ReleaseSteward`

**Queries** (from Frontend/API): stewardships of a domain — every concern in fixed order with its steward, attribution and fallback, with assign and release affordances for callers holding `domains:write`

**Events** (from other contexts):
- From **Capability Mapping**: `BusinessDomainCreated`, `BusinessDomainUpdated` → domain cache; `BusinessDomainDeleted` → domain cache and release of every stewardship of the domain
- From **Auth**: `UserCreated`, `UserDisabled`, `UserEnabled` → user cache

### Relationship Types
- **Conformist** to Capability Mapping (domain lifecycle events into a local, backfilled cache; the events already carry everything the context needs)
- **Published Language** from Auth (user events into a local cache, permission constants)

## Outbound Communication

**Events** (published to event bus):
- `StewardAssigned` (domain, concern, steward, actor, time — states the current steward of the pair), `StewardReleased` (domain, concern, actor, time)

No other context consumes these events. Capability Mapping's domain responses carry the `x-stewardships` link, added by the shared HATEOAS helper as a URL convention; Capability Mapping has no code dependency on this context.

## Business Rules

1. Concerns are the fixed five; any other value is invalid.
2. At most one live stewardship per (domain, concern); assigning to a stewarded pair replaces the steward in one step.
3. A steward is an existing, active user of the tenant, any role; never a team, never free text.
4. Assigning and releasing require `domains:write`; reading requires `domains:read`; an edit grant on the domain confers neither.
5. Every assignment is attributed to its actor and time.
6. Deleting a domain releases its stewardships, attributed to the system.
7. Disabling a user does not release their stewardships.
8. Assigning the current steward again, or releasing an unassigned concern, changes nothing and records nothing.

## Design Constraints

1. Every existence check (domain, user) reads a local, event-fed, backfilled cache; the context issues no query to another context at request time (spec 209).
2. Concerns live in code, not in MetaModel: a tenant-defined concern with no check behind it is accountability for nothing (design doc D4).
3. The fallback to the domain architect is computed on read from the domain cache; recording it would duplicate a Capability Mapping fact.
4. The aggregate has its own identity; the (domain, concern) pair is located through the read model and its uniqueness is enforced at the handler with a unique index on the live-stewardship read model as backstop (a release removes the row, so the index covers only live pairs) (the one-per-owner shape of spec 216).

## Open Questions

1. Whether the domain architect remains a Capability Mapping field or becomes a "model" stewardship here once the attention read side needs to react to architect changes (design decision 8 of spec 226 defers this; it is a roadmap amendment, not a spec).

## Boundary Health
- **Dependency direction**: this context imports only `capabilitymapping/publishedlanguage` and `auth/publishedlanguage`; no context imports this one — `TestContextDependencyGraphIsAcyclic` fails on the first edge back
- **Cache freshness**: a domain rename, architect change or user status change is visible in the stewardships query within the event round-trip
- **Guard**: `TestNoCrossBoundedContextImports`, `TestSQLSchemaOwnership` — only published language and the `stewardship` schema cross or hold the boundary

## Architecture Notes

### Implementation Location
`/backend/internal/stewardship/`

### Key Packages
- `domain/aggregates/` — Stewardship
- `domain/valueobjects/` — Concern, StewardRef
- `application/handlers/` — assign and release, reference checks against the caches
- `application/readmodels/` — stewardship read model, domain cache, user cache
- `application/projectors/` — projector for own events, cache projectors for Capability Mapping and Auth events, domain-deletion reactor
- `infrastructure/api/` — routes and handlers; `SetupRoutes` needs only shared infrastructure (router, buses, database, HATEOAS, auth middleware)
- `publishedlanguage/` — event name constants

### API Style
- REST Level 3 with HATEOAS
- Endpoints: `GET /stewardships?domainId=` (every concern of a domain, with fallback), `GET`/`PUT`/`DELETE /stewardships/{domainId}/{concern}` (read, assign, release)
- Affordances: `x-assign` / `x-release` per concern and `x-candidates` on the collection for callers holding `domains:write`; `x-stewardships` on every domain response for readers of domains, added by the shared HATEOAS helper
- Not in the Arch Assistant tool catalog, as Access Delegation is not

### Persistence
Schema `stewardship`; tables: `stewardships`, `domain_cache`, `user_cache`. All tenant-scoped with row-level security; caches seeded by `backfill` migrations.
