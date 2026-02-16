package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"easi/backend/internal/importing/application/readmodels"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ImportSessionProjector struct {
	readModel *readmodels.ImportSessionReadModel
}

func NewImportSessionProjector(readModel *readmodels.ImportSessionReadModel) *ImportSessionProjector {
	return &ImportSessionProjector{
		readModel: readModel,
	}
}

func (p *ImportSessionProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal import event %s for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func unmarshalEventData[T any](eventData []byte, eventName string) (*T, error) {
	var data T
	if err := json.Unmarshal(eventData, &data); err != nil {
		wrappedErr := fmt.Errorf("unmarshal %s event data: %w", eventName, err)
		log.Printf("failed to unmarshal %s: %v", eventName, wrappedErr)
		return nil, wrappedErr
	}
	return &data, nil
}

func (p *ImportSessionProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "ImportSessionCreated":
		return p.handleImportSessionCreated(ctx, eventData)
	case "ImportStarted":
		return p.handleImportStarted(ctx, eventData)
	case "ImportProgressUpdated":
		return p.handleImportProgressUpdated(ctx, eventData)
	case "ImportCompleted":
		return p.handleImportCompleted(ctx, eventData)
	case "ImportFailed":
		return p.handleImportFailed(ctx, eventData)
	case "ImportSessionCancelled":
		return p.handleImportSessionCancelled(ctx, eventData)
	}
	return nil
}

type importSessionCreatedData struct {
	ID                string                 `json:"id"`
	SourceFormat      string                 `json:"sourceFormat"`
	BusinessDomainID  string                 `json:"businessDomainId"`
	CapabilityEAOwner string                 `json:"capabilityEAOwner"`
	Preview           map[string]interface{} `json:"preview"`
	CreatedAt         time.Time              `json:"createdAt"`
}

func (p *ImportSessionProjector) handleImportSessionCreated(ctx context.Context, eventData []byte) error {
	data, err := unmarshalEventData[importSessionCreatedData](eventData, "ImportSessionCreated")
	if err != nil {
		return err
	}

	preview := readmodels.PreviewDTO{}
	if supported, ok := data.Preview["supported"].(map[string]interface{}); ok {
		preview.Supported = readmodels.SupportedCountsDTO{
			Capabilities:                    getIntFromMap(supported, "capabilities"),
			Components:                      getIntFromMap(supported, "components"),
			ValueStreams:                    getIntFromMap(supported, "valueStreams"),
			ParentChildRelationships:        getIntFromMap(supported, "parentChildRelationships"),
			Realizations:                    getIntFromMap(supported, "realizations"),
			ComponentRelationships:          getIntFromMap(supported, "componentRelationships"),
			CapabilityToValueStreamMappings: getIntFromMap(supported, "capabilityToValueStreamMappings"),
		}
	}
	if unsupported, ok := data.Preview["unsupported"].(map[string]interface{}); ok {
		preview.Unsupported = readmodels.UnsupportedCountsDTO{
			Elements:      getStringIntMap(unsupported, "elements"),
			Relationships: getStringIntMap(unsupported, "relationships"),
		}
	}

	dto := readmodels.ImportSessionDTO{
		ID:                data.ID,
		SourceFormat:      data.SourceFormat,
		BusinessDomainID:  data.BusinessDomainID,
		CapabilityEAOwner: data.CapabilityEAOwner,
		Status:            "pending",
		Preview:           &preview,
		CreatedAt:         data.CreatedAt,
	}

	if err := p.readModel.Insert(ctx, dto); err != nil {
		return fmt.Errorf("project ImportSessionCreated for session %s: %w", data.ID, err)
	}
	return nil
}

type importStartedData struct {
	ID         string `json:"id"`
	TotalItems int    `json:"totalItems"`
}

func (p *ImportSessionProjector) handleImportStarted(ctx context.Context, eventData []byte) error {
	data, err := unmarshalEventData[importStartedData](eventData, "ImportStarted")
	if err != nil {
		return err
	}

	if err := p.readModel.UpdateStatus(ctx, data.ID, "importing"); err != nil {
		return fmt.Errorf("project ImportStarted status update for session %s: %w", data.ID, err)
	}

	progress := readmodels.ProgressDTO{
		Phase:          "creating_components",
		TotalItems:     data.TotalItems,
		CompletedItems: 0,
	}

	if err := p.readModel.UpdateProgress(ctx, data.ID, progress); err != nil {
		return fmt.Errorf("project ImportStarted progress update for session %s: %w", data.ID, err)
	}
	return nil
}

