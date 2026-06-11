package services

import (
	"context"
	"fmt"
	"strings"

	domainservices "easi/backend/internal/enterprisearchitecture/domain/services"
)

type DirectionSources struct {
	EnterpriseCapabilityID string
	Status                 string
	SourceCapabilityIDs    []string
}

type DirectionSourcesProvider interface {
	ActiveDirectionSources(ctx context.Context) ([]DirectionSources, error)
}

type CapabilityNodesProvider interface {
	AllCapabilityNodes(ctx context.Context) ([]domainservices.CapabilityNode, error)
}

type EnterpriseCapabilityNamesProvider interface {
	ActiveEnterpriseCapabilityNames(ctx context.Context) (map[string]string, error)
}

type CompositionService struct {
	directions   DirectionSourcesProvider
	capabilities CapabilityNodesProvider
	ecNames      EnterpriseCapabilityNamesProvider
}

func NewCompositionService(
	directions DirectionSourcesProvider,
	capabilities CapabilityNodesProvider,
	ecNames EnterpriseCapabilityNamesProvider,
) *CompositionService {
	return &CompositionService{directions: directions, capabilities: capabilities, ecNames: ecNames}
}

type CompositionResult struct {
	HasActiveDirection bool
	DirectionStatus    string
	Resolved           []domainservices.ResolvedCapability
	Counts             domainservices.CompositionCounts
}

type SourceEligibility struct {
	CapabilityID                    string
	Eligible                        bool
	IneligibilityReason             *string
	ConflictingEnterpriseCapability *domainservices.CarvedOutBy
}

type PreviewResult struct {
	Resolved          []domainservices.ResolvedCapability
	SourceEligibility []SourceEligibility
	Counts            domainservices.CompositionCounts
}

type SourceConflict struct {
	CapabilityID                        string
	CapabilityName                      string
	ConflictingEnterpriseCapabilityID   string
	ConflictingEnterpriseCapabilityName string
}

type SourceCandidatesQuery struct {
	EnterpriseCapabilityID string
	Search                 string
	BusinessDomainID       string
	Limit                  int
}

type SourceCandidate struct {
	Node                            domainservices.CapabilityNode
	Eligible                        bool
	IneligibilityReason             *string
	ConflictingEnterpriseCapability *domainservices.CarvedOutBy
}

type SourceCandidatesResult struct {
	Candidates []SourceCandidate
	HasMore    bool
}

type compositionInputs struct {
	directions   []domainservices.ActiveDirectionSources
	capabilities []domainservices.CapabilityNode
	statusByEC   map[string]string
}

func (s *CompositionService) CompositionForEC(ctx context.Context, ecID string) (CompositionResult, error) {
	inputs, err := s.loadInputs(ctx)
	if err != nil {
		return CompositionResult{}, err
	}
	status, hasDirection := inputs.statusByEC[ecID]
	resolved := domainservices.ResolveComposition(ecID, inputs.directions, inputs.capabilities)
	return CompositionResult{
		HasActiveDirection: hasDirection,
		DirectionStatus:    status,
		Resolved:           resolved,
		Counts:             domainservices.CountComposition(sourceCountFor(inputs.directions, ecID), resolved),
	}, nil
}

func (s *CompositionService) CountsForAll(ctx context.Context) (map[string]domainservices.CompositionCounts, error) {
	inputs, err := s.loadInputs(ctx)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]domainservices.CompositionCounts, len(inputs.directions))
	for _, d := range inputs.directions {
		resolved := domainservices.ResolveComposition(d.EnterpriseCapabilityID, inputs.directions, inputs.capabilities)
		counts[d.EnterpriseCapabilityID] = domainservices.CountComposition(len(d.SourceCapabilityIDs), resolved)
	}
	return counts, nil
}

func (s *CompositionService) IncludedCapabilityIDsByEC(ctx context.Context) (map[string][]string, error) {
	inputs, err := s.loadInputs(ctx)
	if err != nil {
		return nil, err
	}
	included := make(map[string][]string, len(inputs.directions))
	for _, d := range inputs.directions {
		resolved := domainservices.ResolveComposition(d.EnterpriseCapabilityID, inputs.directions, inputs.capabilities)
		ids := []string{}
		for _, r := range resolved {
			if r.Role != domainservices.RoleCarvedOut {
				ids = append(ids, r.Node.ID)
			}
		}
		included[d.EnterpriseCapabilityID] = ids
	}
	return included, nil
}

func (s *CompositionService) Preview(ctx context.Context, ecID string, sourceIDs []string) (PreviewResult, error) {
	inputs, err := s.loadInputs(ctx)
	if err != nil {
		return PreviewResult{}, err
	}
	others := withoutEC(inputs.directions, ecID)
	synthetic := append([]domainservices.ActiveDirectionSources{{
		EnterpriseCapabilityID: ecID,
		SourceCapabilityIDs:    sourceIDs,
	}}, others...)
	resolved := domainservices.ResolveComposition(ecID, synthetic, inputs.capabilities)

	eligibility := make([]SourceEligibility, len(sourceIDs))
	for i, sourceID := range sourceIDs {
		eligibility[i] = evaluateEligibility(sourceID, ecID, others)
	}
	return PreviewResult{
		Resolved:          resolved,
		SourceEligibility: eligibility,
		Counts:            domainservices.CountComposition(len(sourceIDs), resolved),
	}, nil
}

