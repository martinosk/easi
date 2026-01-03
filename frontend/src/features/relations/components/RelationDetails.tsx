import React from 'react';
import { useAppStore } from '../../../store/appStore';
import { AuditHistorySection } from '../../audit';
import { useRelations } from '../hooks/useRelations';
import { useComponents } from '../../components/hooks/useComponents';
import type { Relation, Component } from '../../../api/types';

interface RelationDetailsProps {
  onEdit: () => void;
}

interface RelationData {
  relation: Relation;
  sourceComponent: Component | undefined;
  targetComponent: Component | undefined;
  referenceLink: string | undefined;
  formattedDate: string;
}

const useRelationData = (selectedEdgeId: string | null): RelationData | null => {
  const { data: relations = [] } = useRelations();
  const { data: components = [] } = useComponents();

  if (!selectedEdgeId) {
    return null;
  }

  const relation = relations.find((r) => r.id === selectedEdgeId);

  if (!relation) {
    return null;
  }

  const sourceComponent = components.find((c) => c.id === relation.sourceComponentId);
  const targetComponent = components.find((c) => c.id === relation.targetComponentId);
  const referenceLink = relation._links.reference;
  const formattedDate = new Date(relation.createdAt).toLocaleString();

  return { relation, sourceComponent, targetComponent, referenceLink, formattedDate };
};

export const RelationDetails: React.FC<RelationDetailsProps> = ({ onEdit }) => {
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);

  const data = useRelationData(selectedEdgeId);

  if (!data) {
    return null;
  }

  const { relation, sourceComponent, targetComponent, referenceLink, formattedDate } = data;

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Relation Details</h3>
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

        {referenceLink && (
          <div className="detail-reference">
            <a
              href={referenceLink}
              target="_blank"
              rel="noopener noreferrer"
              className="reference-link"
            >
              <span className="reference-icon">📚</span>
              Reference Documentation
            </a>
          </div>
        )}

        <AuditHistorySection aggregateId={relation.id} />
      </div>
    </div>
  );
};
