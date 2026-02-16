package readmodels

import (
	"context"
	"database/sql"
	"fmt"

	sharedctx "easi/backend/internal/shared/context"
)

type BlockingDTO struct {
	DomainCapabilityID      string
	BlockedByCapabilityID   string
	BlockedByEnterpriseID   string
	BlockedByCapabilityName string
	BlockedByEnterpriseName string
	IsAncestor              bool
}

type HierarchyConflict struct {
	ConflictingCapabilityID   string
	ConflictingCapabilityName string
	LinkedToCapabilityID      string
	LinkedToCapabilityName    string
	IsAncestor                bool
}

func (rm *EnterpriseCapabilityLinkReadModel) InsertBlocking(ctx context.Context, blocking BlockingDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("resolve tenant for insert blocking record domain capability %s: %w", blocking.DomainCapabilityID, err)
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO enterprisearchitecture.capability_link_blocking
		 (tenant_id, domain_capability_id, blocked_by_capability_id, blocked_by_enterprise_id,
		  blocked_by_capability_name, blocked_by_enterprise_name, is_ancestor)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (tenant_id, domain_capability_id, blocked_by_capability_id) DO NOTHING`,
		tenantID.Value(), blocking.DomainCapabilityID, blocking.BlockedByCapabilityID,
		blocking.BlockedByEnterpriseID, blocking.BlockedByCapabilityName, blocking.BlockedByEnterpriseName,
		blocking.IsAncestor,
	)
	if err != nil {
		return fmt.Errorf("insert blocking record for domain capability %s blocked by %s tenant %s: %w", blocking.DomainCapabilityID, blocking.BlockedByCapabilityID, tenantID.Value(), err)
	}
	return nil
}

func (rm *EnterpriseCapabilityLinkReadModel) DeleteBlockingByBlocker(ctx context.Context, blockedByCapabilityID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("resolve tenant for delete blocking by blocker %s: %w", blockedByCapabilityID, err)
	}

	_, err = rm.db.ExecContext(ctx,
		`DELETE FROM enterprisearchitecture.capability_link_blocking WHERE tenant_id = $1 AND blocked_by_capability_id = $2`,
		tenantID.Value(), blockedByCapabilityID,
	)
	if err != nil {
		return fmt.Errorf("delete blocking records for blocker capability %s tenant %s: %w", blockedByCapabilityID, tenantID.Value(), err)
	}
	return nil
}

func (rm *EnterpriseCapabilityLinkReadModel) DeleteBlockingForCapabilities(ctx context.Context, capabilityIDs []string) error {
	if len(capabilityIDs) == 0 {
		return nil
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("resolve tenant for delete blocking for %d capabilities: %w", len(capabilityIDs), err)
	}

	args := buildAnyClauseArgs(tenantID.Value(), capabilityIDs)
	query := `DELETE FROM enterprisearchitecture.capability_link_blocking WHERE tenant_id = $1 AND blocked_by_capability_id = ANY($2)`
	_, err = rm.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete blocking records for %d capabilities tenant %s: %w", len(capabilityIDs), tenantID.Value(), err)
	}
	return nil
}

func (rm *EnterpriseCapabilityLinkReadModel) CheckHierarchyConflict(ctx context.Context, domainCapabilityID string, targetEnterpriseCapabilityID string) (*HierarchyConflict, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var conflict *HierarchyConflict
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := `
		SELECT blocked_by_capability_id, blocked_by_capability_name, is_ancestor,
		       blocked_by_enterprise_id, blocked_by_enterprise_name
		FROM enterprisearchitecture.capability_link_blocking
		WHERE tenant_id = $1 AND domain_capability_id = $2
		  AND blocked_by_enterprise_id != $3
		LIMIT 1`

		var conflictingID, conflictingName, linkedToID, linkedToName string
		var isAncestor bool

		err := tx.QueryRowContext(ctx, query, tenantID.Value(), domainCapabilityID, targetEnterpriseCapabilityID).Scan(
			&conflictingID, &conflictingName, &isAncestor, &linkedToID, &linkedToName,
		)
		if err == sql.ErrNoRows {
			return nil
		}
		if err != nil {
			return err
		}

		conflict = &HierarchyConflict{
			ConflictingCapabilityID:   conflictingID,
			ConflictingCapabilityName: conflictingName,
			LinkedToCapabilityID:      linkedToID,
			LinkedToCapabilityName:    linkedToName,
			IsAncestor:                isAncestor,
		}
		return nil
	})

	return conflict, err
}
