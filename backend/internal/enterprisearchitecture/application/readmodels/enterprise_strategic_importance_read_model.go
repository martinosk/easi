package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type EnterpriseStrategicImportanceDTO struct {
	ID                     string      `json:"id"`
	EnterpriseCapabilityID string      `json:"enterpriseCapabilityId"`
	PillarID               string      `json:"pillarId"`
	PillarName             string      `json:"pillarName"`
	Importance             int         `json:"importance"`
	ImportanceLabel        string      `json:"importanceLabel"`
	Rationale              string      `json:"rationale,omitempty"`
	SetAt                  time.Time   `json:"setAt"`
	UpdatedAt              *time.Time  `json:"updatedAt,omitempty"`
	Links                  types.Links `json:"_links,omitempty"`
}

type EnterpriseStrategicImportanceReadModel struct {
	db *database.TenantAwareDB
}

func NewEnterpriseStrategicImportanceReadModel(db *database.TenantAwareDB) *EnterpriseStrategicImportanceReadModel {
	return &EnterpriseStrategicImportanceReadModel{db: db}
}

func (rm *EnterpriseStrategicImportanceReadModel) Insert(ctx context.Context, dto EnterpriseStrategicImportanceDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM enterprisearchitecture.enterprise_strategic_importance WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), dto.ID,
	)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO enterprisearchitecture.enterprise_strategic_importance
		 (id, tenant_id, enterprise_capability_id, pillar_id, pillar_name, importance, rationale, set_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		dto.ID, tenantID.Value(), dto.EnterpriseCapabilityID, dto.PillarID, dto.PillarName, dto.Importance, dto.Rationale, dto.SetAt,
	)
	return err
}

func (rm *EnterpriseStrategicImportanceReadModel) Update(ctx context.Context, id string, importance int, rationale string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE enterprisearchitecture.enterprise_strategic_importance SET importance = $1, rationale = $2, updated_at = CURRENT_TIMESTAMP
		 WHERE tenant_id = $3 AND id = $4`,
		importance, rationale, tenantID.Value(), id,
	)
	return err
}

func (rm *EnterpriseStrategicImportanceReadModel) Delete(ctx context.Context, id string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM enterprisearchitecture.enterprise_strategic_importance WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), id,
	)
	return err
}

func (rm *EnterpriseStrategicImportanceReadModel) GetByEnterpriseCapabilityID(ctx context.Context, enterpriseCapabilityID string) ([]EnterpriseStrategicImportanceDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var ratings []EnterpriseStrategicImportanceDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT id, enterprise_capability_id, pillar_id, pillar_name, importance, rationale, set_at, updated_at
			 FROM enterprisearchitecture.enterprise_strategic_importance
			 WHERE tenant_id = $1 AND enterprise_capability_id = $2
			 ORDER BY pillar_name`,
			tenantID.Value(), enterpriseCapabilityID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto EnterpriseStrategicImportanceDTO
			var updatedAt sql.NullTime
			var rationale sql.NullString
			if err := rows.Scan(&dto.ID, &dto.EnterpriseCapabilityID, &dto.PillarID, &dto.PillarName, &dto.Importance, &rationale, &dto.SetAt, &updatedAt); err != nil {
				return err
			}
			if updatedAt.Valid {
				dto.UpdatedAt = &updatedAt.Time
			}
			if rationale.Valid {
				dto.Rationale = rationale.String
			}
			dto.ImportanceLabel = getImportanceLabel(dto.Importance)
			ratings = append(ratings, dto)
		}

		return rows.Err()
	})

	return ratings, err
}

func (rm *EnterpriseStrategicImportanceReadModel) GetByCapabilityAndPillar(ctx context.Context, enterpriseCapabilityID, pillarID string) (*EnterpriseStrategicImportanceDTO, error) {
	return rm.querySingle(ctx,
		`SELECT id, enterprise_capability_id, pillar_id, pillar_name, importance, rationale, set_at, updated_at
		 FROM enterprisearchitecture.enterprise_strategic_importance
		 WHERE tenant_id = $1 AND enterprise_capability_id = $2 AND pillar_id = $3`,
		enterpriseCapabilityID, pillarID,
	)
}

func (rm *EnterpriseStrategicImportanceReadModel) GetByID(ctx context.Context, id string) (*EnterpriseStrategicImportanceDTO, error) {
	return rm.querySingle(ctx,
		`SELECT id, enterprise_capability_id, pillar_id, pillar_name, importance, rationale, set_at, updated_at
		 FROM enterprisearchitecture.enterprise_strategic_importance
		 WHERE tenant_id = $1 AND id = $2`,
		id,
	)
}

func (rm *EnterpriseStrategicImportanceReadModel) querySingle(ctx context.Context, query string, args ...interface{}) (*EnterpriseStrategicImportanceDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto EnterpriseStrategicImportanceDTO
	fields := importanceScanFields{dto: &dto}
	var notFound bool

	queryArgs := append([]interface{}{tenantID.Value()}, args...)

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, query, queryArgs...).Scan(fields.scanTargets()...)
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

	fields.finalize()
	return &dto, nil
}

func getImportanceLabel(importance int) string {
	labels := map[int]string{
		1: "Low",
		2: "Below Average",
		3: "Average",
		4: "Above Average",
		5: "Critical",
	}
	return labels[importance]
}

type importanceScanFields struct {
	dto       *EnterpriseStrategicImportanceDTO
	updatedAt sql.NullTime
	rationale sql.NullString
}

func (f *importanceScanFields) scanTargets() []interface{} {
	return []interface{}{
		&f.dto.ID, &f.dto.EnterpriseCapabilityID, &f.dto.PillarID,
		&f.dto.PillarName, &f.dto.Importance, &f.rationale, &f.dto.SetAt, &f.updatedAt,
	}
}

func (f *importanceScanFields) finalize() {
	if f.updatedAt.Valid {
		f.dto.UpdatedAt = &f.updatedAt.Time
	}
	if f.rationale.Valid {
		f.dto.Rationale = f.rationale.String
	}
	f.dto.ImportanceLabel = getImportanceLabel(f.dto.Importance)
}
