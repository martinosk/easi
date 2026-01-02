import { useState, useCallback } from 'react';
import { useMaturityAnalysis } from '../hooks/useMaturityAnalysis';
import { useMaturityScale } from '../../../hooks/useMaturityScale';
import { useMaturityColorScale } from '../../../hooks/useMaturityColorScale';
import { HelpTooltip } from '../../../components/shared/HelpTooltip';
import type { MaturityAnalysisCandidate, EnterpriseCapabilityId } from '../types';
import './MaturityAnalysisTab.css';

function MaturitySectionLegend() {
  const { data: maturityScale } = useMaturityScale();
  const { getBaseSectionColor } = useMaturityColorScale();

  if (!maturityScale?.sections) return null;

  const sortedSections = [...maturityScale.sections].sort((a, b) => a.order - b.order);

  return (
    <div className="maturity-section-legend">
      <span className="legend-title">
        Maturity Sections
        <HelpTooltip
          content="Color coding for the maturity distribution bar. Sections are configured in Settings."
          iconOnly
        />
      </span>
      <div className="legend-items">
        {sortedSections.map(section => (
          <div key={section.order} className="legend-section-item">
            <span
              className="legend-section-dot"
              style={{ backgroundColor: getBaseSectionColor(section.order) }}
            />
            <span className="legend-section-name">{section.name}</span>
            <span className="legend-section-range">({section.minValue}-{section.maxValue})</span>
          </div>
        ))}
      </div>
    </div>
  );
}

interface MaturityDistributionBarProps {
  distribution: {
    genesis: number;
    customBuild: number;
    product: number;
    commodity: number;
  };
}

function MaturityDistributionBar({ distribution }: MaturityDistributionBarProps) {
  const { data: maturityScale } = useMaturityScale();
  const { getBaseSectionColor } = useMaturityColorScale();

  const total = distribution.genesis + distribution.customBuild + distribution.product + distribution.commodity;
  if (total === 0) return null;

  const sortedSections = maturityScale?.sections
    ? [...maturityScale.sections].sort((a, b) => a.order - b.order)
    : [];

  const distributionByOrder = [
    distribution.genesis,
    distribution.customBuild,
    distribution.product,
    distribution.commodity,
  ];

  const segments = sortedSections
    .map((section, index) => ({
      name: section.name,
      count: distributionByOrder[index] || 0,
      color: getBaseSectionColor(section.order),
    }))
    .filter(s => s.count > 0);

  return (
    <div className="maturity-distribution">
      <div className="distribution-bar">
        {segments.map(segment => (
          <div
            key={segment.name}
            className="distribution-segment"
            style={{
              width: `${(segment.count / total) * 100}%`,
              backgroundColor: segment.color,
            }}
            title={`${segment.name}: ${segment.count}`}
          />
        ))}
      </div>
      <div className="distribution-legend">
        {segments.map(segment => (
          <span key={segment.name} className="legend-item">
            <span className="legend-dot" style={{ backgroundColor: segment.color }} />
            {segment.count}
          </span>
        ))}
      </div>
    </div>
  );
}

interface CandidateCardProps {
  candidate: MaturityAnalysisCandidate;
  onViewDetail: (id: EnterpriseCapabilityId) => void;
}

