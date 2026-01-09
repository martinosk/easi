import { useState, useCallback } from 'react';
import { useMaturityGapDetailHook, useSetTargetMaturity } from '../hooks/useMaturityAnalysis';
import { useMaturityColorScale } from '../../../hooks/useMaturityColorScale';
import { HelpTooltip } from '../../../components/shared/HelpTooltip';
import type { EnterpriseCapabilityId, ImplementationDetail } from '../types';
import './MaturityGapDetailPanel.css';

function getPriorityColor(priority: string): string {
  switch (priority) {
    case 'High':
      return 'var(--color-error, #ef4444)';
    case 'Medium':
      return 'var(--color-warning, #f59e0b)';
    case 'Low':
      return 'var(--color-blue-500, #3b82f6)';
    default:
      return 'var(--color-green-500, #22c55e)';
  }
}

interface ImplementationBarProps {
  implementation: ImplementationDetail;
  targetMaturity: number;
  getColorForValue: (value: number) => string;
}

function ImplementationBar({ implementation, targetMaturity, getColorForValue }: ImplementationBarProps) {
  const percentage = implementation.maturityValue;
  const targetPercentage = targetMaturity;

  return (
    <div className="implementation-bar-container">
      <div className="implementation-info">
        <span className="impl-name">{implementation.domainCapabilityName}</span>
        {implementation.businessDomainName && (
          <span className="impl-domain">{implementation.businessDomainName}</span>
        )}
      </div>
      <div className="bar-wrapper">
        <div className="maturity-bar">
          <div
            className="maturity-fill"
            style={{
              width: `${percentage}%`,
              backgroundColor: getColorForValue(implementation.maturityValue),
            }}
          />
          <div
            className="target-marker"
            style={{ left: `${targetPercentage}%` }}
            title="Target maturity level"
          />
        </div>
        <div className="bar-labels">
          <span className="maturity-value">{implementation.maturityValue}</span>
          <span
            className="gap-badge"
            style={{ color: getPriorityColor(implementation.priority) }}
          >
            {implementation.gap > 0 ? `-${implementation.gap}` : 'On Target'}
          </span>
        </div>
      </div>
    </div>
  );
}

interface PrioritySectionProps {
  title: string;
  priority: string;
  implementations: ImplementationDetail[];
  targetMaturity: number;
  tooltip: string;
  getColorForValue: (value: number) => string;
}

function PrioritySection({ title, priority, implementations, targetMaturity, tooltip, getColorForValue }: PrioritySectionProps) {
  if (implementations.length === 0) return null;

  return (
    <div className="priority-section">
      <div className="priority-header" style={{ borderLeftColor: getPriorityColor(priority) }}>
        <h4 className="priority-title">
          {title}
          <HelpTooltip content={tooltip} iconOnly />
        </h4>
        <span className="priority-count">{implementations.length}</span>
      </div>
      <div className="priority-implementations">
        {implementations.map(impl => (
          <ImplementationBar
            key={impl.domainCapabilityId}
            implementation={impl}
            targetMaturity={targetMaturity}
            getColorForValue={getColorForValue}
          />
        ))}
      </div>
    </div>
  );
}

interface SetTargetMaturityModalProps {
  isOpen: boolean;
  currentValue: number | null;
  onClose: () => void;
  onSave: (value: number) => void;
  isSaving: boolean;
  getColorForValue: (value: number) => string;
  getSectionNameForValue: (value: number) => string;
  bounds: { min: number; max: number };
}

