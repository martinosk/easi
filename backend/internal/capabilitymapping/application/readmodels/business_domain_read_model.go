package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type BusinessDomainDTO struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	Description     string      `json:"description,omitempty"`
	CapabilityCount int         `json:"capabilityCount"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       *time.Time  `json:"updatedAt,omitempty"`
	Links           types.Links `json:"_links,omitempty"`
}

type BusinessDomainReadModel struct {
	db *database.TenantAwareDB
}

func NewBusinessDomainReadModel(db *database.TenantAwareDB) *BusinessDomainReadModel {
	return &BusinessDomainReadModel{db: db}
}

func (rm *BusinessDomainReadModel) Insert(ctx context.Context, dto BusinessDomainDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO business_domains (id, tenant_id, name, description, capability_count, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		dto.ID, tenantID.Value(), dto.Name, dto.Description, 0, dto.CreatedAt,
	)
	return err
}

func (rm *BusinessDomainReadModel) Update(ctx context.Context, id, name, description string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE business_domains SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $3 AND id = $4",
		name, description, tenantID.Value(), id,
	)
	return err
}

func (rm *BusinessDomainReadModel) Delete(ctx context.Context, id string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM business_domains WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), id,
	)
	return err
}

func (rm *BusinessDomainReadModel) IncrementCapabilityCount(ctx context.Context, id string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE business_domains SET capability_count = capability_count + 1 WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), id,
	)
	return err
}

func (rm *BusinessDomainReadModel) DecrementCapabilityCount(ctx context.Context, id string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE business_domains SET capability_count = GREATEST(0, capability_count - 1) WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), id,
	)
	return err
}

func (rm *BusinessDomainReadModel) GetAll(ctx context.Context) ([]BusinessDomainDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var domains []BusinessDomainDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, name, description, capability_count, created_at, updated_at FROM business_domains WHERE tenant_id = $1 ORDER BY name",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto BusinessDomainDTO
			var updatedAt sql.NullTime
			if err := rows.Scan(&dto.ID, &dto.Name, &dto.Description, &dto.CapabilityCount, &dto.CreatedAt, &updatedAt); err != nil {
				return err
			}
			if updatedAt.Valid {
				dto.UpdatedAt = &updatedAt.Time
			}
			domains = append(domains, dto)
		}

		return rows.Err()
	})

	return domains, err
}

func (rm *BusinessDomainReadModel) GetByID(ctx context.Context, id string) (*BusinessDomainDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto BusinessDomainDTO
	var updatedAt sql.NullTime
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			"SELECT id, name, description, capability_count, created_at, updated_at FROM business_domains WHERE tenant_id = $1 AND id = $2",
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.Name, &dto.Description, &dto.CapabilityCount, &dto.CreatedAt, &updatedAt)

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

	if updatedAt.Valid {
		dto.UpdatedAt = &updatedAt.Time
	}

	return &dto, nil
}

func (rm *BusinessDomainReadModel) GetByName(ctx context.Context, name string) (*BusinessDomainDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto BusinessDomainDTO
	var updatedAt sql.NullTime
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			"SELECT id, name, description, capability_count, created_at, updated_at FROM business_domains WHERE tenant_id = $1 AND name = $2",
			tenantID.Value(), name,
		).Scan(&dto.ID, &dto.Name, &dto.Description, &dto.CapabilityCount, &dto.CreatedAt, &updatedAt)

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

	if updatedAt.Valid {
		dto.UpdatedAt = &updatedAt.Time
	}

	return &dto, nil
}

func (rm *BusinessDomainReadModel) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return false, err
	}

	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		if excludeID != "" {
			return tx.QueryRowContext(ctx,
				"SELECT COUNT(*) FROM business_domains WHERE tenant_id = $1 AND name = $2 AND id != $3",
				tenantID.Value(), name, excludeID,
			).Scan(&count)
		}
		return tx.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM business_domains WHERE tenant_id = $1 AND name = $2",
			tenantID.Value(), name,
		).Scan(&count)
	})

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
