import React from 'react';
import { useAppStore } from '../store/appStore';

export const ComponentDetails: React.FC = () => {
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const components = useAppStore((state) => state.components);
  const clearSelection = useAppStore((state) => state.clearSelection);

  if (!selectedNodeId) {
    return null;
  }

  const component = components.find((c) => c.id === selectedNodeId);

  if (!component) {
    return null;
  }

  const archimateLink = component._links.archimate?.href;
  const formattedDate = new Date(component.createdAt).toLocaleString();

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
              <span className="archimate-icon">📚</span>
              ArchiMate Documentation
            </a>
          </div>
        )}
      </div>
    </div>
  );
};
