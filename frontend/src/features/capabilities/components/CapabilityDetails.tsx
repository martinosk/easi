import React, { useState } from 'react';
import { useAppStore } from '../../../store/appStore';
import { EditCapabilityDialog } from './EditCapabilityDialog';
import { DetailField } from '../../../components/shared/DetailField';
import { ColorPicker } from '../../../components/shared/ColorPicker';
import type { Capability, Component, CapabilityRealization, Expert, View, ViewCapability } from '../../../api/types';
import toast from 'react-hot-toast';

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

interface OptionalFieldProps<T> {
  value: T | undefined | null;
  label: string;
  render: (value: T) => React.ReactNode;
}

function OptionalField<T>({ value, label, render }: OptionalFieldProps<T>): React.ReactElement | null {
  if (value === undefined || value === null) return null;
  if (Array.isArray(value) && value.length === 0) return null;
  return <DetailField label={label}>{render(value)}</DetailField>;
}

interface CapabilityFieldsProps {
  capability: Capability;
}

const CapabilityOptionalFields: React.FC<CapabilityFieldsProps> = ({ capability }) => (
  <>
    <OptionalField
      value={capability.description}
      label="Description"
      render={(desc) => desc}
    />
    <OptionalField
      value={capability.status}
      label="Status"
      render={(status) => status}
    />
    <OptionalField
      value={capability.ownershipModel}
      label="Ownership Model"
      render={(model) => model}
    />
    <OptionalField
      value={capability.primaryOwner}
      label="Primary Owner"
      render={(owner) => owner}
    />
    <OptionalField
      value={capability.eaOwner}
      label="EA Owner"
      render={(owner) => owner}
    />
    <OptionalField
      value={capability.experts}
      label="Experts"
      render={(experts) => <ExpertList experts={experts} />}
    />
    <OptionalField
      value={capability.tags}
      label="Tags"
      render={(tags) => <TagList tags={tags} />}
    />
  </>
);

interface ColorPickerFieldProps {
  capabilityInView: ViewCapability;
  currentView: View;
  onColorChange: (color: string) => void;
}

const ColorPickerField: React.FC<ColorPickerFieldProps> = ({ capabilityInView, currentView, onColorChange }) => {
  const currentColor = capabilityInView.customColor || null;
  const isColorPickerEnabled = currentView.colorScheme === 'custom';

  return (
    <DetailField label="Custom Color">
      <div data-testid="color-picker">
        <ColorPicker
          color={currentColor}
          onChange={onColorChange}
          disabled={!isColorPickerEnabled}
          disabledTooltip="Switch to custom color scheme to assign colors"
        />
      </div>
    </DetailField>
  );
};

interface RealizationsFieldProps {
  realizations: CapabilityRealization[];
  components: Component[];
}

const RealizationsField: React.FC<RealizationsFieldProps> = ({ realizations, components }) => {
  if (realizations.length === 0) return null;
  return (
    <DetailField label="Realized By">
      <RealizingComponentsList realizations={realizations} components={components} />
    </DetailField>
  );
};

interface CapabilityContentProps {
  capability: Capability;
  capabilityInView: ViewCapability | undefined;
  currentView: View | null;
  realizations: CapabilityRealization[];
  components: Component[];
  onColorChange: (color: string) => void;
  onEdit: () => void;
  onRemoveFromView: () => void;
}

const CapabilityContent: React.FC<CapabilityContentProps> = ({
  capability,
  capabilityInView,
  currentView,
  realizations,
  components,
  onColorChange,
  onEdit,
  onRemoveFromView,
}) => {
  const formattedDate = new Date(capability.createdAt).toLocaleString();

  return (
    <div className="detail-content">
      <div className="detail-actions">
        <button className="btn btn-secondary btn-small" onClick={onEdit}>Edit</button>
        <button className="btn btn-secondary btn-small" onClick={onRemoveFromView}>Remove from View</button>
      </div>

      <DetailField label="Name">{capability.name}</DetailField>
      <DetailField label="Level"><span className="level-badge">{capability.level}</span></DetailField>
      <CapabilityOptionalFields capability={capability} />
      <DetailField label="Maturity Level">
        <span className={`maturity-badge ${getMaturityBadgeClass(capability.maturityLevel)}`}>
          {capability.maturityLevel || 'Not set'}
        </span>
      </DetailField>
      <DetailField label="Created"><span className="detail-date">{formattedDate}</span></DetailField>
      <DetailField label="ID"><span className="detail-id">{capability.id}</span></DetailField>

      {capabilityInView && currentView && (
        <ColorPickerField
          capabilityInView={capabilityInView}
          currentView={currentView}
          onColorChange={onColorChange}
        />
      )}

      <RealizationsField realizations={realizations} components={components} />
    </div>
  );
};

export const CapabilityDetails: React.FC<CapabilityDetailsProps> = ({ onRemoveFromView }) => {
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const capabilities = useAppStore((state) => state.capabilities);
  const selectCapability = useAppStore((state) => state.selectCapability);
  const capabilityRealizations = useAppStore((state) => state.capabilityRealizations);
  const components = useAppStore((state) => state.components);
  const currentView = useAppStore((state) => state.currentView);
  const updateCapabilityColor = useAppStore((state) => state.updateCapabilityColor);
  const [showEditDialog, setShowEditDialog] = useState(false);

  const capability = capabilities.find((c) => c.id === selectedCapabilityId);
  if (!selectedCapabilityId || !capability) return null;

  const capabilityInView = currentView?.capabilities.find(
    (vc) => vc.capabilityId === selectedCapabilityId
  );

  const capabilityRealizationsForThis = capabilityRealizations.filter(
    (r) => r.capabilityId === capability.id
  );

  const handleColorChange = async (color: string) => {
    if (!currentView) return;

    try {
      await updateCapabilityColor(currentView.id, selectedCapabilityId, color);
    } catch {
      toast.error('Failed to update color');
    }
  };

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Capability Details</h3>
        <button className="detail-close" onClick={() => selectCapability(null)} aria-label="Close details">x</button>
      </div>

      <CapabilityContent
        capability={capability}
        capabilityInView={capabilityInView}
        currentView={currentView}
        realizations={capabilityRealizationsForThis}
        components={components}
        onColorChange={handleColorChange}
        onEdit={() => setShowEditDialog(true)}
        onRemoveFromView={onRemoveFromView}
      />

      <EditCapabilityDialog isOpen={showEditDialog} onClose={() => setShowEditDialog(false)} capability={capability} />
    </div>
  );
};
