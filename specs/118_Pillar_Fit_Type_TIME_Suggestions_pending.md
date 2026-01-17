# Pillar Fit Type & TIME Suggestions

**Status**: pending

**Series**: Application Landscape & TIME (2 of 4)
- Spec 117: Portfolio Metadata Foundation
- **Spec 118**: Pillar Fit Type & TIME Suggestions (this spec)
- Spec 119: TIME Classification & Application Landscape
- Spec 120: Portfolio Rationalization Analysis

## User Value

> "As an enterprise architect, I want to categorize strategy pillars as primarily assessing technical or functional fit, so the system can suggest TIME classifications based on our existing Strategic Fit data."

> "As a portfolio manager, I want to see automated TIME suggestions for each application realization, so I can quickly identify which applications should be invested in, tolerated, migrated, or eliminated."

> "As an analyst, I want to validate that the TIME suggestion algorithm produces sensible results before architects start making official classifications."

## Dependencies

- Spec 098: Strategy Pillars
- Spec 100: Enterprise Capability Groupings
- Spec 103: Strategic Fit Analysis (provides pillar gap data)

---

## Background: Gartner TIME Model

The TIME framework classifies applications based on two dimensions:

```
                        FUNCTIONAL FIT
                     Low           High
                 ┌──────────┬──────────┐
            High │ TOLERATE │  INVEST  │
TECHNICAL        │          │          │
FIT              ├──────────┼──────────┤
            Low  │ELIMINATE │ MIGRATE  │
                 │          │          │
                 └──────────┴──────────┘
```

- **Tolerate**: Good technology, poor business fit. Keep running but don't enhance.
- **Invest**: Good technology, good business fit. This is the standard.
- **Migrate**: Poor technology, good business fit. Move to better platform.
- **Eliminate**: Poor technology, poor business fit. Retire.

**Key insight**: Your existing strategy pillars can represent both technical and functional fitness dimensions. By tagging each pillar, the system can calculate TIME suggestions from Strategic Fit gap data.

---

## Domain Concepts

### Pillar Fit Type

Strategy pillars can be tagged as primarily assessing **technical** or **functional** fit:

| Pillar Example | Fit Type | Rationale |
|----------------|----------|-----------|
| Transformation | TECHNICAL | Modern tech, API-first, good data products |
| Always On | TECHNICAL | Reliability, DR, uptime |
| Grow | FUNCTIONAL | Scales with business, market reach |
| Customer Focus | FUNCTIONAL | Supports customer-facing capabilities |

Pillars without a fit type are excluded from TIME calculation.

### TIME Suggestion

The system calculates a **suggested** TIME classification for each enterprise capability realization (app + enterprise capability pair) based on pillar gap analysis. This is read-only in this spec - architect override comes in Spec 119.

---

## Data Model

### Pillar Fit Type (MetaModel extension)

Add to StrategyPillar:
```
fitType: TECHNICAL | FUNCTIONAL | null
```

If null, the pillar is excluded from TIME suggestion calculation.

### TIME Suggestion Read Model

New read model: `TimeSuggestionReadModel`
```
enterpriseCapabilityId: EnterpriseCapabilityId
enterpriseCapabilityName: string
componentId: ComponentId
componentName: string
suggestedTime: TIME | null
technicalGap: number | null
functionalGap: number | null
confidence: LOW | MEDIUM | HIGH  // based on data completeness
```

TIME enum: `TOLERATE | INVEST | MIGRATE | ELIMINATE`

---

## TIME Suggestion Algorithm

```
For a given enterprise capability realization (app + enterprise cap):

1. Get all pillar gaps for this realization from Strategic Fit analysis
   (gap = capability importance - application fit score)

2. Separate into technical and functional pillars based on fitType

3. Calculate average gaps:
   technicalGap = avg(gaps for TECHNICAL pillars)
   functionalGap = avg(gaps for FUNCTIONAL pillars)

4. Apply threshold (default: 1.5):
   highTechnicalGap = technicalGap >= threshold
   highFunctionalGap = functionalGap >= threshold

5. Determine suggested TIME:
   if NOT highTechnicalGap AND NOT highFunctionalGap: INVEST
   if NOT highTechnicalGap AND highFunctionalGap: TOLERATE
   if highTechnicalGap AND NOT highFunctionalGap: MIGRATE
   if highTechnicalGap AND highFunctionalGap: ELIMINATE

6. Determine confidence:
   - HIGH: All pillars have fit scores and importance ratings
   - MEDIUM: At least one technical AND one functional pillar scored
   - LOW: Insufficient data (fewer than 2 pillars with complete data)

7. If insufficient data (no scored pillars of either type), suggestedTime = null
```

