package projectors

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/accessdelegation/application/readmodels"
	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	viewsPL "easi/backend/internal/architectureviews/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type recordingArtifactNameStore struct {
	upserted []readmodels.ArtifactNameDTO
	deleted  []readmodels.ArtifactRef
	err      error
}

func (s *recordingArtifactNameStore) Upsert(_ context.Context, dto readmodels.ArtifactNameDTO) error {
	if s.err != nil {
		return s.err
	}
	s.upserted = append(s.upserted, dto)
	return nil
}

func (s *recordingArtifactNameStore) Delete(_ context.Context, artifactType, artifactID string) error {
	if s.err != nil {
		return s.err
	}
	s.deleted = append(s.deleted, readmodels.ArtifactRef{ArtifactType: artifactType, ArtifactID: artifactID})
	return nil
}

type supplierEvent struct {
	aggregateID string
	eventType   string
	data        map[string]interface{}
}

func (e supplierEvent) AggregateID() string               { return e.aggregateID }
func (e supplierEvent) EventType() string                 { return e.eventType }
func (e supplierEvent) OccurredAt() time.Time             { return time.Now() }
func (e supplierEvent) EventData() map[string]interface{} { return e.data }

func projectNameCacheEvent(t *testing.T, event domain.DomainEvent) *recordingArtifactNameStore {
	t.Helper()
	store := &recordingArtifactNameStore{}
	require.NoError(t, NewArtifactNameCacheProjector(store).Handle(context.Background(), event))
	return store
}

func TestArtifactNameCacheProjector_CachesArtifactNames(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		payload   map[string]interface{}
		expected  readmodels.ArtifactNameDTO
	}{
		{"capability created", capPL.CapabilityCreated, map[string]interface{}{"id": "cap-1", "name": "Onboarding"},
			readmodels.ArtifactNameDTO{ArtifactType: "capability", ArtifactID: "cap-1", Name: "Onboarding"}},
		{"capability updated", capPL.CapabilityUpdated, map[string]interface{}{"id": "cap-1", "name": "Customer Onboarding"},
			readmodels.ArtifactNameDTO{ArtifactType: "capability", ArtifactID: "cap-1", Name: "Customer Onboarding"}},
		{"business domain created", capPL.BusinessDomainCreated, map[string]interface{}{"id": "bd-1", "name": "Sales"},
			readmodels.ArtifactNameDTO{ArtifactType: "domain", ArtifactID: "bd-1", Name: "Sales"}},
		{"business domain updated", capPL.BusinessDomainUpdated, map[string]interface{}{"id": "bd-1", "name": "Sales & Marketing"},
			readmodels.ArtifactNameDTO{ArtifactType: "domain", ArtifactID: "bd-1", Name: "Sales & Marketing"}},
		{"component created", archPL.ApplicationComponentCreated, map[string]interface{}{"id": "comp-1", "name": "Payment Service"},
			readmodels.ArtifactNameDTO{ArtifactType: "component", ArtifactID: "comp-1", Name: "Payment Service"}},
		{"component updated", archPL.ApplicationComponentUpdated, map[string]interface{}{"id": "comp-1", "name": "Payments"},
			readmodels.ArtifactNameDTO{ArtifactType: "component", ArtifactID: "comp-1", Name: "Payments"}},
		{"vendor created", archPL.VendorCreated, map[string]interface{}{"id": "ven-1", "name": "Acme"},
			readmodels.ArtifactNameDTO{ArtifactType: "vendor", ArtifactID: "ven-1", Name: "Acme"}},
		{"vendor updated", archPL.VendorUpdated, map[string]interface{}{"id": "ven-1", "name": "Acme Corp"},
			readmodels.ArtifactNameDTO{ArtifactType: "vendor", ArtifactID: "ven-1", Name: "Acme Corp"}},
		{"acquired entity created", archPL.AcquiredEntityCreated, map[string]interface{}{"id": "ae-1", "name": "Widget Co"},
			readmodels.ArtifactNameDTO{ArtifactType: "acquired_entity", ArtifactID: "ae-1", Name: "Widget Co"}},
		{"acquired entity updated", archPL.AcquiredEntityUpdated, map[string]interface{}{"id": "ae-1", "name": "Widget Group"},
			readmodels.ArtifactNameDTO{ArtifactType: "acquired_entity", ArtifactID: "ae-1", Name: "Widget Group"}},
		{"internal team created", archPL.InternalTeamCreated, map[string]interface{}{"id": "team-1", "name": "Platform"},
			readmodels.ArtifactNameDTO{ArtifactType: "internal_team", ArtifactID: "team-1", Name: "Platform"}},
		{"internal team updated", archPL.InternalTeamUpdated, map[string]interface{}{"id": "team-1", "name": "Platform Engineering"},
			readmodels.ArtifactNameDTO{ArtifactType: "internal_team", ArtifactID: "team-1", Name: "Platform Engineering"}},
		{"view created", viewsPL.ViewCreated, map[string]interface{}{"id": "view-1", "name": "Integration Map"},
			readmodels.ArtifactNameDTO{ArtifactType: "view", ArtifactID: "view-1", Name: "Integration Map"}},
		{"view renamed", viewsPL.ViewRenamed, map[string]interface{}{"viewId": "view-1", "oldName": "Integration Map", "newName": "Integrations"},
			readmodels.ArtifactNameDTO{ArtifactType: "view", ArtifactID: "view-1", Name: "Integrations"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := projectNameCacheEvent(t, supplierEvent{aggregateID: tt.expected.ArtifactID, eventType: tt.eventType, data: tt.payload})

			require.Len(t, store.upserted, 1)
			assert.Equal(t, tt.expected, store.upserted[0])
			assert.Empty(t, store.deleted)
		})
	}
}

