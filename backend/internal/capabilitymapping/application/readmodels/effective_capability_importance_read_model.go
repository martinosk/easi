package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type EffectiveImportanceDTO struct {
	CapabilityID           string    `json:"capabilityId"`
	PillarID               string    `json:"pillarId"`
	BusinessDomainID       string    `json:"businessDomainId"`
	EffectiveImportance    int       `json:"effectiveImportance"`
	ImportanceLabel        string    `json:"importanceLabel"`
	SourceCapabilityID     string    `json:"sourceCapabilityId"`
	SourceCapabilityName   string    `json:"sourceCapabilityName"`
	IsInherited            bool      `json:"isInherited"`
	ComputedAt             time.Time `json:"computedAt"`
}

type EffectiveCapabilityImportanceReadModel struct {
	db *database.TenantAwareDB
}

func NewEffectiveCapabilityImportanceReadModel(db *database.TenantAwareDB) *EffectiveCapabilityImportanceReadModel {
	return &EffectiveCapabilityImportanceReadModel{db: db}
}

func (rm *EffectiveCapabilityImportanceReadModel) Upsert(ctx context.Context, dto EffectiveImportanceDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO effective_capability_importance
		(tenant_id, capability_id, pillar_id, business_domain_id, effective_importance, importance_label,
		source_capability_id, source_capability_name, is_inherited, computed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (tenant_id, capability_id, pillar_id, business_domain_id)
		DO UPDATE SET
			effective_importance = EXCLUDED.effective_importance,
			importance_label = EXCLUDED.importance_label,
			source_capability_id = EXCLUDED.source_capability_id,
			source_capability_name = EXCLUDED.source_capability_name,
			is_inherited = EXCLUDED.is_inherited,
			computed_at = EXCLUDED.computed_at`,
		tenantID.Value(), dto.CapabilityID, dto.PillarID, dto.BusinessDomainID,
		dto.EffectiveImportance, dto.ImportanceLabel, dto.SourceCapabilityID,
		dto.SourceCapabilityName, dto.IsInherited, dto.ComputedAt,
	)
	return err
}

func (rm *EffectiveCapabilityImportanceReadModel) Delete(ctx context.Context, capabilityID, pillarID, businessDomainID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	if pillarID == "" {
		_, err = rm.db.ExecContext(ctx,
			"DELETE FROM effective_capability_importance WHERE tenant_id = $1 AND capability_id = $2 AND business_domain_id = $3",
			tenantID.Value(), capabilityID, businessDomainID,
		)
	} else {
		_, err = rm.db.ExecContext(ctx,
			"DELETE FROM effective_capability_importance WHERE tenant_id = $1 AND capability_id = $2 AND pillar_id = $3 AND business_domain_id = $4",
			tenantID.Value(), capabilityID, pillarID, businessDomainID,
		)
	}
	return err
}

func (rm *EffectiveCapabilityImportanceReadModel) DeleteByCapability(ctx context.Context, capabilityID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM effective_capability_importance WHERE tenant_id = $1 AND capability_id = $2",
		tenantID.Value(), capabilityID,
	)
	return err
}

func (rm *EffectiveCapabilityImportanceReadModel) DeleteBySourceCapability(ctx context.Context, sourceCapabilityID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM effective_capability_importance WHERE tenant_id = $1 AND source_capability_id = $2",
		tenantID.Value(), sourceCapabilityID,
	)
	return err
}

func (rm *EffectiveCapabilityImportanceReadModel) DeleteByBusinessDomain(ctx context.Context, domainID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM effective_capability_importance WHERE tenant_id = $1 AND business_domain_id = $2",
		tenantID.Value(), domainID,
	)
	return err
}

func (rm *EffectiveCapabilityImportanceReadModel) DeleteByPillarAndDomain(ctx context.Context, pillarID, businessDomainID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM effective_capability_importance WHERE tenant_id = $1 AND pillar_id = $2 AND business_domain_id = $3",
		tenantID.Value(), pillarID, businessDomainID,
	)
	return err
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
			`SELECT capability_id, pillar_id, business_domain_id, effective_importance, importance_label,
			source_capability_id, source_capability_name, is_inherited, computed_at
			FROM effective_capability_importance
			WHERE tenant_id = $1 AND capability_id = $2 AND pillar_id = $3 AND business_domain_id = $4`,
			tenantID.Value(), capabilityID, pillarID, businessDomainID,
		)

		scanErr := row.Scan(&dto.CapabilityID, &dto.PillarID, &dto.BusinessDomainID,
			&dto.EffectiveImportance, &dto.ImportanceLabel, &dto.SourceCapabilityID,
			&dto.SourceCapabilityName, &dto.IsInherited, &dto.ComputedAt)
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
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var results []EffectiveImportanceDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT capability_id, pillar_id, business_domain_id, effective_importance, importance_label,
			source_capability_id, source_capability_name, is_inherited, computed_at
			FROM effective_capability_importance
			WHERE tenant_id = $1 AND capability_id = $2
			ORDER BY pillar_id, business_domain_id`,
			tenantID.Value(), capabilityID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto EffectiveImportanceDTO
			if err := rows.Scan(&dto.CapabilityID, &dto.PillarID, &dto.BusinessDomainID,
				&dto.EffectiveImportance, &dto.ImportanceLabel, &dto.SourceCapabilityID,
				&dto.SourceCapabilityName, &dto.IsInherited, &dto.ComputedAt); err != nil {
				return err
			}
			results = append(results, dto)
		}
		return rows.Err()
	})

	return results, err
}

func (rm *EffectiveCapabilityImportanceReadModel) GetBySourceCapabilityAndPillarAndDomain(ctx context.Context, sourceCapabilityID, pillarID, businessDomainID string) ([]EffectiveImportanceDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var results []EffectiveImportanceDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT capability_id, pillar_id, business_domain_id, effective_importance, importance_label,
			source_capability_id, source_capability_name, is_inherited, computed_at
			FROM effective_capability_importance
			WHERE tenant_id = $1 AND source_capability_id = $2 AND pillar_id = $3 AND business_domain_id = $4
			ORDER BY capability_id`,
			tenantID.Value(), sourceCapabilityID, pillarID, businessDomainID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto EffectiveImportanceDTO
			if err := rows.Scan(&dto.CapabilityID, &dto.PillarID, &dto.BusinessDomainID,
				&dto.EffectiveImportance, &dto.ImportanceLabel, &dto.SourceCapabilityID,
				&dto.SourceCapabilityName, &dto.IsInherited, &dto.ComputedAt); err != nil {
				return err
			}
			results = append(results, dto)
		}
		return rows.Err()
	})

	return results, err
}
