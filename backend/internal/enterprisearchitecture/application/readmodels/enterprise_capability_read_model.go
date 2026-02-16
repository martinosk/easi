package readmodels

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type EnterpriseCapabilityDTO struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Description    string      `json:"description,omitempty"`
	Category       string      `json:"category,omitempty"`
	Active         bool        `json:"active"`
	TargetMaturity *int        `json:"targetMaturity,omitempty"`
	LinkCount      int         `json:"linkCount"`
	DomainCount    int         `json:"domainCount"`
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      *time.Time  `json:"updatedAt,omitempty"`
	Links          types.Links `json:"_links,omitempty"`
}

type EnterpriseCapabilityReadModel struct {
	db *database.TenantAwareDB
}

type UpdateCapabilityParams struct {
	ID          string
	Name        string
	Description string
	Category    string
}

func NewEnterpriseCapabilityReadModel(db *database.TenantAwareDB) *EnterpriseCapabilityReadModel {
	return &EnterpriseCapabilityReadModel{db: db}
}

func (rm *EnterpriseCapabilityReadModel) execByID(ctx context.Context, query string, id string) error {
	return rm.execTenantQuery(ctx, query, func(tid string) []interface{} { return []interface{}{tid, id} })
}

func (rm *EnterpriseCapabilityReadModel) execTenantQuery(ctx context.Context, query string, buildArgs func(tenantID string) []interface{}) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, buildArgs(tenantID.Value())...)
	return err
}

func (rm *EnterpriseCapabilityReadModel) Insert(ctx context.Context, dto EnterpriseCapabilityDTO) error {
	err := rm.execByID(ctx,
		"DELETE FROM enterprisearchitecture.enterprise_capabilities WHERE tenant_id = $1 AND id = $2",
		dto.ID,
	)
	if err != nil {
		return err
	}

	return rm.execTenantQuery(ctx,
		`INSERT INTO enterprisearchitecture.enterprise_capabilities
		 (id, tenant_id, name, description, category, active, link_count, domain_count, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		func(tid string) []interface{} {
			return []interface{}{dto.ID, tid, dto.Name, dto.Description, dto.Category, dto.Active, 0, 0, dto.CreatedAt}
		},
	)
}

func (rm *EnterpriseCapabilityReadModel) Update(ctx context.Context, params UpdateCapabilityParams) error {
	return rm.execTenantQuery(ctx,
		`UPDATE enterprisearchitecture.enterprise_capabilities SET name = $1, description = $2, category = $3, updated_at = CURRENT_TIMESTAMP
		 WHERE tenant_id = $4 AND id = $5`,
		func(tid string) []interface{} {
			return []interface{}{params.Name, params.Description, params.Category, tid, params.ID}
		},
	)
}

func (rm *EnterpriseCapabilityReadModel) Delete(ctx context.Context, id string) error {
	return rm.execByID(ctx, "UPDATE enterprisearchitecture.enterprise_capabilities SET active = false, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $1 AND id = $2", id)
}

func (rm *EnterpriseCapabilityReadModel) IncrementLinkCount(ctx context.Context, id string) error {
	return rm.execByID(ctx, "UPDATE enterprisearchitecture.enterprise_capabilities SET link_count = link_count + 1 WHERE tenant_id = $1 AND id = $2", id)
}

func (rm *EnterpriseCapabilityReadModel) DecrementLinkCount(ctx context.Context, id string) error {
	return rm.execByID(ctx, "UPDATE enterprisearchitecture.enterprise_capabilities SET link_count = GREATEST(0, link_count - 1) WHERE tenant_id = $1 AND id = $2", id)
}

func (rm *EnterpriseCapabilityReadModel) RecalculateDomainCount(ctx context.Context, enterpriseCapabilityID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE enterprisearchitecture.enterprise_capabilities SET domain_count = (
			SELECT COUNT(DISTINCT dcm.business_domain_id)
			FROM enterprisearchitecture.enterprise_capability_links ecl
			JOIN enterprisearchitecture.domain_capability_metadata dcm
				ON dcm.capability_id = ecl.domain_capability_id
				AND dcm.tenant_id = ecl.tenant_id
			WHERE ecl.tenant_id = $1
				AND ecl.enterprise_capability_id = $2
				AND dcm.business_domain_id IS NOT NULL
		)
		WHERE tenant_id = $1 AND id = $2`

	_, err = rm.db.ExecContext(ctx, query, tenantID.Value(), enterpriseCapabilityID)
	return err
}

func (rm *EnterpriseCapabilityReadModel) GetAll(ctx context.Context) ([]EnterpriseCapabilityDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var capabilities []EnterpriseCapabilityDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT "+ecSelectColumns+" FROM enterprisearchitecture.enterprise_capabilities WHERE tenant_id = $1 AND active = true ORDER BY name",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			dto, scanErr := scanEnterpriseCapability(rows)
			if scanErr != nil {
				return scanErr
			}
			capabilities = append(capabilities, dto)
		}

		return rows.Err()
	})

	return capabilities, err
}

