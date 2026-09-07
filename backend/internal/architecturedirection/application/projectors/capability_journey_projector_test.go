package projectors

import (
	"context"
	"testing"
	"time"

	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCapabilityJourneyStore struct {
	inserted        []readmodels.InsertJourneyParams
	statusUpdates   []capabilityJourneyStatusUpdate
	progress        []int
	detailsUpdates  []capabilityJourneyDetailsUpdate
	sourceReplaces  []capabilityJourneySourceReplace
	milestoneAdds   []capabilityJourneyMilestoneUpsert
	milestoneEdits  []capabilityJourneyMilestoneUpsert
	milestoneRemove []string
	milestoneOrders []capabilityJourneyMilestoneOrder
}

type capabilityJourneyStatusUpdate struct {
	journeyID  string
	status     string
	column     readmodels.JourneyTimestampColumn
	occurredAt time.Time
}

type capabilityJourneyDetailsUpdate struct {
	journeyID     string
	note          string
	targetYear    *int
	targetQuarter *int
	resultingName string
}

type capabilityJourneySourceReplace struct {
	journeyID    string
	componentIDs []string
}

type capabilityJourneyMilestoneUpsert struct {
	journeyID     string
	milestoneID   string
	label         string
	targetYear    *int
	targetQuarter *int
	status        string
}

func (m *mockCapabilityJourneyStore) InsertJourney(_ context.Context, p readmodels.InsertJourneyParams) error {
	m.inserted = append(m.inserted, p)
	return nil
}

func (m *mockCapabilityJourneyStore) UpdateStatus(_ context.Context, p readmodels.UpdateJourneyStatusParams) error {
	m.statusUpdates = append(m.statusUpdates, capabilityJourneyStatusUpdate{p.JourneyID, p.Status, p.Column, p.OccurredAt})
	return nil
}

func (m *mockCapabilityJourneyStore) UpdateProgress(_ context.Context, _ string, progress int) error {
	m.progress = append(m.progress, progress)
	return nil
}

func (m *mockCapabilityJourneyStore) UpdateDetails(_ context.Context, p readmodels.UpdateJourneyDetailsParams) error {
	m.detailsUpdates = append(m.detailsUpdates, capabilityJourneyDetailsUpdate{p.JourneyID, p.Note, p.TargetYear, p.TargetQuarter, p.ResultingName})
	return nil
}

func (m *mockCapabilityJourneyStore) ReplaceSources(_ context.Context, journeyID string, componentIDs []string) error {
	m.sourceReplaces = append(m.sourceReplaces, capabilityJourneySourceReplace{journeyID, componentIDs})
	return nil
}

func (m *mockCapabilityJourneyStore) AddMilestone(_ context.Context, p readmodels.JourneyMilestoneUpsertParams) error {
	m.milestoneAdds = append(m.milestoneAdds, capabilityJourneyMilestoneUpsert{p.JourneyID, p.MilestoneID, p.Label, p.TargetYear, p.TargetQuarter, p.Status})
	return nil
}

func (m *mockCapabilityJourneyStore) UpdateMilestone(_ context.Context, p readmodels.JourneyMilestoneUpsertParams) error {
	m.milestoneEdits = append(m.milestoneEdits, capabilityJourneyMilestoneUpsert{p.JourneyID, p.MilestoneID, p.Label, p.TargetYear, p.TargetQuarter, p.Status})
	return nil
}

type capabilityJourneyMilestoneOrder struct {
	journeyID    string
	milestoneIDs []string
}

func (m *mockCapabilityJourneyStore) ReorderMilestones(_ context.Context, journeyID string, milestoneIDs []string) error {
	m.milestoneOrders = append(m.milestoneOrders, capabilityJourneyMilestoneOrder{journeyID, milestoneIDs})
	return nil
}

func (m *mockCapabilityJourneyStore) RemoveMilestone(_ context.Context, _, milestoneID string) error {
	m.milestoneRemove = append(m.milestoneRemove, milestoneID)
	return nil
}

func TestCapabilityJourneyProjector_JourneyPlanned_InsertsJourney(t *testing.T) {
	store := &mockCapabilityJourneyStore{}
	projector := NewCapabilityJourneyProjector(store)

	id, capID, toApp, fromApp := uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String()
	evt := events.NewJourneyPlanned(events.JourneyPlannedFields{
		ID:               id,
		CapabilityID:     capID,
		Kind:             "migration",
		FromComponentIDs: []string{fromApp},
		ToComponentID:    toApp,
		Note:             "moving on",
		TargetPeriod:     &events.TargetPeriodData{Year: 2027, Quarter: 2},
		PlannedBy:        "architect@example.com",
	})

	require.NoError(t, projector.Handle(context.Background(), evt))

	require.Len(t, store.inserted, 1)
	assert.Equal(t, id, store.inserted[0].ID)
	assert.Equal(t, capID, store.inserted[0].CapabilityID)
	assert.Equal(t, "migration", store.inserted[0].Kind)
	assert.Equal(t, []string{fromApp}, store.inserted[0].FromComponentIDs)
	assert.Equal(t, toApp, store.inserted[0].ToComponentID)
	require.NotNil(t, store.inserted[0].TargetYear)
	assert.Equal(t, 2027, *store.inserted[0].TargetYear)
}

func TestCapabilityJourneyProjector_MaturityJourneyPlanned_InsertsTargetMaturity(t *testing.T) {
	store := &mockCapabilityJourneyStore{}
	projector := NewCapabilityJourneyProjector(store)

	target := 65
	evt := events.NewJourneyPlanned(events.JourneyPlannedFields{
		ID:             uuid.New().String(),
		CapabilityID:   uuid.New().String(),
		Kind:           "maturity",
		TargetMaturity: &target,
		PlannedBy:      "architect@example.com",
	})

	require.NoError(t, projector.Handle(context.Background(), evt))

	require.Len(t, store.inserted, 1)
	assert.Equal(t, "maturity", store.inserted[0].Kind)
	assert.Empty(t, store.inserted[0].ToComponentID)
	require.NotNil(t, store.inserted[0].TargetMaturity)
	assert.Equal(t, 65, *store.inserted[0].TargetMaturity)
}

func TestCapabilityJourneyProjector_JourneyStarted_UpdatesStatus(t *testing.T) {
	store := &mockCapabilityJourneyStore{}
	projector := NewCapabilityJourneyProjector(store)

	evt := events.NewJourneyStarted(events.JourneyStartedFields{ID: uuid.New().String(), StartedBy: "a@example.com"})
	require.NoError(t, projector.Handle(context.Background(), evt))

	require.Len(t, store.statusUpdates, 1)
	assert.Equal(t, "in-flight", store.statusUpdates[0].status)
	assert.Equal(t, readmodels.JourneyTimestampStarted, store.statusUpdates[0].column)
}

func TestCapabilityJourneyProjector_JourneyCompleted_UpdatesStatus(t *testing.T) {
	store := &mockCapabilityJourneyStore{}
	projector := NewCapabilityJourneyProjector(store)

	evt := events.NewJourneyCompleted(events.JourneyCompletedFields{ID: uuid.New().String(), CompletedBy: "a@example.com"})
	require.NoError(t, projector.Handle(context.Background(), evt))

	require.Len(t, store.statusUpdates, 1)
	assert.Equal(t, "done", store.statusUpdates[0].status)
	assert.Equal(t, readmodels.JourneyTimestampCompleted, store.statusUpdates[0].column)
}

func TestCapabilityJourneyProjector_JourneyAbandoned_UpdatesStatus(t *testing.T) {
	store := &mockCapabilityJourneyStore{}
	projector := NewCapabilityJourneyProjector(store)

	evt := events.NewJourneyAbandoned(events.JourneyAbandonedFields{ID: uuid.New().String(), AbandonedBy: "a@example.com"})
	require.NoError(t, projector.Handle(context.Background(), evt))

	require.Len(t, store.statusUpdates, 1)
	assert.Equal(t, "abandoned", store.statusUpdates[0].status)
	assert.Equal(t, readmodels.JourneyTimestampAbandoned, store.statusUpdates[0].column)
}

func TestCapabilityJourneyProjector_JourneyProgressUpdated_UpdatesProgress(t *testing.T) {
	store := &mockCapabilityJourneyStore{}
	projector := NewCapabilityJourneyProjector(store)

	evt := events.NewJourneyProgressUpdated(events.JourneyProgressUpdatedFields{ID: uuid.New().String(), Progress: 60, UpdatedBy: "a@example.com"})
	require.NoError(t, projector.Handle(context.Background(), evt))

	require.Len(t, store.progress, 1)
	assert.Equal(t, 60, store.progress[0])
}

func TestCapabilityJourneyProjector_JourneyDetailsUpdated_UpdatesDetails(t *testing.T) {
	store := &mockCapabilityJourneyStore{}
	projector := NewCapabilityJourneyProjector(store)

	evt := events.NewJourneyDetailsUpdated(events.JourneyDetailsUpdatedFields{
		ID: uuid.New().String(), Note: "updated", TargetPeriod: &events.TargetPeriodData{Year: 2028, Quarter: 1}, UpdatedBy: "a@example.com",
	})
	require.NoError(t, projector.Handle(context.Background(), evt))

	require.Len(t, store.detailsUpdates, 1)
	assert.Equal(t, "updated", store.detailsUpdates[0].note)
	require.NotNil(t, store.detailsUpdates[0].targetYear)
	assert.Equal(t, 2028, *store.detailsUpdates[0].targetYear)
}

func TestCapabilityJourneyProjector_JourneySourceApplicationsChanged_ReplacesSources(t *testing.T) {
	store := &mockCapabilityJourneyStore{}
	projector := NewCapabilityJourneyProjector(store)

	newFrom := uuid.New().String()
	evt := events.NewJourneySourceApplicationsChanged(events.JourneySourceApplicationsChangedFields{
		ID: uuid.New().String(), FromComponentIDs: []string{newFrom}, ChangedBy: "a@example.com",
	})
	require.NoError(t, projector.Handle(context.Background(), evt))

	require.Len(t, store.sourceReplaces, 1)
	assert.Equal(t, []string{newFrom}, store.sourceReplaces[0].componentIDs)
}

func TestCapabilityJourneyProjector_MilestoneUpsertEvents_UpsertMilestone(t *testing.T) {
	cases := []struct {
		name       string
		buildEvent func(f events.JourneyMilestoneFields) domain.DomainEvent
		upserts    func(store *mockCapabilityJourneyStore) []capabilityJourneyMilestoneUpsert
		label      string
		status     string
	}{
		{
			name:       "milestone added",
			buildEvent: func(f events.JourneyMilestoneFields) domain.DomainEvent { return events.NewJourneyMilestoneAdded(f) },
			upserts:    func(store *mockCapabilityJourneyStore) []capabilityJourneyMilestoneUpsert { return store.milestoneAdds },
			label:      "first",
			status:     "planned",
		},
		{
			name:       "milestone updated",
			buildEvent: func(f events.JourneyMilestoneFields) domain.DomainEvent { return events.NewJourneyMilestoneUpdated(f) },
			upserts: func(store *mockCapabilityJourneyStore) []capabilityJourneyMilestoneUpsert {
				return store.milestoneEdits
			},
			label:  "first done",
			status: "done",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockCapabilityJourneyStore{}
			projector := NewCapabilityJourneyProjector(store)

			milestoneID := uuid.New().String()
			evt := tc.buildEvent(events.JourneyMilestoneFields{
				ID: uuid.New().String(), MilestoneID: milestoneID, Label: tc.label, Status: tc.status, Actor: "a@example.com",
			})
			require.NoError(t, projector.Handle(context.Background(), evt))

			upserts := tc.upserts(store)
			require.Len(t, upserts, 1)
			assert.Equal(t, milestoneID, upserts[0].milestoneID)
			assert.Equal(t, tc.label, upserts[0].label)
			assert.Equal(t, tc.status, upserts[0].status)
		})
	}
}

