package aggregates

import (
	"testing"

	vo "easi/backend/internal/archassistant/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAIConfiguration_Defaults(t *testing.T) {
	config := NewAIConfiguration("tenant-1")

	assert.NotEmpty(t, config.ID())
	assert.Equal(t, "tenant-1", config.TenantID())
	assert.Equal(t, vo.StatusNotConfigured, config.Status())
	assert.Equal(t, vo.DefaultMaxTokens, config.MaxTokens().Value())
	assert.Equal(t, vo.DefaultTemperature, config.Temperature().Value())
	assert.True(t, config.APIKeyEncrypted().IsEmpty())
	assert.Nil(t, config.SystemPromptOverride())
}

func TestNewAIConfiguration_DefaultProvider(t *testing.T) {
	config := NewAIConfiguration("tenant-1")
	assert.Equal(t, "", config.Provider().Value())
}

func validUpdateParams(t *testing.T) UpdateConfigParams {
	t.Helper()
	endpoint, _ := vo.NewLLMEndpoint("https://api.openai.com")
	model, _ := vo.NewModelName("gpt-4o")
	maxTokens, _ := vo.NewMaxTokens(4096)
	temperature, _ := vo.NewTemperature(0.3)
	provider, _ := vo.NewLLMProvider("openai")
	key := vo.NewEncryptedAPIKey("encrypted-key")
	return UpdateConfigParams{
		Provider:    provider,
		Endpoint:    endpoint,
		APIKeyEncrypted: &key,
		Model:       model,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}
}

func TestUpdateConfig_WithAPIKey_BecomesConfigured(t *testing.T) {
	config := NewAIConfiguration("tenant-1")
	params := validUpdateParams(t)

	err := config.UpdateConfig(params)

	require.NoError(t, err)
	assert.Equal(t, vo.StatusConfigured, config.Status())
	assert.Equal(t, "gpt-4o", config.Model().Value())
	assert.Equal(t, "https://api.openai.com", config.Endpoint().Value())
	assert.Equal(t, "openai", config.Provider().Value())
	assert.False(t, config.APIKeyEncrypted().IsEmpty())
}

func TestUpdateConfig_WithoutAPIKey_WhenNoPreviousKey_Fails(t *testing.T) {
	config := NewAIConfiguration("tenant-1")
	params := validUpdateParams(t)
	params.APIKeyEncrypted = nil

	err := config.UpdateConfig(params)

	assert.ErrorIs(t, err, ErrAPIKeyRequiredForConfigured)
}

func TestUpdateConfig_WithoutAPIKey_PreservesPreviousKey(t *testing.T) {
	config := NewAIConfiguration("tenant-1")
	params := validUpdateParams(t)

	err := config.UpdateConfig(params)
	require.NoError(t, err)

	updateParams := validUpdateParams(t)
	updateParams.APIKeyEncrypted = nil
	newModel, _ := vo.NewModelName("gpt-4-turbo")
	updateParams.Model = newModel

	err = config.UpdateConfig(updateParams)
	require.NoError(t, err)
	assert.Equal(t, "gpt-4-turbo", config.Model().Value())
	assert.Equal(t, vo.StatusConfigured, config.Status())
}

func TestUpdateConfig_SetsSystemPromptOverride(t *testing.T) {
	config := NewAIConfiguration("tenant-1")
	params := validUpdateParams(t)
	prompt := "Custom instructions"
	params.SystemPromptOverride = &prompt

	err := config.UpdateConfig(params)
	require.NoError(t, err)
	assert.Equal(t, &prompt, config.SystemPromptOverride())
}

func TestUpdateConfig_SetsProvider(t *testing.T) {
	config := NewAIConfiguration("tenant-1")
	params := validUpdateParams(t)
	provider, _ := vo.NewLLMProvider("anthropic")
	params.Provider = provider

	err := config.UpdateConfig(params)
	require.NoError(t, err)
	assert.Equal(t, "anthropic", config.Provider().Value())
}

func TestReconstructAIConfiguration(t *testing.T) {
	params := ReconstructParams{
		ID:       "config-id",
		TenantID: "tenant-1",
		Provider: vo.ReconstructLLMProvider("anthropic"),
		Endpoint: vo.ReconstructLLMEndpoint("https://api.anthropic.com"),
		APIKeyEncrypted: vo.NewEncryptedAPIKey("enc-key"),
		Model:    vo.ReconstructModelName("claude-3-opus"),
		MaxTokens: vo.ReconstructMaxTokens(8192),
		Temperature: vo.ReconstructTemperature(0.5),
		Status:   vo.StatusConfigured,
	}

	config := ReconstructAIConfiguration(params)

	assert.Equal(t, "config-id", config.ID())
	assert.Equal(t, "tenant-1", config.TenantID())
	assert.Equal(t, "anthropic", config.Provider().Value())
	assert.Equal(t, "https://api.anthropic.com", config.Endpoint().Value())
	assert.Equal(t, "claude-3-opus", config.Model().Value())
	assert.Equal(t, 8192, config.MaxTokens().Value())
	assert.Equal(t, 0.5, config.Temperature().Value())
	assert.Equal(t, vo.StatusConfigured, config.Status())
}
