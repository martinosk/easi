package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type BuiltByRelationshipDTO struct {
	ID               string      `json:"id"`
	InternalTeamID   string      `json:"internalTeamId"`
	InternalTeamName string      `json:"internalTeamName,omitempty"`
	ComponentID      string      `json:"componentId"`
	ComponentName    string      `json:"componentName,omitempty"`
	Notes            string      `json:"notes,omitempty"`
	CreatedAt        time.Time   `json:"createdAt"`
	Links            types.Links `json:"_links,omitempty"`
}

type BuiltByRelationshipReadModel struct {
	db *database.TenantAwareDB
}

func NewBuiltByRelationshipReadModel(db *database.TenantAwareDB) *BuiltByRelationshipReadModel {
	return &BuiltByRelationshipReadModel{db: db}
}

func (rm *BuiltByRelationshipReadModel) Upsert(ctx context.Context, dto BuiltByRelationshipDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO architecturemodeling.built_by_relationships (id, tenant_id, internal_team_id, component_id, notes, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (tenant_id, component_id) WHERE is_deleted = FALSE
		 DO UPDATE SET internal_team_id = $3, notes = $5, created_at = $6, is_deleted = FALSE, deleted_at = NULL`,
		dto.ID, tenantID.Value(), dto.InternalTeamID, dto.ComponentID, dto.Notes, dto.CreatedAt,
	)
	return err
}

func (rm *BuiltByRelationshipReadModel) UpdateNotesByComponentID(ctx context.Context, componentID string, notes string) error {
	return rm.execWithTenant(ctx,
		`UPDATE architecturemodeling.built_by_relationships SET notes = $1 WHERE tenant_id = $2 AND component_id = $3`,
		func(tenantID string) []interface{} { return []interface{}{notes, tenantID, componentID} },
	)
}

func (rm *BuiltByRelationshipReadModel) DeleteByComponentID(ctx context.Context, componentID string) error {
	return rm.execWithTenant(ctx,
		`DELETE FROM architecturemodeling.built_by_relationships WHERE tenant_id = $1 AND component_id = $2`,
		func(tenantID string) []interface{} { return []interface{}{tenantID, componentID} },
	)
}

func (rm *BuiltByRelationshipReadModel) execWithTenant(ctx context.Context, query string, argsFn func(tenantID string) []interface{}) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, argsFn(tenantID.Value())...)
	return err
}

func (rm *BuiltByRelationshipReadModel) MarkAsDeleted(ctx context.Context, id string, deletedAt time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE architecturemodeling.built_by_relationships SET is_deleted = TRUE, deleted_at = $1 WHERE tenant_id = $2 AND id = $3",
		deletedAt, tenantID.Value(), id,
	)
	return err
}

func (rm *BuiltByRelationshipReadModel) GetByID(ctx context.Context, id string) (*BuiltByRelationshipDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto BuiltByRelationshipDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT r.id, r.internal_team_id, COALESCE(t.name, '') as internal_team_name,
			        r.component_id, COALESCE(c.name, '') as component_name, r.notes, r.created_at
			 FROM architecturemodeling.built_by_relationships r
			 LEFT JOIN architecturemodeling.internal_teams t ON t.tenant_id = r.tenant_id AND t.id = r.internal_team_id AND t.is_deleted = FALSE
			 LEFT JOIN architecturemodeling.application_components c ON c.tenant_id = r.tenant_id AND c.id = r.component_id AND c.is_deleted = FALSE
			 WHERE r.tenant_id = $1 AND r.id = $2 AND r.is_deleted = FALSE`,
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.InternalTeamID, &dto.InternalTeamName, &dto.ComponentID, &dto.ComponentName, &dto.Notes, &dto.CreatedAt)

		if err == sql.ErrNoRows {
			notFound = true
			return nil
		}
		return err
	})

	if err != nil {
		return nil, err
	}
	if notFound {
		return nil, nil
	}

	return &dto, nil
}

func (rm *BuiltByRelationshipReadModel) GetByTeamID(ctx context.Context, teamID string) ([]BuiltByRelationshipDTO, error) {
	return rm.queryList(ctx,
		`WHERE r.tenant_id = $1 AND r.internal_team_id = $2 AND r.is_deleted = FALSE`,
		func(tenantID string) []interface{} { return []interface{}{tenantID, teamID} },
	)
}

func (rm *BuiltByRelationshipReadModel) GetByComponentID(ctx context.Context, componentID string) ([]BuiltByRelationshipDTO, error) {
	return rm.queryList(ctx,
		`WHERE r.tenant_id = $1 AND r.component_id = $2 AND r.is_deleted = FALSE`,
		func(tenantID string) []interface{} { return []interface{}{tenantID, componentID} },
	)
}

func (rm *BuiltByRelationshipReadModel) GetAll(ctx context.Context) ([]BuiltByRelationshipDTO, error) {
	return rm.queryList(ctx,
		`WHERE r.tenant_id = $1 AND r.is_deleted = FALSE`,
		func(tenantID string) []interface{} { return []interface{}{tenantID} },
	)
}

func (rm *BuiltByRelationshipReadModel) queryList(ctx context.Context, whereClause string, argsFn func(tenantID string) []interface{}) ([]BuiltByRelationshipDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	query := `SELECT r.id, r.internal_team_id, COALESCE(t.name, '') as internal_team_name,
		        r.component_id, COALESCE(c.name, '') as component_name, r.notes, r.created_at
		 FROM architecturemodeling.built_by_relationships r
		 LEFT JOIN architecturemodeling.internal_teams t ON t.tenant_id = r.tenant_id AND t.id = r.internal_team_id AND t.is_deleted = FALSE
		 LEFT JOIN architecturemodeling.application_components c ON c.tenant_id = r.tenant_id AND c.id = r.component_id AND c.is_deleted = FALSE
		 ` + whereClause

	relationships := make([]BuiltByRelationshipDTO, 0)
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, argsFn(tenantID.Value())...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto BuiltByRelationshipDTO
			if err := rows.Scan(&dto.ID, &dto.InternalTeamID, &dto.InternalTeamName, &dto.ComponentID, &dto.ComponentName, &dto.Notes, &dto.CreatedAt); err != nil {
				return err
			}
			relationships = append(relationships, dto)
		}
		return rows.Err()
	})

	return relationships, err
}
