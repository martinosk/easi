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
	ValueStreams             []ParsedElement
	Relationships            []ParsedRelationship
	UnsupportedElements      map[string]int
	UnsupportedRelationships map[string]int
}

func (pr *ParseResult) GetPreview() valueobjects.ImportPreview {
	capabilityIDs := collectIDs(pr.Capabilities)
	componentIDs := collectIDs(pr.Components)
	valueStreamIDs := collectIDs(pr.ValueStreams)

	counts := pr.countRelationships(capabilityIDs, componentIDs, valueStreamIDs)

	supported := valueobjects.SupportedCounts{
		Capabilities:                    len(pr.Capabilities),
		Components:                      len(pr.Components),
		ValueStreams:                    len(pr.ValueStreams),
		ParentChildRelationships:        counts.ParentChild,
		Realizations:                    counts.Realization,
		ComponentRelationships:          counts.ComponentRelation,
		CapabilityToValueStreamMappings: counts.CapabilityToValueStream,
	}

	unsupported := valueobjects.UnsupportedCounts{
		Elements:      pr.UnsupportedElements,
		Relationships: pr.UnsupportedRelationships,
	}

	return valueobjects.NewImportPreview(supported, unsupported)
}

type relationshipCounts struct {
	ParentChild             int
	Realization             int
	ComponentRelation       int
	CapabilityToValueStream int
}

type relationshipCounter struct {
	capabilityIDs  map[string]bool
	componentIDs   map[string]bool
	valueStreamIDs map[string]bool
}

func (pr *ParseResult) countRelationships(capabilityIDs, componentIDs, valueStreamIDs map[string]bool) relationshipCounts {
	counter := relationshipCounter{
		capabilityIDs:  capabilityIDs,
		componentIDs:   componentIDs,
		valueStreamIDs: valueStreamIDs,
	}
	counts := relationshipCounts{}
	for _, rel := range pr.Relationships {
		counter.countRelationship(rel, pr, &counts)
	}
	return counts
}

func (c relationshipCounter) countRelationship(rel ParsedRelationship, pr *ParseResult, counts *relationshipCounts) {
	switch rel.Type {
	case "Aggregation", "Composition":
		c.countAggregation(counts)
	case "Realization":
		c.countRealization(rel, pr, counts)
	case "Association":
		c.countAssociation(counts)
	case "Triggering", "Serving":
		c.countTriggeringOrServing(rel, pr, counts)
	}
}

func (c relationshipCounter) countAggregation(counts *relationshipCounts) {
	counts.ParentChild++
}

func (c relationshipCounter) countRealization(rel ParsedRelationship, pr *ParseResult, counts *relationshipCounts) {
	if pr.isCapabilityToValueStreamRealization(rel, c.capabilityIDs, c.valueStreamIDs) {
		counts.CapabilityToValueStream++
	} else {
		counts.Realization++
	}
}

func (c relationshipCounter) countAssociation(counts *relationshipCounts) {
	counts.CapabilityToValueStream++
}

func (c relationshipCounter) countTriggeringOrServing(rel ParsedRelationship, pr *ParseResult, counts *relationshipCounts) {
	if pr.isCapabilityValueStreamRelation(rel, c.capabilityIDs, c.valueStreamIDs) {
		counts.CapabilityToValueStream++
	} else if c.isComponentToComponent(rel) {
		counts.ComponentRelation++
	}
}

func (c relationshipCounter) isComponentToComponent(rel ParsedRelationship) bool {
	return c.componentIDs[rel.SourceRef] && c.componentIDs[rel.TargetRef]
}

func (pr *ParseResult) isCapabilityToValueStreamRealization(rel ParsedRelationship, capabilityIDs, valueStreamIDs map[string]bool) bool {
	return capabilityIDs[rel.SourceRef] && valueStreamIDs[rel.TargetRef]
}

func (pr *ParseResult) isCapabilityValueStreamRelation(rel ParsedRelationship, capabilityIDs, valueStreamIDs map[string]bool) bool {
	return (capabilityIDs[rel.SourceRef] && valueStreamIDs[rel.TargetRef]) ||
		(valueStreamIDs[rel.SourceRef] && capabilityIDs[rel.TargetRef])
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
	Type          string `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	Name          string `xml:"name"`
	Documentation string `xml:"documentation"`
}

type archiMateRelationships struct {
	Relationship []archiMateRelationship `xml:"relationship"`
}

type archiMateRelationship struct {
	Identifier    string `xml:"identifier,attr"`
	Type          string `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	Source        string `xml:"source,attr"`
	Target        string `xml:"target,attr"`
	Name          string `xml:"name"`
	Documentation string `xml:"documentation"`
}

var supportedElementTypes = map[string]bool{
	"Capability":           true,
	"ApplicationComponent": true,
	"ApplicationService":   true,
	"ValueStream":          true,
}

var supportedRelationshipTypes = map[string]bool{
	"Aggregation": true,
	"Composition": true,
	"Realization": true,
	"Association": true,
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
	valueStreamIDs := collectIDs(result.ValueStreams)
	validator := newRelationshipValidator(capabilityIDs, componentIDs, valueStreamIDs)

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
		} else if elem.Type == "ValueStream" {
			result.ValueStreams = append(result.ValueStreams, parsed)
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
	capabilityIDs  map[string]bool
	componentIDs   map[string]bool
	valueStreamIDs map[string]bool
}

func newRelationshipValidator(capabilityIDs, componentIDs, valueStreamIDs map[string]bool) relationshipValidator {
	return relationshipValidator{
		capabilityIDs:  capabilityIDs,
		componentIDs:   componentIDs,
		valueStreamIDs: valueStreamIDs,
	}
}

func (v relationshipValidator) hasValidEndpoints(rel archiMateRelationship) bool {
	switch rel.Type {
	case "Aggregation", "Composition":
		return v.isCapabilityToCapability(rel)
	case "Realization":
		return v.isComponentToCapability(rel) || v.isCapabilityToValueStream(rel)
	case "Association":
		return v.isCapabilityToValueStream(rel)
	case "Triggering", "Serving":
		return v.isComponentToComponent(rel) || v.isCapabilityToValueStream(rel)
	default:
		return true
	}
}

func (v relationshipValidator) isCapabilityToCapability(rel archiMateRelationship) bool {
	return v.capabilityIDs[rel.Source] && v.capabilityIDs[rel.Target]
}

func (v relationshipValidator) isComponentToCapability(rel archiMateRelationship) bool {
	return v.componentIDs[rel.Source] && v.capabilityIDs[rel.Target]
}

func (v relationshipValidator) isComponentToComponent(rel archiMateRelationship) bool {
	return v.componentIDs[rel.Source] && v.componentIDs[rel.Target]
}

func (v relationshipValidator) isCapabilityToValueStream(rel archiMateRelationship) bool {
	return v.capabilityIDs[rel.Source] && v.valueStreamIDs[rel.Target]
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
