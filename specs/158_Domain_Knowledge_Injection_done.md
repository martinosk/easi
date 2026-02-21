# Domain Knowledge Injection

**Status**: done

**Series**: Architecture Assistant Evolution (4 of 6)
- Spec 155: Agent Permission Ceiling
- Spec 156: Generic Tool Executor
- Spec 157: Tool Catalog per Bounded Context
- **Spec 158**: Domain Knowledge Injection (this spec)
- Spec 159: Expand Tool Coverage
- Spec 160: Agent Audit Events

**Depends on:** Spec 157

## User Value

> "As an enterprise architect chatting with the assistant, I want it to understand what business capabilities are, how they relate to business domains and IT systems, and what strategic classification means — so it gives contextually accurate answers instead of generic AI responses."

## Problem

The agent's system prompt says "you help architects explore their application landscape" but teaches nothing about the EASI domain model. It doesn't know:
- Capabilities form an L1-L4 hierarchy
- Only L1 capabilities can be assigned to business domains
- Capability realizations link capabilities to IT systems
- Strategy pillars drive importance ratings and fit scores
- TIME classification (Tolerate/Invest/Migrate/Eliminate) is derived from gap analysis
- Enterprise capabilities are cross-domain groupings distinct from domain capabilities

## Solution: Three Layers

### Layer 1: System Prompt Domain Summary

Expand `systemprompt/builder.go` with a ~250-token static domain model section:

```
EASI Domain Model:

- Capability Hierarchy: Business capabilities form an L1→L4 hierarchy. L1 is the
  top level (e.g. "Customer Management"). L2-L4 are progressively detailed children.
  Only L1 capabilities can be assigned to Business Domains.

- Business Domains: Organizational groupings of L1 capabilities (e.g. "Finance",
  "Customer Experience"). One L1 can belong to multiple domains.

- Capability Realizations: Links between capabilities and application components
  (IT systems). Level: Full, Partial, or Planned. One capability can be realized by
  multiple systems.

- Strategy Pillars: Configurable strategic themes (e.g. "Always On", "Grow",
  "Transform"). Drive two types of scoring:
  - Importance: How critical a capability is for a pillar (1-5, per business domain)
  - Fit Score: How well an application supports a pillar (1-5)
  - Gap = Importance - Fit → classified as liability, concern, or aligned

- Enterprise Capabilities: Cross-domain groupings in the Enterprise Architecture
  view. Link to domain capabilities to discover overlapping or duplicated capabilities across business domains.

- TIME Classification: Investment classification for applications:
  Tolerate, Invest, Migrate, Eliminate. Derived from fit gap analysis.

- Value Streams: Ordered sequences of stages representing business value delivery.
  Stages are mapped to capabilities.

- Component Origins: Applications are linked to their origin — a Vendor (purchased),
  Acquired Entity (from acquisition), or Internal Team (built in-house).
```

### Layer 2: Rich Tool Descriptions

Every `AgentToolSpec.Description` embeds the domain rules relevant to that operation. Examples:

**Before** (current):
> "List all business capabilities. Optionally filter by name substring."

**After**:
> "List business capabilities. Capabilities form an L1→L4 hierarchy representing what the business does (not how). L1 are top-level strategic capabilities and the only level assignable to Business Domains. Each can be realized by application components. Filter by name substring. Returns up to limit results."

**Before**:
> "Link an application to a capability as a realization."

**After**:
> "Record that an application component (IT system) realizes a business capability. Realization level: Full (complete support), Partial (some aspects), Planned (future). One capability can have multiple realizing systems. One system can realize multiple capabilities."

Apply this pattern to all tools in specs 157's context-owned tool declarations.

### Layer 3: Domain Model Query Tool

A read-only tool returning static pre-written domain knowledge — not live data:

```go
AgentToolSpec{
    Name:        "query_domain_model",
    Description: "Get detailed information about the EASI domain model structure, relationships, and business rules. Use when you need to understand how concepts relate before performing operations.",
    Access:      AccessRead,
    Permission:  "assistant:use",
}
```

**Topics** (parameter: `topic`):
- `capability-hierarchy` — L1-L4 rules, parent/child constraints
- `business-domains` — domain structure, L1-only assignment rule
- `realizations` — capability-to-system links, inheritance rules
- `strategy` — pillars, importance, fit scores, gap categories
- `enterprise-capabilities` — cross-domain groupings, linking rules, blocking rules
- `time-classification` — TIME derivation, what each category means
- `value-streams` — stages, capability mapping
- `component-origins` — vendors, acquired entities, internal teams
- `overview` — complete concept map with relationships

This is a custom tool (not generic executor) that returns static markdown content. Lives in `archassistant/infrastructure/toolimpls/domain_knowledge_tool.go`.

## Checklist

- [x] Specification approved
- [x] System prompt expanded with domain model summary in `builder.go`
- [x] All existing tool descriptions enriched with domain context
- [x] `query_domain_model` tool implemented with all 9 topics
- [x] Domain knowledge content reviewed for accuracy
- [x] Unit test: system prompt contains domain model section
- [x] Unit test: domain model tool returns content for each topic
- [x] Build passing
