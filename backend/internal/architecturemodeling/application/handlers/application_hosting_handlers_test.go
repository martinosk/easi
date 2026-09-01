package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyApplicationHostingHandler_Classifies(t *testing.T) {
	component := buildOwnershipComponent(t, nil)
	repo := &mockComponentRepository{loaded: component}
	handler := NewClassifyApplicationHostingHandler(repo)

	_, err := handler.Handle(context.Background(), &commands.ClassifyApplicationHosting{
		ComponentID: component.ID(),
		Hosting:     valueobjects.HostingSaaS,
	})

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, valueobjects.HostingSaaS, repo.saved[0].Hosting().String())
}

func TestClassifyApplicationHostingHandler_RejectsInvalidClassification(t *testing.T) {
	repo := &mockComponentRepository{loaded: buildOwnershipComponent(t, nil)}
	handler := NewClassifyApplicationHostingHandler(repo)

	_, err := handler.Handle(context.Background(), &commands.ClassifyApplicationHosting{
		ComponentID: "comp-1",
		Hosting:     "mainframe",
	})

	assert.ErrorIs(t, err, valueobjects.ErrInvalidHostingClassification)
	assert.Empty(t, repo.saved)
}

func TestClassifyApplicationHostingHandler_LoadErrorFails(t *testing.T) {
	repo := &mockComponentRepository{getErr: errors.New("not found")}
	handler := NewClassifyApplicationHostingHandler(repo)

	_, err := handler.Handle(context.Background(), &commands.ClassifyApplicationHosting{
		ComponentID: "comp-1",
		Hosting:     valueobjects.HostingCloud,
	})

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}

func TestClassifyApplicationHostingHandler_RejectsWrongCommandType(t *testing.T) {
	handler := NewClassifyApplicationHostingHandler(&mockComponentRepository{})

	_, err := handler.Handle(context.Background(), &commands.ClearApplicationComponentOwnership{ComponentID: "comp-1"})

	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}