func TestArtifactNameCacheProjector_ForgetsDeletedArtifacts(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		payload   map[string]interface{}
		expected  readmodels.ArtifactRef
	}{
		{"capability deleted", capPL.CapabilityDeleted, map[string]interface{}{"id": "cap-1"},
			readmodels.ArtifactRef{ArtifactType: "capability", ArtifactID: "cap-1"}},
		{"business domain deleted", capPL.BusinessDomainDeleted, map[string]interface{}{"id": "bd-1"},
			readmodels.ArtifactRef{ArtifactType: "domain", ArtifactID: "bd-1"}},
		{"component deleted", archPL.ApplicationComponentDeleted, map[string]interface{}{"id": "comp-1"},
			readmodels.ArtifactRef{ArtifactType: "component", ArtifactID: "comp-1"}},
		{"vendor deleted", archPL.VendorDeleted, map[string]interface{}{"id": "ven-1"},
			readmodels.ArtifactRef{ArtifactType: "vendor", ArtifactID: "ven-1"}},
		{"acquired entity deleted", archPL.AcquiredEntityDeleted, map[string]interface{}{"id": "ae-1"},
			readmodels.ArtifactRef{ArtifactType: "acquired_entity", ArtifactID: "ae-1"}},
		{"internal team deleted", archPL.InternalTeamDeleted, map[string]interface{}{"id": "team-1"},
			readmodels.ArtifactRef{ArtifactType: "internal_team", ArtifactID: "team-1"}},
		{"view deleted", viewsPL.ViewDeleted, map[string]interface{}{"id": "view-1"},
			readmodels.ArtifactRef{ArtifactType: "view", ArtifactID: "view-1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := projectNameCacheEvent(t, supplierEvent{aggregateID: tt.expected.ArtifactID, eventType: tt.eventType, data: tt.payload})

			require.Len(t, store.deleted, 1)
			assert.Equal(t, tt.expected, store.deleted[0])
			assert.Empty(t, store.upserted)
		})
	}
}

func TestArtifactNameCacheProjector_FallsBackToAggregateID(t *testing.T) {
	store := projectNameCacheEvent(t, supplierEvent{
		aggregateID: "cap-9",
		eventType:   capPL.CapabilityCreated,
		data:        map[string]interface{}{"name": "Fallback"},
	})

	require.Len(t, store.upserted, 1)
	assert.Equal(t, "cap-9", store.upserted[0].ArtifactID)
}

func TestArtifactNameCacheProjector_IgnoresEventsWithNoNameToCache(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		payload   map[string]interface{}
	}{
		{"no artifact type is bound to the event", capPL.CapabilityParentChanged, map[string]interface{}{"id": "cap-1", "name": "Ignored"}},
		{"the event carries no name", capPL.CapabilityCreated, map[string]interface{}{"id": "cap-1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := projectNameCacheEvent(t, supplierEvent{aggregateID: "cap-1", eventType: tt.eventType, data: tt.payload})

			assert.Empty(t, store.upserted)
			assert.Empty(t, store.deleted)
		})
	}
}

func TestArtifactNameCacheProjector_StoreFailure_ReturnsError(t *testing.T) {
	store := &recordingArtifactNameStore{err: errors.New("database error")}
	projector := NewArtifactNameCacheProjector(store)

	err := projector.Handle(context.Background(), supplierEvent{
		aggregateID: "cap-1",
		eventType:   capPL.CapabilityCreated,
		data:        map[string]interface{}{"id": "cap-1", "name": "Onboarding"},
	})

	assert.ErrorContains(t, err, "database error")
}

func TestArtifactNameCacheProjector_SubscribedEventTypes_CoverAllArtifactTypes(t *testing.T) {
	subscribed := NewArtifactNameCacheProjector(&recordingArtifactNameStore{}).SubscribedEventTypes()

	assert.ElementsMatch(t, []string{
		capPL.CapabilityCreated, capPL.CapabilityUpdated, capPL.CapabilityDeleted,
		capPL.BusinessDomainCreated, capPL.BusinessDomainUpdated, capPL.BusinessDomainDeleted,
		archPL.ApplicationComponentCreated, archPL.ApplicationComponentUpdated, archPL.ApplicationComponentDeleted,
		archPL.VendorCreated, archPL.VendorUpdated, archPL.VendorDeleted,
		archPL.AcquiredEntityCreated, archPL.AcquiredEntityUpdated, archPL.AcquiredEntityDeleted,
		archPL.InternalTeamCreated, archPL.InternalTeamUpdated, archPL.InternalTeamDeleted,
		viewsPL.ViewCreated, viewsPL.ViewRenamed, viewsPL.ViewDeleted,
	}, subscribed)
}
