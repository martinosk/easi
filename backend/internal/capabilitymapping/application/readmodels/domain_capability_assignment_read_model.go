package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type AssignmentDTO struct {
	AssignmentID       string            `json:"assignmentId"`
	BusinessDomainID   string            `json:"businessDomainId"`
	BusinessDomainName string            `json:"businessDomainName"`
	CapabilityID       string            `json:"capabilityId"`
	CapabilityName     string            `json:"capabilityName"`
	CapabilityLevel    string            `json:"capabilityLevel"`
	AssignedAt         time.Time         `json:"assignedAt"`
	Links              map[string]string `json:"_links,omitempty"`
}

type DomainCapabilityAssignmentReadModel struct {
	db *database.TenantAwareDB
}

func NewDomainCapabilityAssignmentReadModel(db *database.TenantAwareDB) *DomainCapabilityAssignmentReadModel {
	return &DomainCapabilityAssignmentReadModel{db: db}
}

func (rm *DomainCapabilityAssignmentReadModel) Insert(ctx context.Context, dto AssignmentDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO domain_capability_assignments (assignment_id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_name, capability_level, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		dto.AssignmentID, tenantID.Value(), dto.BusinessDomainID, dto.BusinessDomainName, dto.CapabilityID, dto.CapabilityName, dto.CapabilityLevel, dto.AssignedAt,
	)
	return err
}

func (rm *DomainCapabilityAssignmentReadModel) Delete(ctx context.Context, assignmentID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM domain_capability_assignments WHERE tenant_id = $1 AND assignment_id = $2",
		tenantID.Value(), assignmentID,
	)
	return err
}

const assignmentSelectColumns = "assignment_id, business_domain_id, business_domain_name, capability_id, capability_name, capability_level, assigned_at"

func (rm *DomainCapabilityAssignmentReadModel) GetByDomainID(ctx context.Context, domainID string) ([]AssignmentDTO, error) {
	query := "SELECT " + assignmentSelectColumns + " FROM domain_capability_assignments WHERE tenant_id = $1 AND business_domain_id = $2 ORDER BY capability_name"
	return rm.queryAssignments(ctx, query, domainID)
}

func (rm *DomainCapabilityAssignmentReadModel) GetByCapabilityID(ctx context.Context, capabilityID string) ([]AssignmentDTO, error) {
	query := "SELECT " + assignmentSelectColumns + " FROM domain_capability_assignments WHERE tenant_id = $1 AND capability_id = $2 ORDER BY business_domain_name"
	return rm.queryAssignments(ctx, query, capabilityID)
}

func (rm *DomainCapabilityAssignmentReadModel) queryAssignments(ctx context.Context, query, param string) ([]AssignmentDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var assignments []AssignmentDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, tenantID.Value(), param)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto AssignmentDTO
			if err := rows.Scan(&dto.AssignmentID, &dto.BusinessDomainID, &dto.BusinessDomainName, &dto.CapabilityID, &dto.CapabilityName, &dto.CapabilityLevel, &dto.AssignedAt); err != nil {
				return err
			}
			assignments = append(assignments, dto)
		}

		return rows.Err()
	})

	return assignments, err
}

func (rm *DomainCapabilityAssignmentReadModel) GetByDomainAndCapability(ctx context.Context, domainID, capabilityID string) (*AssignmentDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto AssignmentDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := "SELECT " + assignmentSelectColumns + " FROM domain_capability_assignments WHERE tenant_id = $1 AND business_domain_id = $2 AND capability_id = $3"
		err := tx.QueryRowContext(ctx, query, tenantID.Value(), domainID, capabilityID).Scan(
			&dto.AssignmentID, &dto.BusinessDomainID, &dto.BusinessDomainName, &dto.CapabilityID, &dto.CapabilityName, &dto.CapabilityLevel, &dto.AssignedAt,
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

func (rm *DomainCapabilityAssignmentReadModel) AssignmentExists(ctx context.Context, domainID, capabilityID string) (bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return false, err
	}

	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM domain_capability_assignments WHERE tenant_id = $1 AND business_domain_id = $2 AND capability_id = $3",
			tenantID.Value(), domainID, capabilityID,
		).Scan(&count)
	})

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
