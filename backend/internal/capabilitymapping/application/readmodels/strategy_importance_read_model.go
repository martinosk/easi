package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type StrategyImportanceDTO struct {
	ID                 string      `json:"id"`
	BusinessDomainID   string      `json:"businessDomainId"`
	BusinessDomainName string      `json:"businessDomainName"`
	CapabilityID       string      `json:"capabilityId"`
	CapabilityName     string      `json:"capabilityName"`
	PillarID           string      `json:"pillarId"`
	PillarName         string      `json:"pillarName"`
	Importance         int         `json:"importance"`
	ImportanceLabel    string      `json:"importanceLabel"`
	Rationale          string      `json:"rationale,omitempty"`
	SetAt              time.Time   `json:"setAt"`
	UpdatedAt          *time.Time  `json:"updatedAt,omitempty"`
	Links              types.Links `json:"_links,omitempty"`
}

type StrategyImportanceReadModel struct {
	db *database.TenantAwareDB
}

func NewStrategyImportanceReadModel(db *database.TenantAwareDB) *StrategyImportanceReadModel {
	return &StrategyImportanceReadModel{db: db}
}

func (rm *StrategyImportanceReadModel) Insert(ctx context.Context, dto StrategyImportanceDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO strategy_importance
		(id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_name,
		pillar_id, pillar_name, importance, importance_label, rationale, set_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		dto.ID, tenantID.Value(), dto.BusinessDomainID, dto.BusinessDomainName,
		dto.CapabilityID, dto.CapabilityName, dto.PillarID, dto.PillarName,
		dto.Importance, dto.ImportanceLabel, dto.Rationale, dto.SetAt,
	)
	return err
}

func (rm *StrategyImportanceReadModel) Update(ctx context.Context, dto StrategyImportanceDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE strategy_importance
		SET importance = $1, importance_label = $2, rationale = $3, updated_at = $4
		WHERE tenant_id = $5 AND id = $6`,
		dto.Importance, dto.ImportanceLabel, dto.Rationale, time.Now().UTC(), tenantID.Value(), dto.ID,
	)
	return err
}

func (rm *StrategyImportanceReadModel) Delete(ctx context.Context, id string) error {
	return rm.deleteByColumn(ctx, "id", id)
}

func (rm *StrategyImportanceReadModel) DeleteByCapability(ctx context.Context, capabilityID string) error {
	return rm.deleteByColumn(ctx, "capability_id", capabilityID)
}

func (rm *StrategyImportanceReadModel) DeleteByBusinessDomain(ctx context.Context, domainID string) error {
	return rm.deleteByColumn(ctx, "business_domain_id", domainID)
}

func (rm *StrategyImportanceReadModel) deleteByColumn(ctx context.Context, column, value string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM strategy_importance WHERE tenant_id = $1 AND "+column+" = $2",
		tenantID.Value(), value,
	)
	return err
}

const importanceSelectColumns = `id, business_domain_id, business_domain_name, capability_id, capability_name,
pillar_id, pillar_name, importance, importance_label, rationale, set_at, updated_at`

func (rm *StrategyImportanceReadModel) GetByID(ctx context.Context, id string) (*StrategyImportanceDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto StrategyImportanceDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := "SELECT " + importanceSelectColumns + " FROM strategy_importance WHERE tenant_id = $1 AND id = $2"
		row := tx.QueryRowContext(ctx, query, tenantID.Value(), id)

		var scanErr error
		dto, scanErr = rm.scanImportanceRow(row)
		if scanErr == sql.ErrNoRows {
			notFound = true
			return nil
		}
		return scanErr
	})

	if err != nil {
		return nil, err
	}
	if notFound {
		return nil, nil
	}

	return &dto, nil
}

func (rm *StrategyImportanceReadModel) GetByDomainAndCapability(ctx context.Context, domainID, capabilityID string) ([]StrategyImportanceDTO, error) {
	query := "SELECT " + importanceSelectColumns + " FROM strategy_importance WHERE tenant_id = $1 AND business_domain_id = $2 AND capability_id = $3 ORDER BY pillar_name"
	return rm.queryImportanceList(ctx, query, domainID, capabilityID)
}

func (rm *StrategyImportanceReadModel) GetByDomain(ctx context.Context, domainID string) ([]StrategyImportanceDTO, error) {
	query := "SELECT " + importanceSelectColumns + " FROM strategy_importance WHERE tenant_id = $1 AND business_domain_id = $2 ORDER BY capability_name, pillar_name"
	return rm.queryImportanceList(ctx, query, domainID)
}

func (rm *StrategyImportanceReadModel) GetByCapability(ctx context.Context, capabilityID string) ([]StrategyImportanceDTO, error) {
	query := "SELECT " + importanceSelectColumns + " FROM strategy_importance WHERE tenant_id = $1 AND capability_id = $2 ORDER BY business_domain_name, pillar_name"
	return rm.queryImportanceList(ctx, query, capabilityID)
}

func (rm *StrategyImportanceReadModel) Exists(ctx context.Context, domainID, capabilityID, pillarID string) (bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return false, err
	}

	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM strategy_importance WHERE tenant_id = $1 AND business_domain_id = $2 AND capability_id = $3 AND pillar_id = $4",
			tenantID.Value(), domainID, capabilityID, pillarID,
		).Scan(&count)
	})

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (rm *StrategyImportanceReadModel) queryImportanceList(ctx context.Context, query string, params ...string) ([]StrategyImportanceDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	args := make([]any, 0, len(params)+1)
	args = append(args, tenantID.Value())
	for _, p := range params {
		args = append(args, p)
	}

	var results []StrategyImportanceDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			dto, scanErr := rm.scanImportanceRow(rows)
			if scanErr != nil {
				return scanErr
			}
			results = append(results, dto)
		}
		return rows.Err()
	})

	return results, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (rm *StrategyImportanceReadModel) scanImportanceRow(row rowScanner) (StrategyImportanceDTO, error) {
	var dto StrategyImportanceDTO
	var rationale sql.NullString
	var updatedAt sql.NullTime

	err := row.Scan(&dto.ID, &dto.BusinessDomainID, &dto.BusinessDomainName, &dto.CapabilityID, &dto.CapabilityName,
		&dto.PillarID, &dto.PillarName, &dto.Importance, &dto.ImportanceLabel, &rationale, &dto.SetAt, &updatedAt)
	if err != nil {
		return dto, err
	}

	if rationale.Valid {
		dto.Rationale = rationale.String
	}
	if updatedAt.Valid {
		dto.UpdatedAt = &updatedAt.Time
	}
	return dto, nil
}
