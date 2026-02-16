package readmodels

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type BusinessDomainDTO struct {
	ID                string      `json:"id"`
	Name              string      `json:"name"`
	Description       string      `json:"description,omitempty"`
	DomainArchitectID string      `json:"domainArchitectId,omitempty"`
	CapabilityCount   int         `json:"capabilityCount"`
	CreatedAt         time.Time   `json:"createdAt"`
	UpdatedAt         *time.Time  `json:"updatedAt,omitempty"`
	Links             types.Links `json:"_links,omitempty"`
}

type BusinessDomainUpdate struct {
	Name              string
	Description       string
	DomainArchitectID string
}

type BusinessDomainReadModel struct {
	db *database.TenantAwareDB
}

func NewBusinessDomainReadModel(db *database.TenantAwareDB) *BusinessDomainReadModel {
	return &BusinessDomainReadModel{db: db}
}

func (rm *BusinessDomainReadModel) execTenantQuery(ctx context.Context, query string, args ...interface{}) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("resolve tenant for business domain mutation: %w", err)
	}
	_, err = rm.db.ExecContext(ctx, query, append([]interface{}{tenantID.Value()}, args...)...)
	if err != nil {
		return fmt.Errorf("execute business domain mutation for tenant %s: %w", tenantID.Value(), err)
	}
	return nil
}

func (rm *BusinessDomainReadModel) Insert(ctx context.Context, dto BusinessDomainDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("resolve tenant for insert business domain %s: %w", dto.ID, err)
	}

	var domainArchitectID interface{} = nil
	if dto.DomainArchitectID != "" {
		domainArchitectID = dto.DomainArchitectID
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM capabilitymapping.business_domains WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), dto.ID,
	)
	if err != nil {
		return fmt.Errorf("delete existing business domain %s before insert: %w", dto.ID, err)
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO capabilitymapping.business_domains
		(id, tenant_id, name, description, domain_architect_id, capability_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		dto.ID, tenantID.Value(), dto.Name, dto.Description, domainArchitectID, 0, dto.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert business domain %s for tenant %s: %w", dto.ID, tenantID.Value(), err)
	}
	return nil
}

func (rm *BusinessDomainReadModel) Update(ctx context.Context, id string, update BusinessDomainUpdate) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("resolve tenant for update business domain %s: %w", id, err)
	}

	var archID interface{} = nil
	if update.DomainArchitectID != "" {
		archID = update.DomainArchitectID
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE capabilitymapping.business_domains SET name = $1, description = $2, domain_architect_id = $3, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $4 AND id = $5",
		update.Name, update.Description, archID, tenantID.Value(), id,
	)
	if err != nil {
		return fmt.Errorf("update business domain %s for tenant %s: %w", id, tenantID.Value(), err)
	}
	return nil
}

func (rm *BusinessDomainReadModel) Delete(ctx context.Context, id string) error {
	return rm.execTenantQuery(ctx, "DELETE FROM capabilitymapping.business_domains WHERE tenant_id = $1 AND id = $2", id)
}

func (rm *BusinessDomainReadModel) IncrementCapabilityCount(ctx context.Context, id string) error {
	return rm.execTenantQuery(ctx, "UPDATE capabilitymapping.business_domains SET capability_count = capability_count + 1 WHERE tenant_id = $1 AND id = $2", id)
}

func (rm *BusinessDomainReadModel) DecrementCapabilityCount(ctx context.Context, id string) error {
	return rm.execTenantQuery(ctx, "UPDATE capabilitymapping.business_domains SET capability_count = GREATEST(0, capability_count - 1) WHERE tenant_id = $1 AND id = $2", id)
}

func (rm *BusinessDomainReadModel) GetAll(ctx context.Context) ([]BusinessDomainDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve tenant for list business domains: %w", err)
	}

	var domains []BusinessDomainDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, name, description, domain_architect_id, capability_count, created_at, updated_at FROM capabilitymapping.business_domains WHERE tenant_id = $1 ORDER BY name",
			tenantID.Value(),
		)
		if err != nil {
			return fmt.Errorf("query business domains for tenant %s: %w", tenantID.Value(), err)
		}
		defer rows.Close()

		for rows.Next() {
			dto, err := scanBusinessDomain(rows)
			if err != nil {
				return fmt.Errorf("scan business domain row for tenant %s: %w", tenantID.Value(), err)
			}
			domains = append(domains, *dto)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate business domain rows for tenant %s: %w", tenantID.Value(), err)
		}
		return nil
	})

	return domains, err
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanBusinessDomain(s scanner) (*BusinessDomainDTO, error) {
	var dto BusinessDomainDTO
	var updatedAt sql.NullTime
	var domainArchitectID sql.NullString

	if err := s.Scan(&dto.ID, &dto.Name, &dto.Description, &domainArchitectID, &dto.CapabilityCount, &dto.CreatedAt, &updatedAt); err != nil {
		return nil, err
	}

	if updatedAt.Valid {
		dto.UpdatedAt = &updatedAt.Time
	}
	if domainArchitectID.Valid {
		dto.DomainArchitectID = domainArchitectID.String
	}

	return &dto, nil
}

func (rm *BusinessDomainReadModel) getByCondition(ctx context.Context, whereClause string, arg interface{}) (*BusinessDomainDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve tenant for business domain lookup (%s): %w", whereClause, err)
	}

	var dto *BusinessDomainDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx,
			"SELECT id, name, description, domain_architect_id, capability_count, created_at, updated_at FROM capabilitymapping.business_domains WHERE tenant_id = $1 AND "+whereClause,
			tenantID.Value(), arg,
		)

		result, err := scanBusinessDomain(row)
		if err == sql.ErrNoRows {
			return nil
		}
		if err != nil {
			return fmt.Errorf("scan business domain by condition %s for tenant %s: %w", whereClause, tenantID.Value(), err)
		}
		dto = result
		return nil
	})

	return dto, err
}

func (rm *BusinessDomainReadModel) GetByID(ctx context.Context, id string) (*BusinessDomainDTO, error) {
	return rm.getByCondition(ctx, "id = $2", id)
}

func (rm *BusinessDomainReadModel) GetByName(ctx context.Context, name string) (*BusinessDomainDTO, error) {
	return rm.getByCondition(ctx, "name = $2", name)
}

func (rm *BusinessDomainReadModel) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return false, fmt.Errorf("resolve tenant for business domain name exists %s: %w", name, err)
	}

	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		if excludeID != "" {
			return tx.QueryRowContext(ctx,
				"SELECT COUNT(*) FROM capabilitymapping.business_domains WHERE tenant_id = $1 AND name = $2 AND id != $3",
				tenantID.Value(), name, excludeID,
			).Scan(&count)
		}
		return tx.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM capabilitymapping.business_domains WHERE tenant_id = $1 AND name = $2",
			tenantID.Value(), name,
		).Scan(&count)
	})

	if err != nil {
		return false, fmt.Errorf("query business domain name exists %s for tenant %s: %w", name, tenantID.Value(), err)
	}

	return count > 0, nil
}
