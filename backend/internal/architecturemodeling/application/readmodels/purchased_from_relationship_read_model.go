package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type PurchasedFromRelationshipDTO struct {
	ID          string      `json:"id"`
	VendorID    string      `json:"vendorId"`
	ComponentID string      `json:"componentId"`
	Notes       string      `json:"notes,omitempty"`
	CreatedAt   time.Time   `json:"createdAt"`
	Links       types.Links `json:"_links,omitempty"`
}

type PurchasedFromRelationshipReadModel struct {
	db *database.TenantAwareDB
}

func NewPurchasedFromRelationshipReadModel(db *database.TenantAwareDB) *PurchasedFromRelationshipReadModel {
	return &PurchasedFromRelationshipReadModel{db: db}
}

func (rm *PurchasedFromRelationshipReadModel) Insert(ctx context.Context, dto PurchasedFromRelationshipDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO purchased_from_relationships (id, tenant_id, vendor_id, component_id, notes, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		dto.ID, tenantID.Value(), dto.VendorID, dto.ComponentID, dto.Notes, dto.CreatedAt,
	)
	return err
}

func (rm *PurchasedFromRelationshipReadModel) MarkAsDeleted(ctx context.Context, id string, deletedAt time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE purchased_from_relationships SET is_deleted = TRUE, deleted_at = $1 WHERE tenant_id = $2 AND id = $3",
		deletedAt, tenantID.Value(), id,
	)
	return err
}

func (rm *PurchasedFromRelationshipReadModel) GetByID(ctx context.Context, id string) (*PurchasedFromRelationshipDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto PurchasedFromRelationshipDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			"SELECT id, vendor_id, component_id, notes, created_at FROM purchased_from_relationships WHERE tenant_id = $1 AND id = $2 AND is_deleted = FALSE",
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.VendorID, &dto.ComponentID, &dto.Notes, &dto.CreatedAt)

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

func (rm *PurchasedFromRelationshipReadModel) GetByVendorID(ctx context.Context, vendorID string) ([]PurchasedFromRelationshipDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	relationships := make([]PurchasedFromRelationshipDTO, 0)
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, vendor_id, component_id, notes, created_at FROM purchased_from_relationships WHERE tenant_id = $1 AND vendor_id = $2 AND is_deleted = FALSE",
			tenantID.Value(), vendorID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto PurchasedFromRelationshipDTO
			if err := rows.Scan(&dto.ID, &dto.VendorID, &dto.ComponentID, &dto.Notes, &dto.CreatedAt); err != nil {
				return err
			}
			relationships = append(relationships, dto)
		}

		return rows.Err()
	})

	return relationships, err
}

func (rm *PurchasedFromRelationshipReadModel) GetByComponentID(ctx context.Context, componentID string) ([]PurchasedFromRelationshipDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	relationships := make([]PurchasedFromRelationshipDTO, 0)
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, vendor_id, component_id, notes, created_at FROM purchased_from_relationships WHERE tenant_id = $1 AND component_id = $2 AND is_deleted = FALSE",
			tenantID.Value(), componentID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto PurchasedFromRelationshipDTO
			if err := rows.Scan(&dto.ID, &dto.VendorID, &dto.ComponentID, &dto.Notes, &dto.CreatedAt); err != nil {
				return err
			}
			relationships = append(relationships, dto)
		}

		return rows.Err()
	})

	return relationships, err
}
