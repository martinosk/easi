package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type EnterpriseCapabilityLinkDTO struct {
	ID                     string            `json:"id"`
	EnterpriseCapabilityID string            `json:"enterpriseCapabilityId"`
	DomainCapabilityID     string            `json:"domainCapabilityId"`
	DomainCapabilityName   string            `json:"domainCapabilityName,omitempty"`
	BusinessDomainID       string            `json:"businessDomainId,omitempty"`
	BusinessDomainName     string            `json:"businessDomainName,omitempty"`
	MaturityLevel          *int              `json:"maturityLevel,omitempty"`
	LinkedBy               string            `json:"linkedBy"`
	LinkedAt               time.Time         `json:"linkedAt"`
	Links                  map[string]string `json:"_links,omitempty"`
}

type EnterpriseCapabilityLinkReadModel struct {
	db *database.TenantAwareDB
}

func NewEnterpriseCapabilityLinkReadModel(db *database.TenantAwareDB) *EnterpriseCapabilityLinkReadModel {
	return &EnterpriseCapabilityLinkReadModel{db: db}
}

func (rm *EnterpriseCapabilityLinkReadModel) execByID(ctx context.Context, query string, id string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, tenantID.Value(), id)
	return err
}

func (rm *EnterpriseCapabilityLinkReadModel) Insert(ctx context.Context, dto EnterpriseCapabilityLinkDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		dto.ID, tenantID.Value(), dto.EnterpriseCapabilityID, dto.DomainCapabilityID, dto.LinkedBy, dto.LinkedAt,
	)
	return err
}

func (rm *EnterpriseCapabilityLinkReadModel) Delete(ctx context.Context, id string) error {
	return rm.execByID(ctx, "DELETE FROM enterprise_capability_links WHERE tenant_id = $1 AND id = $2", id)
}

func (rm *EnterpriseCapabilityLinkReadModel) DeleteByDomainCapabilityID(ctx context.Context, domainCapabilityID string) error {
	return rm.execByID(ctx, "DELETE FROM enterprise_capability_links WHERE tenant_id = $1 AND domain_capability_id = $2", domainCapabilityID)
}

func (rm *EnterpriseCapabilityLinkReadModel) GetByEnterpriseCapabilityID(ctx context.Context, enterpriseCapabilityID string) ([]EnterpriseCapabilityLinkDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var links []EnterpriseCapabilityLinkDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT ecl.id, ecl.enterprise_capability_id, ecl.domain_capability_id, ecl.linked_by, ecl.linked_at
			 FROM enterprise_capability_links ecl
			 WHERE ecl.tenant_id = $1 AND ecl.enterprise_capability_id = $2
			 ORDER BY ecl.linked_at DESC`,
			tenantID.Value(), enterpriseCapabilityID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto EnterpriseCapabilityLinkDTO
			if err := rows.Scan(&dto.ID, &dto.EnterpriseCapabilityID, &dto.DomainCapabilityID, &dto.LinkedBy, &dto.LinkedAt); err != nil {
				return err
			}
			links = append(links, dto)
		}

		return rows.Err()
	})

	return links, err
}

func (rm *EnterpriseCapabilityLinkReadModel) querySingle(ctx context.Context, query string, args ...interface{}) (*EnterpriseCapabilityLinkDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto EnterpriseCapabilityLinkDTO
	var notFound bool

	queryArgs := append([]interface{}{tenantID.Value()}, args...)

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, query, queryArgs...).Scan(
			&dto.ID, &dto.EnterpriseCapabilityID, &dto.DomainCapabilityID, &dto.LinkedBy, &dto.LinkedAt,
		)
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

func (rm *EnterpriseCapabilityLinkReadModel) GetByDomainCapabilityID(ctx context.Context, domainCapabilityID string) (*EnterpriseCapabilityLinkDTO, error) {
	return rm.querySingle(ctx,
		`SELECT id, enterprise_capability_id, domain_capability_id, linked_by, linked_at
		 FROM enterprise_capability_links WHERE tenant_id = $1 AND domain_capability_id = $2`,
		domainCapabilityID,
	)
}

func (rm *EnterpriseCapabilityLinkReadModel) GetByID(ctx context.Context, id string) (*EnterpriseCapabilityLinkDTO, error) {
	return rm.querySingle(ctx,
		`SELECT id, enterprise_capability_id, domain_capability_id, linked_by, linked_at
		 FROM enterprise_capability_links WHERE tenant_id = $1 AND id = $2`,
		id,
	)
}
