import { Button, Group, Stack, Text } from '@mantine/core';
import type React from 'react';
import toast from 'react-hot-toast';
import type { CapabilityId, View, ViewCapability, ViewId } from '../../../api/types';
import { ColorPicker } from '../../../components/shared/ColorPicker';
import { DetailField } from '../../../components/shared/DetailField';
import { hasLink } from '../../../utils/hateoas';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useClearCapabilityColor, useUpdateCapabilityColor } from '../../views/hooks/useViews';

const useCapabilityColorHandlers = (viewId: ViewId, capabilityId: CapabilityId) => {
  const updateMutation = useUpdateCapabilityColor();
  const clearMutation = useClearCapabilityColor();

  const handleColorChange = async (color: string) => {
    try {
      await updateMutation.mutateAsync({ viewId, capabilityId, color });
    } catch {
      toast.error('Failed to update color');
    }
  };

  const handleClearColor = async () => {
    try {
      await clearMutation.mutateAsync({ viewId, capabilityId });
    } catch {
      toast.error('Failed to clear color');
    }
  };

  return { handleColorChange, handleClearColor };
};

interface ColorPickerFieldProps {
  capabilityInView: ViewCapability;
  colorScheme: string;
  onColorChange: (color: string) => void;
  onClearColor: () => void;
}

const ColorPickerField: React.FC<ColorPickerFieldProps> = ({
  capabilityInView,
  colorScheme,
  onColorChange,
  onClearColor,
}) => {
  const currentColor = capabilityInView.customColor || null;

  return (
    <DetailField label="Custom Color">
      <Stack gap="xs" data-testid="color-picker">
        <ColorPicker
          color={currentColor}
          onChange={onColorChange}
          disabled={colorScheme !== 'custom'}
          disabledTooltip="Switch to custom color scheme to assign colors"
        />
        {currentColor && hasLink(capabilityInView, 'x-clear-color') && (
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
  capabilityInView: ViewCapability;
  onRemoveFromView: () => void;
}

const SectionBody: React.FC<SectionBodyProps> = ({ view, capabilityInView, onRemoveFromView }) => {
  const { handleColorChange, handleClearColor } = useCapabilityColorHandlers(view.id, capabilityInView.capabilityId);
  const canUpdateColor = hasLink(capabilityInView, 'x-update-color');
  const canRemove = hasLink(capabilityInView, 'x-remove');

  if (!canUpdateColor && !canRemove) return null;

  return (
    <Stack gap="sm" data-testid="view-membership-section">
      <Text size="sm" fw={500}>
        In this view
      </Text>
      {canUpdateColor && (
        <ColorPickerField
          capabilityInView={capabilityInView}
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

interface CapabilityViewMembershipSectionProps {
  capabilityId: CapabilityId;
  onRemoveFromView: () => void;
}

export const CapabilityViewMembershipSection: React.FC<CapabilityViewMembershipSectionProps> = ({
  capabilityId,
  onRemoveFromView,
}) => {
  const { currentView } = useCurrentView();
  const capabilityInView = currentView?.capabilities.find((vc) => vc.capabilityId === capabilityId);
  if (!currentView || !capabilityInView) return null;

  return <SectionBody view={currentView} capabilityInView={capabilityInView} onRemoveFromView={onRemoveFromView} />;
};
