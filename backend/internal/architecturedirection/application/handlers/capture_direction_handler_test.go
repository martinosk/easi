package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDirectionRepository struct {
	saved   []*aggregates.Direction
	saveErr error
}

func (m *mockDirectionRepository) Save(_ context.Context, d *aggregates.Direction) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, d)
	return nil
}

type mockActiveDirectionLookup struct {
	hasActive bool
	err       error
}

func (m *mockActiveDirectionLookup) HasActiveDirectionForEnterpriseCapability(_ context.Context, _ string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.hasActive, nil
}

func validCaptureCmd() *commands.CaptureDirection {
	return &commands.CaptureDirection{
		EnterpriseCapabilityID: uuid.New().String(),
		Type:                   "consolidate",
		SourceCapabilityIDs:    []string{uuid.New().String(), uuid.New().String()},
		Placements:             []commands.PlacementInput{{TargetBusinessDomainID: uuid.New().String()}},
		Horizon:                "next",
		Narrative:              "We consolidate two payroll systems into one.",
	}
}

func TestCaptureDirectionHandler_CreatesDraft(t *testing.T) {
	repo := &mockDirectionRepository{}
	lookup := &mockActiveDirectionLookup{hasActive: false}

	handler := NewCaptureDirectionHandler(repo, lookup)
	result, err := handler.Handle(context.Background(), validCaptureCmd())
	require.NoError(t, err)

	require.Len(t, repo.saved, 1)
	d := repo.saved[0]
	assert.Equal(t, d.ID(), result.CreatedID)
	assert.True(t, d.Status().IsDraft())
}

func TestCaptureDirectionHandler_RejectsSecondActiveDirection(t *testing.T) {
	repo := &mockDirectionRepository{}
	lookup := &mockActiveDirectionLookup{hasActive: true}

	handler := NewCaptureDirectionHandler(repo, lookup)
	_, err := handler.Handle(context.Background(), validCaptureCmd())
	assert.ErrorIs(t, err, ErrActiveDirectionAlreadyExists)
	assert.Empty(t, repo.saved)
}

func TestCaptureDirectionHandler_InvalidType_Fails(t *testing.T) {
	repo := &mockDirectionRepository{}
	lookup := &mockActiveDirectionLookup{}
	handler := NewCaptureDirectionHandler(repo, lookup)

	cmd := validCaptureCmd()
	cmd.Type = "explode"
	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}

func TestCaptureDirectionHandler_InvalidSourceCount_Fails(t *testing.T) {
	repo := &mockDirectionRepository{}
	lookup := &mockActiveDirectionLookup{}
	handler := NewCaptureDirectionHandler(repo, lookup)

	cmd := validCaptureCmd()
	cmd.SourceCapabilityIDs = []string{uuid.New().String()} // only 1 for consolidate
	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, aggregates.ErrInvalidSourceCardinality)
}

func TestCaptureDirectionHandler_LookupError_Fails(t *testing.T) {
	repo := &mockDirectionRepository{}
	lookup := &mockActiveDirectionLookup{err: errors.New("db error")}
	handler := NewCaptureDirectionHandler(repo, lookup)

	_, err := handler.Handle(context.Background(), validCaptureCmd())
	assert.Error(t, err)
}
