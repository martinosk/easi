# Hierarchical Strategic Rating Evaluation

**Status**: ongoing

## User Value

> "As an enterprise architect, I want strategic fit analysis to evaluate gaps against the most specific capability rating in the hierarchy, so that I can capture nuanced strategic importance at different granularity levels."

> "As a portfolio manager, I want to see all gap analyses when an application realizes multiple capabilities in a parent/child chain, so that I can understand the full strategic impact."

## Dependencies

- Spec 099: Domain Capability Strategy Alignment
- Spec 103: Strategic Fit Analysis
- Spec 023: Capability Model (L1-L4 hierarchy)

---

## Business Rules

### Rule 1: Most Specific Rating Wins

When evaluating strategic fit gaps, use the rating from the most specific (lowest level that the app realises) capability that has a rating for the given pillar.

### Rule 2: Rating Inheritance (Fallback)

If a capability is not rated for a pillar, walk up the hierarchy to find the nearest ancestor with a rating for that pillar.

### Rule 3: Multi-Capability Realizations

When an application realizes multiple capabilities within the same parent/child chain, show separate gap analyses for each rated capability in that chain.

---

## Scenarios

### Scenario 1: Child Capability Has Rating, Use Child's Rating

```gherkin
Given capability "Payment Processing" (L1) has importance 3 for pillar "Always On"
And capability "Card Payments" (L2) is a child of "Payment Processing"
And capability "Card Payments" has importance 5 for pillar "Always On"
And application "Payment Gateway" realizes "Card Payments"
And application "Payment Gateway" has fit score 2 for pillar "Always On"
When strategic fit analysis is performed for pillar "Always On"
Then the gap for "Payment Gateway" → "Card Payments" is calculated as 5 - 2 = 3
And the gap is categorized as "liability"
```

### Scenario 2: Child Capability Has No Rating, Use Parent's Rating

```gherkin
Given capability "Payment Processing" (L1) has importance 4 for pillar "Always On"
And capability "Card Payments" (L2) is a child of "Payment Processing"
And capability "Card Payments" has no rating for pillar "Always On"
And application "Payment Gateway" realizes "Card Payments"
And application "Payment Gateway" has fit score 2 for pillar "Always On"
When strategic fit analysis is performed for pillar "Always On"
Then the gap for "Payment Gateway" → "Card Payments" is calculated using inherited importance 4
And the gap is 4 - 2 = 2
And the gap is categorized as "liability"
```

### Scenario 3: Deep Hierarchy Rating Inheritance

```gherkin
Given capability "Customer Management" (L1) has importance 5 for pillar "Grow"
And capability "Customer Onboarding" (L2) is a child of "Customer Management"
And capability "Customer Onboarding" has no rating for pillar "Grow"
And capability "Identity Verification" (L3) is a child of "Customer Onboarding"
And capability "Identity Verification" has no rating for pillar "Grow"
And application "KYC System" realizes "Identity Verification"
And application "KYC System" has fit score 3 for pillar "Grow"
When strategic fit analysis is performed for pillar "Grow"
Then the gap for "KYC System" → "Identity Verification" uses inherited importance 5 from "Customer Management"
And the gap is 5 - 3 = 2
And the gap is categorized as "liability"
```

### Scenario 4: Mid-Hierarchy Rating Takes Precedence Over Parent

```gherkin
Given capability "Customer Management" (L1) has importance 3 for pillar "Transform"
And capability "Customer Onboarding" (L2) is a child of "Customer Management"
And capability "Customer Onboarding" has importance 5 for pillar "Transform"
And capability "Identity Verification" (L3) is a child of "Customer Onboarding"
And capability "Identity Verification" has no rating for pillar "Transform"
And application "KYC System" realizes "Identity Verification"
And application "KYC System" has fit score 2 for pillar "Transform"
When strategic fit analysis is performed for pillar "Transform"
Then the gap for "KYC System" → "Identity Verification" uses inherited importance 5 from "Customer Onboarding"
And the gap is 5 - 2 = 3
And the gap is categorized as "liability"
```

### Scenario 5: Application Realizes Multiple Capabilities in Same Chain - Show All Gaps

