package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestValueStream(t *testing.T) *aggregates.ValueStream {
	t.Helper()
	name, err := valueobjects.NewValueStreamName("Test Value Stream")
	require.NoError(t, err)
	desc := valueobjects.MustNewDescription("Test")
	vs, err := aggregates.NewValueStream(name, desc)
	require.NoError(t, err)
	vs.MarkChangesAsCommitted()
	return vs
}

type mockStageRepository struct {
	stream  *aggregates.ValueStream
	getErr  error
	saveErr error
	saved   []*aggregates.ValueStream
}

func (m *mockStageRepository) GetByID(ctx context.Context, id string) (*aggregates.ValueStream, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.stream, nil
}

func (m *mockStageRepository) Save(ctx context.Context, vs *aggregates.ValueStream) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, vs)
	return nil
}

func TestAddStageHandler_Success(t *testing.T) {
	vs := newTestValueStream(t)
	repo := &mockStageRepository{stream: vs}
	handler := NewAddStageHandler(repo)

	cmd := &commands.AddStage{
		ValueStreamID: vs.ID(),
		Name:          "Discovery",
		Description:   "Initial discovery phase",
	}

	result, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
	assert.NotEmpty(t, result.CreatedID)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, 1, repo.saved[0].StageCount())
}

func TestAddStageHandler_DuplicateName(t *testing.T) {
	vs := newTestValueStream(t)
	name, _ := valueobjects.NewStageName("Discovery")
	desc := valueobjects.MustNewDescription("")
	vs.AddStage(name, desc, nil)
	vs.MarkChangesAsCommitted()

	repo := &mockStageRepository{stream: vs}
	handler := NewAddStageHandler(repo)

	cmd := &commands.AddStage{
		ValueStreamID: vs.ID(),
		Name:          "Discovery",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrStageNameExists)
}

func TestAddStageHandler_InvalidName(t *testing.T) {
	vs := newTestValueStream(t)
	repo := &mockStageRepository{stream: vs}
	handler := NewAddStageHandler(repo)

	cmd := &commands.AddStage{
		ValueStreamID: vs.ID(),
		Name:          "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}

func TestAddStageHandler_InvalidCommand(t *testing.T) {
	repo := &mockStageRepository{}
	handler := NewAddStageHandler(repo)

	_, err := handler.Handle(context.Background(), &commands.RemoveStage{})
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}
