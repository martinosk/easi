package readmodels

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

type PreviewDTO struct {
	Supported   SupportedCountsDTO   `json:"supported"`
	Unsupported UnsupportedCountsDTO `json:"unsupported"`
}

type SupportedCountsDTO struct {
	Capabilities             int `json:"capabilities"`
	Components               int `json:"components"`
	ParentChildRelationships int `json:"parentChildRelationships"`
	Realizations             int `json:"realizations"`
	ComponentRelationships   int `json:"componentRelationships"`
}

type UnsupportedCountsDTO struct {
	Elements      map[string]int `json:"elements"`
	Relationships map[string]int `json:"relationships"`
}

type ProgressDTO struct {
	Phase          string `json:"phase"`
	TotalItems     int    `json:"totalItems"`
	CompletedItems int    `json:"completedItems"`
}

type ImportErrorDTO struct {
	SourceElement string `json:"sourceElement"`
	SourceName    string `json:"sourceName"`
	Error         string `json:"error"`
	Action        string `json:"action"`
}

type ResultDTO struct {
	CapabilitiesCreated       int              `json:"capabilitiesCreated"`
	ComponentsCreated         int              `json:"componentsCreated"`
	RealizationsCreated       int              `json:"realizationsCreated"`
	ComponentRelationsCreated int              `json:"componentRelationsCreated"`
	DomainAssignments         int              `json:"domainAssignments"`
	Errors                    []ImportErrorDTO `json:"errors"`
}

type ImportSessionDTO struct {
	ID               string                    `json:"id"`
	SourceFormat     string                    `json:"sourceFormat"`
	BusinessDomainID string                    `json:"businessDomainId,omitempty"`
	Status           string                    `json:"status"`
	Preview          *PreviewDTO               `json:"preview,omitempty"`
	Progress         *ProgressDTO              `json:"progress,omitempty"`
	Result           *ResultDTO                `json:"result,omitempty"`
	CreatedAt        time.Time                 `json:"createdAt"`
	CompletedAt      *time.Time                `json:"completedAt,omitempty"`
	Links            map[string]sharedAPI.Link `json:"_links,omitempty"`
}

type ImportSessionReadModel struct {
	db *database.TenantAwareDB
}

func NewImportSessionReadModel(db *database.TenantAwareDB) *ImportSessionReadModel {
	return &ImportSessionReadModel{db: db}
}

func (rm *ImportSessionReadModel) Insert(ctx context.Context, dto ImportSessionDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	previewJSON, err := json.Marshal(dto.Preview)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO import_sessions (id, tenant_id, source_format, business_domain_id, status, preview, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		dto.ID, tenantID.Value(), dto.SourceFormat, nullString(dto.BusinessDomainID), dto.Status, previewJSON, dto.CreatedAt,
	)
	return err
}

func (rm *ImportSessionReadModel) UpdateStatus(ctx context.Context, id, status string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE import_sessions SET status = $1 WHERE tenant_id = $2 AND id = $3",
		status, tenantID.Value(), id,
	)
	return err
}

func (rm *ImportSessionReadModel) UpdateProgress(ctx context.Context, id string, progress ProgressDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	progressJSON, err := json.Marshal(progress)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE import_sessions SET progress = $1 WHERE tenant_id = $2 AND id = $3",
		progressJSON, tenantID.Value(), id,
	)
	return err
}

func (rm *ImportSessionReadModel) MarkCompleted(ctx context.Context, id string, result ResultDTO, completedAt time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE import_sessions SET status = 'completed', result = $1, completed_at = $2 WHERE tenant_id = $3 AND id = $4",
		resultJSON, completedAt, tenantID.Value(), id,
	)
	return err
}

func (rm *ImportSessionReadModel) MarkFailed(ctx context.Context, id string, failedAt time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE import_sessions SET status = 'failed', completed_at = $1 WHERE tenant_id = $2 AND id = $3",
		failedAt, tenantID.Value(), id,
	)
	return err
}

func (rm *ImportSessionReadModel) MarkCancelled(ctx context.Context, id string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE import_sessions SET is_cancelled = TRUE WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), id,
	)
	return err
}

func (rm *ImportSessionReadModel) GetByID(ctx context.Context, id string) (*ImportSessionDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto ImportSessionDTO
	var notFound bool
	var businessDomainID sql.NullString
	var previewJSON, progressJSON, resultJSON []byte
	var completedAt sql.NullTime

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT id, source_format, business_domain_id, status, preview, progress, result, created_at, completed_at
			 FROM import_sessions
			 WHERE tenant_id = $1 AND id = $2 AND is_cancelled = FALSE`,
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.SourceFormat, &businessDomainID, &dto.Status, &previewJSON, &progressJSON, &resultJSON, &dto.CreatedAt, &completedAt)

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

	if businessDomainID.Valid {
		dto.BusinessDomainID = businessDomainID.String
	}
	if completedAt.Valid {
		dto.CompletedAt = &completedAt.Time
	}

	if len(previewJSON) > 0 {
		var preview PreviewDTO
		if err := json.Unmarshal(previewJSON, &preview); err == nil {
			dto.Preview = &preview
		}
	}

	if len(progressJSON) > 0 {
		var progress ProgressDTO
		if err := json.Unmarshal(progressJSON, &progress); err == nil {
			dto.Progress = &progress
		}
	}

	if len(resultJSON) > 0 {
		var result ResultDTO
		if err := json.Unmarshal(resultJSON, &result); err == nil {
			dto.Result = &result
		}
	}

	return &dto, nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
