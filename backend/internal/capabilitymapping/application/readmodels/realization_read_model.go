package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type RealizationDTO struct {
	ID                   string            `json:"id"`
	CapabilityID         string            `json:"capabilityId"`
	ComponentID          string            `json:"componentId"`
	ComponentName        string            `json:"componentName,omitempty"`
	RealizationLevel     string            `json:"realizationLevel"`
	Notes                string            `json:"notes,omitempty"`
	Origin               string            `json:"origin"`
	SourceRealizationID  string            `json:"sourceRealizationId,omitempty"`
	SourceCapabilityID   string            `json:"sourceCapabilityId,omitempty"`
	SourceCapabilityName string            `json:"sourceCapabilityName,omitempty"`
	LinkedAt             time.Time         `json:"linkedAt"`
	Links                map[string]string `json:"_links,omitempty"`
}

type RealizationReadModel struct {
	db *database.TenantAwareDB
}

func NewRealizationReadModel(db *database.TenantAwareDB) *RealizationReadModel {
	return &RealizationReadModel{db: db}
}

func (rm *RealizationReadModel) Insert(ctx context.Context, dto RealizationDTO) error {
	return rm.insertRealization(ctx, dto, "INSERT INTO capability_realizations (id, tenant_id, capability_id, component_id, realization_level, notes, origin, source_realization_id, source_capability_id, linked_at, component_name, source_capability_name) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)", true)
}

func (rm *RealizationReadModel) Update(ctx context.Context, id, realizationLevel, notes string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE capability_realizations SET realization_level = $1, notes = $2, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $3 AND id = $4",
		realizationLevel, notes, tenantID.Value(), id,
	)
	return err
}

func (rm *RealizationReadModel) Delete(ctx context.Context, id string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx, "DELETE FROM capability_realizations WHERE tenant_id = $1 AND id = $2", tenantID.Value(), id)
	return err
}

const realizationSelectColumns = `id, capability_id, component_id, realization_level, notes, origin,
	source_realization_id, source_capability_id, linked_at,
	component_name, source_capability_name`

func (rm *RealizationReadModel) GetByCapabilityID(ctx context.Context, capabilityID string) ([]RealizationDTO, error) {
	query := `SELECT ` + realizationSelectColumns + ` FROM capability_realizations WHERE tenant_id = $1 AND capability_id = $2 ORDER BY linked_at DESC`
	return rm.queryRealizations(ctx, query, capabilityID)
}

func (rm *RealizationReadModel) GetByComponentID(ctx context.Context, componentID string) ([]RealizationDTO, error) {
	query := `SELECT ` + realizationSelectColumns + ` FROM capability_realizations WHERE tenant_id = $1 AND component_id = $2 ORDER BY linked_at DESC`
	return rm.queryRealizations(ctx, query, componentID)
}

func (rm *RealizationReadModel) queryRealizations(ctx context.Context, query, param string) ([]RealizationDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var realizations []RealizationDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, tenantID.Value(), param)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			dto, err := rm.scanRealizationRow(rows)
			if err != nil {
				return err
			}
			realizations = append(realizations, dto)
		}

		return rows.Err()
	})

	return realizations, err
}

func (rm *RealizationReadModel) scanRealizationRow(rows *sql.Rows) (RealizationDTO, error) {
	var dto RealizationDTO
	var sourceRealizationID, sourceCapabilityID, componentName, sourceCapabilityName sql.NullString
	err := rows.Scan(
		&dto.ID, &dto.CapabilityID, &dto.ComponentID, &dto.RealizationLevel, &dto.Notes, &dto.Origin,
		&sourceRealizationID, &sourceCapabilityID, &dto.LinkedAt,
		&componentName, &sourceCapabilityName,
	)
	if err != nil {
		return dto, err
	}
	dto.SourceRealizationID = sourceRealizationID.String
	dto.SourceCapabilityID = sourceCapabilityID.String
	dto.ComponentName = componentName.String
	dto.SourceCapabilityName = sourceCapabilityName.String
	return dto, nil
}

