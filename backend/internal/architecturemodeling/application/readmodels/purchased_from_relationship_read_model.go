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
	ID            string      `json:"id"`
	VendorID      string      `json:"vendorId"`
	VendorName    string      `json:"vendorName,omitempty"`
	ComponentID   string      `json:"componentId"`
	ComponentName string      `json:"componentName,omitempty"`
	Notes         string      `json:"notes,omitempty"`
	CreatedAt     time.Time   `json:"createdAt"`
	Links         types.Links `json:"_links,omitempty"`
}

type PurchasedFromRelationshipReadModel struct {
	db *database.TenantAwareDB
}

func NewPurchasedFromRelationshipReadModel(db *database.TenantAwareDB) *PurchasedFromRelationshipReadModel {
	return &PurchasedFromRelationshipReadModel{db: db}
}

func (rm *PurchasedFromRelationshipReadModel) Upsert(ctx context.Context, dto PurchasedFromRelationshipDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO purchased_from_relationships (id, tenant_id, vendor_id, component_id, notes, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (tenant_id, component_id) WHERE is_deleted = FALSE
		 DO UPDATE SET vendor_id = $3, notes = $5, created_at = $6, is_deleted = FALSE, deleted_at = NULL`,
		dto.ID, tenantID.Value(), dto.VendorID, dto.ComponentID, dto.Notes, dto.CreatedAt,
	)
	return err
}

func (rm *PurchasedFromRelationshipReadModel) UpdateNotesByComponentID(ctx context.Context, componentID string, notes string) error {
	return rm.execWithTenant(ctx,
		`UPDATE purchased_from_relationships SET notes = $1 WHERE tenant_id = $2 AND component_id = $3`,
		func(tenantID string) []interface{} { return []interface{}{notes, tenantID, componentID} },
	)
}

func (rm *PurchasedFromRelationshipReadModel) DeleteByComponentID(ctx context.Context, componentID string) error {
	return rm.execWithTenant(ctx,
		`DELETE FROM purchased_from_relationships WHERE tenant_id = $1 AND component_id = $2`,
		func(tenantID string) []interface{} { return []interface{}{tenantID, componentID} },
	)
}

func (rm *PurchasedFromRelationshipReadModel) execWithTenant(ctx context.Context, query string, argsFn func(tenantID string) []interface{}) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, argsFn(tenantID.Value())...)
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
			`SELECT r.id, r.vendor_id, COALESCE(v.name, '') as vendor_name,
			        r.component_id, COALESCE(c.name, '') as component_name, r.notes, r.created_at
			 FROM purchased_from_relationships r
			 LEFT JOIN vendors v ON v.tenant_id = r.tenant_id AND v.id = r.vendor_id AND v.is_deleted = FALSE
			 LEFT JOIN application_components c ON c.tenant_id = r.tenant_id AND c.id = r.component_id AND c.is_deleted = FALSE
			 WHERE r.tenant_id = $1 AND r.id = $2 AND r.is_deleted = FALSE`,
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.VendorID, &dto.VendorName, &dto.ComponentID, &dto.ComponentName, &dto.Notes, &dto.CreatedAt)

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
	return rm.queryList(ctx,
		`WHERE r.tenant_id = $1 AND r.vendor_id = $2 AND r.is_deleted = FALSE`,
		func(tenantID string) []interface{} { return []interface{}{tenantID, vendorID} },
	)
}

func (rm *PurchasedFromRelationshipReadModel) GetByComponentID(ctx context.Context, componentID string) ([]PurchasedFromRelationshipDTO, error) {
	return rm.queryList(ctx,
		`WHERE r.tenant_id = $1 AND r.component_id = $2 AND r.is_deleted = FALSE`,
		func(tenantID string) []interface{} { return []interface{}{tenantID, componentID} },
	)
}

func (rm *PurchasedFromRelationshipReadModel) GetAll(ctx context.Context) ([]PurchasedFromRelationshipDTO, error) {
	return rm.queryList(ctx,
		`WHERE r.tenant_id = $1 AND r.is_deleted = FALSE`,
		func(tenantID string) []interface{} { return []interface{}{tenantID} },
	)
}

func (rm *PurchasedFromRelationshipReadModel) queryList(ctx context.Context, whereClause string, argsFn func(tenantID string) []interface{}) ([]PurchasedFromRelationshipDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	query := `SELECT r.id, r.vendor_id, COALESCE(v.name, '') as vendor_name,
		        r.component_id, COALESCE(c.name, '') as component_name, r.notes, r.created_at
		 FROM purchased_from_relationships r
		 LEFT JOIN vendors v ON v.tenant_id = r.tenant_id AND v.id = r.vendor_id AND v.is_deleted = FALSE
		 LEFT JOIN application_components c ON c.tenant_id = r.tenant_id AND c.id = r.component_id AND c.is_deleted = FALSE
		 ` + whereClause

	relationships := make([]PurchasedFromRelationshipDTO, 0)
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, argsFn(tenantID.Value())...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto PurchasedFromRelationshipDTO
			if err := rows.Scan(&dto.ID, &dto.VendorID, &dto.VendorName, &dto.ComponentID, &dto.ComponentName, &dto.Notes, &dto.CreatedAt); err != nil {
				return err
			}
			relationships = append(relationships, dto)
		}
		return rows.Err()
	})

	return relationships, err
}