func (rm *EnterpriseCapabilityReadModel) GetByID(ctx context.Context, id string) (*EnterpriseCapabilityDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var result *EnterpriseCapabilityDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		dto, scanErr := scanEnterpriseCapability(tx.QueryRowContext(ctx,
			"SELECT "+ecSelectColumns+" FROM enterprisearchitecture.enterprise_capabilities WHERE tenant_id = $1 AND id = $2",
			tenantID.Value(), id,
		))
		if scanErr == sql.ErrNoRows {
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		result = &dto
		return nil
	})

	return result, err
}

const ecSelectColumns = "id, name, description, category, active, target_maturity, link_count, domain_count, created_at, updated_at"

type ecRowScanner interface {
	Scan(dest ...any) error
}

func scanEnterpriseCapability(row ecRowScanner) (EnterpriseCapabilityDTO, error) {
	var dto EnterpriseCapabilityDTO
	var updatedAt sql.NullTime
	var targetMaturity sql.NullInt64
	var description, category sql.NullString

	err := row.Scan(&dto.ID, &dto.Name, &description, &category, &dto.Active, &targetMaturity, &dto.LinkCount, &dto.DomainCount, &dto.CreatedAt, &updatedAt)
	if err != nil {
		return dto, err
	}

	if updatedAt.Valid {
		dto.UpdatedAt = &updatedAt.Time
	}
	if targetMaturity.Valid {
		tm := int(targetMaturity.Int64)
		dto.TargetMaturity = &tm
	}
	if description.Valid {
		dto.Description = description.String
	}
	if category.Valid {
		dto.Category = category.String
	}
	return dto, nil
}

func (rm *EnterpriseCapabilityReadModel) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return false, err
	}

	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		if excludeID != "" {
			return tx.QueryRowContext(ctx,
				"SELECT COUNT(*) FROM enterprisearchitecture.enterprise_capabilities WHERE tenant_id = $1 AND LOWER(name) = LOWER($2) AND id != $3 AND active = true",
				tenantID.Value(), strings.TrimSpace(name), excludeID,
			).Scan(&count)
		}
		return tx.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM enterprisearchitecture.enterprise_capabilities WHERE tenant_id = $1 AND LOWER(name) = LOWER($2) AND active = true",
			tenantID.Value(), strings.TrimSpace(name),
		).Scan(&count)
	})

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (rm *EnterpriseCapabilityReadModel) UpdateTargetMaturity(ctx context.Context, id string, targetMaturity int) error {
	return rm.execTenantQuery(ctx,
		`UPDATE enterprisearchitecture.enterprise_capabilities SET target_maturity = $1, updated_at = CURRENT_TIMESTAMP
		 WHERE tenant_id = $2 AND id = $3`,
		func(tid string) []interface{} { return []interface{}{targetMaturity, tid, id} },
	)
}
