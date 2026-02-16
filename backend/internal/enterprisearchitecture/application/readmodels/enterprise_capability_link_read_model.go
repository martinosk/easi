package readmodels

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"

	"github.com/lib/pq"
)

type EnterpriseCapabilityLinkDTO struct {
	ID                     string      `json:"id"`
	EnterpriseCapabilityID string      `json:"enterpriseCapabilityId"`
	DomainCapabilityID     string      `json:"domainCapabilityId"`
	DomainCapabilityName   string      `json:"domainCapabilityName,omitempty"`
	BusinessDomainID       string      `json:"businessDomainId,omitempty"`
	BusinessDomainName     string      `json:"businessDomainName,omitempty"`
	MaturityLevel          *int        `json:"maturityLevel,omitempty"`
	LinkedBy               string      `json:"linkedBy"`
	LinkedAt               time.Time   `json:"linkedAt"`
	Links                  types.Links `json:"_links,omitempty"`
}

type EnterpriseCapabilityLinkReadModel struct {
	db *database.TenantAwareDB
}

func NewEnterpriseCapabilityLinkReadModel(db *database.TenantAwareDB) *EnterpriseCapabilityLinkReadModel {
	return &EnterpriseCapabilityLinkReadModel{db: db}
}

func buildAnyClauseArgs(tenantID any, ids []string) []any {
	return []any{tenantID, pq.Array(ids)}
}

type linkMutation int

const (
	linkMutationDeleteByID linkMutation = iota
	linkMutationDeleteByDomainCapability
)

var linkMutationQueries = map[linkMutation]string{
	linkMutationDeleteByID:               "DELETE FROM enterprisearchitecture.enterprise_capability_links WHERE tenant_id = $1 AND id = $2",
	linkMutationDeleteByDomainCapability: "DELETE FROM enterprisearchitecture.enterprise_capability_links WHERE tenant_id = $1 AND domain_capability_id = $2",
}

func (rm *EnterpriseCapabilityLinkReadModel) execByID(ctx context.Context, id string, mutation linkMutation) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("resolve tenant for enterprise capability link mutation on %s: %w", id, err)
	}
	_, err = rm.db.ExecContext(ctx, linkMutationQueries[mutation], tenantID.Value(), id)
	if err != nil {
		return fmt.Errorf("execute enterprise capability link mutation on %s for tenant %s: %w", id, tenantID.Value(), err)
	}
	return nil
}

func (rm *EnterpriseCapabilityLinkReadModel) Insert(ctx context.Context, dto EnterpriseCapabilityLinkDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("resolve tenant for insert enterprise capability link %s: %w", dto.ID, err)
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM enterprisearchitecture.enterprise_capability_links WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), dto.ID,
	)
	if err != nil {
		return fmt.Errorf("delete existing enterprise capability link %s before insert: %w", dto.ID, err)
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO enterprisearchitecture.enterprise_capability_links
		 (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		dto.ID, tenantID.Value(), dto.EnterpriseCapabilityID, dto.DomainCapabilityID, dto.LinkedBy, dto.LinkedAt,
	)
	if err != nil {
		return fmt.Errorf("insert enterprise capability link %s for tenant %s: %w", dto.ID, tenantID.Value(), err)
	}
	return nil
}

func (rm *EnterpriseCapabilityLinkReadModel) Delete(ctx context.Context, id string) error {
	return rm.execByID(ctx, id, linkMutationDeleteByID)
}

func (rm *EnterpriseCapabilityLinkReadModel) DeleteByDomainCapabilityID(ctx context.Context, domainCapabilityID string) error {
	return rm.execByID(ctx, domainCapabilityID, linkMutationDeleteByDomainCapability)
}

