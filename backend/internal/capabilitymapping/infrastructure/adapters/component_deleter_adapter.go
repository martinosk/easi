package adapters

import (
	"context"
	"fmt"
	"net/http"

	sharedctx "easi/backend/internal/shared/context"
)

type LoopbackComponentDeleter struct {
	client  *http.Client
	baseURL string
}

func NewLoopbackComponentDeleter(baseURL string) *LoopbackComponentDeleter {
	return &LoopbackComponentDeleter{
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

func (d *LoopbackComponentDeleter) DeleteComponent(ctx context.Context, componentID string) error {
	url := fmt.Sprintf("%s/api/v1/components/%s", d.baseURL, componentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err == nil {
		req.Header.Set("X-Tenant-ID", tenantID.Value())
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to delete component %s: status %d", componentID, resp.StatusCode)
	}

	return nil
}

type NoOpComponentDeleter struct{}

func NewNoOpComponentDeleter() *NoOpComponentDeleter {
	return &NoOpComponentDeleter{}
}

func (d *NoOpComponentDeleter) DeleteComponent(ctx context.Context, componentID string) error {
	return nil
}
