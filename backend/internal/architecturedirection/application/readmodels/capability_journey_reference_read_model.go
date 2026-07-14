package readmodels

import (
	"context"
	"database/sql"
)

func (rm *CapabilityJourneyReadModel) CacheReferenceName(ctx context.Context, entityType, entityID, name string) error {
	return rm.tenantExec(ctx,
		`INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (tenant_id, entity_type, entity_id) DO UPDATE SET name = EXCLUDED.name`,
		func(t string) []any { return []any{t, entityType, entityID, name} },
	)
}

func (rm *CapabilityJourneyReadModel) UpdateCapabilityName(ctx context.Context, capabilityID, name string) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.capability_journeys
		 SET capability_name = CASE WHEN capability_id = $1 THEN $2 ELSE capability_name END,
		     capability_stale = CASE WHEN capability_id = $1 THEN FALSE ELSE capability_stale END,
		     target_parent_name = CASE WHEN target_parent_id = $1 THEN $2 ELSE target_parent_name END,
		     target_parent_stale = CASE WHEN target_parent_id = $1 THEN FALSE ELSE target_parent_stale END
		 WHERE tenant_id = $3 AND (capability_id = $1 OR target_parent_id = $1)`,
		func(t string) []any { return []any{capabilityID, name, t} },
	)
}

func (rm *CapabilityJourneyReadModel) MarkCapabilityStale(ctx context.Context, capabilityID string) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.capability_journeys
		 SET capability_stale = (capability_id = $1) OR capability_stale,
		     target_parent_stale = (target_parent_id = $1) OR target_parent_stale
		 WHERE tenant_id = $2 AND (capability_id = $1 OR target_parent_id = $1)`,
		func(t string) []any { return []any{capabilityID, t} },
	)
}

func (rm *CapabilityJourneyReadModel) UpdateComponentName(ctx context.Context, componentID, name string) error {
	return rm.execOnComponentTables(ctx,
		`UPDATE architecturedirection.capability_journeys
		 SET to_component_name = $1, to_component_stale = FALSE
		 WHERE tenant_id = $2 AND to_component_id = $3`,
		`UPDATE architecturedirection.capability_journey_sources
		 SET component_name = $1, component_stale = FALSE
		 WHERE tenant_id = $2 AND component_id = $3`,
		func(tenantID string) []any { return []any{name, tenantID, componentID} },
	)
}

func (rm *CapabilityJourneyReadModel) MarkComponentStale(ctx context.Context, componentID string) error {
	return rm.execOnComponentTables(ctx,
		`UPDATE architecturedirection.capability_journeys SET to_component_stale = TRUE
		 WHERE tenant_id = $1 AND to_component_id = $2`,
		`UPDATE architecturedirection.capability_journey_sources SET component_stale = TRUE
		 WHERE tenant_id = $1 AND component_id = $2`,
		func(tenantID string) []any { return []any{tenantID, componentID} },
	)
}

func (rm *CapabilityJourneyReadModel) execOnComponentTables(ctx context.Context, journeyQuery, sourceQuery string, args func(tenantID string) []any) error {
	return rm.tenantTx(ctx, func(tx *sql.Tx, tenantID string) error {
		if _, err := tx.ExecContext(ctx, journeyQuery, args(tenantID)...); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, sourceQuery, args(tenantID)...)
		return err
	})
}

func (rm *CapabilityJourneyReadModel) UpdateDomainName(ctx context.Context, domainID, name string) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.capability_journeys SET target_domain_name = $1, target_domain_stale = FALSE
		 WHERE tenant_id = $2 AND target_domain_id = $3`,
		func(t string) []any { return []any{name, t, domainID} },
	)
}

func (rm *CapabilityJourneyReadModel) MarkDomainStale(ctx context.Context, domainID string) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.capability_journeys SET target_domain_stale = TRUE
		 WHERE tenant_id = $1 AND target_domain_id = $2`,
		func(t string) []any { return []any{t, domainID} },
	)
}

func (rm *CapabilityJourneyReadModel) UpdatePlannedByName(ctx context.Context, email, name string) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.capability_journeys SET planned_by_name = $1
		 WHERE tenant_id = $2 AND planned_by = $3`,
		func(t string) []any { return []any{name, t, email} },
	)
}
