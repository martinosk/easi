import React, { useState } from 'react';
import { useAppStore } from '../store/appStore';
import { EditCapabilityDialog } from './EditCapabilityDialog';
import { DeleteCapabilityDialog } from './DeleteCapabilityDialog';
import { DetailField } from './DetailField';

interface CapabilityDetailsProps {
  onRemoveFromCanvas: () => void;
}

const getMaturityBadgeClass = (maturityLevel?: string): string => {
  const level = maturityLevel?.toLowerCase();
  const maturityClasses: Record<string, string> = {
    'genesis': 'badge-genesis',
    'custom build': 'badge-custom-build',
    'product': 'badge-product',
    'commodity': 'badge-commodity',
  };
  return maturityClasses[level || ''] || 'badge-default';
};

export const CapabilityDetails: React.FC<CapabilityDetailsProps> = ({ onRemoveFromCanvas }) => {
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const capabilities = useAppStore((state) => state.capabilities);
  const selectCapability = useAppStore((state) => state.selectCapability);
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);

  const capability = capabilities.find((c) => c.id === selectedCapabilityId);
  if (!selectedCapabilityId || !capability) return null;

  const formattedDate = new Date(capability.createdAt).toLocaleString();

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Capability Details</h3>
        <button className="detail-close" onClick={() => selectCapability(null)} aria-label="Close details">x</button>
      </div>

      <div className="detail-content">
        <div className="detail-actions">
          <button className="btn btn-secondary btn-small" onClick={() => setShowEditDialog(true)}>Edit</button>
          <button className="btn btn-secondary btn-small" onClick={onRemoveFromCanvas}>Remove from Canvas</button>
          <button className="btn btn-danger btn-small" onClick={() => setShowDeleteDialog(true)}>Delete</button>
        </div>

        <DetailField label="Name">{capability.name}</DetailField>
        <DetailField label="Level"><span className="level-badge">{capability.level}</span></DetailField>
        {capability.description && <DetailField label="Description">{capability.description}</DetailField>}
        <DetailField label="Maturity Level">
          <span className={`maturity-badge ${getMaturityBadgeClass(capability.maturityLevel)}`}>
            {capability.maturityLevel || 'Not set'}
          </span>
        </DetailField>
        {capability.status && <DetailField label="Status">{capability.status}</DetailField>}
        {capability.ownershipModel && <DetailField label="Ownership Model">{capability.ownershipModel}</DetailField>}
        {capability.primaryOwner && <DetailField label="Primary Owner">{capability.primaryOwner}</DetailField>}
        {capability.eaOwner && <DetailField label="EA Owner">{capability.eaOwner}</DetailField>}
        {capability.experts && capability.experts.length > 0 && (
          <DetailField label="Experts">
            <ul className="expert-list">
              {capability.experts.map((expert, idx) => (
                <li key={idx} className="expert-item">
                  <strong>{expert.name}</strong> - {expert.role}
                  {expert.contact && <span className="expert-contact"> ({expert.contact})</span>}
                </li>
              ))}
            </ul>
          </DetailField>
        )}
        {capability.tags && capability.tags.length > 0 && (
          <DetailField label="Tags">
            <div className="tag-list">
              {capability.tags.map((tag, idx) => <span key={idx} className="tag-badge">{tag}</span>)}
            </div>
          </DetailField>
        )}
        <DetailField label="Created"><span className="detail-date">{formattedDate}</span></DetailField>
        <DetailField label="ID"><span className="detail-id">{capability.id}</span></DetailField>
      </div>

      <EditCapabilityDialog isOpen={showEditDialog} onClose={() => setShowEditDialog(false)} capability={capability} />
      <DeleteCapabilityDialog isOpen={showDeleteDialog} onClose={() => setShowDeleteDialog(false)} capability={capability} />
    </div>
  );
};
