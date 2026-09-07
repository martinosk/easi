# 218 — Modelling Rule Set: Capability Naming Validation

> **Status:** pending
> **Depends on:** [217_CustomFieldSchemaInMetaModel](217_CustomFieldSchemaInMetaModel_pending.md)
> **Roadmap alignment:** SD5 / H1-3

---

## Problem Statement

Capability names carry the shared language of the whole model, yet the only validation today is "not empty". Weak names (verb phrases, application names, org-unit names) degrade the map for everyone. The roadmap introduces a modelling rule set owned by MetaModel; its first rule is a capability naming standard, evaluated by AI as advice the modeller can override — the domain never blocks on a language model.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Admin / metamodel steward** | Configure the naming rule and its guidance text |
| **Enterprise Architect** | Get advice on weak names without being blocked by it |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: AI-advised capability naming with override

  Scenario: A conforming name passes silently
    Given the naming rule is enabled
    When an architect creates capability "Demand Forecasting"
    Then the capability is created without any advisory interruption

  Scenario: A weak name triggers advice
    Given the naming rule is enabled
    When an architect submits capability name "Do SAP Stuff"
    Then an advisory explains why the name is weak and suggests alternatives
    And the architect can revise the name or proceed anyway

  Scenario: Overriding the advice is recorded
    Given an architect received a negative advisory for "Do SAP Stuff"
    When they proceed with the name unchanged
    Then the capability is created
    And the override is visible in the capability's audit trail

  Scenario: Rule disabled means no advice
    Given the naming rule is disabled
    When an architect creates a capability with any non-empty name
    Then no advisory is requested or shown

  Scenario: Advisory unavailable never blocks
    Given the naming rule is enabled and the AI channel fails
    When an architect submits a capability name
    Then the capability is created and the advisory is skipped
```

---

## Business Rules & Invariants

1. **MetaModel owns the rule** — a per-tenant modelling rule set; v1 contains one rule: capability naming (enabled flag + guidance text steering the evaluation).
2. **Advisory, never blocking** — evaluation happens at the edge before submission; capability commands never depend on an AI result, and AI-channel failure degrades to no advice.
3. **Override is a domain fact** — proceeding against a negative advisory raises `CapabilityNamingGuidanceOverridden` on the capability (advisory summary, actor), visible through the generic audit trail.
4. **Arch Assistant is the AI channel** — evaluation uses the tenant's configured AI provider and the MetaModel guidance, consumed from Arch Assistant's local cache of MetaModel rule events.
5. **Rename included** — the same advisory applies to renaming an existing capability.

---

## Acceptance Criteria

- [ ] Stewards configure the naming rule (enable/disable, guidance text) in MetaModel; changes are published events
- [ ] Create and rename flows in Capability Mapping request an advisory exactly when the rule is enabled
- [ ] A negative advisory offers revise and proceed; proceeding records the override event
- [ ] AI-channel failure or timeout never delays or blocks the command path
- [ ] Tenants without an AI configuration behave as if the rule is disabled

---

## Architecture

### Ownership

MetaModel owns the rule configuration. Arch Assistant owns the evaluation endpoint. Capability Mapping owns the override fact on its capability aggregate.

### Domain Model

MetaModel: a small per-tenant modelling-rules aggregate; published event `CapabilityNamingRuleChanged` (enabled, guidance). Capability Mapping: `CapabilityNamingGuidanceOverridden` event raised alongside create/rename when the command carries an override acknowledgment; no aggregate state change. Arch Assistant: stateless evaluation using the tenant AI configuration; subscribes to the MetaModel rule event into a local cache (the TenantCreated-subscription pattern).

### API Surface

MetaModel: rule read/write under `meta-model:read`/`meta-model:write`. Arch Assistant: a name-review operation (capability name, level, parent context in; verdict, reasoning, suggestions out) under the caller's `capabilities:write`. Capability create/update commands accept an optional override acknowledgment carrying the advisory summary.

### Persistence

MetaModel rule state and Arch Assistant's rule cache; no new Capability Mapping tables — the override lives in the event stream and surfaces through Audit.

### Frontend

Capability create/edit dialogs: on submit with the rule enabled, request the advisory; render verdict with suggestions; "Use anyway" proceeds with the acknowledgment attached. Rule configuration UI beside the existing MetaModel settings.

### Cross-Context Integration

MetaModel publishes the rule event; Arch Assistant consumes it. Capability Mapping is untouched by MetaModel — the advisory result travels through the client, keeping the command path AI-free.

---

## Design Decisions

1. **Advisory at the edge, override as an event** — satisfies the standing invariant that the domain never blocks on a model while still making the override auditable. Alternatives: validation inside `CapabilityName` (rejected: blocks the domain on an LLM), server-side enforcement gate in the handler (rejected: same problem plus latency on every create).
2. **Arch Assistant hosts the evaluation, not MetaModel** — MetaModel owns vocabulary, not AI plumbing; Arch Assistant already holds per-tenant provider configuration, budgets, and permission ceilings. 
3. **Advisory result travels through the client** — avoids a CM→AA runtime dependency; the acknowledgment on the command is data, and a fabricated acknowledgment is equivalent to an override, which is permitted anyway.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Client-carried advisory | The override event trusts the client's summary text | The event records actor and timestamp; the advisory is advice, not access control |
| Per-submit LLM call | Latency and cost on capability creation | Called only when the rule is enabled; failure degrades silently to no advice |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
