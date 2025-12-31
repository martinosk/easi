package readmodels

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

func buildInClauseArgs(tenantID interface{}, ids []string) (placeholders string, args []interface{}) {
	placeholderList := make([]string, len(ids))
	args = make([]interface{}, len(ids)+1)
	args[0] = tenantID
	for i, id := range ids {
		placeholderList[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}
	return strings.Join(placeholderList, ", "), args
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
		SELECT blocked_by_capability_id, blocked_by_capability_name, is_ancestor,
		       blocked_by_enterprise_id, blocked_by_enterprise_name
		FROM capability_link_blocking
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

		blockingQuery := `
		SELECT blocked_by_capability_id, blocked_by_capability_name, is_ancestor,
		       blocked_by_enterprise_id, blocked_by_enterprise_name
		FROM capability_link_blocking
		WHERE tenant_id = $1 AND domain_capability_id = $2
		LIMIT 1`

		var conflictingID, conflictingName, linkedToID, linkedToName string
		var isAncestor bool

		err = tx.QueryRowContext(ctx, blockingQuery, tenantID.Value(), domainCapabilityID).Scan(
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
	if len(domainCapabilityIDs) == 0 {
		return nil, nil
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	inClause, args := buildInClauseArgs(tenantID.Value(), domainCapabilityIDs)
	statusMap := initializeStatusMap(domainCapabilityIDs)

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		if err := rm.populateLinkedStatus(ctx, tx, inClause, args, statusMap); err != nil {
			return err
		}
		return rm.populateBlockingStatus(ctx, tx, inClause, args, statusMap)
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

func (rm *EnterpriseCapabilityLinkReadModel) populateLinkedStatus(ctx context.Context, tx *sql.Tx, inClause string, args []interface{}, statusMap map[string]*CapabilityLinkStatusDTO) error {
	query := fmt.Sprintf(`
		SELECT ecl.domain_capability_id, ecl.enterprise_capability_id, ec.name
		FROM enterprise_capability_links ecl
		JOIN enterprise_capabilities ec ON ec.id = ecl.enterprise_capability_id AND ec.tenant_id = $1
		WHERE ecl.tenant_id = $1 AND ecl.domain_capability_id IN (%s)`, inClause)

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

func (rm *EnterpriseCapabilityLinkReadModel) populateBlockingStatus(ctx context.Context, tx *sql.Tx, inClause string, args []interface{}, statusMap map[string]*CapabilityLinkStatusDTO) error {
	query := fmt.Sprintf(`
		SELECT domain_capability_id, blocked_by_capability_id, blocked_by_capability_name,
		       is_ancestor, blocked_by_enterprise_id, blocked_by_enterprise_name
		FROM capability_link_blocking
		WHERE tenant_id = $1 AND domain_capability_id IN (%s)`, inClause)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var domainCapID, blockingCapID, blockingCapName, enterpriseCapID, enterpriseCapName string
		var isAncestor bool
		if err := rows.Scan(&domainCapID, &blockingCapID, &blockingCapName, &isAncestor, &enterpriseCapID, &enterpriseCapName); err != nil {
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

type BlockingDTO struct {
	DomainCapabilityID        string
	BlockedByCapabilityID     string
	BlockedByEnterpriseID     string
	BlockedByCapabilityName   string
	BlockedByEnterpriseName   string
	IsAncestor                bool
}

func (rm *EnterpriseCapabilityLinkReadModel) InsertBlocking(ctx context.Context, blocking BlockingDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO capability_link_blocking
		 (tenant_id, domain_capability_id, blocked_by_capability_id, blocked_by_enterprise_id,
		  blocked_by_capability_name, blocked_by_enterprise_name, is_ancestor)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (tenant_id, domain_capability_id, blocked_by_capability_id) DO NOTHING`,
		tenantID.Value(), blocking.DomainCapabilityID, blocking.BlockedByCapabilityID,
		blocking.BlockedByEnterpriseID, blocking.BlockedByCapabilityName, blocking.BlockedByEnterpriseName,
		blocking.IsAncestor,
	)
	return err
}

func (rm *EnterpriseCapabilityLinkReadModel) DeleteBlockingByBlocker(ctx context.Context, blockedByCapabilityID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`DELETE FROM capability_link_blocking WHERE tenant_id = $1 AND blocked_by_capability_id = $2`,
		tenantID.Value(), blockedByCapabilityID,
	)
	return err
}

func (rm *EnterpriseCapabilityLinkReadModel) DeleteBlockingForCapabilities(ctx context.Context, capabilityIDs []string) error {
	if len(capabilityIDs) == 0 {
		return nil
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	inClause, args := buildInClauseArgs(tenantID.Value(), capabilityIDs)
	query := fmt.Sprintf(`DELETE FROM capability_link_blocking WHERE tenant_id = $1 AND blocked_by_capability_id IN (%s)`, inClause)
	_, err = rm.db.ExecContext(ctx, query, args...)
	return err
}

func (rm *EnterpriseCapabilityLinkReadModel) GetAncestorIDs(ctx context.Context, capabilityID string) ([]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var ancestors []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := `
		WITH RECURSIVE ancestors AS (
			SELECT id, parent_id, 1 as depth
			FROM capabilities
			WHERE tenant_id = $1 AND id = $2
			UNION ALL
			SELECT c.id, c.parent_id, a.depth + 1
			FROM capabilities c
			INNER JOIN ancestors a ON c.id = a.parent_id AND c.tenant_id = $1
			WHERE a.depth < 10
		)
		SELECT id FROM ancestors WHERE id != $2`

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

func (rm *EnterpriseCapabilityLinkReadModel) GetDescendantIDs(ctx context.Context, capabilityID string) ([]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var descendants []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := `
		WITH RECURSIVE descendants AS (
			SELECT id, 1 as depth
			FROM capabilities
			WHERE tenant_id = $1 AND id = $2
			UNION ALL
			SELECT c.id, d.depth + 1
			FROM capabilities c
			INNER JOIN descendants d ON c.parent_id = d.id AND c.tenant_id = $1
			WHERE d.depth < 10
		)
		SELECT id FROM descendants WHERE id != $2`

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

func (rm *EnterpriseCapabilityLinkReadModel) GetSubtreeCapabilityIDs(ctx context.Context, rootID string) ([]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var subtree []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := `
		WITH RECURSIVE subtree AS (
			SELECT id, 1 as depth
			FROM capabilities
			WHERE tenant_id = $1 AND id = $2
			UNION ALL
			SELECT c.id, s.depth + 1
			FROM capabilities c
			INNER JOIN subtree s ON c.parent_id = s.id AND c.tenant_id = $1
			WHERE s.depth < 10
		)
		SELECT id FROM subtree`

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

func (rm *EnterpriseCapabilityLinkReadModel) GetCapabilityName(ctx context.Context, capabilityID string) (string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}

	var name string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT name FROM capabilities WHERE tenant_id = $1 AND id = $2`,
			tenantID.Value(), capabilityID,
		).Scan(&name)
	})

	return name, err
}

func (rm *EnterpriseCapabilityLinkReadModel) GetEnterpriseCapabilityName(ctx context.Context, enterpriseCapabilityID string) (string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}

	var name string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT name FROM enterprise_capabilities WHERE tenant_id = $1 AND id = $2`,
			tenantID.Value(), enterpriseCapabilityID,
		).Scan(&name)
	})

	return name, err
}

func (rm *EnterpriseCapabilityLinkReadModel) GetLinksForCapabilities(ctx context.Context, capabilityIDs []string) ([]EnterpriseCapabilityLinkDTO, error) {
	if len(capabilityIDs) == 0 {
		return nil, nil
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	inClause, args := buildInClauseArgs(tenantID.Value(), capabilityIDs)

	var links []EnterpriseCapabilityLinkDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := fmt.Sprintf(`SELECT id, enterprise_capability_id, domain_capability_id, linked_by, linked_at
				  FROM enterprise_capability_links
				  WHERE tenant_id = $1 AND domain_capability_id IN (%s)`, inClause)

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
