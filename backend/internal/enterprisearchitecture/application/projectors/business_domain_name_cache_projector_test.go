package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type businessDomainNameUpsert struct {
	BusinessDomainID string
	Name             string
}

type mockBusinessDomainNameCacheWriter struct {
	upserted  []businessDomainNameUpsert
	deleted   []string
	upsertErr error
	deleteErr error
}

func (m *mockBusinessDomainNameCacheWriter) Upsert(ctx context.Context, businessDomainID, name string) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.upserted = append(m.upserted, businessDomainNameUpsert{BusinessDomainID: businessDomainID, Name: name})
	return nil
}

func (m *mockBusinessDomainNameCacheWriter) Delete(ctx context.Context, businessDomainID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deleted = append(m.deleted, businessDomainID)
	return nil
}

func TestBusinessDomainNameCacheProjector_CreatedAndUpdated_UpsertName(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
	}{
		{"created", cmPL.BusinessDomainCreated},
		{"updated", cmPL.BusinessDomainUpdated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockBusinessDomainNameCacheWriter{}
			projector := NewBusinessDomainNameCacheProjector(mock)

			domainID := uuid.New().String()
			eventData, err := json.Marshal(businessDomainNameEvent{ID: domainID, Name: "Customer Experience"})
			require.NoError(t, err)

			require.NoError(t, projector.ProjectEvent(context.Background(), tt.eventType, eventData))

			require.Len(t, mock.upserted, 1)
			assert.Equal(t, domainID, mock.upserted[0].BusinessDomainID)
			assert.Equal(t, "Customer Experience", mock.upserted[0].Name)
		})
	}
}

func TestBusinessDomainNameCacheProjector_Deleted_RemovesName(t *testing.T) {
	mock := &mockBusinessDomainNameCacheWriter{}
	projector := NewBusinessDomainNameCacheProjector(mock)

	domainID := uuid.New().String()
	eventData, err := json.Marshal(businessDomainIdentifiedEvent{ID: domainID})
	require.NoError(t, err)

	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.BusinessDomainDeleted, eventData))

	require.Len(t, mock.deleted, 1)
	assert.Equal(t, domainID, mock.deleted[0])
}

func TestBusinessDomainNameCacheProjector_UnknownEvent_Ignored(t *testing.T) {
	mock := &mockBusinessDomainNameCacheWriter{}
	projector := NewBusinessDomainNameCacheProjector(mock)

	err := projector.ProjectEvent(context.Background(), "SomeOtherEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mock.upserted)
	assert.Empty(t, mock.deleted)
}

func TestBusinessDomainNameCacheProjector_InvalidJSON_ReturnsError(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
	}{
		{"created", cmPL.BusinessDomainCreated},
		{"updated", cmPL.BusinessDomainUpdated},
		{"deleted", cmPL.BusinessDomainDeleted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockBusinessDomainNameCacheWriter{}
			projector := NewBusinessDomainNameCacheProjector(mock)

			err := projector.ProjectEvent(context.Background(), tt.eventType, []byte("invalid"))
			assert.Error(t, err)
		})
	}
}

func TestBusinessDomainNameCacheProjector_ErrorPropagation(t *testing.T) {
	tests := []struct {
		name      string
		mock      *mockBusinessDomainNameCacheWriter
		eventType string
		eventData any
	}{
		{
			"upsert error",
			&mockBusinessDomainNameCacheWriter{upsertErr: errors.New("db error")},
			cmPL.BusinessDomainCreated,
			businessDomainNameEvent{ID: uuid.New().String(), Name: "Finance"},
		},
		{
			"delete error",
			&mockBusinessDomainNameCacheWriter{deleteErr: errors.New("db error")},
			cmPL.BusinessDomainDeleted,
			businessDomainIdentifiedEvent{ID: uuid.New().String()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projector := NewBusinessDomainNameCacheProjector(tt.mock)
			eventData, err := json.Marshal(tt.eventData)
			require.NoError(t, err)

			err = projector.ProjectEvent(context.Background(), tt.eventType, eventData)
			assert.Error(t, err)
		})
	}
}
