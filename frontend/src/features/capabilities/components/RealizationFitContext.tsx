import { useFitComparisons } from '../../components/hooks/useFitScores';
import type { ComponentId, CapabilityId, BusinessDomainId, FitComparison, FitCategory } from '../../../api/types';
import './RealizationFitContext.css';

interface FitComparisonDisplayProps {
  comparison: FitComparison;
}

function getCategoryLabel(category: FitCategory, gap: number): string {
  switch (category) {
    case 'liability':
      return 'LIABILITY';
    case 'concern':
      return 'Minor';
    case 'aligned':
      return 'OK';
    default:
      return `Gap ${gap}`;
  }
}

function FitComparisonDisplay({ comparison }: FitComparisonDisplayProps) {
  return (
    <div className={`fit-comparison fit-${comparison.category}`}>
      <span className="pillar-label">{comparison.pillarName}:</span>
      <span className="fit-values">
        Fit <span className="fit-value">{comparison.fitScore}</span> vs Imp{' '}
        <span className="importance-value">{comparison.importance}</span>
      </span>
      <span className="gap-indicator">
        → Gap {comparison.gap}{' '}
        <span className={`gap-label gap-${comparison.category}`}>
          ({getCategoryLabel(comparison.category, comparison.gap)})
        </span>
      </span>
    </div>
  );
}

interface RealizationFitContextProps {
  componentId: ComponentId;
  capabilityId: CapabilityId;
  businessDomainId: BusinessDomainId;
}

export function RealizationFitContext({
  componentId,
  capabilityId,
  businessDomainId,
}: RealizationFitContextProps) {
  const { data: comparisons = [] } = useFitComparisons(componentId, capabilityId, businessDomainId);

  const validComparisons = comparisons.filter((c) => c.importance > 0 && c.fitScore > 0);

  if (validComparisons.length === 0) return null;

  return (
    <div className="realization-fit-context">
      <div className="fit-context-label">Fit vs Importance:</div>
      <div className="fit-comparisons">
        {validComparisons.map((c) => (
          <FitComparisonDisplay key={c.pillarId} comparison={c} />
        ))}
      </div>
    </div>
  );
}
