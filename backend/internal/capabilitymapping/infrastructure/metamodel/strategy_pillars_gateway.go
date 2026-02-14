package metamodel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	sharedctx "easi/backend/internal/shared/context"
)

type StrategyPillarDTO = mmPL.StrategyPillarDTO
type StrategyPillarsConfigDTO = mmPL.StrategyPillarsConfigDTO
type StrategyPillarsGateway = mmPL.StrategyPillarsGateway

var DefaultStrategyPillarsConfig = mmPL.DefaultStrategyPillarsConfig

type pillarCacheEntry struct {
	config    *StrategyPillarsConfigDTO
	expiresAt time.Time
}

type httpStrategyPillarsGateway struct {
	baseURL    string
	httpClient *http.Client
	cacheTTL   time.Duration

	mu    sync.RWMutex
	cache map[string]*pillarCacheEntry
}

func NewStrategyPillarsGateway(baseURL string) StrategyPillarsGateway {
	return &httpStrategyPillarsGateway{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		cacheTTL:   5 * time.Minute,
		cache:      make(map[string]*pillarCacheEntry),
	}
}

func NewStrategyPillarsGatewayWithClient(baseURL string, client *http.Client, cacheTTL time.Duration) StrategyPillarsGateway {
	return &httpStrategyPillarsGateway{
		baseURL:    baseURL,
		httpClient: client,
		cacheTTL:   cacheTTL,
		cache:      make(map[string]*pillarCacheEntry),
	}
}

func (g *httpStrategyPillarsGateway) GetStrategyPillars(ctx context.Context) (*StrategyPillarsConfigDTO, error) {
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

func (g *httpStrategyPillarsGateway) GetActivePillar(ctx context.Context, pillarID string) (*StrategyPillarDTO, error) {
	config, err := g.GetStrategyPillars(ctx)
	if err != nil {
		return nil, err
	}

	for _, pillar := range config.Pillars {
		if pillar.ID == pillarID && pillar.Active {
			return &pillar, nil
		}
	}

	return nil, nil
}

func (g *httpStrategyPillarsGateway) InvalidateCache(tenantID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.cache, tenantID)
}

func (g *httpStrategyPillarsGateway) getFromCache(tenantID string) *StrategyPillarsConfigDTO {
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

func (g *httpStrategyPillarsGateway) setInCache(tenantID string, config *StrategyPillarsConfigDTO) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.cache[tenantID] = &pillarCacheEntry{
		config:    config,
		expiresAt: time.Now().Add(g.cacheTTL),
	}
}

func (g *httpStrategyPillarsGateway) fetchFromAPI(ctx context.Context, tenantID string) (*StrategyPillarsConfigDTO, error) {
	url := fmt.Sprintf("%s/api/v1/meta-model/strategy-pillars", g.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Tenant-ID", tenantID)
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch strategy pillars: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return DefaultStrategyPillarsConfig(), nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var response StrategyPillarsConfigDTO
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}
