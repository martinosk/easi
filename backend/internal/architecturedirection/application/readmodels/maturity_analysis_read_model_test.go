package readmodels

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func maturityRow(id string, value int, domainID string) capabilityMaturityRow {
	return capabilityMaturityRow{
		CapabilityID:       id,
		CapabilityName:     "Capability " + id,
		BusinessDomainID:   domainID,
		BusinessDomainName: domainID,
		MaturityValue:      value,
	}
}

func TestBuildMaturityCandidate_AggregatesIncludedCapabilities(t *testing.T) {
	target := 80
	ec := ecHeaderRow{ID: "ec-1", Name: "Customer Identity", Category: "core", TargetMaturity: &target}
	meta := map[string]capabilityMaturityRow{
		"cap-1": maturityRow("cap-1", 20, "dom-1"),
		"cap-2": maturityRow("cap-2", 60, "dom-1"),
		"cap-3": maturityRow("cap-3", 90, "dom-2"),
	}

	candidate := buildMaturityCandidate(ec, []string{"cap-1", "cap-2", "cap-3"}, meta)

	assert.Equal(t, 3, candidate.ImplementationCount)
	assert.Equal(t, 2, candidate.DomainCount)
	assert.Equal(t, 90, candidate.MaxMaturity)
	assert.Equal(t, 20, candidate.MinMaturity)
	assert.Equal(t, 57, candidate.AverageMaturity, "average is rounded, matching the previous SQL ::int cast")
	assert.Equal(t, 60, candidate.MaxGap, "gap = target - min maturity")
	assert.Equal(t, MaturityDistributionDTO{Genesis: 1, CustomBuild: 0, Product: 1, Commodity: 1}, candidate.MaturityDistribution)
}

func TestBuildMaturityCandidate_NoTargetUsesMaxMaturity(t *testing.T) {
	ec := ecHeaderRow{ID: "ec-1", Name: "Customer Identity"}
	meta := map[string]capabilityMaturityRow{
		"cap-1": maturityRow("cap-1", 30, ""),
		"cap-2": maturityRow("cap-2", 70, ""),
	}

	candidate := buildMaturityCandidate(ec, []string{"cap-1", "cap-2"}, meta)

	assert.Equal(t, 40, candidate.MaxGap, "without a target, the gap is max - min maturity")
	assert.Nil(t, candidate.TargetMaturity)
}

func TestBuildMaturityCandidate_NoIncludedCapabilities(t *testing.T) {
	candidate := buildMaturityCandidate(ecHeaderRow{ID: "ec-1", Name: "Empty"}, nil, nil)

	assert.Equal(t, 0, candidate.ImplementationCount)
	assert.Equal(t, 0, candidate.MaxGap)
	assert.Equal(t, 0, candidate.AverageMaturity)
}

func TestSortMaturityCandidates(t *testing.T) {
	candidates := []MaturityAnalysisCandidateDTO{
		{EnterpriseCapabilityName: "A", MaxGap: 10, ImplementationCount: 5},
		{EnterpriseCapabilityName: "B", MaxGap: 40, ImplementationCount: 1},
		{EnterpriseCapabilityName: "C", MaxGap: 10, ImplementationCount: 9},
	}

	sortMaturityCandidates(candidates, "")
	assert.Equal(t, []string{"B", "C", "A"}, candidateNames(candidates), "default sort: max gap, then implementations")

	sortMaturityCandidates(candidates, "implementations")
	assert.Equal(t, []string{"C", "A", "B"}, candidateNames(candidates), "implementations sort: count, then max gap")
}

func candidateNames(candidates []MaturityAnalysisCandidateDTO) []string {
	names := make([]string, len(candidates))
	for i, c := range candidates {
		names[i] = c.EnterpriseCapabilityName
	}
	return names
}

func TestBuildGapImplementations_SortedByMaturityWithGapAndPriority(t *testing.T) {
	meta := map[string]capabilityMaturityRow{
		"cap-1": maturityRow("cap-1", 10, "dom-1"),
		"cap-2": maturityRow("cap-2", 80, ""),
		"cap-3": maturityRow("cap-3", 60, ""),
	}

	implementations := buildGapImplementations([]string{"cap-2", "cap-1", "cap-3"}, meta, 80)

	require.Len(t, implementations, 3)
	assert.Equal(t, "cap-1", implementations[0].DomainCapabilityID, "sorted ascending by maturity")
	assert.Equal(t, 70, implementations[0].Gap)
	assert.Equal(t, "High", implementations[0].Priority)
	assert.Equal(t, "Genesis", implementations[0].MaturitySection)
	assert.Equal(t, "Medium", implementations[1].Priority)
	assert.Equal(t, 0, implementations[2].Gap)
	assert.Equal(t, "None", implementations[2].Priority)
}
