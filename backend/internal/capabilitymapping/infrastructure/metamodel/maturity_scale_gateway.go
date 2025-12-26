package metamodel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	sharedctx "easi/backend/internal/shared/context"
)

type MaturitySectionDTO struct {
	Order    int    `json:"order"`
	Name     string `json:"name"`
	MinValue int    `json:"minValue"`
	MaxValue int    `json:"maxValue"`
}

type MaturityScaleConfigDTO struct {
	Sections []MaturitySectionDTO `json:"sections"`
}

type MaturityScaleGateway interface {
	GetMaturityScaleConfig(ctx context.Context) (*MaturityScaleConfigDTO, error)
	InvalidateCache(tenantID string)
}

type cacheEntry struct {
	config    *MaturityScaleConfigDTO
	expiresAt time.Time
}

type httpMaturityScaleGateway struct {
	baseURL    string
	httpClient *http.Client
	cacheTTL   time.Duration

	mu    sync.RWMutex
	cache map[string]*cacheEntry
}

func NewMaturityScaleGateway(baseURL string) MaturityScaleGateway {
	return &httpMaturityScaleGateway{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		cacheTTL:   5 * time.Minute,
		cache:      make(map[string]*cacheEntry),
	}
}

func NewMaturityScaleGatewayWithClient(baseURL string, client *http.Client, cacheTTL time.Duration) MaturityScaleGateway {
	return &httpMaturityScaleGateway{
		baseURL:    baseURL,
		httpClient: client,
		cacheTTL:   cacheTTL,
		cache:      make(map[string]*cacheEntry),
	}
}

func (g *httpMaturityScaleGateway) GetMaturityScaleConfig(ctx context.Context) (*MaturityScaleConfigDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	if cached := g.getFromCache(tenantID.Value()); cached != nil {
		return cached, nil
	}

	config, err := g.fetchFromAPI(ctx, tenantID.Value())
	if err != nil {
		return nil, err
	}

	g.setInCache(tenantID.Value(), config)
	return config, nil
}

func (g *httpMaturityScaleGateway) InvalidateCache(tenantID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.cache, tenantID)
}

func (g *httpMaturityScaleGateway) getFromCache(tenantID string) *MaturityScaleConfigDTO {
	g.mu.RLock()
	defer g.mu.RUnlock()

	entry, exists := g.cache[tenantID]
	if !exists {
		return nil
	}

	if time.Now().After(entry.expiresAt) {
		return nil
	}

	return entry.config
}

func (g *httpMaturityScaleGateway) setInCache(tenantID string, config *MaturityScaleConfigDTO) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.cache[tenantID] = &cacheEntry{
		config:    config,
		expiresAt: time.Now().Add(g.cacheTTL),
	}
}

type apiResponse struct {
	Sections []MaturitySectionDTO `json:"sections"`
}

func (g *httpMaturityScaleGateway) fetchFromAPI(ctx context.Context, tenantID string) (*MaturityScaleConfigDTO, error) {
	url := fmt.Sprintf("%s/api/v1/meta-model/maturity-scale", g.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Tenant-ID", tenantID)
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch maturity scale config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var response apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &MaturityScaleConfigDTO{
		Sections: response.Sections,
	}, nil
}

func DefaultMaturityScaleConfig() *MaturityScaleConfigDTO {
	return &MaturityScaleConfigDTO{
		Sections: []MaturitySectionDTO{
			{Order: 1, Name: "Genesis", MinValue: 0, MaxValue: 24},
			{Order: 2, Name: "Custom Built", MinValue: 25, MaxValue: 49},
			{Order: 3, Name: "Product", MinValue: 50, MaxValue: 74},
			{Order: 4, Name: "Commodity", MinValue: 75, MaxValue: 99},
		},
	}
}
