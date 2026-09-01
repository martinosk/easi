package readmodels

import (
	"context"
	"database/sql"
	"fmt"

	sharedctx "easi/backend/internal/shared/context"
)

type OwnershipRecord struct {
	State     string
	OwnerKind string
	OwnerID   string
}

func (rm *ApplicationComponentReadModel) SetOwnership(ctx context.Context, componentID string, record OwnershipRecord) error {
	return rm.execByID(ctx,
		"UPDATE architecturemodeling.application_components SET ownership_state = $3, owner_kind = $4, owner_id = $5 WHERE tenant_id = $1 AND id = $2",
		componentID, record.State, record.OwnerKind, record.OwnerID,
	)
}

func (rm *ApplicationComponentReadModel) ClearOwnership(ctx context.Context, componentID string) error {
	return rm.execByID(ctx,
		"UPDATE architecturemodeling.application_components SET ownership_state = 'unknown', owner_kind = NULL, owner_id = NULL WHERE tenant_id = $1 AND id = $2",
		componentID,
	)
}

func (rm *ApplicationComponentReadModel) FindComponentIDsByTeamOwner(ctx context.Context, teamID string) ([]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve tenant for find components owned by team %s: %w", teamID, err)
	}

	var componentIDs []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id FROM architecturemodeling.application_components WHERE tenant_id = $1 AND owner_kind = 'team' AND owner_id = $2 AND is_deleted = FALSE",
			tenantID.Value(), teamID,
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
			componentIDs = append(componentIDs, id)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("find components owned by team %s for tenant %s: %w", teamID, tenantID.Value(), err)
	}
	return componentIDs, nil
}
