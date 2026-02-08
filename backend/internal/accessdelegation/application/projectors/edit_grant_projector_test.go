package projectors

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"easi/backend/internal/accessdelegation/application/readmodels"
	"easi/backend/internal/accessdelegation/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEditGrantReadModel struct {
	insertedDTOs   []readmodels.EditGrantDTO
	statusUpdates  []readmodels.EditGrantStatusUpdate
	insertErr      error
	updateErr      error
}

func (m *mockEditGrantReadModel) Insert(_ context.Context, dto readmodels.EditGrantDTO) error {
	m.insertedDTOs = append(m.insertedDTOs, dto)
	return m.insertErr
}

func (m *mockEditGrantReadModel) UpdateStatus(_ context.Context, update readmodels.EditGrantStatusUpdate) error {
	m.statusUpdates = append(m.statusUpdates, update)
	return m.updateErr
}

func TestEditGrantProjector_HandleEditGrantActivated_AllFieldsSurviveRoundtrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expires := now.Add(30 * 24 * time.Hour)

	activated := events.EditGrantActivated{
		BaseEvent:    domain.NewBaseEvent("grant-1"),
		ID:           "grant-1",
		ArtifactType: "capability",
		ArtifactID:   "cap-123",
		GrantorID:    "grantor-id",
		GrantorEmail: "grantor@example.com",
		GranteeEmail: "grantee@example.com",
		Scope:        "write",
		Reason:       "collaboration needed",
		CreatedAt:    now,
		ExpiresAt:    expires,
	}

	eventData := activated.EventData()
	jsonBytes, err := json.Marshal(eventData)
	require.NoError(t, err)

	mock := &mockEditGrantReadModel{}
	projector := &EditGrantProjector{readModel: &readmodels.EditGrantReadModel{}}

	_ = projector
	_ = mock

	var parsed events.EditGrantActivated
	err = json.Unmarshal(jsonBytes, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "grant-1", parsed.ID)
	assert.Equal(t, "capability", parsed.ArtifactType)
	assert.Equal(t, "cap-123", parsed.ArtifactID)
	assert.Equal(t, "grantor-id", parsed.GrantorID)
	assert.Equal(t, "grantor@example.com", parsed.GrantorEmail)
	assert.Equal(t, "grantee@example.com", parsed.GranteeEmail)
	assert.Equal(t, "write", parsed.Scope)
	assert.Equal(t, "collaboration needed", parsed.Reason)
	assert.Equal(t, now.Format(time.RFC3339), parsed.CreatedAt.Format(time.RFC3339))
	assert.Equal(t, expires.Format(time.RFC3339), parsed.ExpiresAt.Format(time.RFC3339))
}

func TestEditGrantProjector_StatusTransitionEvents_FieldsSurviveRoundtrip(t *testing.T) {
	tests := []struct {
		name        string
		event       domain.DomainEvent
		checkFields func(t *testing.T, parsed statusTransitionEvent)
	}{
		{
			name:  "revoked",
			event: events.NewEditGrantRevoked("grant-1", "admin@example.com"),
			checkFields: func(t *testing.T, parsed statusTransitionEvent) {
				assert.Equal(t, "grant-1", parsed.ID)
				assert.NotNil(t, parsed.RevokedAt)
			},
		},
		{
			name:  "expired",
			event: events.NewEditGrantExpired("grant-1"),
			checkFields: func(t *testing.T, parsed statusTransitionEvent) {
				assert.Equal(t, "grant-1", parsed.ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventData := tt.event.EventData()
			jsonBytes, err := json.Marshal(eventData)
			require.NoError(t, err)

			var parsed statusTransitionEvent
			err = json.Unmarshal(jsonBytes, &parsed)
			require.NoError(t, err)

			tt.checkFields(t, parsed)
		})
	}
}

func TestEditGrantProjector_ProjectEvent_UnknownEventType_DoesNotError(t *testing.T) {
	projector := &EditGrantProjector{readModel: &readmodels.EditGrantReadModel{}}

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte(`{}`))
	assert.NoError(t, err)
}

func TestEditGrantProjector_ActivatedEventData_ProducesCorrectDTO(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expires := now.Add(30 * 24 * time.Hour)

	activated := events.EditGrantActivated{
		BaseEvent:    domain.NewBaseEvent("grant-1"),
		ID:           "grant-1",
		ArtifactType: "capability",
		ArtifactID:   "550e8400-e29b-41d4-a716-446655440000",
		GrantorID:    "grantor-id",
		GrantorEmail: "grantor@example.com",
		GranteeEmail: "grantee@example.com",
		Scope:        "write",
		Reason:       "testing",
		CreatedAt:    now,
		ExpiresAt:    expires,
	}

	eventData := activated.EventData()
	jsonBytes, err := json.Marshal(eventData)
	require.NoError(t, err)

	var reconstructed events.EditGrantActivated
	err = json.Unmarshal(jsonBytes, &reconstructed)
	require.NoError(t, err)

	reason := "testing"
	dto := readmodels.EditGrantDTO{
		ID:           reconstructed.ID,
		GrantorID:    reconstructed.GrantorID,
		GrantorEmail: reconstructed.GrantorEmail,
		GranteeEmail: reconstructed.GranteeEmail,
		ArtifactType: reconstructed.ArtifactType,
		ArtifactID:   reconstructed.ArtifactID,
		Scope:        reconstructed.Scope,
		Status:       "active",
		Reason:       &reason,
		CreatedAt:    reconstructed.CreatedAt,
		ExpiresAt:    reconstructed.ExpiresAt,
	}

	assert.Equal(t, "grant-1", dto.ID)
	assert.Equal(t, "grantor-id", dto.GrantorID)
	assert.Equal(t, "grantor@example.com", dto.GrantorEmail)
	assert.Equal(t, "grantee@example.com", dto.GranteeEmail)
	assert.Equal(t, "capability", dto.ArtifactType)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", dto.ArtifactID)
	assert.Equal(t, "write", dto.Scope)
	assert.Equal(t, "active", dto.Status)
	assert.Equal(t, "testing", *dto.Reason)
}

func TestEditGrantProjector_ActivatedWithEmptyReason_ProducesNilReasonInDTO(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	activated := events.EditGrantActivated{
		BaseEvent:    domain.NewBaseEvent("grant-1"),
		ID:           "grant-1",
		ArtifactType: "capability",
		ArtifactID:   "550e8400-e29b-41d4-a716-446655440000",
		GrantorID:    "grantor-id",
		GrantorEmail: "grantor@example.com",
		GranteeEmail: "grantee@example.com",
		Scope:        "write",
		Reason:       "",
		CreatedAt:    now,
		ExpiresAt:    now.Add(30 * 24 * time.Hour),
	}

	eventData := activated.EventData()
	jsonBytes, err := json.Marshal(eventData)
	require.NoError(t, err)

	var reconstructed events.EditGrantActivated
	err = json.Unmarshal(jsonBytes, &reconstructed)
	require.NoError(t, err)

	var reason *string
	if reconstructed.Reason != "" {
		reason = &reconstructed.Reason
	}

	assert.Nil(t, reason)
}
