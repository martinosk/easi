package aggregates

import (
	"errors"
	"time"

	"easi/backend/internal/importing/domain/events"
	"easi/backend/internal/importing/domain/valueobjects"
	"easi/backend/internal/shared/domain"
)

var (
	ErrImportAlreadyStarted      = errors.New("import has already been started")
	ErrImportNotStarted          = errors.New("import has not been started")
	ErrCannotCancelStartedImport = errors.New("cannot cancel import that has already started")
	ErrImportAlreadyCompleted    = errors.New("import has already completed")
)

type ParsedElement struct {
	SourceID    string
	Name        string
	Description string
	ParentID    string
}

type ParsedRelationship struct {
	SourceID      string
	Type          string
	SourceRef     string
	TargetRef     string
	Name          string
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

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getIntMap(m map[string]interface{}, key string) map[string]int {
	result := make(map[string]int)
	raw, ok := m[key].(map[string]interface{})
	if !ok {
		return result
	}
	for k, v := range raw {
		result[k] = getInt(map[string]interface{}{k: v}, k)
	}
	return result
}

func deserializeSupportedCounts(data map[string]interface{}) valueobjects.SupportedCounts {
	s, ok := data["supported"].(map[string]interface{})
	if !ok {
		return valueobjects.SupportedCounts{}
	}
	return valueobjects.SupportedCounts{
		Capabilities:             getInt(s, "capabilities"),
		Components:               getInt(s, "components"),
		ParentChildRelationships: getInt(s, "parentChildRelationships"),
		Realizations:             getInt(s, "realizations"),
	}
}

func deserializeUnsupportedCounts(data map[string]interface{}) valueobjects.UnsupportedCounts {
	u, ok := data["unsupported"].(map[string]interface{})
	if !ok {
		return valueobjects.UnsupportedCounts{
			Elements:      make(map[string]int),
			Relationships: make(map[string]int),
		}
	}
	return valueobjects.UnsupportedCounts{
		Elements:      getIntMap(u, "elements"),
		Relationships: getIntMap(u, "relationships"),
	}
}

func deserializePreview(data map[string]interface{}) valueobjects.ImportPreview {
	supported := deserializeSupportedCounts(data)
	unsupported := deserializeUnsupportedCounts(data)
	return valueobjects.NewImportPreview(supported, unsupported)
}

func toMapSlice(data interface{}) []map[string]interface{} {
	if slice, ok := data.([]interface{}); ok {
		result := make([]map[string]interface{}, 0, len(slice))
		for _, item := range slice {
			if m, ok := item.(map[string]interface{}); ok {
				result = append(result, m)
			}
		}
		return result
	}
	if slice, ok := data.([]map[string]interface{}); ok {
		return slice
	}
	return nil
}

func deserializeCapabilities(data map[string]interface{}) []ParsedElement {
	maps := toMapSlice(data["capabilities"])
	result := make([]ParsedElement, 0, len(maps))
	for _, m := range maps {
		result = append(result, ParsedElement{
			SourceID:    getString(m, "sourceId"),
			Name:        getString(m, "name"),
			Description: getString(m, "description"),
			ParentID:    getString(m, "parentId"),
		})
	}
	return result
}

func deserializeComponents(data map[string]interface{}) []ParsedElement {
	maps := toMapSlice(data["components"])
	result := make([]ParsedElement, 0, len(maps))
	for _, m := range maps {
		result = append(result, ParsedElement{
			SourceID:    getString(m, "sourceId"),
			Name:        getString(m, "name"),
			Description: getString(m, "description"),
		})
	}
	return result
}

func deserializeRelationships(data map[string]interface{}) []ParsedRelationship {
	maps := toMapSlice(data["relationships"])
	result := make([]ParsedRelationship, 0, len(maps))
	for _, m := range maps {
		result = append(result, ParsedRelationship{
			SourceID:      getString(m, "sourceId"),
			Type:          getString(m, "type"),
			SourceRef:     getString(m, "sourceRef"),
			TargetRef:     getString(m, "targetRef"),
			Name:          getString(m, "name"),
			Documentation: getString(m, "documentation"),
		})
	}
	return result
}

func deserializeParsedData(data map[string]interface{}) ParsedData {
	return ParsedData{
		Capabilities:  deserializeCapabilities(data),
		Components:    deserializeComponents(data),
		Relationships: deserializeRelationships(data),
	}
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
