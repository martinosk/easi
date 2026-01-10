import React, { useState, useMemo } from 'react';
import { useAppStore } from '../../../store/appStore';
import { EditCapabilityDialog } from './EditCapabilityDialog';
import { DetailField } from '../../../components/shared/DetailField';
import { ColorPicker } from '../../../components/shared/ColorPicker';
import { RealizationFitContext } from './RealizationFitContext';
import { AuditHistorySection } from '../../audit';
import { useCapabilities, useCapabilityRealizations } from '../hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useUpdateCapabilityColor } from '../../views/hooks/useViews';
import { useCurrentView } from '../../../hooks/useCurrentView';
import { useMaturityScale } from '../../../hooks/useMaturityScale';
import { useStrategyImportanceByCapability } from '../../business-domains/hooks/useStrategyImportance';
import { deriveLegacyMaturityValue, getDefaultSections } from '../../../utils/maturity';
import type { Capability, Component, CapabilityRealization, Expert, View, ViewCapability, ViewId, CapabilityId, StrategyImportance, ComponentId } from '../../../api/types';
import toast from 'react-hot-toast';

interface CapabilityDetailsProps {
  onRemoveFromView: () => void;
}

const getMaturityBadgeClass = (maturityLevel?: string): string => {
  const level = maturityLevel?.toLowerCase();
  const maturityClasses: Record<string, string> = {
    'genesis': 'badge-genesis',
    'custom build': 'badge-custom-build',
    'custom built': 'badge-custom-build',
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
  capabilityId: CapabilityId;
  importanceRatings: StrategyImportance[];
}

const RealizingComponentsList: React.FC<RealizingComponentsProps> = ({
  realizations,
  components,
  capabilityId,
  importanceRatings,
}) => {
  const uniqueDomainIds = useMemo(() => {
    const ids = new Set(importanceRatings.map((r) => r.businessDomainId));
    return Array.from(ids);
  }, [importanceRatings]);

  if (uniqueDomainIds.length === 0) {
    return (
      <ul className="realization-list">
        {realizations.map((r) => (
          <li key={r.id} className="realization-item-with-fit">
            <div className="realization-header">
              <span className="realization-name">{getComponentName(components, r.componentId)}</span>
              <span className="realization-level">{getLevelBadge(r.realizationLevel)}</span>
            </div>
          </li>
        ))}
      </ul>
    );
  }

  return (
    <ul className="realization-list">
      {realizations.map((r) => (
        <li key={r.id} className="realization-item-with-fit">
          <div className="realization-header">
            <span className="realization-name">{getComponentName(components, r.componentId)}</span>
            <span className="realization-level">{getLevelBadge(r.realizationLevel)}</span>
          </div>
          {uniqueDomainIds.map((domainId) => (
            <RealizationFitContext
              key={`${r.componentId}-${domainId}`}
              componentId={r.componentId as ComponentId}
              capabilityId={capabilityId}
              businessDomainId={domainId}
            />
          ))}
        </li>
      ))}
    </ul>
  );
};

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
  capabilityId: CapabilityId;
  importanceRatings: StrategyImportance[];
}

const RealizationsField: React.FC<RealizationsFieldProps> = ({
  realizations,
  components,
  capabilityId,
  importanceRatings,
}) => {
  if (realizations.length === 0) return null;
  return (
    <DetailField label="Realized By">
      <RealizingComponentsList
        realizations={realizations}
        components={components}
        capabilityId={capabilityId}
        importanceRatings={importanceRatings}
      />
    </DetailField>
  );
};

interface CapabilityContentProps {
  capability: Capability;
  capabilityInView: ViewCapability | undefined;
  currentView: View | null;
  realizations: CapabilityRealization[];
  components: Component[];
  importanceRatings: StrategyImportance[];
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
  importanceRatings,
  onColorChange,
  onEdit,
  onRemoveFromView,
}) => {
  const formattedDate = new Date(capability.createdAt).toLocaleString();
  const { data: maturityScale } = useMaturityScale();
  const sections = maturityScale?.sections ?? getDefaultSections();

  const effectiveMaturityValue = capability.maturityValue ??
    (capability.maturityLevel ? deriveLegacyMaturityValue(capability.maturityLevel, sections) : 12);
  const sectionName = sections.find(s => effectiveMaturityValue >= s.minValue && effectiveMaturityValue <= s.maxValue)?.name ||
    'Unknown';
  const maturityDisplay = `${sectionName} (${effectiveMaturityValue})`;

  const canEdit = capability._links?.edit !== undefined;
  const canRemoveFromView = capabilityInView?._links?.['x-remove'] !== undefined;
  const showActionButtons = canEdit || canRemoveFromView;

  return (
    <div className="detail-content">
      {showActionButtons && (
        <div className="detail-actions">
          {canEdit && <button className="btn btn-secondary btn-small" onClick={onEdit}>Edit</button>}
          {canRemoveFromView && <button className="btn btn-secondary btn-small" onClick={onRemoveFromView}>Remove from View</button>}
        </div>
      )}

      <DetailField label="Name">{capability.name}</DetailField>
      <DetailField label="Level"><span className="level-badge">{capability.level}</span></DetailField>
      <CapabilityOptionalFields capability={capability} />
      <DetailField label="Maturity Level">
        <span className={`maturity-badge ${getMaturityBadgeClass(sectionName)}`}>
          {maturityDisplay}
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

      <RealizationsField
        realizations={realizations}
        components={components}
        capabilityId={capability.id as CapabilityId}
        importanceRatings={importanceRatings}
      />

      <AuditHistorySection aggregateId={capability.id} />
    </div>
  );
};

export const CapabilityDetails: React.FC<CapabilityDetailsProps> = ({ onRemoveFromView }) => {
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const { data: capabilities = [] } = useCapabilities();
  const { data: components = [] } = useComponents();
  const { currentView } = useCurrentView();
  const { data: capabilityRealizationsForThis = [] } = useCapabilityRealizations(selectedCapabilityId ?? undefined);
  const { data: importanceRatings = [] } = useStrategyImportanceByCapability(selectedCapabilityId as CapabilityId | undefined);
  const updateCapabilityColorMutation = useUpdateCapabilityColor();
  const [showEditDialog, setShowEditDialog] = useState(false);

  const capability = capabilities.find((c) => c.id === selectedCapabilityId);
  if (!selectedCapabilityId || !capability) return null;

  const capabilityInView = currentView?.capabilities.find(
    (vc) => vc.capabilityId === selectedCapabilityId
  );

  const handleColorChange = async (color: string) => {
    if (!currentView) return;

    try {
      await updateCapabilityColorMutation.mutateAsync({
        viewId: currentView.id as ViewId,
        capabilityId: selectedCapabilityId as CapabilityId,
        color
      });
    } catch {
      toast.error('Failed to update color');
    }
  };

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Capability Details</h3>
      </div>

      <CapabilityContent
        capability={capability}
        capabilityInView={capabilityInView}
        currentView={currentView}
        realizations={capabilityRealizationsForThis}
        components={components}
        importanceRatings={importanceRatings}
        onColorChange={handleColorChange}
        onEdit={() => setShowEditDialog(true)}
        onRemoveFromView={onRemoveFromView}
      />

      <EditCapabilityDialog isOpen={showEditDialog} onClose={() => setShowEditDialog(false)} capability={capability} />
    </div>
  );
};
