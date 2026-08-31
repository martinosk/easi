# Design: Retiring the Enterprise Capability

> Status: approved 2026-08-31; slice 210 shipped, slices 211–213 ready to implement
> Author: agent + maosk
> Date: 2026-08-31
> Amended 2026-08-31 (user decision): the enterprise capability is **retired outright**, not demoted to the subject of a Direction as this document first proposed. Slice 210 shipped under the original reading and stands unchanged — see D1. Slices 211–213 are re-planned; D3, D4, D5 and D8 record what replaced them.
> Phase-1 design document for roadmap move H1-1 (decisions SD1, SD2 in [`../architecture/ROADMAP.md`](../architecture/ROADMAP.md)).

## Problem Statement

The enterprise capability was introduced as a cross-domain strategic grouping — a curated theme like "Digital Customer Engagement" that domain capabilities roll up into. It never earned its keep. Nobody curates the taxonomy. Spec 172 already reduced the aggregate to a shell (name, description, category, target maturity, active) whose only association with real capabilities runs through a Direction. The strategic-importance rating has no UI at all — backend and agent tools only.

This document originally proposed saving the concept by demoting it to "the subject a Direction is about". That preserved the wrong thing: the concept is not mis-scoped, it is unwanted. What remains is to remove it without losing the analysis that happened to be displayed alongside it.

What makes retirement affordable is that the enterprise architecture page's four tabs are of very unequal worth, and only one of them actually depends on the enterprise capability:

| Tab | Depends on the enterprise capability? | Verdict |
|-----|---------------------------------------|---------|
| Capabilities (the catalog) | Is the enterprise capability | Retire |
| Strategic fit | No — `GET /strategic-fit-analysis/{pillarId}` is owned by Capability Mapping and scores domain capabilities against pillars | Keep, re-home the UI |
| TIME suggestions | No — computed per `(capability, application)` realisation from fit gaps | Keep, re-home beside the assessment |
| Maturity analysis | Yes — the enterprise capability holds the target maturity | Rework |

So the whole cost of retiring the enterprise capability is finding maturity analysis a home. And there is a natural one. Maturing a capability is a planned change, pursued over a period, in steps — which is exactly what a journey already is. Critically, maturing a capability is **not** necessarily a technical change: it can come from process, ownership, data quality, skills or governance work. A journey's milestones already carry arbitrary steps, and the `move` kind already proves a journey need not be application-shaped.

## Research Summary

- **Precedent for the relocation (slice 210)** — the Platform→Auth merge (f93e65fa): the event store is one global `events` table, so stored events never move; read-model tables re-parent via `ALTER TABLE … SET SCHEMA`; published event type strings stay byte-identical so consumers only swap which package's constant they subscribe to; the architecture guard tests discover contexts dynamically and needed zero edits.
- **Enterprise capability surface** — aggregates `EnterpriseCapability` and `EnterpriseStrategicImportance` (composite id per EC+pillar, importance 1–5 with rationale); seven published events; routes under `/enterprise-capabilities` and `/time-suggestions`; permission group `enterprise-arch:*`; eight agent tools; one frontend feature (`enterprise-architecture`) with a four-tab page.
- **Direction, Standard Application and Composition are enterprise-capability-shaped** — a Direction's subject is an enterprise capability, a Standard Application is set per enterprise capability, and Composition exists only to compose one from source capabilities. `DirectionPanel` and `StandardApplicationPanel` render **only** inside `EnterpriseCapabilityDetailPanel`. Remove the enterprise capability and all three lose both their subject and their only surface.
- **Journey kinds already tolerate a non-application journey** — `migration`, `consolidation` and `carve-out` require from-applications and a to-application; `move` requires none of them and instead carries `targetDomain`, `targetParent` and `resultingName`. A journey kind may define its own target fields.
- **Journeys attach to one domain capability** (spec 182, D1), have status, target period, progress and ordered milestones, and compose across the capability hierarchy (spec 195).
- **Maturity analysis today** — for each enterprise capability, compare its target maturity against the current maturity (`capability_node_cache` / `domain_capability_metadata`, 0–99 over Genesis / Custom Build / Product / Commodity) of every domain capability its direction composes; gap drives an investment priority.
- **TIME suggestion today** — `TimeSuggestionCalculator` over fit gaps, emitting `TimeClassification` + confidence per `(capability, component)`. `TimeAssessment` in Architecture Direction is keyed identically and holds `TimeGrade`, the same four values under a different Go type (coverage smell C2).
- **Cache overlap** — post-relocation, `ea_realization_cache` duplicates Architecture Direction's `realization_cache` and `domain_capability_metadata` overlaps `capability_node_cache`; `ea_importance_cache`, `ea_fit_score_cache` and `ea_strategy_pillar_cache` are unique.
- **OnePagers** carries `enterprise-capability` as one of six subject types, across the subject-type value object, the relation catalog, the built-in field catalog, the subject index projector and the subject-deleted reactor. It is the only cross-context consumer of the retiring events.
- **Architecture Direction after retirement** holds `CapabilityJourney`, `TimeAssessment` and `RealizationRole` — all keyed on a domain capability or a capability-application pair. One language: planning and tracking capability change.

## Ubiquitous Language

