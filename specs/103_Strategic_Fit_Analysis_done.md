# Strategic Fit Analysis

**Status**: ongoing

## User Value

> "As an enterprise architect, I want to score how well each application realization fits the strategic pillars that matter for that capability, so I can identify where our IT landscape is misaligned with our strategy."

> "As a portfolio manager, I want to see which applications are strategic liabilities - realizing important capabilities but with poor fit for our strategic goals."

## Dependencies

- Spec 098: Strategy Pillars
- Spec 100: Enterprise Capability Groupings
- Capability Realization Links (existing)

---

## Domain Concepts

### Pillar Fit Configuration

Architects configure which strategy pillars should have fit scoring enabled. Not all pillars require fit scoring - only those where the technical fitness of realizations matters.

Example:
- "Always On" pillar → Enable fit scoring (uptime, reliability matters)
- "Grow" pillar → Enable fit scoring (scalability, flexibility matters)
- "Transform" pillar → May not need fit scoring (it's about change, not current state)

### Application Fit Score

A score (1-5 scale) on an application (component) indicating how well that application fits a particular strategy pillar. This represents the application's inherent qualities.

Key insight: The fit score is at the **application level**, not the realization level:
- App X has "Always On" fit: 2 (poor reliability, no DR)
- App X has "Grow" fit: 4 (good scalability)

The **analysis** is at the realization level - comparing the app's fit against each capability's importance:
- App X (fit: 2) realizes Capability A (importance: 5) → Gap: 3 (liability)
- App X (fit: 2) realizes Capability B (importance: 2) → Gap: 0 (aligned)

### Strategic Liability

A realization where:
- The capability has high strategic importance for a pillar
- The realizing application has low fit score for that same pillar

Formula: `Gap = Capability Importance - Application Fit`

---

## Data Model

### Pillar Fit Configuration

Add to strategy pillar configuration:
```
fitScoringEnabled: boolean  // Can applications be scored for this pillar?
fitCriteria: string         // Optional: What aspects to evaluate (for guidance)
```

### Application Fit Score

New aggregate: `ApplicationFitScore`
```
componentId: ComponentId
pillarId: PillarId
score: 1-5
rationale: string (max 500 chars)
scoredAt: timestamp
scoredBy: userId
```

Uniqueness: One score per application per pillar.

---

## User Experience

### Configuration: Pillar Fit Settings

In Settings → Strategy Pillars, add toggle for each pillar:

```
┌─────────────────────────────────────────────────────────────┐
│  Strategy Pillars Configuration                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Always On                                      [Edit]      │
│  Core capabilities that must always be operational          │
│  ☑ Enable fit scoring for realizations                      │
│  Fit criteria: Reliability, uptime SLA, disaster recovery   │
│                                                             │
│  ─────────────────────────────────────────────────────────  │
│                                                             │
│  Grow                                           [Edit]      │
│  Capabilities driving business growth                       │
│  ☑ Enable fit scoring for realizations                      │
│  Fit criteria: Scalability, performance, flexibility        │
│                                                             │
│  ─────────────────────────────────────────────────────────  │
│                                                             │
│  Transform                                      [Edit]      │
│  Capabilities enabling digital transformation               │
│  ☐ Enable fit scoring for realizations                      │
│  (Transformation is about change, not current fitness)      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Scoring: Application Detail Panel

When viewing an application (component), show/edit fit scores for enabled pillars:

```
┌─────────────────────────────────────────────────────────────┐
│  CRM System (Application)                                   │
│  Type: Application · Status: Active                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Strategic Fit Scores                                       │
│  Rate how well this application supports each pillar        │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ Always On                                           │    │
│  │ ●●●●○ 4/5                              [Edit]      │    │
│  │ "99.9% SLA, active DR, automated failover"          │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │ Grow                                                │    │
│  │ ●●○○○ 2/5                              [Edit]      │    │
│  │ "Limited API capacity, vertical scaling only"       │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  Realizes 5 capabilities across 3 domains                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Viewing: Capability Realization with Fit Context

When viewing a capability's realizations, show the app's fit vs the capability's importance:

```
┌─────────────────────────────────────────────────────────────┐
│  Customer Onboarding (Capability)                           │
│  Strategic Importance: Always On ★★★★★ | Grow ★★★☆☆        │
├─────────────────────────────────────────────────────────────┤
│  Realized by:                                               │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ CRM System                                  [Full]  │    │
│  │                                                     │    │
│  │ Fit vs Importance:                                  │    │
│  │   Always On: Fit 4 vs Imp 5 → Gap 1 (Minor)        │    │
│  │   Grow:      Fit 2 vs Imp 3 → Gap 1 (Minor)        │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │ Legacy Portal                             [Partial] │    │
│  │                                                     │    │
│  │ Fit vs Importance:                                  │    │
│  │   Always On: Fit 2 vs Imp 5 → Gap 3 ⚠️ LIABILITY   │    │
│  │   Grow:      Fit 1 vs Imp 3 → Gap 2 ⚠️ LIABILITY   │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Analysis: Strategic Fit Dashboard

New tab in Enterprise Architecture: "Strategic Fit"

```
┌─────────────────────────────────────────────────────────────┐
│  Enterprise Architecture                                    │
├─────────────────────────────────────────────────────────────┤
│  [Capabilities] [Maturity Analysis] [Unlinked] [Strat. Fit] │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Strategic Fit Analysis                                     │
│  Identify realizations with poor strategic fit              │
│                                                             │
│  Filter by pillar: [Always On ▼]                           │
│                                                             │
│  Summary: 12 liabilities · 45 realizations analyzed         │
│                                                             │
│  Strategic Liabilities (Importance > Fit)                   │
│  ─────────────────────────────────────────────────────────  │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ Legacy Portal → Customer Onboarding                 │    │
│  │ Importance: ★★★★★ (5)  Fit: ●●○○○ (2)  Gap: 3     │    │
│  │ "No DR, manual failover"                            │    │
│  │ Enterprise: CUSTOMER MANAGEMENT                     │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │ Old Billing System → Invoice Processing             │    │
│  │ Importance: ★★★★☆ (4)  Fit: ●○○○○ (1)  Gap: 3     │    │
│  │ "Mainframe, no redundancy"                          │    │
│  │ Enterprise: FINANCIAL OPERATIONS                    │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  Well-Aligned (Importance ≤ Fit)                           │
│  ─────────────────────────────────────────────────────────  │
│  ... (collapsed by default)                                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## API Requirements

### Configuration
- `PUT /strategy-pillars/{id}/fit-configuration` - Enable/disable fit scoring, set criteria

### Scoring (Application Level)
- `GET /components/{id}/fit-scores` - Get all fit scores for an application
- `PUT /components/{id}/fit-scores/{pillarId}` - Set/update fit score for a pillar
- `DELETE /components/{id}/fit-scores/{pillarId}` - Remove fit score

### Analysis (Realization Level)
- `GET /strategic-fit-analysis?pillarId=xxx` - Get liabilities and aligned realizations
  - Joins: realization → component fit scores + capability importance
  - Returns realizations with gap calculations
  - Categorizes as liability (gap ≥ 2), minor concern (gap = 1), or aligned (gap ≤ 0)

---

## Calculations

### Fit Score Scale
- 5: Excellent fit - fully supports the pillar's goals
- 4: Good fit - minor gaps
- 3: Adequate fit - acceptable but room for improvement
- 2: Poor fit - significant gaps
- 1: Critical - does not support pillar goals

### Strategic Liability Detection
```
For each realization R of capability C:
  importance = StrategyImportance(C, domain, pillar)
  fit = RealizationFitScore(R, pillar)
  gap = importance - fit

  if gap >= 2: "Strategic Liability"
  if gap == 1: "Minor Concern"
  if gap <= 0: "Well Aligned"
```

---

## Checklist

- [x] Specification approved
- [x] Pillar fit configuration (domain model, API)
- [x] Pillar fit configuration (settings UI)
- [x] ApplicationFitScore aggregate and events
- [x] Application fit score CRUD handlers
- [x] Strategic fit analysis read model (joins realizations + fit scores + importance)
- [x] Analysis API endpoint
- [x] Fit scoring UI in application/component detail panel
- [x] Fit context display in capability realization view
- [x] Strategic Fit dashboard tab
- [x] Tests passing
- [x] User sign-off
