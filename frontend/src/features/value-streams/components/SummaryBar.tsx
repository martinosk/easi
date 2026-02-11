interface SummaryBarProps {
  stageCount: number;
  capabilityCount: number;
}

export function SummaryBar({ stageCount, capabilityCount }: SummaryBarProps) {
  return (
    <div className="vsd-summary" data-testid="summary-bar">
      <div className="vsd-summary-item">
        <span className="vsd-summary-value">{stageCount}</span>
        <span className="vsd-summary-label">{stageCount === 1 ? 'Stage' : 'Stages'}</span>
      </div>
      <div className="vsd-summary-item">
        <span className="vsd-summary-value">{capabilityCount}</span>
        <span className="vsd-summary-label">{capabilityCount === 1 ? 'Capability' : 'Capabilities'}</span>
      </div>
    </div>
  );
}
