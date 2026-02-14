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
	Capabilities                    int `json:"capabilities"`
	Components                      int `json:"components"`
	ValueStreams                    int `json:"valueStreams"`
	ParentChildRelationships        int `json:"parentChildRelationships"`
	Realizations                    int `json:"realizations"`
	ComponentRelationships          int `json:"componentRelationships"`
	CapabilityToValueStreamMappings int `json:"capabilityToValueStreamMappings"`
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
	ID                string                    `json:"id"`
	SourceFormat      string                    `json:"sourceFormat"`
	BusinessDomainID  string                    `json:"businessDomainId,omitempty"`
	CapabilityEAOwner string                    `json:"capabilityEAOwner,omitempty"`
	Status            string                    `json:"status"`
	Preview           *PreviewDTO               `json:"preview,omitempty"`
	Progress          *ProgressDTO              `json:"progress,omitempty"`
	Result            *ResultDTO                `json:"result,omitempty"`
	CreatedAt         time.Time                 `json:"createdAt"`
	CompletedAt       *time.Time                `json:"completedAt,omitempty"`
	Links             map[string]sharedAPI.Link `json:"_links,omitempty"`
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
		`INSERT INTO import_sessions (id, tenant_id, source_format, business_domain_id, capability_ea_owner, status, preview, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		dto.ID, tenantID.Value(), dto.SourceFormat, nullString(dto.BusinessDomainID), nullString(dto.CapabilityEAOwner), dto.Status, previewJSON, dto.CreatedAt,
	)
	return err
}

func (rm *ImportSessionReadModel) UpdateStatus(ctx context.Context, id, status string) error {
	return rm.execUpdate(ctx, id, "UPDATE import_sessions SET status = $1 WHERE tenant_id = $2 AND id = $3", status)
}

func (rm *ImportSessionReadModel) UpdateProgress(ctx context.Context, id string, progress ProgressDTO) error {
	progressJSON, err := json.Marshal(progress)
	if err != nil {
		return err
	}
	return rm.execUpdate(ctx, id, "UPDATE import_sessions SET progress = $1 WHERE tenant_id = $2 AND id = $3", progressJSON)
}

func (rm *ImportSessionReadModel) MarkCompleted(ctx context.Context, id string, result ResultDTO, completedAt time.Time) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return rm.execUpdateWithTime(ctx, id, "UPDATE import_sessions SET status = 'completed', result = $1, completed_at = $2 WHERE tenant_id = $3 AND id = $4", resultJSON, completedAt)
}

func (rm *ImportSessionReadModel) MarkFailed(ctx context.Context, id string, failedAt time.Time) error {
	return rm.execUpdateWithTime(ctx, id, "UPDATE import_sessions SET status = 'failed', completed_at = $1 WHERE tenant_id = $2 AND id = $3", nil, failedAt)
}

func (rm *ImportSessionReadModel) MarkCancelled(ctx context.Context, id string) error {
	return rm.execUpdate(ctx, id, "UPDATE import_sessions SET is_cancelled = TRUE WHERE tenant_id = $1 AND id = $2", nil)
}

func (rm *ImportSessionReadModel) execUpdate(ctx context.Context, id string, query string, value any) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	if value != nil {
		_, err = rm.db.ExecContext(ctx, query, value, tenantID.Value(), id)
	} else {
		_, err = rm.db.ExecContext(ctx, query, tenantID.Value(), id)
	}
	return err
}

func (rm *ImportSessionReadModel) execUpdateWithTime(ctx context.Context, id string, query string, value any, timestamp time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	if value != nil {
		_, err = rm.db.ExecContext(ctx, query, value, timestamp, tenantID.Value(), id)
	} else {
		_, err = rm.db.ExecContext(ctx, query, timestamp, tenantID.Value(), id)
	}
	return err
}

type importSessionRow struct {
	dto               ImportSessionDTO
	businessDomainID  sql.NullString
	capabilityEAOwner sql.NullString
	completedAt       sql.NullTime
	previewJSON       []byte
	progressJSON      []byte
	resultJSON        []byte
}

func (r *importSessionRow) scanTargets() []any {
	return []any{
		&r.dto.ID, &r.dto.SourceFormat, &r.businessDomainID, &r.capabilityEAOwner,
		&r.dto.Status, &r.previewJSON, &r.progressJSON, &r.resultJSON,
		&r.dto.CreatedAt, &r.completedAt,
	}
}

func (r *importSessionRow) toDTO() *ImportSessionDTO {
	r.dto.BusinessDomainID = r.businessDomainID.String
	r.dto.CapabilityEAOwner = r.capabilityEAOwner.String
	if r.completedAt.Valid {
		r.dto.CompletedAt = &r.completedAt.Time
	}
	r.dto.Preview = unmarshalJSON[PreviewDTO](r.previewJSON)
	r.dto.Progress = unmarshalJSON[ProgressDTO](r.progressJSON)
	r.dto.Result = unmarshalJSON[ResultDTO](r.resultJSON)
	return &r.dto
}

func unmarshalJSON[T any](data []byte) *T {
	if len(data) == 0 {
		return nil
	}
	var result T
	if json.Unmarshal(data, &result) == nil {
		return &result
	}
	return nil
}

func (rm *ImportSessionReadModel) GetByID(ctx context.Context, id string) (*ImportSessionDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var row importSessionRow
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT id, source_format, business_domain_id, capability_ea_owner, status, preview, progress, result, created_at, completed_at
			 FROM import_sessions
			 WHERE tenant_id = $1 AND id = $2 AND is_cancelled = FALSE`,
			tenantID.Value(), id,
		).Scan(row.scanTargets()...)

		if err == sql.ErrNoRows {
			notFound = true
			return nil
		}
		return err
	})

	if err != nil || notFound {
		return nil, err
	}

	return row.toDTO(), nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
