# Strategic Roadmap

The plan-of-record for EASI's capability and bounded-context evolution: settled strategic decisions, the standing invariants every design is checked against, and the horizon plan with spec traceability. Settled 2026-08-31 against the [capability coverage assessment](https://claude.ai/code/artifact/923d1471-69c5-4023-9e58-9f8d0aba036f); the [roadmap artifact](https://claude.ai/code/artifact/2d6d9e21-9b4b-4280-9674-f57a83209eaf) is the stakeholder view of this register. Codes like C2, G1, F5 reference findings in the coverage assessment.

## Decision Register

| ID | Decision | Addresses |
|----|----------|-----------|
| SD1 | EnterpriseArchitecture dissolves into Architecture Direction. EnterpriseCapability is demoted from curated catalog to the subject of a Direction — it comes into existence when direction work starts. TIME suggestion and assessment unify on one value object. | C2, C3 |
| SD2 | TargetMaturity and EnterpriseStrategicImportance move into Architecture Direction as a separate Assessment aggregate with its own events; Direction consumes, never embeds. The maturity concept stays reshapeable or removable without touching Direction. | A4 |
| SD3 | CMDB sync (ServiceNow) auto-registers discovered applications in ownership state Unknown/orphaned, stamped with provenance; stewards promote them. Sync never overwrites a curated field. | G2 |
| SD4 | Value Streams evolves into the Business Operating Model context: BusinessRole and BusinessObject aggregates, linked to capabilities through a reified but hidden, unnamed process aggregate. Capability Mapping does not grow. | A3, B4 |
| SD5 | MetaModel owns the custom-field schema (moved from OnePagers) and the modelling rule set: capability-naming standards (AI-evaluated, override recorded), composition depth, required attributes. OnePagers trends toward pure presentation. | G1, D1 |
| SD6 | ApplicationComponent carries an ownership state machine (Unknown/orphaned → Nominated → Owned or Managed; owner is an identity or team reference, never free text), a hosting classification (on-prem, cloud, SaaS, third-party hosted, unknown), and composition capped at two levels, exposed as HATEOAS affordances. | B1, A2, B3 |
| SD7 | Importing evolves into the Integrations anti-corruption layer: external identity mapping, sync runs, per-attribute reconciliation, provenance on every synced fact. All writes go through published commands. ServiceNow is the first adapter, MS Forms the second. | G2 |
| SD8 | Every new analysis surface is a pure read-side context over published events — stewardship, dashboards, export, cross-context impact. Adding an analysis costs a projector, never a contract. | C4, C6, F5, G3 |

## Standing Invariants

Every spec and every architectural recommendation is checked against these. A design that breaks one is a proposed amendment to this register, decided by a human.

1. **Events-only integration** — contexts integrate through published events and published commands; nothing else crosses a boundary (spec 209, test-enforced).
2. **Analysis is a read side** — a new analysis surface is a projection over published events, never a new dependency between write-side contexts.
3. **External data carries provenance** — facts enter only through the Integrations context, stamped with source, external id and sync time; sync never overwrites a curated fact.
4. **Judgements are separable** — assessment-like concepts get their own aggregate and events; consumers reference, never embed.
5. **MetaModel owns the vocabulary** — scales, pillars, custom attributes and modelling rules; AI checks are advisory at the edge with recorded overrides, and the domain never blocks on a model.
6. **Contexts stay small** — when one context starts speaking two languages, split it. Architecture Direction's split line after absorbing EnterpriseArchitecture is judgement (assess, decide) versus plan (journeys, tracking).

## Horizon Plan & Traceability

A horizon move becomes a Phase-1 design doc (`docs/specs/`) and vertical-slice specs per `easi-spec-driven-development`. When a spec advancing a move lands, its row is updated here.

### Horizon 1 — pay the structural debts (committed)

| Move | Scope | Decisions | Specs | Status |
|------|-------|-----------|-------|--------|
| H1-1 | Dissolve EnterpriseArchitecture into Architecture Direction: unified TIME value object, EnterpriseCapability as Direction subject, Assessment aggregate, retire the context | SD1, SD2 | — | not started |
| H1-2 | Application record: ownership state machine with stats projection, hosting classification, two-level composition | SD6 | — | not started |
| H1-3 | MetaModel: custom-field schema moves in (backfilled), modelling rule set v1 with AI naming validation | SD5 | — | not started |

Exit: coverage assessment re-scored — boundary smells 3 → 1, contexts 15 → 14.

### Horizon 2 — open the model

| Move | Scope | Decisions | Specs | Status |
|------|-------|-----------|-------|--------|
| H2-4 | Business Operating Model: BusinessRole, BusinessObject, hidden process, capability links | SD4 | — | not started |
| H2-5 | Integrations context and ServiceNow adapter: identity map, sync runs, reconciliation, provenance | SD7, SD3 | — | not started |
| H2-6 | Stewardship read side: orphaned, stale and incomplete items ranked, routed via invite-to-edit | SD8 | — | not started |

### Horizon 3 — depth on demand (pulled by usage, not scheduled)

| Move | Scope | Decisions | Specs | Status |
|------|-------|-----------|-------|--------|
| H3-7 | MS Forms adapter feeding one-pager facts | SD7 | — | not started |
| H3-8 | Export read side (CSV, image, BI feed) | SD8 | — | not started |
| H3-9 | Dashboards & KPIs read side | SD8 | — | not started |
| H3-10 | Maturity 2.0: MetaModel-defined dimensions, per-dimension ratings, evidence-assisted suggestions | SD2 | — | not started |
| H3-11 | Principles register grown from the modelling rule set | SD5 | — | not started |

## Parked

Parking is a decision. Each capability waits for its pull condition; the trailhead makes it cheap to start.

| Code | Capability | Trailhead | Pull condition |
|------|-----------|-----------|----------------|
| B3 | Technology portfolio | Hosting classification (H1-2) | Someone needs versions and end-of-life tracking |
| B4 | Full data architecture | BusinessObject (H2-4) | A real steward of systems-of-record questions |
| B6 | Cost & investment transparency | — | A finance-side source of truth to integrate with |
| E3 | Initiative / project linkage | — | A delivery-side system to link to |
| E4 | Scenario & what-if analysis | Single-target Direction model | Consolidation decisions start being contested |
| B1† | Attribute values out of OnePagers | Schema in MetaModel (H1-3) | Integrations needs first-class attributes other contexts can see |

## Maintenance

- A spec advancing a move records it in its `Roadmap alignment` field and updates the move's row here when done.
- At each horizon exit, the coverage assessment is re-scored and both artifacts are refreshed.
- Amendments to decisions or invariants happen in this file by PR, dated, never implicitly in a spec.
