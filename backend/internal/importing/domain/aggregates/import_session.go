package aggregates

import (
	"errors"
	"time"

	"easi/backend/internal/importing/domain/events"
	"easi/backend/internal/importing/domain/valueobjects"
	"easi/backend/internal/shared/domain"
)

var (
	ErrImportAlreadyStarted       = errors.New("import has already been started")
	ErrImportNotStarted           = errors.New("import has not been started")
	ErrCannotCancelStartedImport  = errors.New("cannot cancel import that has already started")
	ErrImportAlreadyCompleted     = errors.New("import has already completed")
)

type ParsedElement struct {
	SourceID    string
	Name        string
	Description string
	ParentID    string
}

type ParsedRelationship struct {
	SourceID   string
	Type       string
	SourceRef  string
	TargetRef  string
	Name       string
	Documentation string
}

type ParsedData struct {
	Capabilities  []ParsedElement
	Components    []ParsedElement
	Relationships []ParsedRelationship
}

type ImportResult struct {
	CapabilitiesCreated       int
	ComponentsCreated         int
	RealizationsCreated       int
	ComponentRelationsCreated int
	DomainAssignments         int
	Errors                    []valueobjects.ImportError
}

type ImportSession struct {
	domain.AggregateRoot
	id               valueobjects.ImportSessionID
	sourceFormat     valueobjects.SourceFormat
	businessDomainID string
	status           valueobjects.ImportStatus
	preview          valueobjects.ImportPreview
	progress         valueobjects.ImportProgress
	parsedData       ParsedData
	result           ImportResult
	createdAt        time.Time
	completedAt      *time.Time
	isCancelled      bool
}

func NewImportSession(
	sourceFormat valueobjects.SourceFormat,
	businessDomainID string,
	preview valueobjects.ImportPreview,
	parsedData ParsedData,
) (*ImportSession, error) {
	session := &ImportSession{
		AggregateRoot: domain.NewAggregateRoot(),
		id:            valueobjects.NewImportSessionID(),
	}

	previewMap := map[string]interface{}{
		"supported": map[string]interface{}{
			"capabilities":             preview.Supported().Capabilities,
			"components":               preview.Supported().Components,
			"parentChildRelationships": preview.Supported().ParentChildRelationships,
			"realizations":             preview.Supported().Realizations,
		},
		"unsupported": map[string]interface{}{
			"elements":      preview.Unsupported().Elements,
			"relationships": preview.Unsupported().Relationships,
		},
	}

	parsedDataMap := map[string]interface{}{
		"capabilities":  serializeElements(parsedData.Capabilities),
		"components":    serializeElements(parsedData.Components),
		"relationships": serializeRelationships(parsedData.Relationships),
	}

	event := events.NewImportSessionCreated(
		session.id.Value(),
		sourceFormat.Value(),
		businessDomainID,
		previewMap,
		parsedDataMap,
	)

	session.apply(event)
	session.RaiseEvent(event)

	return session, nil
}

func (s *ImportSession) ID() string {
	return s.id.Value()
}

func (s *ImportSession) SourceFormat() valueobjects.SourceFormat {
	return s.sourceFormat
}

func (s *ImportSession) BusinessDomainID() string {
	return s.businessDomainID
}

func (s *ImportSession) Status() valueobjects.ImportStatus {
	return s.status
}

func (s *ImportSession) Preview() valueobjects.ImportPreview {
	return s.preview
}

func (s *ImportSession) Progress() valueobjects.ImportProgress {
	return s.progress
}

func (s *ImportSession) ParsedData() ParsedData {
	return s.parsedData
}

func (s *ImportSession) Result() ImportResult {
	return s.result
}

func (s *ImportSession) CreatedAt() time.Time {
	return s.createdAt
}

func (s *ImportSession) CompletedAt() *time.Time {
	return s.completedAt
}

func (s *ImportSession) IsCancelled() bool {
	return s.isCancelled
}

func (s *ImportSession) StartImport() error {
	if !s.status.IsPending() {
		return ErrImportAlreadyStarted
	}

	event := events.NewImportStarted(s.id.Value(), s.preview.TotalSupportedItems())
	s.apply(event)
	s.RaiseEvent(event)

	return nil
}

func (s *ImportSession) UpdateProgress(progress valueobjects.ImportProgress) error {
	if !s.status.IsImporting() {
		return ErrImportNotStarted
	}

	event := events.NewImportProgressUpdated(
		s.id.Value(),
		progress.Phase(),
		progress.TotalItems(),
		progress.CompletedItems(),
	)
	s.apply(event)
	s.RaiseEvent(event)

	return nil
}

