package readmodels

import (
	"context"
	"database/sql"
)

type ComponentID string
type BusinessDomainID string

type ReferenceEntity string

const (
	ReferenceEntityCapability     ReferenceEntity = "capability"
	ReferenceEntityApplication    ReferenceEntity = "application"
	ReferenceEntityBusinessDomain ReferenceEntity = "business_domain"
)

func (rm *CapabilityJourneyReadModel) CacheReferenceName(ctx context.Context, entity ReferenceEntity, entityID, name string) error {
	return rm.tenantExec(ctx,
		`INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (tenant_id, entity_type, entity_id) DO UPDATE SET name = EXCLUDED.name`,
		func(t string) []any { return []any{t, string(entity), entityID, name} },
	)
}

func (rm *CapabilityJourneyReadModel) UpdateCapabilityName(ctx context.Context, capabilityID CapabilityID, name string) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.capability_journeys
		 SET capability_name = CASE WHEN capability_id = $1 THEN $2 ELSE capability_name END,
		     capability_stale = CASE WHEN capability_id = $1 THEN FALSE ELSE capability_stale END,
		     target_parent_name = CASE WHEN target_parent_id = $1 THEN $2 ELSE target_parent_name END,
		     target_parent_stale = CASE WHEN target_parent_id = $1 THEN FALSE ELSE target_parent_stale END
		 WHERE tenant_id = $3 AND (capability_id = $1 OR target_parent_id = $1)`,
		func(t string) []any { return []any{string(capabilityID), name, t} },
	)
}

func (rm *CapabilityJourneyReadModel) MarkCapabilityStale(ctx context.Context, capabilityID CapabilityID) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.capability_journeys
		 SET capability_stale = (capability_id = $1) OR capability_stale,
		     target_parent_stale = (target_parent_id = $1) OR target_parent_stale
		 WHERE tenant_id = $2 AND (capability_id = $1 OR target_parent_id = $1)`,
		func(t string) []any { return []any{string(capabilityID), t} },
	)
}

type componentReferenceQueries struct {
	journeys string
	sources  string
}

var componentNameQueries = componentReferenceQueries{
	journeys: `UPDATE architecturedirection.capability_journeys
	           SET to_component_name = $1, to_component_stale = FALSE
	           WHERE tenant_id = $2 AND to_component_id = $3`,
	sources: `UPDATE architecturedirection.capability_journey_sources
	          SET component_name = $1, component_stale = FALSE
	          WHERE tenant_id = $2 AND component_id = $3`,
}

var componentStaleQueries = componentReferenceQueries{
	journeys: `UPDATE architecturedirection.capability_journeys SET to_component_stale = TRUE
	           WHERE tenant_id = $1 AND to_component_id = $2`,
	sources: `UPDATE architecturedirection.capability_journey_sources SET component_stale = TRUE
	          WHERE tenant_id = $1 AND component_id = $2`,
}

func (rm *CapabilityJourneyReadModel) UpdateComponentName(ctx context.Context, componentID ComponentID, name string) error {
	return rm.execOnComponentTables(ctx, componentNameQueries,
		func(t string) []any { return []any{name, t, string(componentID)} },
	)
}

func (rm *CapabilityJourneyReadModel) MarkComponentStale(ctx context.Context, componentID ComponentID) error {
	return rm.execOnComponentTables(ctx, componentStaleQueries,
		func(t string) []any { return []any{t, string(componentID)} },
	)
}

func (rm *CapabilityJourneyReadModel) execOnComponentTables(ctx context.Context, queries componentReferenceQueries, args func(tenantID string) []any) error {
	return rm.tenantTx(ctx, func(tx *sql.Tx, tenantID string) error {
		if _, err := tx.ExecContext(ctx, queries.journeys, args(tenantID)...); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, queries.sources, args(tenantID)...)
		return err
	})
}

func (rm *CapabilityJourneyReadModel) UpdateDomainName(ctx context.Context, domainID BusinessDomainID, name string) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.capability_journeys SET target_domain_name = $1, target_domain_stale = FALSE
		 WHERE tenant_id = $2 AND target_domain_id = $3`,
		func(t string) []any { return []any{name, t, string(domainID)} },
	)
}

func (rm *CapabilityJourneyReadModel) MarkDomainStale(ctx context.Context, domainID BusinessDomainID) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.capability_journeys SET target_domain_stale = TRUE
		 WHERE tenant_id = $1 AND target_domain_id = $2`,
		func(t string) []any { return []any{t, string(domainID)} },
	)
}

func (rm *CapabilityJourneyReadModel) UpdatePlannedByName(ctx context.Context, email, name string) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.capability_journeys SET planned_by_name = $1
		 WHERE tenant_id = $2 AND planned_by = $3`,
		func(t string) []any { return []any{name, t, email} },
	)
}
