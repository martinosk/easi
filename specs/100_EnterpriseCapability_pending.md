# Enterprise Capability

## Description
Implement Enterprise Capabilities as an optional grouping mechanism for domain capabilities. Enterprise Capabilities enable discovery and analysis of capabilities that exist across multiple business domains under different names.

## Purpose
Support bottom-up capability consolidation by allowing architects to:
1. Discover overlapping capabilities across domains ("We have 5 Payroll implementations")
2. Track standardization requirements ("Payroll should be consolidated")
3. Analyze cross-domain maturity gaps
4. Prioritize investments based on enterprise-wide view

## Bounded Context
**Enterprise Architecture** - A new bounded context focused on cross-domain capability analysis and standardization tracking.

See [Bounded Context Canvas](../docs/bounded-contexts/EnterpriseArchitecture.md) for full context definition.

**Location**: `/backend/internal/enterprisearchitecture/`

## Dependencies
- Spec 053: Business Domain Aggregate (CapabilityMapping context)
- Spec 023: Capability Model (CapabilityMapping context)
- Spec 098: Strategy Pillars (MetaModel context - pillar definitions)

## Problem Statement

### Current Reality
Organizations often have:
- Siloed capability mapping per business domain
- No standard naming conventions
- Same logical capability with different names:
  - IT Support: "Payroll Management"
  - Customer Service: "Salary Processing"
  - Finance: "Compensation Admin"
  - HR: "Pay & Benefits"
- No visibility into overlap or consolidation opportunities

### Solution: Enterprise Capability as Optional Grouping

Enterprise Capabilities provide a **bottom-up discovery** mechanism:
- Domain capabilities can exist independently (no mandatory linkage)
- When architects discover overlap, they create an Enterprise Capability
- Domain capabilities are then optionally linked to it
- Cross-domain analysis becomes possible

## Strategic Design Decisions

### Bottom-Up, Not Top-Down
- Enterprise Capabilities emerge from discovery, not imposed taxonomy
- Architects create them when they identify overlap
- Linking is optional and incremental
- Domain capabilities keep their local names

### Bounded Context Placement
Enterprise Capability lives in a new **Enterprise Architecture context** because:
- It has a distinct ubiquitous language ("enterprise capability", "standardization candidate", "maturity gap")
- It serves different stakeholders (enterprise architects vs. domain architects)
- It represents analysis/discovery concerns, not operational modeling or configuration
- It provides foundation for Enterprise Strategy context (consolidation workflows)
- It requires cross-domain visibility that neither MetaModel nor CapabilityMapping naturally provides

**Why not MetaModel?** MetaModel is about "what options exist" (configuration). Enterprise Capability is about "what overlaps exist" (analysis).

**Why not CapabilityMapping?** CapabilityMapping is about domain-specific capability modeling. Enterprise Capability is about cross-domain consolidation analysis.

### Strategy Pillar Scope Clarification

| Pillar Scope | Where It Lives | Question It Answers |
|--------------|----------------|---------------------|
| **Domain-scoped** (Spec 098) | DomainCapabilityStrategyAlignment | "How important is this capability for THIS domain's strategy?" |
| **Enterprise-scoped** | EnterpriseCapabilityStrategyAlignment | "Should this capability be standardized across the enterprise?" |

The **Standardization** pillar is most meaningful at enterprise level:
- "Payroll" marked for Standardization = "We should consolidate our 5 payroll implementations"

---

## Aggregate: EnterpriseCapability

### Identity
- Aggregate ID: `EnterpriseCapabilityID` (UUID-based value object with "ent-cap-" prefix)
- Reference: `TenantID` (scoped to tenant)

### Properties
- `name`: EnterpriseCapabilityName (canonical name)
- `description`: EnterpriseCapabilityDescription (what this capability represents)
- `category`: CapabilityCategory (optional grouping)

### Create Enterprise Capability

**Command**: CreateEnterpriseCapability
```
- name: string (non-empty, max 200 chars)
- description: string (optional, max 1000 chars)
- category: string (optional, max 100 chars)
```

