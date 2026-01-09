import { useState, useMemo } from 'react';
import { useStrategicFitAnalysis } from '../hooks/useStrategicFitAnalysis';
import { useStrategyPillarsConfig } from '../../../hooks/useStrategyPillarsSettings';
import type { RealizationFit, StrategicFitSummary, ApiError } from '../../../api/types';
import './StrategicFitTab.css';

const SCORE_RANGE = [1, 2, 3, 4, 5] as const;

function getAnalysisErrorMessage(error: unknown): string {
  if (error instanceof Error && 'statusCode' in error) {
    const apiError = error as ApiError;
    switch (apiError.statusCode) {
      case 400:
        return 'Fit scoring is not enabled for this pillar.';
      case 403:
        return 'You do not have permission to view strategic fit analysis.';
      case 404:
        return 'Strategy pillar not found.';
      default:
        return apiError.message || 'Failed to load analysis';
    }
  }
  return error instanceof Error ? error.message : 'Failed to load analysis';
}

function getCategoryColor(category: 'liability' | 'concern' | 'aligned'): string {
  switch (category) {
    case 'liability':
      return 'var(--red-600)';
    case 'concern':
      return 'var(--yellow-600)';
    case 'aligned':
      return 'var(--green-600)';
    default:
      return 'var(--color-gray-600)';
  }
}

interface SummaryCardProps {
  summary: StrategicFitSummary;
}

function SummaryCard({ summary }: SummaryCardProps) {
  return (
    <div className="fit-summary-card">
      <div className="summary-stats">
        <div className="summary-stat liability">
          <span className="stat-value">{summary.liabilityCount}</span>
          <span className="stat-label">Liabilities</span>
        </div>
        <div className="summary-stat concern">
          <span className="stat-value">{summary.concernCount}</span>
          <span className="stat-label">Concerns</span>
        </div>
        <div className="summary-stat aligned">
          <span className="stat-value">{summary.alignedCount}</span>
          <span className="stat-label">Aligned</span>
        </div>
      </div>
      <div className="summary-meta">
        <span className="meta-item">
          {summary.scoredRealizations} of {summary.totalRealizations} realizations scored
        </span>
        {summary.averageGap > 0 && (
          <span className="meta-item">
            Average gap: {summary.averageGap.toFixed(1)}
          </span>
        )}
      </div>
    </div>
  );
}

interface RealizationFitCardProps {
  realization: RealizationFit;
}

function RealizationFitCard({ realization }: RealizationFitCardProps) {
  return (
    <div className={`realization-fit-card category-${realization.category}`}>
      <div className="fit-card-header">
        <div className="fit-card-names">
          <span className="component-name">{realization.componentName}</span>
          <span className="arrow-separator">→</span>
          <span className="capability-name">{realization.capabilityName}</span>
        </div>
        {realization.businessDomainName && (
          <span className="domain-badge">{realization.businessDomainName}</span>
        )}
      </div>
      <div className="fit-card-scores">
        <div className="score-item">
          <span className="score-label">
            Importance
            {realization.isImportanceInherited && realization.importanceSourceCapabilityName && (
              <span className="inherited-indicator" title={`Inherited from ${realization.importanceSourceCapabilityName}`}>
                {' '}(from {realization.importanceSourceCapabilityName})
              </span>
            )}
          </span>
          <span className="score-value importance">
            {SCORE_RANGE.map((i) => (
              <span
                key={i}
                className={`score-star ${i <= realization.importance ? 'filled' : ''}`}
              >
                ★
              </span>
            ))}
            <span className="score-number">({realization.importance})</span>
          </span>
        </div>
        <div className="score-item">
          <span className="score-label">Fit</span>
          <span className="score-value fit">
            {SCORE_RANGE.map((i) => (
              <span
                key={i}
                className={`score-dot ${i <= realization.fitScore ? 'filled' : ''}`}
              />
            ))}
            <span className="score-number">({realization.fitScore})</span>
          </span>
        </div>
        <div className="score-item">
          <span className="score-label">Gap</span>
          <span className={`score-value gap gap-${realization.gap >= 2 ? 'high' : realization.gap === 1 ? 'medium' : 'low'}`}>
            {realization.gap}
          </span>
        </div>
      </div>
      {realization.fitRationale && (
        <div className="fit-rationale">"{realization.fitRationale}"</div>
      )}
    </div>
  );
}

