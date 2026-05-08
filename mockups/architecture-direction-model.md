# Target Architecture — Conceptual Model

The conceptual model behind target-architecture modelling and the path from current to target. The output of mockup-driven exploration in `architecture-direction.html`. Precedes any feature specs.

## Why

EA at DFDS today: ~1300 physical capabilities, six business domains, multiple architects working largely in silos. Modelling the AS-IS architecture works fine — duplication across domains is honest. Modelling the TARGET architecture and the path to it does not, because the existing tooling gives a single "enterprise capability" concept that conflates two genuinely different activities.

This document names the layered model that resolves the conflation.

## What this is for (and what it isn't)

EASI's job is to make **informed alignment decisions cheap**. The audience is *anyone* in the organisation making a daily decision — a product manager picking which app to invest in, a domain owner deciding whether to build something, an engineer scoping a project — who needs to answer one question quickly:

> *"Is what I'm about to do leading us toward the target architecture, or away from it?"*

The model exists to make that question answerable in five seconds, by a non-architect, on a regular work day.

It is **not** a project-planning tool. Not a Gantt chart. Not a delivery tracker. Not a roadmap with target dates. Not a portfolio of work packages.

> **Test for any proposed addition to the model:** does it help an individual make a daily alignment decision in five seconds? If not, drop it.

This test rules out: fine-grained dates, capability increments, work-package decomposition, plateau time-spans, formal gap analysis, realisation tracking. Those belong in adjacent tools — not here.

## Two layers

**Physical capability** — operationally real. Belongs to exactly one business domain. Has a local name (e.g. *Customer Care* in Passenger, *Customer Service* in Terminal), its own processes, people, data, and is realized by one or more applications. Already exists in EASI as the `Capability` aggregate.

**Logical capability** — an abstract label that groups multiple physicals for *reasoning* purposes. Has a canonical name (e.g. *Customer Service*). Has no processes, people, or data of its own — it is descriptive, not operational. Used to compare across domains, run portfolio analysis, and assign a standard application across a distributed business reality.

> Today's EASI `EnterpriseCapability` is a Logical capability, mis-named. Migrate by renaming.

## Relationships

```
BusinessDomain      1 ──< N  PhysicalCapability
PhysicalCapability  N >──< M Application            (realizes — current state)
LogicalCapability   1 ──< N  PhysicalCapability     (each physical maps to 0..1 logical)
Direction           1 ──< N  PhysicalCapability     (acts on)
StandardApp         1 ────── 1 LogicalCapability    (designation)
StandardApp         1 ────── 1 Application
```

## Two modelling activities

| | What changes | Subject of change | What changes in reality |
|---|---|---|---|
| **1. Physical consolidation** | The business architecture | Physical capabilities | Capabilities merge / split / move; processes and systems realign |
| **2. Logical grouping** | The landscape map | A label spanning physicals | Nothing physical — only how we reason about the picture |

Both activities are legitimate and neither contains the other. Most strategic conversations involve a mix.

## New concepts

### Direction

A change proposal acting on physical capabilities.

| Field | Values |
|---|---|
| `type` | `consolidate` (N → 1) · `decompose` (1 → N) · `stay` (confirmed no change) |
| `placements` | Target domain(s) for the resulting physical capability/ies |
| `horizon` | `now` · `next` · `later` |
| `status` | `draft` → `proposed` → `agreed`  (or `rejected`) |
| `narrative` | Stakeholder-facing 1–2 sentences |
| `sourceCapabilities` | The physical capabilities the direction acts on |

`stay` is rare — used only when the group has explicitly evaluated and decided no change is needed. Most "no change" cases are simply absence of a Direction.

### Standard App designation

Attached to a Logical capability. Says: "for this logical grouping, the standard application is X."

| Field | Values |
|---|---|
| `logicalCapability` | The grouping this standard applies to |
| `application` | The designated standard app |
| `horizon` | `now` · `next` · `later` |
| `status` | `draft` → `proposed` → `agreed`  (or `rejected`) |
| `narrative` | Why this app, what it covers |

## Key consequences

**1. Type-2 (app-standardize, business stays distributed) is a Logical-layer concern.** It is a Logical capability with a Standard App designation — no Direction. The earlier mockup's `direction.scope = application-only` collapses into "no direction; logical grouping with standard app." This is cleaner: physical reality is untouched, the standard exists at the level of abstraction where it actually applies.

**2. A Discover candidate has two valid resolutions, not one.** A multi-domain theme can resolve into:

- a **physical consolidation Direction** (changes the world), or
- a **logical grouping** (organizes the picture, optionally with a standard app)

Plus the third option: reject. The group chooses explicitly. Different stakeholders, different costs, different timelines.

**3. Phased paths are sequences, not new concepts.** A theme can be: nothing today → a Logical grouping with standard app (Phase 1 / Now) → also a physical consolidation Direction (Phase 2 / Later). The model needs no new concepts; horizons make the sequence visible. This matches the realistic EA pattern of "standardize the application now to buy time, consolidate the business later when the org is ready."

## What this leaves alone

Existing EASI concepts that stay as-is:

- `Capability` aggregate → unchanged. *Is* the Physical capability.
- `BusinessDomain` → unchanged.
- `Application`, `CapabilityRealization` → unchanged.
- `EnterpriseCapability` → **renamed to `LogicalCapability`**. Data migrates 1:1.
- `EnterpriseCapabilityLink` → renamed to mapping; semantics unchanged.
- Strategic importance, target maturity, TIME framework attached to EnterpriseCapability → carry over to LogicalCapability without modification.

## Decision flows

```
Discover candidate (multi-domain theme, no existing decision)
    ├──► Promote to physical consolidation Direction
    ├──► Promote to logical grouping  (+ optional Standard App designation)
    └──► Reject

Logical capability
    ├──► Add / replace Standard App designation
    ├──► Convert to / spawn a physical consolidation Direction       (Phase 2)
    └──► Dissolve  (e.g. capabilities reclassified)

Direction
    ├──► Advance status (draft → proposed → agreed)
    ├──► Adjust placements / horizon
    └──► Reject
```

## Group decision semantics

`status` values across both Direction and Standard App designation:

- `draft` — single architect's proposal, group has not yet discussed
- `proposed` — formally on the group's agenda, under discussion
- `agreed` — group has reached consensus
- `rejected` — group has decided against (terminal; preserved for audit)

Status transitions are events. Comments / decision narrative attach at each transition.

## Out of scope

- Project / change-management tracking (this is a strategic communication tool, not a project tool)
- Realization tracking — when a Direction is `agreed`, how do we know when it has been executed in reality?
- Cross-tenant or organisation-wide governance
- Decomposition entry-point UX — decomposition starts from a single physical capability, not from a discovered theme. Different flow, separate spec.

## Feature specs that follow this

Each becomes a self-contained spec under `/specs/`:

1. Rename `EnterpriseCapability` → `LogicalCapability` (data migration only, no behaviour change)
2. Add `Direction` aggregate (type, placements, horizon, status, narrative, source capabilities)
3. Add `StandardAppDesignation`
4. Discover view — consolidation candidates with two-path promotion
5. Direction map — physical movement visualisation
6. Logical capability map — the logical layer, with standard apps and physical mappings
7. Target Architecture synthesis view — physical + logical + roadmap horizons composed
