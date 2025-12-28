# Rate Domain Capability Strategic Importance

**Status**: pending

## User Value

> "As a domain architect, I want to rate how important each capability is for our strategic pillars, so I can communicate strategic priorities and guide investment decisions."

### Key Insight

Strategic importance is **domain-specific**. The same capability can have different importance in different domains:
- Digital Banking rates "Customer Data Management" as Critical (5) for Transform
- Traditional Lending rates the same capability as Average (3) for Transform

This enables federated governance where each domain manages its own strategic priorities.

## Dependencies

- Spec 098: Strategy Pillars Settings
- Spec 053: Business Domain Aggregate
- Spec 023: Capability Model

---

## Domain Model

### DomainCapabilityStrategyImportance Aggregate

Manages importance ratings for a capability within a specific business domain context.

**Properties**:
- Business Domain reference
- Capability reference
- Pillar reference
- Importance (1-5 scale)
- Rationale (optional, max 500 chars) - explains "why" this rating

**Commands**:
- Set importance (creates rating with optional rationale)
- Update importance (change rating or rationale)
- Remove importance (deletes rating)

**Business Rules**:
- Domain, capability, and pillar must all exist
- Pillar must be active
- One rating per (domain + capability + pillar) combination
- Importance must be 1-5

**Cascade Behavior**:
- When capability deleted → remove all its importance ratings
- When business domain deleted → remove all ratings for that domain
- When pillar soft-deleted → keep existing ratings but prevent new ones

### Importance Scale

| Value | Label | Meaning |
|-------|-------|---------|
| 1 | Low | Nice to have |
| 2 | Below Average | Minor support for strategy |
| 3 | Average | Supports strategy |
| 4 | Above Average | Important for strategy |
| 5 | Critical | Essential for strategy success |

---

## User Experience

### Entry Point: Capability Details Panel

Show importance ratings within the domain context:

```
┌─────────────────────────────────────────────────────────────┐
│  Customer Onboarding                            [Edit] [🗑] │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Level: L1                                                  │
│  Maturity: Product (65)                                     │
│                                                             │
│  ─────────────────────────────────────────────────────────  │
│  Strategic Importance (Digital Banking)                     │
│  ─────────────────────────────────────────────────────────  │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ Always On          ★★★★★ Critical           [Edit]  │    │
│  │ Core revenue stream, any downtime causes losses     │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │ Transform          ★★★★☆ Above Average      [Edit]  │    │
│  │ Key enabler for digital-first strategy              │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  [+ Rate Another Pillar]                                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Set/Edit Importance Dialog

```
┌─────────────────────────────────────────────────────────────┐
│  Set Strategic Importance                            [X]    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Capability: Customer Onboarding                            │
│  Domain: Digital Banking                                    │
│                                                             │
│  Strategy Pillar                                            │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  ○ Always On                                        │    │
│  │  ○ Grow                                             │    │
│  │  ● Transform                                        │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  How important is this capability for "Transform"?          │
│                                                             │
│     1      2      3      4      5                           │
│     ○      ○      ○      ●      ○                           │
│     Low         Average       Critical                      │
│                                                             │
│  Why? (optional)                                            │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ Key enabler for our digital-first customer          │    │
│  │ acquisition strategy.                               │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│  [Remove]                          [Cancel]  [Save]         │
└─────────────────────────────────────────────────────────────┘
```

### Business Domain Capability List

Show importance at a glance with filtering:

```
┌─────────────────────────────────────────────────────────────┐
│  Digital Banking                                            │
├─────────────────────────────────────────────────────────────┤
│  Capabilities                                               │
│                                                             │
│  Filter: [All Pillars ▼]  [Critical only ☐]                 │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ ▼ Customer Onboarding         [Product] [★★★★★]    │    │
│  │   ├─ Identity Verification    [Custom]  [★★★★☆]    │    │
│  │   └─ Document Processing      [Genesis] [★★★☆☆]    │    │
│  │ ▼ Payment Processing          [Product] [★★★★★]    │    │
│  │ ▶ Risk Assessment             [Custom]  [★★☆☆☆]    │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## API Requirements

**Operations needed**:
- List importance ratings for a capability in a domain
- Set importance (create rating)
- Update importance
- Remove importance
- List all ratings for a domain (portfolio view)
- List all ratings for a capability across domains (cross-domain comparison)

**Anti-Corruption Layer**:
Capability Mapping context maintains a local cache of available strategy pillars, synced via events from MetaModel context.

**Permissions**:
- Read: `PermCapabilityRead`
- Write: `PermCapabilityWrite`

---

## Checklist

- [ ] Specification approved
- [ ] Domain model implemented
- [ ] Anti-corruption layer for pillars
- [ ] API implemented
- [ ] Capability details UI updated
- [ ] Business domain filtering
- [ ] Tests passing
- [ ] User sign-off
