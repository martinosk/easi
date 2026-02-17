package repositories

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/archassistant/domain"
	"easi/backend/internal/archassistant/domain/aggregates"
	vo "easi/backend/internal/archassistant/domain/valueobjects"
	"easi/backend/internal/infrastructure/database"
)

type AIConfigurationRepository struct {
	db *database.TenantAwareDB
}

func NewAIConfigurationRepository(db *database.TenantAwareDB) domain.AIConfigurationRepository {
	return &AIConfigurationRepository{db: db}
}

func (r *AIConfigurationRepository) GetByTenantID(ctx context.Context) (*aggregates.AIConfiguration, error) {
	var config *aggregates.AIConfiguration
	err := r.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx, `
			SELECT id, tenant_id, provider, endpoint, api_key_encrypted, model,
			       max_tokens, temperature, system_prompt_override, status, updated_at
			FROM archassistant.ai_configurations
			LIMIT 1
		`)
		c, err := scanConfig(row)
		if err == sql.ErrNoRows {
			return nil
		}
		if err != nil {
			return err
		}
		config = c
		return nil
	})
	return config, err
}

func (r *AIConfigurationRepository) Save(ctx context.Context, config *aggregates.AIConfiguration) error {
	return r.db.WithTenantContext(ctx, func(conn *sql.Conn) error {
		_, err := conn.ExecContext(ctx, `
			INSERT INTO archassistant.ai_configurations
				(id, tenant_id, provider, endpoint, api_key_encrypted, model, max_tokens, temperature, system_prompt_override, status, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (tenant_id) DO UPDATE SET
				provider = EXCLUDED.provider,
				endpoint = EXCLUDED.endpoint,
				api_key_encrypted = EXCLUDED.api_key_encrypted,
				model = EXCLUDED.model,
				max_tokens = EXCLUDED.max_tokens,
				temperature = EXCLUDED.temperature,
				system_prompt_override = EXCLUDED.system_prompt_override,
				status = EXCLUDED.status,
				updated_at = EXCLUDED.updated_at
		`,
			config.ID(),
			config.TenantID(),
			config.Provider().Value(),
			config.Endpoint().Value(),
			config.APIKeyEncrypted().Value(),
			config.Model().Value(),
			config.MaxTokens().Value(),
			config.Temperature().Value(),
			nilIfEmpty(config.SystemPromptOverride()),
			config.Status().Value(),
			config.UpdatedAt(),
		)
		return err
	})
}

func nilIfEmpty(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanConfig(s scanner) (*aggregates.AIConfiguration, error) {
	var (
		id, tenantID, providerStr, endpointStr, apiKeyEnc, modelStr, statusStr string
		maxTokens                                                               int
		temperature                                                             float64
		systemPromptOverride                                                    *string
		updatedAt                                                               time.Time
	)

	err := s.Scan(&id, &tenantID, &providerStr, &endpointStr, &apiKeyEnc, &modelStr,
		&maxTokens, &temperature, &systemPromptOverride, &statusStr, &updatedAt)
	if err != nil {
		return nil, err
	}

	status, _ := vo.ConfigurationStatusFromString(statusStr)

	return aggregates.ReconstructAIConfiguration(aggregates.ReconstructParams{
		ID:                   id,
		TenantID:             tenantID,
		Provider:             vo.ReconstructLLMProvider(providerStr),
		Endpoint:             vo.ReconstructLLMEndpoint(endpointStr),
		APIKeyEncrypted:      vo.NewEncryptedAPIKey(apiKeyEnc),
		Model:                vo.ReconstructModelName(modelStr),
		MaxTokens:            vo.ReconstructMaxTokens(maxTokens),
		Temperature:          vo.ReconstructTemperature(temperature),
		SystemPromptOverride: systemPromptOverride,
		Status:               status,
		UpdatedAt:            updatedAt,
	}), nil
}
