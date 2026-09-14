package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingCommandBus struct {
	dispatched []cqrs.Command
	failOn     string
}

func (b *recordingCommandBus) Dispatch(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	b.dispatched = append(b.dispatched, cmd)
	if b.failOn != "" && b.failOn == cmd.CommandName() {
		return cqrs.EmptyResult(), errors.New("dispatch failed")
	}
	return cqrs.EmptyResult(), nil
}

func (b *recordingCommandBus) Register(_ string, _ cqrs.CommandHandler) {}

type noRelations struct{}

func (noRelations) GetBySourceID(_ context.Context, _ string) ([]readmodels.ComponentRelationDTO, error) {
	return nil, nil
}

func (noRelations) GetByTargetID(_ context.Context, _ string) ([]readmodels.ComponentRelationDTO, error) {
	return nil, nil
}

type fakeComponentRecord struct {
	record *readmodels.ApplicationComponentDTO
	err    error
}

func (f *fakeComponentRecord) GetByID(_ context.Context, _ string) (*readmodels.ApplicationComponentDTO, error) {
	return f.record, f.err
}

func deleteHandlerFor(t *testing.T, record *readmodels.ApplicationComponentDTO, bus *recordingCommandBus) (*DeleteApplicationComponentHandler, *mockComponentRepository) {
	t.Helper()
	repo := &mockComponentRepository{loaded: buildOwnershipComponent(t, nil)}
	return NewDeleteApplicationComponentHandler(repo, noRelations{}, &fakeComponentRecord{record: record}, bus), repo
}

func TestDeleteApplicationComponentHandler_DeletesComposedPartsAndReleasesAggregatedParts(t *testing.T) {
	quoting, billing := valueobjects.NewComponentID().Value(), valueobjects.NewComponentID().Value()
	bus := &recordingCommandBus{}
	handler, repo := deleteHandlerFor(t, &readmodels.ApplicationComponentDTO{
		Parts: []readmodels.ContainmentPartDTO{
			{ID: quoting, Name: "Quoting", Kind: valueobjects.ContainmentComposition},
			{ID: billing, Name: "Billing", Kind: valueobjects.ContainmentAggregation},
		},
	}, bus)

	_, err := handler.Handle(context.Background(), &commands.DeleteApplicationComponent{ID: repo.loaded.ID()})

	require.NoError(t, err)
	assert.Equal(t, []cqrs.Command{
		&commands.DeleteApplicationComponent{ID: quoting},
		&commands.DetachComponent{ComponentID: billing},
	}, bus.dispatched)
	require.Len(t, repo.saved, 1)
	assert.True(t, repo.saved[0].IsDeleted())
}

func TestDeleteApplicationComponentHandler_DetachesItselfWhenItIsAPart(t *testing.T) {
	bus := &recordingCommandBus{}
	handler, repo := deleteHandlerFor(t, &readmodels.ApplicationComponentDTO{
		PartOf: &readmodels.ContainmentParentDTO{ID: valueobjects.NewComponentID().Value(), Name: "CRM Suite", Kind: valueobjects.ContainmentComposition},
	}, bus)

	_, err := handler.Handle(context.Background(), &commands.DeleteApplicationComponent{ID: repo.loaded.ID()})

	require.NoError(t, err)
	assert.Equal(t, []cqrs.Command{&commands.DetachComponent{ComponentID: repo.loaded.ID()}}, bus.dispatched)
	require.Len(t, repo.saved, 1)
	assert.True(t, repo.saved[0].IsDeleted())
}

func TestDeleteApplicationComponentHandler_LeavesParentIntactWhenReleasingAPartFails(t *testing.T) {
	bus := &recordingCommandBus{failOn: "DeleteApplicationComponent"}
	handler, repo := deleteHandlerFor(t, &readmodels.ApplicationComponentDTO{
		Parts: []readmodels.ContainmentPartDTO{{ID: valueobjects.NewComponentID().Value(), Name: "Quoting", Kind: valueobjects.ContainmentComposition}},
	}, bus)

	_, err := handler.Handle(context.Background(), &commands.DeleteApplicationComponent{ID: repo.loaded.ID()})

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
	assert.False(t, repo.loaded.IsDeleted())
}

func TestDeleteApplicationComponentHandler_StandaloneComponentDeletesWithoutContainmentCommands(t *testing.T) {
	bus := &recordingCommandBus{}
	handler, repo := deleteHandlerFor(t, &readmodels.ApplicationComponentDTO{}, bus)

	_, err := handler.Handle(context.Background(), &commands.DeleteApplicationComponent{ID: repo.loaded.ID()})

	require.NoError(t, err)
	assert.Empty(t, bus.dispatched)
	require.Len(t, repo.saved, 1)
}

func TestDeleteApplicationComponentHandler_AlreadyDeletedIsNoOp(t *testing.T) {
	bus := &recordingCommandBus{}
	handler, repo := deleteHandlerFor(t, &readmodels.ApplicationComponentDTO{}, bus)
	require.NoError(t, repo.loaded.Delete())
	repo.loaded.MarkChangesAsCommitted()

	_, err := handler.Handle(context.Background(), &commands.DeleteApplicationComponent{ID: repo.loaded.ID()})

	require.NoError(t, err)
	assert.Empty(t, bus.dispatched)
	assert.Empty(t, repo.saved)
}

func TestDeleteApplicationComponentHandler_RejectsWrongCommandType(t *testing.T) {
	handler, _ := deleteHandlerFor(t, nil, &recordingCommandBus{})

	_, err := handler.Handle(context.Background(), &commands.DetachComponent{ComponentID: "c1"})

	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}
