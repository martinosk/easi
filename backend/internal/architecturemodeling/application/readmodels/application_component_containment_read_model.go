package readmodels

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	sharedctx "easi/backend/internal/shared/context"
)

var ErrContainmentsAggregateConflict = errors.New("tenant already has a component containments aggregate")

type ContainmentParentDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type ContainmentPartDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type ContainmentRecord struct {
	PartID   string
	ParentID string
	Kind     string
}

func (rm *ApplicationComponentReadModel) SetParent(ctx context.Context, record ContainmentRecord) error {
	return rm.execByID(ctx,
		"UPDATE architecturemodeling.application_components SET parent_component_id = $3, containment_kind = $4 WHERE tenant_id = $1 AND id = $2",
		record.PartID, record.ParentID, record.Kind,
	)
}

func (rm *ApplicationComponentReadModel) ClearParent(ctx context.Context, partID string) error {
	return rm.execByID(ctx,
		"UPDATE architecturemodeling.application_components SET parent_component_id = NULL, containment_kind = NULL WHERE tenant_id = $1 AND id = $2",
		partID,
	)
}

func (rm *ApplicationComponentReadModel) RegisterContainmentsAggregate(ctx context.Context, aggregateID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("resolve tenant for register containments aggregate %s: %w", aggregateID, err)
	}
	if _, err := rm.db.ExecContext(ctx,
		"INSERT INTO architecturemodeling.component_containment_aggregates (tenant_id, aggregate_id) VALUES ($1, $2) ON CONFLICT (tenant_id) DO NOTHING",
		tenantID.Value(), aggregateID,
	); err != nil {
		return fmt.Errorf("register containments aggregate %s for tenant %s: %w", aggregateID, tenantID.Value(), err)
	}
	registered, found, err := rm.FindContainmentsAggregateID(ctx)
	if err != nil {
		return err
	}
	if !found || registered != aggregateID {
		return ErrContainmentsAggregateConflict
	}
	return nil
}

func (rm *ApplicationComponentReadModel) FindContainmentsAggregateID(ctx context.Context) (string, bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", false, fmt.Errorf("resolve tenant for find containments aggregate: %w", err)
	}
	var id string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			"SELECT aggregate_id FROM architecturemodeling.component_containment_aggregates WHERE tenant_id = $1",
			tenantID.Value(),
		).Scan(&id)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("find containments aggregate for tenant %s: %w", tenantID.Value(), err)
	}
	return id, true, nil
}

func (rm *ApplicationComponentReadModel) PartsOf(ctx context.Context, parentID string) ([]ContainmentPartDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve tenant for parts of component %s: %w", parentID, err)
	}
	var parts []ContainmentPartDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		partsByParent, err := rm.fetchPartsByParent(ctx, tx, tenantID.Value(), []string{parentID})
		parts = partsByParent[parentID]
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("parts of component %s for tenant %s: %w", parentID, tenantID.Value(), err)
	}
	return parts, nil
}

func (rm *ApplicationComponentReadModel) fetchPartsByParent(ctx context.Context, tx *sql.Tx, tenantID string, parentIDs []string) (map[string][]ContainmentPartDTO, error) {
	rows, err := tx.QueryContext(ctx,
		"SELECT parent_component_id, id, name, containment_kind FROM architecturemodeling.application_components WHERE tenant_id = $1 AND parent_component_id = ANY($2) AND is_deleted = FALSE ORDER BY LOWER(name) ASC, id ASC",
		tenantID, pq.Array(parentIDs),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	partsByParent := map[string][]ContainmentPartDTO{}
	for rows.Next() {
		var parentID string
		var part ContainmentPartDTO
		if err := rows.Scan(&parentID, &part.ID, &part.Name, &part.Kind); err != nil {
			return nil, err
		}
		partsByParent[parentID] = append(partsByParent[parentID], part)
	}
	return partsByParent, rows.Err()
}

func (rm *ApplicationComponentReadModel) loadPartsForComponents(ctx context.Context, tx *sql.Tx, tenantID string, components []ApplicationComponentDTO) error {
	if len(components) == 0 {
		return nil
	}
	ids := make([]string, len(components))
	for i, c := range components {
		ids[i] = c.ID
	}
	partsByParent, err := rm.fetchPartsByParent(ctx, tx, tenantID, ids)
	if err != nil {
		return err
	}
	for i := range components {
		components[i].Parts = partsByParent[components[i].ID]
	}
	return nil
}