```gherkin
Given capability "Payment Processing" (L1) has importance 4 for pillar "Always On"
And capability "Card Payments" (L2) is a child of "Payment Processing"
And capability "Card Payments" has importance 5 for pillar "Always On"
And application "Payment Gateway" realizes both "Payment Processing" AND "Card Payments"
And application "Payment Gateway" has fit score 2 for pillar "Always On"
When strategic fit analysis is performed for pillar "Always On"
Then two gap entries are shown:
  | Capability         | Importance | Gap |
  | Payment Processing | 4          | 2   |
  | Card Payments      | 5          | 3   |
And both entries reference the same application "Payment Gateway"
```

### Scenario 6: Application Realizes Multiple Capabilities - Mixed Rated and Unrated

```gherkin
Given capability "Customer Management" (L1) has importance 4 for pillar "Grow"
And capability "Customer Onboarding" (L2) is a child of "Customer Management"
And capability "Customer Onboarding" has importance 5 for pillar "Grow"
And capability "Document Collection" (L3) is a child of "Customer Onboarding"
And capability "Document Collection" has no rating for pillar "Grow"
And application "Onboarding Portal" realizes "Customer Onboarding", "Document Collection"
And application "Onboarding Portal" has fit score 3 for pillar "Grow"
When strategic fit analysis is performed for pillar "Grow"
Then two gap entries are shown:
  | Capability           | Importance | Source                    | Gap |
  | Customer Onboarding  | 5          | direct rating             | 2   |
  | Document Collection  | 5          | inherited from L2 parent  | 2   |
```

### Scenario 7: Capability Chain With No Ratings Anywhere

```gherkin
Given capability "Support Operations" (L1) has no rating for pillar "Transform"
And capability "Ticket Management" (L2) is a child of "Support Operations"
And capability "Ticket Management" has no rating for pillar "Transform"
And application "Helpdesk System" realizes "Ticket Management"
And application "Helpdesk System" has fit score 4 for pillar "Transform"
When strategic fit analysis is performed for pillar "Transform"
Then no gap entry is shown for "Helpdesk System" → "Ticket Management"
Because there is no importance rating in the capability hierarchy for this pillar
```

### Scenario 8: Application Realizes Capabilities in Different Branches

```gherkin
Given capability "Sales" (L1) has importance 4 for pillar "Grow"
And capability "Lead Management" (L2) is a child of "Sales"
And capability "Lead Management" has importance 3 for pillar "Grow"
And capability "Marketing" (L1) has importance 5 for pillar "Grow"
And capability "Campaign Management" (L2) is a child of "Marketing"
And capability "Campaign Management" has no rating for pillar "Grow"
And application "CRM System" realizes "Lead Management" AND "Campaign Management"
And application "CRM System" has fit score 3 for pillar "Grow"
When strategic fit analysis is performed for pillar "Grow"
Then two gap entries are shown:
  | Capability          | Importance | Source               | Gap |
  | Lead Management     | 3          | direct rating        | 0   |
  | Campaign Management | 5          | inherited from L1    | 2   |
And the entries are in different capability branches (no parent/child relationship)
```

### Scenario 9: Only Parent Rated, Child Realized - Shows Under Child Name

```gherkin
Given capability "Finance" (L1) has importance 5 for pillar "Always On"
And capability "Accounts Payable" (L2) is a child of "Finance"
And capability "Accounts Payable" has no rating for pillar "Always On"
And application "AP System" realizes only "Accounts Payable"
And application "AP System" has fit score 1 for pillar "Always On"
When strategic fit analysis is performed for pillar "Always On"
Then one gap entry is shown for "AP System" → "Accounts Payable"
And the entry shows:
  | Field            | Value                    |
  | Capability       | Accounts Payable         |
  | Importance       | 5                        |
  | Importance Source| Inherited from Finance   |
  | Fit Score        | 1                        |
  | Gap              | 4                        |
  | Category         | liability                |
```

### Scenario 10: Pillar-Specific Rating Inheritance

```gherkin
Given capability "HR" (L1) has importance 5 for pillar "Always On"
And capability "HR" (L1) has importance 2 for pillar "Transform"
And capability "Recruitment" (L2) is a child of "HR"
And capability "Recruitment" has importance 4 for pillar "Transform"
And capability "Recruitment" has no rating for pillar "Always On"
And application "HR Suite" realizes "Recruitment"
And application "HR Suite" has fit score 3 for pillar "Always On"
And application "HR Suite" has fit score 2 for pillar "Transform"
When strategic fit analysis is performed for pillar "Always On"
Then the gap uses inherited importance 5 from "HR" (since Recruitment has no Always On rating)
And the gap is 5 - 3 = 2
When strategic fit analysis is performed for pillar "Transform"
Then the gap uses direct importance 4 from "Recruitment"
And the gap is 4 - 2 = 2
```

