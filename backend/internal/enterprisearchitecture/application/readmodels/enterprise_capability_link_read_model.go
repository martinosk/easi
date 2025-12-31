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

type HierarchyConflict struct {
	ConflictingCapabilityID   string
	ConflictingCapabilityName string
	LinkedToCapabilityID      string
	LinkedToCapabilityName    string
	IsAncestor                bool
}

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
	Links                   map[string]string `json:"_links,omitempty"`
}

type LinkedCapability struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (rm *EnterpriseCapabilityLinkReadModel) CheckHierarchyConflict(ctx context.Context, domainCapabilityID string, targetEnterpriseCapabilityID string) (*HierarchyConflict, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var conflict *HierarchyConflict
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := `
		WITH RECURSIVE ancestors AS (
			SELECT id, name, parent_id, 1 as depth
			FROM capabilities
			WHERE tenant_id = $1 AND id = $2
			UNION ALL
			SELECT c.id, c.name, c.parent_id, a.depth + 1
			FROM capabilities c
			INNER JOIN ancestors a ON c.id = a.parent_id AND c.tenant_id = $1
			WHERE a.depth < 10
		),
		descendants AS (
			SELECT id, name, parent_id, 1 as depth
			FROM capabilities
			WHERE tenant_id = $1 AND id = $2
			UNION ALL
			SELECT c.id, c.name, c.parent_id, d.depth + 1
			FROM capabilities c
			INNER JOIN descendants d ON c.parent_id = d.id AND c.tenant_id = $1
			WHERE d.depth < 10
		),
		related AS (
			SELECT id, name, TRUE as is_ancestor FROM ancestors WHERE id != $2
			UNION
			SELECT id, name, FALSE as is_ancestor FROM descendants WHERE id != $2
		)
		SELECT r.id, r.name, r.is_ancestor, ecl.enterprise_capability_id, ec.name
		FROM related r
		INNER JOIN enterprise_capability_links ecl ON ecl.domain_capability_id = r.id AND ecl.tenant_id = $1
		INNER JOIN enterprise_capabilities ec ON ec.id = ecl.enterprise_capability_id AND ec.tenant_id = $1
		WHERE ecl.enterprise_capability_id != $3
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
		var enterpriseCapID, enterpriseCapName string
		err := tx.QueryRowContext(ctx,
			`SELECT ecl.enterprise_capability_id, ec.name
			 FROM enterprise_capability_links ecl
			 JOIN enterprise_capabilities ec ON ec.id = ecl.enterprise_capability_id AND ec.tenant_id = $1
			 WHERE ecl.tenant_id = $1 AND ecl.domain_capability_id = $2`,
			tenantID.Value(), domainCapabilityID,
		).Scan(&enterpriseCapID, &enterpriseCapName)

		if err == nil {
			result.Status = LinkStatusLinked
			result.LinkedTo = &LinkedCapability{ID: enterpriseCapID, Name: enterpriseCapName}
			return nil
		}
		if err != sql.ErrNoRows {
			return err
		}

		hierarchyQuery := `
		WITH RECURSIVE ancestors AS (
			SELECT id, name, parent_id, 1 as depth
			FROM capabilities
			WHERE tenant_id = $1 AND id = $2
			UNION ALL
			SELECT c.id, c.name, c.parent_id, a.depth + 1
			FROM capabilities c
			INNER JOIN ancestors a ON c.id = a.parent_id AND c.tenant_id = $1
			WHERE a.depth < 10
		),
		descendants AS (
			SELECT id, name, parent_id, 1 as depth
			FROM capabilities
			WHERE tenant_id = $1 AND id = $2
			UNION ALL
			SELECT c.id, c.name, c.parent_id, d.depth + 1
			FROM capabilities c
			INNER JOIN descendants d ON c.parent_id = d.id AND c.tenant_id = $1
			WHERE d.depth < 10
		),
		related AS (
			SELECT id, name, TRUE as is_ancestor FROM ancestors WHERE id != $2
			UNION
			SELECT id, name, FALSE as is_ancestor FROM descendants WHERE id != $2
		)
		SELECT r.id, r.name, r.is_ancestor, ecl.enterprise_capability_id, ec.name
		FROM related r
		INNER JOIN enterprise_capability_links ecl ON ecl.domain_capability_id = r.id AND ecl.tenant_id = $1
		INNER JOIN enterprise_capabilities ec ON ec.id = ecl.enterprise_capability_id AND ec.tenant_id = $1
		LIMIT 1`

		var conflictingID, conflictingName, linkedToID, linkedToName string
		var isAncestor bool

		err = tx.QueryRowContext(ctx, hierarchyQuery, tenantID.Value(), domainCapabilityID).Scan(
			&conflictingID, &conflictingName, &isAncestor, &linkedToID, &linkedToName,
		)

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
		result.BlockingEnterpriseCapID = &linkedToID

		return nil
	})

	return result, err
}

func (rm *EnterpriseCapabilityLinkReadModel) GetBatchLinkStatus(ctx context.Context, domainCapabilityIDs []string) ([]CapabilityLinkStatusDTO, error) {
	results := make([]CapabilityLinkStatusDTO, 0, len(domainCapabilityIDs))
	for _, id := range domainCapabilityIDs {
		status, err := rm.GetLinkStatus(ctx, id)
		if err != nil {
			return nil, err
		}
		results = append(results, *status)
	}
	return results, nil
}