func TestCapabilityJourneyProjector_JourneyMilestoneRemoved_RemovesMilestone(t *testing.T) {
	store := &mockCapabilityJourneyStore{}
	projector := NewCapabilityJourneyProjector(store)

	milestoneID := uuid.New().String()
	evt := events.NewJourneyMilestoneRemoved(events.JourneyMilestoneRemovedFields{
		ID: uuid.New().String(), MilestoneID: milestoneID, RemovedBy: "a@example.com",
	})
	require.NoError(t, projector.Handle(context.Background(), evt))

	require.Equal(t, []string{milestoneID}, store.milestoneRemove)
}

func TestCapabilityJourneyProjector_UnknownEvent_Ignored(t *testing.T) {
	store := &mockCapabilityJourneyStore{}
	projector := NewCapabilityJourneyProjector(store)

	require.NoError(t, projector.ProjectEvent(context.Background(), "SomethingElseHappened", []byte(`{}`)))
	assert.Empty(t, store.inserted)
}

func TestCapabilityJourneyProjector_JourneyMilestonesReordered_RewritesOrder(t *testing.T) {
	store := &mockCapabilityJourneyStore{}
	projector := NewCapabilityJourneyProjector(store)

	journeyID := uuid.New().String()
	order := []string{uuid.New().String(), uuid.New().String()}
	evt := events.NewJourneyMilestonesReordered(events.JourneyMilestonesReorderedFields{
		ID: journeyID, MilestoneIDs: order, ReorderedBy: "a@example.com",
	})
	require.NoError(t, projector.Handle(context.Background(), evt))

	require.Equal(t, []capabilityJourneyMilestoneOrder{{journeyID, order}}, store.milestoneOrders)
}
