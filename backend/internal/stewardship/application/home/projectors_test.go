package home

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "easi/backend/internal/shared/eventsourcing"
)

type recordingCaches struct {
	calls []string
}

func (r *recordingCaches) record(format string, args ...any) error {
	r.calls = append(r.calls, fmt.Sprintf(format, args...))
	return nil
}

func (r *recordingCaches) SaveCapability(_ context.Context, c CachedCapability) error {
	return r.record("SaveCapability %s %s %s %s", c.ID, c.Name, c.Level, c.ParentID)
}
func (r *recordingCaches) RenameCapability(_ context.Context, id CapabilityID, name string) error {
	return r.record("RenameCapability %s %s", id, name)
}
func (r *recordingCaches) SetCapabilityMetadata(_ context.Context, m CachedCapabilityMetadata) error {
	return r.record("SetCapabilityMetadata %s %s %s", m.CapabilityID, m.Status, m.EAOwner)
}
func (r *recordingCaches) MoveCapability(_ context.Context, p CachedPlacement) error {
	return r.record("MoveCapability %s %s %s", p.CapabilityID, p.ParentID, p.Level)
}
func (r *recordingCaches) SetCapabilityLevel(_ context.Context, id CapabilityID, level string) error {
	return r.record("SetCapabilityLevel %s %s", id, level)
}
func (r *recordingCaches) DeleteCapability(_ context.Context, id CapabilityID) error {
	return r.record("DeleteCapability %s", id)
}
func (r *recordingCaches) AssignToDomain(_ context.Context, a CachedAssignment) error {
	return r.record("AssignToDomain %s %s", a.CapabilityID, a.DomainID)
}
func (r *recordingCaches) UnassignFromDomain(_ context.Context, a CachedAssignment) error {
	return r.record("UnassignFromDomain %s %s", a.CapabilityID, a.DomainID)
}
func (r *recordingCaches) DeleteAssignmentsOfCapability(_ context.Context, id CapabilityID) error {
	return r.record("DeleteAssignmentsOfCapability %s", id)
}
func (r *recordingCaches) DeleteAssignmentsToDomain(_ context.Context, id DomainID) error {
	return r.record("DeleteAssignmentsToDomain %s", id)
}
func (r *recordingCaches) SaveApplication(_ context.Context, id, name string) error {
	return r.record("SaveApplication %s %s", id, name)
}
func (r *recordingCaches) SetOwnership(_ context.Context, o CachedOwnership) error {
	return r.record("SetOwnership %s %s %s %s", o.ComponentID, o.State, o.OwnerKind, o.OwnerID)
}
func (r *recordingCaches) DeleteApplication(_ context.Context, id string) error {
	return r.record("DeleteApplication %s", id)
}
func (r *recordingCaches) SaveRealization(_ context.Context, c CachedRealization) error {
	return r.record("SaveRealization %s %s %s", c.ID, c.CapabilityID, c.ComponentID)
}
func (r *recordingCaches) DeleteRealization(_ context.Context, id string) error {
	return r.record("DeleteRealization %s", id)
}
func (r *recordingCaches) DeleteRealizationsOfCapability(_ context.Context, id string) error {
	return r.record("DeleteRealizationsOfCapability %s", id)
}
func (r *recordingCaches) DeleteRealizationsOfComponent(_ context.Context, id string) error {
	return r.record("DeleteRealizationsOfComponent %s", id)
}
func (r *recordingCaches) SaveTimeAssessment(_ context.Context, a CachedTimeAssessment) error {
	return r.record("SaveTimeAssessment %s %s %s %s", a.CapabilityID, a.ComponentID, a.Grade, a.AssessedAt.Format(time.RFC3339))
}
func (r *recordingCaches) DeleteTimeAssessment(_ context.Context, capabilityID, componentID string) error {
	return r.record("DeleteTimeAssessment %s %s", capabilityID, componentID)
}
func (r *recordingCaches) SaveEditGrant(_ context.Context, g CachedEditGrant) error {
	return r.record("SaveEditGrant %s %s %s %s %s", g.ID, g.ArtifactType, g.ArtifactID, g.GranteeEmail, g.ExpiresAt.Format(time.RFC3339))
}
func (r *recordingCaches) DeleteEditGrant(_ context.Context, id string) error {
	return r.record("DeleteEditGrant %s", id)
}

type eventHandler interface {
	Handle(ctx context.Context, event domain.DomainEvent) error
}

func storedEvent(t *testing.T, eventType string, data map[string]any) domain.DomainEvent {
	t.Helper()
	payload, err := json.Marshal(data)
	require.NoError(t, err)
	return domain.NewGenericDomainEvent("aggregate", eventType, payload, time.Now().UTC())
}

type projectorCase struct {
	eventType string
	data      map[string]any
	want      string
}

func assertProjects(t *testing.T, newProjector func(*recordingCaches) eventHandler, cases []projectorCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.eventType, func(t *testing.T) {
			caches := &recordingCaches{}

			require.NoError(t, newProjector(caches).Handle(context.Background(), storedEvent(t, tc.eventType, tc.data)))

			assert.Equal(t, []string{tc.want}, caches.calls)
		})
	}
}

