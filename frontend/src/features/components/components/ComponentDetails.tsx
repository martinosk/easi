import React, { useMemo } from 'react';
import { useAppStore } from '../../../store/appStore';
import { DetailField } from '../../../components/shared/DetailField';
import { ColorPicker } from '../../../components/shared/ColorPicker';
import { ComponentFitScores } from './ComponentFitScores';
import { useCapabilities, useCapabilitiesByComponent } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../hooks/useComponents';
import { useUpdateComponentColor, useClearComponentColor } from '../../views/hooks/useViews';
import { useCurrentView } from '../../../hooks/useCurrentView';
import type { CapabilityRealization, Capability, ViewComponent, Component, View, ViewId, ComponentId } from '../../../api/types';
import toast from 'react-hot-toast';

interface ComponentDetailsProps {
  onEdit: (componentId: string) => void;
  onRemoveFromView?: () => void;
}

export interface ComponentDetailsContentProps {
  component: Component;
  realizations: CapabilityRealization[];
  capabilities: Capability[];
  onEdit: (componentId: string) => void;
  componentInView?: ViewComponent;
  currentView?: View | null;
  isInCurrentView?: boolean;
  onRemoveFromView?: () => void;
  onColorChange?: (color: string) => void;
  onClearColor?: () => void;
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

interface ReferenceLinkFieldProps {
  href: string | undefined;
}

const ReferenceLinkField: React.FC<ReferenceLinkFieldProps> = ({ href }) => {
  if (!href) return null;
  return (
    <div className="detail-reference">
      <a href={href} target="_blank" rel="noopener noreferrer" className="reference-link">
        Reference Documentation
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

const ComponentContentInternal: React.FC<ComponentContentProps> = ({
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

      <ReferenceLinkField href={component._links.reference} />

      {componentInView && currentView && onColorChange && onClearColor && (
        <ColorPickerField
          componentInView={componentInView}
          colorScheme={currentView.colorScheme || 'maturity'}
          onColorChange={onColorChange}
          onClearColor={onClearColor}
        />
      )}

      <RealizationsField realizations={realizations} capabilities={capabilities} />

      <ComponentFitScores componentId={component.id} />
    </div>
  );
};

export const ComponentDetailsContent: React.FC<ComponentDetailsContentProps> = ({
  component,
  realizations,
  capabilities,
  onEdit,
  componentInView,
  currentView,
  isInCurrentView = false,
  onRemoveFromView,
  onColorChange,
  onClearColor,
}) => {
  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Application Details</h3>
      </div>

      <ComponentContentInternal
        component={component}
        componentInView={componentInView}
        currentView={currentView ?? null}
        realizations={realizations}
        capabilities={capabilities}
        isInCurrentView={isInCurrentView}
        onColorChange={onColorChange ?? (() => {})}
        onClearColor={onClearColor ?? (() => {})}
        onEdit={onEdit}
        onRemoveFromView={onRemoveFromView}
      />
    </div>
  );
};

export const ComponentDetails: React.FC<ComponentDetailsProps> = ({ onEdit, onRemoveFromView }) => {
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const { data: components = [] } = useComponents();
  const { currentView } = useCurrentView();
  const { data: capabilities = [] } = useCapabilities();
  const { data: componentRealizations = [] } = useCapabilitiesByComponent(selectedNodeId ?? undefined);
  const updateComponentColorMutation = useUpdateComponentColor();
  const clearComponentColorMutation = useClearComponentColor();

  const component = useMemo(() =>
    components.find((c) => c.id === selectedNodeId),
    [components, selectedNodeId]
  );
  if (!selectedNodeId || !component) return null;

  const componentInView = currentView?.components.find((vc) => vc.componentId === selectedNodeId);
  const isInCurrentView = currentView?.components.some((vc) => vc.componentId === selectedNodeId) || false;

  const handleColorChange = async (color: string) => {
    if (!currentView) return;
    try {
      await updateComponentColorMutation.mutateAsync({
        viewId: currentView.id as ViewId,
        componentId: selectedNodeId as ComponentId,
        color
      });
    } catch {
      toast.error('Failed to update color');
    }
  };

  const handleClearColor = async () => {
    if (!currentView) return;
    try {
      await clearComponentColorMutation.mutateAsync({
        viewId: currentView.id as ViewId,
        componentId: selectedNodeId as ComponentId
      });
    } catch {
      toast.error('Failed to clear color');
    }
  };

  return (
    <ComponentDetailsContent
      component={component}
      realizations={componentRealizations}
      capabilities={capabilities}
      onEdit={onEdit}
      componentInView={componentInView}
      currentView={currentView}
      isInCurrentView={isInCurrentView}
      onRemoveFromView={onRemoveFromView}
      onColorChange={handleColorChange}
      onClearColor={handleClearColor}
    />
  );
};
