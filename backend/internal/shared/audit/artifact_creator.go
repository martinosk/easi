package audit

import (
	"context"
	"database/sql"
	"fmt"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type ArtifactCreator struct {
	AggregateID string `json:"aggregateId"`
	CreatorID   string `json:"creatorId"`
}

type ArtifactCreatorsResponse struct {
	Data  []ArtifactCreator `json:"data"`
	Links map[string]any    `json:"_links"`
}

type ArtifactCreatorReader interface {
	GetArtifactCreators(ctx context.Context) ([]ArtifactCreator, error)
}

type ArtifactCreatorReadModel struct {
	db *database.TenantAwareDB
}

func NewArtifactCreatorReadModel(db *database.TenantAwareDB) *ArtifactCreatorReadModel {
	return &ArtifactCreatorReadModel{db: db}
}

func (rm *ArtifactCreatorReadModel) GetArtifactCreators(ctx context.Context) ([]ArtifactCreator, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant from context: %w", err)
	}

	var creators []ArtifactCreator

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT DISTINCT aggregate_id, actor_id
			FROM events
			WHERE tenant_id = $1
			  AND version = 1
			  AND event_type IN (
			    'ApplicationComponentCreated',
			    'CapabilityCreated',
			    'VendorCreated',
			    'InternalTeamCreated',
			    'AcquiredEntityCreated'
			  )
		`, tenantID.Value())
		if err != nil {
			return fmt.Errorf("failed to query artifact creators: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var creator ArtifactCreator
			if err := rows.Scan(&creator.AggregateID, &creator.CreatorID); err != nil {
				return fmt.Errorf("failed to scan artifact creator: %w", err)
			}
			creators = append(creators, creator)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, err
	}

	return creators, nil
}
