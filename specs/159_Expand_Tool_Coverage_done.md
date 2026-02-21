# Expand Tool Coverage

**Status**: pending

**Series**: Architecture Assistant Evolution (5 of 6)
- Spec 155: Agent Permission Ceiling
- Spec 156: Generic Tool Executor
- Spec 157: Tool Catalog per Bounded Context
- Spec 158: Domain Knowledge Injection
- **Spec 159**: Expand Tool Coverage (this spec)
- Spec 160: Agent Audit Events

**Depends on:** Spec 157, Spec 158

## User Value

> "As an enterprise architect, I want to ask the assistant about strategic fit analysis, enterprise capabilities, capability dependencies, and component origins — not just applications and basic capabilities — so it covers the full scope of my architecture work."

## Problem

Only 24 of ~80+ API endpoints are exposed as agent tools. Major gaps:
- Enterprise Architecture (enterprise capabilities, links, TIME classification)
- Strategy (importance ratings, fit scores, gap analysis)
- Capability dependencies
- Capability realizations (detailed CRUD)
- MetaModel (strategy pillars, maturity scale — read-only)
- Component origins (vendors, acquired entities, internal teams)

## Solution

Add `AgentToolSpec` entries to each context's `publishedlanguage/agent_tools.go`. Each tool below includes a domain-aware description per spec 158.
The Architecture test TestToolCatalog_InfoUnexposedRoutes will output which routes are missing tool specs.
When this spec is implemented, TestToolCatalog_InfoUnexposedRoutes must fail if there's routes without tool specs.  
This requires that those excluded on purpose are explicitly excluded.

### Enterprise Architecture Context (new)

| Tool | Access | API |
|---|---|---|
| `list_enterprise_capabilities` | Read | `GET /enterprise-capabilities` |
| `get_enterprise_capability_details` | Read | `GET /enterprise-capabilities/{id}` |
| `create_enterprise_capability` | Create | `POST /enterprise-capabilities` |
| `update_enterprise_capability` | Update | `PUT /enterprise-capabilities/{id}` |
| `delete_enterprise_capability` | Delete | `DELETE /enterprise-capabilities/{id}` |
| `link_capability_to_enterprise` | Create | `POST /enterprise-capabilities/{id}/links` |
| `unlink_capability_from_enterprise` | Delete | `DELETE /enterprise-capability-links/{id}` |
| `get_enterprise_strategic_importance` | Read | `GET /enterprise-capabilities/{id}/importance` |
| `set_enterprise_strategic_importance` | Create | `POST /enterprise-capabilities/{id}/importance` |
| `get_time_suggestions` | Read | `GET /enterprise-capabilities/{id}/time-suggestions` |

### Capability Mapping Context (additions)

| Tool | Access | API |
|---|---|---|
| `list_capability_dependencies` | Read | `GET /capability-dependencies` |
| `create_capability_dependency` | Create | `POST /capability-dependencies` |
| `delete_capability_dependency` | Delete | `DELETE /capability-dependencies/{id}` |
| `get_capability_children` | Read | `GET /capabilities/{id}/children` |
| `get_strategy_importance` | Read | `GET /capabilities/{id}/importance` |
| `set_strategy_importance` | Create | `PUT /strategy-importance` |
| `get_application_fit_scores` | Read | `GET /components/{id}/fit-scores` |
| `set_application_fit_score` | Create | `PUT /components/{id}/fit-scores/{pillarId}` |
| `get_strategic_fit_analysis` | Read | `GET /strategic-fit-analysis` |

### MetaModel Context (read-only)

| Tool | Access | API |
|---|---|---|
| `get_strategy_pillars` | Read | `GET /strategy-pillars` |
| `get_maturity_scale` | Read | `GET /maturity-scale` |

No write tools. MetaModel writes are permanently excluded by the agent permission ceiling (spec 155).

### Architecture Modeling Context (additions)

| Tool | Access | API |
|---|---|---|
| `list_vendors` | Read | `GET /vendors` |
| `get_vendor_details` | Read | `GET /vendors/{id}` |
| `list_acquired_entities` | Read | `GET /acquired-entities` |
| `get_acquired_entity_details` | Read | `GET /acquired-entities/{id}` |
| `list_internal_teams` | Read | `GET /internal-teams` |
| `get_internal_team_details` | Read | `GET /internal-teams/{id}` |
| `get_component_origin` | Read | `GET /components/{id}/origin` |

### What Stays Excluded

- Users, invitations, audit — permanently blocked by permission ceiling
- Access delegation — permanently blocked
- MetaModel writes — permanently blocked
- Architecture views/layouts — not useful in conversation (visual artifacts)
- Import endpoints — bulk operations, dangerous in agent context
- Releases — not relevant to architecture exploration

## Checklist

- [x] Specification approved
- [x] Enterprise Architecture tools added (10 specs)
- [x] Capability Mapping tools added (9 specs)
- [x] MetaModel read-only tools added (2 specs)
- [x] Architecture Modeling origin tools added (7 specs)
- [x] All descriptions include domain context per spec 158
- [x] Architecture guard tests pass (valid routes, valid permissions)
- [x] Manual testing: agent can answer enterprise architecture questions
- [x] Manual testing: agent can answer strategy/fit questions
- [x] Build passing