---

## Solution Architecture

### Design Rationale

The hierarchical rating resolution logic is complex business logic that must be testable and easy to reason about. Placing this logic directly in read model SQL queries would make it difficult to test and maintain. Instead, the architecture separates concerns:

- **Domain layer** owns the hierarchy traversal logic (testable, reusable)
- **Projectors** pre-compute effective importance values on data changes
- **Read models** perform simple joins against pre-computed data (fast queries)

### Component Overview

```
┌─────────────────────────────────────────────────────────────┐
│  Domain Service: HierarchicalRatingResolver                 │
│  - Encapsulates Rules 1 and 2 (most specific wins, fallback)│
│  - Pure logic, no infrastructure dependencies               │
│  - Unit testable with all 10 scenarios                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Projector: EffectiveImportanceProjector                    │
│  - Listens to importance and hierarchy change events        │
│  - Uses domain service to compute effective values          │
│  - Writes to pre-computed read model table                  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Read Model: EffectiveCapabilityImportance                  │
│  - Stores resolved importance per capability/pillar/domain  │
│  - Includes source capability and inheritance flag          │
│  - Strategic fit analysis joins here instead of raw ratings │
└─────────────────────────────────────────────────────────────┘
```

### Domain Service: HierarchicalRatingResolver

**Location**: `capabilitymapping/domain/services`

**Responsibility**: Given a capability, pillar, and business domain, resolve the effective importance by walking up the hierarchy until a rating is found.

**Method**: `ResolveEffectiveImportance(capabilityID, pillarID, businessDomainID)`

**Returns**: A value object containing:
- The resolved importance value (1-5)
- The source capability ID (where the rating originated)
- Whether the rating is inherited (boolean)
- Returns nil if no rating exists anywhere in the hierarchy

**Dependencies**:
- CapabilityHierarchyService (existing) - to walk up the parent chain
- StrategyImportanceRepository - to check for ratings at each level

### Value Object: EffectiveImportance

**Location**: `capabilitymapping/domain/valueobjects`

**Fields**:
- Importance (reuse existing Importance value object)
- SourceCapabilityID (CapabilityID)
- IsInherited (boolean)

### Read Model Table: effective_capability_importance

**Purpose**: Pre-computed materialized view of resolved importance for every capability that has an effective rating (direct or inherited).

**Key**: (tenant_id, capability_id, pillar_id, business_domain_id)

**Fields**:
- effective_importance and importance_label
- source_capability_id and source_capability_name
- is_inherited flag

### Projector: EffectiveImportanceProjector

**Location**: `capabilitymapping/application/projectors`

**Event Handlers**:

| Event | Action |
|-------|--------|
| StrategyImportanceSet | Recompute for the rated capability and all its descendants |
| StrategyImportanceUpdated | Recompute for the rated capability and all its descendants |
| StrategyImportanceRemoved | Recompute for the capability and all its descendants |
| CapabilityParentChanged | Recompute for the moved capability and all its descendants |
| CapabilityDeleted | Delete all effective importance entries for the capability |

The projector uses CapabilityHierarchyService.GetDescendants() to find all capabilities that may be affected by a rating change.

### Modified Strategic Fit Analysis Read Model

**Change**: Replace the current join to `strategy_importance` via `l1_capability_id` with a direct join to `effective_capability_importance` using the realized `capability_id`.

**Result**:
- Each realization row gets its own effective importance (Rule 3: multi-capability realizations)
- The importance reflects the most specific rating with fallback (Rules 1 and 2)
- Query remains simple - complexity is handled by pre-computation

### API Response Enhancement

The strategic fit analysis response should include importance source information:
- `importanceSource`: The name of the capability where the rating originated
- `isInherited`: Whether the importance was inherited from an ancestor

This enables scenarios 6, 8, and 9 which show the source of inherited ratings.

### Testing Strategy

| Layer | Test Approach |
|-------|---------------|
| HierarchicalRatingResolver | Unit tests covering all 10 scenarios using in-memory test doubles |
| EffectiveImportanceProjector | Integration tests verifying correct table updates on events |
| Strategic Fit Analysis | Integration tests verifying correct gap calculations |
| API | E2E tests for full response validation |

---

## Checklist

- [x] Specification approved
- [x] Unit tests for all scenarios
- [x] Integration tests
- [ ] User sign-off
