package readmodels

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type DomainCapabilityMetadataDTO struct {
	CapabilityID       string
	CapabilityName     string
	CapabilityLevel    string
	ParentID           string
	L1CapabilityID     string
	BusinessDomainID   string
	BusinessDomainName string
}

type DomainCapabilityMetadataReadModel struct {
	db *database.TenantAwareDB
}

func NewDomainCapabilityMetadataReadModel(db *database.TenantAwareDB) *DomainCapabilityMetadataReadModel {
	return &DomainCapabilityMetadataReadModel{db: db}
}

func (rm *DomainCapabilityMetadataReadModel) Insert(ctx context.Context, dto DomainCapabilityMetadataDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO domain_capability_metadata
		 (tenant_id, capability_id, capability_name, capability_level, parent_id, l1_capability_id, business_domain_id, business_domain_name)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (tenant_id, capability_id) DO UPDATE SET
		 capability_name = EXCLUDED.capability_name,
		 capability_level = EXCLUDED.capability_level,
		 parent_id = EXCLUDED.parent_id,
		 l1_capability_id = EXCLUDED.l1_capability_id,
		 business_domain_id = EXCLUDED.business_domain_id,
		 business_domain_name = EXCLUDED.business_domain_name`,
		tenantID.Value(), dto.CapabilityID, dto.CapabilityName, dto.CapabilityLevel,
		nullIfEmpty(dto.ParentID), dto.L1CapabilityID,
		nullIfEmpty(dto.BusinessDomainID), nullIfEmpty(dto.BusinessDomainName),
	)
	return err
}

func (rm *DomainCapabilityMetadataReadModel) Delete(ctx context.Context, capabilityID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM domain_capability_metadata WHERE tenant_id = $1 AND capability_id = $2",
		tenantID.Value(), capabilityID,
	)
	return err
}

func (rm *DomainCapabilityMetadataReadModel) GetByID(ctx context.Context, capabilityID string) (*DomainCapabilityMetadataDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto DomainCapabilityMetadataDTO
	var parentID, businessDomainID, businessDomainName sql.NullString
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT capability_id, capability_name, capability_level, parent_id, l1_capability_id, business_domain_id, business_domain_name
			 FROM domain_capability_metadata WHERE tenant_id = $1 AND capability_id = $2`,
			tenantID.Value(), capabilityID,
		).Scan(&dto.CapabilityID, &dto.CapabilityName, &dto.CapabilityLevel, &parentID, &dto.L1CapabilityID, &businessDomainID, &businessDomainName)

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

	dto.ParentID = parentID.String
	dto.BusinessDomainID = businessDomainID.String
	dto.BusinessDomainName = businessDomainName.String

	return &dto, nil
}

func (rm *DomainCapabilityMetadataReadModel) GetCapabilityName(ctx context.Context, capabilityID string) (string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}

	var name string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			"SELECT capability_name FROM domain_capability_metadata WHERE tenant_id = $1 AND capability_id = $2",
			tenantID.Value(), capabilityID,
		).Scan(&name)
	})

	return name, err
}

func (rm *DomainCapabilityMetadataReadModel) GetAncestorIDs(ctx context.Context, capabilityID string) ([]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var ancestors []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := `
		WITH RECURSIVE ancestors AS (
			SELECT capability_id, parent_id, 1 as depth
			FROM domain_capability_metadata
			WHERE tenant_id = $1 AND capability_id = $2
			UNION ALL
			SELECT m.capability_id, m.parent_id, a.depth + 1
			FROM domain_capability_metadata m
			INNER JOIN ancestors a ON m.capability_id = a.parent_id AND m.tenant_id = $1
			WHERE a.depth < 10
		)
		SELECT capability_id FROM ancestors WHERE capability_id != $2`

		rows, err := tx.QueryContext(ctx, query, tenantID.Value(), capabilityID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			ancestors = append(ancestors, id)
		}
		return rows.Err()
	})

	return ancestors, err
}

func (rm *DomainCapabilityMetadataReadModel) GetDescendantIDs(ctx context.Context, capabilityID string) ([]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var descendants []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := `
		WITH RECURSIVE descendants AS (
			SELECT capability_id, 1 as depth
			FROM domain_capability_metadata
			WHERE tenant_id = $1 AND capability_id = $2
			UNION ALL
			SELECT m.capability_id, d.depth + 1
			FROM domain_capability_metadata m
			INNER JOIN descendants d ON m.parent_id = d.capability_id AND m.tenant_id = $1
			WHERE d.depth < 10
		)
		SELECT capability_id FROM descendants WHERE capability_id != $2`

		rows, err := tx.QueryContext(ctx, query, tenantID.Value(), capabilityID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			descendants = append(descendants, id)
		}
		return rows.Err()
	})

	return descendants, err
}

func (rm *DomainCapabilityMetadataReadModel) GetSubtreeCapabilityIDs(ctx context.Context, rootID string) ([]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var subtree []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := `
		WITH RECURSIVE subtree AS (
			SELECT capability_id, 1 as depth
			FROM domain_capability_metadata
			WHERE tenant_id = $1 AND capability_id = $2
			UNION ALL
			SELECT m.capability_id, s.depth + 1
			FROM domain_capability_metadata m
			INNER JOIN subtree s ON m.parent_id = s.capability_id AND m.tenant_id = $1
			WHERE s.depth < 10
		)
		SELECT capability_id FROM subtree`

		rows, err := tx.QueryContext(ctx, query, tenantID.Value(), rootID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			subtree = append(subtree, id)
		}
		return rows.Err()
	})

	return subtree, err
}

