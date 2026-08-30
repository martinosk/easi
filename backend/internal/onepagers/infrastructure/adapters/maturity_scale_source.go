package adapters

import (
	"context"
	"fmt"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/readmodels"
)

type MaturityScaleCache interface {
	Sections(ctx context.Context) ([]readmodels.MaturityScaleSection, error)
}

type maturityScaleSource struct {
	cache MaturityScaleCache
}

func NewOnePagerMaturityScaleSource(db *database.TenantAwareDB) ports.MaturityScaleSource {
	return NewMaturityScaleSource(readmodels.NewMaturityScaleCacheReadModel(db))
}

func NewMaturityScaleSource(cache MaturityScaleCache) ports.MaturityScaleSource {
	return maturityScaleSource{cache: cache}
}

func (s maturityScaleSource) Sections(ctx context.Context) ([]ports.MaturitySection, error) {
	cached, err := s.cache.Sections(ctx)
	if err != nil {
		return nil, fmt.Errorf("read cached maturity scale for one-pagers: %w", err)
	}
	if cached == nil {
		return nil, nil
	}
	sections := make([]ports.MaturitySection, len(cached))
	for i, section := range cached {
		sections[i] = ports.MaturitySection{Name: section.Name, MinValue: section.MinValue, MaxValue: section.MaxValue}
	}
	return sections, nil
}
