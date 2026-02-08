import React from 'react';
import { DetailField } from '../../../components/shared/DetailField';
import { AuditHistorySection } from '../../audit';
import { hasLink } from '../../../utils/hateoas';
import type { AcquiredEntity, OriginRelationship } from '../../../api/types';

interface AcquiredEntityDetailsProps {
  entity: AcquiredEntity;
  relationships: OriginRelationship[];
  canRemoveFromView: boolean;
  onEdit: () => void;
  onRemoveFromView: () => void;
}

const formatDate = (dateString: string | undefined): string => {
  if (!dateString) return 'Not set';
  try {
    return new Date(dateString).toLocaleDateString();
  } catch {
    return dateString;
  }
};

const getIntegrationStatusLabel = (status: string): string => {
  const labels: Record<string, string> = {
    NotStarted: 'Not Started',
    InProgress: 'In Progress',
    Completed: 'Completed',
    OnHold: 'On Hold',
  };
  return labels[status] || status;
};

const getIntegrationStatusColor = (status: string): string => {
  const colors: Record<string, string> = {
    NotStarted: '#6b7280',
    InProgress: '#f59e0b',
    Completed: '#10b981',
    OnHold: '#ef4444',
  };
  return colors[status] || '#6b7280';
};

export const AcquiredEntityDetails: React.FC<AcquiredEntityDetailsProps> = ({
  entity,
  relationships,
  canRemoveFromView,
  onEdit,
  onRemoveFromView,
}) => {
  const canEdit = hasLink(entity, 'edit');
  const formattedCreatedAt = new Date(entity.createdAt).toLocaleString();
  const showActionButtons = canEdit || canRemoveFromView;

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Acquired Entity Details</h3>
      </div>

      <div className="detail-content">
        {showActionButtons && (
          <div className="detail-actions">
            {canEdit && (
              <button className="btn btn-secondary btn-small" onClick={onEdit}>
                Edit
              </button>
            )}
            {canRemoveFromView && (
              <button className="btn btn-secondary btn-small" onClick={onRemoveFromView}>
                Remove from View
              </button>
            )}
          </div>
        )}

        <DetailField label="Name">{entity.name}</DetailField>

        <DetailField label="Acquisition Date">
          {formatDate(entity.acquisitionDate)}
        </DetailField>

        <DetailField label="Integration Status">
          <span
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '6px',
            }}
          >
            <span
              style={{
                width: '8px',
                height: '8px',
                borderRadius: '50%',
                backgroundColor: getIntegrationStatusColor(entity.integrationStatus),
              }}
            />
            {getIntegrationStatusLabel(entity.integrationStatus)}
          </span>
        </DetailField>

        {entity.notes && <DetailField label="Notes">{entity.notes}</DetailField>}

        <DetailField label="Created">
          <span className="detail-date">{formattedCreatedAt}</span>
        </DetailField>

        <DetailField label="Type">Acquired Entity</DetailField>

        {relationships.length > 0 && (
          <DetailField label={`Applications (${relationships.length})`}>
            <ul className="realization-list">
              {relationships.map((rel) => (
                <li key={rel.id} className="realization-item">
                  <span className="realization-name">{rel.componentName}</span>
                  <span className="realization-level">Acquired via</span>
                </li>
              ))}
            </ul>
          </DetailField>
        )}

        <AuditHistorySection aggregateId={entity.id} />
      </div>
    </div>
  );
};
