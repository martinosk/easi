package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validRemoveCmd() *commands.RemoveTimeAssessment {
	return &commands.RemoveTimeAssessment{
		CapabilityID: uuid.New().String(),
		ComponentID:  uuid.New().String(),
		RemovedBy:    "architect@example.com",
	}
}

func TestRemoveTimeAssessmentHandler_ExistingAssessment_RemovesAndSaves(t *testing.T) {
	existing := buildExistingTimeAssessment(t, uuid.New().String(), uuid.New().String(), valueobjects.TimeGradeInvest)
	repo := &mockTimeAssessmentRepository{loaded: existing}
	lookup := &mockExistingTimeAssessmentLookup{id: existing.ID(), exists: true}

	handler := NewRemoveTimeAssessmentHandler(repo, lookup)
	result, err := handler.Handle(context.Background(), validRemoveCmd())

	require.NoError(t, err)
	require.True(t, repo.getCalled)
	require.Len(t, repo.saved, 1)
	assert.True(t, repo.saved[0].IsRemoved())
	assert.Equal(t, existing.ID(), result.CreatedID)
}

func TestRemoveTimeAssessmentHandler_NoAssessmentForPair_Fails(t *testing.T) {
	repo := &mockTimeAssessmentRepository{}
	lookup := &mockExistingTimeAssessmentLookup{exists: false}

	handler := NewRemoveTimeAssessmentHandler(repo, lookup)
	_, err := handler.Handle(context.Background(), validRemoveCmd())

	assert.ErrorIs(t, err, ErrTimeAssessmentNotFoundForPair)
	assert.Empty(t, repo.saved)
	assert.False(t, repo.getCalled)
}

func TestRemoveTimeAssessmentHandler_LookupError_Fails(t *testing.T) {
	repo := &mockTimeAssessmentRepository{}
	lookup := &mockExistingTimeAssessmentLookup{err: errors.New("db down")}

	handler := NewRemoveTimeAssessmentHandler(repo, lookup)
	_, err := handler.Handle(context.Background(), validRemoveCmd())

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}

func TestRemoveTimeAssessmentHandler_LoadError_Fails(t *testing.T) {
	repo := &mockTimeAssessmentRepository{getErr: errors.New("db down")}
	lookup := &mockExistingTimeAssessmentLookup{id: uuid.New().String(), exists: true}

	handler := NewRemoveTimeAssessmentHandler(repo, lookup)
	_, err := handler.Handle(context.Background(), validRemoveCmd())

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}

func TestRemoveTimeAssessmentHandler_AlreadyRemoved_Fails(t *testing.T) {
	existing := buildExistingTimeAssessment(t, uuid.New().String(), uuid.New().String(), valueobjects.TimeGradeInvest)
	require.NoError(t, existing.Remove("a@example.com"))
	repo := &mockTimeAssessmentRepository{loaded: existing}
	lookup := &mockExistingTimeAssessmentLookup{id: existing.ID(), exists: true}

	handler := NewRemoveTimeAssessmentHandler(repo, lookup)
	_, err := handler.Handle(context.Background(), validRemoveCmd())

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}
