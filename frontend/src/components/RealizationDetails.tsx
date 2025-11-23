import React, { useState } from 'react';
import { useAppStore } from '../store/appStore';
import { EditRealizationDialog } from './EditRealizationDialog';
import { DetailField } from './DetailField';

export const RealizationDetails: React.FC = () => {
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const capabilityRealizations = useAppStore((state) => state.capabilityRealizations);
  const capabilities = useAppStore((state) => state.capabilities);
  const components = useAppStore((state) => state.components);
  const selectEdge = useAppStore((state) => state.selectEdge);

  const [showEditDialog, setShowEditDialog] = useState(false);

  if (!selectedEdgeId || !selectedEdgeId.startsWith('realization-')) {
    return null;
  }

  const realizationId = selectedEdgeId.replace('realization-', '');
  const realization = capabilityRealizations.find((r) => r.id === realizationId);

  if (!realization) {
    return null;
  }

  const capability = capabilities.find((c) => c.id === realization.capabilityId);
  const component = components.find((c) => c.id === realization.componentId);
  const formattedDate = new Date(realization.linkedAt).toLocaleString();
  const isInherited = realization.origin === 'Inherited';

  const getLevelDisplay = (level: string): string => {
    switch (level) {
      case 'Full':
        return 'Full (100%)';
      case 'Partial':
        return 'Partial';
      case 'Planned':
        return 'Planned';
      default:
        return level;
    }
  };

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Realization Details</h3>
        <button
          className="detail-close"
          onClick={() => selectEdge(null)}
          aria-label="Close details"
        >
          x
        </button>
      </div>

      <div className="detail-content">
        {!isInherited && (
          <div className="detail-actions">
            <button
              className="btn btn-secondary btn-small"
              onClick={() => setShowEditDialog(true)}
            >
              Edit
            </button>
          </div>
        )}

        <DetailField label="Capability">{capability?.name || 'Unknown'}</DetailField>
        <DetailField label="Application">{component?.name || 'Unknown'}</DetailField>
        <DetailField label="Realization Level">
          <span className={`level-badge level-${realization.realizationLevel.toLowerCase()}`}>
            {getLevelDisplay(realization.realizationLevel)}
          </span>
        </DetailField>
        <DetailField label="Origin">
          <span className={`origin-badge ${isInherited ? 'origin-inherited' : 'origin-direct'}`}>
            {realization.origin}
          </span>
        </DetailField>
        {realization.notes && (
          <DetailField label="Notes">{realization.notes}</DetailField>
        )}
        <DetailField label="Linked">
          <span className="detail-date">{formattedDate}</span>
        </DetailField>
        <DetailField label="ID">
          <span className="detail-id">{realization.id}</span>
        </DetailField>

        {isInherited && (
          <div className="detail-info">
            This is an inherited realization. It was automatically created when
            an application was linked to a child capability. To edit or delete,
            modify the original direct realization.
          </div>
        )}
      </div>

      <EditRealizationDialog
        isOpen={showEditDialog}
        onClose={() => setShowEditDialog(false)}
        realization={realization}
      />
    </div>
  );
};
