package toolimpls

import (
	"context"
	"fmt"

	"easi/backend/internal/archassistant/application/tools"
)

type domainKnowledgeTool struct{}

func (t *domainKnowledgeTool) Execute(_ context.Context, args map[string]interface{}) tools.ToolResult {
	topic, errResult := requireString(args, "topic")
	if errResult != nil {
		return *errResult
	}
	content, ok := domainTopics[topic]
	if !ok {
		return tools.ToolResult{
			Content: fmt.Sprintf("Unknown topic: %q. Available topics: capability-hierarchy, business-domains, realizations, strategy, enterprise-capabilities, time-classification, value-streams, component-origins, overview", topic),
			IsError: true,
		}
	}
	return tools.ToolResult{Content: content}
}

var domainTopics = map[string]string{
	"capability-hierarchy": `# Capability Hierarchy

Business capabilities represent **what** the business does, not how it does it.
They form a strict L1→L4 hierarchy:

- **L1** (top-level): Strategic capabilities (e.g. "Customer Management", "Order Fulfillment").
  L1 is the only level that can be assigned to Business Domains.
- **L2**: Children of L1 — more specific business functions.
- **L3**: Children of L2 — detailed operational capabilities.
- **L4**: Children of L3 — the most granular level.

Rules:
- Every capability has exactly one level (L1-L4).
- L1 capabilities have no parent.
- L2-L4 must have a parent one level above (L2→L1, L3→L2, L4→L3).
- A capability can have multiple children at the next level.
- Deleting a capability with children is not allowed — remove children first.`,

	"business-domains": `# Business Domains

Business domains are organizational groupings that represent major areas of the business
(e.g. "Finance", "Customer Experience", "Supply Chain").

Rules:
- Only **L1 capabilities** can be assigned to a business domain.
- One L1 capability can belong to **multiple** business domains.
- Domains help organize the capability landscape for different stakeholder views.
- Removing an L1 from a domain does not delete the capability itself.`,

	"realizations": `# Capability Realizations

A realization is a link between a business capability and an application component (IT system),
recording that the system supports or implements that capability.

Realization levels:
- **Full**: The system completely supports the capability.
- **Partial**: The system supports some aspects of the capability.
- **Planned**: The system is planned to support the capability in the future.

Rules:
- One capability can be realized by **multiple** systems.
- One system can realize **multiple** capabilities.
- Realizations can exist at any capability level (L1-L4).
- Removing a realization does not affect the capability or the application.`,

	"strategy": `# Strategy Pillars, Importance & Fit Scores

Strategy pillars are configurable strategic themes that drive portfolio analysis
(e.g. "Always On", "Grow Revenue", "Digital Transformation").

Two scoring dimensions per pillar:
- **Importance** (1-5): How critical a capability is for achieving the pillar's goals.
  Scored per business domain — the same capability may have different importance in different domains.
- **Fit Score** (1-5): How well an application component supports the pillar.
  Scored per application.

Gap analysis:
- **Gap = Importance - Fit Score**
- Positive gap → the capability is more important than the system's support level.
- Classification:
  - **Aligned**: Gap ≤ 0 (system meets or exceeds the need)
  - **Concern**: Gap = 1 (minor shortfall)
  - **Liability**: Gap ≥ 2 (significant shortfall requiring attention)`,

	"enterprise-capabilities": `# Enterprise Capabilities

Enterprise capabilities provide a cross-domain view of the organization's capability landscape.
They live in the Enterprise Architecture view and group related domain capabilities
that serve similar business functions across different business domains.

Purpose:
- Discover **overlapping** capabilities across business domains.
- Identify **duplication** where multiple domains implement similar functions.
- Drive rationalization decisions — consolidate or differentiate.

Rules:
- Enterprise capabilities link to domain capabilities (L1-L4).
- One domain capability can be linked to multiple enterprise capabilities.
- Enterprise capabilities exist independently of business domains.`,

	"time-classification": `# TIME Classification

TIME is an investment classification framework for applications:

- **Tolerate**: The application works but has known limitations. No active investment.
  Maintain minimally until replacement is available.
- **Invest**: The application is strategic and should receive continued investment
  to enhance capabilities and address gaps.
- **Migrate**: The application should be replaced. Plan migration to a target system.
  Minimize new development — only critical fixes.
- **Eliminate**: The application is being actively retired. Migrate remaining users
  and decommission.

TIME classification is derived from fit gap analysis:
- Applications with large gaps across strategic pillars trend toward Migrate/Eliminate.
- Applications with strong fit scores trend toward Invest.
- Applications with acceptable but not strategic fit trend toward Tolerate.`,

	"value-streams": `# Value Streams

Value streams represent the sequence of activities (stages) that deliver business value
to a customer or stakeholder.

Structure:
- A value stream has an ordered list of **stages**.
- Each stage represents a step in the value delivery process.
- Stages are mapped to **capabilities** — showing which business capabilities
  are needed at each step.

Purpose:
- Visualize end-to-end value delivery.
- Identify which capabilities (and their realizing systems) support each stage.
- Spot capability gaps in the value chain.`,

	"component-origins": `# Component Origins

Every application component (IT system) has an origin that describes how it was obtained:

- **Vendor** (purchased): A commercial off-the-shelf (COTS) product from a software vendor.
  Linked to a vendor entity with details like vendor name.
- **Acquired Entity** (from acquisition): A system inherited through a business acquisition.
  Linked to an acquired entity with details about the acquisition.
- **Internal Team** (built in-house): A system developed internally by the organization.
  Linked to the team or department responsible.

Purpose:
- Track total cost of ownership by origin type.
- Plan vendor consolidation strategies.
- Identify systems from acquisitions that may need migration or integration.`,

	"overview": `# EASI Domain Model Overview

## Core Concepts and Relationships

**Business Capabilities** (L1→L4 hierarchy)
  └── Assigned to → **Business Domains** (L1 only)
  └── Realized by → **Application Components** (via Realizations: Full/Partial/Planned)
  └── Mapped to → **Value Stream Stages**
  └── Linked to → **Enterprise Capabilities** (cross-domain groupings)

**Application Components** (IT systems)
  └── Have → **Relations** to other applications (depends_on, uses, etc.)
  └── Have → **Fit Scores** per Strategy Pillar (1-5)
  └── Have → **TIME Classification** (Tolerate/Invest/Migrate/Eliminate)
  └── Have → **Component Origin** (Vendor/Acquired Entity/Internal Team)

**Strategy Pillars** (strategic themes)
  └── Drive → **Importance Scores** on capabilities (1-5, per domain)
  └── Drive → **Fit Scores** on applications (1-5)
  └── Gap = Importance - Fit → Aligned/Concern/Liability

**Value Streams**
  └── Contain → **Stages** (ordered sequence)
  └── Stages map to → **Capabilities**

**Business Domains** group L1 capabilities for organizational views.
**Enterprise Capabilities** group domain capabilities for cross-domain analysis.`,
}
