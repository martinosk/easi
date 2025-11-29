import React from 'react';
import { useAppStore } from '../../../store/appStore';
import { DetailField } from '../../../components/shared/DetailField';
import { ColorPicker } from '../../../components/shared/ColorPicker';
import type { CapabilityRealization, Capability, ViewComponent, Component, View } from '../../../api/types';
import toast from 'react-hot-toast';

interface ComponentDetailsProps {
  onEdit: (componentId: string) => void;
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

interface ColorPickerFieldProps {
  componentInView: ViewComponent;
  colorScheme: string;
  onColorChange: (color: string) => void;
  onClearColor: () => void;
}

const ColorPickerField: React.FC<ColorPickerFieldProps> = ({
  componentInView,
  colorScheme,
  onColorChange,
  onClearColor,
}) => {
  const currentColor = componentInView.customColor || null;
  const isColorPickerEnabled = colorScheme === 'custom';

  return (
    <DetailField label="Custom Color">
      <div data-testid="color-picker">
        <ColorPicker
          color={currentColor}
          onChange={onColorChange}
          disabled={!isColorPickerEnabled}
          disabledTooltip="Switch to custom color scheme to assign colors"
        />
        {currentColor && (
          <button
            className="btn btn-secondary btn-small"
            onClick={onClearColor}
            style={{ marginTop: '8px' }}
          >
            Clear Color
          </button>
        )}
      </div>
    </DetailField>
  );
};

interface RealizationsFieldProps {
  realizations: CapabilityRealization[];
  capabilities: Capability[];
}

const RealizationsField: React.FC<RealizationsFieldProps> = ({ realizations, capabilities }) => {
  if (realizations.length === 0) return null;

  const directRealizations = realizations.filter((r) => r.origin === 'Direct');
  const inheritedRealizations = realizations.filter((r) => r.origin === 'Inherited');

  return (
    <DetailField label="Realizes Capabilities">
      <ul className="realization-list">
        <RealizationListItems realizations={directRealizations} capabilities={capabilities} origin="Direct" />
        <RealizationListItems realizations={inheritedRealizations} capabilities={capabilities} origin="Inherited" />
      </ul>
    </DetailField>
  );
};

interface ArchimateLinkFieldProps {
  href: string | undefined;
}

const ArchimateLinkField: React.FC<ArchimateLinkFieldProps> = ({ href }) => {
  if (!href) return null;
  return (
    <div className="detail-archimate">
      <a href={href} target="_blank" rel="noopener noreferrer" className="archimate-link">
        ArchiMate Documentation
      </a>
    </div>
  );
};

interface ActionButtonsProps {
  componentId: string;
  isInCurrentView: boolean;
  onEdit: (componentId: string) => void;
  onRemoveFromView?: () => void;
}

const ActionButtons: React.FC<ActionButtonsProps> = ({ componentId, isInCurrentView, onEdit, onRemoveFromView }) => (
  <div className="detail-actions">
    <button className="btn btn-secondary btn-small" onClick={() => onEdit(componentId)}>Edit</button>
    {isInCurrentView && onRemoveFromView && (
      <button className="btn btn-secondary btn-small" onClick={onRemoveFromView}>Remove from View</button>
    )}
  </div>
);

interface ComponentContentProps {
  component: Component;
  componentInView: ViewComponent | undefined;
  currentView: View | null;
  realizations: CapabilityRealization[];
  capabilities: Capability[];
  isInCurrentView: boolean;
  onColorChange: (color: string) => void;
  onClearColor: () => void;
  onEdit: (componentId: string) => void;
  onRemoveFromView?: () => void;
}

const ComponentContent: React.FC<ComponentContentProps> = ({
  component,
  componentInView,
  currentView,
  realizations,
  capabilities,
  isInCurrentView,
  onColorChange,
  onClearColor,
  onEdit,
  onRemoveFromView,
}) => {
  const formattedDate = new Date(component.createdAt).toLocaleString();

  return (
    <div className="detail-content">
      <ActionButtons
        componentId={component.id}
        isInCurrentView={isInCurrentView}
        onEdit={onEdit}
        onRemoveFromView={onRemoveFromView}
      />

      <DetailField label="Name">{component.name}</DetailField>

      <OptionalField value={component.description} label="Description" render={(desc) => desc} />

      <DetailField label="Created"><span className="detail-date">{formattedDate}</span></DetailField>
      <DetailField label="Type">Application Component</DetailField>
      <DetailField label="ID"><span className="detail-id">{component.id}</span></DetailField>

      <ArchimateLinkField href={component._links.archimate} />

      {componentInView && currentView && (
        <ColorPickerField
          componentInView={componentInView}
          colorScheme={currentView.colorScheme || 'archimate'}
          onColorChange={onColorChange}
          onClearColor={onClearColor}
        />
      )}

      <RealizationsField realizations={realizations} capabilities={capabilities} />
    </div>
  );
};

export const ComponentDetails: React.FC<ComponentDetailsProps> = ({ onEdit, onRemoveFromView }) => {
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const components = useAppStore((state) => state.components);
  const currentView = useAppStore((state) => state.currentView);
  const clearSelection = useAppStore((state) => state.clearSelection);
  const capabilityRealizations = useAppStore((state) => state.capabilityRealizations);
  const capabilities = useAppStore((state) => state.capabilities);
  const updateComponentColor = useAppStore((state) => state.updateComponentColor);
  const clearComponentColor = useAppStore((state) => state.clearComponentColor);

  const component = components.find((c) => c.id === selectedNodeId);
  if (!selectedNodeId || !component) return null;

  const componentInView = currentView?.components.find((vc) => vc.componentId === selectedNodeId);
  const isInCurrentView = currentView?.components.some((vc) => vc.componentId === selectedNodeId) || false;

  const componentRealizations = capabilityRealizations.filter((r) => r.componentId === component.id);

  const handleColorChange = async (color: string) => {
    if (!currentView) return;
    try {
      await updateComponentColor(currentView.id, selectedNodeId, color);
    } catch {
      toast.error('Failed to update color');
    }
  };

  const handleClearColor = async () => {
    if (!currentView) return;
    try {
      await clearComponentColor(currentView.id, selectedNodeId);
    } catch {
      toast.error('Failed to clear color');
    }
  };

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Application Details</h3>
        <button className="detail-close" onClick={clearSelection} aria-label="Close details">x</button>
      </div>

      <ComponentContent
        component={component}
        componentInView={componentInView}
        currentView={currentView}
        realizations={componentRealizations}
        capabilities={capabilities}
        isInCurrentView={isInCurrentView}
        onColorChange={handleColorChange}
        onClearColor={handleClearColor}
        onEdit={onEdit}
        onRemoveFromView={onRemoveFromView}
      />
    </div>
  );
};
