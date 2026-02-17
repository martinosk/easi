package domain

import (
	"context"

	"easi/backend/internal/archassistant/domain/aggregates"
)

type AIConfigurationRepository interface {
	GetByTenantID(ctx context.Context) (*aggregates.AIConfiguration, error)
	Save(ctx context.Context, config *aggregates.AIConfiguration) error
}
