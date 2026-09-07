# Documentation Index

Quick reference for navigating EASI documentation.

## By Role/Task

| Working on... | Read this |
|---------------|-----------|
| Planning a feature or bounded-context change | [docs/architecture/ROADMAP.md](architecture/ROADMAP.md) |
| Designing a feature family before writing its specs | [docs/specs/](specs/) — Phase-1 design docs |
| Backend API endpoint | [docs/backend/api-standards.md](backend/api-standards.md) |
| Backend event handling | [docs/backend/standard-patterns.md](backend/standard-patterns.md) |
| Cross-context events / bridges | [docs/backend/cross-context-events.md](backend/cross-context-events.md) |
| Backend anti-patterns | [docs/backend/antipatterns.md](backend/antipatterns.md) |
| Backend tests | [docs/backend/testing.md](backend/testing.md) |
| Database migration | [docs/backend/database.md](backend/database.md) |
| Frontend component | [docs/frontend/standard-patterns.md](frontend/standard-patterns.md) |
| CodeScene architectural components | [docs/architecture/components.csv](architecture/components.csv) (generated — see [architecture README](architecture/README.md#codescene-components)) |

## By Bounded Context

| Context | Canvas | Classification | Status |
|---------|--------|----------------|--------|
| CapabilityMapping | [docs/architecture/CapabilityMapping.md](architecture/CapabilityMapping.md) | Core | Implemented |
| ValueStreams | [docs/architecture/README.md](architecture/README.md) | Core | Implemented |
| ArchitectureModeling | [docs/architecture/ArchitectureModeling.md](architecture/ArchitectureModeling.md) | Supporting | Implemented |
| ArchitectureViews | [docs/architecture/ArchitectureViews.md](architecture/ArchitectureViews.md) | Supporting | Implemented |
| MetaModel | [docs/architecture/MetaModel.md](architecture/MetaModel.md) | Supporting | Implemented |
| AccessDelegation | [docs/architecture/README.md](architecture/README.md) | Supporting | Implemented |
| ArchAssistant | [docs/architecture/ArchAssistant.md](architecture/ArchAssistant.md) | Supporting | Implemented |
| ArchitectureDirection | [docs/architecture/ArchitectureDirection.md](architecture/ArchitectureDirection.md) | Core | Implemented |
| OnePagers | [docs/architecture/OnePagers.md](architecture/OnePagers.md) | Supporting | Implemented |
| Releases | [docs/architecture/Releases.md](architecture/Releases.md) | Generic | Implemented |

Full context map: [docs/architecture/README.md](architecture/README.md)

## Design Documents (Phase 1)

Planning has three tiers, per [`easi-spec-driven-development`](/.claude/skills/easi-spec-driven-development/SKILL.md):

1. **[ROADMAP.md](architecture/ROADMAP.md)** — what has been decided and in what order. Settled decisions, standing invariants, horizon plan. One file, whole product.
2. **`docs/specs/{family}.md`** — how one feature family or horizon move is designed, and why it is cut into the slices it is. Holds the research, the ubiquitous language, the numbered decisions (D1, D2, …) that every slice in the family conforms to, and the slice map. Approved by a human before any spec is written.
3. **`/specs/{NNN}_{Name}_{status}.md`** — one independently deployable vertical slice each.

The middle tier exists so a decision shared by many slices is stated once and can be amended by name — spec 186's header reads *"Supersedes design decision D7 of Configurable One-Pagers"*.

| Design doc | Covers | Specs |
|------------|--------|-------|
| [capability-journeys.md](specs/capability-journeys.md) | Journeys, TIME assessment per realisation, realisation roles, board lenses | 180–184, 195–197, 205 |
| [configurable-one-pagers.md](specs/configurable-one-pagers.md) | One-pager configuration, fields, facts, completeness | 175, 177, 185, 186, 188 |
| [enterprise-capability.md](specs/enterprise-capability.md) | Retiring the enterprise capability; maturity as a journey; one TIME vocabulary | 210–213 |

## Core Rules

Always apply: [CLAUDE.md](/CLAUDE.md) (root)

## API Documentation

- OpenAPI spec served at: `http://localhost:8080/swagger/index.html`

## User Documentation

- IDP Configuration: [docs/user/idp-configuration.md](user/idp-configuration.md)
