package api

import (
	"encoding/base64"
	"encoding/json"
	"time"

	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/valueobjects"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type QualityRowDTO struct {
	SubjectType   string      `json:"subjectType"`
	SubjectID     string      `json:"subjectId"`
	Name          string      `json:"name"`
	Completeness  string      `json:"completeness"`
	RequiredCount int         `json:"requiredCount"`
	FilledCount   int         `json:"filledCount"`
	MissingCount  int         `json:"missingCount"`
	CreatorID     string      `json:"creatorId"`
	CreatorEmail  string      `json:"creatorEmail"`
	CreatedAt     time.Time   `json:"createdAt"`
	LastUpdatedAt time.Time   `json:"lastUpdatedAt"`
	Links         types.Links `json:"_links,omitempty"`
}

var qualitySubjectGrantPermission = map[string]sharedctx.ResourceName{
	"capability":      "capabilities",
	"application":     "components",
	"acquired-entity": "components",
	"vendor":          "components",
	"internal-team":   "components",
}

func toQualityRow(record readmodels.SubjectIndexRecord, actor sharedctx.Actor, links *OnePagerLinks) QualityRowDTO {
	return QualityRowDTO{
		SubjectType:   record.SubjectType,
		SubjectID:     record.SubjectID,
		Name:          record.Name,
		Completeness:  record.Signal(),
		RequiredCount: record.RequiredCount,
		FilledCount:   record.FilledCount,
		MissingCount:  record.MissingCount(),
		CreatorID:     record.CreatorActorID,
		CreatorEmail:  record.CreatorEmail,
		CreatedAt:     record.CreatedAt,
		LastUpdatedAt: record.LastUpdatedAt,
		Links:         qualityRowLinks(links, record, actor),
	}
}

func qualityRowLinks(links *OnePagerLinks, record readmodels.SubjectIndexRecord, actor sharedctx.Actor) types.Links {
	permission, supported := qualitySubjectGrantPermission[record.SubjectType]
	if !supported {
		return nil
	}
	rowLinks := types.Links{}
	links.AddEditGrantsLink(rowLinks, actor, permission)
	return rowLinks
}

type qualityCursor struct {
	SubjectType   string    `json:"st"`
	SubjectID     string    `json:"sid"`
	Name          string    `json:"n"`
	CreatorEmail  string    `json:"e"`
	CreatedAt     time.Time `json:"c"`
	LastUpdatedAt time.Time `json:"u"`
	RequiredCount int       `json:"rq"`
	FilledCount   int       `json:"fl"`
}

func encodeQualityCursor(record readmodels.SubjectIndexRecord) string {
	cursor := qualityCursor{
		SubjectType:   record.SubjectType,
		SubjectID:     record.SubjectID,
		Name:          record.Name,
		CreatorEmail:  record.CreatorEmail,
		CreatedAt:     record.CreatedAt,
		LastUpdatedAt: record.LastUpdatedAt,
		RequiredCount: record.RequiredCount,
		FilledCount:   record.FilledCount,
	}
	data, _ := json.Marshal(cursor)
	return base64.URLEncoding.EncodeToString(data)
}

func decodeQualityCursor(token string) (*readmodels.SubjectIndexRecord, error) {
	if token == "" {
		return nil, nil
	}
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	var cursor qualityCursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, err
	}
	return &readmodels.SubjectIndexRecord{
		SubjectType:   cursor.SubjectType,
		SubjectID:     cursor.SubjectID,
		Name:          cursor.Name,
		CreatorEmail:  cursor.CreatorEmail,
		CreatedAt:     cursor.CreatedAt,
		LastUpdatedAt: cursor.LastUpdatedAt,
		RequiredCount: cursor.RequiredCount,
		FilledCount:   cursor.FilledCount,
	}, nil
}

func parseQualitySort(raw string) (string, bool) {
	switch raw {
	case "":
		return readmodels.SortCompleteness, true
	case readmodels.SortCompleteness, readmodels.SortCreator, readmodels.SortName, readmodels.SortCreated, readmodels.SortUpdated:
		return raw, true
	default:
		return "", false
	}
}

func parseQualityOrder(raw string) (string, bool) {
	switch raw {
	case "":
		return readmodels.OrderAsc, true
	case readmodels.OrderAsc, readmodels.OrderDesc:
		return raw, true
	default:
		return "", false
	}
}

var subjectTypesByReadPermission = []struct {
	permission   string
	subjectTypes []string
}{
	{authPL.PermCapabilitiesRead.String(), []string{"capability"}},
	{authPL.PermEnterpriseArchRead.String(), []string{"enterprise-capability"}},
	{authPL.PermComponentsRead.String(), []string{"application", "acquired-entity", "vendor", "internal-team"}},
}

func readableSubjectTypes(actor sharedctx.Actor) []string {
	var readable []string
	for _, grant := range subjectTypesByReadPermission {
		if actor.HasPermission(grant.permission) {
			readable = append(readable, grant.subjectTypes...)
		}
	}
	return orderedSubjectTypes(readable)
}

func orderedSubjectTypes(subjectTypes []string) []string {
	if len(subjectTypes) == 0 {
		return nil
	}
	allowed := make(map[string]bool, len(subjectTypes))
	for _, subjectType := range subjectTypes {
		allowed[subjectType] = true
	}
	ordered := make([]string, 0, len(subjectTypes))
	for _, subjectType := range valueobjects.AllSubjectTypes() {
		if allowed[subjectType.Value()] {
			ordered = append(ordered, subjectType.Value())
		}
	}
	return ordered
}