---

## User Experience

### Configuration: Pillar Fit Type

In Settings → Strategy Pillars, add fit type selector:

```
┌─────────────────────────────────────────────────────────────────┐
│  Strategy Pillars Configuration                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Transformation                                      [Edit]      │
│  Capabilities enabling digital transformation                    │
│  ☑ Enable fit scoring                                           │
│  Fit type: [TECHNICAL ▼]                                        │
│  Fit criteria: Modern architecture, API-first, good data...     │
│                                                                  │
│  ─────────────────────────────────────────────────────────────  │
│                                                                  │
│  Grow                                                [Edit]      │
│  Capabilities driving business growth                            │
│  ☑ Enable fit scoring                                           │
│  Fit type: [FUNCTIONAL ▼]                                       │
│  Fit criteria: Scalability, market reach, flexibility...        │
│                                                                  │
│  ─────────────────────────────────────────────────────────────  │
│                                                                  │
│  Innovation                                          [Edit]      │
│  New market opportunities                                        │
│  ☐ Enable fit scoring                                           │
│  Fit type: [Not set ▼]  (excluded from TIME calculation)        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Viewing: TIME Suggestions (Preview)

Temporary view for validating algorithm before Spec 119 delivers the full UI:

```
┌─────────────────────────────────────────────────────────────────┐
│  TIME Suggestions (Preview)                                      │
│  System-calculated suggestions based on Strategic Fit gaps       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Filter: [All ▼]  Confidence: [All ▼]                           │
│                                                                  │
│  PAYROLL PROCESSING                                              │
│  ├─ SAP HR           → INVEST    (Tech: -0.5, Func: -0.2) HIGH  │
│  └─ Legacy Payroll   → MIGRATE   (Tech: 2.1, Func: 0.3)  HIGH   │
│                                                                  │
│  CUSTOMER ONBOARDING                                             │
│  ├─ CRM System       → TOLERATE  (Tech: 0.2, Func: 1.8)  MEDIUM │
│  └─ Portal           → ELIMINATE (Tech: 2.5, Func: 2.0)  HIGH   │
│                                                                  │
│  EXPENSE MANAGEMENT                                              │
│  ├─ Excel            → ?         (insufficient data)     LOW    │
│  └─ SAP Concur       → INVEST    (Tech: -0.3, Func: 0.1) HIGH   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## API Requirements

All endpoints follow REST Level 3 with HATEOAS. Responses include `_links` for navigation.

### Pillar Fit Type (MetaModel)
- `PUT /strategy-pillars/{id}` - Update pillar including fitType
- `GET /strategy-pillars` - Returns fitType in pillar list

### TIME Suggestions (Enterprise Architecture)
- `GET /time-suggestions` - Get all calculated TIME suggestions
- `GET /time-suggestions/by-enterprise-capability/{enterpriseCapabilityId}` - Filter by enterprise capability
- `GET /time-suggestions/by-component/{componentId}` - Filter by component
  - All return list of TimeSuggestionReadModel with `_links`

---

## Events

### Extended Events
- `StrategyPillarUpdated` - Now includes fitType changes

---

## Checklist

- [ ] Specification approved
- [ ] Pillar fitType field (domain model, events)
- [ ] Pillar fitType API (update pillar endpoint)
- [ ] Pillar fitType UI in strategy pillars settings
- [ ] TIME suggestion algorithm implementation
- [ ] TIME suggestion read model and projector
- [ ] TIME suggestions API endpoint
- [ ] TIME suggestions preview UI (temporary)
- [ ] Tests passing
- [ ] User sign-off
