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

type MaturitySectionDTO struct {
	Name  string        `json:"name"`
	Order int           `json:"order"`
	Range MaturityRange `json:"range"`
}

type MaturityRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type CapabilityDTO struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Description     string              `json:"description,omitempty"`
	ParentID        string              `json:"parentId,omitempty"`
	Level           string              `json:"level"`
	MaturityValue   int                 `json:"maturityValue"`
	MaturitySection *MaturitySectionDTO `json:"maturitySection,omitempty"`
	OwnershipModel  string              `json:"ownershipModel,omitempty"`
	PrimaryOwner    string              `json:"primaryOwner,omitempty"`
	EAOwner         string              `json:"eaOwner,omitempty"`
	Status          string              `json:"status,omitempty"`
	Experts         []ExpertDTO         `json:"experts,omitempty"`
	Tags            []string            `json:"tags,omitempty"`
	CreatedAt       time.Time           `json:"createdAt"`
	Links           map[string]string   `json:"_links,omitempty"`
}

type ExpertDTO struct {
	Name    string    `json:"name"`
	Role    string    `json:"role"`
	Contact string    `json:"contact"`
	AddedAt time.Time `json:"addedAt"`
}

type CapabilityMetadataUpdate struct {
	MaturityValue  int
	OwnershipModel string
	PrimaryOwner   string
	EAOwner        string
	Status         string
}

type capabilityScanResult struct {
	dto            CapabilityDTO
	parentID       sql.NullString
	maturityValue  sql.NullInt64
	ownershipModel sql.NullString
	primaryOwner   sql.NullString
	eaOwner        sql.NullString
}

type CapabilityReadModel struct {
	db *database.TenantAwareDB
}

func NewCapabilityReadModel(db *database.TenantAwareDB) *CapabilityReadModel {
	return &CapabilityReadModel{db: db}
}

func (rm *CapabilityReadModel) Insert(ctx context.Context, dto CapabilityDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	var parentIDValue interface{} = nil
	if dto.ParentID != "" {
		parentIDValue = dto.ParentID
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO capabilities (id, tenant_id, name, description, parent_id, level, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		dto.ID, tenantID.Value(), dto.Name, dto.Description, parentIDValue, dto.Level, "Initial", "Active", dto.CreatedAt,
	)
	return err
}

func (rm *CapabilityReadModel) UpdateMetadata(ctx context.Context, id string, metadata CapabilityMetadataUpdate) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE capabilities SET maturity_value = $1, ownership_model = $2, primary_owner = $3, ea_owner = $4, status = $5, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $6 AND id = $7",
		metadata.MaturityValue, metadata.OwnershipModel, metadata.PrimaryOwner, metadata.EAOwner, metadata.Status, tenantID.Value(), id,
	)
	return err
}

type ExpertInfo struct {
	CapabilityID string
	Name         string
	Role         string
	Contact      string
	AddedAt      time.Time
}

func (rm *CapabilityReadModel) AddExpert(ctx context.Context, info ExpertInfo) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO capability_experts (capability_id, tenant_id, expert_name, expert_role, contact_info, added_at) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (tenant_id, capability_id, expert_name) DO NOTHING",
		info.CapabilityID, tenantID.Value(), info.Name, info.Role, info.Contact, info.AddedAt,
	)
	return err
}

func (rm *CapabilityReadModel) AddTag(ctx context.Context, capabilityID, tag string, addedAt time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO capability_tags (capability_id, tenant_id, tag, added_at) VALUES ($1, $2, $3, $4) ON CONFLICT (tenant_id, capability_id, tag) DO NOTHING",
		capabilityID, tenantID.Value(), tag, addedAt,
	)
	return err
}

func (rm *CapabilityReadModel) Update(ctx context.Context, id, name, description string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE capabilities SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $3 AND id = $4",
		name, description, tenantID.Value(), id,
	)
	return err
}

func (rm *CapabilityReadModel) UpdateParent(ctx context.Context, id, parentID, level string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	var parentIDValue interface{} = nil
	if parentID != "" {
		parentIDValue = parentID
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE capabilities SET parent_id = $1, level = $2, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $3 AND id = $4",
		parentIDValue, level, tenantID.Value(), id,
	)
	return err
}

func (rm *CapabilityReadModel) Delete(ctx context.Context, id string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM capability_tags WHERE tenant_id = $1 AND capability_id = $2",
		tenantID.Value(), id,
	)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM capability_experts WHERE tenant_id = $1 AND capability_id = $2",
		tenantID.Value(), id,
	)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM capabilities WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), id,
	)
	return err
}