type importProgressUpdatedData struct {
	ID             string `json:"id"`
	Phase          string `json:"phase"`
	TotalItems     int    `json:"totalItems"`
	CompletedItems int    `json:"completedItems"`
}

func (p *ImportSessionProjector) handleImportProgressUpdated(ctx context.Context, eventData []byte) error {
	data, err := unmarshalEventData[importProgressUpdatedData](eventData, "ImportProgressUpdated")
	if err != nil {
		return err
	}

	progress := readmodels.ProgressDTO{
		Phase:          data.Phase,
		TotalItems:     data.TotalItems,
		CompletedItems: data.CompletedItems,
	}

	if err := p.readModel.UpdateProgress(ctx, data.ID, progress); err != nil {
		return fmt.Errorf("project ImportProgressUpdated for session %s: %w", data.ID, err)
	}
	return nil
}

type importCompletedData struct {
	ID                        string                   `json:"id"`
	CapabilitiesCreated       int                      `json:"capabilitiesCreated"`
	ComponentsCreated         int                      `json:"componentsCreated"`
	ValueStreamsCreated       int                      `json:"valueStreamsCreated"`
	RealizationsCreated       int                      `json:"realizationsCreated"`
	ComponentRelationsCreated int                      `json:"componentRelationsCreated"`
	CapabilityMappings        int                      `json:"capabilityMappings"`
	DomainAssignments         int                      `json:"domainAssignments"`
	Errors                    []map[string]interface{} `json:"errors"`
	CompletedAt               time.Time                `json:"completedAt"`
}

func (p *ImportSessionProjector) handleImportCompleted(ctx context.Context, eventData []byte) error {
	data, err := unmarshalEventData[importCompletedData](eventData, "ImportCompleted")
	if err != nil {
		return err
	}

	errors := make([]readmodels.ImportErrorDTO, 0, len(data.Errors))
	for _, e := range data.Errors {
		errors = append(errors, readmodels.ImportErrorDTO{
			SourceElement: getString(e, "sourceElement"),
			SourceName:    getString(e, "sourceName"),
			Error:         getString(e, "error"),
			Action:        getString(e, "action"),
		})
	}

	result := readmodels.ResultDTO{
		CapabilitiesCreated:       data.CapabilitiesCreated,
		ComponentsCreated:         data.ComponentsCreated,
		ValueStreamsCreated:       data.ValueStreamsCreated,
		RealizationsCreated:       data.RealizationsCreated,
		ComponentRelationsCreated: data.ComponentRelationsCreated,
		CapabilityMappings:        data.CapabilityMappings,
		DomainAssignments:         data.DomainAssignments,
		Errors:                    errors,
	}

	if err := p.readModel.MarkCompleted(ctx, data.ID, result, data.CompletedAt); err != nil {
		return fmt.Errorf("project ImportCompleted for session %s: %w", data.ID, err)
	}
	return nil
}

type importFailedData struct {
	ID       string    `json:"id"`
	FailedAt time.Time `json:"failedAt"`
}

type importCancelledData struct {
	ID string `json:"id"`
}

func (p *ImportSessionProjector) handleImportFailed(ctx context.Context, eventData []byte) error {
	data, err := unmarshalEventData[importFailedData](eventData, "ImportFailed")
	if err != nil {
		return err
	}
	return p.readModel.MarkFailed(ctx, data.ID, data.FailedAt)
}

func (p *ImportSessionProjector) handleImportSessionCancelled(ctx context.Context, eventData []byte) error {
	data, err := unmarshalEventData[importCancelledData](eventData, "ImportSessionCancelled")
	if err != nil {
		return err
	}
	return p.readModel.MarkCancelled(ctx, data.ID)
}

func getIntFromMap(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}

func getStringIntMap(m map[string]interface{}, key string) map[string]int {
	result := make(map[string]int)
	if nested, ok := m[key].(map[string]interface{}); ok {
		for k, v := range nested {
			if count, ok := v.(float64); ok {
				result[k] = int(count)
			}
			if count, ok := v.(int); ok {
				result[k] = count
			}
		}
	}
	return result
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
