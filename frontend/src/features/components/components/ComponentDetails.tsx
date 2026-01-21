import React, { useMemo, useState } from 'react';
import { useAppStore } from '../../../store/appStore';
import { DetailField } from '../../../components/shared/DetailField';
import { ColorPicker } from '../../../components/shared/ColorPicker';
import { ComponentFitScores } from './ComponentFitScores';
import { ComponentExpertsList } from './ComponentExpertsList';
import { AddComponentExpertDialog } from './AddComponentExpertDialog';
import { ComponentOriginsSection } from './ComponentOriginsSection';
import { AuditHistorySection } from '../../audit';
import { useCapabilities, useCapabilitiesByComponent } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../hooks/useComponents';
import { useUpdateComponentColor, useClearComponentColor } from '../../views/hooks/useViews';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { hasLink } from '../../../utils/hateoas';
import type { CapabilityRealization, Capability, ViewComponent, Component, View } from '../../../api/types';
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
  onAddExpert?: () => void;
  isAddExpertOpen?: boolean;
  onCloseAddExpert?: () => void;
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
  const canUpdateColor = componentInView._links?.['x-update-color'] !== undefined;
  const canClearColor = componentInView._links?.['x-clear-color'] !== undefined;

  if (!canUpdateColor) return null;

  return (
    <DetailField label="Custom Color">
      <div data-testid="color-picker">
        <ColorPicker
          color={currentColor}
          onChange={onColorChange}
          disabled={!isColorPickerEnabled}
          disabledTooltip="Switch to custom color scheme to assign colors"
        />
        {currentColor && canClearColor && (
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

interface TypeFieldProps {
  referenceUrl: string | undefined;
}

const TypeField: React.FC<TypeFieldProps> = ({ referenceUrl }) => {
  const hasReference = referenceUrl && referenceUrl.trim() !== '';

  return (
    <DetailField label="Type">
      {hasReference ? (
        <a href={referenceUrl} target="_blank" rel="noopener noreferrer" className="type-link">
          Application Component
        </a>
      ) : (
        'Application Component'
      )}
    </DetailField>
  );
};

interface ActionButtonsProps {
  componentId: string;
  canEdit: boolean;
  canRemoveFromView: boolean;
  onEdit: (componentId: string) => void;
  onRemoveFromView?: () => void;
}

const ActionButtons: React.FC<ActionButtonsProps> = ({ componentId, canEdit, canRemoveFromView, onEdit, onRemoveFromView }) => {
  if (!canEdit && !canRemoveFromView) return null;

  return (
    <div className="detail-actions">
      {canEdit && <button className="btn btn-secondary btn-small" onClick={() => onEdit(componentId)}>Edit</button>}
      {canRemoveFromView && onRemoveFromView && (
        <button className="btn btn-secondary btn-small" onClick={onRemoveFromView}>Remove from View</button>
      )}
    </div>
  );
};

interface ConditionalColorPickerProps {
  componentInView: ViewComponent | undefined;
  currentView: View | null;
  onColorChange?: (color: string) => void;
  onClearColor?: () => void;
}

const hasRequiredColorPickerProps = (
  componentInView: ViewComponent | undefined,
  currentView: View | null,
  onColorChange?: (color: string) => void,
  onClearColor?: () => void,
): componentInView is ViewComponent =>
  componentInView !== undefined && currentView !== null && onColorChange !== undefined && onClearColor !== undefined;

const ConditionalColorPicker: React.FC<ConditionalColorPickerProps> = ({
  componentInView,
  currentView,
  onColorChange,
  onClearColor,
}) => {
  if (!hasRequiredColorPickerProps(componentInView, currentView, onColorChange, onClearColor)) return null;

  return (
    <ColorPickerField
      componentInView={componentInView}
      colorScheme={currentView!.colorScheme || 'maturity'}
      onColorChange={onColorChange!}
      onClearColor={onClearColor!}
    />
  );
};

interface ComponentContentProps {
  component: Component;
  componentInView: ViewComponent | undefined;
  currentView: View | null;
  realizations: CapabilityRealization[];
  capabilities: Capability[];
  isInCurrentView: boolean;
  onColorChange?: (color: string) => void;
  onClearColor?: () => void;
  onEdit: (componentId: string) => void;
  onRemoveFromView?: () => void;
  onAddExpert?: () => void;
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
  onAddExpert,
}) => {
  const formattedDate = new Date(component.createdAt).toLocaleString();
  const canEdit = component._links?.edit !== undefined;
  const canRemoveFromView = isInCurrentView && componentInView?._links?.['x-remove'] !== undefined;
  const canAddExpert = hasLink(component, 'x-add-expert');

  return (
    <div className="detail-content">
      <ActionButtons
        componentId={component.id}
        canEdit={canEdit}
        canRemoveFromView={canRemoveFromView}
        onEdit={onEdit}
        onRemoveFromView={onRemoveFromView}
      />

      <DetailField label="Name">{component.name}</DetailField>

      <OptionalField value={component.description} label="Description" render={(desc) => desc} />

      {onAddExpert && (
        <ComponentExpertsList
          componentId={component.id}
          experts={component.experts}
          canAddExpert={canAddExpert}
          onAddClick={onAddExpert}
        />
      )}

      <DetailField label="Created"><span className="detail-date">{formattedDate}</span></DetailField>
      <TypeField referenceUrl={component._links.describedby?.href} />

      <ConditionalColorPicker
        componentInView={componentInView}
        currentView={currentView}
        onColorChange={onColorChange}
        onClearColor={onClearColor}
      />

      <RealizationsField realizations={realizations} capabilities={capabilities} />

      <ComponentOriginsSection componentId={component.id} />

      <ComponentFitScores componentId={component.id} />

      <AuditHistorySection aggregateId={component.id} />
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
  onAddExpert,
  isAddExpertOpen,
  onCloseAddExpert,
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
        onColorChange={onColorChange}
        onClearColor={onClearColor}
        onEdit={onEdit}
        onRemoveFromView={onRemoveFromView}
        onAddExpert={onAddExpert}
      />

      {isAddExpertOpen !== undefined && onCloseAddExpert && (
        <AddComponentExpertDialog
          isOpen={isAddExpertOpen}
          onClose={onCloseAddExpert}
          componentId={component.id}
        />
      )}
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
  const [isAddExpertOpen, setIsAddExpertOpen] = useState(false);

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
        viewId: currentView.id,
        componentId: selectedNodeId,
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
        viewId: currentView.id,
        componentId: selectedNodeId
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
      onAddExpert={() => setIsAddExpertOpen(true)}
      isAddExpertOpen={isAddExpertOpen}
      onCloseAddExpert={() => setIsAddExpertOpen(false)}
    />
  );
};