| Term | Meaning after this design |
|------|---------------------------|
| Capability | The domain capability. Unchanged, and now the only thing plans and assessments hang off. |
| Journey | A planned change to one capability over time, with a status, a target period and ordered milestones. Kinds: migration, consolidation, carve-out, move, and now maturity. |
| Maturity journey | A journey whose declared outcome is a maturity level rather than an application change. Its milestones carry the work, which need not be technical. |
| Maturity gap | Current maturity against the maturity a journey targets. A gap exists because someone planned to close it. |
| TIME grade | The single TIME vocabulary (Invest, Tolerate, Migrate, Eliminate), recorded per realisation. |
| TIME suggestion | A computed, confidence-qualified advisory TIME grade per realisation, derived from fit gaps. Advice; never a recorded judgement. |
| Strategic fit | Capability importance against application fit per strategy pillar. Owned by Capability Mapping, untouched by this move. |
| ~~Enterprise capability~~ | Retired. |
| ~~Direction~~, ~~Standard application~~, ~~Composition~~ | Retired with it — each was a statement about an enterprise capability. |

## Proposed Approach

Relocate first, replace second, remove last — four slices.

Slice 210 (shipped) moved the whole EnterpriseArchitecture context into Architecture Direction mechanically — packages, routes, tables via `SET SCHEMA`, published constants relocated with unchanged strings — following the Platform→Auth playbook. Contexts went 15 → 14 with zero behaviour change. Its value to the retirement is that every enterprise-capability concern now sits inside one context, making the removal a single-context deletion rather than a cross-boundary unpicking.

Slice 211 adds the maturity journey: a new journey kind carrying the maturity it will deliver, so a capability's maturity ambition is expressed as a plan with steps and a period rather than a number on a catalog entry. Slice 212 unifies the two TIME types and moves the suggestion beside the assessment it advises, so the suggestion survives the page that used to host it. Slice 213 then deletes the enterprise capability and everything hanging off it, and gives strategic fit analysis a page of its own.

Only after 211 and 212 have shipped does 213 remove anything a user relies on.

## Key Decisions

| ID | Decision | Rationale |
|----|----------|-----------|
| D1 | **Relocate wholesale, then remodel** — applied in spec 210 | The Platform→Auth precedent shows a mechanical move is one safe, atomic change; thinning the context first would churn cross-context contracts twice. Alternative (carve concepts across the boundary one by one) rejected. Unaffected by the amendment: the relocation is what makes the retirement single-context. |
| D2 | **Published event strings were unchanged through the relocation** — applied in spec 210 | `EnterpriseCapabilityCreated` etc. kept their bytes; consumers swapped import constants only. The wire and the stored history stayed stable across the move. Superseded for the retirement by D9. |
| D3 | **The enterprise capability is retired outright** (user decision 2026-08-31) | Nobody curates the taxonomy, spec 172 already hollowed the aggregate, and the strategic-importance rating never had a UI. Supersedes this document's original proposal to demote it to the subject of a Direction — that reading kept the concept alive to serve concepts that themselves only existed to serve it. |
| D4 | **Direction, Standard Application and Composition retire with it** (user decision 2026-08-31) | Each is a statement *about* an enterprise capability, and none has a surface outside the enterprise capability detail panel. Alternative — re-home Direction on the domain capability — rejected: a Direction and a journey would then both describe intended change on the same capability, which is the ubiquitous-language failure this move exists to end. |
| D5 | **Maturity is a journey kind, not a rating** (user decision 2026-08-31) | A target maturity on a catalog entry is a number nobody revisits; a journey ties the ambition to a period, milestones and a progress state, and is already the artifact architects work in. Alternative — target maturity as capability metadata beside current maturity — rejected: it preserves catalog-wide gap analysis but re-creates the inert number; investment candidates are better found through TIME suggestions and fit gaps, which are computed from evidence rather than typed in. |
| D6 | **Maturity journeys carry no applications** | The uplift may be process, ownership, data quality or skills work, not a system change. The kind requires zero from-applications and no to-application, following `move`'s precedent, and validates its own target field instead. |
| D7 | **One TIME value object, with the suggestion beside the assessment** | `TimeGrade` survives as the only TIME type — it is the recorded-judgement type with UI and constraints attached — and `TimeClassification` is deleted. The suggestion is composed into assessment reads at query time, never stored: advice goes stale whenever fit scores move. A separate suggestions page is what kept advice away from the decision. |
| D8 | **Strategic fit analysis is untouched behind the API** (user decision 2026-08-31) | It is already owned by Capability Mapping, scores domain capabilities against pillars, and never referenced an enterprise capability. Only its UI home moves — to a page of its own, not into a capability drawer, which would lose the cross-capability comparison that makes it useful. |
| D9 | **Retired event strings die rather than change; stored events stay** | The retirement removes the concepts, so the published constants are deleted along with OnePagers' handling of them. Stored events remain in `infrastructure.events`, unread — event stores are append-only, replaying retired streams is nobody's job once the deserializers are gone, and a history-rewriting migration is risk without benefit. |
| D10 | **Spec 169 must be re-scoped** | Discovery Candidates resolve as `resolved-as-enterprise-capability` or `resolved-as-direction`; both vocabularies disappear here. Re-scope its resolution paths (candidate → capability, or candidate → journey) when 169 is picked up. |

## Slice Map

| Spec | Slice | Kind |
|------|-------|------|
| 210 | Relocate EnterpriseArchitecture into Architecture Direction | Shipped. Mechanical merge, zero behaviour change; contexts 15 → 14 |
| 211 | Maturity journeys | Additive: a new journey kind expressing maturity uplift (D5, D6) |
| 212 | One TIME vocabulary, suggestion beside the assessment | AD-internal remodel; frees the suggestion from the retiring page (D7) |
| 213 | Retire the enterprise capability | Product-visible removal; strategic fit gets its own page (D3, D4, D8, D9) |