func (rm *CapabilityReadModel) GetByID(ctx context.Context, id string) (*CapabilityDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto *CapabilityDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		result, found, err := scanCapabilityRow(ctx, tx, tenantID.Value(), id)
		if err != nil || !found {
			return err
		}

		result.dto.Experts, err = rm.fetchExperts(ctx, tx, tenantID.Value(), id)
		if err != nil {
			return err
		}

		result.dto.Tags, err = rm.fetchTags(ctx, tx, tenantID.Value(), id)
		if err != nil {
			return err
		}
		dto = result.toDTO()
		return nil
	})
	return dto, err
}

func scanCapabilityRow(ctx context.Context, tx *sql.Tx, tenantID, id string) (*capabilityScanResult, bool, error) {
	var result capabilityScanResult
	err := tx.QueryRowContext(ctx,
		"SELECT id, name, description, parent_id, level, maturity_value, ownership_model, primary_owner, ea_owner, status, created_at FROM capabilities WHERE tenant_id = $1 AND id = $2",
		tenantID, id,
	).Scan(&result.dto.ID, &result.dto.Name, &result.dto.Description, &result.parentID, &result.dto.Level, &result.maturityValue, &result.ownershipModel, &result.primaryOwner, &result.eaOwner, &result.dto.Status, &result.dto.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &result, true, nil
}

func (rm *CapabilityReadModel) fetchExperts(ctx context.Context, tx *sql.Tx, tenantID, capabilityID string) ([]ExpertDTO, error) {
	rows, err := tx.QueryContext(ctx,
		"SELECT expert_name, expert_role, contact_info, added_at FROM capability_experts WHERE tenant_id = $1 AND capability_id = $2",
		tenantID, capabilityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var experts []ExpertDTO
	for rows.Next() {
		var expert ExpertDTO
		if err := rows.Scan(&expert.Name, &expert.Role, &expert.Contact, &expert.AddedAt); err != nil {
			return nil, err
		}
		experts = append(experts, expert)
	}
	return experts, rows.Err()
}

func (rm *CapabilityReadModel) fetchTags(ctx context.Context, tx *sql.Tx, tenantID, capabilityID string) ([]string, error) {
	rows, err := tx.QueryContext(ctx,
		"SELECT tag FROM capability_tags WHERE tenant_id = $1 AND capability_id = $2 ORDER BY tag",
		tenantID, capabilityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *capabilityScanResult) toDTO() *CapabilityDTO {
	dto := &r.dto
	if r.parentID.Valid {
		dto.ParentID = r.parentID.String
	}
	if r.maturityValue.Valid {
		dto.MaturityValue = int(r.maturityValue.Int64)
		dto.MaturitySection = calculateMaturitySection(dto.MaturityValue)
	}
	if r.ownershipModel.Valid {
		dto.OwnershipModel = r.ownershipModel.String
	}
	if r.primaryOwner.Valid {
		dto.PrimaryOwner = r.primaryOwner.String
	}
	if r.eaOwner.Valid {
		dto.EAOwner = r.eaOwner.String
	}
	return dto
}

func calculateMaturitySection(value int) *MaturitySectionDTO {
	switch {
	case value <= 24:
		return &MaturitySectionDTO{
			Name:  "Genesis",
			Order: 1,
			Range: MaturityRange{Min: 0, Max: 24},
		}
	case value <= 49:
		return &MaturitySectionDTO{
			Name:  "Custom Build",
			Order: 2,
			Range: MaturityRange{Min: 25, Max: 49},
		}
	case value <= 74:
		return &MaturitySectionDTO{
			Name:  "Product",
			Order: 3,
			Range: MaturityRange{Min: 50, Max: 74},
		}
	default:
		return &MaturitySectionDTO{
			Name:  "Commodity",
			Order: 4,
			Range: MaturityRange{Min: 75, Max: 99},
		}
	}
}

func (rm *CapabilityReadModel) GetAll(ctx context.Context) ([]CapabilityDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	return rm.queryCapabilityList(ctx,
		"SELECT id, name, description, parent_id, level, maturity_value, ownership_model, primary_owner, ea_owner, status, created_at FROM capabilities WHERE tenant_id = $1 ORDER BY level, name",
		tenantID.Value(),
	)
}

func (rm *CapabilityReadModel) GetChildren(ctx context.Context, parentID string) ([]CapabilityDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	return rm.queryCapabilityList(ctx,
		"SELECT id, name, description, parent_id, level, maturity_value, ownership_model, primary_owner, ea_owner, status, created_at FROM capabilities WHERE tenant_id = $1 AND parent_id = $2 ORDER BY name",
		tenantID.Value(), parentID,
	)
}

func (rm *CapabilityReadModel) queryCapabilityList(ctx context.Context, query string, args ...interface{}) ([]CapabilityDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var capabilities []CapabilityDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		caps, err := rm.scanCapabilityRows(ctx, tx, query, args...)
		if err != nil {
			return err
		}
		capabilities = caps

		if len(capabilities) == 0 {
			return nil
		}

		capabilityMap := buildCapabilityMap(capabilities)
		if err := rm.fetchExpertsBatch(ctx, tx, tenantID.Value(), capabilityMap); err != nil {
			return err
		}
		return rm.fetchTagsBatch(ctx, tx, tenantID.Value(), capabilityMap)
	})
	return capabilities, err
}

func (rm *CapabilityReadModel) scanCapabilityRows(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) ([]CapabilityDTO, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var capabilities []CapabilityDTO
	for rows.Next() {
		var result capabilityScanResult
		if err := rows.Scan(
			&result.dto.ID, &result.dto.Name, &result.dto.Description, &result.parentID,
			&result.dto.Level, &result.maturityValue,
			&result.ownershipModel, &result.primaryOwner, &result.eaOwner, &result.dto.Status, &result.dto.CreatedAt,
		); err != nil {
			return nil, err
		}
		capabilities = append(capabilities, *result.toDTO())
	}
	return capabilities, rows.Err()
}

func buildCapabilityMap(capabilities []CapabilityDTO) map[string]*CapabilityDTO {
	capabilityMap := make(map[string]*CapabilityDTO, len(capabilities))
	for i := range capabilities {
		capabilityMap[capabilities[i].ID] = &capabilities[i]
	}
	return capabilityMap
}

func buildInClause(ids []string) (placeholders string, args []interface{}) {
	args = make([]interface{}, len(ids))
	ph := make([]string, len(ids))
	for i, id := range ids {
		ph[i] = fmt.Sprintf("$%d", i+2)
		args[i] = id
	}
	return strings.Join(ph, ", "), args
}

type batchRowProcessor func(rows *sql.Rows, capabilityMap map[string]*CapabilityDTO) error

func (rm *CapabilityReadModel) fetchRelatedBatch(ctx context.Context, tx *sql.Tx, tenantID string, capabilityMap map[string]*CapabilityDTO, queryTemplate string, processor batchRowProcessor) error {
	if len(capabilityMap) == 0 {
		return nil
	}

	ids := make([]string, 0, len(capabilityMap))
	for id := range capabilityMap {
		ids = append(ids, id)
	}

	placeholders, idArgs := buildInClause(ids)
	query := fmt.Sprintf(queryTemplate, placeholders)

	rows, err := tx.QueryContext(ctx, query, append([]interface{}{tenantID}, idArgs...)...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := processor(rows, capabilityMap); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (rm *CapabilityReadModel) fetchExpertsBatch(ctx context.Context, tx *sql.Tx, tenantID string, capabilityMap map[string]*CapabilityDTO) error {
	return rm.fetchRelatedBatch(ctx, tx, tenantID, capabilityMap,
		"SELECT capability_id, expert_name, expert_role, contact_info, added_at FROM capability_experts WHERE tenant_id = $1 AND capability_id IN (%s)",
		func(rows *sql.Rows, m map[string]*CapabilityDTO) error {
			var capabilityID string
			var expert ExpertDTO
			if err := rows.Scan(&capabilityID, &expert.Name, &expert.Role, &expert.Contact, &expert.AddedAt); err != nil {
				return err
			}
			if cap, ok := m[capabilityID]; ok {
				cap.Experts = append(cap.Experts, expert)
			}
			return nil
		})
}

func (rm *CapabilityReadModel) fetchTagsBatch(ctx context.Context, tx *sql.Tx, tenantID string, capabilityMap map[string]*CapabilityDTO) error {
	return rm.fetchRelatedBatch(ctx, tx, tenantID, capabilityMap,
		"SELECT capability_id, tag FROM capability_tags WHERE tenant_id = $1 AND capability_id IN (%s) ORDER BY tag",
		func(rows *sql.Rows, m map[string]*CapabilityDTO) error {
			var capabilityID, tag string
			if err := rows.Scan(&capabilityID, &tag); err != nil {
				return err
			}
			if cap, ok := m[capabilityID]; ok {
				cap.Tags = append(cap.Tags, tag)
			}
			return nil
		})
}
