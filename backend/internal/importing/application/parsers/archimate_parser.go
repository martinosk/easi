package parsers

import (
	"encoding/xml"
	"io"

	"easi/backend/internal/importing/domain/valueobjects"
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

type ParseResult struct {
	Capabilities             []ParsedElement
	Components               []ParsedElement
	Relationships            []ParsedRelationship
	UnsupportedElements      map[string]int
	UnsupportedRelationships map[string]int
}

func (pr *ParseResult) GetPreview() valueobjects.ImportPreview {
	parentChildCount := 0
	realizationCount := 0
	componentRelationCount := 0
	for _, rel := range pr.Relationships {
		switch rel.Type {
		case "Aggregation", "Composition":
			parentChildCount++
		case "Realization":
			realizationCount++
		case "Triggering", "Serving":
			componentRelationCount++
		}
	}

	supported := valueobjects.SupportedCounts{
		Capabilities:             len(pr.Capabilities),
		Components:               len(pr.Components),
		ParentChildRelationships: parentChildCount,
		Realizations:             realizationCount,
		ComponentRelationships:   componentRelationCount,
	}

	unsupported := valueobjects.UnsupportedCounts{
		Elements:      pr.UnsupportedElements,
		Relationships: pr.UnsupportedRelationships,
	}

	return valueobjects.NewImportPreview(supported, unsupported)
}

type ArchiMateParser struct{}

func NewArchiMateParser() *ArchiMateParser {
	return &ArchiMateParser{}
}

type archiMateModel struct {
	XMLName       xml.Name               `xml:"model"`
	Elements      archiMateElements      `xml:"elements"`
	Relationships archiMateRelationships `xml:"relationships"`
}

type archiMateElements struct {
	Element []archiMateElement `xml:"element"`
}

type archiMateElement struct {
	Identifier    string `xml:"identifier,attr"`
	Type          string `xml:"type,attr"`
	Name          string `xml:"name"`
	Documentation string `xml:"documentation"`
}

type archiMateRelationships struct {
	Relationship []archiMateRelationship `xml:"relationship"`
}

type archiMateRelationship struct {
	Identifier    string `xml:"identifier,attr"`
	Type          string `xml:"type,attr"`
	Source        string `xml:"source,attr"`
	Target        string `xml:"target,attr"`
	Name          string `xml:"name"`
	Documentation string `xml:"documentation"`
}

var supportedElementTypes = map[string]bool{
	"Capability":           true,
	"ApplicationComponent": true,
	"ApplicationService":   true,
}

var supportedRelationshipTypes = map[string]bool{
	"Aggregation": true,
	"Composition": true,
	"Realization": true,
	"Triggering":  true,
	"Serving":     true,
}

func (p *ArchiMateParser) Parse(reader io.Reader) (*ParseResult, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var model archiMateModel
	if err := xml.Unmarshal(data, &model); err != nil {
		return nil, err
	}

	result := &ParseResult{
		UnsupportedElements:      make(map[string]int),
		UnsupportedRelationships: make(map[string]int),
	}

	classifyElements(model.Elements.Element, result)

	capabilityIDs := collectIDs(result.Capabilities)
	componentIDs := collectIDs(result.Components)
	validator := newRelationshipValidator(capabilityIDs, componentIDs)

	classifyRelationships(model.Relationships.Relationship, validator, result)

	return result, nil
}

func classifyElements(elements []archiMateElement, result *ParseResult) {
	for _, elem := range elements {
		if !supportedElementTypes[elem.Type] {
			result.UnsupportedElements[elem.Type]++
			continue
		}

		parsed := ParsedElement{
			SourceID:    elem.Identifier,
			Name:        elem.Name,
			Description: elem.Documentation,
		}

		if elem.Type == "Capability" {
			result.Capabilities = append(result.Capabilities, parsed)
		} else {
			result.Components = append(result.Components, parsed)
		}
	}
}

func collectIDs(elements []ParsedElement) map[string]bool {
	ids := make(map[string]bool, len(elements))
	for _, e := range elements {
		ids[e.SourceID] = true
	}
	return ids
}

type relationshipValidator struct {
	capabilityIDs map[string]bool
	componentIDs  map[string]bool
}

func newRelationshipValidator(capabilityIDs, componentIDs map[string]bool) relationshipValidator {
	return relationshipValidator{capabilityIDs: capabilityIDs, componentIDs: componentIDs}
}

func (v relationshipValidator) hasValidEndpoints(rel archiMateRelationship) bool {
	switch rel.Type {
	case "Aggregation", "Composition":
		return v.capabilityIDs[rel.Source] && v.capabilityIDs[rel.Target]
	case "Realization":
		return v.componentIDs[rel.Source] && v.capabilityIDs[rel.Target]
	case "Triggering", "Serving":
		return v.componentIDs[rel.Source] && v.componentIDs[rel.Target]
	default:
		return true
	}
}

func classifyRelationships(relationships []archiMateRelationship, validator relationshipValidator, result *ParseResult) {
	for _, rel := range relationships {
		if !supportedRelationshipTypes[rel.Type] || !validator.hasValidEndpoints(rel) {
			result.UnsupportedRelationships[rel.Type]++
			continue
		}

		result.Relationships = append(result.Relationships, ParsedRelationship{
			SourceID:      rel.Identifier,
			Type:          rel.Type,
			SourceRef:     rel.Source,
			TargetRef:     rel.Target,
			Name:          rel.Name,
			Documentation: rel.Documentation,
		})
	}
}