func (s *ImportSession) Complete(result ImportResult) error {
	if !s.status.IsImporting() {
		return ErrImportNotStarted
	}

	var errorMaps []map[string]interface{}
	for _, e := range result.Errors {
		errorMaps = append(errorMaps, map[string]interface{}{
			"sourceElement": e.SourceElement(),
			"sourceName":    e.SourceName(),
			"error":         e.Error(),
			"action":        e.Action(),
		})
	}

	event := events.NewImportCompleted(
		s.id.Value(),
		result.CapabilitiesCreated,
		result.ComponentsCreated,
		result.RealizationsCreated,
		result.DomainAssignments,
		errorMaps,
	)
	s.apply(event)
	s.RaiseEvent(event)

	return nil
}

func (s *ImportSession) Fail(reason string) error {
	if !s.status.IsImporting() {
		return ErrImportNotStarted
	}

	event := events.NewImportFailed(s.id.Value(), reason)
	s.apply(event)
	s.RaiseEvent(event)

	return nil
}

func (s *ImportSession) Cancel() error {
	if !s.status.IsPending() {
		return ErrCannotCancelStartedImport
	}

	event := events.NewImportSessionCancelled(s.id.Value())
	s.apply(event)
	s.RaiseEvent(event)

	return nil
}

func (s *ImportSession) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.ImportSessionCreated:
		s.id, _ = valueobjects.NewImportSessionIDFromString(e.ID)
		s.sourceFormat, _ = valueobjects.NewSourceFormat(e.SourceFormat)
		s.businessDomainID = e.BusinessDomainID
		s.status = valueobjects.ImportStatusPending()
		s.preview = deserializePreview(e.Preview)
		s.parsedData = deserializeParsedData(e.ParsedData)
		s.createdAt = e.CreatedAt

	case events.ImportStarted:
		s.status = valueobjects.ImportStatusImporting()
		s.progress, _ = valueobjects.NewImportProgress(
			valueobjects.PhaseCreatingComponents,
			e.TotalItems,
			0,
		)

	case events.ImportProgressUpdated:
		s.progress, _ = valueobjects.NewImportProgress(
			e.Phase,
			e.TotalItems,
			e.CompletedItems,
		)

	case events.ImportCompleted:
		s.status = valueobjects.ImportStatusCompleted()
		s.result = ImportResult{
			CapabilitiesCreated: e.CapabilitiesCreated,
			ComponentsCreated:   e.ComponentsCreated,
			RealizationsCreated: e.RealizationsCreated,
			DomainAssignments:   e.DomainAssignments,
			Errors:              deserializeErrors(e.Errors),
		}
		completedAt := e.CompletedAt
		s.completedAt = &completedAt

	case events.ImportFailed:
		s.status = valueobjects.ImportStatusFailed()
		failedAt := e.FailedAt
		s.completedAt = &failedAt

	case events.ImportSessionCancelled:
		s.isCancelled = true
	}
}

func LoadImportSessionFromHistory(domainEvents []domain.DomainEvent) (*ImportSession, error) {
	session := &ImportSession{
		AggregateRoot: domain.NewAggregateRoot(),
	}
	session.LoadFromHistory(domainEvents, func(event domain.DomainEvent) {
		session.apply(event)
	})
	return session, nil
}

func serializeElements(elements []ParsedElement) []map[string]interface{} {
	result := make([]map[string]interface{}, len(elements))
	for i, e := range elements {
		result[i] = map[string]interface{}{
			"sourceId":    e.SourceID,
			"name":        e.Name,
			"description": e.Description,
			"parentId":    e.ParentID,
		}
	}
	return result
}

func serializeRelationships(rels []ParsedRelationship) []map[string]interface{} {
	result := make([]map[string]interface{}, len(rels))
	for i, r := range rels {
		result[i] = map[string]interface{}{
			"sourceId":      r.SourceID,
			"type":          r.Type,
			"sourceRef":     r.SourceRef,
			"targetRef":     r.TargetRef,
			"name":          r.Name,
			"documentation": r.Documentation,
		}
	}
	return result
}