func (rm *DomainCapabilityMetadataReadModel) UpdateBusinessDomainForL1Subtree(ctx context.Context, l1CapabilityID, businessDomainID, businessDomainName string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE domain_capability_metadata
		 SET business_domain_id = $1, business_domain_name = $2
		 WHERE tenant_id = $3 AND l1_capability_id = $4`,
		nullIfEmpty(businessDomainID), nullIfEmpty(businessDomainName), tenantID.Value(), l1CapabilityID,
	)
	return err
}

func (rm *DomainCapabilityMetadataReadModel) UpdateParentAndL1(ctx context.Context, capabilityID, newParentID, newLevel, newL1CapabilityID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE domain_capability_metadata
		 SET parent_id = $1, capability_level = $2, l1_capability_id = $3
		 WHERE tenant_id = $4 AND capability_id = $5`,
		nullIfEmpty(newParentID), newLevel, newL1CapabilityID, tenantID.Value(), capabilityID,
	)
	return err
}

func (rm *DomainCapabilityMetadataReadModel) GetBusinessDomainForL1(ctx context.Context, l1CapabilityID string) (string, string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", "", err
	}

	var businessDomainID, businessDomainName sql.NullString
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT business_domain_id, business_domain_name
			 FROM domain_capability_metadata
			 WHERE tenant_id = $1 AND capability_id = $2`,
			tenantID.Value(), l1CapabilityID,
		).Scan(&businessDomainID, &businessDomainName)
	})

	if err == sql.ErrNoRows {
		return "", "", nil
	}
	if err != nil {
		return "", "", err
	}

	return businessDomainID.String, businessDomainName.String, nil
}

func (rm *DomainCapabilityMetadataReadModel) GetEnterpriseCapabilitiesLinkedToCapabilities(ctx context.Context, capabilityIDs []string) ([]string, error) {
	if len(capabilityIDs) == 0 {
		return nil, nil
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	placeholders := make([]string, len(capabilityIDs))
	args := make([]interface{}, len(capabilityIDs)+1)
	args[0] = tenantID.Value()
	for i, id := range capabilityIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	var enterpriseCapabilityIDs []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := fmt.Sprintf(`
			SELECT DISTINCT enterprise_capability_id
			FROM enterprise_capability_links
			WHERE tenant_id = $1 AND domain_capability_id IN (%s)`,
			strings.Join(placeholders, ", "))

		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			enterpriseCapabilityIDs = append(enterpriseCapabilityIDs, id)
		}
		return rows.Err()
	})

	return enterpriseCapabilityIDs, err
}

func (rm *DomainCapabilityMetadataReadModel) RecalculateL1ForSubtree(ctx context.Context, capabilityID string) error {
	subtreeIDs, err := rm.GetSubtreeCapabilityIDs(ctx, capabilityID)
	if err != nil {
		return err
	}

	root, err := rm.GetByID(ctx, capabilityID)
	if err != nil {
		return err
	}
	if root == nil {
		return nil
	}

	newL1ID := rm.findL1Ancestor(ctx, capabilityID, root.CapabilityLevel)
	businessDomainID, businessDomainName, _ := rm.GetBusinessDomainForL1(ctx, newL1ID)

	for _, id := range subtreeIDs {
		if err := rm.updateL1AndBusinessDomain(ctx, id, newL1ID, businessDomainID, businessDomainName); err != nil {
			return err
		}
	}

	return nil
}

func (rm *DomainCapabilityMetadataReadModel) findL1Ancestor(ctx context.Context, capabilityID, level string) string {
	if level == "L1" {
		return capabilityID
	}

	current, err := rm.GetByID(ctx, capabilityID)
	if err != nil || current == nil || current.ParentID == "" {
		return capabilityID
	}

	for depth := 0; depth < 10 && current.ParentID != ""; depth++ {
		parent, err := rm.GetByID(ctx, current.ParentID)
		if err != nil || parent == nil {
			break
		}
		if parent.CapabilityLevel == "L1" {
			return parent.CapabilityID
		}
		current = parent
	}

	return capabilityID
}

func (rm *DomainCapabilityMetadataReadModel) updateL1AndBusinessDomain(ctx context.Context, capabilityID, l1CapabilityID, businessDomainID, businessDomainName string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE domain_capability_metadata
		 SET l1_capability_id = $1, business_domain_id = $2, business_domain_name = $3
		 WHERE tenant_id = $4 AND capability_id = $5`,
		l1CapabilityID, nullIfEmpty(businessDomainID), nullIfEmpty(businessDomainName), tenantID.Value(), capabilityID,
	)
	return err
}

func (rm *DomainCapabilityMetadataReadModel) LookupBusinessDomainName(ctx context.Context, businessDomainID string) (string, error) {
	if businessDomainID == "" {
		return "", nil
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}

	var name string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			"SELECT business_domain_name FROM domain_capability_assignments WHERE tenant_id = $1 AND business_domain_id = $2 LIMIT 1",
			tenantID.Value(), businessDomainID,
		).Scan(&name)
	})

	if err == sql.ErrNoRows {
		return businessDomainID, nil
	}
	if err != nil {
		return businessDomainID, err
	}

	return name, nil
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
