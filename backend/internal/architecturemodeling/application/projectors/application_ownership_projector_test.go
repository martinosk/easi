package projectors

import (
	"context"
	"testing"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ownershipWrite struct {
	componentID string
	record      readmodels.OwnershipRecord
}

type fakeOwnershipWriter struct {
	writes  []ownershipWrite
	cleared []string
}

func (f *fakeOwnershipWriter) SetOwnership(_ context.Context, componentID string, record readmodels.OwnershipRecord) error {
	f.writes = append(f.writes, ownershipWrite{componentID, record})
	return nil
}

func (f *fakeOwnershipWriter) ClearOwnership(_ context.Context, componentID string) error {
	f.cleared = append(f.cleared, componentID)
	return nil
}

func TestApplicationOwnershipProjector_WritesOwnership(t *testing.T) {
	cases := []struct {
		name  string
		event domain.DomainEvent
		want  readmodels.OwnershipRecord
	}{
		{
			name:  "nominated",
			event: events.NewApplicationOwnerNominated("comp-1", valueobjects.OwnerKindUser, "user-1"),
			want:  readmodels.OwnershipRecord{State: valueobjects.OwnershipStateNominated, OwnerKind: valueobjects.OwnerKindUser, OwnerID: "user-1"},
		},
		{
			name:  "confirmed",
			event: events.NewApplicationOwnershipConfirmed("comp-1", valueobjects.OwnerKindTeam, "team-1", valueobjects.OwnershipStateManaged),
			want:  readmodels.OwnershipRecord{State: valueobjects.OwnershipStateManaged, OwnerKind: valueobjects.OwnerKindTeam, OwnerID: "team-1"},
		},
		{
			name:  "assigned",
			event: events.NewApplicationOwnerAssigned("comp-1", valueobjects.OwnerKindUser, "user-1", valueobjects.OwnershipStateOwned),
			want:  readmodels.OwnershipRecord{State: valueobjects.OwnershipStateOwned, OwnerKind: valueobjects.OwnerKindUser, OwnerID: "user-1"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			writer := &fakeOwnershipWriter{}
			projector := NewApplicationOwnershipProjector(writer)

			require.NoError(t, projector.Handle(context.Background(), tc.event))

			require.Len(t, writer.writes, 1)
			assert.Equal(t, ownershipWrite{"comp-1", tc.want}, writer.writes[0])
		})
	}
}

func TestApplicationOwnershipProjector_Cleared(t *testing.T) {
	writer := &fakeOwnershipWriter{}
	projector := NewApplicationOwnershipProjector(writer)

	require.NoError(t, projector.Handle(context.Background(), events.NewApplicationOwnershipCleared("comp-1")))

	assert.Empty(t, writer.writes)
	assert.Equal(t, []string{"comp-1"}, writer.cleared)
}

func TestApplicationOwnershipProjector_IgnoresOtherEvents(t *testing.T) {
	writer := &fakeOwnershipWriter{}
	projector := NewApplicationOwnershipProjector(writer)

	require.NoError(t, projector.Handle(context.Background(), events.NewApplicationComponentCreated("comp-1", "Billing", "")))

	assert.Empty(t, writer.writes)
	assert.Empty(t, writer.cleared)
}