func deserializePreview(data map[string]interface{}) valueobjects.ImportPreview {
	supported := valueobjects.SupportedCounts{}
	unsupported := valueobjects.UnsupportedCounts{
		Elements:      make(map[string]int),
		Relationships: make(map[string]int),
	}

	if s, ok := data["supported"].(map[string]interface{}); ok {
		if v, ok := s["capabilities"].(int); ok {
			supported.Capabilities = v
		}
		if v, ok := s["capabilities"].(float64); ok {
			supported.Capabilities = int(v)
		}
		if v, ok := s["components"].(int); ok {
			supported.Components = v
		}
		if v, ok := s["components"].(float64); ok {
			supported.Components = int(v)
		}
		if v, ok := s["parentChildRelationships"].(int); ok {
			supported.ParentChildRelationships = v
		}
		if v, ok := s["parentChildRelationships"].(float64); ok {
			supported.ParentChildRelationships = int(v)
		}
		if v, ok := s["realizations"].(int); ok {
			supported.Realizations = v
		}
		if v, ok := s["realizations"].(float64); ok {
			supported.Realizations = int(v)
		}
	}

	if u, ok := data["unsupported"].(map[string]interface{}); ok {
		if elems, ok := u["elements"].(map[string]interface{}); ok {
			for k, v := range elems {
				if count, ok := v.(int); ok {
					unsupported.Elements[k] = count
				}
				if count, ok := v.(float64); ok {
					unsupported.Elements[k] = int(count)
				}
			}
		}
		if rels, ok := u["relationships"].(map[string]interface{}); ok {
			for k, v := range rels {
				if count, ok := v.(int); ok {
					unsupported.Relationships[k] = count
				}
				if count, ok := v.(float64); ok {
					unsupported.Relationships[k] = int(count)
				}
			}
		}
	}

	return valueobjects.NewImportPreview(supported, unsupported)
}

func deserializeParsedData(data map[string]interface{}) ParsedData {
	result := ParsedData{}

	if caps, ok := data["capabilities"].([]interface{}); ok {
		for _, c := range caps {
			if m, ok := c.(map[string]interface{}); ok {
				result.Capabilities = append(result.Capabilities, ParsedElement{
					SourceID:    getString(m, "sourceId"),
					Name:        getString(m, "name"),
					Description: getString(m, "description"),
					ParentID:    getString(m, "parentId"),
				})
			}
		}
	} else if caps, ok := data["capabilities"].([]map[string]interface{}); ok {
		for _, m := range caps {
			result.Capabilities = append(result.Capabilities, ParsedElement{
				SourceID:    getString(m, "sourceId"),
				Name:        getString(m, "name"),
				Description: getString(m, "description"),
				ParentID:    getString(m, "parentId"),
			})
		}
	}

	if comps, ok := data["components"].([]interface{}); ok {
		for _, c := range comps {
			if m, ok := c.(map[string]interface{}); ok {
				result.Components = append(result.Components, ParsedElement{
					SourceID:    getString(m, "sourceId"),
					Name:        getString(m, "name"),
					Description: getString(m, "description"),
				})
			}
		}
	} else if comps, ok := data["components"].([]map[string]interface{}); ok {
		for _, m := range comps {
			result.Components = append(result.Components, ParsedElement{
				SourceID:    getString(m, "sourceId"),
				Name:        getString(m, "name"),
				Description: getString(m, "description"),
			})
		}
	}

	if rels, ok := data["relationships"].([]interface{}); ok {
		for _, r := range rels {
			if m, ok := r.(map[string]interface{}); ok {
				result.Relationships = append(result.Relationships, ParsedRelationship{
					SourceID:      getString(m, "sourceId"),
					Type:          getString(m, "type"),
					SourceRef:     getString(m, "sourceRef"),
					TargetRef:     getString(m, "targetRef"),
					Name:          getString(m, "name"),
					Documentation: getString(m, "documentation"),
				})
			}
		}
	} else if rels, ok := data["relationships"].([]map[string]interface{}); ok {
		for _, m := range rels {
			result.Relationships = append(result.Relationships, ParsedRelationship{
				SourceID:      getString(m, "sourceId"),
				Type:          getString(m, "type"),
				SourceRef:     getString(m, "sourceRef"),
				TargetRef:     getString(m, "targetRef"),
				Name:          getString(m, "name"),
				Documentation: getString(m, "documentation"),
			})
		}
	}

	return result
}

func deserializeErrors(errors []map[string]interface{}) []valueobjects.ImportError {
	result := make([]valueobjects.ImportError, len(errors))
	for i, e := range errors {
		result[i] = valueobjects.NewImportError(
			getString(e, "sourceElement"),
			getString(e, "sourceName"),
			getString(e, "error"),
			getString(e, "action"),
		)
	}
	return result
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