**Validation Rules**:
- Name must be unique within tenant (case-insensitive)
- Name cannot be empty

**Event**: EnterpriseCapabilityCreated

### Update Enterprise Capability

**Command**: UpdateEnterpriseCapability
```
- id: EnterpriseCapabilityID
- name: string (non-empty, max 200 chars)
- description: string (optional, max 1000 chars)
- category: string (optional, max 100 chars)
- version: int (optimistic locking)
```

**Event**: EnterpriseCapabilityUpdated

### Delete Enterprise Capability

**Command**: DeleteEnterpriseCapability
```
- id: EnterpriseCapabilityID
```

**Behavior**:
- Soft delete (marks as inactive)
- Existing links from domain capabilities are preserved for history
- New links to this capability are blocked

**Event**: EnterpriseCapabilityDeleted

---

## Aggregate: EnterpriseCapabilityLink

Links a domain capability to an enterprise capability. This is a separate aggregate to:
- Enable independent lifecycle management
- Support querying from both directions
- Maintain loose coupling

### Identity
- Aggregate ID: `EnterpriseLinkID` (UUID-based value object with "ent-link-" prefix)

### Properties
- `enterpriseCapabilityID`: reference to enterprise capability
- `capabilityID`: reference to domain capability
- `linkedAt`: timestamp
- `linkedBy`: UserEmail

### Link Capability to Enterprise Capability

**Command**: LinkCapabilityToEnterpriseCapability
```
- enterpriseCapabilityID: EnterpriseCapabilityID (required)
- capabilityID: CapabilityID (required)
```

**Validation Rules**:
- Enterprise capability must exist and be active
- Domain capability must exist
- Capability not already linked to this enterprise capability
- **A domain capability can only be linked to ONE enterprise capability** (prevents confusion)

**Event**: CapabilityLinkedToEnterpriseCapability

### Unlink Capability from Enterprise Capability

**Command**: UnlinkCapabilityFromEnterpriseCapability
```
- linkID: EnterpriseLinkID
```

**Event**: CapabilityUnlinkedFromEnterpriseCapability

---

## Aggregate: EnterpriseCapabilityStrategyAlignment

Strategy pillar alignment at enterprise level (separate from domain-scoped alignment in Spec 098).

### Identity
- Aggregate ID: `EnterpriseAlignmentID` (UUID-based value object with "ent-strat-" prefix)

### Properties
- `enterpriseCapabilityID`: reference to enterprise capability
- `pillarID`: reference to strategy pillar
- `strategicImportance`: 1-5 scale

### Align Enterprise Capability to Strategy Pillar

**Command**: AlignEnterpriseCapabilityToStrategyPillar
```
- enterpriseCapabilityID: EnterpriseCapabilityID (required)
- pillarID: PillarID (required)
- strategicImportance: int (1-5, required)
```

**Validation Rules**:
- Enterprise capability must exist
- Pillar must exist and be active
- Combination must be unique

**Event**: EnterpriseCapabilityAlignedToStrategyPillar

### Update Strategic Importance

**Command**: UpdateEnterpriseStrategicImportance
```
- alignmentID: EnterpriseAlignmentID
- strategicImportance: int (1-5)
```

**Event**: EnterpriseCapabilityStrategyAlignmentUpdated

### Unalign

**Command**: UnalignEnterpriseCapabilityFromStrategyPillar
```
- alignmentID: EnterpriseAlignmentID
```

**Event**: EnterpriseCapabilityUnalignedFromStrategyPillar

---

## Value Objects

### EnterpriseCapabilityID
UUID-based identifier with "ent-cap-" prefix.

### EnterpriseCapabilityName
**Validation**: Non-empty, max 200 characters, trimmed whitespace.

### EnterpriseCapabilityDescription
**Validation**: Max 1000 characters, allows empty.

### CapabilityCategory
**Validation**: Max 100 characters, allows empty. Optional grouping for enterprise capabilities.