interface RealizationSectionProps {
  title: string;
  realizations: RealizationFit[];
  defaultExpanded?: boolean;
  category: 'liability' | 'concern' | 'aligned';
}

function RealizationSection({ title, realizations, defaultExpanded = false, category }: RealizationSectionProps) {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);

  if (realizations.length === 0) return null;

  return (
    <div className={`realization-section section-${category}`}>
      <button
        type="button"
        className="section-header"
        onClick={() => setIsExpanded(!isExpanded)}
        aria-expanded={isExpanded}
        aria-label={`${title}, ${realizations.length} items`}
      >
        <span className="section-title" style={{ color: getCategoryColor(category) }}>
          {title} ({realizations.length})
        </span>
        <span className={`expand-icon ${isExpanded ? 'expanded' : ''}`} aria-hidden="true">▸</span>
      </button>
      {isExpanded && (
        <div className="section-content">
          {realizations.map((r) => (
            <RealizationFitCard key={r.realizationId} realization={r} />
          ))}
        </div>
      )}
    </div>
  );
}

export function StrategicFitTab() {
  const { data: pillarsConfig, isLoading: pillarsLoading } = useStrategyPillarsConfig();
  const [selectedPillarId, setSelectedPillarId] = useState<string | null>(null);
  const { data: analysis, isLoading: analysisLoading, error } = useStrategicFitAnalysis(selectedPillarId);

  const enabledPillars = useMemo(() => {
    if (!pillarsConfig?.data) return [];
    return pillarsConfig.data.filter((p) => p.active && p.fitScoringEnabled);
  }, [pillarsConfig]);

  const handlePillarChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedPillarId(e.target.value || null);
  };

  if (pillarsLoading) {
    return (
      <div className="loading-state">
        <div className="loading-spinner" />
        <span>Loading pillars...</span>
      </div>
    );
  }

  if (enabledPillars.length === 0) {
    return (
      <div className="strategic-fit-tab">
        <div className="empty-state">
          <svg className="empty-state-icon" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
          <h3 className="empty-state-title">No Pillars with Fit Scoring</h3>
          <p className="empty-state-description">
            Enable fit scoring for strategy pillars in Settings to analyze strategic alignment.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="strategic-fit-tab">
      <div className="fit-header">
        <div className="fit-description">
          <h3>Strategic Fit Analysis</h3>
          <p>Identify realizations where application fit does not match strategic importance</p>
        </div>
        <div className="pillar-selector">
          <label htmlFor="pillar-select">Filter by pillar:</label>
          <select
            id="pillar-select"
            value={selectedPillarId || ''}
            onChange={handlePillarChange}
            className="pillar-select"
            aria-label="Select strategy pillar for fit analysis"
          >
            <option value="">Select a pillar</option>
            {enabledPillars.map((pillar) => (
              <option key={pillar.id} value={pillar.id}>
                {pillar.name}
              </option>
            ))}
          </select>
        </div>
      </div>

      {!selectedPillarId ? (
        <div className="select-pillar-prompt">
          <p>Select a strategy pillar to view the fit analysis</p>
        </div>
      ) : analysisLoading ? (
        <div className="loading-state">
          <div className="loading-spinner" />
          <span>Loading analysis...</span>
        </div>
      ) : error ? (
        <div className="error-message">
          {getAnalysisErrorMessage(error)}
        </div>
      ) : analysis ? (
        <div className="fit-analysis-content">
          <SummaryCard summary={analysis.summary} />

          <RealizationSection
            title="Strategic Liabilities"
            realizations={analysis.liabilities}
            defaultExpanded={true}
            category="liability"
          />

          <RealizationSection
            title="Concerns"
            realizations={analysis.concerns}
            defaultExpanded={analysis.liabilities.length === 0}
            category="concern"
          />

          <RealizationSection
            title="Well Aligned"
            realizations={analysis.aligned}
            defaultExpanded={false}
            category="aligned"
          />
        </div>
      ) : null}
    </div>
  );
}
