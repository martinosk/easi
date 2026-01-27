# Pillar Fit Type & TIME Suggestions

**Status**: done

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

The system calculates a **suggested** TIME classification for each capability realization (app + domain capability pair) based on pillar gap analysis. This operates on domain capabilities where strategic importance is configured. This is read-only in this spec - architect override comes in Spec 119.

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
capabilityId: CapabilityId
capabilityName: string
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
For a given capability realization (app + domain capability):

1. Get all pillar gaps for this realization from Strategic Fit analysis
   (gap = capability importance - application fit score)

2. Separate into technical and functional pillars based on fitType

3. Calculate average gaps:
   technicalGap = avg(gaps for TECHNICAL pillars)
   functionalGap = avg(gaps for FUNCTIONAL pillars)

4. Apply threshold (default: 15.0):
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
- `GET /time-suggestions?capabilityId={id}` - Filter by domain capability
- `GET /time-suggestions?componentId={id}` - Filter by component
  - All return list of TimeSuggestionReadModel with `_links`

---

## Events

### Extended Events
- `StrategyPillarUpdated` - Now includes fitType changes
- `PillarFitConfigurationUpdated` - Includes fitType in fit configuration updates

---

## Implementation Notes

### Backend

#### Value Objects
- `FitType` - TECHNICAL | FUNCTIONAL with validation
- `TimeClassification` - Tolerate | Invest | Migrate | Eliminate
- `TimeSuggestionConfidence` - High | Medium | Low | Insufficient

#### Domain Services
- `TimeSuggestionCalculator` - Implements the TIME algorithm with configurable gap threshold (default: 15.0)

#### Read Models
- `TimeSuggestionReadModel` - Aggregates data from:
  - `capability_realizations` (component-capability pairs)
  - `capabilities` (domain capability names)
  - `domain_capability_metadata` (business domain context)
  - `effective_capability_importance` (strategic importance by pillar)
  - `application_fit_scores` (component fit scores by pillar)
  - Strategy pillars gateway (pillar fit types)

#### API Handlers
- `TimeSuggestionsHandlers` - GET endpoint with optional filters
- Updated `StrategyPillarsHandlers` - fitType in request/response

### Frontend

#### Types
- `FitType` = 'TECHNICAL' | 'FUNCTIONAL' | ''
- `TimeClassification` = 'Tolerate' | 'Invest' | 'Migrate' | 'Eliminate'
- `TimeSuggestionConfidence` = 'High' | 'Medium' | 'Low' | 'Insufficient'
- `TimeSuggestion` interface with all fields

#### API
- `strategyPillarsApi` - Updated with fitType support
- `enterpriseArchApi.getTimeSuggestions()` - New endpoint

#### Components
- `StrategyPillarsSettings` - Added fit type dropdown selector (only visible when fit scoring enabled)
- `TimeSuggestionsTab` - New tab in Enterprise Architecture page with:
  - Summary statistics (total realizations, high confidence count, eliminate count)
  - TIME legend with classification descriptions
  - Sortable/groupable suggestions table
  - Technical and functional gap display with color coding
  - Confidence badges
  - Empty state with requirements explanation

#### Hooks
- `useTimeSuggestions` - React Query hook for fetching suggestions

### Code Quality
- All components achieve CodeScene code health score of 10.0
- 825 frontend tests passing
- All backend tests passing

---

## Checklist

- [x] Specification approved
- [x] Pillar fitType field (domain model, events)
- [x] Pillar fitType API (update pillar endpoint)
- [x] Pillar fitType UI in strategy pillars settings
- [x] TIME suggestion algorithm implementation
- [x] TIME suggestion read model and projector
- [x] TIME suggestions API endpoint
- [x] TIME suggestions preview UI (temporary)
- [x] Tests passing
- [ ] User sign-off
