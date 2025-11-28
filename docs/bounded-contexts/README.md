# EASI Bounded Contexts

This directory contains Bounded Context Canvases for all contexts in the EASI enterprise architecture platform. Each canvas follows the [DDD Crew Bounded Context Canvas](https://github.com/ddd-crew/bounded-context-canvas) template.

## Overview

EASI is built using Strategic Domain-Driven Design principles with clear bounded context boundaries. The system is organized into 5 bounded contexts:

### 1. Architecture Modeling (Implemented)
**Purpose:** Manage IT application landscape - what systems exist and how they interact.

**Location:** `/backend/internal/architecturemodeling/`

**Key Responsibilities:**
- Application component inventory
- Component relationships and dependencies
- System integration mapping

**Strategic Classification:** Supporting Domain

[Full Canvas →](./ArchitectureModeling.md)

---

### 2. Architecture Views (Implemented)
**Purpose:** Create and manage visual representations of architecture for different stakeholder perspectives.

**Location:** `/backend/internal/architectureviews/`

**Key Responsibilities:**
- Multiple architecture views
- Custom layouts and styling
- Visual presentation management

**Strategic Classification:** Supporting Domain

[Full Canvas →](./ArchitectureViews.md)

---

### 3. Capability Mapping (Implemented + Business Domains Extension)
**Purpose:** Map business capabilities to IT systems, track maturity, dependencies, and strategic alignment.

**Location:** `/backend/internal/capabilitymapping/`

**Key Responsibilities:**
- Business capability taxonomy (L1-L4 hierarchy)
- Capability-to-system realization mapping
- Capability maturity and strategic alignment
- Business domain groupings (L1 capabilities)
- Capability dependencies

**Strategic Classification:** Core Domain

[Full Canvas →](./CapabilityMapping.md)

---

### 4. Releases (Implemented)
**Purpose:** Track and communicate EASI platform releases and version history.

**Location:** `/backend/internal/releases/`

**Key Responsibilities:**
- Version tracking
- Release notes
- Platform version reporting

**Strategic Classification:** Generic Subdomain

[Full Canvas →](./Releases.md)

---

### 5. Enterprise Strategy (Future)
**Purpose:** Govern strategic architectural decisions about business domain evolution.

**Location:** `/backend/internal/enterprisestrategy/` (future)

**Key Responsibilities:**
- Domain consolidation (merging domains)
- Domain decomposition (splitting domains)
- Strategic area retirement
- Strategic decision audit trail
- Impact analysis for structural changes

**Strategic Classification:** Core Domain

[Full Canvas →](./EnterpriseStrategy.md)

---

## Context Map

### Integration Patterns

```
┌─────────────────────────────────────────────────────────────────┐
│                    Architecture Modeling                         │
│                  (Application Component SoR)                     │
│                  [Supporting Domain]                             │
└─────────────┬────────────────────────────┬──────────────────────┘
              │ Events                     │ Events
              │ (ComponentDeleted)         │ (ComponentDeleted)
              ↓                            ↓
    ┌─────────────────────┐      ┌──────────────────────┐
    │ Architecture Views  │      │ Capability Mapping   │
    │  (Views/Layouts)    │      │ (Cap-to-System Map)  │
    │ [Supporting Domain] │      │   [Core Domain]      │
    └─────────────────────┘      └──────────┬───────────┘
                                            │ Commands
                                            │ (Strategic Ops)
                                            ↓
                               ┌────────────────────────┐
                               │ Enterprise Strategy    │
                               │ (Strategic Governance) │
                               │   [Core Domain]        │
                               │      [Future]          │
                               └────────────────────────┘

                    ┌─────────────────┐
                    │    Releases     │
                    │ (Version Info)  │
                    │ [Generic Domain]│
                    └─────────────────┘
                    (Isolated - no integration)
```

### Relationship Types

| Upstream Context | Downstream Context | Relationship Type | Integration Pattern |
|------------------|-------------------|-------------------|---------------------|
| Architecture Modeling | Architecture Views | Customer-Supplier | Event-driven (component deletions) |
| Architecture Modeling | Capability Mapping | Customer-Supplier | Event-driven (component deletions) + Query (read models) |
| Capability Mapping | Enterprise Strategy | Partnership | Command-driven (strategic ops) + Event-driven (tracking) |

### Key Integration Points

**Architecture Modeling → Architecture Views**
- Events: `ApplicationComponentDeleted`, `ComponentRelationDeleted`
- Purpose: Keep views consistent when components are removed

**Architecture Modeling → Capability Mapping**
- Events: `ApplicationComponentDeleted`
- Queries: Read `ApplicationComponentReadModel` for system realization
- Purpose: Link capabilities to systems, cleanup when systems deleted

**Capability Mapping → Enterprise Strategy** (Future)
- Commands: `CreateBusinessDomain`, `AssignCapabilityToDomain`, etc.
- Events: `CapabilityAssignedToDomain`, `BusinessDomainDeleted`
- Purpose: Strategic domain operations coordinated via saga pattern

## Domain Classification

### Core Domains (Competitive Advantage)
1. **Capability Mapping** - Sophisticated capability-to-system mapping with strategic analysis
2. **Enterprise Strategy** (future) - Strategic governance of domain evolution

### Supporting Domains (Essential but not differentiating)
1. **Architecture Modeling** - Standard application inventory
2. **Architecture Reviews** - View management and visualization

### Generic Domains (Commodity)
1. **Releases** - Simple version tracking

## Context Autonomy

Each bounded context has:
- **Own Event Store**: Separate event streams in PostgreSQL
- **Own Read Models**: Denormalized projections for query performance
- **Own Aggregates**: Independent transactional boundaries
- **Own API**: REST Level 3 with context-specific endpoints
- **Tenant Isolation**: Multi-tenancy at context level (except Releases)

**No Shared Databases**: Contexts communicate via events and queries, never direct database access.

**No Circular Dependencies**: Dependency graph is acyclic (Architecture Modeling → Capability Mapping → Enterprise Strategy).

## Implementation Status

| Context | Status | Location | CQRS/ES |
|---------|--------|----------|---------|
| Architecture Modeling | ✅ Implemented | `/backend/internal/architecturemodeling/` | Yes |
| Architecture Reviews | ✅ Implemented | `/backend/internal/architectureviews/` | Yes |
| Capability Mapping | ✅ Implemented | `/backend/internal/capabilitymapping/` | Yes |
| Business Domains (in Capability Mapping) | 📝 Specified (specs 053-058) | `/backend/internal/capabilitymapping/` | Yes |
| Releases | ✅ Implemented | `/backend/internal/releases/` | No (simple CRUD) |
| Enterprise Strategy | 📝 Specified (spec 059) | `/backend/internal/enterprisestrategy/` (future) | Yes (future) |
s
