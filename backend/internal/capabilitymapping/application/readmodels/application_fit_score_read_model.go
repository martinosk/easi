package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type ApplicationFitScoreDTO struct {
	ID            string      `json:"id"`
	ComponentID   string      `json:"componentId"`
	ComponentName string      `json:"componentName"`
	PillarID      string      `json:"pillarId"`
	PillarName    string      `json:"pillarName"`
	Score         int         `json:"score"`
	ScoreLabel    string      `json:"scoreLabel"`
	Rationale     string      `json:"rationale,omitempty"`
	ScoredAt      time.Time   `json:"scoredAt"`
	ScoredBy      string      `json:"scoredBy"`
	UpdatedAt     *time.Time  `json:"updatedAt,omitempty"`
	Links         types.Links `json:"_links,omitempty"`
}

type ApplicationFitScoreReadModel struct {
	db *database.TenantAwareDB
}

func NewApplicationFitScoreReadModel(db *database.TenantAwareDB) *ApplicationFitScoreReadModel {
	return &ApplicationFitScoreReadModel{db: db}
}

func (rm *ApplicationFitScoreReadModel) Insert(ctx context.Context, dto ApplicationFitScoreDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM capabilitymapping.application_fit_scores WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), dto.ID,
	)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO capabilitymapping.application_fit_scores
		(id, tenant_id, component_id, component_name, pillar_id, pillar_name,
		score, score_label, rationale, scored_at, scored_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		dto.ID, tenantID.Value(), dto.ComponentID, dto.ComponentName,
		dto.PillarID, dto.PillarName, dto.Score, dto.ScoreLabel, dto.Rationale,
		dto.ScoredAt, dto.ScoredBy,
	)
	return err
}

func (rm *ApplicationFitScoreReadModel) Update(ctx context.Context, dto ApplicationFitScoreDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE capabilitymapping.application_fit_scores
		SET score = $1, score_label = $2, rationale = $3, updated_at = $4
		WHERE tenant_id = $5 AND id = $6`,
		dto.Score, dto.ScoreLabel, dto.Rationale, time.Now().UTC(), tenantID.Value(), dto.ID,
	)
	return err
}

func (rm *ApplicationFitScoreReadModel) Delete(ctx context.Context, id string) error {
	return rm.deleteByColumn(ctx, "id", id)
}

func (rm *ApplicationFitScoreReadModel) DeleteByComponent(ctx context.Context, componentID string) error {
	return rm.deleteByColumn(ctx, "component_id", componentID)
}

func (rm *ApplicationFitScoreReadModel) deleteByColumn(ctx context.Context, column, value string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM capabilitymapping.application_fit_scores WHERE tenant_id = $1 AND "+column+" = $2",
		tenantID.Value(), value,
	)
	return err
}

const fitScoreSelectColumns = `id, component_id, component_name, pillar_id, pillar_name,
score, score_label, rationale, scored_at, scored_by, updated_at`

func (rm *ApplicationFitScoreReadModel) GetByID(ctx context.Context, id string) (*ApplicationFitScoreDTO, error) {
	return rm.querySingleFitScore(ctx,
		"SELECT "+fitScoreSelectColumns+" FROM capabilitymapping.application_fit_scores WHERE tenant_id = $1 AND id = $2",
		id,
	)
}

func (rm *ApplicationFitScoreReadModel) GetByComponentID(ctx context.Context, componentID string) ([]ApplicationFitScoreDTO, error) {
	query := "SELECT " + fitScoreSelectColumns + " FROM capabilitymapping.application_fit_scores WHERE tenant_id = $1 AND component_id = $2 ORDER BY pillar_name"
	return rm.queryFitScoreList(ctx, query, componentID)
}

func (rm *ApplicationFitScoreReadModel) GetByComponentAndPillar(ctx context.Context, componentID, pillarID string) (*ApplicationFitScoreDTO, error) {
	return rm.querySingleFitScore(ctx,
		"SELECT "+fitScoreSelectColumns+" FROM capabilitymapping.application_fit_scores WHERE tenant_id = $1 AND component_id = $2 AND pillar_id = $3",
		componentID, pillarID,
	)
}

func (rm *ApplicationFitScoreReadModel) buildTenantArgs(ctx context.Context, params ...string) ([]any, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	args := make([]any, 0, len(params)+1)
	args = append(args, tenantID.Value())
	for _, p := range params {
		args = append(args, p)
	}
	return args, nil
}

func (rm *ApplicationFitScoreReadModel) querySingleFitScore(ctx context.Context, query string, params ...string) (*ApplicationFitScoreDTO, error) {
	args, err := rm.buildTenantArgs(ctx, params...)
	if err != nil {
		return nil, err
	}

	var result *ApplicationFitScoreDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		dto, scanErr := rm.scanFitScoreRow(tx.QueryRowContext(ctx, query, args...))
		if scanErr == sql.ErrNoRows {
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		result = &dto
		return nil
	})

	return result, err
}

func (rm *ApplicationFitScoreReadModel) Exists(ctx context.Context, componentID, pillarID string) (bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return false, err
	}

	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM capabilitymapping.application_fit_scores WHERE tenant_id = $1 AND component_id = $2 AND pillar_id = $3",
			tenantID.Value(), componentID, pillarID,
		).Scan(&count)
	})

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (rm *ApplicationFitScoreReadModel) GetByPillarID(ctx context.Context, pillarID string) ([]ApplicationFitScoreDTO, error) {
	query := "SELECT " + fitScoreSelectColumns + " FROM capabilitymapping.application_fit_scores WHERE tenant_id = $1 AND pillar_id = $2 ORDER BY component_name"
	return rm.queryFitScoreList(ctx, query, pillarID)
}

func (rm *ApplicationFitScoreReadModel) queryFitScoreList(ctx context.Context, query string, params ...string) ([]ApplicationFitScoreDTO, error) {
	args, err := rm.buildTenantArgs(ctx, params...)
	if err != nil {
		return nil, err
	}

	var results []ApplicationFitScoreDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			dto, scanErr := rm.scanFitScoreRow(rows)
			if scanErr != nil {
				return scanErr
			}
			results = append(results, dto)
		}
		return rows.Err()
	})

	return results, err
}

type fitScoreRowScanner interface {
	Scan(dest ...any) error
}

func (rm *ApplicationFitScoreReadModel) scanFitScoreRow(row fitScoreRowScanner) (ApplicationFitScoreDTO, error) {
	var dto ApplicationFitScoreDTO
	var rationale sql.NullString
	var updatedAt sql.NullTime

	err := row.Scan(&dto.ID, &dto.ComponentID, &dto.ComponentName, &dto.PillarID, &dto.PillarName,
		&dto.Score, &dto.ScoreLabel, &rationale, &dto.ScoredAt, &dto.ScoredBy, &updatedAt)
	if err != nil {
		return dto, err
	}

	if rationale.Valid {
		dto.Rationale = rationale.String
	}
	if updatedAt.Valid {
		dto.UpdatedAt = &updatedAt.Time
	}
	return dto, nil
}
