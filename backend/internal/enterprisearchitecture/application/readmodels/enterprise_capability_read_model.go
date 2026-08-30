package readmodels

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/lib/pq"

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
		 (id, tenant_id, name, description, category, active, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		func(tid string) []interface{} {
			return []interface{}{dto.ID, tid, dto.Name, dto.Description, dto.Category, dto.Active, dto.CreatedAt}
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

func (rm *EnterpriseCapabilityReadModel) GetAll(ctx context.Context) ([]EnterpriseCapabilityDTO, error) {
	return rm.listActive(ctx, nil)
}

func (rm *EnterpriseCapabilityReadModel) GetByIDs(ctx context.Context, ids []string) ([]EnterpriseCapabilityDTO, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return rm.listActive(ctx, ids)
}

func (rm *EnterpriseCapabilityReadModel) listActive(ctx context.Context, ids []string) ([]EnterpriseCapabilityDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	if ids == nil {
		return rm.queryEnterpriseCapabilities(ctx, ecActiveSelect+" ORDER BY name", tenantID.Value())
	}
	return rm.queryEnterpriseCapabilities(ctx, ecActiveSelect+" AND id = ANY($2) ORDER BY name", tenantID.Value(), pq.Array(ids))
}

func (rm *EnterpriseCapabilityReadModel) queryEnterpriseCapabilities(ctx context.Context, query string, args ...any) ([]EnterpriseCapabilityDTO, error) {
	var capabilities []EnterpriseCapabilityDTO
	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

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

func (rm *EnterpriseCapabilityReadModel) ActiveEnterpriseCapabilityNames(ctx context.Context) (map[string]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	names := map[string]string{}
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, name FROM enterprisearchitecture.enterprise_capabilities WHERE tenant_id = $1 AND active = true",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var id, name string
			if err := rows.Scan(&id, &name); err != nil {
				return err
			}
			names[id] = name
		}
		return rows.Err()
	})
	return names, err
}

func (rm *EnterpriseCapabilityReadModel) Count(ctx context.Context) (int, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return 0, err
	}

	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM enterprisearchitecture.enterprise_capabilities WHERE tenant_id = $1",
			tenantID.Value(),
		).Scan(&count)
	})
	return count, err
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

const ecSelectColumns = "id, name, description, category, active, target_maturity, created_at, updated_at"

const ecActiveSelect = "SELECT " + ecSelectColumns + " FROM enterprisearchitecture.enterprise_capabilities WHERE tenant_id = $1 AND active = true"

type ecRowScanner interface {
	Scan(dest ...any) error
}

func scanEnterpriseCapability(row ecRowScanner) (EnterpriseCapabilityDTO, error) {
	var dto EnterpriseCapabilityDTO
	var updatedAt sql.NullTime
	var targetMaturity sql.NullInt64
	var description, category sql.NullString

	err := row.Scan(&dto.ID, &dto.Name, &description, &category, &dto.Active, &targetMaturity, &dto.CreatedAt, &updatedAt)
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
