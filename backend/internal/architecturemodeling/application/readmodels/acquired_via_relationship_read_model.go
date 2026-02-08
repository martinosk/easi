package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type AcquiredViaRelationshipDTO struct {
	ID                 string      `json:"id"`
	AcquiredEntityID   string      `json:"acquiredEntityId"`
	AcquiredEntityName string      `json:"acquiredEntityName,omitempty"`
	ComponentID        string      `json:"componentId"`
	ComponentName      string      `json:"componentName,omitempty"`
	Notes              string      `json:"notes,omitempty"`
	CreatedAt          time.Time   `json:"createdAt"`
	Links              types.Links `json:"_links,omitempty"`
}

type AcquiredViaRelationshipReadModel struct {
	db *database.TenantAwareDB
}

func NewAcquiredViaRelationshipReadModel(db *database.TenantAwareDB) *AcquiredViaRelationshipReadModel {
	return &AcquiredViaRelationshipReadModel{db: db}
}

func (rm *AcquiredViaRelationshipReadModel) Insert(ctx context.Context, dto AcquiredViaRelationshipDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO acquired_via_relationships (id, tenant_id, acquired_entity_id, component_id, notes, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		dto.ID, tenantID.Value(), dto.AcquiredEntityID, dto.ComponentID, dto.Notes, dto.CreatedAt,
	)
	return err
}

func (rm *AcquiredViaRelationshipReadModel) Upsert(ctx context.Context, dto AcquiredViaRelationshipDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO acquired_via_relationships (id, tenant_id, acquired_entity_id, component_id, notes, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (tenant_id, component_id) WHERE is_deleted = FALSE
		 DO UPDATE SET acquired_entity_id = $3, notes = $5, created_at = $6, is_deleted = FALSE, deleted_at = NULL`,
		dto.ID, tenantID.Value(), dto.AcquiredEntityID, dto.ComponentID, dto.Notes, dto.CreatedAt,
	)
	return err
}

func (rm *AcquiredViaRelationshipReadModel) UpdateByComponentID(ctx context.Context, dto AcquiredViaRelationshipDTO) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		`UPDATE acquired_via_relationships
		 SET acquired_entity_id = $1, notes = $2, created_at = $3, is_deleted = FALSE, deleted_at = NULL
		 WHERE tenant_id = $4 AND component_id = $5`,
		dto.AcquiredEntityID, dto.Notes, dto.CreatedAt, tenantID.Value(), dto.ComponentID,
	)
	return err
}

func (rm *AcquiredViaRelationshipReadModel) UpdateNotesByComponentID(ctx context.Context, componentID string, notes string) error {
	return rm.execWithTenant(ctx,
		`UPDATE acquired_via_relationships SET notes = $1 WHERE tenant_id = $2 AND component_id = $3`,
		func(tenantID string) []interface{} { return []interface{}{notes, tenantID, componentID} },
	)
}

func (rm *AcquiredViaRelationshipReadModel) DeleteByComponentID(ctx context.Context, componentID string) error {
	return rm.execWithTenant(ctx,
		`DELETE FROM acquired_via_relationships WHERE tenant_id = $1 AND component_id = $2`,
		func(tenantID string) []interface{} { return []interface{}{tenantID, componentID} },
	)
}

func (rm *AcquiredViaRelationshipReadModel) execWithTenant(ctx context.Context, query string, argsFn func(tenantID string) []interface{}) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, argsFn(tenantID.Value())...)
	return err
}

func (rm *AcquiredViaRelationshipReadModel) MarkAsDeleted(ctx context.Context, id string, deletedAt time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE acquired_via_relationships SET is_deleted = TRUE, deleted_at = $1 WHERE tenant_id = $2 AND id = $3",
		deletedAt, tenantID.Value(), id,
	)
	return err
}

func (rm *AcquiredViaRelationshipReadModel) GetByID(ctx context.Context, id string) (*AcquiredViaRelationshipDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto AcquiredViaRelationshipDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT r.id, r.acquired_entity_id, COALESCE(ae.name, '') as acquired_entity_name,
			        r.component_id, COALESCE(c.name, '') as component_name, r.notes, r.created_at
			 FROM acquired_via_relationships r
			 LEFT JOIN acquired_entities ae ON ae.tenant_id = r.tenant_id AND ae.id = r.acquired_entity_id AND ae.is_deleted = FALSE
			 LEFT JOIN application_components c ON c.tenant_id = r.tenant_id AND c.id = r.component_id AND c.is_deleted = FALSE
			 WHERE r.tenant_id = $1 AND r.id = $2 AND r.is_deleted = FALSE`,
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.AcquiredEntityID, &dto.AcquiredEntityName, &dto.ComponentID, &dto.ComponentName, &dto.Notes, &dto.CreatedAt)

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

func (rm *AcquiredViaRelationshipReadModel) GetByEntityID(ctx context.Context, entityID string) ([]AcquiredViaRelationshipDTO, error) {
	return rm.queryList(ctx,
		`WHERE r.tenant_id = $1 AND r.acquired_entity_id = $2 AND r.is_deleted = FALSE`,
		func(tenantID string) []interface{} { return []interface{}{tenantID, entityID} },
	)
}

func (rm *AcquiredViaRelationshipReadModel) GetByComponentID(ctx context.Context, componentID string) ([]AcquiredViaRelationshipDTO, error) {
	return rm.queryList(ctx,
		`WHERE r.tenant_id = $1 AND r.component_id = $2 AND r.is_deleted = FALSE`,
		func(tenantID string) []interface{} { return []interface{}{tenantID, componentID} },
	)
}

func (rm *AcquiredViaRelationshipReadModel) GetAll(ctx context.Context) ([]AcquiredViaRelationshipDTO, error) {
	return rm.queryList(ctx,
		`WHERE r.tenant_id = $1 AND r.is_deleted = FALSE`,
		func(tenantID string) []interface{} { return []interface{}{tenantID} },
	)
}

func (rm *AcquiredViaRelationshipReadModel) queryList(ctx context.Context, whereClause string, argsFn func(tenantID string) []interface{}) ([]AcquiredViaRelationshipDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	query := `SELECT r.id, r.acquired_entity_id, COALESCE(ae.name, '') as acquired_entity_name,
		        r.component_id, COALESCE(c.name, '') as component_name, r.notes, r.created_at
		 FROM acquired_via_relationships r
		 LEFT JOIN acquired_entities ae ON ae.tenant_id = r.tenant_id AND ae.id = r.acquired_entity_id AND ae.is_deleted = FALSE
		 LEFT JOIN application_components c ON c.tenant_id = r.tenant_id AND c.id = r.component_id AND c.is_deleted = FALSE
		 ` + whereClause

	relationships := make([]AcquiredViaRelationshipDTO, 0)
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, argsFn(tenantID.Value())...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto AcquiredViaRelationshipDTO
			if err := rows.Scan(&dto.ID, &dto.AcquiredEntityID, &dto.AcquiredEntityName, &dto.ComponentID, &dto.ComponentName, &dto.Notes, &dto.CreatedAt); err != nil {
				return err
			}
			relationships = append(relationships, dto)
		}
		return rows.Err()
	})

	return relationships, err
}
