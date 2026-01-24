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

func (rm *BuiltByRelationshipReadModel) Insert(ctx context.Context, dto BuiltByRelationshipDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO built_by_relationships (id, tenant_id, internal_team_id, component_id, notes, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		dto.ID, tenantID.Value(), dto.InternalTeamID, dto.ComponentID, dto.Notes, dto.CreatedAt,
	)
	return err
}

func (rm *BuiltByRelationshipReadModel) UpdateByComponentID(ctx context.Context, dto BuiltByRelationshipDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE built_by_relationships
		 SET internal_team_id = $1, notes = $2, created_at = $3, is_deleted = FALSE, deleted_at = NULL
		 WHERE tenant_id = $4 AND component_id = $5`,
		dto.InternalTeamID, dto.Notes, dto.CreatedAt, tenantID.Value(), dto.ComponentID,
	)
	return err
}

func (rm *BuiltByRelationshipReadModel) UpdateNotesByComponentID(ctx context.Context, componentID string, notes string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE built_by_relationships SET notes = $1 WHERE tenant_id = $2 AND component_id = $3`,
		notes, tenantID.Value(), componentID,
	)
	return err
}

func (rm *BuiltByRelationshipReadModel) DeleteByComponentID(ctx context.Context, componentID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM built_by_relationships WHERE tenant_id = $1 AND component_id = $2",
		tenantID.Value(), componentID,
	)
	return err
}

func (rm *BuiltByRelationshipReadModel) MarkAsDeleted(ctx context.Context, id string, deletedAt time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE built_by_relationships SET is_deleted = TRUE, deleted_at = $1 WHERE tenant_id = $2 AND id = $3",
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
			 FROM built_by_relationships r
			 LEFT JOIN internal_teams t ON t.tenant_id = r.tenant_id AND t.id = r.internal_team_id AND t.is_deleted = FALSE
			 LEFT JOIN application_components c ON c.tenant_id = r.tenant_id AND c.id = r.component_id AND c.is_deleted = FALSE
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
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	relationships := make([]BuiltByRelationshipDTO, 0)
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT r.id, r.internal_team_id, COALESCE(t.name, '') as internal_team_name,
			        r.component_id, COALESCE(c.name, '') as component_name, r.notes, r.created_at
			 FROM built_by_relationships r
			 LEFT JOIN internal_teams t ON t.tenant_id = r.tenant_id AND t.id = r.internal_team_id AND t.is_deleted = FALSE
			 LEFT JOIN application_components c ON c.tenant_id = r.tenant_id AND c.id = r.component_id AND c.is_deleted = FALSE
			 WHERE r.tenant_id = $1 AND r.internal_team_id = $2 AND r.is_deleted = FALSE`,
			tenantID.Value(), teamID,
		)
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

func (rm *BuiltByRelationshipReadModel) GetByComponentID(ctx context.Context, componentID string) ([]BuiltByRelationshipDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	relationships := make([]BuiltByRelationshipDTO, 0)
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT r.id, r.internal_team_id, COALESCE(t.name, '') as internal_team_name,
			        r.component_id, COALESCE(c.name, '') as component_name, r.notes, r.created_at
			 FROM built_by_relationships r
			 LEFT JOIN internal_teams t ON t.tenant_id = r.tenant_id AND t.id = r.internal_team_id AND t.is_deleted = FALSE
			 LEFT JOIN application_components c ON c.tenant_id = r.tenant_id AND c.id = r.component_id AND c.is_deleted = FALSE
			 WHERE r.tenant_id = $1 AND r.component_id = $2 AND r.is_deleted = FALSE`,
			tenantID.Value(), componentID,
		)
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

func (rm *BuiltByRelationshipReadModel) GetAll(ctx context.Context) ([]BuiltByRelationshipDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	relationships := make([]BuiltByRelationshipDTO, 0)
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT r.id, r.internal_team_id, COALESCE(t.name, '') as internal_team_name,
			        r.component_id, COALESCE(c.name, '') as component_name, r.notes, r.created_at
			 FROM built_by_relationships r
			 LEFT JOIN internal_teams t ON t.tenant_id = r.tenant_id AND t.id = r.internal_team_id AND t.is_deleted = FALSE
			 LEFT JOIN application_components c ON c.tenant_id = r.tenant_id AND c.id = r.component_id AND c.is_deleted = FALSE
			 WHERE r.tenant_id = $1 AND r.is_deleted = FALSE`,
			tenantID.Value(),
		)
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
