package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"easi/backend/internal/accessdelegation/application/readmodels"
	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	viewsPL "easi/backend/internal/architectureviews/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

const (
	CapabilityArtifact     = "capability"
	ComponentArtifact      = "component"
	ViewArtifact           = "view"
	DomainArtifact         = "domain"
	VendorArtifact         = "vendor"
	AcquiredEntityArtifact = "acquired_entity"
	InternalTeamArtifact   = "internal_team"
)

type ArtifactNameStore interface {
	Upsert(ctx context.Context, dto readmodels.ArtifactNameDTO) error
	Delete(ctx context.Context, artifactType, artifactID string) error
}

type artifactNameBinding struct {
	artifactType string
	forgotten    bool
}

var artifactNameBindings = map[string]artifactNameBinding{
	capPL.CapabilityCreated:            {artifactType: CapabilityArtifact},
	capPL.CapabilityUpdated:            {artifactType: CapabilityArtifact},
	capPL.CapabilityDeleted:            {artifactType: CapabilityArtifact, forgotten: true},
	capPL.BusinessDomainCreated:        {artifactType: DomainArtifact},
	capPL.BusinessDomainUpdated:        {artifactType: DomainArtifact},
	capPL.BusinessDomainDeleted:        {artifactType: DomainArtifact, forgotten: true},
	archPL.ApplicationComponentCreated: {artifactType: ComponentArtifact},
	archPL.ApplicationComponentUpdated: {artifactType: ComponentArtifact},
	archPL.ApplicationComponentDeleted: {artifactType: ComponentArtifact, forgotten: true},
	archPL.VendorCreated:               {artifactType: VendorArtifact},
	archPL.VendorUpdated:               {artifactType: VendorArtifact},
	archPL.VendorDeleted:               {artifactType: VendorArtifact, forgotten: true},
	archPL.AcquiredEntityCreated:       {artifactType: AcquiredEntityArtifact},
	archPL.AcquiredEntityUpdated:       {artifactType: AcquiredEntityArtifact},
	archPL.AcquiredEntityDeleted:       {artifactType: AcquiredEntityArtifact, forgotten: true},
	archPL.InternalTeamCreated:         {artifactType: InternalTeamArtifact},
	archPL.InternalTeamUpdated:         {artifactType: InternalTeamArtifact},
	archPL.InternalTeamDeleted:         {artifactType: InternalTeamArtifact, forgotten: true},
	viewsPL.ViewCreated:                {artifactType: ViewArtifact},
	viewsPL.ViewRenamed:                {artifactType: ViewArtifact},
	viewsPL.ViewDeleted:                {artifactType: ViewArtifact, forgotten: true},
}

type ArtifactNameCacheProjector struct {
	cache ArtifactNameStore
}

func NewArtifactNameCacheProjector(cache ArtifactNameStore) *ArtifactNameCacheProjector {
	return &ArtifactNameCacheProjector{cache: cache}
}

func (p *ArtifactNameCacheProjector) SubscribedEventTypes() []string {
	eventTypes := make([]string, 0, len(artifactNameBindings))
	for eventType := range artifactNameBindings {
		eventTypes = append(eventTypes, eventType)
	}
	sort.Strings(eventTypes)
	return eventTypes
}

func (p *ArtifactNameCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	binding, subscribed := artifactNameBindings[event.EventType()]
	if !subscribed {
		return nil
	}

	payload, err := decodeArtifactNamePayload(event)
	if err != nil {
		return err
	}

	artifactID := payload.artifactID(event.AggregateID())
	if artifactID == "" {
		return nil
	}
	if binding.forgotten {
		return p.cache.Delete(ctx, binding.artifactType, artifactID)
	}

	name := payload.artifactName()
	if name == "" {
		return nil
	}
	return p.cache.Upsert(ctx, readmodels.ArtifactNameDTO{
		ArtifactType: binding.artifactType,
		ArtifactID:   artifactID,
		Name:         name,
	})
}

type artifactNamePayload struct {
	ID      string `json:"id"`
	ViewID  string `json:"viewId"`
	Name    string `json:"name"`
	NewName string `json:"newName"`
}

func decodeArtifactNamePayload(event domain.DomainEvent) (artifactNamePayload, error) {
	var payload artifactNamePayload
	data, err := json.Marshal(event.EventData())
	if err != nil {
		return payload, fmt.Errorf("marshal %s payload: %w", event.EventType(), err)
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return payload, fmt.Errorf("unmarshal %s payload: %w", event.EventType(), err)
	}
	return payload, nil
}

func (p artifactNamePayload) artifactID(aggregateID string) string {
	return firstNonEmpty(p.ID, p.ViewID, aggregateID)
}

func (p artifactNamePayload) artifactName() string {
	return firstNonEmpty(p.Name, p.NewName)
}

func firstNonEmpty(candidates ...string) string {
	for _, candidate := range candidates {
		if candidate != "" {
			return candidate
		}
	}
	return ""
}
