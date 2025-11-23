import React from 'react';
import { useAppStore } from '../store/appStore';
import type { CapabilityRealization, Capability } from '../api/types';

interface ComponentDetailsProps {
  onEdit: () => void;
  onRemoveFromView?: () => void;
}

const getLevelBadge = (level: string): string => {
  const badges: Record<string, string> = {
    Full: '100%',
    Partial: 'Partial',
    Planned: 'Planned',
  };
  return badges[level] || level;
};

const getCapabilityName = (capabilities: Capability[], capabilityId: string): string => {
  const cap = capabilities.find((c) => c.id === capabilityId);
  return cap ? `${cap.level}: ${cap.name}` : 'Unknown';
};

interface RealizationListProps {
  realizations: CapabilityRealization[];
  capabilities: Capability[];
  origin: 'Direct' | 'Inherited';
}

const RealizationListItems: React.FC<RealizationListProps> = ({ realizations, capabilities, origin }) => (
  <>
    {realizations.map((r) => (
      <li key={r.id} className={`realization-item${origin === 'Inherited' ? ' inherited' : ''}`}>
        <span className="realization-name">{getCapabilityName(capabilities, r.capabilityId)}</span>
        <span className="realization-level">{getLevelBadge(r.realizationLevel)}</span>
        <span className={`realization-origin origin-${origin.toLowerCase()}`}>{origin.toLowerCase()}</span>
      </li>
    ))}
  </>
);

export const ComponentDetails: React.FC<ComponentDetailsProps> = ({ onEdit, onRemoveFromView }) => {
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const components = useAppStore((state) => state.components);
  const currentView = useAppStore((state) => state.currentView);
  const clearSelection = useAppStore((state) => state.clearSelection);
  const capabilityRealizations = useAppStore((state) => state.capabilityRealizations);
  const capabilities = useAppStore((state) => state.capabilities);

  const component = components.find((c) => c.id === selectedNodeId);
  if (!selectedNodeId || !component) {
    return null;
  }

  const isInCurrentView = currentView?.components.some(
    (vc) => vc.componentId === selectedNodeId
  );

  const archimateLink = component._links.archimate?.href;
  const formattedDate = new Date(component.createdAt).toLocaleString();

  const componentRealizations = capabilityRealizations.filter(
    (r) => r.componentId === component.id
  );

  const directRealizations = componentRealizations.filter((r) => r.origin === 'Direct');
  const inheritedRealizations = componentRealizations.filter((r) => r.origin === 'Inherited');

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Component Details</h3>
        <button
          className="detail-close"
          onClick={clearSelection}
          aria-label="Close details"
        >
          ×
        </button>
      </div>

      <div className="detail-content">
        <div className="detail-actions">
          <button className="btn btn-secondary btn-small" onClick={onEdit}>
            Edit
          </button>
          {isInCurrentView && onRemoveFromView && (
            <button className="btn btn-secondary btn-small" onClick={onRemoveFromView}>
              Remove from View
            </button>
          )}
        </div>

        <div className="detail-field">
          <label className="detail-label">Name</label>
          <div className="detail-value">{component.name}</div>
        </div>

        {component.description && (
          <div className="detail-field">
            <label className="detail-label">Description</label>
            <div className="detail-value">{component.description}</div>
          </div>
        )}

        <div className="detail-field">
          <label className="detail-label">Created</label>
          <div className="detail-value detail-date">{formattedDate}</div>
        </div>

        <div className="detail-field">
          <label className="detail-label">Type</label>
          <div className="detail-value">Application Component</div>
        </div>

        <div className="detail-field">
          <label className="detail-label">ID</label>
          <div className="detail-value detail-id">{component.id}</div>
        </div>

        {archimateLink && (
          <div className="detail-archimate">
            <a
              href={archimateLink}
              target="_blank"
              rel="noopener noreferrer"
              className="archimate-link"
            >
              ArchiMate Documentation
            </a>
          </div>
        )}

        {componentRealizations.length > 0 && (
          <div className="detail-field">
            <label className="detail-label">Realizes Capabilities</label>
            <ul className="realization-list">
              <RealizationListItems realizations={directRealizations} capabilities={capabilities} origin="Direct" />
              <RealizationListItems realizations={inheritedRealizations} capabilities={capabilities} origin="Inherited" />
            </ul>
          </div>
        )}
      </div>
    </div>
  );
};