func (s *CompositionService) FirstSourceConflict(ctx context.Context, ecID string, sourceIDs []string) (*SourceConflict, error) {
	inputs, err := s.loadInputs(ctx)
	if err != nil {
		return nil, err
	}
	nodesByID := indexNodes(inputs.capabilities)
	for _, sourceID := range sourceIDs {
		conflict := domainservices.EvaluateSourceEligibility(sourceID, ecID, inputs.directions)
		if conflict == nil {
			continue
		}
		name := sourceID
		if node, ok := nodesByID[sourceID]; ok {
			name = node.Name
		}
		return &SourceConflict{
			CapabilityID:                        sourceID,
			CapabilityName:                      name,
			ConflictingEnterpriseCapabilityID:   conflict.EnterpriseCapabilityID,
			ConflictingEnterpriseCapabilityName: conflict.EnterpriseCapabilityName,
		}, nil
	}
	return nil, nil
}

func (s *CompositionService) SourceCandidates(ctx context.Context, query SourceCandidatesQuery) (SourceCandidatesResult, error) {
	inputs, err := s.loadInputs(ctx)
	if err != nil {
		return SourceCandidatesResult{}, err
	}
	others := withoutEC(inputs.directions, query.EnterpriseCapabilityID)

	matches := filterCandidates(inputs.capabilities, query)
	limited := matches
	if query.Limit > 0 && len(matches) > query.Limit {
		limited = matches[:query.Limit]
	}

	candidates := make([]SourceCandidate, len(limited))
	for i, node := range limited {
		eligibility := evaluateEligibility(node.ID, query.EnterpriseCapabilityID, others)
		candidates[i] = SourceCandidate{
			Node:                            node,
			Eligible:                        eligibility.Eligible,
			IneligibilityReason:             eligibility.IneligibilityReason,
			ConflictingEnterpriseCapability: eligibility.ConflictingEnterpriseCapability,
		}
	}
	return SourceCandidatesResult{Candidates: candidates, HasMore: len(matches) > len(limited)}, nil
}

func (s *CompositionService) loadInputs(ctx context.Context) (compositionInputs, error) {
	directions, err := s.directions.ActiveDirectionSources(ctx)
	if err != nil {
		return compositionInputs{}, fmt.Errorf("load active direction sources: %w", err)
	}
	capabilities, err := s.capabilities.AllCapabilityNodes(ctx)
	if err != nil {
		return compositionInputs{}, fmt.Errorf("load capability nodes: %w", err)
	}
	ecNames, err := s.ecNames.ActiveEnterpriseCapabilityNames(ctx)
	if err != nil {
		return compositionInputs{}, fmt.Errorf("load active enterprise capability names: %w", err)
	}

	active := make([]domainservices.ActiveDirectionSources, 0, len(directions))
	statusByEC := make(map[string]string, len(directions))
	for _, d := range directions {
		name, ecIsActive := ecNames[d.EnterpriseCapabilityID]
		if !ecIsActive {
			continue
		}
		active = append(active, domainservices.ActiveDirectionSources{
			EnterpriseCapabilityID:   d.EnterpriseCapabilityID,
			EnterpriseCapabilityName: name,
			SourceCapabilityIDs:      d.SourceCapabilityIDs,
		})
		statusByEC[d.EnterpriseCapabilityID] = d.Status
	}
	return compositionInputs{directions: active, capabilities: capabilities, statusByEC: statusByEC}, nil
}

func evaluateEligibility(capabilityID, ecID string, others []domainservices.ActiveDirectionSources) SourceEligibility {
	conflict := domainservices.EvaluateSourceEligibility(capabilityID, ecID, others)
	if conflict == nil {
		return SourceEligibility{CapabilityID: capabilityID, Eligible: true}
	}
	reason := fmt.Sprintf("Already an explicit source of an active direction on '%s'", conflict.EnterpriseCapabilityName)
	return SourceEligibility{
		CapabilityID:                    capabilityID,
		Eligible:                        false,
		IneligibilityReason:             &reason,
		ConflictingEnterpriseCapability: conflict,
	}
}

func filterCandidates(nodes []domainservices.CapabilityNode, query SourceCandidatesQuery) []domainservices.CapabilityNode {
	search := strings.ToLower(query.Search)
	matches := make([]domainservices.CapabilityNode, 0)
	for _, node := range nodes {
		if !strings.Contains(strings.ToLower(node.Name), search) {
			continue
		}
		if query.BusinessDomainID != "" && node.BusinessDomainID != query.BusinessDomainID {
			continue
		}
		matches = append(matches, node)
	}
	return matches
}

func withoutEC(directions []domainservices.ActiveDirectionSources, ecID string) []domainservices.ActiveDirectionSources {
	others := make([]domainservices.ActiveDirectionSources, 0, len(directions))
	for _, d := range directions {
		if d.EnterpriseCapabilityID != ecID {
			others = append(others, d)
		}
	}
	return others
}

func sourceCountFor(directions []domainservices.ActiveDirectionSources, ecID string) int {
	for _, d := range directions {
		if d.EnterpriseCapabilityID == ecID {
			return len(d.SourceCapabilityIDs)
		}
	}
	return 0
}

func indexNodes(nodes []domainservices.CapabilityNode) map[string]domainservices.CapabilityNode {
	index := make(map[string]domainservices.CapabilityNode, len(nodes))
	for _, n := range nodes {
		index[n.ID] = n
	}
	return index
}
