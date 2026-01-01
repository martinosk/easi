import { useState, useCallback } from 'react';
import { useMaturityAnalysis } from '../hooks/useMaturityAnalysis';
import type { MaturityAnalysisCandidate, EnterpriseCapabilityId } from '../types';
import './MaturityAnalysisTab.css';

function getMaturitySectionColor(section: string): string {
  switch (section) {
    case 'Genesis':
      return 'var(--color-purple-500, #8b5cf6)';
    case 'Custom Build':
      return 'var(--color-blue-500, #3b82f6)';
    case 'Product':
      return 'var(--color-green-500, #22c55e)';
    case 'Commodity':
      return 'var(--color-gray-500, #6b7280)';
    default:
      return 'var(--color-gray-400)';
  }
}

function getMaturitySection(value: number): string {
  if (value <= 24) return 'Genesis';
  if (value <= 49) return 'Custom Build';
  if (value <= 74) return 'Product';
  return 'Commodity';
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
  const total = distribution.genesis + distribution.customBuild + distribution.product + distribution.commodity;
  if (total === 0) return null;

  const segments = [
    { name: 'Genesis', count: distribution.genesis, color: getMaturitySectionColor('Genesis') },
    { name: 'Custom Build', count: distribution.customBuild, color: getMaturitySectionColor('Custom Build') },
    { name: 'Product', count: distribution.product, color: getMaturitySectionColor('Product') },
    { name: 'Commodity', count: distribution.commodity, color: getMaturitySectionColor('Commodity') },
  ].filter(s => s.count > 0);

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
  const targetSection = candidate.targetMaturity !== null
    ? getMaturitySection(candidate.targetMaturity)
    : null;

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
          <span className="stat-label">Target Maturity</span>
          {candidate.targetMaturity !== null && targetSection ? (
            <span className="stat-value">
              {candidate.targetMaturity}
              <span className="maturity-section" style={{ color: getMaturitySectionColor(targetSection) }}>
                ({targetSection})
              </span>
            </span>
          ) : (
            <span className="stat-value stat-not-set">Not set</span>
          )}
        </div>
        <div className="stat-group">
          <span className="stat-label">Implementations</span>
          <span className="stat-value">{candidate.implementationCount}</span>
        </div>
        <div className="stat-group">
          <span className="stat-label">Domains</span>
          <span className="stat-value">{candidate.domainCount}</span>
        </div>
        <div className="stat-group">
          <span className="stat-label">Max Gap</span>
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
                <span className="summary-label">Capabilities</span>
              </div>
              <div className="summary-stat">
                <span className="summary-value">{summary.totalImplementations}</span>
                <span className="summary-label">Implementations</span>
              </div>
              <div className="summary-stat">
                <span className="summary-value">{summary.averageGap}</span>
                <span className="summary-label">Avg Gap</span>
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
