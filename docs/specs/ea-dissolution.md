# Design: Dissolving EnterpriseArchitecture into Architecture Direction

Phase-1 design document for roadmap move H1-1 (decisions SD1, SD2 in `docs/architecture/ROADMAP.md`). Slices: specs 210–213.

## Problem Statement

EnterpriseArchitecture fails the ubiquitous-language test: an EnterpriseCapability cannot be discussed without Architecture Direction's vocabulary (directions, sources, standard applications, composition). Spec 172 already made the Direction the only association between domain capabilities and an EC, leaving the EC aggregate a shell (name, description, category, target maturity, active). Spec 207 moved every analysis over ECs into AD. What remains in EA is two aggregates and a TIME-suggestion calculator whose four output values duplicate AD's TimeGrade vocabulary (coverage smell C2), plus a strategic-importance rating that has **no frontend UI at all** — backend and agent tools only.

The context must dissolve into AD; the enterprise-capability concept demotes from curated catalog to the subject of a Direction; the ratings become a decoupled Assessment concept so the maturity idea can be reshaped or removed without touching Direction (SD2).

## Research Summary

- **Precedent** — the Platform→Auth merge (f93e65fa): the event store is one global `events` table, so stored events never move; read-model tables re-parent via `ALTER TABLE … SET SCHEMA`; published event type strings stay byte-identical so consumers only swap which package's constant they subscribe to (one line each); the architecture guard tests discover contexts dynamically and needed zero edits.
- **EA surface** — aggregates `EnterpriseCapability` and `EnterpriseStrategicImportance` (composite id per EC+pillar, importance 1–5 with rationale); seven published events (`EnterpriseCapabilityCreated/Updated/Deleted/TargetMaturitySet`, `EnterpriseStrategicImportanceSet/Updated/Removed`); routes under `/enterprise-capabilities` and `/time-suggestions`; permission group `enterprise-arch:*`; eight agent tools.
- **Consumers** — AD (`enterprise_capability_cache` projector + a reactor rejecting the active Direction on EC deletion) and OnePagers (subject index for subject type `enterprise-capability`, `targetMaturity` attribute, facts archival on deletion). Nothing else.
- **TIME suggestion** — `TimeSuggestionCalculator` over four EA-local caches (`ea_realization_cache`, `domain_capability_metadata`, `ea_importance_cache`, `ea_fit_score_cache`) plus pillar fit types; gap threshold 1.5; outputs `TimeClassification` (Tolerate/Invest/Migrate/Eliminate) + confidence. AD's `TimeGrade` holds the identical four values.
- **Cache overlap** — post-merge, `ea_realization_cache` duplicates AD's `realization_cache` and `domain_capability_metadata` overlaps AD's `capability_node_cache` (name, domain, maturity value); `ea_importance_cache`, `ea_fit_score_cache`, `ea_strategy_pillar_cache` are unique.
- **Frontend** — one page (`EnterpriseArchPage`, four tabs: capabilities, maturity-analysis, strategic-fit, time-suggestions); the EC detail panel already embeds AD's `DirectionPanel`; target maturity is edited from the maturity-analysis tab; strategic importance has no UI caller.
- **Spec 169 (pending)** — its "resolve candidate as enterprise capability" path assumes EC is a standalone catalog entity, contradicting the spec-172 model; it needs re-scoping.

## Ubiquitous Language

| Term | Meaning after this design |
|------|---------------------------|
| Enterprise Capability | The subject a Direction is about. Comes into existence when direction work starts; not a curated catalog entry. |
| Assessment | AD's judgement record about an enterprise capability: target maturity and per-pillar strategic importance, with rationale. Own aggregate, own events; Direction consumes, never embeds. |
| TIME grade | The single TIME vocabulary (Invest, Tolerate, Migrate, Eliminate) used both by the recorded assessment and the computed suggestion. |
| TIME suggestion | A computed, confidence-qualified advisory TIME grade per realisation, derived from fit gaps. Advice; never a recorded judgement. |

## Proposed Approach

Relocate first, remodel second. Slice 210 moves the whole context into AD mechanically — packages, routes, tables via `SET SCHEMA`, published constants relocated with unchanged strings — following the Platform→Auth playbook, and drops AD's now-redundant `enterprise_capability_cache` (the underlying table becomes AD-local). The EC-deleted reaction becomes an in-context reaction. OnePagers swaps seven import lines. Contexts go 15 → 14 here, with zero behaviour change.

The remaining slices are AD-internal remodels: 211 carves target maturity and strategic importance out of the relocated aggregates into the Assessment aggregate (SD2), keeping published event strings unchanged so OnePagers is untouched; 212 unifies TIME on one value object, moves the suggestion calculator beside TimeAssessment, surfaces the suggestion in the assessment flow, and consolidates the duplicate caches; 213 demotes EC creation into the direction-capture flow and retires the standalone catalog affordances.

## Key Decisions

1. **Relocate wholesale, then remodel** — the precedent shows a mechanical move is one safe, atomic change; thinning EA first would churn cross-context contracts twice. Alternative (carve concepts across the boundary one by one) rejected.
2. **Published event type strings never change** — `EnterpriseCapabilityCreated` etc. keep their bytes; consumers swap import constants only. The wire and the stored history stay stable through all four slices.
3. **Assessment is one aggregate per enterprise capability** — holding target maturity and the per-pillar importance entries. Seeded by a one-time routine dispatching AD commands from current read-model state (the mechanism Importing proved); no synthetic events written directly to the store. Direction never references Assessment; read models join where a surface needs both.
4. **Strategic importance survives as data, not as UI** — it moves into Assessment (agent tools re-homed) but no UI is added; whether it earns a surface is a later product decision the decoupling explicitly enables (SD2).
5. **One TIME value object** — AD's `TimeGrade` becomes the single type; the suggestion calculator emits it with its confidence. `TimeClassification` is deleted. Cache consolidation follows: suggestion queries move onto `realization_cache`/`capability_node_cache` plus the two unique fit/importance caches, which re-parent into AD.
6. **EC creation happens only through direction capture** — drafting a direction either selects an existing subject or names a new one inline; standalone EC create/delete endpoints and agent tools retire. Existing subject-less ECs remain until directed or deleted.
7. **Permission strings stay `enterprise-arch:*`** — renaming a permission group is user-visible churn with no modelling value; not part of this move.
8. **Spec 169 must be re-scoped** — its EC-linking resolution path contradicts the direction-is-the-association model; re-scope to Direction-only resolution (or require an agreed Direction on the target EC) when 169 is picked up.

## Slice Map

| Spec | Slice | Kind |
|------|-------|------|
| 210 | Relocate EnterpriseArchitecture into Architecture Direction | Mechanical merge, zero behaviour change; contexts 15 → 14 |
| 211 | Assessment aggregate (target maturity + strategic importance) | AD-internal remodel; OnePagers untouched via stable event strings |
| 212 | One TIME vocabulary and consolidated caches | AD-internal remodel; resolves smell C2 |
| 213 | Enterprise capability as Direction subject | Product-visible demotion; retires catalog affordances |