function CandidateCard({ candidate, onViewDetail }: CandidateCardProps) {
  const { getColorForValue, getSectionNameForValue } = useMaturityColorScale();

  const targetSection = candidate.targetMaturity !== null
    ? getSectionNameForValue(candidate.targetMaturity)
    : null;
  const targetColor = candidate.targetMaturity !== null
    ? getColorForValue(candidate.targetMaturity)
    : undefined;

  return (
    <div className="candidate-card">
      <div className="candidate-header">
        <div className="candidate-info">
          <h3 className="candidate-name">{candidate.enterpriseCapabilityName}</h3>
          {candidate.category && (
            <span className="category-badge">{candidate.category}</span>
          )}
        </div>
        <button
          type="button"
          className="btn btn-sm btn-secondary"
          onClick={() => onViewDetail(candidate.enterpriseCapabilityId as EnterpriseCapabilityId)}
        >
          View Details
        </button>
      </div>

      <div className="candidate-stats">
        <div className="stat-group">
          <span className="stat-label">
            Target Maturity
            <HelpTooltip content="Click View Details to set the target maturity level" iconOnly />
          </span>
          {candidate.targetMaturity !== null && targetSection ? (
            <span className="stat-value">
              {candidate.targetMaturity}
              <span className="maturity-section" style={{ color: targetColor }}>
                ({targetSection})
              </span>
            </span>
          ) : (
            <span className="stat-value stat-not-set">Not set</span>
          )}
        </div>
        <div className="stat-group">
          <span className="stat-label">
            Implementations
            <HelpTooltip content="Number of domain capabilities linked to this enterprise capability" iconOnly />
          </span>
          <span className="stat-value">{candidate.implementationCount}</span>
        </div>
        <div className="stat-group">
          <span className="stat-label">
            Domains
            <HelpTooltip content="Number of distinct business domains containing implementations" iconOnly />
          </span>
          <span className="stat-value">{candidate.domainCount}</span>
        </div>
        <div className="stat-group">
          <span className="stat-label">
            Max Gap
            <HelpTooltip content="Largest maturity difference from target among all implementations" iconOnly />
          </span>
          <span className={`stat-value gap-value ${candidate.maxGap > 40 ? 'gap-high' : candidate.maxGap >= 15 ? 'gap-medium' : ''}`}>
            {candidate.maxGap}
          </span>
        </div>
      </div>

      <div className="candidate-maturity-range">
        <span className="range-label">Maturity Range:</span>
        <span className="range-value">
          {candidate.minMaturity} - {candidate.maxMaturity}
        </span>
        <span className="range-avg">(avg: {candidate.averageMaturity})</span>
      </div>

      <MaturityDistributionBar distribution={candidate.maturityDistribution} />
    </div>
  );
}

interface MaturityAnalysisTabProps {
  onViewDetail: (id: EnterpriseCapabilityId) => void;
}

export function MaturityAnalysisTab({ onViewDetail }: MaturityAnalysisTabProps) {
  const [sortBy, setSortBy] = useState<string>('gap');
  const { candidates, summary, isLoading, error } = useMaturityAnalysis(sortBy);

  const handleSortChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setSortBy(e.target.value);
  }, []);

  if (isLoading) {
    return (
      <div className="loading-state">
        <div className="loading-spinner" />
        <span>Loading maturity analysis...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="error-message">
        Failed to load maturity analysis: {error.message}
      </div>
    );
  }

  return (
    <div className="maturity-analysis-tab">
      <div className="analysis-header">
        <div className="analysis-summary">
          {summary && (
            <>
              <div className="summary-stat">
                <span className="summary-value">{summary.candidateCount}</span>
                <span className="summary-label">
                  Capabilities
                  <HelpTooltip
                    content="Enterprise capabilities with linked domain capabilities that can be analyzed for maturity variance"
                    iconOnly
                  />
                </span>
              </div>
              <div className="summary-stat">
                <span className="summary-value">{summary.totalImplementations}</span>
                <span className="summary-label">
                  Implementations
                  <HelpTooltip
                    content="Total domain capabilities linked to these enterprise capabilities"
                    iconOnly
                  />
                </span>
              </div>
              <div className="summary-stat">
                <span className="summary-value">{summary.averageGap}</span>
                <span className="summary-label">
                  Avg Gap
                  <HelpTooltip
                    content="Average difference between implementation maturity and target (or highest implementation if no target set)"
                    iconOnly
                  />
                </span>
              </div>
            </>
          )}
        </div>
        <div className="analysis-controls">
          <label htmlFor="sort-select" className="sort-label">Sort by:</label>
          <select
            id="sort-select"
            value={sortBy}
            onChange={handleSortChange}
            className="sort-select"
          >
            <option value="gap">Max Gap</option>
            <option value="implementations">Implementations</option>
          </select>
        </div>
      </div>

      <MaturitySectionLegend />

      {candidates.length === 0 ? (
        <div className="empty-state">
          <svg className="empty-state-icon" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          <h3 className="empty-state-title">No Enterprise Capabilities</h3>
          <p className="empty-state-description">
            Create enterprise capabilities to set target maturity and analyze gaps.
          </p>
        </div>
      ) : (
        <div className="candidates-grid">
          {candidates.map(candidate => (
            <CandidateCard
              key={candidate.enterpriseCapabilityId}
              candidate={candidate}
              onViewDetail={onViewDetail}
            />
          ))}
        </div>
      )}
    </div>
  );
}
