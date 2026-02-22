package adapters

import (
	"context"
	"fmt"

	"easi/backend/internal/archassistant/domain"
	"easi/backend/internal/archassistant/publishedlanguage"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/crypto"
)

type AIConfigProviderAdapter struct {
	repo domain.AIConfigurationRepository
}

func NewAIConfigProviderAdapter(repo domain.AIConfigurationRepository) publishedlanguage.AIConfigProvider {
	return &AIConfigProviderAdapter{repo: repo}
}

func (a *AIConfigProviderAdapter) GetDecryptedConfig(ctx context.Context) (*publishedlanguage.AIConfigInfo, error) {
	config, err := a.repo.GetByTenantID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AI configuration: %w", err)
	}
	if config == nil {
		return nil, publishedlanguage.ErrNotConfigured
	}
	if !config.Status().IsConfigured() {
		return nil, publishedlanguage.ErrNotConfigured
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	apiKey, err := crypto.Decrypt(config.APIKeyEncrypted().Value(), tenantID.Value())
	if err != nil {
		return nil, err
	}

	return &publishedlanguage.AIConfigInfo{
		Provider:    config.Provider().Value(),
		Endpoint:    config.Endpoint().Value(),
		APIKey:      apiKey,
		Model:       config.Model().Value(),
		MaxTokens:   config.MaxTokens().Value(),
		Temperature: config.Temperature().Value(),
	}, nil
}

func (a *AIConfigProviderAdapter) IsConfigured(ctx context.Context) (bool, error) {
	config, err := a.repo.GetByTenantID(ctx)
	if err != nil {
		return false, err
	}
	if config == nil {
		return false, nil
	}
	return config.Status().IsConfigured(), nil
}
