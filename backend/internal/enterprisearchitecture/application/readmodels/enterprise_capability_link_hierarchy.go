package readmodels

import (
	"context"
	"database/sql"

	sharedctx "easi/backend/internal/shared/context"
)

type HierarchyDirection int

const (
	HierarchyAncestors HierarchyDirection = iota
	HierarchyDescendants
	HierarchySubtree
)

var hierarchyQueries = map[HierarchyDirection]string{
	HierarchyAncestors: `
		WITH RECURSIVE cte AS (
			SELECT capability_id, parent_id, 1 as depth
			FROM enterprisearchitecture.domain_capability_metadata
			WHERE tenant_id = $1 AND capability_id = $2
			UNION ALL
			SELECT m.capability_id, m.parent_id, c.depth + 1
			FROM enterprisearchitecture.domain_capability_metadata m
			INNER JOIN cte c ON m.capability_id = c.parent_id AND m.tenant_id = $1
			WHERE c.depth < 10
		)
		SELECT capability_id FROM cte WHERE capability_id != $2`,
	HierarchyDescendants: `
		WITH RECURSIVE cte AS (
			SELECT capability_id, 1 as depth
			FROM enterprisearchitecture.domain_capability_metadata
			WHERE tenant_id = $1 AND capability_id = $2
			UNION ALL
			SELECT m.capability_id, c.depth + 1
			FROM enterprisearchitecture.domain_capability_metadata m
			INNER JOIN cte c ON m.parent_id = c.capability_id AND m.tenant_id = $1
			WHERE c.depth < 10
		)
		SELECT capability_id FROM cte WHERE capability_id != $2`,
	HierarchySubtree: `
		WITH RECURSIVE cte AS (
			SELECT capability_id, 1 as depth
			FROM enterprisearchitecture.domain_capability_metadata
			WHERE tenant_id = $1 AND capability_id = $2
			UNION ALL
			SELECT m.capability_id, c.depth + 1
			FROM enterprisearchitecture.domain_capability_metadata m
			INNER JOIN cte c ON m.parent_id = c.capability_id AND m.tenant_id = $1
			WHERE c.depth < 10
		)
		SELECT capability_id FROM cte`,
}

func (rm *EnterpriseCapabilityLinkReadModel) QueryHierarchy(ctx context.Context, capabilityID string, direction HierarchyDirection) ([]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var result []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, hierarchyQueries[direction], tenantID.Value(), capabilityID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			result = append(result, id)
		}
		return rows.Err()
	})

	return result, err
}

type NameKind int

const (
	NameDomainCapability NameKind = iota
	NameEnterpriseCapability
)

var nameQueries = map[NameKind]string{
	NameDomainCapability:     `SELECT capability_name FROM enterprisearchitecture.domain_capability_metadata WHERE tenant_id = $1 AND capability_id = $2`,
	NameEnterpriseCapability: `SELECT name FROM enterprisearchitecture.enterprise_capabilities WHERE tenant_id = $1 AND id = $2`,
}

func (rm *EnterpriseCapabilityLinkReadModel) QueryName(ctx context.Context, id string, kind NameKind) (string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}

	var name string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, nameQueries[kind], tenantID.Value(), id).Scan(&name)
	})
	return name, err
}
