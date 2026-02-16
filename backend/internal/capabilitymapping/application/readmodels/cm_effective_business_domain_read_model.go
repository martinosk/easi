package readmodels

import (
	"context"
	"database/sql"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type CMEffectiveBusinessDomainDTO struct {
	CapabilityID       string
	BusinessDomainID   string
	BusinessDomainName string
	L1CapabilityID     string
}

type CMEffectiveBusinessDomainReadModel struct {
	db *database.TenantAwareDB
}

func NewCMEffectiveBusinessDomainReadModel(db *database.TenantAwareDB) *CMEffectiveBusinessDomainReadModel {
	return &CMEffectiveBusinessDomainReadModel{db: db}
}

func (rm *CMEffectiveBusinessDomainReadModel) Upsert(ctx context.Context, dto CMEffectiveBusinessDomainDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM capabilitymapping.cm_effective_business_domain WHERE tenant_id = $1 AND capability_id = $2",
		tenantID.Value(), dto.CapabilityID,
	)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO capabilitymapping.cm_effective_business_domain
		 (tenant_id, capability_id, business_domain_id, business_domain_name, l1_capability_id)
		 VALUES ($1, $2, $3, $4, $5)`,
		tenantID.Value(), dto.CapabilityID,
		cmNullIfEmpty(dto.BusinessDomainID), cmNullIfEmpty(dto.BusinessDomainName),
		dto.L1CapabilityID,
	)
	return err
}

func (rm *CMEffectiveBusinessDomainReadModel) Delete(ctx context.Context, capabilityID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM capabilitymapping.cm_effective_business_domain WHERE tenant_id = $1 AND capability_id = $2",
		tenantID.Value(), capabilityID,
	)
	return err
}

func (rm *CMEffectiveBusinessDomainReadModel) GetByCapabilityID(ctx context.Context, capabilityID string) (*CMEffectiveBusinessDomainDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto CMEffectiveBusinessDomainDTO
	var businessDomainID, businessDomainName sql.NullString
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT capability_id, business_domain_id, business_domain_name, l1_capability_id
			 FROM capabilitymapping.cm_effective_business_domain WHERE tenant_id = $1 AND capability_id = $2`,
			tenantID.Value(), capabilityID,
		).Scan(&dto.CapabilityID, &businessDomainID, &businessDomainName, &dto.L1CapabilityID)

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

	dto.BusinessDomainID = businessDomainID.String
	dto.BusinessDomainName = businessDomainName.String

	return &dto, nil
}

func (rm *CMEffectiveBusinessDomainReadModel) UpdateBusinessDomainForL1Subtree(ctx context.Context, l1CapabilityID string, bdID string, bdName string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE capabilitymapping.cm_effective_business_domain
		 SET business_domain_id = $2, business_domain_name = $3
		 WHERE tenant_id = $1 AND l1_capability_id = $4`,
		tenantID.Value(), cmNullIfEmpty(bdID), cmNullIfEmpty(bdName), l1CapabilityID,
	)
	return err
}

func cmNullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
