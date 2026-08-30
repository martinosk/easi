package readmodels

import (
	"context"
	"database/sql"

	domainservices "easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type CapabilityNodeDTO struct {
	CapabilityID       string
	CapabilityName     string
	CapabilityLevel    string
	ParentID           string
	L1CapabilityID     string
	BusinessDomainID   string
	BusinessDomainName string
}

type ParentL1Update struct {
	CapabilityID      string
	NewParentID       string
	NewLevel          string
	NewL1CapabilityID string
}

type BusinessDomainRef struct {
	ID   string
	Name string
}

type CapabilityNodeCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewCapabilityNodeCacheReadModel(db *database.TenantAwareDB) *CapabilityNodeCacheReadModel {
	return &CapabilityNodeCacheReadModel{db: db}
}

func (rm *CapabilityNodeCacheReadModel) Insert(ctx context.Context, dto CapabilityNodeDTO) error {
	return rm.execForTenant(ctx,
		`INSERT INTO architecturedirection.capability_node_cache
		 (tenant_id, capability_id, capability_name, capability_level, parent_id, l1_capability_id, business_domain_id, business_domain_name)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (tenant_id, capability_id) DO UPDATE SET
		 capability_name = EXCLUDED.capability_name,
		 capability_level = EXCLUDED.capability_level,
		 parent_id = EXCLUDED.parent_id,
		 l1_capability_id = EXCLUDED.l1_capability_id,
		 business_domain_id = EXCLUDED.business_domain_id,
		 business_domain_name = EXCLUDED.business_domain_name`,
		dto.CapabilityID, dto.CapabilityName, dto.CapabilityLevel,
		nullIfEmpty(dto.ParentID), dto.L1CapabilityID,
		nullIfEmpty(dto.BusinessDomainID), nullIfEmpty(dto.BusinessDomainName),
	)
}

func (rm *CapabilityNodeCacheReadModel) Delete(ctx context.Context, capabilityID string) error {
	return rm.execForTenant(ctx,
		`DELETE FROM architecturedirection.capability_node_cache WHERE tenant_id = $1 AND capability_id = $2`,
		capabilityID,
	)
}

func (rm *CapabilityNodeCacheReadModel) GetByID(ctx context.Context, capabilityID string) (*CapabilityNodeDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	var dto CapabilityNodeDTO
	var parentID, businessDomainID, businessDomainName sql.NullString
	var notFound bool
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT capability_id, capability_name, capability_level, parent_id, l1_capability_id, business_domain_id, business_domain_name
			 FROM architecturedirection.capability_node_cache WHERE tenant_id = $1 AND capability_id = $2`,
			tenantID.Value(), capabilityID,
		).Scan(&dto.CapabilityID, &dto.CapabilityName, &dto.CapabilityLevel, &parentID, &dto.L1CapabilityID, &businessDomainID, &businessDomainName)
		if err == sql.ErrNoRows {
			notFound = true
			return nil
		}
		return err
	})
	if err != nil || notFound {
		return nil, err
	}
	dto.ParentID = parentID.String
	dto.BusinessDomainID = businessDomainID.String
	dto.BusinessDomainName = businessDomainName.String
	return &dto, nil
}

func (rm *CapabilityNodeCacheReadModel) AllCapabilityNodes(ctx context.Context) ([]domainservices.CapabilityNode, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	var nodes []domainservices.CapabilityNode
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT capability_id, capability_name, capability_level, parent_id, business_domain_id, business_domain_name
			 FROM architecturedirection.capability_node_cache
			 WHERE tenant_id = $1 ORDER BY capability_name`,
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var node domainservices.CapabilityNode
			var parentID, businessDomainID, businessDomainName sql.NullString
			if err := rows.Scan(&node.ID, &node.Name, &node.Level, &parentID, &businessDomainID, &businessDomainName); err != nil {
				return err
			}
			node.ParentID = parentID.String
			node.BusinessDomainID = businessDomainID.String
			node.BusinessDomainName = businessDomainName.String
			nodes = append(nodes, node)
		}
		return rows.Err()
	})
	return nodes, err
}

func (rm *CapabilityNodeCacheReadModel) UpdateParentAndL1(ctx context.Context, update ParentL1Update) error {
	return rm.execForTenant(ctx,
		`UPDATE architecturedirection.capability_node_cache
		 SET parent_id = $2, capability_level = $3, l1_capability_id = $4
		 WHERE tenant_id = $1 AND capability_id = $5`,
		nullIfEmpty(update.NewParentID), update.NewLevel, update.NewL1CapabilityID, update.CapabilityID,
	)
}

