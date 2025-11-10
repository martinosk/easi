import React from 'react';
import { useAppStore } from '../store/appStore';

interface RelationDetailsProps {
  onEdit: () => void;
}

export const RelationDetails: React.FC<RelationDetailsProps> = ({ onEdit }) => {
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const relations = useAppStore((state) => state.relations);
  const components = useAppStore((state) => state.components);
  const clearSelection = useAppStore((state) => state.clearSelection);

  if (!selectedEdgeId) {
    return null;
  }

  const relation = relations.find((r) => r.id === selectedEdgeId);

  if (!relation) {
    return null;
  }

  const sourceComponent = components.find(
    (c) => c.id === relation.sourceComponentId
  );
  const targetComponent = components.find(
    (c) => c.id === relation.targetComponentId
  );

  const archimateLink = relation._links.archimate?.href;
  const formattedDate = new Date(relation.createdAt).toLocaleString();

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Relation Details</h3>
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
        </div>

        {relation.name && (
          <div className="detail-field">
            <label className="detail-label">Name</label>
            <div className="detail-value">{relation.name}</div>
          </div>
        )}

        <div className="detail-field">
          <label className="detail-label">Type</label>
          <div className="detail-value">
            <span className={`relation-type-badge relation-type-${relation.relationType.toLowerCase()}`}>
              {relation.relationType}
            </span>
          </div>
        </div>

        <div className="detail-field">
          <label className="detail-label">Source</label>
          <div className="detail-value">
            {sourceComponent?.name || relation.sourceComponentId}
          </div>
        </div>

        <div className="detail-field">
          <label className="detail-label">Target</label>
          <div className="detail-value">
            {targetComponent?.name || relation.targetComponentId}
          </div>
        </div>

        {relation.description && (
          <div className="detail-field">
            <label className="detail-label">Description</label>
            <div className="detail-value">{relation.description}</div>
          </div>
        )}

        <div className="detail-field">
          <label className="detail-label">Created</label>
          <div className="detail-value detail-date">{formattedDate}</div>
        </div>

        <div className="detail-field">
          <label className="detail-label">ID</label>
          <div className="detail-value detail-id">{relation.id}</div>
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
