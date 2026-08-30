package readmodels

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type ArtifactRef struct {
	ArtifactType string
	ArtifactID   string
}

type ArtifactNameDTO struct {
	ArtifactType string
	ArtifactID   string
	Name         string
}

type ArtifactNameCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewArtifactNameCacheReadModel(db *database.TenantAwareDB) *ArtifactNameCacheReadModel {
	return &ArtifactNameCacheReadModel{db: db}
}

func (rm *ArtifactNameCacheReadModel) Upsert(ctx context.Context, dto ArtifactNameDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO accessdelegation.artifact_name_cache (tenant_id, artifact_type, artifact_id, name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (tenant_id, artifact_type, artifact_id) DO UPDATE SET name = EXCLUDED.name`,
		tenantID.Value(), dto.ArtifactType, dto.ArtifactID, dto.Name,
	)
	if err != nil {
		return fmt.Errorf("cache name of %s %s: %w", dto.ArtifactType, dto.ArtifactID, err)
	}
	return nil
}

func (rm *ArtifactNameCacheReadModel) Delete(ctx context.Context, artifactType, artifactID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`DELETE FROM accessdelegation.artifact_name_cache
		 WHERE tenant_id = $1 AND artifact_type = $2 AND artifact_id = $3`,
		tenantID.Value(), artifactType, artifactID,
	)
	if err != nil {
		return fmt.Errorf("remove cached name of %s %s: %w", artifactType, artifactID, err)
	}
	return nil
}

func (rm *ArtifactNameCacheReadModel) NamesByIDs(ctx context.Context, artifactType string, artifactIDs []string) (map[string]string, error) {
	names := map[string]string{}
	if len(artifactIDs) == 0 {
		return names, nil
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT artifact_id, name FROM accessdelegation.artifact_name_cache
			 WHERE tenant_id = $1 AND artifact_type = $2 AND artifact_id = ANY($3)`,
			tenantID.Value(), artifactType, pq.Array(artifactIDs),
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var artifactID, name string
			if err := rows.Scan(&artifactID, &name); err != nil {
				return err
			}
			names[artifactID] = name
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("read cached names of %s artifacts: %w", artifactType, err)
	}
	return names, nil
}
