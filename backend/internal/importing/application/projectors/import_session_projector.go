package projectors

import (
	"context"
	"encoding/json"
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
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func unmarshalEventData[T any](eventData []byte, eventName string) (*T, error) {
	var data T
	if err := json.Unmarshal(eventData, &data); err != nil {
		log.Printf("Failed to unmarshal %s: %v", eventName, err)
		return nil, err
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

func (p *ImportSessionProjector) handleImportSessionCreated(ctx context.Context, eventData []byte) error {
	var data struct {
		ID                string                 `json:"id"`
		SourceFormat      string                 `json:"sourceFormat"`
		BusinessDomainID  string                 `json:"businessDomainId"`
		CapabilityEAOwner string                 `json:"capabilityEAOwner"`
		Preview           map[string]interface{} `json:"preview"`
		CreatedAt         time.Time              `json:"createdAt"`
	}
	if err := json.Unmarshal(eventData, &data); err != nil {
		log.Printf("Failed to unmarshal ImportSessionCreated: %v", err)
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

	return p.readModel.Insert(ctx, dto)
}

func (p *ImportSessionProjector) handleImportStarted(ctx context.Context, eventData []byte) error {
	var data struct {
		ID         string `json:"id"`
		TotalItems int    `json:"totalItems"`
	}
	if err := json.Unmarshal(eventData, &data); err != nil {
		log.Printf("Failed to unmarshal ImportStarted: %v", err)
		return err
	}

	if err := p.readModel.UpdateStatus(ctx, data.ID, "importing"); err != nil {
		return err
	}

	progress := readmodels.ProgressDTO{
		Phase:          "creating_components",
		TotalItems:     data.TotalItems,
		CompletedItems: 0,
	}

	return p.readModel.UpdateProgress(ctx, data.ID, progress)
}

func (p *ImportSessionProjector) handleImportProgressUpdated(ctx context.Context, eventData []byte) error {
	var data struct {
		ID             string `json:"id"`
		Phase          string `json:"phase"`
		TotalItems     int    `json:"totalItems"`
		CompletedItems int    `json:"completedItems"`
	}
	if err := json.Unmarshal(eventData, &data); err != nil {
		log.Printf("Failed to unmarshal ImportProgressUpdated: %v", err)
		return err
	}

	progress := readmodels.ProgressDTO{
		Phase:          data.Phase,
		TotalItems:     data.TotalItems,
		CompletedItems: data.CompletedItems,
	}

	return p.readModel.UpdateProgress(ctx, data.ID, progress)
}

func (p *ImportSessionProjector) handleImportCompleted(ctx context.Context, eventData []byte) error {
	var data struct {
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
	if err := json.Unmarshal(eventData, &data); err != nil {
		log.Printf("Failed to unmarshal ImportCompleted: %v", err)
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

	return p.readModel.MarkCompleted(ctx, data.ID, result, data.CompletedAt)
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
