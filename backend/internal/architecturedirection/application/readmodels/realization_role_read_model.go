package readmodels

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/shared/types"
)

var ErrRealizationRolesAggregateConflict = errors.New("a different realization roles aggregate is already registered for this capability")

type RealizationRoleDTO struct {
	CapabilityID   string      `json:"capabilityId"`
	CapabilityName string      `json:"capabilityName"`
	ComponentID    string      `json:"componentId"`
	ComponentName  string      `json:"componentName"`
	Role           string      `json:"role"`
	AssignedBy     string      `json:"assignedBy"`
	AssignedAt     time.Time   `json:"assignedAt"`
	Links          types.Links `json:"_links,omitempty"`
}

type UpsertRealizationRoleParams struct {
	CapabilityID         string
	ComponentID          string
	RealizationID        string
	Role                 string
	AssignedBy           string
	AssignedAt           time.Time
	AggregateID          string
	DisplacedComponentID string
}

type RealizationRoleReadModel struct {
	db *database.TenantAwareDB
}

func NewRealizationRoleReadModel(db *database.TenantAwareDB) *RealizationRoleReadModel {
	return &RealizationRoleReadModel{db: db}
}

func (rm *RealizationRoleReadModel) UpsertRole(ctx context.Context, p UpsertRealizationRoleParams) error {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return err
	}
	tx, err := rm.db.BeginTxWithTenant(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if p.DisplacedComponentID != "" {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM architecturedirection.realization_roles
			 WHERE tenant_id = $1 AND capability_id = $2 AND component_id = $3`,
			tenantID, p.CapabilityID, p.DisplacedComponentID,
		); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO architecturedirection.realization_roles
		 (tenant_id, capability_id, component_id, realization_id, role, assigned_by, assigned_at, capability_name, component_name, aggregate_id)
		 SELECT $1, $2, $3, $4, $5, $6, $7, cap.name, comp.name, $8
		 FROM (SELECT 1) AS stub
		 LEFT JOIN architecturedirection.reference_name_cache cap
		   ON cap.tenant_id = $1 AND cap.entity_type = 'capability' AND cap.entity_id = $2
		 LEFT JOIN architecturedirection.reference_name_cache comp
		   ON comp.tenant_id = $1 AND comp.entity_type = 'application' AND comp.entity_id = $3
		 ON CONFLICT (tenant_id, capability_id, component_id) DO UPDATE SET
		   realization_id = EXCLUDED.realization_id,
		   role = EXCLUDED.role,
		   assigned_by = EXCLUDED.assigned_by,
		   assigned_at = EXCLUDED.assigned_at,
		   capability_name = EXCLUDED.capability_name,
		   component_name = EXCLUDED.component_name,
		   aggregate_id = EXCLUDED.aggregate_id,
		   updated_at = CURRENT_TIMESTAMP`,
		tenantID, p.CapabilityID, p.ComponentID, p.RealizationID, p.Role, p.AssignedBy, p.AssignedAt, p.AggregateID,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (rm *RealizationRoleReadModel) DeleteRole(ctx context.Context, capabilityID, componentID string) error {
	return rm.tenantExec(ctx,
		`DELETE FROM architecturedirection.realization_roles WHERE tenant_id = $1 AND capability_id = $2 AND component_id = $3`,
		func(t string) []any { return []any{t, capabilityID, componentID} },
	)
}

func (rm *RealizationRoleReadModel) DeleteByCapabilityID(ctx context.Context, capabilityID string) error {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return err
	}
	tx, err := rm.db.BeginTxWithTenant(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM architecturedirection.realization_roles WHERE tenant_id = $1 AND capability_id = $2`,
		tenantID, capabilityID,
	); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM architecturedirection.realization_role_aggregates WHERE tenant_id = $1 AND capability_id = $2`,
		tenantID, capabilityID,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (rm *RealizationRoleReadModel) DeleteByComponentID(ctx context.Context, componentID string) error {
	return rm.tenantExec(ctx,
		`DELETE FROM architecturedirection.realization_roles WHERE tenant_id = $1 AND component_id = $2`,
		func(t string) []any { return []any{t, componentID} },
	)
}

func (rm *RealizationRoleReadModel) CacheCapabilityName(ctx context.Context, capabilityID, name string) error {
	return rm.cacheReferenceName(ctx, "capability", capabilityID, name)
}

func (rm *RealizationRoleReadModel) UpdateCapabilityName(ctx context.Context, capabilityID, name string) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.realization_roles SET capability_name = $1
		 WHERE tenant_id = $2 AND capability_id = $3`,
		func(t string) []any { return []any{name, t, capabilityID} },
	)
}

func (rm *RealizationRoleReadModel) CacheComponentName(ctx context.Context, componentID, name string) error {
	return rm.cacheReferenceName(ctx, "application", componentID, name)
}

func (rm *RealizationRoleReadModel) UpdateComponentName(ctx context.Context, componentID, name string) error {
	return rm.tenantExec(ctx,
		`UPDATE architecturedirection.realization_roles SET component_name = $1
		 WHERE tenant_id = $2 AND component_id = $3`,
		func(t string) []any { return []any{name, t, componentID} },
	)
}

func (rm *RealizationRoleReadModel) cacheReferenceName(ctx context.Context, entityType, entityID, name string) error {
	return rm.tenantExec(ctx,
		`INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (tenant_id, entity_type, entity_id) DO UPDATE SET name = EXCLUDED.name`,
		func(t string) []any { return []any{t, entityType, entityID, name} },
	)
}

func (rm *RealizationRoleReadModel) tenantExec(ctx context.Context, query string, argsFn func(tenantID string) []any) error {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, argsFn(tenantID)...)
	return err
}

func (rm *RealizationRoleReadModel) RegisterAggregate(ctx context.Context, capabilityID, aggregateID string) error {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return err
	}
	if _, err := rm.db.ExecContext(ctx,
		`INSERT INTO architecturedirection.realization_role_aggregates (tenant_id, capability_id, aggregate_id)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (tenant_id, capability_id) DO NOTHING`,
		tenantID, capabilityID, aggregateID,
	); err != nil {
		return err
	}
	registered, found, err := rm.FindAggregateIDForCapability(ctx, capabilityID)
	if err != nil {
		return err
	}
	if !found || registered != aggregateID {
		return ErrRealizationRolesAggregateConflict
	}
	return nil
}

func (rm *RealizationRoleReadModel) FindAggregateIDForCapability(ctx context.Context, capabilityID string) (string, bool, error) {
	var id string
	found, err := rm.findSingleRow(ctx,
		`SELECT aggregate_id FROM architecturedirection.realization_role_aggregates
		 WHERE tenant_id = $1 AND capability_id = $2`,
		capabilityID, &id)
	return id, found, err
}

func (rm *RealizationRoleReadModel) FindPairByRealizationID(ctx context.Context, realizationID string) (string, string, bool, error) {
	var capabilityID, componentID string
	found, err := rm.findSingleRow(ctx,
		`SELECT capability_id, component_id FROM architecturedirection.realization_roles
		 WHERE tenant_id = $1 AND realization_id = $2`,
		realizationID, &capabilityID, &componentID)
	return capabilityID, componentID, found, err
}

func (rm *RealizationRoleReadModel) findSingleRow(ctx context.Context, query, key string, dest ...any) (bool, error) {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return false, err
	}
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query, tenantID, key).Scan(dest...)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

const realizationRoleSelectColumns = `capability_id, COALESCE(capability_name, ''), component_id, COALESCE(component_name, ''),
	role, assigned_by, assigned_at`

func (rm *RealizationRoleReadModel) GetByPair(ctx context.Context, capabilityID, componentID string) (*RealizationRoleDTO, error) {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return nil, err
	}
	var dto *RealizationRoleDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx,
			`SELECT `+realizationRoleSelectColumns+`
			 FROM architecturedirection.realization_roles
			 WHERE tenant_id = $1 AND capability_id = $2 AND component_id = $3`,
			tenantID, capabilityID, componentID,
		)
		fetched, scanErr := scanRealizationRole(row)
		if errors.Is(scanErr, sql.ErrNoRows) {
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		dto = &fetched
		return nil
	})
	return dto, err
}

func (rm *RealizationRoleReadModel) GetByCapabilityIDs(ctx context.Context, capabilityIDs []string) ([]RealizationRoleDTO, error) {
	if len(capabilityIDs) == 0 {
		return []RealizationRoleDTO{}, nil
	}
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return nil, err
	}
	return rm.roles(ctx, "tenant_id = $1 AND capability_id = ANY($2)", tenantID, pq.Array(capabilityIDs))
}

func (rm *RealizationRoleReadModel) GetAll(ctx context.Context) ([]RealizationRoleDTO, error) {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return nil, err
	}
	return rm.roles(ctx, "tenant_id = $1", tenantID)
}

func (rm *RealizationRoleReadModel) roles(ctx context.Context, where string, args ...any) ([]RealizationRoleDTO, error) {
	roles := []RealizationRoleDTO{}
	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT `+realizationRoleSelectColumns+`
			 FROM architecturedirection.realization_roles
			 WHERE `+where+`
			 ORDER BY assigned_at DESC`,
			args...,
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			dto, scanErr := scanRealizationRole(rows)
			if scanErr != nil {
				return scanErr
			}
			roles = append(roles, dto)
		}
		return rows.Err()
	})
	return roles, err
}

type realizationRoleRowScanner interface {
	Scan(dest ...any) error
}

func scanRealizationRole(row realizationRoleRowScanner) (RealizationRoleDTO, error) {
	var dto RealizationRoleDTO
	err := row.Scan(&dto.CapabilityID, &dto.CapabilityName, &dto.ComponentID, &dto.ComponentName,
		&dto.Role, &dto.AssignedBy, &dto.AssignedAt)
	return dto, err
}