func (rm *CapabilityNodeCacheReadModel) UpdateLevel(ctx context.Context, capabilityID, newLevel string) error {
	return rm.execForTenant(ctx,
		`UPDATE architecturedirection.capability_node_cache SET capability_level = $2 WHERE tenant_id = $1 AND capability_id = $3`,
		newLevel, capabilityID,
	)
}

func (rm *CapabilityNodeCacheReadModel) UpdateBusinessDomainForL1Subtree(ctx context.Context, l1CapabilityID string, domain BusinessDomainRef) error {
	return rm.execForTenant(ctx,
		`UPDATE architecturedirection.capability_node_cache
		 SET business_domain_id = $2, business_domain_name = $3
		 WHERE tenant_id = $1 AND l1_capability_id = $4`,
		nullIfEmpty(domain.ID), nullIfEmpty(domain.Name), l1CapabilityID,
	)
}

func (rm *CapabilityNodeCacheReadModel) UpdateBusinessDomainNameForDomain(ctx context.Context, businessDomainID, name string) error {
	return rm.execForTenant(ctx,
		`UPDATE architecturedirection.capability_node_cache SET business_domain_name = $2 WHERE tenant_id = $1 AND business_domain_id = $3`,
		name, businessDomainID,
	)
}

func (rm *CapabilityNodeCacheReadModel) UpdateMaturityValue(ctx context.Context, capabilityID string, maturityValue int) error {
	return rm.execForTenant(ctx,
		`UPDATE architecturedirection.capability_node_cache SET maturity_value = $2 WHERE tenant_id = $1 AND capability_id = $3`,
		maturityValue, capabilityID,
	)
}

func (rm *CapabilityNodeCacheReadModel) RecalculateL1ForSubtree(ctx context.Context, capabilityID string) error {
	root, err := rm.GetByID(ctx, capabilityID)
	if err != nil || root == nil {
		return err
	}
	subtreeIDs, err := rm.subtreeCapabilityIDs(ctx, capabilityID)
	if err != nil {
		return err
	}
	l1ID := rm.l1AncestorOf(ctx, root)
	domain, err := rm.businessDomainOf(ctx, l1ID)
	if err != nil {
		return err
	}
	for _, id := range subtreeIDs {
		if err := rm.execForTenant(ctx,
			`UPDATE architecturedirection.capability_node_cache
			 SET l1_capability_id = $2, business_domain_id = $3, business_domain_name = $4
			 WHERE tenant_id = $1 AND capability_id = $5`,
			l1ID, nullIfEmpty(domain.ID), nullIfEmpty(domain.Name), id,
		); err != nil {
			return err
		}
	}
	return nil
}

func (rm *CapabilityNodeCacheReadModel) BusinessDomainName(ctx context.Context, businessDomainID string) (string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}
	var name string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT name FROM architecturedirection.reference_name_cache
			 WHERE tenant_id = $1 AND entity_type = 'business_domain' AND entity_id = $2`,
			tenantID.Value(), businessDomainID,
		).Scan(&name)
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	})
	return name, err
}

func (rm *CapabilityNodeCacheReadModel) subtreeCapabilityIDs(ctx context.Context, rootID string) ([]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	var ids []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`WITH RECURSIVE subtree AS (
				SELECT capability_id, 1 AS depth FROM architecturedirection.capability_node_cache
				WHERE tenant_id = $1 AND capability_id = $2
				UNION ALL
				SELECT n.capability_id, s.depth + 1 FROM architecturedirection.capability_node_cache n
				INNER JOIN subtree s ON n.parent_id = s.capability_id AND n.tenant_id = $1
				WHERE s.depth < 10
			)
			SELECT capability_id FROM subtree`,
			tenantID.Value(), rootID,
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			ids = append(ids, id)
		}
		return rows.Err()
	})
	return ids, err
}

func (rm *CapabilityNodeCacheReadModel) l1AncestorOf(ctx context.Context, node *CapabilityNodeDTO) string {
	current := node
	for depth := 0; depth < 10; depth++ {
		if current.CapabilityLevel == "L1" || current.ParentID == "" {
			return current.CapabilityID
		}
		parent, err := rm.GetByID(ctx, current.ParentID)
		if err != nil || parent == nil {
			return node.CapabilityID
		}
		current = parent
	}
	return node.CapabilityID
}

func (rm *CapabilityNodeCacheReadModel) businessDomainOf(ctx context.Context, capabilityID string) (BusinessDomainRef, error) {
	node, err := rm.GetByID(ctx, capabilityID)
	if err != nil || node == nil {
		return BusinessDomainRef{}, err
	}
	return BusinessDomainRef{ID: node.BusinessDomainID, Name: node.BusinessDomainName}, nil
}

func (rm *CapabilityNodeCacheReadModel) execForTenant(ctx context.Context, query string, args ...any) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, append([]any{tenantID.Value()}, args...)...)
	return err
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
