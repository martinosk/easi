package readmodels

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

const effectiveImportanceSelectColumns = `capability_id, pillar_id, business_domain_id, effective_importance, importance_label,
		source_capability_id, source_capability_name, is_inherited, rationale, computed_at`

type EffectiveImportanceDTO struct {
	CapabilityID         string    `json:"capabilityId"`
	PillarID             string    `json:"pillarId"`
	BusinessDomainID     string    `json:"businessDomainId"`
	EffectiveImportance  int       `json:"effectiveImportance"`
	ImportanceLabel      string    `json:"importanceLabel"`
	SourceCapabilityID   string    `json:"sourceCapabilityId"`
	SourceCapabilityName string    `json:"sourceCapabilityName"`
	IsInherited          bool      `json:"isInherited"`
	Rationale            string    `json:"rationale"`
	ComputedAt           time.Time `json:"computedAt"`
}

type effectiveImportanceRowScanner interface {
	Scan(dest ...any) error
}

type EffectiveCapabilityImportanceReadModel struct {
	db *database.TenantAwareDB
}

func NewEffectiveCapabilityImportanceReadModel(db *database.TenantAwareDB) *EffectiveCapabilityImportanceReadModel {
	return &EffectiveCapabilityImportanceReadModel{db: db}
}

func (rm *EffectiveCapabilityImportanceReadModel) Upsert(ctx context.Context, dto EffectiveImportanceDTO) error {
	return rm.execTenantQuery(ctx,
		`INSERT INTO effective_capability_importance
		(tenant_id, capability_id, pillar_id, business_domain_id, effective_importance, importance_label,
		source_capability_id, source_capability_name, is_inherited, rationale, computed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (tenant_id, capability_id, pillar_id, business_domain_id)
		DO UPDATE SET
			effective_importance = EXCLUDED.effective_importance,
			importance_label = EXCLUDED.importance_label,
			source_capability_id = EXCLUDED.source_capability_id,
			source_capability_name = EXCLUDED.source_capability_name,
			is_inherited = EXCLUDED.is_inherited,
			rationale = EXCLUDED.rationale,
			computed_at = EXCLUDED.computed_at`,
		dto.CapabilityID, dto.PillarID, dto.BusinessDomainID,
		dto.EffectiveImportance, dto.ImportanceLabel, dto.SourceCapabilityID,
		dto.SourceCapabilityName, dto.IsInherited, dto.Rationale, dto.ComputedAt,
	)
}

func (rm *EffectiveCapabilityImportanceReadModel) Delete(ctx context.Context, capabilityID, pillarID, businessDomainID string) error {
	if pillarID == "" {
		return rm.execTenantQuery(ctx,
			"DELETE FROM effective_capability_importance WHERE tenant_id = $1 AND capability_id = $2 AND business_domain_id = $3",
			capabilityID, businessDomainID,
		)
	}
	return rm.execTenantQuery(ctx,
		"DELETE FROM effective_capability_importance WHERE tenant_id = $1 AND capability_id = $2 AND pillar_id = $3 AND business_domain_id = $4",
		capabilityID, pillarID, businessDomainID,
	)
}

func (rm *EffectiveCapabilityImportanceReadModel) DeleteByCapability(ctx context.Context, capabilityID string) error {
	return rm.execTenantQuery(ctx,
		"DELETE FROM effective_capability_importance WHERE tenant_id = $1 AND capability_id = $2",
		capabilityID,
	)
}

func (rm *EffectiveCapabilityImportanceReadModel) DeleteBySourceCapability(ctx context.Context, sourceCapabilityID string) error {
	return rm.execTenantQuery(ctx,
		"DELETE FROM effective_capability_importance WHERE tenant_id = $1 AND source_capability_id = $2",
		sourceCapabilityID,
	)
}

func (rm *EffectiveCapabilityImportanceReadModel) DeleteByBusinessDomain(ctx context.Context, domainID string) error {
	return rm.execTenantQuery(ctx,
		"DELETE FROM effective_capability_importance WHERE tenant_id = $1 AND business_domain_id = $2",
		domainID,
	)
}

func (rm *EffectiveCapabilityImportanceReadModel) DeleteByPillarAndDomain(ctx context.Context, pillarID, businessDomainID string) error {
	return rm.execTenantQuery(ctx,
		"DELETE FROM effective_capability_importance WHERE tenant_id = $1 AND pillar_id = $2 AND business_domain_id = $3",
		pillarID, businessDomainID,
	)
}

func (rm *EffectiveCapabilityImportanceReadModel) GetByCapabilityPillarDomain(ctx context.Context, capabilityID, pillarID, businessDomainID string) (*EffectiveImportanceDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto EffectiveImportanceDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx,
			fmt.Sprintf(`SELECT %s FROM effective_capability_importance
			WHERE tenant_id = $1 AND capability_id = $2 AND pillar_id = $3 AND business_domain_id = $4`, effectiveImportanceSelectColumns),
			tenantID.Value(), capabilityID, pillarID, businessDomainID,
		)

		scanErr := scanEffectiveImportanceRow(row, &dto)
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

func (rm *EffectiveCapabilityImportanceReadModel) GetByCapability(ctx context.Context, capabilityID string) ([]EffectiveImportanceDTO, error) {
	return rm.queryEffectiveImportanceList(ctx,
		fmt.Sprintf(`SELECT %s FROM effective_capability_importance
		WHERE tenant_id = $1 AND capability_id = $2
		ORDER BY pillar_id, business_domain_id`, effectiveImportanceSelectColumns),
		capabilityID,
	)
}

func (rm *EffectiveCapabilityImportanceReadModel) GetBySourceCapabilityAndPillarAndDomain(ctx context.Context, sourceCapabilityID, pillarID, businessDomainID string) ([]EffectiveImportanceDTO, error) {
	return rm.queryEffectiveImportanceList(ctx,
		fmt.Sprintf(`SELECT %s FROM effective_capability_importance
		WHERE tenant_id = $1 AND source_capability_id = $2 AND pillar_id = $3 AND business_domain_id = $4
		ORDER BY capability_id`, effectiveImportanceSelectColumns),
		sourceCapabilityID, pillarID, businessDomainID,
	)
}

func (rm *EffectiveCapabilityImportanceReadModel) execTenantQuery(ctx context.Context, query string, args ...interface{}) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, append([]interface{}{tenantID.Value()}, args...)...)
	return err
}

func (rm *EffectiveCapabilityImportanceReadModel) queryEffectiveImportanceList(ctx context.Context, query string, args ...string) ([]EffectiveImportanceDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	queryArgs := make([]interface{}, 0, len(args)+1)
	queryArgs = append(queryArgs, tenantID.Value())
	for _, a := range args {
		queryArgs = append(queryArgs, a)
	}

	var results []EffectiveImportanceDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, queryArgs...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto EffectiveImportanceDTO
			if err := scanEffectiveImportanceRow(rows, &dto); err != nil {
				return err
			}
			results = append(results, dto)
		}
		return rows.Err()
	})

	return results, err
}

func scanEffectiveImportanceRow(row effectiveImportanceRowScanner, dto *EffectiveImportanceDTO) error {
	return row.Scan(&dto.CapabilityID, &dto.PillarID, &dto.BusinessDomainID,
		&dto.EffectiveImportance, &dto.ImportanceLabel, &dto.SourceCapabilityID,
		&dto.SourceCapabilityName, &dto.IsInherited, &dto.Rationale, &dto.ComputedAt)
}
