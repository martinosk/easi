package aggregates

import (
	"errors"
	"time"

	vo "easi/backend/internal/archassistant/domain/valueobjects"

	"github.com/google/uuid"
)

var ErrAPIKeyRequiredForConfigured = errors.New("API key is required to mark configuration as configured")

type AIConfiguration struct {
	id                   string
	tenantID             string
	provider             vo.LLMProvider
	endpoint             vo.LLMEndpoint
	apiKeyEncrypted      vo.EncryptedAPIKey
	model                vo.ModelName
	maxTokens            vo.MaxTokens
	temperature          vo.Temperature
	systemPromptOverride *string
	status               vo.ConfigurationStatus
	updatedAt            time.Time
}

func NewAIConfiguration(tenantID string) *AIConfiguration {
	return &AIConfiguration{
		id:        uuid.New().String(),
		tenantID:  tenantID,
		maxTokens: vo.DefaultMaxTokensValue(),
		temperature: vo.DefaultTemperatureValue(),
		status:    vo.StatusNotConfigured,
		updatedAt: time.Now(),
	}
}

type ReconstructParams struct {
	ID                   string
	TenantID             string
	Provider             vo.LLMProvider
	Endpoint             vo.LLMEndpoint
	APIKeyEncrypted      vo.EncryptedAPIKey
	Model                vo.ModelName
	MaxTokens            vo.MaxTokens
	Temperature          vo.Temperature
	SystemPromptOverride *string
	Status               vo.ConfigurationStatus
	UpdatedAt            time.Time
}

func ReconstructAIConfiguration(p ReconstructParams) *AIConfiguration {
	return &AIConfiguration{
		id:                   p.ID,
		tenantID:             p.TenantID,
		provider:             p.Provider,
		endpoint:             p.Endpoint,
		apiKeyEncrypted:      p.APIKeyEncrypted,
		model:                p.Model,
		maxTokens:            p.MaxTokens,
		temperature:          p.Temperature,
		systemPromptOverride: p.SystemPromptOverride,
		status:               p.Status,
		updatedAt:            p.UpdatedAt,
	}
}

type UpdateConfigParams struct {
	Provider             vo.LLMProvider
	Endpoint             vo.LLMEndpoint
	APIKeyEncrypted      *vo.EncryptedAPIKey
	Model                vo.ModelName
	MaxTokens            vo.MaxTokens
	Temperature          vo.Temperature
	SystemPromptOverride *string
}

func (c *AIConfiguration) UpdateConfig(params UpdateConfigParams) error {
	c.provider = params.Provider
	c.endpoint = params.Endpoint
	c.model = params.Model
	c.maxTokens = params.MaxTokens
	c.temperature = params.Temperature
	c.systemPromptOverride = params.SystemPromptOverride

	if params.APIKeyEncrypted != nil {
		c.apiKeyEncrypted = *params.APIKeyEncrypted
	}

	if c.apiKeyEncrypted.IsEmpty() {
		return ErrAPIKeyRequiredForConfigured
	}

	c.status = vo.StatusConfigured
	c.updatedAt = time.Now()
	return nil
}

func (c *AIConfiguration) ID() string                         { return c.id }
func (c *AIConfiguration) TenantID() string                   { return c.tenantID }
func (c *AIConfiguration) Provider() vo.LLMProvider           { return c.provider }
func (c *AIConfiguration) Endpoint() vo.LLMEndpoint           { return c.endpoint }
func (c *AIConfiguration) APIKeyEncrypted() vo.EncryptedAPIKey { return c.apiKeyEncrypted }
func (c *AIConfiguration) Model() vo.ModelName                { return c.model }
func (c *AIConfiguration) MaxTokens() vo.MaxTokens            { return c.maxTokens }
func (c *AIConfiguration) Temperature() vo.Temperature        { return c.temperature }
func (c *AIConfiguration) SystemPromptOverride() *string       { return c.systemPromptOverride }
func (c *AIConfiguration) Status() vo.ConfigurationStatus     { return c.status }
func (c *AIConfiguration) UpdatedAt() time.Time               { return c.updatedAt }
