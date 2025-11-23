import React, { useState } from 'react';
import { useAppStore } from '../store/appStore';
import { EditCapabilityDialog } from './EditCapabilityDialog';
import { DetailField } from './DetailField';
import type { Component, CapabilityRealization, Expert } from '../api/types';

interface CapabilityDetailsProps {
  onRemoveFromView: () => void;
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

const getLevelBadge = (level: string): string => {
  const badges: Record<string, string> = {
    Full: '100%',
    Partial: 'Partial',
    Planned: 'Planned',
  };
  return badges[level] || level;
};

const getComponentName = (components: Component[], componentId: string): string => {
  const comp = components.find((c) => c.id === componentId);
  return comp?.name || 'Unknown';
};

const ExpertList: React.FC<{ experts: Expert[] }> = ({ experts }) => (
  <ul className="expert-list">
    {experts.map((expert, idx) => (
      <li key={idx} className="expert-item">
        <strong>{expert.name}</strong> - {expert.role}
        {expert.contact && <span className="expert-contact"> ({expert.contact})</span>}
      </li>
    ))}
  </ul>
);

const TagList: React.FC<{ tags: string[] }> = ({ tags }) => (
  <div className="tag-list">
    {tags.map((tag, idx) => <span key={idx} className="tag-badge">{tag}</span>)}
  </div>
);

interface RealizingComponentsProps {
  realizations: CapabilityRealization[];
  components: Component[];
}

const RealizingComponentsList: React.FC<RealizingComponentsProps> = ({ realizations, components }) => (
  <ul className="realization-list">
    {realizations.map((r) => (
      <li key={r.id} className="realization-item">
        <span className="realization-name">{getComponentName(components, r.componentId)}</span>
        <span className="realization-level">{getLevelBadge(r.realizationLevel)}</span>
      </li>
    ))}
  </ul>
);

export const CapabilityDetails: React.FC<CapabilityDetailsProps> = ({ onRemoveFromView }) => {
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const capabilities = useAppStore((state) => state.capabilities);
  const selectCapability = useAppStore((state) => state.selectCapability);
  const capabilityRealizations = useAppStore((state) => state.capabilityRealizations);
  const components = useAppStore((state) => state.components);
  const [showEditDialog, setShowEditDialog] = useState(false);

  const capability = capabilities.find((c) => c.id === selectedCapabilityId);
  if (!selectedCapabilityId || !capability) return null;

  const formattedDate = new Date(capability.createdAt).toLocaleString();
  const capabilityRealizationsForThis = capabilityRealizations.filter(
    (r) => r.capabilityId === capability.id
  );

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Capability Details</h3>
        <button className="detail-close" onClick={() => selectCapability(null)} aria-label="Close details">x</button>
      </div>

      <div className="detail-content">
        <div className="detail-actions">
          <button className="btn btn-secondary btn-small" onClick={() => setShowEditDialog(true)}>Edit</button>
          <button className="btn btn-secondary btn-small" onClick={onRemoveFromView}>Remove from View</button>
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
            <ExpertList experts={capability.experts} />
          </DetailField>
        )}
        {capability.tags && capability.tags.length > 0 && (
          <DetailField label="Tags">
            <TagList tags={capability.tags} />
          </DetailField>
        )}
        <DetailField label="Created"><span className="detail-date">{formattedDate}</span></DetailField>
        <DetailField label="ID"><span className="detail-id">{capability.id}</span></DetailField>

        {capabilityRealizationsForThis.length > 0 && (
          <DetailField label="Realized By">
            <RealizingComponentsList realizations={capabilityRealizationsForThis} components={components} />
          </DetailField>
        )}
      </div>

      <EditCapabilityDialog isOpen={showEditDialog} onClose={() => setShowEditDialog(false)} capability={capability} />
    </div>
  );
};
