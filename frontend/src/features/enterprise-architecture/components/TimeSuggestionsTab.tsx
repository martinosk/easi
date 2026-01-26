import { useMemo, useState, useCallback } from 'react';
import { useTimeSuggestions } from '../hooks/useTimeSuggestions';
import { HelpTooltip } from '../../../components/shared/HelpTooltip';
import type { TimeSuggestion, TimeClassification, TimeSuggestionConfidence } from '../types';
import './TimeSuggestionsTab.css';

const TIME_CLASSIFICATIONS: {
  value: TimeClassification;
  color: string;
  description: string;
}[] = [
  { value: 'Tolerate', color: 'var(--green-500)', description: 'Keep as-is, good fit' },
  { value: 'Invest', color: 'var(--blue-500)', description: 'Enhance technical quality' },
  { value: 'Migrate', color: 'var(--yellow-500)', description: 'Replace functional implementation' },
  { value: 'Eliminate', color: 'var(--red-500)', description: 'Phase out entirely' },
];

function TimeLegend() {
  return (
    <div className="time-legend">
      <span className="time-legend-title">
        TIME Classifications
        <HelpTooltip
          content="Gartner TIME model: Tolerate (good fit), Invest (improve technical), Migrate (replace functional), Eliminate (phase out)"
          iconOnly
        />
      </span>
      <div className="time-legend-items">
        {TIME_CLASSIFICATIONS.map(item => (
          <div key={item.value} className="time-legend-item">
            <span className="time-legend-dot" style={{ backgroundColor: item.color }} />
            <span className="time-legend-name">{item.value}</span>
            <span className="time-legend-description">- {item.description}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function TimeBadge({ time }: { time: TimeClassification | null }) {
  if (!time) {
    return <span className="time-badge unknown">N/A</span>;
  }
  return <span className={`time-badge ${time.toLowerCase()}`}>{time}</span>;
}

function ConfidenceBadge({ confidence }: { confidence: TimeSuggestionConfidence }) {
  return <span className={`confidence-badge ${confidence.toLowerCase()}`}>{confidence}</span>;
}

function GapCell({ gap, label }: { gap: number | null; label: string }) {
  if (gap === null) {
    return <td className="gap-cell"><span className="gap-value gap-none">-</span></td>;
  }
  const sign = gap > 0 ? '+' : '';
  const colorClass = gap > 0 ? 'gap-positive' : gap < 0 ? 'gap-negative' : 'gap-neutral';
  return (
    <td className="gap-cell" title={`${label}: ${sign}${gap.toFixed(1)}`}>
      <span className={`gap-value ${colorClass}`}>{sign}{gap.toFixed(1)}</span>
    </td>
  );
}

function SuggestionRow({ suggestion }: { suggestion: TimeSuggestion }) {
  return (
    <tr>
      <td><div className="entity-cell"><span className="entity-name">{suggestion.capabilityName}</span></div></td>
      <td><div className="entity-cell"><span className="entity-name">{suggestion.componentName}</span></div></td>
      <GapCell gap={suggestion.technicalGap} label="Technical Gap" />
      <GapCell gap={suggestion.functionalGap} label="Functional Gap" />
      <td className="gap-cell"><TimeBadge time={suggestion.suggestedTime} /></td>
      <td className="gap-cell"><ConfidenceBadge confidence={suggestion.confidence} /></td>
    </tr>
  );
}

interface SummaryStats {
  total: number;
  byClassification: Record<TimeClassification | 'Unknown', number>;
  highConfidence: number;
}

function calculateSummary(suggestions: TimeSuggestion[]): SummaryStats {
  const byClassification: Record<TimeClassification | 'Unknown', number> = {
    Tolerate: 0, Invest: 0, Migrate: 0, Eliminate: 0, Unknown: 0,
  };
  let highConfidence = 0;
  for (const s of suggestions) {
    if (s.suggestedTime) byClassification[s.suggestedTime]++;
    else byClassification.Unknown++;
    if (s.confidence === 'High') highConfidence++;
  }
  return { total: suggestions.length, byClassification, highConfidence };
}

type GroupBy = 'none' | 'capability' | 'component';

interface HeaderProps {
  summary: SummaryStats;
  groupBy: GroupBy;
  onGroupByChange: (e: React.ChangeEvent<HTMLSelectElement>) => void;
}

function TimeSuggestionsHeader({ summary, groupBy, onGroupByChange }: HeaderProps) {
  return (
    <div className="time-header">
      <div className="time-summary">
        <div className="time-summary-stat">
          <span className="time-summary-value">{summary.total}</span>
          <span className="time-summary-label">
            Total Realizations
            <HelpTooltip content="Component-capability combinations with both strategic importance and fit scores" iconOnly />
          </span>
        </div>
        <div className="time-summary-stat">
          <span className="time-summary-value">{summary.highConfidence}</span>
          <span className="time-summary-label">
            High Confidence
            <HelpTooltip content="Suggestions with both technical and functional gaps available" iconOnly />
          </span>
        </div>
        <div className="time-summary-stat">
          <span className="time-summary-value">{summary.byClassification.Eliminate}</span>
          <span className="time-summary-label">
            Eliminate
            <HelpTooltip content="Components suggested for phase-out due to both technical and functional gaps" iconOnly />
          </span>
        </div>
      </div>
      <div className="time-filters">
        <label htmlFor="group-select" className="time-filter-label">Group by:</label>
        <select id="group-select" value={groupBy} onChange={onGroupByChange} className="time-filter-select">
          <option value="none">No grouping</option>
          <option value="capability">Enterprise Capability</option>
          <option value="component">Component</option>
        </select>
      </div>
    </div>
  );
}

function TimeSuggestionsEmptyState() {
  return (
    <div className="empty-state">
      <svg className="empty-state-icon" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
      </svg>
      <h3 className="empty-state-title">No TIME Suggestions Available</h3>
      <p className="empty-state-description">
        TIME suggestions require:<br />
        • Enterprise capabilities with strategic importance configured<br />
        • Components with fit scores<br />
        • Strategy pillars with fit types (Technical/Functional) enabled
      </p>
    </div>
  );
}

function SuggestionsTable({ suggestions }: { suggestions: TimeSuggestion[] }) {
  return (
    <div className="suggestions-table-container">
      <table className="suggestions-table">
        <thead>
          <tr>
            <th>Capability</th>
            <th>Component</th>
            <th className="text-center">
              Technical Gap
              <HelpTooltip content="Difference between strategic importance and fit score for technical pillars. Positive = underperforming" iconOnly />
            </th>
            <th className="text-center">
              Functional Gap
              <HelpTooltip content="Difference between strategic importance and fit score for functional pillars. Positive = underperforming" iconOnly />
            </th>
            <th className="text-center">
              Suggested TIME
              <HelpTooltip content="Recommended action based on technical and functional gap analysis" iconOnly />
            </th>
            <th className="text-center">
              Confidence
              <HelpTooltip content="High: both gaps available, Medium: one gap available, Low/Insufficient: limited data" iconOnly />
            </th>
          </tr>
        </thead>
        <tbody>
          {suggestions.map(s => (
            <SuggestionRow key={`${s.capabilityId}-${s.componentId}`} suggestion={s} />
          ))}
        </tbody>
      </table>
    </div>
  );
}

export function TimeSuggestionsTab() {
  const [groupBy, setGroupBy] = useState<GroupBy>('none');
  const { suggestions, isLoading, error } = useTimeSuggestions();

  const summary = useMemo(() => calculateSummary(suggestions), [suggestions]);

  const sortedSuggestions = useMemo(() => {
    const sorted = [...suggestions];
    if (groupBy === 'capability') {
      sorted.sort((a, b) => a.capabilityName.localeCompare(b.capabilityName));
    } else if (groupBy === 'component') {
      sorted.sort((a, b) => a.componentName.localeCompare(b.componentName));
    }
    return sorted;
  }, [suggestions, groupBy]);

  const handleGroupByChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setGroupBy(e.target.value as GroupBy);
  }, []);

  if (isLoading) {
    return (
      <div className="loading-state">
        <div className="loading-spinner" />
        <span>Loading TIME suggestions...</span>
      </div>
    );
  }

  if (error) {
    return <div className="error-message">Failed to load TIME suggestions: {error.message}</div>;
  }

  return (
    <div className="time-suggestions-tab">
      <TimeSuggestionsHeader summary={summary} groupBy={groupBy} onGroupByChange={handleGroupByChange} />
      <TimeLegend />
      {suggestions.length === 0 ? <TimeSuggestionsEmptyState /> : <SuggestionsTable suggestions={sortedSuggestions} />}
    </div>
  );
}