func (rm *RealizationReadModel) GetByID(ctx context.Context, id string) (*RealizationDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto RealizationDTO
	var sourceRealizationID, sourceCapabilityID, componentName, sourceCapabilityName sql.NullString
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := `SELECT ` + realizationSelectColumns + ` FROM capability_realizations WHERE tenant_id = $1 AND id = $2`
		err := tx.QueryRowContext(ctx, query, tenantID.Value(), id).Scan(
			&dto.ID, &dto.CapabilityID, &dto.ComponentID, &dto.RealizationLevel, &dto.Notes, &dto.Origin,
			&sourceRealizationID, &sourceCapabilityID, &dto.LinkedAt,
			&componentName, &sourceCapabilityName,
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

	dto.SourceRealizationID = sourceRealizationID.String
	dto.SourceCapabilityID = sourceCapabilityID.String
	dto.ComponentName = componentName.String
	dto.SourceCapabilityName = sourceCapabilityName.String

	return &dto, nil
}

func (rm *RealizationReadModel) DeleteBySourceRealizationID(ctx context.Context, sourceRealizationID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx, "DELETE FROM capability_realizations WHERE tenant_id = $1 AND source_realization_id = $2", tenantID.Value(), sourceRealizationID)
	return err
}

func (rm *RealizationReadModel) InsertInherited(ctx context.Context, dto RealizationDTO) error {
	return rm.insertRealization(ctx, dto, `INSERT INTO capability_realizations (id, tenant_id, capability_id, component_id, realization_level, notes, origin, source_realization_id, source_capability_id, linked_at, component_name, source_capability_name)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (tenant_id, capability_id, component_id) DO NOTHING`, false)
}

func (rm *RealizationReadModel) insertRealization(ctx context.Context, dto RealizationDTO, query string, includeID bool) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	args := rm.buildInsertArgs(tenantID.Value(), dto, includeID)
	_, err = rm.db.ExecContext(ctx, query, args...)
	return err
}

func (rm *RealizationReadModel) buildInsertArgs(tenantID string, dto RealizationDTO, includeID bool) []interface{} {
	commonArgs := []interface{}{
		tenantID, dto.CapabilityID, dto.ComponentID, dto.RealizationLevel,
		dto.Notes, dto.Origin, rm.toNullableString(dto.SourceRealizationID),
		rm.toNullableString(dto.SourceCapabilityID), dto.LinkedAt,
		rm.toNullableString(dto.ComponentName), rm.toNullableString(dto.SourceCapabilityName),
	}

	if includeID {
		return append([]interface{}{dto.ID}, commonArgs...)
	}
	return commonArgs
}

func (rm *RealizationReadModel) toNullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (rm *RealizationReadModel) UpdateComponentName(ctx context.Context, componentID, componentName string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE capability_realizations SET component_name = $1 WHERE tenant_id = $2 AND component_id = $3",
		componentName, tenantID.Value(), componentID,
	)
	return err
}

