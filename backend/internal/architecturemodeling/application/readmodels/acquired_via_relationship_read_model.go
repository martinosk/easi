package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type AcquiredViaRelationshipDTO struct {
	ID               string      `json:"id"`
	AcquiredEntityID string      `json:"acquiredEntityId"`
	ComponentID      string      `json:"componentId"`
	Notes            string      `json:"notes,omitempty"`
	CreatedAt        time.Time   `json:"createdAt"`
	Links            types.Links `json:"_links,omitempty"`
}

type AcquiredViaRelationshipReadModel struct {
	db *database.TenantAwareDB
}

func NewAcquiredViaRelationshipReadModel(db *database.TenantAwareDB) *AcquiredViaRelationshipReadModel {
	return &AcquiredViaRelationshipReadModel{db: db}
}

func (rm *AcquiredViaRelationshipReadModel) Insert(ctx context.Context, dto AcquiredViaRelationshipDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO acquired_via_relationships (id, tenant_id, acquired_entity_id, component_id, notes, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		dto.ID, tenantID.Value(), dto.AcquiredEntityID, dto.ComponentID, dto.Notes, dto.CreatedAt,
	)
	return err
}

func (rm *AcquiredViaRelationshipReadModel) MarkAsDeleted(ctx context.Context, id string, deletedAt time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE acquired_via_relationships SET is_deleted = TRUE, deleted_at = $1 WHERE tenant_id = $2 AND id = $3",
		deletedAt, tenantID.Value(), id,
	)
	return err
}

func (rm *AcquiredViaRelationshipReadModel) GetByID(ctx context.Context, id string) (*AcquiredViaRelationshipDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto AcquiredViaRelationshipDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			"SELECT id, acquired_entity_id, component_id, notes, created_at FROM acquired_via_relationships WHERE tenant_id = $1 AND id = $2 AND is_deleted = FALSE",
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.AcquiredEntityID, &dto.ComponentID, &dto.Notes, &dto.CreatedAt)

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

func (rm *AcquiredViaRelationshipReadModel) GetByEntityID(ctx context.Context, entityID string) ([]AcquiredViaRelationshipDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var relationships []AcquiredViaRelationshipDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, acquired_entity_id, component_id, notes, created_at FROM acquired_via_relationships WHERE tenant_id = $1 AND acquired_entity_id = $2 AND is_deleted = FALSE",
			tenantID.Value(), entityID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto AcquiredViaRelationshipDTO
			if err := rows.Scan(&dto.ID, &dto.AcquiredEntityID, &dto.ComponentID, &dto.Notes, &dto.CreatedAt); err != nil {
				return err
			}
			relationships = append(relationships, dto)
		}

		return rows.Err()
	})

	return relationships, err
}

func (rm *AcquiredViaRelationshipReadModel) GetByComponentID(ctx context.Context, componentID string) ([]AcquiredViaRelationshipDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var relationships []AcquiredViaRelationshipDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, acquired_entity_id, component_id, notes, created_at FROM acquired_via_relationships WHERE tenant_id = $1 AND component_id = $2 AND is_deleted = FALSE",
			tenantID.Value(), componentID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto AcquiredViaRelationshipDTO
			if err := rows.Scan(&dto.ID, &dto.AcquiredEntityID, &dto.ComponentID, &dto.Notes, &dto.CreatedAt); err != nil {
				return err
			}
			relationships = append(relationships, dto)
		}

		return rows.Err()
	})

	return relationships, err
}
