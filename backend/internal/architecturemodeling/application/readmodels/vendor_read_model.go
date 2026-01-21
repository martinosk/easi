package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type VendorDTO struct {
	ID                    string      `json:"id"`
	Name                  string      `json:"name"`
	ImplementationPartner string      `json:"implementationPartner,omitempty"`
	Notes                 string      `json:"notes,omitempty"`
	CreatedAt             time.Time   `json:"createdAt"`
	Links                 types.Links `json:"_links,omitempty"`
}

type VendorReadModel struct {
	db *database.TenantAwareDB
}

func NewVendorReadModel(db *database.TenantAwareDB) *VendorReadModel {
	return &VendorReadModel{db: db}
}

func (rm *VendorReadModel) Insert(ctx context.Context, dto VendorDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO vendors (id, tenant_id, name, implementation_partner, notes, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		dto.ID, tenantID.Value(), dto.Name, dto.ImplementationPartner, dto.Notes, dto.CreatedAt,
	)
	return err
}

func (rm *VendorReadModel) Update(ctx context.Context, id, name, implementationPartner, notes string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE vendors SET name = $1, implementation_partner = $2, notes = $3, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $4 AND id = $5",
		name, implementationPartner, notes, tenantID.Value(), id,
	)
	return err
}

func (rm *VendorReadModel) MarkAsDeleted(ctx context.Context, id string, deletedAt time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE vendors SET is_deleted = TRUE, deleted_at = $1 WHERE tenant_id = $2 AND id = $3",
		deletedAt, tenantID.Value(), id,
	)
	return err
}

func (rm *VendorReadModel) GetByID(ctx context.Context, id string) (*VendorDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto VendorDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			"SELECT id, name, implementation_partner, notes, created_at FROM vendors WHERE tenant_id = $1 AND id = $2 AND is_deleted = FALSE",
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.Name, &dto.ImplementationPartner, &dto.Notes, &dto.CreatedAt)

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

func (rm *VendorReadModel) GetAll(ctx context.Context) ([]VendorDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	vendors := make([]VendorDTO, 0)
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, name, implementation_partner, notes, created_at FROM vendors WHERE tenant_id = $1 AND is_deleted = FALSE ORDER BY LOWER(name) ASC",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto VendorDTO
			if err := rows.Scan(&dto.ID, &dto.Name, &dto.ImplementationPartner, &dto.Notes, &dto.CreatedAt); err != nil {
				return err
			}
			vendors = append(vendors, dto)
		}

		return rows.Err()
	})

	return vendors, err
}

func (rm *VendorReadModel) GetAllPaginated(ctx context.Context, limit int, afterCursor string, afterName string) ([]VendorDTO, bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, false, err
	}

	queryLimit := limit + 1
	vendors := make([]VendorDTO, 0)

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		var rows *sql.Rows
		var err error

		if afterCursor == "" {
			rows, err = tx.QueryContext(ctx,
				"SELECT id, name, implementation_partner, notes, created_at FROM vendors WHERE tenant_id = $1 AND is_deleted = FALSE ORDER BY LOWER(name) ASC, id ASC LIMIT $2",
				tenantID.Value(), queryLimit,
			)
		} else {
			rows, err = tx.QueryContext(ctx,
				"SELECT id, name, implementation_partner, notes, created_at FROM vendors WHERE tenant_id = $1 AND is_deleted = FALSE AND (LOWER(name) > LOWER($2) OR (LOWER(name) = LOWER($2) AND id > $3)) ORDER BY LOWER(name) ASC, id ASC LIMIT $4",
				tenantID.Value(), afterName, afterCursor, queryLimit,
			)
		}
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto VendorDTO
			if err := rows.Scan(&dto.ID, &dto.Name, &dto.ImplementationPartner, &dto.Notes, &dto.CreatedAt); err != nil {
				return err
			}
			vendors = append(vendors, dto)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, false, err
	}

	hasMore := len(vendors) > limit
	if hasMore {
		vendors = vendors[:limit]
	}

	return vendors, hasMore, nil
}
