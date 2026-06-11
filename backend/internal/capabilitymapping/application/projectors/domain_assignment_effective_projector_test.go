package projectors

import (
	"context"
	"encoding/json"
	"testing"

	"easi/backend/internal/capabilitymapping/domain/events"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakePillarsGateway struct {
	config *mmPL.StrategyPillarsConfigDTO
}

func (f *fakePillarsGateway) GetStrategyPillars(_ context.Context) (*mmPL.StrategyPillarsConfigDTO, error) {
	return f.config, nil
}

func (f *fakePillarsGateway) GetActivePillar(_ context.Context, _ string) (*mmPL.StrategyPillarDTO, error) {
	return nil, nil
}

func (f *fakePillarsGateway) InvalidateCache(_ string) {}

func TestDomainAssignmentEffective_DefaultSentinelPillarsAreSkipped(t *testing.T) {
	projector := NewDomainAssignmentEffectiveProjector(nil, nil, &fakePillarsGateway{
		config: mmPL.DefaultStrategyPillarsConfig(),
	})

	event := events.NewCapabilityAssignedToDomain(uuid.New().String(), uuid.New().String(), uuid.New().String())
	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "CapabilityAssignedToDomain", eventData)

	require.NoError(t, err)
}