func TestCapabilityCacheProjector(t *testing.T) {
	assertProjects(t, func(c *recordingCaches) eventHandler { return NewCapabilityCacheProjector(c) }, []projectorCase{
		{"CapabilityCreated", map[string]any{"id": "c1", "name": "Invoicing", "level": "L2", "parentId": "c0"}, "SaveCapability c1 Invoicing L2 c0"},
		{"CapabilityUpdated", map[string]any{"id": "c1", "name": "Billing"}, "RenameCapability c1 Billing"},
		{"CapabilityMetadataUpdated", map[string]any{"id": "c1", "status": "Deprecated", "eaOwner": "u1"}, "SetCapabilityMetadata c1 Deprecated u1"},
		{"CapabilityParentChanged", map[string]any{"capabilityId": "c1", "newParentId": "c9", "newLevel": "L3"}, "MoveCapability c1 c9 L3"},
		{"CapabilityLevelChanged", map[string]any{"capabilityId": "c1", "newLevel": "L3"}, "SetCapabilityLevel c1 L3"},
		{"CapabilityDeleted", map[string]any{"id": "c1"}, "DeleteCapability c1"},
	})
}

func TestDomainAssignmentCacheProjector(t *testing.T) {
	assertProjects(t, func(c *recordingCaches) eventHandler { return NewDomainAssignmentCacheProjector(c) }, []projectorCase{
		{"CapabilityAssignedToDomain", map[string]any{"capabilityId": "c1", "businessDomainId": "d1"}, "AssignToDomain c1 d1"},
		{"CapabilityUnassignedFromDomain", map[string]any{"capabilityId": "c1", "businessDomainId": "d1"}, "UnassignFromDomain c1 d1"},
		{"CapabilityDeleted", map[string]any{"id": "c1"}, "DeleteAssignmentsOfCapability c1"},
		{"BusinessDomainDeleted", map[string]any{"id": "d1"}, "DeleteAssignmentsToDomain d1"},
	})
}

func TestApplicationCacheProjector(t *testing.T) {
	assertProjects(t, func(c *recordingCaches) eventHandler { return NewApplicationCacheProjector(c) }, []projectorCase{
		{"ApplicationComponentCreated", map[string]any{"id": "a1", "name": "CRM"}, "SaveApplication a1 CRM"},
		{"ApplicationComponentUpdated", map[string]any{"id": "a1", "name": "CRM 2"}, "SaveApplication a1 CRM 2"},
		{"ApplicationComponentDeleted", map[string]any{"id": "a1"}, "DeleteApplication a1"},
		{"ApplicationOwnerNominated", map[string]any{"componentId": "a1", "ownerKind": "user", "ownerId": "u1"}, "SetOwnership a1 nominated user u1"},
		{"ApplicationOwnershipConfirmed", map[string]any{"componentId": "a1", "ownerKind": "user", "ownerId": "u1", "ownershipState": "owned"}, "SetOwnership a1 owned user u1"},
		{"ApplicationOwnerAssigned", map[string]any{"componentId": "a1", "ownerKind": "team", "ownerId": "t1", "ownershipState": "managed"}, "SetOwnership a1 managed team t1"},
		{"ApplicationOwnershipCleared", map[string]any{"componentId": "a1"}, "SetOwnership a1 unknown  "},
	})
}

func TestRealizationCacheProjector(t *testing.T) {
	assertProjects(t, func(c *recordingCaches) eventHandler { return NewRealizationCacheProjector(c) }, []projectorCase{
		{"SystemLinkedToCapability", map[string]any{"id": "r1", "capabilityId": "c1", "componentId": "a1"}, "SaveRealization r1 c1 a1"},
		{"SystemRealizationDeleted", map[string]any{"id": "r1"}, "DeleteRealization r1"},
		{"CapabilityDeleted", map[string]any{"id": "c1"}, "DeleteRealizationsOfCapability c1"},
		{"ApplicationComponentDeleted", map[string]any{"id": "a1"}, "DeleteRealizationsOfComponent a1"},
	})
}

func TestTimeAssessmentCacheProjector(t *testing.T) {
	assertProjects(t, func(c *recordingCaches) eventHandler { return NewTimeAssessmentCacheProjector(c) }, []projectorCase{
		{"TimeAssessmentRecorded", map[string]any{"id": "t1", "capabilityId": "c1", "componentId": "a1", "grade": "Invest", "occurredOn": "2026-09-01T10:00:00Z"}, "SaveTimeAssessment c1 a1 Invest 2026-09-01T10:00:00Z"},
		{"TimeAssessmentRemoved", map[string]any{"id": "t1", "capabilityId": "c1", "componentId": "a1"}, "DeleteTimeAssessment c1 a1"},
	})
}

func TestEditGrantCacheProjector(t *testing.T) {
	assertProjects(t, func(c *recordingCaches) eventHandler { return NewEditGrantCacheProjector(c) }, []projectorCase{
		{"EditGrantActivated", map[string]any{"id": "g1", "artifactType": "component", "artifactId": "a1", "granteeEmail": "ole@example.com", "grantorEmail": "x@example.com", "reason": "r", "expiresAt": "2026-10-21T10:00:00Z"}, "SaveEditGrant g1 component a1 ole@example.com 2026-10-21T10:00:00Z"},
		{"EditGrantRevoked", map[string]any{"id": "g1", "revokedBy": "x"}, "DeleteEditGrant g1"},
		{"EditGrantExpired", map[string]any{"id": "g1"}, "DeleteEditGrant g1"},
	})
}

func TestProjectors_IgnoreUntrackedEvents(t *testing.T) {
	caches := &recordingCaches{}

	require.NoError(t, NewEditGrantCacheProjector(caches).Handle(context.Background(), storedEvent(t, "SomethingElse", map[string]any{})))

	assert.Empty(t, caches.calls)
}
