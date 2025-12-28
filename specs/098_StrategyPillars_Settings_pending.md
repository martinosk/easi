# Customize Strategy Pillars

**Status**: pending

## User Value

> "As an enterprise architect, I want to define my organization's strategic pillars (e.g., 'Digital First', 'Cost Optimization', 'Innovation') instead of using hardcoded values, so capabilities can be aligned to our actual strategic initiatives."

## Dependencies

- Spec 090: MetaModel Bounded Context

---

## Domain Model

### StrategyPillarConfiguration Aggregate

Single aggregate per tenant managing all pillar definitions.

**Pillar Definition** (value object):
- Name (required, unique per tenant, max 100 chars)
- Description (optional, max 500 chars)
- Display order (for UI sorting)
- Active flag (for soft delete)

**Commands**:
- Add pillar
- Update pillar (name, description, order)
- Remove pillar (soft delete)

**Business Rules**:
- Pillar names must be unique within tenant (case-insensitive)
- Maximum 20 pillars per tenant
- At least 1 active pillar must remain
- Soft delete preserves pillar for historical importance ratings

**Initialization**:
When a tenant is created, initialize with default pillars (Always On, Grow, Transform) for backward compatibility.

---

## User Experience

### Entry Point
Settings → Strategy Pillars (new tab alongside Maturity Scale)

### Main View

```
┌─────────────────────────────────────────────────────────────┐
│  Settings                                                   │
├─────────────────────────────────────────────────────────────┤
│  [Maturity Scale] [Strategy Pillars]                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Strategy Pillars                              [Edit]       │
│                                                             │
│  Define the strategic pillars used to categorize            │
│  capabilities across your organization.                     │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ 1. Always On                                        │    │
│  │    Core capabilities that must always be operational│    │
│  ├─────────────────────────────────────────────────────┤    │
│  │ 2. Grow                                             │    │
│  │    Capabilities driving business growth             │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │ 3. Transform                                        │    │
│  │    Capabilities enabling digital transformation     │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Edit Mode

```
┌─────────────────────────────────────────────────────────────┐
│  Strategy Pillars                        [Cancel] [Save]    │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ 1. [Always On_____________]                 [🗑]    │    │
│  │    [Core capabilities that...]                      │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │ 2. [Grow__________________]                 [🗑]    │    │
│  │    [Capabilities driving...]                        │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │ 3. [Transform_____________]                 [🗑]    │    │
│  │    [Capabilities enabling...]                       │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  [+ Add Pillar]                                             │
│                                                             │
│  Maximum 20 pillars allowed.                                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Interactions
- Edit pillar name and description inline
- Delete pillar (disabled if only 1 remains)
- Add new pillar at bottom
- Save persists all changes
- Cancel discards changes

### Validation Feedback
- Name required, must be unique
- Show error if max pillars reached

---

## API Requirements

**Operations needed**:
- List all pillars for tenant (include inactive optionally)
- Get single pillar
- Create pillar
- Update pillar (with optimistic locking)
- Delete pillar (soft delete)

**Permissions**:
- Read: `PermMetaModelRead`
- Write: `PermMetaModelWrite`

---

## Checklist

- [x] Specification approved
- [x] Domain model implemented
- [x] API implemented
- [x] Settings UI implemented
- [x] Tests passing
- [ ] User sign-off
