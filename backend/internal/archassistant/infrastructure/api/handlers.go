package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"easi/backend/internal/archassistant/domain"
	"easi/backend/internal/archassistant/domain/aggregates"
	vo "easi/backend/internal/archassistant/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/crypto"
)

type AIConfigHandlers struct {
	repo domain.AIConfigurationRepository
}

func NewAIConfigHandlers(repo domain.AIConfigurationRepository) *AIConfigHandlers {
	return &AIConfigHandlers{repo: repo}
}

type AIConfigResponse struct {
	ID                   string                    `json:"id"`
	Provider             string                    `json:"provider"`
	Endpoint             string                    `json:"endpoint"`
	APIKeyStatus         string                    `json:"apiKeyStatus"`
	Model                string                    `json:"model"`
	MaxTokens            int                       `json:"maxTokens"`
	Temperature          float64                   `json:"temperature"`
	SystemPromptOverride *string                   `json:"systemPromptOverride,omitempty"`
	Status               string                    `json:"status"`
	UpdatedAt            time.Time                 `json:"updatedAt"`
	Links                map[string]sharedAPI.Link `json:"_links"`
}

type UpdateAIConfigRequest struct {
	Provider             string  `json:"provider"`
	Endpoint             string  `json:"endpoint"`
	APIKey               string  `json:"apiKey,omitempty"`
	Model                string  `json:"model"`
	MaxTokens            int     `json:"maxTokens"`
	Temperature          float64 `json:"temperature"`
	SystemPromptOverride *string `json:"systemPromptOverride,omitempty"`
}

type TestConnectionResponse struct {
	Success   bool   `json:"success"`
	Model     string `json:"model,omitempty"`
	LatencyMs int64  `json:"latencyMs,omitempty"`
	Error     string `json:"error,omitempty"`
}

// GetConfig godoc
// @Summary Get assistant configuration
// @Description Retrieves the AI assistant configuration for the current tenant. If no configuration exists yet, returns a default not_configured response.
// @Tags assistant-config
// @Produce json
// @Success 200 {object} AIConfigResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /assistant-config [get]
func (h *AIConfigHandlers) GetConfig(w http.ResponseWriter, r *http.Request) {
	config, err := h.repo.GetByTenantID(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve configuration")
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())

	if config == nil {
		resp := AIConfigResponse{
			Status:      vo.StatusNotConfigured.Value(),
			MaxTokens:   vo.DefaultMaxTokens,
			Temperature: vo.DefaultTemperature,
			Links:       configLinks(actor, false),
		}
		sharedAPI.RespondJSON(w, http.StatusOK, resp)
		return
	}

	resp := toResponse(config, actor)
	sharedAPI.RespondJSON(w, http.StatusOK, resp)
}

// UpdateConfig godoc
// @Summary Create or update assistant configuration
// @Description Creates or updates the AI assistant configuration for the current tenant. API key is optional on update and existing encrypted key is preserved when omitted.
// @Tags assistant-config
// @Accept json
// @Produce json
// @Param request body UpdateAIConfigRequest true "Assistant configuration update request"
// @Success 200 {object} AIConfigResponse
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /assistant-config [put]
func (h *AIConfigHandlers) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[UpdateAIConfigRequest](w, r)
	if !ok {
		return
	}

	params, err := buildUpdateParams(req)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	config, err := h.loadOrCreateConfig(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "")
		return
	}

	if err := h.applyAPIKeyEncryption(r.Context(), req.APIKey, &params); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to encrypt API key")
		return
	}

	if err := config.UpdateConfig(params); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	if err := h.repo.Save(r.Context(), config); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to save configuration")
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())
	sharedAPI.RespondJSON(w, http.StatusOK, toResponse(config, actor))
}

func (h *AIConfigHandlers) loadOrCreateConfig(ctx context.Context) (*aggregates.AIConfiguration, error) {
	config, err := h.repo.GetByTenantID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve configuration: %w", err)
	}
	if config == nil {
		tenantID, err := sharedctx.GetTenant(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get tenant: %w", err)
		}
		config = aggregates.NewAIConfiguration(tenantID.Value())
	}
	return config, nil
}

func (h *AIConfigHandlers) applyAPIKeyEncryption(ctx context.Context, rawKey string, params *aggregates.UpdateConfigParams) error {
	if rawKey == "" {
		return nil
	}
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	encrypted, err := crypto.Encrypt(rawKey, tenantID.Value())
	if err != nil {
		return err
	}
	key := vo.NewEncryptedAPIKey(encrypted)
	params.APIKeyEncrypted = &key
	return nil
}

