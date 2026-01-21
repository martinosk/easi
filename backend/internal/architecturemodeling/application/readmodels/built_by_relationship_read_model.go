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
	ID             string      `json:"id"`
	InternalTeamID string      `json:"internalTeamId"`
	ComponentID    string      `json:"componentId"`
	Notes          string      `json:"notes,omitempty"`
	CreatedAt      time.Time   `json:"createdAt"`
	Links          types.Links `json:"_links,omitempty"`
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
			"SELECT id, internal_team_id, component_id, notes, created_at FROM built_by_relationships WHERE tenant_id = $1 AND id = $2 AND is_deleted = FALSE",
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.InternalTeamID, &dto.ComponentID, &dto.Notes, &dto.CreatedAt)

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
			"SELECT id, internal_team_id, component_id, notes, created_at FROM built_by_relationships WHERE tenant_id = $1 AND internal_team_id = $2 AND is_deleted = FALSE",
			tenantID.Value(), teamID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto BuiltByRelationshipDTO
			if err := rows.Scan(&dto.ID, &dto.InternalTeamID, &dto.ComponentID, &dto.Notes, &dto.CreatedAt); err != nil {
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
			"SELECT id, internal_team_id, component_id, notes, created_at FROM built_by_relationships WHERE tenant_id = $1 AND component_id = $2 AND is_deleted = FALSE",
			tenantID.Value(), componentID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto BuiltByRelationshipDTO
			if err := rows.Scan(&dto.ID, &dto.InternalTeamID, &dto.ComponentID, &dto.Notes, &dto.CreatedAt); err != nil {
				return err
			}
			relationships = append(relationships, dto)
		}

		return rows.Err()
	})

	return relationships, err
}