func (rm *RealizationReadModel) UpdateSourceCapabilityName(ctx context.Context, capabilityID, capabilityName string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE capability_realizations cr
		 SET source_capability_name = $1
		 FROM capability_realizations source_r
		 WHERE cr.tenant_id = $2
		   AND cr.source_realization_id = source_r.id
		   AND source_r.tenant_id = cr.tenant_id
		   AND source_r.capability_id = $3`,
		capabilityName, tenantID.Value(), capabilityID,
	)
	return err
}

func (rm *RealizationReadModel) GetSourceCapabilityID(ctx context.Context, sourceRealizationID string) (string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}

	var capabilityID string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			"SELECT capability_id FROM capability_realizations WHERE tenant_id = $1 AND id = $2",
			tenantID.Value(), sourceRealizationID,
		).Scan(&capabilityID)
	})

	if err == sql.ErrNoRows {
		return "", nil
	}
	return capabilityID, err
}

type CapabilityRealizationsGroup struct {
	CapabilityID   string           `json:"capabilityId"`
	CapabilityName string           `json:"capabilityName"`
	Level          string           `json:"level"`
	Realizations   []RealizationDTO `json:"realizations"`
}

type domainRealizationRow struct {
	capID, capName, capLevel                                                     string
	realizationID, realCapID, componentID, realizationLevel, notes, origin       sql.NullString
	sourceRealizationID, sourceCapabilityID, componentName, sourceCapabilityName sql.NullString
	linkedAt                                                                     sql.NullTime
}

func (r *domainRealizationRow) toDTO() RealizationDTO {
	return RealizationDTO{
		ID:                   r.realizationID.String,
		CapabilityID:         r.realCapID.String,
		ComponentID:          r.componentID.String,
		RealizationLevel:     r.realizationLevel.String,
		Notes:                r.notes.String,
		Origin:               r.origin.String,
		SourceRealizationID:  r.sourceRealizationID.String,
		SourceCapabilityID:   r.sourceCapabilityID.String,
		LinkedAt:             r.linkedAt.Time,
		ComponentName:        r.componentName.String,
		SourceCapabilityName: r.sourceCapabilityName.String,
	}
}

type realizationGroupBuilder struct {
	groupMap map[string]*CapabilityRealizationsGroup
	order    []string
}

func newRealizationGroupBuilder() *realizationGroupBuilder {
	return &realizationGroupBuilder{
		groupMap: make(map[string]*CapabilityRealizationsGroup),
		order:    make([]string, 0),
	}
}

func (b *realizationGroupBuilder) addRow(row domainRealizationRow) {
	group, exists := b.groupMap[row.capID]
	if !exists {
		group = &CapabilityRealizationsGroup{
			CapabilityID:   row.capID,
			CapabilityName: row.capName,
			Level:          row.capLevel,
			Realizations:   []RealizationDTO{},
		}
		b.groupMap[row.capID] = group
		b.order = append(b.order, row.capID)
	}

	if row.realizationID.Valid {
		group.Realizations = append(group.Realizations, row.toDTO())
	}
}

func (b *realizationGroupBuilder) build() []CapabilityRealizationsGroup {
	groups := make([]CapabilityRealizationsGroup, 0, len(b.order))
	for _, capID := range b.order {
		groups = append(groups, *b.groupMap[capID])
	}
	return groups
}

const domainRealizationsQuery = `
	WITH RECURSIVE domain_capabilities AS (
		SELECT c.id, c.name, c.level, dca.capability_name as root_name
		FROM capabilities c
		INNER JOIN domain_capability_assignments dca
			ON c.id = dca.capability_id AND c.tenant_id = dca.tenant_id
		WHERE dca.tenant_id = $1
			AND dca.business_domain_id = $2

		UNION ALL

		SELECT c.id, c.name, c.level, dc.root_name
		FROM capabilities c
		INNER JOIN domain_capabilities dc ON c.parent_id = dc.id
		WHERE c.tenant_id = $1
	)
	SELECT dc.id, dc.name, dc.level,
		cr.id as realization_id, cr.capability_id, cr.component_id,
		cr.realization_level, cr.notes, cr.origin,
		cr.source_realization_id, cr.source_capability_id, cr.linked_at,
		cr.component_name, cr.source_capability_name
	FROM domain_capabilities dc
	LEFT JOIN capability_realizations cr
		ON dc.id = cr.capability_id AND cr.tenant_id = $1
	WHERE CAST(SUBSTRING(dc.level FROM 2) AS INTEGER) <= $3
	ORDER BY dc.root_name, dc.level, dc.name, cr.linked_at DESC`

func (rm *RealizationReadModel) GetByBusinessDomainAndDepth(ctx context.Context, domainID string, maxDepth int) ([]CapabilityRealizationsGroup, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	builder := newRealizationGroupBuilder()

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, domainRealizationsQuery, tenantID.Value(), domainID, maxDepth)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var row domainRealizationRow
			err := rows.Scan(
				&row.capID, &row.capName, &row.capLevel,
				&row.realizationID, &row.realCapID, &row.componentID,
				&row.realizationLevel, &row.notes, &row.origin,
				&row.sourceRealizationID, &row.sourceCapabilityID, &row.linkedAt,
				&row.componentName, &row.sourceCapabilityName,
			)
			if err != nil {
				return err
			}
			builder.addRow(row)
		}
		return rows.Err()
	})

	if err != nil {
		return nil, err
	}
	return builder.build(), nil
}
