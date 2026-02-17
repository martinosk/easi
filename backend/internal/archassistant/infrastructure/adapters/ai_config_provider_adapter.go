package adapters

import (
	"context"
	"errors"

	"easi/backend/internal/archassistant/domain"
	"easi/backend/internal/archassistant/publishedlanguage"
	"easi/backend/internal/shared/crypto"
	sharedctx "easi/backend/internal/shared/context"
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
		return nil, err
	}
	if config == nil {
		return nil, errors.New("AI configuration not found")
	}
	if !config.Status().IsConfigured() {
		return nil, errors.New("AI configuration is not configured")
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
