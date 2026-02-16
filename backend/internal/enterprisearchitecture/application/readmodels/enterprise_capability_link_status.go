package readmodels

import (
	"context"
	"database/sql"

	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type LinkStatus string

const (
	LinkStatusAvailable       LinkStatus = "available"
	LinkStatusLinked          LinkStatus = "linked"
	LinkStatusBlockedByParent LinkStatus = "blocked_by_parent"
	LinkStatusBlockedByChild  LinkStatus = "blocked_by_child"
)

type CapabilityLinkStatusDTO struct {
	CapabilityID            string            `json:"capabilityId"`
	Status                  LinkStatus        `json:"status"`
	LinkedTo                *LinkedCapability `json:"linkedTo,omitempty"`
	BlockingCapability      *LinkedCapability `json:"blockingCapability,omitempty"`
	BlockingEnterpriseCapID *string           `json:"blockingEnterpriseCapabilityId,omitempty"`
	Links                   types.Links       `json:"_links,omitempty"`
}

type LinkedCapability struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type statusLookup struct {
	tx                 *sql.Tx
	tenantID           string
	domainCapabilityID string
}

func (rm *EnterpriseCapabilityLinkReadModel) GetLinkStatus(ctx context.Context, domainCapabilityID string) (*CapabilityLinkStatusDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	result := &CapabilityLinkStatusDTO{
		CapabilityID: domainCapabilityID,
		Status:       LinkStatusAvailable,
	}

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		lookup := statusLookup{tx: tx, tenantID: tenantID.Value(), domainCapabilityID: domainCapabilityID}
		if found, err := rm.scanLinkedStatus(ctx, lookup, result); err != nil || found {
			return err
		}
		return rm.scanBlockingStatus(ctx, lookup, result)
	})

	return result, err
}

func (rm *EnterpriseCapabilityLinkReadModel) scanLinkedStatus(ctx context.Context, lookup statusLookup, result *CapabilityLinkStatusDTO) (bool, error) {
	var enterpriseCapID, enterpriseCapName string
	err := lookup.tx.QueryRowContext(ctx,
		`SELECT ecl.enterprise_capability_id, ec.name
		 FROM enterprisearchitecture.enterprise_capability_links ecl
		 JOIN enterprisearchitecture.enterprise_capabilities ec ON ec.id = ecl.enterprise_capability_id AND ec.tenant_id = $1
		 WHERE ecl.tenant_id = $1 AND ecl.domain_capability_id = $2`,
		lookup.tenantID, lookup.domainCapabilityID,
	).Scan(&enterpriseCapID, &enterpriseCapName)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	result.Status = LinkStatusLinked
	result.LinkedTo = &LinkedCapability{ID: enterpriseCapID, Name: enterpriseCapName}
	return true, nil
}

func (rm *EnterpriseCapabilityLinkReadModel) scanBlockingStatus(ctx context.Context, lookup statusLookup, result *CapabilityLinkStatusDTO) error {
	var conflictingID, conflictingName, enterpriseCapID string
	var isAncestor bool

	err := lookup.tx.QueryRowContext(ctx,
		`SELECT blocked_by_capability_id, blocked_by_capability_name, is_ancestor, blocked_by_enterprise_id
		 FROM enterprisearchitecture.capability_link_blocking
		 WHERE tenant_id = $1 AND domain_capability_id = $2
		 LIMIT 1`,
		lookup.tenantID, lookup.domainCapabilityID,
	).Scan(&conflictingID, &conflictingName, &isAncestor, &enterpriseCapID)

	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	if isAncestor {
		result.Status = LinkStatusBlockedByParent
	} else {
		result.Status = LinkStatusBlockedByChild
	}
	result.BlockingCapability = &LinkedCapability{ID: conflictingID, Name: conflictingName}
	result.BlockingEnterpriseCapID = &enterpriseCapID
	return nil
}

func (rm *EnterpriseCapabilityLinkReadModel) GetBatchLinkStatus(ctx context.Context, domainCapabilityIDs []string) ([]CapabilityLinkStatusDTO, error) {
	if len(domainCapabilityIDs) == 0 {
		return nil, nil
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	args := buildAnyClauseArgs(tenantID.Value(), domainCapabilityIDs)
	statusMap := initializeStatusMap(domainCapabilityIDs)

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		if err := rm.populateLinkedStatus(ctx, tx, args, statusMap); err != nil {
			return err
		}
		return rm.populateBlockingStatus(ctx, tx, args, statusMap)
	})

	if err != nil {
		return nil, err
	}
	return collectResults(domainCapabilityIDs, statusMap), nil
}

func initializeStatusMap(ids []string) map[string]*CapabilityLinkStatusDTO {
	statusMap := make(map[string]*CapabilityLinkStatusDTO, len(ids))
	for _, id := range ids {
		statusMap[id] = &CapabilityLinkStatusDTO{
			CapabilityID: id,
			Status:       LinkStatusAvailable,
		}
	}
	return statusMap
}

func (rm *EnterpriseCapabilityLinkReadModel) populateLinkedStatus(ctx context.Context, tx *sql.Tx, args []any, statusMap map[string]*CapabilityLinkStatusDTO) error {
	query := `
		SELECT ecl.domain_capability_id, ecl.enterprise_capability_id, ec.name
		FROM enterprisearchitecture.enterprise_capability_links ecl
		JOIN enterprisearchitecture.enterprise_capabilities ec ON ec.id = ecl.enterprise_capability_id AND ec.tenant_id = $1
		WHERE ecl.tenant_id = $1 AND ecl.domain_capability_id = ANY($2)`

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var domainCapID, enterpriseCapID, enterpriseCapName string
		if err := rows.Scan(&domainCapID, &enterpriseCapID, &enterpriseCapName); err != nil {
			return err
		}
		if status, ok := statusMap[domainCapID]; ok {
			status.Status = LinkStatusLinked
			status.LinkedTo = &LinkedCapability{ID: enterpriseCapID, Name: enterpriseCapName}
		}
	}
	return rows.Err()
}

func (rm *EnterpriseCapabilityLinkReadModel) populateBlockingStatus(ctx context.Context, tx *sql.Tx, args []any, statusMap map[string]*CapabilityLinkStatusDTO) error {
	query := `
		SELECT domain_capability_id, blocked_by_capability_id, blocked_by_capability_name,
		       is_ancestor, blocked_by_enterprise_id
		FROM enterprisearchitecture.capability_link_blocking
		WHERE tenant_id = $1 AND domain_capability_id = ANY($2)`

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var domainCapID, blockingCapID, blockingCapName, enterpriseCapID string
		var isAncestor bool
		if err := rows.Scan(&domainCapID, &blockingCapID, &blockingCapName, &isAncestor, &enterpriseCapID); err != nil {
			return err
		}
		status, ok := statusMap[domainCapID]
		if !ok || status.Status == LinkStatusLinked {
			continue
		}
		if isAncestor {
			status.Status = LinkStatusBlockedByParent
		} else {
			status.Status = LinkStatusBlockedByChild
		}
		status.BlockingCapability = &LinkedCapability{ID: blockingCapID, Name: blockingCapName}
		status.BlockingEnterpriseCapID = &enterpriseCapID
	}
	return rows.Err()
}

func collectResults(ids []string, statusMap map[string]*CapabilityLinkStatusDTO) []CapabilityLinkStatusDTO {
	results := make([]CapabilityLinkStatusDTO, 0, len(ids))
	for _, id := range ids {
		results = append(results, *statusMap[id])
	}
	return results
}
