package projectors

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeContainmentWriter struct {
	set         []readmodels.ContainmentRecord
	cleared     []string
	registered  []string
	registerErr error
}

func (f *fakeContainmentWriter) SetParent(_ context.Context, record readmodels.ContainmentRecord) error {
	f.set = append(f.set, record)
	return nil
}

func (f *fakeContainmentWriter) ClearParent(_ context.Context, partID string) error {
	f.cleared = append(f.cleared, partID)
	return nil
}

func (f *fakeContainmentWriter) RegisterContainmentsAggregate(_ context.Context, aggregateID string) error {
	f.registered = append(f.registered, aggregateID)
	return f.registerErr
}

func TestComponentContainmentProjector_AttachedRegistersAggregateAndSetsParent(t *testing.T) {
	writer := &fakeContainmentWriter{}
	projector := NewComponentContainmentProjector(writer)

	require.NoError(t, projector.Handle(context.Background(), events.NewComponentAttached("agg-1", "quoting", "crm", valueobjects.ContainmentComposition)))

	assert.Equal(t, []string{"agg-1"}, writer.registered)
	assert.Equal(t, []readmodels.ContainmentRecord{{PartID: "quoting", ParentID: "crm", Kind: valueobjects.ContainmentComposition}}, writer.set)
	assert.Empty(t, writer.cleared)
}

func TestComponentContainmentProjector_AttachedFailsWhenRegistryConflicts(t *testing.T) {
	conflict := errors.New("tenant already has a component containments aggregate")
	writer := &fakeContainmentWriter{registerErr: conflict}
	projector := NewComponentContainmentProjector(writer)

	err := projector.Handle(context.Background(), events.NewComponentAttached("agg-2", "quoting", "crm", valueobjects.ContainmentComposition))

	assert.ErrorIs(t, err, conflict)
	assert.Empty(t, writer.set)
}

func TestComponentContainmentProjector_DetachedClearsParent(t *testing.T) {
	writer := &fakeContainmentWriter{}
	projector := NewComponentContainmentProjector(writer)

	require.NoError(t, projector.Handle(context.Background(), events.NewComponentDetached("agg-1", "quoting", "crm")))

	assert.Equal(t, []string{"quoting"}, writer.cleared)
	assert.Empty(t, writer.set)
	assert.Empty(t, writer.registered)
}

func TestComponentContainmentProjector_IgnoresOtherEvents(t *testing.T) {
	writer := &fakeContainmentWriter{}
	projector := NewComponentContainmentProjector(writer)

	require.NoError(t, projector.Handle(context.Background(), events.NewApplicationComponentCreated("comp-1", "Billing", "")))

	assert.Empty(t, writer.set)
	assert.Empty(t, writer.cleared)
	assert.Empty(t, writer.registered)
}
