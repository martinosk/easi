import { Button, Group, Stack, Text } from '@mantine/core';
import type React from 'react';
import toast from 'react-hot-toast';
import type { ComponentId, View, ViewComponent, ViewId } from '../../../api/types';
import { toComponentId } from '../../../api/types';
import { ColorPicker } from '../../../components/shared/ColorPicker';
import { DetailField } from '../../../components/shared/DetailField';
import { hasLink } from '../../../utils/hateoas';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useClearComponentColor, useUpdateComponentColor } from '../../views/hooks/useViews';

const useComponentColorHandlers = (viewId: ViewId, componentId: ComponentId) => {
  const updateMutation = useUpdateComponentColor();
  const clearMutation = useClearComponentColor();

  const handleColorChange = async (color: string) => {
    try {
      await updateMutation.mutateAsync({ viewId, componentId, color });
    } catch {
      toast.error('Failed to update color');
    }
  };

  const handleClearColor = async () => {
    try {
      await clearMutation.mutateAsync({ viewId, componentId });
    } catch {
      toast.error('Failed to clear color');
    }
  };

  return { handleColorChange, handleClearColor };
};

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

  return (
    <DetailField label="Custom Color">
      <Stack gap="xs" data-testid="color-picker">
        <ColorPicker
          color={currentColor}
          onChange={onColorChange}
          disabled={colorScheme !== 'custom'}
          disabledTooltip="Switch to custom color scheme to assign colors"
        />
        {currentColor && hasLink(componentInView, 'x-clear-color') && (
          <Group justify="flex-start">
            <Button variant="default" size="xs" onClick={onClearColor}>
              Clear Color
            </Button>
          </Group>
        )}
      </Stack>
    </DetailField>
  );
};

interface SectionBodyProps {
  view: View;
  componentInView: ViewComponent;
  onRemoveFromView: () => void;
}

const SectionBody: React.FC<SectionBodyProps> = ({ view, componentInView, onRemoveFromView }) => {
  const { handleColorChange, handleClearColor } = useComponentColorHandlers(view.id, componentInView.componentId);
  const canUpdateColor = hasLink(componentInView, 'x-update-color');
  const canRemove = hasLink(componentInView, 'x-remove');

  if (!canUpdateColor && !canRemove) return null;

  return (
    <Stack gap="sm" data-testid="view-membership-section">
      <Text size="sm" fw={500}>
        In this view
      </Text>
      {canUpdateColor && (
        <ColorPickerField
          componentInView={componentInView}
          colorScheme={view.colorScheme || 'maturity'}
          onColorChange={handleColorChange}
          onClearColor={handleClearColor}
        />
      )}
      {canRemove && (
        <Group justify="flex-start">
          <Button variant="default" size="xs" onClick={onRemoveFromView}>
            Remove from View
          </Button>
        </Group>
      )}
    </Stack>
  );
};

interface ComponentViewMembershipSectionProps {
  componentId: string;
  onRemoveFromView: () => void;
}

export const ComponentViewMembershipSection: React.FC<ComponentViewMembershipSectionProps> = ({
  componentId,
  onRemoveFromView,
}) => {
  const { currentView } = useCurrentView();
  const componentInView = currentView?.components.find((vc) => vc.componentId === toComponentId(componentId));
  if (!currentView || !componentInView) return null;

  return <SectionBody view={currentView} componentInView={componentInView} onRemoveFromView={onRemoveFromView} />;
};