func (rm *EnterpriseCapabilityLinkReadModel) CountByEnterpriseCapabilityID(ctx context.Context, enterpriseCapabilityID string) (int, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return 0, err
	}

	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM enterprisearchitecture.enterprise_capability_links WHERE tenant_id = $1 AND enterprise_capability_id = $2`,
			tenantID.Value(), enterpriseCapabilityID,
		).Scan(&count)
	})
	return count, err
}

type LinkQueryKind int

const (
	LinkQueryByID LinkQueryKind = iota
	LinkQueryByDomainCapabilityID
)

var linkSingleQueries = map[LinkQueryKind]string{
	LinkQueryByID: `SELECT id, enterprise_capability_id, domain_capability_id, linked_by, linked_at
		 FROM enterprisearchitecture.enterprise_capability_links WHERE tenant_id = $1 AND id = $2`,
	LinkQueryByDomainCapabilityID: `SELECT id, enterprise_capability_id, domain_capability_id, linked_by, linked_at
		 FROM enterprisearchitecture.enterprise_capability_links WHERE tenant_id = $1 AND domain_capability_id = $2`,
}

func (rm *EnterpriseCapabilityLinkReadModel) querySingle(ctx context.Context, id string, kind LinkQueryKind) (*EnterpriseCapabilityLinkDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto EnterpriseCapabilityLinkDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, linkSingleQueries[kind], tenantID.Value(), id).Scan(
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
	return rm.querySingle(ctx, domainCapabilityID, LinkQueryByDomainCapabilityID)
}

func (rm *EnterpriseCapabilityLinkReadModel) GetByID(ctx context.Context, id string) (*EnterpriseCapabilityLinkDTO, error) {
	return rm.querySingle(ctx, id, LinkQueryByID)
}

func (rm *EnterpriseCapabilityLinkReadModel) GetByEnterpriseCapabilityID(ctx context.Context, enterpriseCapabilityID string) ([]EnterpriseCapabilityLinkDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var links []EnterpriseCapabilityLinkDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT ecl.id, ecl.enterprise_capability_id, ecl.domain_capability_id, ecl.linked_by, ecl.linked_at,
			        dcm.capability_name, dcm.business_domain_id, dcm.business_domain_name
			 FROM enterprisearchitecture.enterprise_capability_links ecl
			 JOIN enterprisearchitecture.domain_capability_metadata dcm
			     ON dcm.capability_id = ecl.domain_capability_id
			     AND dcm.tenant_id = ecl.tenant_id
			 WHERE ecl.tenant_id = $1 AND ecl.enterprise_capability_id = $2
			 ORDER BY dcm.business_domain_name NULLS LAST, dcm.capability_name`,
			tenantID.Value(), enterpriseCapabilityID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto EnterpriseCapabilityLinkDTO
			var domainCapName sql.NullString
			var businessDomainID, businessDomainName sql.NullString
			if err := rows.Scan(
				&dto.ID, &dto.EnterpriseCapabilityID, &dto.DomainCapabilityID,
				&dto.LinkedBy, &dto.LinkedAt,
				&domainCapName, &businessDomainID, &businessDomainName,
			); err != nil {
				return err
			}
			if domainCapName.Valid {
				dto.DomainCapabilityName = domainCapName.String
			}
			if businessDomainID.Valid {
				dto.BusinessDomainID = businessDomainID.String
			}
			if businessDomainName.Valid {
				dto.BusinessDomainName = businessDomainName.String
			}
			links = append(links, dto)
		}

		return rows.Err()
	})

	return links, err
}

func (rm *EnterpriseCapabilityLinkReadModel) GetLinksForCapabilities(ctx context.Context, capabilityIDs []string) ([]EnterpriseCapabilityLinkDTO, error) {
	if len(capabilityIDs) == 0 {
		return nil, nil
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	args := buildAnyClauseArgs(tenantID.Value(), capabilityIDs)

	var links []EnterpriseCapabilityLinkDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := `SELECT id, enterprise_capability_id, domain_capability_id, linked_by, linked_at
				  FROM enterprisearchitecture.enterprise_capability_links
				  WHERE tenant_id = $1 AND domain_capability_id = ANY($2)`

		rows, err := tx.QueryContext(ctx, query, args...)
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