### EnterpriseLinkID
UUID-based identifier with "ent-link-" prefix.

### EnterpriseAlignmentID
UUID-based identifier with "ent-strat-" prefix.

---

## Read Models

### enterprise_capabilities

```sql
CREATE TABLE enterprise_capabilities (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    description VARCHAR(1000),
    category VARCHAR(100),
    is_active BOOLEAN NOT NULL DEFAULT true,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL,
    created_by VARCHAR(255),
    modified_at TIMESTAMPTZ NOT NULL,
    modified_by VARCHAR(255)
);

CREATE UNIQUE INDEX idx_enterprise_cap_name ON enterprise_capabilities(tenant_id, name) WHERE is_active = true;
CREATE INDEX idx_enterprise_cap_tenant ON enterprise_capabilities(tenant_id);
CREATE INDEX idx_enterprise_cap_category ON enterprise_capabilities(tenant_id, category) WHERE is_active = true;
```

### enterprise_capability_links

```sql
CREATE TABLE enterprise_capability_links (
    id VARCHAR(50) PRIMARY KEY,
    enterprise_capability_id VARCHAR(50) NOT NULL,
    capability_id VARCHAR(50) NOT NULL,
    linked_at TIMESTAMPTZ NOT NULL,
    linked_by VARCHAR(255),
    CONSTRAINT fk_link_enterprise FOREIGN KEY (enterprise_capability_id)
        REFERENCES enterprise_capabilities(id),
    CONSTRAINT fk_link_capability FOREIGN KEY (capability_id)
        REFERENCES capabilities(id)
);

CREATE UNIQUE INDEX idx_enterprise_link_unique ON enterprise_capability_links(capability_id);
CREATE INDEX idx_enterprise_link_enterprise ON enterprise_capability_links(enterprise_capability_id);
CREATE INDEX idx_enterprise_link_capability ON enterprise_capability_links(capability_id);
```

### enterprise_capability_strategy_alignments

```sql
CREATE TABLE enterprise_capability_strategy_alignments (
    id VARCHAR(50) PRIMARY KEY,
    enterprise_capability_id VARCHAR(50) NOT NULL,
    pillar_id VARCHAR(50) NOT NULL,
    strategic_importance INT NOT NULL CHECK (strategic_importance BETWEEN 1 AND 5),
    aligned_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT fk_ent_align_enterprise FOREIGN KEY (enterprise_capability_id)
        REFERENCES enterprise_capabilities(id),
    CONSTRAINT fk_ent_align_pillar FOREIGN KEY (pillar_id)
        REFERENCES strategy_pillars(id)
);

CREATE UNIQUE INDEX idx_ent_cap_pillar_unique
    ON enterprise_capability_strategy_alignments(enterprise_capability_id, pillar_id);
CREATE INDEX idx_ent_align_enterprise
    ON enterprise_capability_strategy_alignments(enterprise_capability_id);
CREATE INDEX idx_ent_align_pillar
    ON enterprise_capability_strategy_alignments(pillar_id);
```

---

## Query Views for Analysis

### Enterprise Capability Overview (with linked capabilities)

```sql
CREATE VIEW enterprise_capability_overview AS
SELECT
    ec.id as enterprise_capability_id,
    ec.name as enterprise_capability_name,
    ec.category,
    COUNT(ecl.id) as linked_capability_count,
    COUNT(DISTINCT c.business_domain_id) as domain_count
FROM enterprise_capabilities ec
LEFT JOIN enterprise_capability_links ecl ON ec.id = ecl.enterprise_capability_id
LEFT JOIN capabilities c ON ecl.capability_id = c.id
WHERE ec.is_active = true
GROUP BY ec.id, ec.name, ec.category;
```

### Maturity Gap Analysis (across linked capabilities)