function SetTargetMaturityModal({ isOpen, currentValue, onClose, onSave, isSaving, getColorForValue, getSectionNameForValue, bounds }: SetTargetMaturityModalProps) {
  const [value, setValue] = useState<number>(currentValue ?? Math.floor((bounds.min + bounds.max) / 2));

  const handleSubmit = useCallback((e: React.FormEvent) => {
    e.preventDefault();
    onSave(value);
  }, [value, onSave]);

  if (!isOpen) return null;

  const section = getSectionNameForValue(value);

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        <h3 className="modal-title">Set Target Maturity</h3>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="target-maturity-input" className="form-label">
              Target Maturity Value
            </label>
            <div className="slider-container">
              <input
                type="range"
                id="target-maturity-slider"
                min={bounds.min}
                max={bounds.max}
                value={value}
                onChange={(e) => setValue(Number(e.target.value))}
                className="slider"
              />
              <div className="slider-value-display">
                <span className="slider-value">{value}</span>
                <span
                  className="slider-section"
                  style={{ color: getColorForValue(value) }}
                >
                  {section}
                </span>
              </div>
            </div>
            <input
              type="number"
              id="target-maturity-input"
              min={bounds.min}
              max={bounds.max}
              value={value}
              onChange={(e) => setValue(Math.min(bounds.max, Math.max(bounds.min, Number(e.target.value))))}
              className="form-input"
            />
          </div>
          <div className="modal-actions">
            <button type="button" className="btn btn-secondary" onClick={onClose} disabled={isSaving}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={isSaving}>
              {isSaving ? 'Saving...' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

interface MaturityGapDetailPanelProps {
  enterpriseCapabilityId: EnterpriseCapabilityId;
  onBack: () => void;
}

export function MaturityGapDetailPanel({ enterpriseCapabilityId, onBack }: MaturityGapDetailPanelProps) {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const { detail, isLoading, error } = useMaturityGapDetailHook(enterpriseCapabilityId);
  const setTargetMaturityMutation = useSetTargetMaturity();
  const { getColorForValue, getSectionNameForValue, bounds } = useMaturityColorScale();

  const handleOpenModal = useCallback(() => {
    setIsModalOpen(true);
  }, []);

  const handleCloseModal = useCallback(() => {
    setIsModalOpen(false);
  }, []);

  const handleSaveTargetMaturity = useCallback(async (value: number) => {
    await setTargetMaturityMutation.mutateAsync({
      enterpriseCapabilityId,
      targetMaturity: value,
    });
    setIsModalOpen(false);
  }, [enterpriseCapabilityId, setTargetMaturityMutation]);

  if (isLoading) {
    return (
      <div className="maturity-gap-detail-panel">
        <div className="loading-state">
          <div className="loading-spinner" />
          <span>Loading details...</span>
        </div>
      </div>
    );
  }

  if (error || !detail) {
    return (
      <div className="maturity-gap-detail-panel">
        <button type="button" className="back-button" onClick={onBack}>
          ← Back to Analysis
        </button>
        <div className="error-message">
          {error ? `Failed to load details: ${error.message}` : 'Capability not found'}
        </div>
      </div>
    );
  }

  const targetMaturity = detail.targetMaturity ?? Math.max(...detail.implementations.map(i => i.maturityValue));
  const targetSection = detail.targetMaturity !== null
    ? getSectionNameForValue(detail.targetMaturity)
    : null;

  return (
    <div className="maturity-gap-detail-panel">
      <button type="button" className="back-button" onClick={onBack}>
        ← Back to Analysis
      </button>

      <div className="detail-header">
        <div className="header-info">
          <h2 className="detail-title">{detail.enterpriseCapabilityName}</h2>
          {detail.category && <span className="category-badge">{detail.category}</span>}
        </div>
      </div>

      <div className="target-maturity-section">
        <div className="target-info">
          <span className="target-label">Target Maturity:</span>
          {detail.targetMaturity !== null && targetSection ? (
            <span className="target-value">
              {detail.targetMaturity}
              <span
                className="target-section"
                style={{ color: getColorForValue(detail.targetMaturity) }}
              >
                ({targetSection})
              </span>
            </span>
          ) : (
            <span className="target-not-set">Not set (using max: {targetMaturity})</span>
          )}
        </div>
        <button
          type="button"
          className="btn btn-sm btn-secondary"
          onClick={handleOpenModal}
        >
          {detail.targetMaturity !== null ? 'Edit Target' : 'Set Target'}
        </button>
      </div>

      <div className="implementations-section">
        <h3 className="section-title">
          Implementations ({detail.implementations.length})
          <HelpTooltip
            content="Each bar shows current maturity level. The vertical line marks the target. Gap is the difference between current and target maturity."
            iconOnly
          />
        </h3>

        <PrioritySection
          title="High Priority (Gap > 40)"
          priority="High"
          implementations={detail.investmentPriorities.high}
          targetMaturity={targetMaturity}
          tooltip="Implementations that need significant work to reach the target"
          getColorForValue={getColorForValue}
        />

        <PrioritySection
          title="Medium Priority (Gap 15-40)"
          priority="Medium"
          implementations={detail.investmentPriorities.medium}
          targetMaturity={targetMaturity}
          tooltip="Implementations that need moderate work to reach the target"
          getColorForValue={getColorForValue}
        />

        <PrioritySection
          title="Low Priority (Gap 1-14)"
          priority="Low"
          implementations={detail.investmentPriorities.low}
          targetMaturity={targetMaturity}
          tooltip="Implementations that need minor work to reach the target"
          getColorForValue={getColorForValue}
        />

        <PrioritySection
          title="On Target"
          priority="None"
          implementations={detail.investmentPriorities.onTarget}
          targetMaturity={targetMaturity}
          tooltip="Implementations that meet or exceed the target maturity level"
          getColorForValue={getColorForValue}
        />
      </div>

      <SetTargetMaturityModal
        isOpen={isModalOpen}
        currentValue={detail.targetMaturity}
        onClose={handleCloseModal}
        onSave={handleSaveTargetMaturity}
        isSaving={setTargetMaturityMutation.isPending}
        getColorForValue={getColorForValue}
        getSectionNameForValue={getSectionNameForValue}
        bounds={bounds}
      />
    </div>
  );
}
