# Value Stream Future Enhancements

## Status
**pending**

---

## Context

These are potential enhancements to the core Value Streams feature (spec 130). Each item is independent and can be picked up as a separate vertical slice once the base feature is validated with users.f

---

## Enhancements

### Value Stream Maturity Scoring
Aggregate capability maturity levels into a value stream-level score. Each stage's maturity is derived from its mapped capabilities' maturity levels, giving a per-stage and overall value stream maturity view. Helps identify which stages are well-supported vs. immature.

### Value Stream Dependencies
Model dependencies between value streams (e.g., "Order-to-Cash" depends on "Procure-to-Pay"). Enables impact analysis across value streams.

### Strategic Importance Rating
Rate value streams against strategy pillars, mirroring the existing `StrategyImportance` pattern for capabilities. Enables strategic portfolio analysis at the value stream level.

### Business Domain Crossing Analysis
Analyze which business domains a value stream spans, derived from its mapped capabilities' domain assignments. Surfaces cross-domain flows and governance implications.

### Cross-Value-Stream Overlap Matrix
Dedicated heatmap/matrix view showing capabilities x value streams. Identifies the most cross-cutting capabilities and the most capability-intensive value streams.

### Stage Types / Phase Classification
Tag stages with a type such as "Triggering", "Value-Adding", "Enabling", or "Handoff". Adds analytical richness for Lean-style value stream analysis. Purely additive — an optional field on stages.

### Stage-Capability Contribution Levels
Add a contribution level ("Primary", "Supporting") to stage-capability mappings. Follows the existing `CapabilityRealization` pattern. Enables finer-grained analysis of which capabilities are critical vs. auxiliary for each stage.

### Value Stream Canvas Visualization
Place value streams on the existing architecture canvas alongside components and capabilities. Requires adapting the free-form canvas to support sequential flow elements — a significant design effort.

### Value Stream Templates
Pre-built templates for common value streams (e.g., "Hire-to-Retire", "Order-to-Cash", "Procure-to-Pay"). Reduces time-to-value for new users by providing starting points.

### Audit History UI
Surface event-sourced audit history for value streams in the UI. Follow the existing audit trail patterns.

### Stakeholder / Customer Journey Mapping
Extend value streams with an external stakeholder perspective, mapping stages to customer touchpoints. Bridges internal process flows with customer experience analysis.

### Time / Duration Modeling
Add estimated or measured duration to stages. Enables cycle time analysis and bottleneck identification. Closer to process mining than pure EA.

---

## Checklist
- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API Documentation updated in OpenAPI specification
- [ ] User sign-off