```sql
CREATE VIEW enterprise_capability_maturity_gaps AS
SELECT
    ec.id as enterprise_capability_id,
    ec.name as enterprise_capability_name,
    bd.id as business_domain_id,
    bd.name as business_domain_name,
    c.id as capability_id,
    c.name as capability_name,
    c.maturity_value as current_maturity,
    ms.section_name as maturity_section
FROM enterprise_capabilities ec
JOIN enterprise_capability_links ecl ON ec.id = ecl.enterprise_capability_id
JOIN capabilities c ON ecl.capability_id = c.id
JOIN business_domains bd ON c.business_domain_id = bd.id
LEFT JOIN maturity_scale_configs msc ON msc.tenant_id = ec.tenant_id
LEFT JOIN LATERAL (
    SELECT name as section_name
    FROM get_maturity_section(msc.id, c.maturity_value)
) ms ON true
WHERE ec.is_active = true;
```

### Standardization Candidates

```sql
CREATE VIEW standardization_candidates AS
SELECT
    ec.id as enterprise_capability_id,
    ec.name as enterprise_capability_name,
    ecsa.strategic_importance as standardization_importance,
    COUNT(ecl.id) as implementation_count,
    MIN(c.maturity_value) as min_maturity,
    MAX(c.maturity_value) as max_maturity,
    MAX(c.maturity_value) - MIN(c.maturity_value) as maturity_spread
FROM enterprise_capabilities ec
JOIN enterprise_capability_strategy_alignments ecsa ON ec.id = ecsa.enterprise_capability_id
JOIN strategy_pillars sp ON ecsa.pillar_id = sp.id AND sp.name = 'Standardization'
JOIN enterprise_capability_links ecl ON ec.id = ecl.enterprise_capability_id
JOIN capabilities c ON ecl.capability_id = c.id
WHERE ec.is_active = true
GROUP BY ec.id, ec.name, ecsa.strategic_importance
HAVING COUNT(ecl.id) > 1
ORDER BY ecsa.strategic_importance DESC, maturity_spread DESC;
```

---

## Event Handlers

### OnCapabilityDeleted
When a domain capability is deleted, automatically unlink from enterprise capability.

### OnEnterpriseCapabilityDeleted
When an enterprise capability is deleted (soft), existing links remain for history but show as "unlinked from deleted enterprise capability" in UI.

---

## Use Cases

### UC1: Discover Overlap
1. Architect notices "Payroll Management" in IT Support domain
2. Searches for similar capabilities, finds "Salary Processing" in Customer Service
3. Creates Enterprise Capability: "Payroll"
4. Links both domain capabilities to it
5. System shows: "2 implementations across 2 domains"

### UC2: Track Standardization
1. Enterprise architect decides Payroll should be standardized
2. Aligns "Payroll" enterprise capability with "Standardization" pillar at importance 5
3. Standardization Candidates view shows:
   - Payroll: 2 implementations, maturity spread 50 (Genesis vs Product)
4. Prioritizes investment to consolidate

### UC3: Gap Analysis
1. Query Enterprise Capability Maturity Gaps view
2. Shows per-domain maturity for all linked capabilities
3. Identifies: IT Support's Payroll at Genesis needs investment to reach Product

---

## Checklist
- [ ] Specification ready
- [ ] EnterpriseCapability Aggregate
  - [ ] Value objects implemented
  - [ ] Aggregate with event sourcing
  - [ ] Events defined
  - [ ] Command handlers
  - [ ] Projector
  - [ ] Repository
  - [ ] Unit tests
- [ ] EnterpriseCapabilityLink Aggregate
  - [ ] Value objects
  - [ ] Aggregate with event sourcing
  - [ ] Events defined
  - [ ] Command handlers
  - [ ] Projector
  - [ ] Repository
  - [ ] Unit tests
- [ ] EnterpriseCapabilityStrategyAlignment Aggregate
  - [ ] Value objects
  - [ ] Aggregate with event sourcing
  - [ ] Events defined
  - [ ] Command handlers
  - [ ] Projector
  - [ ] Repository
  - [ ] Unit tests
- [ ] Read model migrations
- [ ] Query views
- [ ] Event handlers (cascade delete)
- [ ] User sign-off
