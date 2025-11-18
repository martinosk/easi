package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type CapabilityDTO struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Description    string            `json:"description,omitempty"`
	ParentID       string            `json:"parentId,omitempty"`
	Level          string            `json:"level"`
	StrategyPillar string            `json:"strategyPillar,omitempty"`
	PillarWeight   int               `json:"pillarWeight,omitempty"`
	MaturityLevel  string            `json:"maturityLevel,omitempty"`
	OwnershipModel string            `json:"ownershipModel,omitempty"`
	PrimaryOwner   string            `json:"primaryOwner,omitempty"`
	EAOwner        string            `json:"eaOwner,omitempty"`
	Status         string            `json:"status,omitempty"`
	Experts        []ExpertDTO       `json:"experts,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	Links          map[string]string `json:"_links,omitempty"`
}

type ExpertDTO struct {
	Name    string    `json:"name"`
	Role    string    `json:"role"`
	Contact string    `json:"contact"`
	AddedAt time.Time `json:"addedAt"`
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

func (rm *CapabilityReadModel) UpdateMetadata(ctx context.Context, id, strategyPillar string, pillarWeight int, maturityLevel, ownershipModel, primaryOwner, eaOwner, status string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE capabilities SET strategy_pillar = $1, pillar_weight = $2, maturity_level = $3, ownership_model = $4, primary_owner = $5, ea_owner = $6, status = $7, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $8 AND id = $9",
		strategyPillar, pillarWeight, maturityLevel, ownershipModel, primaryOwner, eaOwner, status, tenantID.Value(), id,
	)
	return err
}

func (rm *CapabilityReadModel) AddExpert(ctx context.Context, capabilityID, expertName, expertRole, contactInfo string, addedAt time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO capability_experts (capability_id, tenant_id, expert_name, expert_role, contact_info, added_at) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (tenant_id, capability_id, expert_name) DO NOTHING",
		capabilityID, tenantID.Value(), expertName, expertRole, contactInfo, addedAt,
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

func (rm *CapabilityReadModel) GetByID(ctx context.Context, id string) (*CapabilityDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto CapabilityDTO
	var parentID, strategyPillar, ownershipModel, primaryOwner, eaOwner sql.NullString
	var pillarWeight sql.NullInt64
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			"SELECT id, name, description, parent_id, level, strategy_pillar, pillar_weight, maturity_level, ownership_model, primary_owner, ea_owner, status, created_at FROM capabilities WHERE tenant_id = $1 AND id = $2",
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.Name, &dto.Description, &parentID, &dto.Level, &strategyPillar, &pillarWeight, &dto.MaturityLevel, &ownershipModel, &primaryOwner, &eaOwner, &dto.Status, &dto.CreatedAt)

		if err == sql.ErrNoRows {
			notFound = true
			return nil
		}
		if err != nil {
			return err
		}

		rows, err := tx.QueryContext(ctx,
			"SELECT expert_name, expert_role, contact_info, added_at FROM capability_experts WHERE tenant_id = $1 AND capability_id = $2",
			tenantID.Value(), id,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var expert ExpertDTO
			if err := rows.Scan(&expert.Name, &expert.Role, &expert.Contact, &expert.AddedAt); err != nil {
				return err
			}
			dto.Experts = append(dto.Experts, expert)
		}

		tagRows, err := tx.QueryContext(ctx,
			"SELECT tag FROM capability_tags WHERE tenant_id = $1 AND capability_id = $2 ORDER BY tag",
			tenantID.Value(), id,
		)
		if err != nil {
			return err
		}
		defer tagRows.Close()

		for tagRows.Next() {
			var tag string
			if err := tagRows.Scan(&tag); err != nil {
				return err
			}
			dto.Tags = append(dto.Tags, tag)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	if notFound {
		return nil, nil
	}

	if parentID.Valid {
		dto.ParentID = parentID.String
	}
	if strategyPillar.Valid {
		dto.StrategyPillar = strategyPillar.String
	}
	if pillarWeight.Valid {
		dto.PillarWeight = int(pillarWeight.Int64)
	}
	if ownershipModel.Valid {
		dto.OwnershipModel = ownershipModel.String
	}
	if primaryOwner.Valid {
		dto.PrimaryOwner = primaryOwner.String
	}
	if eaOwner.Valid {
		dto.EAOwner = eaOwner.String
	}

	return &dto, nil
}

func (rm *CapabilityReadModel) GetAll(ctx context.Context) ([]CapabilityDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var capabilities []CapabilityDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, name, description, parent_id, level, created_at FROM capabilities WHERE tenant_id = $1 ORDER BY level, name",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto CapabilityDTO
			var parentID sql.NullString
			if err := rows.Scan(&dto.ID, &dto.Name, &dto.Description, &parentID, &dto.Level, &dto.CreatedAt); err != nil {
				return err
			}
			if parentID.Valid {
				dto.ParentID = parentID.String
			}
			capabilities = append(capabilities, dto)
		}

		return rows.Err()
	})

	return capabilities, err
}

func (rm *CapabilityReadModel) GetChildren(ctx context.Context, parentID string) ([]CapabilityDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var capabilities []CapabilityDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, name, description, parent_id, level, created_at FROM capabilities WHERE tenant_id = $1 AND parent_id = $2 ORDER BY name",
			tenantID.Value(), parentID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto CapabilityDTO
			var parentIDVal sql.NullString
			if err := rows.Scan(&dto.ID, &dto.Name, &dto.Description, &parentIDVal, &dto.Level, &dto.CreatedAt); err != nil {
				return err
			}
			if parentIDVal.Valid {
				dto.ParentID = parentIDVal.String
			}
			capabilities = append(capabilities, dto)
		}

		return rows.Err()
	})

	return capabilities, err
}