func buildUpdateParams(req UpdateAIConfigRequest) (aggregates.UpdateConfigParams, error) {
	provider, err := vo.NewLLMProvider(req.Provider)
	if err != nil {
		return aggregates.UpdateConfigParams{}, err
	}

	endpointStr := req.Endpoint
	if endpointStr == "" {
		endpointStr = provider.DefaultEndpoint()
	}
	endpoint, err := vo.NewLLMEndpoint(endpointStr)
	if err != nil {
		return aggregates.UpdateConfigParams{}, err
	}

	model, err := vo.NewModelName(req.Model)
	if err != nil {
		return aggregates.UpdateConfigParams{}, err
	}

	maxTokens, err := vo.NewMaxTokens(req.MaxTokens)
	if err != nil {
		return aggregates.UpdateConfigParams{}, err
	}

	temperature, err := vo.NewTemperature(req.Temperature)
	if err != nil {
		return aggregates.UpdateConfigParams{}, err
	}

	return aggregates.UpdateConfigParams{
		Provider:             provider,
		Endpoint:             endpoint,
		Model:                model,
		MaxTokens:            maxTokens,
		Temperature:          temperature,
		SystemPromptOverride: req.SystemPromptOverride,
	}, nil
}

// TestConnection godoc
// @Summary Test assistant provider connection
// @Description Tests connectivity to the configured LLM provider endpoint using the stored encrypted API key for the current tenant.
// @Tags assistant-config
// @Produce json
// @Success 200 {object} TestConnectionResponse
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /assistant-config/test [post]
func (h *AIConfigHandlers) TestConnection(w http.ResponseWriter, r *http.Request) {
	config, err := h.repo.GetByTenantID(r.Context())
	if err != nil || config == nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "No configuration found")
		return
	}

	if !config.Status().IsConfigured() {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "Configuration is not complete")
		return
	}

	tenantID, err := sharedctx.GetTenant(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to get tenant")
		return
	}

	apiKey, err := crypto.Decrypt(config.APIKeyEncrypted().Value(), tenantID.Value())
	if err != nil {
		sharedAPI.RespondJSON(w, http.StatusOK, TestConnectionResponse{
			Success: false,
			Error:   "Failed to decrypt API key",
		})
		return
	}

	result := testLLMConnection(config.Provider(), config.Endpoint().Value(), apiKey, config.Model().Value())
	sharedAPI.RespondJSON(w, http.StatusOK, result)
}

func testLLMConnection(provider vo.LLMProvider, endpoint, apiKey, model string) TestConnectionResponse {
	reqBody := map[string]interface{}{
		"model":      model,
		"messages":   []map[string]string{{"role": "user", "content": "Hello"}},
		"max_tokens": 5,
	}
	body, _ := json.Marshal(reqBody)

	path := "/v1/chat/completions"
	if provider.IsAnthropic() {
		path = "/v1/messages"
	}

	req, err := http.NewRequest("POST", endpoint+path, bytes.NewReader(body))
	if err != nil {
		return TestConnectionResponse{Success: false, Error: fmt.Sprintf("Failed to create request: %v", err)}
	}
	req.Header.Set("Content-Type", "application/json")

	if provider.IsAnthropic() {
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	} else {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	return executeLLMRequest(req)
}

func executeLLMRequest(req *http.Request) TestConnectionResponse {
	client := &http.Client{Timeout: 30 * time.Second}
	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return TestConnectionResponse{Success: false, Error: fmt.Sprintf("Connection failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return TestConnectionResponse{
			Success: false,
			Error:   fmt.Sprintf("LLM returned status %d: %s", resp.StatusCode, string(respBody)),
		}
	}

	var llmResp struct {
		Model string `json:"model"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&llmResp)

	return TestConnectionResponse{
		Success:   true,
		Model:     llmResp.Model,
		LatencyMs: latency,
	}
}

func toResponse(config *aggregates.AIConfiguration, actor sharedctx.Actor) AIConfigResponse {
	apiKeyStatus := "not_configured"
	if !config.APIKeyEncrypted().IsEmpty() {
		apiKeyStatus = "configured"
	}

	return AIConfigResponse{
		ID:                   config.ID(),
		Provider:             config.Provider().Value(),
		Endpoint:             config.Endpoint().Value(),
		APIKeyStatus:         apiKeyStatus,
		Model:                config.Model().Value(),
		MaxTokens:            config.MaxTokens().Value(),
		Temperature:          config.Temperature().Value(),
		SystemPromptOverride: config.SystemPromptOverride(),
		Status:               config.Status().Value(),
		UpdatedAt:            config.UpdatedAt(),
		Links:                configLinks(actor, config.Status().IsConfigured()),
	}
}

func configLinks(actor sharedctx.Actor, isConfigured bool) map[string]sharedAPI.Link {
	links := map[string]sharedAPI.Link{
		"self": {Href: "/api/v1/assistant-config", Method: "GET"},
	}
	if actor.HasPermission("metamodel:write") {
		links["update"] = sharedAPI.Link{Href: "/api/v1/assistant-config", Method: "PUT"}
		if isConfigured {
			links["test"] = sharedAPI.Link{Href: "/api/v1/assistant-config/test", Method: "POST"}
		}
	}
	return links
}
